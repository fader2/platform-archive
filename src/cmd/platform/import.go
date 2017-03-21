package main

import (
	"github.com/labstack/echo"
	"github.com/urfave/cli"
	"log"

	// todo rename to absolute pkg name
	"api"
)

var importCmd = cli.Command{
	Name:   "import",
	Action: runImport,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "input",
		},
	},
}

func runImport(ctx *cli.Context) {
	e := echo.New()
	setup(e, ctx)

	err := api.ImportWorkspace(ctx.String("input"))
	if err != nil {
		log.Fatal(err)
	}
}
