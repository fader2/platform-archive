package core

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/fader2/platform/config"
	"github.com/fader2/platform/utils"

	"log"

	"encoding/json"

	"io"

	"github.com/CloudyKit/jet"
	"github.com/fader2/platform/addons"
	"github.com/julienschmidt/httprouter"
	lua "github.com/yuin/gopher-lua"
)

var (
	DefNotFoundTplHandler = func(
		w http.ResponseWriter,
		r *http.Request,
		ps httprouter.Params,
	) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Not found template"))
	}
)

type AssetsLoader interface {
	Open(name string) (io.ReadCloser, error)
	Exists(name string) (string, bool)
}

func EntrypointHandler(
	assets AssetsLoader,
	cfg *config.Config,
	route config.Route,
	tpls *jet.Set,
) func(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
) {
	return func(
		w http.ResponseWriter,
		r *http.Request,
		ps httprouter.Params,
	) {
		if config.IsMaintenance() {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Header().Set("Retry-After", "120")
			w.Write([]byte("Maintenance. Please retry after 120 sec."))
			return
		}

		/*
			- find template
			- setup context and lua executor
				- setup request params
			- execute lua
			- execute tpl
		*/

		// Find tpl
		view, err := tpls.GetTemplate(route.Handler)
		if err != nil {
			// not found template
			DefNotFoundTplHandler(w, r, ps)
			return
		}

		// setup ctx and lua engine
		L := lua.NewState()
		defer L.Close()

		addons.PreloadLuaModules(L)

		vars := make(jet.VarMap)
		ctx := NewContext(
			&route,
			w,
			r,
			vars,
		)
		luaSetNewCtx(L, ctx)
		config.LuaSetReadOnlyCfg(L, cfg)

		// set request options
		for _, param := range ps {
			vars.Set(param.Key, param.Value)
		}

		// execute all middlewares
		for _, middleware := range route.Middlewares {
			fullPathToMiddleware, exists := assets.Exists(middleware)
			if !exists {
				log.Printf(
					"error open middleware %s, not exists",
					middleware,
				)
				continue
			}
			f, err := assets.Open(fullPathToMiddleware)
			if err != nil {
				log.Printf(
					"error open middleware %s, %s",
					middleware,
					err,
				)
				continue
			}
			defer f.Close()

			d := new(bytes.Buffer)
			io.Copy(d, f)
			if err := L.DoString(d.String()); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal server"))
				log.Printf(
					"error execute middleware %s, %s",
					middleware,
					err,
				)
				return
			}

			if ctx.AbortFromMiddleware {
				break
			}
		}

		if ctx.Abort {
			return
		}

		if ctx.ResponseStatus == -1 {
			ctx.ResponseStatus = http.StatusOK
		}
		w.WriteHeader(ctx.ResponseStatus)
		view.Execute(w, vars, ctx)
	}
}

func NewContext(
	route *config.Route,
	w http.ResponseWriter,
	r *http.Request,
	vars jet.VarMap,
) *Context {
	return &Context{
		Vars:           vars,
		Route:          route,
		w:              w,
		r:              r,
		ResponseStatus: -1,
	}
}

type Context struct {
	Vars jet.VarMap

	w http.ResponseWriter
	r *http.Request

	Route          *config.Route
	Err            error // in lua script was error
	ResponseStatus int   // in lua script set response status

	Abort               bool // in lua script was executed render
	AbortFromMiddleware bool // in lua script set was abort
}

func (c *Context) Blob(code int, contentType string, b []byte) (err error) {
	c.w.Header().Set("Content-Type", contentType)
	c.w.WriteHeader(code)
	_, err = c.w.Write(b)
	return
}

////////////////////////////////////////////////////////////////////////////////
// Lua implement
////////////////////////////////////////////////////////////////////////////////

const luaCtxTypeName = "ctx"

func luaSetNewCtx(L *lua.LState, c *Context) {
	L.SetGlobal("ctx", L.NewFunction(luaCtxBuilder(c)))
	mt := L.NewTypeMetatable(luaCtxTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), luaCtxMethods))
	return
}

func luaCtxBuilder(ctx *Context) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		ud := L.NewUserData()
		ud.Value = ctx

		L.SetMetatable(ud, L.GetTypeMetatable(luaCtxTypeName))
		L.Push(ud)
		return 1
	}
}

func luaCheckCtx(L *lua.LState) *Context {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*Context); ok {
		return v
	}
	reason := fmt.Sprintf("expected Context, got %T", ud.Value)
	L.ArgError(1, reason)
	return nil
}

var luaCtxMethods = map[string]lua.LGFunction{
	"Set": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		k := L.CheckString(2)
		lv := L.CheckAny(3)
		v := utils.ToValueFromLValue(lv)
		if v == nil {
			log.Printf("ctx.Set(): not supported type, got %T, key %s", lv, k)
			return 0
		}
		ctx.Vars.Set(k, v)
		return 0
	},
	"Get": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		k := L.CheckString(2)
		v := ctx.Vars[k]
		lv := utils.ToLValueOrNil(v, L)
		if lv == nil {
			log.Printf("ctx.Get(): not supported type, got %T, key %s", v, k)
			return 0
		}
		L.Push(lv)
		return 1
	},
	"Status": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		ctx.ResponseStatus = L.CheckInt(2)
		return 0
	},
	"NoContent": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		ctx.Abort = true
		ctx.AbortFromMiddleware = true

		ctx.w.WriteHeader(L.CheckInt(2))
		return 0
	},
	"Redirect": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		ctx.Abort = true
		ctx.AbortFromMiddleware = true

		ctx.w.Header().Set("Location", L.CheckString(2))
		ctx.w.WriteHeader(http.StatusFound)
		return 0
	},
	"JSON": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		ctx.Abort = true
		ctx.AbortFromMiddleware = true

		if L.GetTop() < 3 {
			L.RaiseError("ctx.JSON(): unexpected number of arguments, got %d, want 3", L.GetTop())
			return 0
		}

		status := L.CheckInt(2)
		v := utils.ToValueFromLValue(L.CheckAny(3))

		data, _ := json.Marshal(v)
		ctx.Blob(
			status, // code
			"application/json;charset=utf-8",
			data,
		)

		return 0
	},
	"Blob": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		ctx.Abort = true
		ctx.AbortFromMiddleware = true

		if L.GetTop() < 4 {
			L.RaiseError("ctx.Blob(): unexpected number of arguments, got %d, want 4", L.GetTop())
			return 0
		}

		status := L.CheckInt(2)
		contentType := L.CheckString(3)
		ud := L.CheckUserData(4)
		data, ok := ud.Value.([]byte)
		if !ok {
			L.RaiseError(
				"not supported data type %T, got []byte",
				ud.Value,
			)
			return 0
		}

		ctx.Blob(
			status, // code
			contentType,
			data,
		)

		return 0
	},
}
