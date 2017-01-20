package main

import (
	"github.com/gomatic/servicer/gateway"
)

//
const MAJOR = "1.3"

//
func main() {
	gateway.Main(run, "", "")
}
