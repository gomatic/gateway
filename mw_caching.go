package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/urfave/negroni"
)

// Caches 502 failures for five minutes.
func gatewayFailureCache(f http.HandlerFunc) negroni.HandlerFunc {
	timeout := 2 * time.Minute                   // TODO parameterize this 2m timeout
	routes := cache.New(timeout, 30*time.Second) // TODO parameterize this 30s interval

	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		path := r.URL.Path
		_, name, _, err := named(path)
		if when, hit := routes.Get(name); hit && err == nil {
			until := -time.Now().Sub(when.(time.Time).Add(timeout))
			hs := w.Header()
			hs.Set("Cache-Control", fmt.Sprintf("public, max-age=%v", int(until.Seconds())))
			w.WriteHeader(502)

			fmt.Fprintln(w, "Bad Gateway")
			log.Printf(
				"502 cached %s for route %s: %+v remaining",
				name,
				path,
				until,
			)
		} else {
			f(w, r)
			switch rw := w.(type) {
			case negroni.ResponseWriter:
				if 502 == rw.Status() {
					routes.Add(name, time.Now(), timeout)
				}
			default:
				log.Printf("%+v", w.Header())
			}
		}
		next(w, r)
	}
}
