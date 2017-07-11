package boltdb

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sync"
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
	"github.com/fader2/platform/utils"
	lua "github.com/yuin/gopher-lua"
)

const NAME = "boltdb"

var Instance *Addon

func init() {
	Instance = &Addon{
		Databases: make(map[string]*bolt.DB),
	}
	addons.Register(Instance)

	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})
}

type Addon struct {
	workspace string

	Databases      map[string]*bolt.DB
	DatabasesMutex sync.Mutex
}

func (a *Addon) Name() string {
	return NAME
}

func (a *Addon) Bootstrap(cfg *config.Config, tpls *jet.Set) (err error) {
	a.workspace = cfg.Workspace

	// templates
	tpls.AddGlobalFunc("fragment", func(args jet.Arguments) reflect.Value {
		args.RequireNumOfArguments("fragment", 1, 1)
		if args.Get(0).Interface() == nil {
			return reflect.ValueOf("")
		}
		slug, ok := args.Get(0).Interface().(string)
		if !ok {
			args.Panicf("not expected type %T", slug)
			return reflect.ValueOf("")
		}
		Instance.DatabasesMutex.Lock()
		db, exists := Instance.Databases[consts.TPL_FRAGMENTS_BUCKET_NAME]
		Instance.DatabasesMutex.Unlock()
		if !exists {
			args.Panicf("not init database name %s", consts.TPL_FRAGMENTS_BUCKET_NAME)
			return reflect.ValueOf(nil)
		}

		s := NewBlobStorage(db, consts.TPL_FRAGMENTS_BUCKET_NAME)
		blob, err := objects.GetBlob(s, objects.UUIDFromString(slug))
		if err != nil {
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
	"init": func(L *lua.LState) int {
		log.Println("lua: " + NAME + " Init")
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
	"opens": func(L *lua.LState) int {
		log.Println("lua: " + NAME + " Opens")

		lv := L.CheckAny(1)
		if lv.Type() != lua.LTTable {
			return 0
		}
		arr := utils.ToValueFromLValue(lv).([]interface{})
		log.Printf("setup %d boltdb databases\n", len(arr))

		for _, name := range arr {
			dbpath := filepath.Join(
				config.AppConfig.Workspace,
				fmt.Sprintf("%s.dat", name),
			)
			log.Printf("setup boltdb databases %s\n", dbpath)
			db, err := bolt.Open(
				dbpath,
				0600,
				&bolt.Options{
					Timeout: 1 * time.Second,
				},
			)
			if err != nil {
				L.RaiseError("setup database %s: %s", name, err)
			}

			db.Update(func(tx *bolt.Tx) error {
				tx.CreateBucketIfNotExists([]byte(name.(string)))
				return nil
			})

			Instance.DatabasesMutex.Lock()
			Instance.Databases[name.(string)] = db
			Instance.DatabasesMutex.Unlock()
		}
		return 0
	},
	"bucket": func(L *lua.LState) int {
		name := L.CheckString(1)
		Instance.DatabasesMutex.Lock()
		db, exists := Instance.Databases[name]
		Instance.DatabasesMutex.Unlock()
		if !exists {
			L.RaiseError("not init database name %s", name)
			return 0
		}
		s := NewBlobStorage(db, name)
		return newUserData(s)(L)
	},
}
