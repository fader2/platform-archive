package consts

import lua "github.com/yuin/gopher-lua"

func SetupToLua(L *lua.LState) {
	L.SetGlobal("GET", lua.LString(HTTP_GET))
	L.SetGlobal("POST", lua.LString(HTTP_POST))

	L.SetGlobal("DOMAIN", lua.LString(DOMAIN))
	L.SetGlobal("COOKIE_EXPIRES", lua.LString(COOKIE_EXPIRES))
	L.SetGlobal("COOKIE_SECURE", lua.LString(COOKIE_SECURE))
	L.SetGlobal("TPL_FRAGMENTS_BUCKET_NAME", lua.LString(TPL_FRAGMENTS_BUCKET_NAME))

	L.SetGlobal("JWT_EXPIRATION_DELTA", lua.LString(JWT_EXPIRATION_DELTA))
}
