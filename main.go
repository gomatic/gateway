package main

import (
	"github.com/gomatic/servicer/gateway"
	"github.com/urfave/cli"
)

//
const VERSION = "1.3"

//
func main() {
	gateway.Main(run, func(app *cli.App) error {
		app.Name = "gateway"
		app.Usage = "A microservice gateway."
		app.Version = VERSION
		return nil
	})
}
