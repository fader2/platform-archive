package boltdb

import (
	"errors"
	"log"

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

func newLuaRoute(s *BoltdbStorage) func(L *lua.LState) int {
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

		k := L.CheckString(2)
		lv := L.CheckAny(3)
		// lv := utils.ToValueFromLValue()

		b := objects.EmptyBlob()
		if err := luaSetDataFromLValue(k, lv, b); err != nil {
			L.RaiseError("error set data from content type: %s", err)
			return 0
		}
		b.Meta.Set("Original-Name", k)

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
		k := L.CheckString(2)

		b, err := objects.GetBlob(
			ls.s,
			objects.UUIDFromString(k),
		)
		if err != nil {
			L.RaiseError("error get blob object: %s", err)
			return 0
		}

		log.Println(">> content type", b.Meta.Get("Content-Type"))
		log.Println(">> original name", b.Meta.Get("Original-Name"))

		return luaPushValue(L, b)
	},
	"Del": func(L *lua.LState) int { return 0 },
}

func luaSetDataFromLValue(k string, in lua.LValue, b *objects.Blob) error {
	v := utils.ToValueFromLValue(in)
	ct := objects.TypeFrom(v)
	b.Meta.Set("Content-Type", ct.String())
	b.ID = objects.UUIDFromString(k)

	switch ct {
	case objects.TString:
		b.Data = []byte(v.(string))
	case objects.TNumber:
		b.Data = utils.Float64bytes(utils.ToFloat64(v))
	case objects.TBool:
		if v.(bool) {
			b.Data = utils.Float64bytes(1)
		} else {
			b.Data = utils.Float64bytes(0)
		}

	default:
		return errors.New("not supported content type " + ct.String())
	}

	return nil
}

func luaPushValue(L *lua.LState, b *objects.Blob) int {
	ct := b.Meta.Get("Content-Type")
	switch objects.ContentType(ct) {
	case objects.TNumber:
		v := utils.Float64frombytes(b.Data)
		L.Push(lua.LNumber(v))
		return 1
	case objects.TString:
		L.Push(lua.LString(string(b.Data)))
		return 1
	case objects.TBool:
		v := utils.Float64frombytes(b.Data)
		L.Push(lua.LBool(v == 1))
		return 1
	default:
		L.RaiseError("not supported content type", ct)
	}

	return 0
}
