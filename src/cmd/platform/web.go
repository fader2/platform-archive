package main

import (
	"api"

	"github.com/urfave/cli"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/tylerb/graceful"
)

var logger = *log.New(os.Stderr, "[main]", 1)

const (
	FADER_HTTPPORT = "FADER_HTTPPORT"
	FADER_HTTPHOST = "FADER_HTTPHOST"
	FADER_DBPATH   = "FADER_DBPATH"
	FADER_LOGLEVEL = "FADER_LOGLEVEL"
	FADER_INITFILE = "FADER_INITFILE"
)

var (
	runWeb = cli.Command{
		Name:   "web",
		Action: startWeb,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name: "watch",
			},
			cli.StringFlag{
				Name:  "port",
				Value: "1323",
			},
			cli.StringFlag{
				Name: "host",
			},
		},
	}
)

func setup(e *echo.Echo, ctx *cli.Context) {
	settings := settingsFromENV(ctx)

	if err := api.Setup(e, settings); err != nil {
		logger.Fatalln("[FATAL] setup,", err)
	}
}

func startWeb(ctx *cli.Context) {

	logger.Println("Start http api...")

	e := echo.New()
	setup(e, ctx)

	// ---------------------------
	// HTTP server
	// ---------------------------

	serverSignal := make(chan struct{}, 1)

	go func() {
		addr := ctx.String("host") + ":" + ctx.String("port")

		logger.Println("HTTP listener address: ", addr)

		server := standard.New(addr)
		server.SetHandler(e)

		err := graceful.ListenAndServe(server.Server, 5*time.Second)

		logger.Println("[ERR] stop http server", err)

		serverSignal <- struct{}{}
	}()

	// ---------------------------
	// Run listener of OS
	// ---------------------------

	osSignal := make(chan os.Signal, 2)
	close := make(chan struct{})
	signal.Notify(osSignal, os.Interrupt, syscall.SIGTERM)

	go func() {

		defer func() {
			close <- struct{}{}
		}()

		select {
		case <-osSignal:
			logger.Println("[INF] signal completion of the process")
		case <-serverSignal:
			logger.Println("[INF] shutdown http server")
		}

	}()

	<-close
	os.Exit(0)
}

func settingsFromENV(ctx *cli.Context) *api.Settings {
	return &api.Settings{
		ApiHost:      os.Getenv(FADER_HTTPHOST),
		ApiPort:      os.Getenv(FADER_HTTPPORT),
		DatabasePath: os.Getenv(FADER_DBPATH),
		LogLevel:     api.LogLevelFrom(os.Getenv(FADER_LOGLEVEL)),
		InitFile:     os.Getenv(FADER_INITFILE),

		Watch: ctx.Bool("watch"),
	}
}
