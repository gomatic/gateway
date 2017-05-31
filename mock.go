package main

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gomatic/servicer"
)

//
type forwarding struct {
	Host, Port, Domain string
	From, Uri          *url.URL
	To                 string
	Headers            http.Header
}

//
type mock struct {
	Settings servicer.Settings
	Forward  forwarding
}

//
func (m mock) String() string {
	j, _ := json.Marshal(m)
	return string(j)
}
