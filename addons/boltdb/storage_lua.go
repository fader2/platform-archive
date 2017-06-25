package boltdb

import (
	"github.com/fader2/platform/objects"
	"github.com/fader2/platform/utils"
	lua "github.com/yuin/gopher-lua"
)

var luaBoltdbTypeName = "storage_boltdb"

type luaBoltdb struct {
	s *BoltdbStorage
}

func setupCustomTypes(L *lua.LState) {
	// FormFile
	mt := L.NewTypeMetatable(luaBoltdbTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), luaBoltdbMethods))
}

func newUserData(s *BoltdbStorage) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		ud := L.NewUserData()
		ud.Value = &luaBoltdb{
			s: s,
		}
		L.SetMetatable(ud, L.GetTypeMetatable(luaBoltdbTypeName))
		L.Push(ud)
		return 1
	}
}

func checkLuaBoltdb(L *lua.LState) *luaBoltdb {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*luaBoltdb); ok {
		return v
	}
	L.ArgError(1, "not expectd type object")
	return nil
}

var luaBoltdbMethods = map[string]lua.LGFunction{
	"Set": func(L *lua.LState) int {
		ls := checkLuaBoltdb(L)

		name := L.CheckString(2)
		lv := L.CheckAny(3)
		v := utils.ToValueFromLValue(lv)

		b := objects.EmptyBlob()
		b.SetIDFromName(name)
		b.SetDataFromValue(v)
		b.SetOrigName(name)

		id, err := objects.SetBlob(ls.s, b)
		if err != nil {
			L.RaiseError("error save object", err)
			return 0
		}

		L.Push(lua.LString(id.String()))
		return 1
	},
	"Get": func(L *lua.LState) int {
		ls := checkLuaBoltdb(L)
		name := L.CheckString(2)
		id := objects.UUIDFromString(name)

		b, err := objects.GetBlob(ls.s, id)
		if err != nil {
			L.RaiseError(
				"error get blob object by name %q, id=%q: %s",
				name,
				id.String(),
				err,
			)
			return 0
		}

		return b.PushDataToLState(L)
	},
	"Del": func(L *lua.LState) int { return 0 },
}
