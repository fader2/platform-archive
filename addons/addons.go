package addons

import (
	"log"

	"github.com/fader2/platform/config"

	"github.com/CloudyKit/jet"
	lua "github.com/yuin/gopher-lua"
)

var Addons = make(map[string]Addon)

type Addon interface {
	Name() string

	// LuaModule Loads module components
	LuaModule() lua.LGFunction

	// AssetsLoader returns the file loader belonging to the extension
	AssetsLoader() jet.Loader

	// Bootstrap bootstrap addon
	Bootstrap(cfg *config.Config) error
}

func Register(addon Addon) {
	Addons[addon.Name()] = addon
	log.Println("add addon", addon.Name())
}

func PreloadLuaModules(L *lua.LState) {
	for _, addon := range Addons {
		L.PreloadModule(addon.Name(), addon.LuaModule())
	}
}

func Bootstrap(cfg *config.Config) {
	for _, addon := range Addons {
		addon.Bootstrap(cfg)
	}
}

func AppendJetLoaders(loaders ...jet.Loader) []jet.Loader {
	for _, addon := range Addons {
		loaders = append(loaders, addon.AssetsLoader())
	}
	return loaders
}
