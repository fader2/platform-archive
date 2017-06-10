package api

import (
	"interfaces"
	"log"

	"fmt"

	"encoding/json"

	uuid "github.com/satori/go.uuid"
	lua "github.com/yuin/gopher-lua"
)

var exports = map[string]lua.LGFunction{
	"MemorySet": basicFn_InMemorySet,
	"MemoryGet": basicFn_InMemoryGet,
	"MemoryDel": basicFn_InMemoryDel,

	"ListBuckets":         basicFn_ListBuckets,
	"ListFilesByBucketID": basicFn_listFilesFromBucketID,

	// file manager
	"FindFileByName": basicFn_FindFileByName,
	"FindFile":       basicFn_FindFile,
	"CreateFile":     basicFn_CreateFile,
	"CreateFileFrom": basicFn_CreateFileFrom,
	"UpdateFileFrom": basicFn_UpdateFileFrom,
	"DeleteFile":     basicFn_DeleteFile,

	// bucket manager
	"FindBucketByName": basicFn_FindBucketByName,
	"FindBucket":       basicFn_FindBucket,

	"CreateBucket": basicFn_CreateBucket,
	"UpdateBucket": basicFn_UpdateBucket,
	"DeleteBucket": basicFn_DeleteBucket,
	//"CreateBucketFrom": func(L *lua.LState) int { return 0 },
	//"UpdateBucketFrom": func(L *lua.LState) int { return 0 },
}

////////////////////////////////////////////////////////////////////////////////
// Lua interfaces.DataUsed
////////////////////////////////////////////////////////////////////////////////

var luaDataUsed = "DataUsed"

func newDataUsed(v interfaces.DataUsed) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		ud := L.NewUserData()
		ud.Value = v
		L.SetMetatable(ud, L.GetTypeMetatable(luaDataUsed))
		L.Push(ud)
		return 1
	}
}

func checkDataUsed(L *lua.LState, argNum int) interfaces.DataUsed {
	ud := L.CheckUserData(argNum)
	switch ud.Value.(type) {
	case interfaces.DataUsed:
		return ud.Value.(interfaces.DataUsed)
	case int:
		return interfaces.DataUsed(uint32(ud.Value.(int)))
	case int64:
		return interfaces.DataUsed(uint32(ud.Value.(int64)))
	}

	// todo unreacheble code
	L.ArgError(1, "interfaces.DataUsed expected")
	return interfaces.DataUsed(0)
}

// types

var dataUsedMethods = map[string]lua.LGFunction{
	"Has": func(L *lua.LState) int {
		ud := L.CheckUserData(1)
		self, ok := ud.Value.(interfaces.DataUsed)
		if !ok {
			L.ArgError(1, "interfaces.DataUsed expected")
			return 0
		}
		ud = L.CheckUserData(2)
		needle, ok := ud.Value.(interfaces.DataUsed)
		if !ok {
			L.ArgError(2, "interfaces.DataUsed expected")
			return 0
		}

		L.Push(lua.LBool(self&needle != 0))
		return 1
	},
	"Add": func(L *lua.LState) int {
		var self *lua.LUserData
		var v interfaces.DataUsed

		for i := 1; i <= L.GetTop(); i++ {
			ud := L.CheckUserData(i)
			if ud == nil {
				L.ArgError(i, "interfaces.DataUsed expected, got nil")
				continue
			}
			if i == 1 {
				self = ud
			}
			_v, ok := ud.Value.(interfaces.DataUsed)
			if !ok {
				L.ArgError(i, "interfaces.DataUsed expected")
				return 0
			}
			v = v | _v
		}

		self.Value = v
		L.Push(self)

		return 1
	},
}

////////////////////////////////////////////////////////////////////////////////
// luaRoute
////////////////////////////////////////////////////////////////////////////////

var luaRouteTypeName = "route"

func newLuaRoute(route *interfaces.RouteMatch) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		ud := L.NewUserData()
		ud.Value = &luaRoute{
			Name:          route.Handler.Name,
			Path:          route.Handler.Path,
			Bucket:        route.Handler.Bucket,
			File:          route.Handler.File,
			LuaScript:     route.Handler.LuaScript,
			LuaArgsScript: route.Handler.LuaArgsScript,
			route:         route.Route,
		}
		L.SetMetatable(ud, L.GetTypeMetatable(luaRouteTypeName))
		L.Push(ud)
		return 1
	}
}

