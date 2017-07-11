package boltdb

import (
	"github.com/boltdb/bolt"
	"github.com/fader2/platform/objects"
	"github.com/fader2/platform/utils"
	uuid "github.com/satori/go.uuid"
	lua "github.com/yuin/gopher-lua"
)

var luaBoltdbTypeName = "storage_boltdb"

type luaBoltdb struct {
	*BoltdbStorage
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
			s,
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
	"set": luaUpsertObject(),
	"setRefID": func(L *lua.LState) int {
		s := checkLuaBoltdb(L)
		name := L.CheckString(2)
		lv := L.CheckUserData(3)
		var id uuid.UUID
		switch v := lv.Value.(type) {
		case string:
			id = uuid.FromStringOrNil(v)
		case uuid.UUID:
			id = v
		default:
			L.RaiseError("unsupported type %T", v)
		}

		err := s.db.Update(func(tx *bolt.Tx) error {
			return tx.Bucket([]byte(s.bucket)).
				Put(utils.SHA256(name), id.Bytes())
		})
		if err != nil {
			L.RaiseError("save ref %s", err)
			return 0
		}

		return 0
	},
	"get": func(L *lua.LState) int {
		store := checkLuaBoltdb(L)
		name := L.CheckString(2)
		id := objects.UUIDFromString(name)

		b, err := objects.GetBlob(store, id)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}

		return b.PushDataToLState(L)
	},
	"getRefID": func(L *lua.LState) int {
		s := checkLuaBoltdb(L)
		name := L.CheckString(2)
		var idRaw []byte
		err := s.db.View(func(tx *bolt.Tx) error {
			idRaw = tx.Bucket([]byte(s.bucket)).
				Get(utils.SHA256(name))
			return nil
		})
		if err != nil {
			L.RaiseError("save ref %s", err)
			return 0
		}

		L.Push(lua.LString(uuid.FromBytesOrNil(idRaw).String()))
		return 1
	},
	"del": func(L *lua.LState) int {
		s := checkLuaBoltdb(L)
		name := L.CheckString(2)
		id := objects.UUIDFromString(name)
		s.db.Update(func(tx *bolt.Tx) error {
			return tx.Bucket([]byte(s.bucket)).Delete(id.Bytes())
		})
		return 0
	},
	"delRefID": func(L *lua.LState) int {
		s := checkLuaBoltdb(L)
		name := L.CheckString(2)
		s.db.Update(func(tx *bolt.Tx) error {
			return tx.Bucket([]byte(s.bucket)).Delete(utils.SHA256(name))
		})
		return 0
	},
}

func luaUpsertObject() func(L *lua.LState) int {
	return func(L *lua.LState) int {
		s := checkLuaBoltdb(L)

		name := L.CheckString(2)
		lv := L.CheckAny(3)
		v := utils.ToValueFromLValue(lv)

		id, err := upsertObject(s.BoltdbStorage, name, v)
		if err != nil {
			L.RaiseError("error save object", err)
			return 0
		}

		L.Push(lua.LString(id.String()))
		return 1
	}
}

func upsertObject(
	s *BoltdbStorage,
	name string,
	v interface{},
) (uuid.UUID, error) {

	b := objects.EmptyBlob()
	b.SetIDFromName(name)
	b.SetDataFromValue(v)
	b.SetOrigName(name)

	id, err := objects.SetBlob(s, b)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
