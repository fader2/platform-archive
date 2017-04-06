// Copyright (c) Fader, IP. All Rights Reserved.
// See LICENSE for license information.

package main

import (
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := &cli.App{
		Commands: []cli.Command{
			runWeb,
			importCmd,
			exportCmd,
		},
	}
	app.Run(os.Args)
}