type luaRoute struct {
	Name   string
	Path   string
	Bucket string
	File   string

	AllowedAlicenses []string
	AllowedMethods   []string

	LuaScript     string
	LuaArgsScript string

	route interfaces.Route
}

func checkRoute(L *lua.LState) *luaRoute {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*luaRoute); ok {
		return v
	}
	L.ArgError(1, "route expected")
	return nil
}

// luaRoute methods

var routeMethods = map[string]lua.LGFunction{
	"Name":   rotueGetName,
	"Path":   rotueGetPath,
	"Bucket": rotueGetBucket,
	"File":   rotueGetFile,
	"Has":    routeHasRoute,

	"Lua":     rotueGetLuaScript,
	"LuaArgs": rotueGetLuaArgsScript,

	// generate URL of the current routes in the parameters
	// TODO: renate to URLPath
	"URL":     routeGetURLFromParams,
	"URLPath": routeGetURLFromParams,
}

func routeHasRoute(L *lua.LState) int {
	r := checkRoute(L)

	if L.GetTop() != 2 {
		// if exists current route then return true
		L.Push(lua.LBool(r.route != nil))
		return 1
	}

	routeName := L.CheckString(2)
	L.Push(lua.LBool(routeName == r.Name))
	return 1
}

func rotueGetName(L *lua.LState) int {
	r := checkRoute(L)
	L.Push(lua.LString(r.Name))
	return 1
}

func rotueGetPath(L *lua.LState) int {
	r := checkRoute(L)
	L.Push(lua.LString(r.Path))
	return 1
}

func rotueGetBucket(L *lua.LState) int {
	r := checkRoute(L)
	L.Push(lua.LString(r.Bucket))
	return 1
}

func rotueGetFile(L *lua.LState) int {
	r := checkRoute(L)
	L.Push(lua.LString(r.File))
	return 1
}

func rotueGetLuaScript(L *lua.LState) int {
	r := checkRoute(L)
	L.Push(lua.LString(r.LuaScript))
	return 1
}

func rotueGetLuaArgsScript(L *lua.LState) int {
	r := checkRoute(L)
	L.Push(lua.LString(r.LuaArgsScript))
	return 1
}

func routeGetURLFromParams(L *lua.LState) int {
	r := checkRoute(L)

	if r.route == nil {
		// TODO: error
		log.Println("empty router")
		return 0
	}

	var args []string

	if L.GetTop() > 1 {
		args = make([]string, L.GetTop()-1)
		for i := 2; i <= L.GetTop(); i++ {
			args[i-2] = L.CheckString(i)
		}
	}

	url, err := r.route.URLPath(args...)
	// TODO: URL as custom object
	if err != nil {
		// TODO: error
		log.Println("build url", err)
		return 0
	}
	log.Println("build url:", url.String())
	L.Push(lua.LString(url.String()))
	return 1
}

////////////////////////////////////////////////////////////////////////////////
// Bucket and file utils
////////////////////////////////////////////////////////////////////////////////

func basicFn_ListBuckets(L *lua.LState) int {
	ud := L.NewUserData()
	ud.Value = listOfBuckets()
	L.Push(ud)
	return 1
}

////////////////////////////////////////////////////////////////////////////////
// Bucket file
////////////////////////////////////////////////////////////////////////////////

var luaBucketTypeName = "bucket"

func BucketAsLuaBucket(L *lua.LState, bucket *interfaces.Bucket) *lua.LUserData {
	ud := L.NewUserData()
	if bucket == nil {
		_f := interfaces.NewBucket()
		_f.BucketID = uuid.NewV4()
		ud.Value = &luaBucket{_f}
	} else {
		ud.Value = &luaBucket{bucket}
	}
	L.SetMetatable(ud, L.GetTypeMetatable(luaBucketTypeName))
	return ud
}

func newLuaBucket(bucket *interfaces.Bucket) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		L.Push(BucketAsLuaBucket(L, bucket))
		return 1
	}

}

