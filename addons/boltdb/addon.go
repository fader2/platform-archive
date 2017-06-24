package boltdb

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"encoding/gob"

	"github.com/CloudyKit/jet"
	"github.com/CloudyKit/jet/loaders/httpfs"
	"github.com/boltdb/bolt"
	"github.com/fader2/platform/addons"
	"github.com/fader2/platform/addons/boltdb/assets/templates"
	"github.com/fader2/platform/config"
	"github.com/fader2/platform/consts"
	"github.com/fader2/platform/objects"
	lua "github.com/yuin/gopher-lua"
)

const NAME = "boltdb"

var Instance *Addon

func init() {
	Instance = &Addon{}
	addons.Register(Instance)

	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})
}

type Addon struct {
	DB *bolt.DB
}

func (a *Addon) Name() string {
	return NAME
}

func (a *Addon) Bootstrap(cfg *config.Config, tpls *jet.Set) (err error) {
	dbpath := filepath.Join(cfg.Workspace, "_boltdb.db")
	a.DB, err = bolt.Open(
		dbpath,
		0600,
		&bolt.Options{
			Timeout: 1 * time.Second,
		},
	)
	if err != nil {
		return err
	}

	// templates
	tpls.AddGlobalFunc("get", func(args jet.Arguments) reflect.Value {
		args.RequireNumOfArguments("get", 1, 1)
		if args.Get(0).Interface() == nil {
			return reflect.ValueOf("")
		}
		slug, ok := args.Get(0).Interface().(string)
		if !ok {
			args.Panicf("not expected type %T", slug)
			return reflect.ValueOf("")
		}
		s := NewBlobStorage(a.DB, consts.TPL_WIDGETS_BUCKET_NAME)
		blob, err := objects.GetBlob(s, objects.UUIDFromString(slug))
		if err != nil {
			args.Panicf("find widget by name %q: %s", slug, err)
			return reflect.ValueOf("")
		}
		return reflect.ValueOf(string(blob.Data))
	})

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
		switch name {
		case consts.TPL_WIDGETS_BUCKET_NAME:
		// TODO: более лучший механихм работы с
		default:
			name = "__buckets:" + name
		}
		s := NewBlobStorage(Instance.DB, name)
		return newLuaRoute(s)(L)
	},
}
