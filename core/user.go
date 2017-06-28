package core

import (
	"fmt"

	"github.com/fader2/platform/objects"
	lua "github.com/yuin/gopher-lua"
)

const luaUserTypeName = "user"

type luaUser struct {
	*objects.User
	Exists bool
}

func luaAddCustomTypeUser(L *lua.LState) {
	mt := L.NewTypeMetatable(luaUserTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), luaUserMethods))
	return
}

func luaUserBuilder(ctx *luaUser) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		ud := L.NewUserData()
		ud.Value = ctx

		L.SetMetatable(ud, L.GetTypeMetatable(luaUserTypeName))
		L.Push(ud)
		return 1
	}
}

func luaCheckUser(L *lua.LState) *luaUser {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*luaUser); ok {
		return v
	}
	reason := fmt.Sprintf("expected luaUser, got %T", ud.Value)
	L.ArgError(1, reason)
	return nil
}

var luaUserMethods = map[string]lua.LGFunction{
	"IsGuest": func(L *lua.LState) int {
		u := luaCheckUser(L)
		L.Push(lua.LBool(u.IsGuest()))
		return 1
	},
	"IsAccess": func(L *lua.LState) int {
		u := luaCheckUser(L)
		L.Push(lua.LBool(u.IsAccess(L.CheckString(2))))
		return 1
	},
	"Type": func(L *lua.LState) int {
		u := luaCheckUser(L)
		L.Push(lua.LString(u.Type()))
		return 1
	},
	"IsExists": func(L *lua.LState) int {
		u := luaCheckUser(L)
		L.Push(lua.LBool(u.Exists))
		return 1
	},
}
