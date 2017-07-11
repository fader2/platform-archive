package core

import (
	"strings"

	"github.com/fader2/platform/objects"
	lua "github.com/yuin/gopher-lua"
)

func luaCheckStore(num int, L *lua.LState) (s objects.Storer) {
	ud := L.CheckUserData(num)
	var ok bool
	s, ok = ud.Value.(objects.Storer)
	if !ok {
		return nil
	}
	return
}

func luaGlobalHelpFuncs(L *lua.LState) {
	L.SetGlobal("uuidFrom", L.NewFunction(func(L *lua.LState) int {
		v := L.CheckString(1)
		id := objects.UUIDFromString(v)
		L.Push(lua.LString(id.String()))
		return 1
	}))

	L.SetGlobal("split", L.NewFunction(func(L *lua.LState) int {
		arr := strings.Split(L.CheckString(1), L.CheckString(2))
		ud := L.NewTable()

		for _, str := range arr {
			ud.Append(lua.LString(str))
		}
		L.Push(ud)
		return 1
	}))

	L.SetGlobal("join", L.NewFunction(func(L *lua.LState) int {
		t := L.CheckTable(1)
		sep := L.CheckString(2)
		arr := []string{}
		t.ForEach(func(k, v lua.LValue) {
			arr = append(arr, v.String())
		})
		L.Push(lua.LString(strings.Join(arr, sep)))
		return 1
	}))
}
