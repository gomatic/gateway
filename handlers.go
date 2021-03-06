package main

import (
	"fmt"
	"log"
	"net/http"
)

//
func ok(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s\n", version)
	return
}

//
func notFound(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	return
}

//
func robots(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "User-agent: *\nDisallow: /")
	return
}

//
func fail(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, "Failing...")
	log.Panicf("failure request: %+v", req)
	return
}
