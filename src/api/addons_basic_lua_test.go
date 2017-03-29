package api

import (
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
	"interfaces"
	"os"
	"testing"
)

func Test_basicFn_InMemoryDel(t *testing.T) {

}

func testSetup() error {
	e := echo.New()
	stngs := &Settings{DatabasePath: "./.testdb"}

	err := Setup(e, stngs)
	if err != nil {
		return err
	}

	return nil
}

func cleanTestDb() {
	os.RemoveAll("./.testdb")
}

func Test_basicFn_CreateFile(t *testing.T) {
	err := testSetup()
	assert.NoError(t, err, "Test setup")
	defer cleanTestDb()

	var L = lua.NewState()
	defer L.Close()
	SetupAddons()
	setupLuaModules(L)

	bucketName := "createFileBucket"

	buck := &interfaces.Bucket{
		BucketID:   uuid.NewV4(),
		BucketName: bucketName,
	}

	err = bucketManager.CreateBucket(buck)
	assert.NoError(t, err, "Create bucket")

	data := L.NewUserData()
	data.Value = buck.BucketID.String()
	L.SetGlobal("bucket_id", data)

	err = L.DoString(`
local std = require("basic")

newFile = File:new()
newFile:SetBucketID(bucket_id)
newFile:SetFileName("testFile")

ok = std.CreateFile(newFile)

id = newFile:FileID()
print(id)

`)
	assert.NoError(t, err)
	ok, hasOk := luaGetBool(L, "ok")
	assert.Equal(t, true, ok, "test create file done")
	assert.Equal(t, true, hasOk, "ok has")
	id, hasId := luaGetString(L, "id")
	assert.NotEmpty(t, id, "id should be not empty")
	assert.Equal(t, true, hasId, "id has")

}

func Test_basicFn_CreateFileFrom(t *testing.T) {
	err := testSetup()
	assert.NoError(t, err, "Test setup")
	defer cleanTestDb()

	var L = lua.NewState()
	defer L.Close()
	SetupAddons()
	setupLuaModules(L)

	bucketName := "createFileFromBucket"
	fileName := "cretefileformName"

	buck := &interfaces.Bucket{
		BucketID:   uuid.NewV4(),
		BucketName: bucketName,
	}

	err = bucketManager.CreateBucket(buck)
	assert.NoError(t, err, "Create bucket")

	data := L.NewUserData()
	data.Value = buck.BucketID.String()
	L.SetGlobal("bucket_id", data)

	err = L.DoString(`
local std = require("basic")

newFile = File:new()
newFile:SetBucketID(bucket_id)
newFile:SetFileName("` + fileName + `")
newFile:SetRawData("alo")

used = std.RawData

ok = std.CreateFileFrom(newFile,used)

id = newFile:FileID()
bucket_id = newFile:BucketID()
print(id)
`)
	assert.NoError(t, err)

	ok, hasOk := luaGetBool(L, "ok")
	assert.Equal(t, true, ok, "test create file done")
	assert.Equal(t, true, hasOk, "ok has")

	id, hasId := luaGetString(L, "id")
	assert.NotEmpty(t, id, "id should be not empty")
	assert.Equal(t, true, hasId, "id has")

	buckId, hasId := luaGetString(L, "bucket_id")
	assert.NotEmpty(t, buckId, "bucket_id should be not empty")
	assert.Equal(t, true, hasId, "bucket_id has")

	assert.Equal(t, buck.BucketID, uuid.FromStringOrNil(buckId), "bucket ids sholud be equal")

	file, err := fileManager.FindFileByName(bucketName, fileName, interfaces.FullFile)
	assert.NoError(t, err)
	assert.Equal(t, "alo", string(file.RawData), "raw data should be alo")

}

func Test_basicFn_UpdateFile(t *testing.T) {
	err := testSetup()
	defer cleanTestDb()
	assert.NoError(t, err, "Test setup")

	var L = lua.NewState()
	defer L.Close()
	SetupAddons()
	setupLuaModules(L)

	bucketName := "createUpdateBucket"

	buck := &interfaces.Bucket{
		BucketID:   uuid.NewV4(),
		BucketName: bucketName,
	}

	err = bucketManager.CreateBucket(buck)
	assert.NoError(t, err, "Create bucket")

	data := L.NewUserData()
	data.Value = buck.BucketID.String()
	L.SetGlobal("bucket_id", data)

	err = L.DoString(`
local std = require("basic")

newFile = File:new()
newFile:SetBucketID(bucket_id)
newFile:SetFileName("testFile")

ok = std.CreateFile(newFile)
id = newFile:FileID()

vvv = std.RawData

newFile:SetRawData("opacha")

std.UpdateFileFrom(newFile, vvv)

`)

	file, err := fileManager.FindFileByName(bucketName, "testFile", interfaces.RawData)
	assert.NoError(t, err)
	assert.NotNil(t, file, "file not nil")
	assert.Equal(t, "opacha", string(file.RawData))
}

func Test_basicFn_FindFileByName(t *testing.T) {
	err := testSetup()
	defer cleanTestDb()
	assert.NoError(t, err, "Test setup")

	var L = lua.NewState()
	defer L.Close()
	SetupAddons()
	setupLuaModules(L)

	bucketName := "createffbnBucket"

	buck := &interfaces.Bucket{
		BucketID:   uuid.NewV4(),
		BucketName: bucketName,
	}

	err = bucketManager.CreateBucket(buck)
	assert.NoError(t, err, "Create bucket")

	file := interfaces.NewFile()
	file.BucketID = buck.BucketID
	file.FileName = "find_me_file"
	file.FileID = uuid.NewV4()

	err = fileManager.CreateFile(file)
	assert.NoError(t, err, "Create file")

	newFile, err := fileManager.FindFile(file.FileID, interfaces.FullFile)
	assert.NoError(t, err, "find file")
	assert.Equal(t, file.FileName, newFile.FileName)

	err = L.DoString(`
local std = require("basic")

file = std.FindFileByName("` + bucketName + `", "find_me_file")

name = file:FileName()

`)
	assert.NoError(t, err)
	id, hasName := luaGetString(L, "name")
	assert.NotEmpty(t, id, "id should be not empty")
	assert.Equal(t, true, hasName, "id has")
}

func TestLua(t *testing.T) {
	code := `id = "is_id"`

	var L = lua.NewState()
	defer L.Close()

	err := L.DoString(code)
	assert.NoError(t, err, "lua do string")

	lVal := L.GetGlobal("id")
	t.Log(lVal, lVal.String(), lVal.Type())
}