type luaBucket struct {
	*interfaces.Bucket
}

func checkBucket(L *lua.LState) *luaBucket {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*luaBucket); ok {
		return v
	}
	L.ArgError(1, "route expected")
	return nil
}

var bucketMethods = map[string]lua.LGFunction{
	// Setters
	"SetBucketName": func(L *lua.LState) int {
		if L.GetTop() != 2 {
			return 0
		}

		obj := checkBucket(L)
		obj.BucketName = L.CheckString(2)
		return 0
	},
	"SetLuaScript": func(L *lua.LState) int {
		if L.GetTop() != 2 {
			return 0
		}

		obj := checkBucket(L)
		obj.LuaScript = []byte(L.CheckString(2))

		return 0
	},
	"SetRawData": func(L *lua.LState) int {
		if L.GetTop() != 2 {
			return 0
		}

		obj := checkBucket(L)
		obj.RawData = []byte(L.CheckString(2))
		return 0
	},
	"SetRawDataAsBytes": func(L *lua.LState) int {
		if L.GetTop() != 2 {
			return 0
		}

		obj := checkBucket(L)
		obj.RawData = L.CheckUserData(2).Value.([]byte)
		return 0
	},
	"SetStructuralData": func(L *lua.LState) int {
		if L.GetTop() != 2 {
			return 0
		}

		// obj := checkBucket(L)
		L.ArgError(2, "SetStructuralData: not implemented")
		// obj.Bucket.StructuralData = L.CheckUserData(2).Value.([]byte)
		return 0
	},
	"SetOwners": func(L *lua.LState) int {
		if L.GetTop() != 2 {
			return 0
		}

		// obj := checkBucket(L)

		L.ArgError(2, "SetOwners: not implemented")
		return 0
	},

	// Getter

	"BucketID": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		obj := checkBucket(L)
		L.Push(lua.LString(obj.BucketID.String()))

		return 1
	},
	"BucketName": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		obj := checkBucket(L)

		L.Push(lua.LString(obj.BucketName))
		return 1
	},
	"LuaScript": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		obj := checkBucket(L)
		L.Push(lua.LString(string(obj.LuaScript)))

		return 1
	},
	"MetaData": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		// obj := checkBucket(L)

		return 0
	},
	"StructuralData": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		// obj := checkBucket(L)

		return 0
	},
	"RawData": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		obj := checkBucket(L)
		L.Push(lua.LString(string(obj.RawData)))

		return 1
	},
	"Owners": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		// obj := checkBucket(L)

		return 0
	},
}

////////////////////////////////////////////////////////////////////////////////
// File type
////////////////////////////////////////////////////////////////////////////////

var luaFileTypeName = "file"

func FileAsLuaFile(L *lua.LState, file *interfaces.File) *lua.LUserData {
	ud := L.NewUserData()
	if file == nil {
		_f := interfaces.NewFile()
		_f.FileID = uuid.NewV4()
		ud.Value = &luaFile{_f}
	} else {
		ud.Value = &luaFile{file}
	}
	L.SetMetatable(ud, L.GetTypeMetatable(luaFileTypeName))
	return ud
}

func newLuaFile(file *interfaces.File) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		L.Push(FileAsLuaFile(L, file))
		return 1
	}

}

type luaFile struct {
	*interfaces.File
}

func (f *luaFile) IsText() bool {
	_type := interfaces.GetUserTypeFromContentType(f.ContentType)
	return _type == interfaces.TextFile
}

func (f *luaFile) IsRaw() bool {
	_type := interfaces.GetUserTypeFromContentType(f.ContentType)
	return _type == interfaces.RawFile
}

func (f *luaFile) IsImage() bool {
	_type := interfaces.GetUserTypeFromContentType(f.ContentType)
	return _type == interfaces.ImageFile
}

func checkFile(L *lua.LState) *luaFile {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*luaFile); ok {
		return v
	}
	L.ArgError(1, "file expected")
	return nil
}

// luaRoute methods

