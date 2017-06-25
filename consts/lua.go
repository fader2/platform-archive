package consts

import lua "github.com/yuin/gopher-lua"

func SetupToLua(L *lua.LState) {
	L.SetGlobal("GET", lua.LString(HTTP_GET))
	L.SetGlobal("POST", lua.LString(HTTP_POST))

	L.SetGlobal("DOMAIN", lua.LString(DOMAIN))
	L.SetGlobal("DEF_COOKIE_EXPIRES", lua.LString(DEF_COOKIE_EXPIRES))
	L.SetGlobal("DEF_COOKIE_SECURE", lua.LString(DEF_COOKIE_SECURE))
	L.SetGlobal("TPL_FRAGMENTS_BUCKET_NAME", lua.LString(TPL_FRAGMENTS_BUCKET_NAME))
}
