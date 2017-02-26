// Copyright (c) Fader, IP. All Rights Reserved.
// See LICENSE for license information.

package main

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
