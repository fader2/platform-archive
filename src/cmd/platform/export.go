package main

import (
	"github.com/labstack/echo"
	"github.com/urfave/cli"
	"log"

	// todo rename to absolute pkg name
	"api"
)

var exportCmd = cli.Command{
	Name:   "export",
	Action: runExport,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "output",
		},
	},
}

func runExport(ctx *cli.Context) {
	e := echo.New()
	setup(e, ctx)

	err := api.ExportWorkspace(ctx.String("output"))
	if err != nil {
		log.Fatal(err)
	}
}
