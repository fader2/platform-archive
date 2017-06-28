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
	uuid "github.com/satori/go.uuid"
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

func luaAddCustomTypeContextAndHelpers(L *lua.LState, c *Context) {
	L.SetGlobal("ctx", L.NewFunction(luaCtxBuilder(c)))
	L.SetGlobal("GenToken", L.NewFunction(func(L *lua.LState) int {
		id, err := uuid.FromString(L.CheckString(1))
		if err != nil {
			L.RaiseError("invalid ID", err)
			return 0
		}
		token, err := GenerateToken(id)
		if err != nil {
			L.RaiseError("invalid token", err)
			return 0
		}

		L.Push(lua.LString(token))
		return 1
	}))
	L.SetGlobal("Auth", L.NewFunction(
		func(L *lua.LState) int {
			ls := L.CheckUserData(1)
			s, ok := ls.Value.(objects.Storer)
			if !ok {
				L.RaiseError("expected objects.Storer, got %T", ls.Value)
				return 0
			}
			token := L.CheckString(2)
			usr, err := Authenticate(token, s)

			if err == consts.ErrNotFound {
				return luaUserBuilder(&luaUser{
					User:   objects.EmptyUser(objects.UnknownUserType),
					Exists: false,
				})(L)
			}
			if err != nil {
				L.RaiseError("auth user by token", err)
				return 0
			}

			return luaUserBuilder(&luaUser{
				User:   usr,
				Exists: true,
			})(L)
		},
	))

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
		ctx.Vars.Set(k, v)
		return 0
	},
	"Get": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		k := L.CheckString(2)
		v := ctx.Vars[k]
		lv := utils.ToLValueOrNil(v, L)
		L.Push(lv)
		return 1
	},
	"QueryParam": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		key := L.CheckString(2)
		L.Push(lua.LString(ctx.r.URL.Query().Get(key)))
		return 1
	},
	"FormValue": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		key := L.CheckString(2)
		L.Push(lua.LString(ctx.r.FormValue(key)))
		return 1
	},
	"FormFile": func(L *lua.LState) int {
		log.Println("FormFile not implemented")
		return 0
	},
	"Path": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		L.Push(lua.LString(ctx.r.URL.Path))
		return 1
	},
	"CookieValue": func(L *lua.LState) int {
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
	"DumpVars": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		log.Println("========================")
		log.Println("DUMP VARS FROM CONTEXT")
		for k, v := range ctx.Vars {
			log.Printf("\t%q: \t\t%+v\n", k, v)
		}
		log.Println("========================")
		return 0
	},
	"SetCookie": func(L *lua.LState) int {
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
	"DelCookie": func(L *lua.LState) int {
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

	"IsPost": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		L.Push(lua.LBool(ctx.r.Method == http.MethodPost))
		return 1
	},
	"IsGet": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		L.Push(lua.LBool(ctx.r.Method == http.MethodGet))
		return 1
	},
	"IsDelete": func(L *lua.LState) int {
		ctx := luaCheckCtx(L)
		L.Push(lua.LBool(ctx.r.Method == http.MethodDelete))
		return 1
	},
}
