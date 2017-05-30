package main

import (
	"github.com/gomatic/go-vbuild"
	"github.com/gomatic/servicer/gateway"
	"github.com/urfave/cli"
)

//
func main() {
	build.Version.Update(1, 3)
	gateway.Main(run, func(app *cli.App) error {
		app.Name = "gateway"
		app.Usage = "A microservice gateway."
		app.Version = build.Version.String()
		return nil
	})
}
