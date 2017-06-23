package boltdb

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/CloudyKit/jet"
	"github.com/CloudyKit/jet/loaders/httpfs"
	"github.com/boltdb/bolt"
	"github.com/fader2/platform/addons"
	"github.com/fader2/platform/addons/boltdb/assets/templates"
	"github.com/fader2/platform/config"
	lua "github.com/yuin/gopher-lua"
)

const NAME = "boltdb"

var addon *Addon

func init() {
	addon = &Addon{}
	addons.Register(addon)
}

type Addon struct {
	db *bolt.DB
}

func (a *Addon) Name() string {
	return NAME
}

func (a *Addon) Bootstrap(cfg *config.Config) (err error) {
	dbpath := filepath.Join(cfg.Workspace, "_boltdb.db")
	a.db, err = bolt.Open(
		dbpath,
		0600,
		&bolt.Options{
			Timeout: 1 * time.Second,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (a *Addon) LuaModule() lua.LGFunction {
	return func(L *lua.LState) int {
		mod := L.SetFuncs(L.NewTable(), exports)
		L.SetField(mod, "name", lua.LString(a.Name()))

		setupCustomTypes(L)

		L.Push(mod)
		return 1
	}
}

func (a *Addon) AssetsLoader() jet.Loader {
	return httpfs.NewLoader(templates.Assets)
}

var exports = map[string]lua.LGFunction{
	"Init": func(L *lua.LState) int {
		f, err := templates.Assets.Open(
			"addons." + NAME + "___bootstrap.lua",
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
	"Bucket": func(L *lua.LState) int {
		name := L.CheckString(1)
		s := NewBlobStorage(addon.db, "__buckets:"+name)
		return newLuaRoute(s)(L)
	},
}
