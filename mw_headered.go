package main

import (
	"crypto/md5"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gomatic/servicer"
	"github.com/gorilla/context"
)

var version = fmt.Sprintf("%s.%s\n", VERSION, servicer.VERSION)

//
func headered(w http.ResponseWriter, req *http.Request) {
	now := time.Now()
	utc := now.UTC()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).
		AddDate(0, 0, 1)

	nows := fmt.Sprintf("%8x", utc.UnixNano())
	rnds := fmt.Sprintf("%8x", rand.Int63())
	url5 := md5.Sum([]byte(req.URL.String()))
	pth5 := md5.Sum([]byte(req.URL.Path))
	qry5 := md5.Sum([]byte(fmt.Sprintf("%+v", req.URL.Query())))
	rnd5 := md5.Sum([]byte(fmt.Sprintf("%v.%v", nows, rnds)))

	// ETags are currently just generated daily based on the time and the URL.
	etag := fmt.Sprintf(`W/"%x"`, md5.Sum([]byte(fmt.Sprintf("%v.%v", tomorrow, url5))))

	ifNoneMatch := req.Header.Get("If-None-Match")
	if etag == ifNoneMatch {
		log.Println("Not Modified")
		http.Error(w, "Not Modified", http.StatusNotModified)
		return
	}
	hs := w.Header()

	hs.Set("Etag", etag)

	requestId := fmt.Sprintf("%x%x%x%x", pth5[:4], qry5[:4], url5[:4], rnd5[:4])
	context.Set(req, "X-Request-Id", requestId)

	hs.Set("Server", "gw/sp")

	hs.Set("X-Powered-By", version)
	hs.Set("X-Request-Id", requestId)

	hs.Set("Last-Modified", utc.Format(time.RFC850))
}
