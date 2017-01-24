package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
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
	"github.com/vulcand/oxy/forward"
)

var (
	signingSecret = []byte("secret")
	signingMethod = jwt.SigningMethodHS256
	allowOrigins  = []string{}
	iss           = "gateway"
)

func init() {
	if s, exists := os.LookupEnv("SECRET"); exists {
		signingSecret = []byte(s)
	}
}

//
func named(path string) (string, error) {
	parts := strings.Split(path, "/")
	version := parts[1]

	if len(parts) < 2 || len(version) < 1 || version != "v1" {
		return "", errors.New("")
	}

	return parts[2], nil
}

//
func run(settings servicer.Settings) error {

	name = settings.Name

	// Ensure that the deafult port is not the port of this servicer.
	notApi := 3000
	if settings.Api.Port == notApi {
		notApi = 5000
	}
	thisPort := strconv.Itoa(settings.Api.Port)
	defaultPort := strconv.Itoa(notApi)
	fwd, err := forward.New()
	if err != nil {
		return err
	}

	//
	v1 := func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		name, err := named(path)
		if err != nil {
			log.Printf("ERROR %+v: Invalid request: %+v : %+v", http.StatusInternalServerError, req, path)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		host, port, domain := "", "", settings.Dns.Namespace + "." + settings.Dns.Domain

		// Adhere to: http://kubernetes.io/docs/user-guide/services/#dns
		// {{.Name}}.{{.Namespace}}.svc.cluster.local
		// e.g. business.dev.svc.cluster.local
		if cname, addrs, err := net.LookupSRV(name, "tcp", domain); err != nil {
			log.Print(err)
		} else {
			log.Printf("SRV: %s:%+v", cname, addrs)
			host = cname
			if len(addrs) != 0 {
				port = strconv.Itoa(int(addrs[0].Port))
			}
		}
		if host == "" {
			if h, exists := os.LookupEnv(fmt.Sprintf("%s_SERVICE_HOST", strings.ToUpper(name))); !exists {
				host = "127.0.0.1"
			} else {
				host = h
			}
		}
		if port == "" {
			if p, exists := os.LookupEnv(fmt.Sprintf("%s_SERVICE_PORT", strings.ToUpper(name))); !exists {
				port = defaultPort
			} else {
				port = p
			}
		}

		if port == thisPort {
			if host == "127.0.0.1" || host == "localhost" || host == "gateway" {
				log.Printf("ERROR %+v: This request will call itself: %+v", http.StatusConflict, req)
				w.WriteHeader(http.StatusConflict)
				return
			}
		}
		from := req.URL
		to := fmt.Sprintf("http://%s:%s%s?%s", host, port, from.Path, from.RawQuery)
		log.Printf("forwarding %s to %s", from, to)

		uri, err := url.ParseRequestURI(to)
		if err != nil {
			log.Printf("ERROR %+v: %v", http.StatusInternalServerError, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		hs := w.Header()
		hs.Set("Content-Type", "application/json")

		req.URL = uri
		if settings.Output.Mocking {
			// TODO forward to a mock service that can be configured to return different results.
			m := mock{
				Settings: settings,
				Forward: forwarding{
					Host:    host,
					Port:    port,
					Domain:  domain,
					From:    from,
					To:      to,
					Uri:     uri,
					Headers: hs,
				},
			}
			fmt.Fprintln(w, m.String())
		} else {
			fwd.ServeHTTP(w, req)
		}
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
	root.HandleFunc("/health", ok)
	root.HandleFunc("/health/", ok)
	root.HandleFunc("/v1/health", ok)
	root.HandleFunc("/v1/health/", ok)
	root.HandleFunc("/robots.txt", robots)
	root.HandleFunc("/", notFound)

	// Testing/debug routes

	if settings.Output.Debugging {
		root.HandleFunc("/token", token)
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
