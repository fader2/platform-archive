package api

import (
	"encoding/json"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/yuin/gopher-lua"
	"interfaces"
	"os"
	"testing"
)

func TestContextMethodIs(t *testing.T) {
	var L = lua.NewState()
	defer L.Close()
	SetupAddons()
	setupLuaModules(L)
	setupLuaContext("POST", "/", nil, L)

	err := L.DoString(`
	isPost = ctx():IsPOST()
`)

	assert.NoError(t, err, "execute lua")
	isPost, has := luaGetBool(L, "isPost")
	assert.Equal(t, true, isPost, "is post")
	assert.Equal(t, true, has, "has isPost variable")
	L.Close()

	L = lua.NewState()
	SetupAddons()
	setupLuaModules(L)
	setupLuaContext("GET", "/", nil, L)

	err = L.DoString(`
	isGet = ctx():IsGET()
`)

	assert.NoError(t, err, "execute lua")
	isGet, has := luaGetBool(L, "isGet")
	assert.Equal(t, true, isGet, "is post")
	assert.Equal(t, true, has, "has isGet variable")
	L.Close()

	L = lua.NewState()
	SetupAddons()
	setupLuaModules(L)
	//setupLuaContext("GET", "/", nil, L)

	err = L.DoString(`
	isPost = ctx():IsGET()
`)

	assert.Error(t, err, "execute lua without context")
}

func TestContextRenderJSON(t *testing.T) {
	var L = lua.NewState()
	defer L.Close()
	SetupAddons()
	setupLuaModules(L)
	_, resp := setupLuaContextWithResp("POST", "/", nil, L)

	err := L.DoString(`
	a = {}
	a["key"] = "value"
	a["int"] = 3
	a["float"] = 3.14
	ctx():JSON(200, a)
`)

	assert.NoError(t, err, "lua json render")
	type respStruct struct {
		Key     string  `json:"key"`
		Integer int     `json:"int"`
		Float   float64 `json:"float"`
	}
	var r respStruct
	bts := resp.Body.Bytes()
	err = json.Unmarshal(bts, &r)
	assert.NoError(t, err, "unmarshal json")
	assert.Equal(t, 3, r.Integer, "integer")
	assert.Equal(t, 3.14, r.Float, "float")
	assert.Equal(t, "value", r.Key, "string")

}

func TestContextExport(t *testing.T) {
	err := setupDb(&Settings{DatabasePath: "./.test_db"})
	assert.NoError(t, err, "setup db")
	defer os.Remove("./.test_db")

	var L = lua.NewState()
	defer L.Close()
	setupLuaContextWithResp("POST", "/", nil, L)

	buck := &interfaces.Bucket{
		BucketID:   uuid.NewV4(),
		BucketName: "testBucket",
	}
	err = dbManager.CreateBucket(buck)
	assert.NoError(t, err, "create bucket")
	file := interfaces.NewFile()
	file.BucketID = buck.BucketID
	file.FileID = uuid.NewV4()
	file.FileName = "testFile"
	file.RawData = []byte("raw data")

	err = dbManager.CreateFile(file)
	assert.NoError(t, err, "create file")

	//todo
	err = L.DoString(`
	ctx():AppExport("export.test.zip")
`)

	assert.NoError(t, err, "app export")
}
