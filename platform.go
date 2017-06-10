package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"log"

	"encoding/json"

	"github.com/CloudyKit/jet"
	"github.com/fader2/platform/core"
	"github.com/julienschmidt/httprouter"
	billy "gopkg.in/src-d/go-billy.v2"
	"gopkg.in/src-d/go-billy.v2/osfs"
)

var version = ""

const (
	faderLuaFileName = "fader.lua"
)

var workspace = flag.String("workspace", "_workspace", "Path to work directory")
var port = flag.Int("port", 8383, "Port listening for the frontend")

var (
	cfg    *core.Config
	tpls   *jet.Set
	routes *httprouter.Router
	fs     billy.Filesystem
)

func main() {
	flag.Parse()

	tpls = jet.NewHTMLSet(*workspace)
	fs = osfs.New(*workspace).Dir("")
	routes = httprouter.New()
	loadSetting()
	showCfg()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: routes,
	}
	server.ListenAndServe()

	////////////////////////////////////////////////////////////////////////////
	// graceful exit
	////////////////////////////////////////////////////////////////////////////

	done := make(chan struct{}, 1)
	signals := make(chan os.Signal, 3)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGUSR1)
	go func() {
		for sig := range signals {

			switch sig {
			case syscall.SIGCONT:
				log.Printf("captured %v, processing...", sig)
				continueFrontend()
			case syscall.SIGSTOP:
				log.Printf("captured %v, processing...", sig)
				stopFrontend()
			case syscall.SIGUSR1:
				log.Printf("captured %v, processing...", sig)
				stopFrontend()
				loadSetting()
				showCfg()
				continueFrontend()
			default:
				log.Printf("captured %v, exiting...", sig)
				destroy()
				close(done)
			}
		}
	}()

	// but not yet graceful

	<-done
	log.Println("bye bye")
	os.Exit(1)
}

func destroy() {
	log.Println("destroy")

	// TODO: destroy
}

func loadSetting() {
	log.Println("load settings")
	var err error
	fader, err := fs.Open(faderLuaFileName)
	if err != nil {
		log.Fatalf("open %q: %s", faderLuaFileName, err)
	}
	faderSettings := new(bytes.Buffer)
	io.Copy(faderSettings, fader)
	cfg, err = core.LoadConfigFromLua(faderSettings.Bytes())
	if err != nil {
		log.Fatal("init fader settings (from lua):", err)
	}

	tpls.SetDevelopmentMode(cfg.Dev)
	cfg.Workspace = *workspace

	for _, route := range cfg.Routs {
		routes.Handle(
			strings.ToUpper(route.Method),
			route.Path,
			core.EntrypointHandler(
				cfg,
				route,
				tpls,
			),
		)
		log.Printf(
			"[ROUTE] '%s %s' (for %v) to '%s' (midls %v)\n",
			strings.ToUpper(route.Method),
			route.Path,
			route.Roles,
			route.Handler,
			route.Middlewares,
		)
	}
}

func continueFrontend() {
	log.Println("run frontend server")

	core.SetMaintenance(false)
}

func stopFrontend() {
	log.Println("stop frontend server")
	core.SetMaintenance(true)
}

func showCfg() {
	log.Println("==================================")
	log.Println("options:")
	log.Println("\tfrontend port:", *port)
	log.Println("\tworkspace:", *workspace)
	log.Println("")
	log.Println("settings (DUMP):")
	cfgJSON, _ := json.MarshalIndent(cfg, "", "    ")
	log.Println(string(cfgJSON))
	log.Println("==================================")
}