var fileMethods = map[string]lua.LGFunction{
	// Setters
	"SetFileName": func(L *lua.LState) int {
		if L.GetTop() != 2 {
			return 0
		}

		log.Println("Set file name", L.CheckString(2))
		file := checkFile(L)
		file.FileName = L.CheckString(2)
		return 0
	},
	"SetBucketID": func(L *lua.LState) int {
		if L.GetTop() != 2 {
			return 0
		}

		file := checkFile(L)
		var bucketID uuid.UUID

		lv := L.Get(2)

		//
		switch lv.Type() {
		case lua.LTString:
			bucketID = uuid.FromStringOrNil(lv.(lua.LString).String())
		case lua.LTUserData:
			switch v := lv.(*lua.LUserData).Value.(type) {
			case uuid.UUID:
				bucketID = v
			case string:
				bucketID = uuid.FromStringOrNil(v)
			default:
				L.ArgError(
					2,
					fmt.Sprintf("CreateFile: not supported bucket ID type %T", v),
				)
				return 0
			}
		default:
			L.ArgError(
				2,
				fmt.Sprintf(
					"CreateFile: not supported bucket ID type %v",
					lv.Type(),
				),
			)
			return 0
		}

		if uuid.Equal(uuid.Nil, bucketID) {
			L.ArgError(
				2,
				"CreateFile: is nil ID",
			)
			return 0
		}
		//

		file.BucketID = bucketID
		return 0
	},
	"SetLuaScript": func(L *lua.LState) int {
		if L.GetTop() != 2 {
			return 0
		}

		file := checkFile(L)
		file.LuaScript = []byte(L.CheckString(2))

		return 0
	},
	"SetRawData": func(L *lua.LState) int {
		if L.GetTop() != 2 {
			return 0
		}

		file := checkFile(L)
		fmt.Println("Set raw data", L.CheckString(2))
		file.RawData = []byte(L.CheckString(2))
		return 0
	},
	"SetRawDataAsBytes": func(L *lua.LState) int {
		if L.GetTop() != 2 {
			return 0
		}

		file := checkFile(L)
		file.RawData = L.CheckUserData(2).Value.([]byte)
		return 0
	},
	"SetRawDataFromFile": func(L *lua.LState) int {
		file := checkFile(L)

		// method args
		ud := L.CheckUserData(2)
		var fileInfo *luaFormFile
		var ok bool
		if fileInfo, ok = ud.Value.(*luaFormFile); !ok {
			reason := fmt.Sprintf("form file expected, got %T", ud.Value)
			L.ArgError(2, reason)
			return 0
		}

		// main

		file.File.ContentType = fileInfo.ContentType
		file.File.RawData = fileInfo.Data
		file.File.FileName = fileInfo.FileName

		return 0
	},
	"SetStructuralData": func(L *lua.LState) int {
		if L.GetTop() != 2 {
			return 0
		}

		file := checkFile(L)

		lv := L.Get(2)
		switch lv.Type() {
		case lua.LTString:
			v := string(lv.(lua.LString))
			err := json.Unmarshal([]byte(v), &file.StructuralData)

			if err != nil {
				reason := fmt.Sprintf("error unmarshal json, %s", err)
				L.ArgError(2, reason)
				return 0
			}
		case lua.LTTable:
			v, ok := ToValueFromLValue(lv).(map[string]interface{})
			if ok {
				file.StructuralData = v
			} else {
				reason := fmt.Sprintf(
					"error transform data %T to map[string]interface{}",
					lv,
				)
				L.ArgError(2, reason)
			}

		default:
			reason := fmt.Sprintf("not supported type %T", lv)
			L.ArgError(2, reason)
		}

		return 0
	},
	"SetContentType": func(L *lua.LState) int {
		if L.GetTop() != 2 {
			return 0
		}

		file := checkFile(L)
		file.ContentType = L.CheckString(2)
		return 0
	},
	"SetOwners": func(L *lua.LState) int {
		if L.GetTop() != 2 {
			return 0
		}

		// file := checkFile(L)

		L.ArgError(2, "SetOwners: not implemented")
		return 0
	},
	"AsPrivate": func(L *lua.LState) int {
		if L.GetTop() != 2 {
			return 0
		}

		file := checkFile(L)
		file.IsPrivate = true
		return 0
	},
	"AsPublic": func(L *lua.LState) int {
		if L.GetTop() != 2 {
			return 0
		}

		file := checkFile(L)
		file.IsPrivate = false
		return 0
	},
	"AsReadOnly": func(L *lua.LState) int {
		if L.GetTop() != 2 {
			return 0
		}

		file := checkFile(L)
		file.IsReadOnly = true
		return 0
	},

	"AsNotReadOnly": func(L *lua.LState) int {
		if L.GetTop() != 2 {
			return 0
		}

		file := checkFile(L)
		file.IsReadOnly = false
		return 0
	},

	// Getter

	"FileID": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		file := checkFile(L)

		L.Push(lua.LString(file.FileID.String()))
		return 1
	},
	"FileName": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		file := checkFile(L)

		L.Push(lua.LString(file.FileName))
		return 1
	},
	"BucketID": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		file := checkFile(L)
		L.Push(lua.LString(file.BucketID.String()))

		return 1
	},
	"LuaScript": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		file := checkFile(L)
		L.Push(lua.LString(string(file.LuaScript)))

		return 1
	},
	"MetaData": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		file := checkFile(L)
		L.Push(ToLValueOrNil(file.MetaData, L))
		return 1
	},
	"StructuralData": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		file := checkFile(L)
		L.Push(ToLValueOrNil(file.StructuralData, L))
		return 1
	},
	"RawData": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		file := checkFile(L)
		L.Push(lua.LString(string(file.RawData)))

		return 1
	},
	"ContentType": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		file := checkFile(L)
		L.Push(lua.LString(file.ContentType))

		return 1
	},
	"Owners": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		// file := checkFile(L)

		return 0
	},
	"IsPrivate": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		file := checkFile(L)
		L.Push(lua.LBool(file.IsPrivate))

		return 1
	},
	"IsReadOnly": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		file := checkFile(L)
		L.Push(lua.LBool(file.IsPrivate))

		return 1
	},
	"IsImage": func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		file := checkFile(L)
		_type := interfaces.GetUserTypeFromContentType(file.ContentType)
		L.Push(lua.LBool(_type == interfaces.ImageFile))

		return 1
	},
	"IsText": func(L *lua.LState) int {
		L.Push(lua.LBool(false))
		return 1

		if L.GetTop() != 1 {
			return 0
		}

		file := checkFile(L)
		_type := interfaces.GetUserTypeFromContentType(file.ContentType)
		L.Push(lua.LBool(_type == interfaces.TextFile))

		return 1
	},
	"IsRaw": func(L *lua.LState) int {
		L.Push(lua.LBool(false))
		return 1

		if L.GetTop() != 1 {
			return 0
		}

		file := checkFile(L)
		_type := interfaces.GetUserTypeFromContentType(file.ContentType)
		L.Push(lua.LBool(_type == interfaces.RawFile))

		return 1
	},
}

