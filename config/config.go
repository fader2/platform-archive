package config

import (
	"crypto/rsa"
	"fmt"
	"log"
	"sync/atomic"

	"sync"

	"github.com/CloudyKit/jet"
	"github.com/fader2/platform/utils"
	lua "github.com/yuin/gopher-lua"
)

var (
	maintenance int32

	AppConfig *Config
)

type Config struct {
	Dev bool

	Workspace string
	AppLua    string
	AppPath   string
	Version   string

	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey

	Vars      jet.VarMap
	VarsMutex sync.RWMutex

	Routs         []Route
	NotFoundPage  Route
	ForbiddenPage Route
}

type Route struct {
	Method      string   // HTTP method
	Path        string   // URL path
	Handler     string   // path to file fo handler
	Middlewares []string // path to lua scripts
	Roles       []string // permissible roles to the resource
}

func New() *Config {
	return &Config{
		Vars: make(jet.VarMap),
	}
}

// Utils

func SetMaintenance(v bool) {
	if v {
		atomic.StoreInt32(&maintenance, 1)
	} else {
		atomic.StoreInt32(&maintenance, 0)
	}
}

func IsMaintenance() bool {
	return atomic.LoadInt32(&maintenance) == 1
}

////////////////////////////////////////////////////////////////////////////////
// Lua implement
////////////////////////////////////////////////////////////////////////////////

const luaCfgTypeName = "cfg"

func LuaSetCfg(L *lua.LState, c *Config) {
	L.SetGlobal("cfg", L.NewFunction(LuaCfgFromCfg(c)))
	mt := L.NewTypeMetatable(luaCfgTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), luaCfgMethods))
	return
}

// LuaSetReadOnlyCfg set a global new type object. Config read only
func LuaSetReadOnlyCfg(L *lua.LState, c *Config) {
	L.SetGlobal("cfg", L.NewFunction(LuaCfgFromCfg(c)))
	mt := L.NewTypeMetatable(luaCfgTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), luaCfgReadOnlyMethods))
	return
}

func LuaCfgFromCfg(cfg *Config) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		ud := L.NewUserData()
		ud.Value = cfg

		L.SetMetatable(ud, L.GetTypeMetatable(luaCfgTypeName))
		L.Push(ud)
		return 1
	}
}

// LuaCheckCfg returns *Config if it is first argument
// helpful function, used in Lua modules
func LuaCheckCfg(L *lua.LState) *Config {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*Config); ok {
		return v
	}
	reason := fmt.Sprintf("expected *Config, got %T", ud.Value)
	L.ArgError(1, reason)
	return nil
}

var luaCfgReadOnlyMethods = map[string]lua.LGFunction{
	"Dev": func(L *lua.LState) int {
		cfg := LuaCheckCfg(L)
		L.Push(lua.LBool(cfg.Dev))
		return 1
	},
	"Get": luaCfg_GetVar,
	"Workspace": func(L *lua.LState) int {
		cfg := LuaCheckCfg(L)
		L.Push(lua.LString(cfg.Workspace))
		return 1
	},
	"DumpVars": func(L *lua.LState) int {
		cfg := LuaCheckCfg(L)
		log.Println("========================")
		log.Println("DUMP VARS FROM CONFIG")
		for k, v := range cfg.Vars {
			log.Printf("\t%q: \t\t%+v\n", k, v)
		}
		log.Println("========================")
		return 0
	},
}

var luaCfgMethods = map[string]lua.LGFunction{
	"Get": luaCfg_GetVar,
	"Set": luaCfg_SetVar,
	/*
		AddRoute добавить роут
		- method
		- path
		- handler
		- middlewares
	*/
	"AddRoute": func(L *lua.LState) int {
		cfg := LuaCheckCfg(L)

		cfg.Routs = append(cfg.Routs, luaCfg_RouteFromLuaFn(L))

		return 0
	},
	"Dev": func(L *lua.LState) int {
		cfg := LuaCheckCfg(L)
		cfg.Dev = L.CheckBool(2)
		return 0
	},
	"NotFoundPage": func(L *lua.LState) int {
		cfg := LuaCheckCfg(L)

		cfg.NotFoundPage = luaCfg_RouteFromLuaFn(L)

		return 0
	},
	"ForbiddenPage": func(L *lua.LState) int {
		cfg := LuaCheckCfg(L)

		cfg.ForbiddenPage = luaCfg_RouteFromLuaFn(L)

		return 0
	},
}

// Returns a route based on arguments (Internal use in adding a new routing)
func luaCfg_RouteFromLuaFn(L *lua.LState) (route Route) {
	if L.GetTop() < 4 {
		L.ArgError(1, "unexpected number of arguments")
		return
	}
	route = Route{
		Method:  L.CheckString(2),
		Path:    L.CheckString(3),
		Handler: L.CheckString(4),
	}

	if L.GetTop() == 4 {
		return
	}

	////////////////////////////////////////////////////////
	// Middlewares
	////////////////////////////////////////////////////////

	anylv := L.CheckAny(5) // middlewares array
	anyi := utils.ToValueFromLValue(anylv)
	if arr, ok := anyi.([]interface{}); ok {
		for _, item := range arr {
			if midl, ok := item.(string); ok {
				route.Middlewares = append(route.Middlewares, midl)
			} else {
				log.Printf("unexpected type middleware %T (wang string)\n", item)
			}
		}
	} else if midl, ok := anyi.(string); ok {
		route.Middlewares = append(route.Middlewares, midl)
	} else {
		log.Printf("unexpected type array of middlewares %T \n", anyi)
	}

	if L.GetTop() == 5 {
		return
	}

	////////////////////////////////////////////////////////
	// Roles
	////////////////////////////////////////////////////////

	anylv = L.CheckAny(6)
	anyi = utils.ToValueFromLValue(anylv)
	if arr, ok := anyi.([]interface{}); ok {
		for _, item := range arr {
			if role, ok := item.(string); ok {
				route.Roles = append(route.Roles, role)
			} else {
				log.Printf("unexpected type middleware %T (wang string)\n", item)
			}
		}
	} else if role, ok := anyi.(string); ok {
		route.Middlewares = append(route.Middlewares, role)
	} else {
		log.Printf("unexpected type array of middlewares %T \n", anyi)
	}
	return
}

func luaCfg_GetVar(L *lua.LState) int {
	cfg := LuaCheckCfg(L)
	cfg.VarsMutex.RLock()
	defer cfg.VarsMutex.RUnlock()

	k := L.CheckString(2)
	v := cfg.Vars[k]
	lv := utils.ToLValueOrNil(v, L)
	if lv == nil {
		log.Printf("cfg.Get(): not supported type, got %T, key %s", v, k)
		return 0
	}
	L.Push(lv)
	return 1
}

func luaCfg_SetVar(L *lua.LState) int {
	cfg := LuaCheckCfg(L)
	cfg.VarsMutex.Lock()
	defer cfg.VarsMutex.Unlock()

	k := L.CheckString(2)
	lv := L.CheckAny(3)
	v := utils.ToValueFromLValue(lv)
	if v == nil {
		log.Printf("cfg.Set(): not supported type, got %T, key %s", lv, k)
		return 0
	}
	cfg.Vars.Set(k, v)
	return 1
}
