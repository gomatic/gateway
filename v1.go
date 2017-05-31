package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"github.com/gomatic/servicer"
	"github.com/vulcand/oxy/forward"
)

//
func v1(settings servicer.Settings) (func(w http.ResponseWriter, req *http.Request), error) {
	name = settings.Name
	thisPort := strconv.Itoa(settings.Api.Port)
	fwd, err := forward.New()
	if err != nil {
		return nil, err
	}

	return func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		_, name, servicePath, err := named(path)
		if err != nil {
			log.Printf("ERROR %+v: Invalid request: %+v : %+v", http.StatusInternalServerError, req, path)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		host, port, domain := "", "", settings.Dns.Namespace+"."+settings.Dns.Domain

		// Adhere to: http://kubernetes.io/docs/user-guide/services/#dns
		// {{.Name}}.{{.Namespace}}.svc.cluster.local
		// e.g. business.dev.svc.cluster.local
		if cname, addrs, err := net.LookupSRV(name, "tcp", domain); err != nil {
			log.Print(err)
		} else {
			log.Printf("SRV: %s:%+v", cname, addrs)
			host = cname
			if len(addrs) != 0 {
				// TODO better handling of multiple addresses
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
			p, exists := os.LookupEnv(fmt.Sprintf("%s_SERVICE_PORT", strings.ToUpper(name)))
			if !exists {
				log.Printf("ERROR %+v: No port registered for %v", http.StatusNotFound, name)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			port = p
		}

		if port == thisPort {
			if host == "127.0.0.1" || host == "localhost" || host == "gateway" {
				log.Printf("ERROR %+v: This request will call itself: %+v", http.StatusConflict, req)
				w.WriteHeader(http.StatusConflict)
				return
			}
		}
		from := req.URL
		to := fmt.Sprintf("http://%s:%s/%s?%s", host, port, servicePath, from.RawQuery)
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
		req.RequestURI = uri.RequestURI()
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
	}, nil
}
