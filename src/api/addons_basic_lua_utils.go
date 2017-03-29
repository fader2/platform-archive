package api

import (
	"github.com/yuin/gopher-lua"
)

func luaGetString(L *lua.LState, name string) (string, bool) {
	val := L.GetGlobal(name)
	if res, ok := val.(lua.LString); ok {
		return res.String(), true
	} else {
		return "", false
	}
}

func luaGetBool(L *lua.LState, name string) (bool, bool) {
	val := L.GetGlobal(name)
	if _, ok := val.(lua.LBool); ok {
		return lua.LVAsBool(val), true
	} else {
		return false, false
	}
}