////////////////////////////////////////////////////////////////////////////////
// File manager
////////////////////////////////////////////////////////////////////////////////

// список файлов в бакете
func basicFn_listFilesFromBucketID(L *lua.LState) int {
	var bid uuid.UUID
	if L.GetTop() == 2 {
		switch v := L.CheckUserData(2).Value.(type) {
		case uuid.UUID:
			bid = v
		case string:
			bid = uuid.FromStringOrNil(v)
		}
	}

	ud := L.NewUserData()
	ud.Value = filesByBucketID(bid)
	L.Push(ud)
	return 1
}

// найти файл по имени бакета и имени файла
func basicFn_FindFileByName(L *lua.LState) int {
	/*
		bucketName, fileName string,
		used DataUsed,
	*/
	var bucketName, fileName string
	var used interfaces.DataUsed

	bucketName = L.CheckString(1)
	fileName = L.CheckString(2)
	used = interfaces.FullFile

	file, err := fileManager.FindFileByName(
		bucketName,
		fileName,
		used,
	)

	if err != nil {
		L.RaiseError("FindFileByName: find file by name %s, err %s", bucketName+" "+fileName, err)
		return 0
	}

	return newLuaFile(file)(L)
}

func basicFn_FindFile(L *lua.LState) int {
	var id uuid.UUID
	if L.GetTop() == 1 {
		lv := L.Get(1)
		switch lv.Type() {
		case lua.LTString:
			id = uuid.FromStringOrNil(lv.(lua.LString).String())
		case lua.LTUserData:
			switch v := lv.(*lua.LUserData).Value.(type) {
			case uuid.UUID:
				id = v
			case string:
				id = uuid.FromStringOrNil(v)
			default:
				L.ArgError(
					1,
					fmt.Sprintf("FindFile: not supported file ID type %T", v),
				)
				return 0
			}
		default:
			L.ArgError(
				1,
				fmt.Sprintf(
					"FindFile: not supported file ID type %v",
					lv.Type(),
				),
			)
			return 0
		}
	}

	if uuid.Equal(uuid.Nil, id) {
		L.ArgError(
			1,
			"FindFile: is nil ID",
		)
		return 0
	}

	file, err := fileManager.FindFile(id, interfaces.FullFile)

	if err != nil {
		L.RaiseError("FindFile: find file by ID %s, err %s", id, err)
		return 0
	}

	return newLuaFile(file)(L)
}

