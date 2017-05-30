package example

import lua "github.com/yuin/gopher-lua"

var exports = map[string]lua.LGFunction{
	"myfunc": myfunc,
}

func myfunc(L *lua.LState) int {
	L.Push(lua.LString("myfunc value"))
	return 1
}
