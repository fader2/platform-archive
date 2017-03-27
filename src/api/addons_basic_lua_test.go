package api

import (
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
	"interfaces"
	"testing"
)

func Test_basicFn_InMemoryDel(t *testing.T) {

}

func testSetup() error {
	e := echo.New()
	stngs := &Settings{}

	err := Setup(e, stngs)
	if err != nil {
		return err
	}

	return nil
}

func Test_basicFn_CreateFile(t *testing.T) {
	err := testSetup()
	assert.NoError(t, err, "Test setup")

	var L = lua.NewState()
	defer L.Close()
	SetupAddons()
	setupLuaModules(L)

	buck := &interfaces.Bucket{
		BucketID:   uuid.NewV4(),
		BucketName: "testBucket",
	}

	data := L.NewUserData()
	data.Value = buck.BucketID.String()
	L.SetGlobal("bucket_id", data)

	err = L.DoString(`
local std = require("basic")

newFile = File:new()
newFile:SetBucketID(bucket_id)
newFile:SetFileName("testFile")

ok = std.CreateFile(newFile)

print(ok)
print(id)

id = newFile:FileID()


`)
	assert.NoError(t, err)
	lValue := L.Get(-1)
	t.Log(lValue.String(), lValue.Type())
}
