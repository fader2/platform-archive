package foo

import (
	"bytes"
	"io"
	"os"

	"github.com/CloudyKit/jet"
	"github.com/CloudyKit/jet/loaders/httpfs"
	"github.com/fader2/platform/addons"
	"github.com/fader2/platform/addons/foo/assets/templates"
	"github.com/fader2/platform/config"
	lua "github.com/yuin/gopher-lua"
)

const NAME = "foo"

func init() {
	addons.Register(&Addon{})
}

type Addon struct {
}

func (a *Addon) Name() string {
	return NAME
}

func (a *Addon) Bootstrap(cfg *config.Config, tpls *jet.Set) error {
	// TODO: bootstrap
	return nil
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

var exports = map[string]lua.LGFunction{
	"init": func(L *lua.LState) int {
		f, err := templates.Assets.Open(
			NAME + "/bootstrap.lua",
		)
		if os.IsNotExist(err) {
			return 0
		}
		if err != nil {
			L.RaiseError("bootstrap %s: %s", NAME, err)
			return 0
		}
		defer f.Close()

		bootstrap := new(bytes.Buffer)
		io.Copy(bootstrap, f)
		if err := L.DoString(bootstrap.String()); err != nil {
			L.RaiseError("bootstrap %s: load cfg: %s", NAME, err)
		}
		return 0
	},
}
