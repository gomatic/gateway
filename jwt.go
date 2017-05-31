package main

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
)

//
func getString(claims jwt.MapClaims, name string) string {
	if claim, exists := claims[name]; exists {
		return fmt.Sprintf("%+v", claim)
	}
	return ""
}

//
func getTime(claims jwt.MapClaims, name string) int64 {
	if claim, exists := claims[name]; exists {
		switch t := claim.(type) {
		case string:
			i, err := strconv.ParseInt(t, 10, 64)
			if err == nil {
				return int64(i)
			}
		case int64:
			return int64(t)
		case int32:
			return int64(t)
		case int8:
			return int64(t)
		case int:
			return int64(t)
		case float64:
			return int64(t)
		case float32:
			return int64(t)
		default:
			log.Printf("claim is not an integer: %[1]T %[1]v", t)
		}
	}
	return -1
}

//
func validate(w http.ResponseWriter, req *http.Request) {
	user := context.Get(req, "user")
	if user == nil {
		w.WriteHeader(http.StatusNotAcceptable)
		fmt.Fprintln(w, "Not authenticated.")
		log.Println("No 'user' on request.")
		return
	}
	token, ok := user.(*jwt.Token)
	if !ok {
		w.WriteHeader(http.StatusNotAcceptable)
		fmt.Fprintln(w, "Not authenticated.")
		log.Printf("Unexpected token type. %+v", user)
		return
	}

	var (
		claim jwt.MapClaims
		std   jwt.StandardClaims
	)
	switch c := token.Claims.(type) {
	case jwt.StandardClaims:
		std = c
		claim = jwt.MapClaims{
			"aud": c.Audience,
			"exp": c.ExpiresAt,
			"jti": c.Id,
			"iat": c.IssuedAt,
			"iss": c.Issuer,
			"nbf": c.NotBefore,
			"sub": c.Subject,
		}
	case jwt.MapClaims:
		claim = c
		std = jwt.StandardClaims{
			Audience:  getString(c, "aud"),
			ExpiresAt: getTime(c, "exp"),
			Id:        getString(c, "jti"),
			IssuedAt:  getTime(c, "iat"),
			Issuer:    getString(c, "iss"),
			NotBefore: getTime(c, "nbf"),
			Subject:   getString(c, "sub"),
		}
	default:
		log.Printf("Unexpected claim: %+v\n", claim)
		return
	}

	// Validate aud

	provided, err := hex.DecodeString(std.Audience)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		fmt.Fprintln(w, "Not authenticated.")
		log.Printf("Insecure audience. %+v", user)
		return
	}

	var auth string

	if a, exists := claim["auth"]; exists {
		auth, ok = a.(string)
		if !ok {
			log.Printf("auth key type-problem %[1]T %+[1]v", a)
		}
	}

	h := hmac.New(md5.New, []byte(signingSecret))
	h.Write([]byte(std.Subject))
	h.Write([]byte(auth))
	h.Write([]byte(std.Issuer))
	h.Write([]byte(fmt.Sprintf("%d", std.IssuedAt)))
	expected := h.Sum(nil)
	if !hmac.Equal(provided, expected) {
		w.WriteHeader(http.StatusNotAcceptable)
		fmt.Fprintln(w, "Not authenticated.")
		log.Printf("Invalid audience. %+v", user)
		return
	}

	now := time.Now().UTC()
	exp := fmt.Sprintf("%+v", time.Unix(std.ExpiresAt, 0).Sub(now))
	iat := fmt.Sprintf("%+v", time.Unix(std.IssuedAt, 0).Sub(now))
	nbf := fmt.Sprintf("%+v", time.Unix(std.NotBefore, 0).Sub(now))
	rel := fmt.Sprintf("\n\texp: %v\n\tiat: %v\n\tnbf: %v", exp, iat, nbf)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Authenticated.%s\n", rel)
	log.Printf("Authenticated. Claim: %+v%s", claim, rel)
	return
}
