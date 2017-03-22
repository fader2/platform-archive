package main

import (
	"github.com/labstack/echo"
	"github.com/urfave/cli"
	"log"

	// todo rename to absolute pkg name
	"api"
)

var legacyImportCmd = cli.Command{
	Name:   "import64",
	Action: legacyImport,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "input",
		},
	},
}

func legacyImport(ctx *cli.Context) {
	e := echo.New()
	setup(e, ctx)

	err := api.ImportBase64File(ctx.String("input"))
	if err != nil {
		log.Fatal(err)
	}
}
