package cache

import (
	"log"

	"github.com/CloudyKit/jet"
	"github.com/CloudyKit/jet/loaders/httpfs"
	"github.com/boltdb/bolt"
	"github.com/fader2/platform/addons"
	"github.com/fader2/platform/addons/cache/assets/templates"
	lua "github.com/yuin/gopher-lua"
)

func init() {
	db, err := bolt.Open("_cache.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	addons.Register(&Addon{
		db,
	})
}

type Addon struct {
	db *bolt.DB
}

func (a *Addon) Name() string {
	return "cache"
}

func (a *Addon) LuaModule() lua.LGFunction {
	return func(L *lua.LState) int {
		mod := L.SetFuncs(L.NewTable(), exports)
		L.SetField(mod, "name", lua.LString(a.Name()))

		L.Push(mod)
		return 1
	}
}

func (a *Addon) AssetsLoader() jet.Loader {
	return httpfs.NewLoader(templates.Assets)
}

var exports = map[string]lua.LGFunction{}
