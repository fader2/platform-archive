// Copyright (c) Fader, IP. All Rights Reserved.
// See LICENSE for license information.

/*package main

import (
	"context"
	"flag"
	"fs"
	"log"

	"os"

	"github.com/SentimensRG/sigctx"
	"github.com/fsnotify/fsnotify"
)

var (
	flagWatcher = flag.Bool("watch", false, "run watcher for FADER_WORKSPACE")
	watcher     *fsnotify.Watcher
)

func main() {
	flag.Parse()

	ctx := sigctx.New()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if *flagWatcher {
		w := fs.NewFSWatcher()

		if err := w.Run(ctx, os.Getenv("FADER_WORKSPACE")); err != nil {
			log.Fatal(err)
		}
	}

	<-ctx.Done()
}
*/

package main

import (
	"api"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/tylerb/graceful"
)

var flagPortListener = flag.String("port", "1323", "http listener port")
var flagHostListener = flag.String("host", "", "http listener host or ip")
var flagDatabasePath = flag.String("db", "", "path to file database boltdb")

var falgWatch = flag.Bool("watch", false, "enable file system workspace synchromization")
var flagWorkspace = flag.String("workspace", "FaderWorkspace", "Workspace path. Ignored if watch = false")

var logger = *log.New(os.Stderr, "[main]", 1)

const (
	FADER_HTTPPORT = "FADER_HTTPPORT"
	FADER_HTTPHOST = "FADER_HTTPHOST"
	FADER_DBPATH   = "FADER_DBPATH"
	FADER_LOGLEVEL = "FADER_LOGLEVEL"
	FADER_INITFILE = "FADER_INITFILE"
)

func main() {
	flag.Parse()

	logger.Println("Start http api...")

	e := echo.New()

	settings := settingsFromENV()

	if err := api.Setup(e, settings); err != nil {
		logger.Fatalln("[FATAL] setup,", err)
	}

	// ---------------------------
	// HTTP server
	// ---------------------------

	serverSignal := make(chan struct{}, 1)

	go func() {
		addr := *flagHostListener + ":" + *flagPortListener

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

func settingsFromENV() *api.Settings {
	return &api.Settings{
		ApiHost:      os.Getenv(FADER_HTTPHOST),
		ApiPort:      os.Getenv(FADER_HTTPPORT),
		DatabasePath: os.Getenv(FADER_DBPATH),
		LogLevel:     api.LogLevelFrom(os.Getenv(FADER_LOGLEVEL)),
		InitFile:     os.Getenv(FADER_INITFILE),

		Watch: *falgWatch,
	}
}