func basicFn_CreateFileFrom(L *lua.LState) int {
	file := checkFile(L)

	mode := L.CheckUserData(2).Value.(interfaces.DataUsed)

	err := fileManager.CreateFileFrom(file.File, mode)
	if err != nil {
		L.RaiseError("create file %s, err %s", file.FileID, err)
		L.Push(lua.LBool(false))
		return 1
	}

	L.Push(lua.LBool(true))
	return 1
}

func basicFn_UpdateFileFrom(L *lua.LState) int {
	file := checkFile(L)
	mode := L.CheckUserData(2).Value.(interfaces.DataUsed)

	err := fileManager.UpdateFileFrom(file.File, mode)

	log.Printf("update file %s, name %s, mode %v", file.FileID, file.FileName, mode)

	if err != nil {
		L.RaiseError("update file %s, err %s", file.FileID, err)
		L.Push(lua.LBool(false))
		return 1
	}

	L.Push(lua.LBool(true))
	return 1
}

func basicFn_DeleteFile(L *lua.LState) int {
	var id uuid.UUID

	lv := L.Get(1)
	switch lv.Type() {
	case lua.LTString:
		id = uuid.FromStringOrNil(lv.(lua.LString).String())
	case lua.LTUserData:
		switch v := lv.(*lua.LUserData).Value.(type) {
		case uuid.UUID:
			id = v
		case string:
			id = uuid.FromStringOrNil(v)
		default:
			L.ArgError(
				1,
				fmt.Sprintf("DeleteFile: not supported ID type %T", v),
			)
			L.Push(lua.LBool(false))
			return 1
		}
	default:
		L.ArgError(
			1,
			fmt.Sprintf(
				"DeleteFile: not supported ID type %v",
				lv.Type(),
			),
		)
		L.Push(lua.LBool(false))
		return 1
	}

	if uuid.Equal(uuid.Nil, id) {
		L.ArgError(
			1,
			"DeleteFile: empty ID",
		)
		L.Push(lua.LBool(false))
		return 1
	}

	if err := fileManager.DeleteFile(id); err != nil {
		L.RaiseError(
			"DeleteFile: error delete file %v, err: %s",
			id,
			err,
		)
		L.Push(lua.LBool(false))
		return 1
	}

	log.Println("DeleteFile: success delete file ID", id)

	L.Push(lua.LBool(true))
	return 1
}

////////////////////////////////////////////////////////////////////////////////
// Bucket manager
////////////////////////////////////////////////////////////////////////////////

func basicFn_CreateBucket(L *lua.LState) int {
	bucketFile := interfaces.NewBucket()
	bucketFile.BucketID = uuid.NewV4()
	bucketFile.BucketName = L.CheckString(1)

	if err := bucketManager.CreateBucket(bucketFile); err != nil {
		L.RaiseError("create bucket %s, err %s", bucketFile.BucketName, err)
		L.Push(lua.LBool(false))
		return 0
	}

	L.Push(lua.LBool(true))
	return 1
}

