package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"log"

	"encoding/json"

	"github.com/CloudyKit/jet"
	"github.com/CloudyKit/jet/loaders/multi"
	"github.com/fader2/platform/addons"
	"github.com/fader2/platform/config"
	"github.com/fader2/platform/core"
	"github.com/julienschmidt/httprouter"
	lua "github.com/yuin/gopher-lua"
	billy "gopkg.in/src-d/go-billy.v2"
	"gopkg.in/src-d/go-billy.v2/osfs"
)

var version = ""

const (
	appLuaFileName = "app.lua"
	appFolderName  = "app"
)

var workspace = flag.String("workspace", "_workspace", "Path to work directory")
var port = flag.Int("port", 8383, "Port listening for the frontend")

var (
	tpls   *jet.Set
	routes *httprouter.Router
	fs     billy.Filesystem

	frontendServer *http.Server
)

func main() {
	flag.Parse()

	appRootPath := filepath.Join(*workspace, appFolderName)
	appLuaFile := filepath.Join(*workspace, appLuaFileName)

	assets = multi.NewLoader(
		addons.AppendJetLoaders(
			jet.NewOSFileSystemLoader(appRootPath),
		)...,
	)

	tpls = jet.NewHTMLSetLoader(assets)

	fs = osfs.New(*workspace).Dir("")
	loadSetting()
	config.AppConfig.Workspace = *workspace
	config.AppConfig.AppPath = appRootPath
	config.AppConfig.AppLua = appLuaFile
	config.AppConfig.Version = version
	if err := BootstrapAddons(config.AppConfig, tpls); err != nil {
		log.Fatal(err)
	}

	showCfg()
	go continueFrontend()

	////////////////////////////////////////////////////////////////////////////
	// graceful exit
	////////////////////////////////////////////////////////////////////////////

	done := make(chan struct{}, 1)
	signals := make(chan os.Signal, 3)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGUSR1)
	go func() {
		for sig := range signals {

			switch sig {
			case syscall.SIGUSR1:
				log.Printf("captured %v, processing...", sig)
				stopFrontend()
				loadSetting()
				showCfg()
				go continueFrontend()
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

	routes = httprouter.New()
	config.AppConfig = config.New()

	// load app.lua file
	var err error
	fader, err := fs.Open(appLuaFileName)
	if err != nil {
		log.Fatalf("open %q: %s", appLuaFileName, err)
	}
	_data := new(bytes.Buffer)
	io.Copy(_data, fader)

	// execute app.lua file
	L := lua.NewState()
	defer L.Close()
	config.LuaSetCfg(L, config.AppConfig)
	addons.PreloadLuaModules(L)
	if err = L.DoString(_data.String()); err != nil {
		log.Fatal("init settings (from lua):", err)
		return
	}

	tpls.SetDevelopmentMode(config.AppConfig.Dev)

	// setup routes from cfg
	for _, route := range config.AppConfig.Routs {
		routes.Handle(
			strings.ToUpper(route.Method),
			route.Path,
			core.EntrypointHandler(
				assets,
				config.AppConfig,
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
	config.SetMaintenance(false)
	startFrontendServer()
}

func startFrontendServer() {
	frontendServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: routes,
	}
	if err := frontendServer.ListenAndServe(); err != nil {
		log.Println("listen frontend server:", err)
	}
}

func stopFrontendServer() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	frontendServer.Shutdown(ctx)
	log.Println("frontend server gracefully stopped")
}

func stopFrontend() {
	log.Println("stop frontend server")
	config.SetMaintenance(true)
	stopFrontendServer()
}

func showCfg() {
	log.Println("==================================")
	log.Println("options:")
	log.Println("\tfrontend port:", *port)
	log.Println("\tworkspace:", *workspace)
	log.Println("")
	log.Println("settings (DUMP):")
	cfgJSON, _ := json.MarshalIndent(config.AppConfig, "", "    ")
	log.Println(string(cfgJSON))
	log.Println("==================================")
}
