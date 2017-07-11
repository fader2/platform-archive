package tpls

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/CloudyKit/jet"
	"github.com/CloudyKit/jet/loaders/httpfs"
	"github.com/fader2/platform/addons"
	"github.com/fader2/platform/addons/tpls/assets/templates"
	"github.com/fader2/platform/config"
	"github.com/microcosm-cc/bluemonday"
	lua "github.com/yuin/gopher-lua"

	"github.com/fader2/platform/objects"
	"github.com/russross/blackfriday"
)

const NAME = "tpls"

var Instance *Addon

func init() {
	Instance = &Addon{}
	addons.Register(Instance)
}

type Addon struct {
}

func (a *Addon) Name() string {
	return NAME
}

func (a *Addon) Bootstrap(cfg *config.Config, tpls *jet.Set) error {
	// TODO: bootstrap
	tpls.AddGlobalFunc("base64", func(a jet.Arguments) reflect.Value {
		a.RequireNumOfArguments("base64", 1, 1)

		buffer := bytes.NewBuffer(nil)
		fmt.Fprint(buffer, a.Get(0))

		return reflect.ValueOf(base64.URLEncoding.EncodeToString(buffer.Bytes()))
	})
	tpls.AddGlobalFunc("md", func(a jet.Arguments) reflect.Value {
		a.RequireNumOfArguments("md", 1, 1)

		buffer := bytes.NewBuffer(nil)
		fmt.Fprint(buffer, a.Get(0))

		unsafe := blackfriday.MarkdownCommon(buffer.Bytes())
		html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
		return reflect.ValueOf(string(html))
	})
	tpls.AddGlobalFunc("uuidFrom", func(a jet.Arguments) reflect.Value {
		a.RequireNumOfArguments("uuidFrom", 1, 1)
		return reflect.ValueOf(objects.UUIDFromString(a.Get(0).String()).String())
	})
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
