package core

import (
	"fmt"
	"log"

	"github.com/fader2/platform/objects"
	uuid "github.com/satori/go.uuid"
	lua "github.com/yuin/gopher-lua"
)

var (
	luaUserTypeName = "user"
)

type user struct {
	*objects.User
}

func newEmptyUserWithAccess(id uuid.UUID, access ...string) (u *user) {
	u = &user{
		User: objects.EmptyUser(objects.UnknownUserType),
	}
	u.User.ID = id
	u.User.Info.Pasport.Access = access
	return
}

func registerUserType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaUserTypeName)
	L.SetGlobal("user", mt)
	L.SetField(mt, "new", L.NewFunction(func(L *lua.LState) int {
		ud := L.NewUserData()
		if L.GetTop() == 1 {
			ud.Value = newEmptyUserWithAccess(objects.UUIDFromString(L.CheckString(1)), "guest")
		} else {
			ud.Value = newEmptyUserWithAccess(uuid.NewV4(), "guest")
		}

		L.SetMetatable(ud, L.GetTypeMetatable(luaUserTypeName))
		L.Push(ud)
		return 1
	}))
	L.SetField(mt, "findByLogin", L.NewFunction(func(L *lua.LState) int {
		s := luaCheckStore(1, L)
		id := objects.UUIDFromString(L.CheckString(2))
		if uuid.Equal(uuid.Nil, id) {
			L.ArgError(2, "empty or invalud UUID")
			return 0
		}

		return luaFindUserByID(id, s, L)
	}))
	L.SetField(mt, "find", L.NewFunction(func(L *lua.LState) int {
		s := luaCheckStore(1, L)
		id := uuid.FromStringOrNil(L.CheckString(2))
		if uuid.Equal(uuid.Nil, id) {
			L.ArgError(2, "empty or invalud UUID")
			return 0
		}

		// main

		return luaFindUserByID(id, s, L)
	}))
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), luaUserMethods))
	return
}

func luaCheckUser(L *lua.LState, numarg ...int) *user {
	var ud *lua.LUserData
	if len(numarg) == 1 {
		ud = L.CheckUserData(numarg[0])
	} else {
		ud = L.CheckUserData(1)
	}

	if v, ok := ud.Value.(*user); ok {
		return v
	}
	reason := fmt.Sprintf("expected user, got %T", ud.Value)
	L.ArgError(1, reason)
	return nil
}

var luaUserMethods = map[string]lua.LGFunction{
	"isGuest": func(L *lua.LState) int {
		u := luaCheckUser(L)
		L.Push(lua.LBool(u.IsGuest()))
		return 1
	},
	"id": func(L *lua.LState) int {
		u := luaCheckUser(L)
		L.Push(lua.LString(u.User.ID.String()))
		return 1
	},
	"save": func(L *lua.LState) int {
		u := luaCheckUser(L)
		s := luaCheckStore(2, L)
		id, err := objects.SetUser(s, u.User)
		log.Println("find user by ID", u.User.ID, err)
		if err != nil {
			L.RaiseError("error save user. %s", err)
			return 0
		}

		L.Push(lua.LString(id.String()))
		return 1
	},
	"meta": func(L *lua.LState) int {
		u := luaCheckUser(L)
		k := L.CheckString(2)
		if L.GetTop() == 2 {
			L.Push(lua.LString(u.Meta.Get(k)))
			return 1
		}
		if L.GetTop() == 3 {
			v := L.CheckString(3)
			u.Meta.Set(k, v)
			return 0
		}

		L.RaiseError("user:meta() - not expected number of arguments")
		return 0
	},
	"email": func(L *lua.LState) int {
		u := luaCheckUser(L)
		if L.GetTop() == 2 {
			u.Info.Pasport.Email = L.CheckString(2)
			return 0
		}
		L.Push(lua.LString(u.Info.Pasport.Email))
		return 1
	},
	"login": func(L *lua.LState) int {
		u := luaCheckUser(L)
		if L.GetTop() == 2 {
			u.Info.Pasport.UserName = L.CheckString(2)
			return 0
		}
		L.Push(lua.LString(u.Info.Pasport.UserName))
		return 1
	},
	"pwd": func(L *lua.LState) int {
		u := luaCheckUser(L)
		if L.GetTop() == 2 {
			u.User.SetPWD(L.CheckString(2))
			return 0
		}
		return 0
	},
	"matchPwd": func(L *lua.LState) int {
		u := luaCheckUser(L)
		if L.GetTop() == 2 {
			L.Push(lua.LBool(u.User.MatchPWD(L.CheckString(2))))
			return 1
		}
		return 0
	},
}

func luaFindUserByID(id uuid.UUID, s objects.Storer, L *lua.LState) int {
	u, err := objects.GetUser(s, id)
	if err != nil {
		ud := L.NewUserData()
		ud.Value = newEmptyUserWithAccess(id, "guest")
		L.SetMetatable(ud, L.GetTypeMetatable(luaUserTypeName))
		L.Push(ud)
		L.Push(lua.LBool(false))
		return 2
	}

	ud := L.NewUserData()
	ud.Value = &user{
		User: u,
	}
	L.SetMetatable(ud, L.GetTypeMetatable(luaUserTypeName))
	L.Push(ud)
	L.Push(lua.LBool(true))
	return 2
}
