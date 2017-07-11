package core

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/CloudyKit/jet"
	"github.com/fader2/platform/config"
	"github.com/fader2/platform/consts"
	"github.com/fader2/platform/objects"
	"github.com/fader2/platform/utils"
	lua "github.com/yuin/gopher-lua"
)

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

	SessionUser *objects.User // session user

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

func registerContextType(L *lua.LState, c *Context) {
	mt := L.NewTypeMetatable(luaCtxTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), luaCtxMethods))

	L.SetGlobal("ctx", newContextLuaValue(c, L))
	return
}

func newContextLuaValue(ctx *Context, L *lua.LState) lua.LValue {
	ud := L.NewUserData()
	ud.Value = ctx
	L.SetMetatable(ud, L.GetTypeMetatable(luaCtxTypeName))
	return ud
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
	"set": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		k := L.CheckString(2)
		lv := L.CheckAny(3)
		v := utils.ToValueFromLValue(lv)
		ctx.Vars.Set(k, v)
		return 0
	},
	"get": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		k := L.CheckString(2)
		v := ctx.Vars[k]
		lv := utils.ToLValueOrNil(v, L)
		L.Push(lv)
		return 1
	},
	"session": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		if L.GetTop() == 2 {
			u := luaCheckUser(L, 2)
			ctx.SessionUser = u.User
			return 0
		}

		if ctx.SessionUser == nil {
			L.Push(lua.LNil)
			return 1
		}

		ud := L.NewUserData()
		ud.Value = &user{ctx.SessionUser}
		L.SetMetatable(ud, L.GetTypeMetatable(luaUserTypeName))
		L.Push(ud)
		return 1
	},

	"queryParam": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		key := L.CheckString(2)
		L.Push(lua.LString(ctx.r.URL.Query().Get(key)))
		return 1
	},
	"formValue": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		key := L.CheckString(2)
		L.Push(lua.LString(ctx.r.FormValue(key)))
		return 1
	},
	"formFile": func(L *lua.LState) int {
		log.Println("FormFile not implemented")
		return 0
	},
	"path": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		L.Push(lua.LString(ctx.r.URL.Path))
		return 1
	},
	"cookieValue": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		k := L.CheckString(2)
		ck, err := ctx.r.Cookie(k)
		if err == http.ErrNoCookie {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(lua.LString(ck.Value))
		return 1
	},
	"dumpVars": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		log.Println("========================")
		log.Println("DUMP VARS FROM CONTEXT")
		for k, v := range ctx.Vars {
			log.Printf("\t%q: \t\t%+v\n", k, v)
		}
		log.Println("========================")
		return 0
	},
	"setCookie": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		k := L.CheckString(2)
		v := L.CheckString(3)

		delta, _ := time.ParseDuration(config.AppConfig.CfgString(consts.COOKIE_EXPIRES))
		exp := time.Now().Add(delta)
		http.SetCookie(ctx.w, &http.Cookie{
			Name:     k,
			Value:    v,
			Path:     "/",
			Domain:   config.AppConfig.CfgString(consts.DOMAIN),
			Expires:  exp,
			Secure:   config.AppConfig.CfgBool(consts.COOKIE_SECURE),
			HttpOnly: true,
		})
		return 0
	},
	"delCookie": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		k := L.CheckString(2) // cookie name

		http.SetCookie(ctx.w, &http.Cookie{
			Name:     k,
			Value:    "",
			Path:     "/",
			Domain:   config.AppConfig.CfgString(consts.DOMAIN),
			Expires:  time.Unix(0, 0),
			Secure:   config.AppConfig.CfgBool(consts.COOKIE_SECURE),
			HttpOnly: true,
		})
		return 0
	},
	"status": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		ctx.ResponseStatus = L.CheckInt(2)
		return 0
	},
	"noContent": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		ctx.Abort = true
		ctx.AbortFromMiddleware = true

		ctx.w.WriteHeader(L.CheckInt(2))
		return 0
	},
	"redirect": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		ctx.Abort = true
		ctx.AbortFromMiddleware = true

		ctx.w.Header().Set("Location", L.CheckString(2))
		ctx.w.WriteHeader(http.StatusFound)
		return 0
	},
	"json": func(L *lua.LState) int {
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
	"blob": func(L *lua.LState) int {
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
				"not supported data type %T, want []byte",
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

	"isPost": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		L.Push(lua.LBool(ctx.r.Method == http.MethodPost))
		return 1
	},
	"isGet": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		L.Push(lua.LBool(ctx.r.Method == http.MethodGet))
		return 1
	},
	"isDelete": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		L.Push(lua.LBool(ctx.r.Method == http.MethodDelete))
		return 1
	},
}
