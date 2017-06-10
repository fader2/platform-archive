package bar

import (
	"github.com/CloudyKit/jet"
	"github.com/CloudyKit/jet/loaders/httpfs"
	"github.com/fader2/platform/addons"
	"github.com/fader2/platform/addons/bar/assets/templates"
	lua "github.com/yuin/gopher-lua"
)

func init() {
	addons.Register(&Addon{})
}

type Addon struct {
}

func (a *Addon) Name() string {
	return "bar"
}

func (a *Addon) LuaModule() lua.LGFunction {
	return func(L *lua.LState) int {
		// register functions to the table
		mod := L.SetFuncs(L.NewTable(), exports)
		// register other stuff
		L.SetField(mod, "name", lua.LString(a.Name()))

		// returns the module
		L.Push(mod)
		return 1
	}
}

func (a *Addon) AssetsLoader() jet.Loader {
	return httpfs.NewLoader(templates.Assets)
}

var exports = map[string]lua.LGFunction{}