func basicFn_UpdateBucket(L *lua.LState) int {
	bucket := checkBucket(L)
	if bucket == nil {
		L.ArgError(1, fmt.Sprintf("UpdateBucket: first argument error. Need bucket."))
		return 0
	}
	used := checkDataUsed(L, 2)
	if used == interfaces.DataUsed(0) {
		used = interfaces.FullBucket
	}

	err := bucketManager.UpdateBucket(bucket.Bucket, used)
	if err != nil {
		L.RaiseError("update bucket %s, err %s", bucket.BucketName, err)
		L.Push(lua.LBool(false))
		return 1
	}

	L.Push(lua.LBool(true))
	return 1
}

// todo
func basicFn_DeleteBucket(L *lua.LState) int {
	//bucketName := L.CheckString(1)

	return 0
}

func basicFn_FindBucketByName(L *lua.LState) int {
	name := L.CheckString(1)
	bucket, err := bucketManager.FindBucketByName(name, interfaces.FullBucket)
	if err != nil {
		L.RaiseError("FindBucketByName: find bucket by name %s, err %s", name, err)
		return 0
	}
	return newLuaBucket(bucket)(L)
}

func basicFn_FindBucket(L *lua.LState) int {
	var id uuid.UUID
	if L.GetTop() == 1 {
		lv := L.Get(1)
		switch lv.Type() {
		case lua.LTString:
			id = uuid.FromStringOrNil(lv.(lua.LString).String())
		case lua.LTUserData:
			switch v := lv.(*lua.LUserData).Value.(type) {
			case uuid.UUID:
				id = v
			case string:
				id = uuid.FromStringOrNil(v)
			default:
				L.ArgError(
					1,
					fmt.Sprintf("FindBucket: not supported file ID type %T", v),
				)
				return 0
			}
		default:
			L.ArgError(
				1,
				fmt.Sprintf(
					"FindBucket: not supported file ID type %v",
					lv.Type(),
				),
			)
			return 0
		}
	}

	if uuid.Equal(uuid.Nil, id) {
		L.ArgError(
			1,
			"FindBucket: is nil ID",
		)
		return 0
	}

	bucket, err := bucketManager.FindBucket(id, interfaces.FullBucket)
	if err != nil {
		L.RaiseError("FindBucket: find bucket by ID %s, err %s", id, err)
		return 0
	}

	return newLuaBucket(bucket)(L)
}

// Memory storage

func basicFn_InMemorySet(L *lua.LState) int {
	key := L.CheckString(1)
	lv := L.CheckUserData(2)
	obj, ok := lv.Value.(interfaces.MsgpackMarshaller)
	if !ok {
		L.RaiseError(
			"InMemorySet: expected MsgpackMarshaller obj, got %T",
			lv.Value,
		)
		L.Push(lua.LBool(false))
		return 1
	}
	if err := inMemStore.Set(key, obj); err != nil {
		L.RaiseError(
			"InMemorySet: error setting, %s",
			err,
		)
		L.Push(lua.LBool(false))
		return 1
	}
	L.Push(lua.LBool(true))
	return 1
}

func basicFn_InMemoryGet(L *lua.LState) int {
	key := L.CheckString(1)
	lv := L.CheckUserData(2)
	obj, ok := lv.Value.(interfaces.MsgpackMarshaller)
	if !ok {
		L.RaiseError(
			"InMemoryGet: expected MsgpackMarshaller obj, got %T",
			lv.Value,
		)
		L.Push(lua.LBool(false))
		return 1
	}
	if err := inMemStore.Get(key, obj); err != nil {
		L.RaiseError(
			"InMemoryGet: errro getting, %s",
			err,
		)
		L.Push(lua.LBool(false))
		return 1
	}

	L.Push(lua.LBool(true))
	return 1
}

func basicFn_InMemoryDel(L *lua.LState) int {
	key := L.CheckString(1)
	if err := inMemStore.Del(key); err != nil {
		L.RaiseError(
			"InMemoryDel: error deleting, %s",
			err,
		)
		L.Push(lua.LBool(false))
		return 1
	}

	L.Push(lua.LBool(true))
	return 1
}

func basicFn_CreateFile(L *lua.LState) int {
	file := checkFile(L)

	err := fileManager.CreateFile(file.File)

	if err != nil {
		L.RaiseError("create file %s, err %s", file.FileID, err)
		L.Push(lua.LBool(false))
		return 1
	}

	L.Push(lua.LBool(true))
	return 1
}
