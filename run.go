package main

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/didip/tollbooth"
	"github.com/gomatic/servicer"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
)

var (
	signingSecret = []byte("secret")
	signingMethod = jwt.SigningMethodHS256
	allowOrigins  = []string{}
)

func init() {
	if s, exists := os.LookupEnv("SECRET"); exists {
		signingSecret = []byte(s)
	}
}

//
func named(path string) (string, string, string, error) {
	parts := strings.Split(path, "/")
	version := parts[1]

	if len(parts) < 2 {
		return "", "", "", errors.New("Invalid gateway path")
	}

	if len(version) < 1 || version != "v1" {
		return "", "", "", errors.New("Invalid service version")
	}

	return version, parts[2], strings.Join(parts[3:], "/"), nil
}

//
func run(settings servicer.Settings) error {

	name = settings.Name

	//
	v1, err := v1(settings)
	if err != nil {
		return nil
	}

	//
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return signingSecret, nil
		},
		SigningMethod: signingMethod,
	})

	// Unauthenticated routes

	root := http.NewServeMux()

	root.HandleFunc("/ok", ok)
	root.HandleFunc("/index.html", ok)
	root.HandleFunc("/health", ok)
	root.HandleFunc("/health.html", ok)
	root.HandleFunc("/health/", ok)
	root.HandleFunc("/healthcheck", ok)
	root.HandleFunc("/healthcheck.html", ok)
	root.HandleFunc("/healthcheck/", ok)
	root.HandleFunc("/robots.txt", robots)
	root.HandleFunc("/", notFound)

	// Testing/debug routes

	if settings.Output.Debugging {
		root.HandleFunc("/fail", fail)
	}

	// Authenticated and secured routes

	secure := negroni.New()
	recovery := negroni.NewRecovery()
	recovery.PrintStack = false
	secure.Use(recovery)
	secure.Use(negroni.NewLogger())
	secure.UseHandlerFunc(headered)
	if len(allowOrigins) > 0 {
		secure.Use(cors.New(cors.Options{
			AllowedOrigins:   allowOrigins,
			AllowCredentials: true,
			AllowedHeaders:   []string{"Content-Type"},
			AllowedMethods:   []string{"HEAD", "GET", "POST", "OPTIONS", "DELETE", "PATCH"},
			MaxAge:           300,
			Debug:            settings.Output.Debugging,
		}))
	}
	secure.UseFunc(jwtMiddleware.HandlerWithNext)
	limit := 20
	if limited, exists := os.LookupEnv("RATE_LIMIT"); exists {
		if l, err := strconv.Atoi(limited); err == nil {
			limit = l
		}
	}
	secure.Use(gatewayFailureCache(http.HandlerFunc(v1)))

	root.Handle("/v1/validate", negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		negroni.Wrap(http.HandlerFunc(validate)),
	))

	root.Handle("/v1/", tollbooth.LimitHandler(tollbooth.NewLimiter(int64(limit), time.Second), secure))

	s := &http.Server{
		Addr:           settings.Api.String(),
		Handler:        root,
		ReadTimeout:    settings.Timeout.Read,
		WriteTimeout:   settings.Timeout.Write,
		MaxHeaderBytes: 1 << 20,
	}

	return s.ListenAndServe()
}
