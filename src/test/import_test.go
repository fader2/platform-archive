package test

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"interfaces"
	"io/ioutil"
	"os"
	"path/filepath"
	boltStore "store/boltdb"
	"strings"
	"testing"
	"time"

	"github.com/boltdb/bolt"
)

var (
	testDb *bolt.DB
)

func initTestDb(dbPaths ...string) {
	var e error
	dbPath := "_app.db"
	if len(dbPaths) > 0 {
		dbPath = dbPaths[0]
	}
	testDb, e = bolt.Open(dbPath, FilesPermission, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	if e != nil {
		panic(e)
	}
}

func closeTestDb() {
	p := testDb.Path()
	testDb.Close()
	os.RemoveAll(p)
}

func testFile(file *interfaces.File, used interfaces.DataUsed, data string) error {
	switch used {
	case interfaces.StructuralData:
		bts, err := json.Marshal(file.StructuralData)
		if err != nil {
			return err
		}
		if string(bts) == data {
			return nil
		} else {
			return fmt.Errorf("%s not equal %s", string(bts), data)
		}
	case interfaces.MetaData:
		bts, err := json.Marshal(file.MetaData)
		if err != nil {
			return err
		}
		if string(bts) == data {
			return nil
		} else {
			return fmt.Errorf("%s not equal %s", string(bts), data)
		}
	case interfaces.RawData:
		if string(file.RawData) == data {
			return nil
		} else {
			return fmt.Errorf("%s not equal %s", string(file.RawData), data)
		}
	case interfaces.LuaScript:
		if string(file.LuaScript) == data {
			return nil
		} else {
			return fmt.Errorf("%s not equal %s", string(file.LuaScript), data)
		}
	}
	return nil
}

var tw = map[string]struct {
	isDir bool
	used  interfaces.DataUsed
	data  string
}{
	"/bucket1/file/structural_data.json": {false, interfaces.StructuralData, `{"testkey":"testdata"}`},
	"/bucket2/emptyfile":                 {true, interfaces.FullFile, ""},
	"/bucket1/filename/filename":         {false, interfaces.RawData, "data"},
	"/bucket3/file.html/file.html":       {false, interfaces.RawData, "<!html>"},
	"/bucket3/file/meta.json":            {false, interfaces.MetaData, `{"testmeta":"testmeta"}`},
}

func createFileInWorkspace(workspaceRoot, path string, isDir bool, data string) (bool, error) {
	filePath := filepath.Join(workspaceRoot, path)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return true, nil
	} else if !isDir {
		dir, _ := filepath.Split(filePath)
		err := os.MkdirAll(dir, FilesPermission)
		if err != nil {
			return false, err
		}
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, FilesPermission)
		if err != nil {
			return false, err
		}
		defer f.Close()
		_, err = f.WriteString(data)
		if err != nil {
			return false, err
		}
	} else if isDir {
		err = os.MkdirAll(filePath, FilesPermission)
		if err != nil {
			return false, err
		}
	}
	return false, nil
}

func writeTestWorkspace(workspaceRoot string) error {
	for filePath, file := range tw {
		if ok, err := createFileInWorkspace(workspaceRoot, filePath, file.isDir, file.data); ok || err != nil {
			return fmt.Errorf("File exists or an error %v", err)
		}
	}
	return nil
}

func TestCreateOrGetBucket(t *testing.T) {
	initTestDb()
	defer closeTestDb()

	var (
		buckName = "buck"
	)

	// test create
	bucket, isNew, err := createOrGetBucket(testDb, buckName)
	assert.NoError(t, err, "createing new bucket")
	assert.Equal(t, true, isNew, "is new bucket")
	assert.NotNil(t, bucket)

	// test get
	bucket1, isNew1, err1 := createOrGetBucket(testDb, buckName)
	assert.NoError(t, err1, "get existing bucket")
	assert.Equal(t, false, isNew1, "is old bucket")
	assert.NotNil(t, bucket1)
}

func TestCreateOrGetFile(t *testing.T) {
	initTestDb()
	defer closeTestDb()

	var (
		buckName = "buck"
		fileName = "file"
	)

	_, _, err := createOrGetBucket(testDb, buckName)
	assert.NoError(t, err, "createing new bucket")

	// test create
	file, isNew, err := createOrGetFile(testDb, buckName, fileName)
	assert.NoError(t, err, "createing new file")
	assert.Equal(t, true, isNew, "is new file")
	assert.NotNil(t, file)

	// test get
	file1, isNew1, err1 := createOrGetFile(testDb, buckName, fileName)
	assert.NoError(t, err1, "get existing file")
	assert.Equal(t, false, isNew1, "is old file")
	assert.NotNil(t, file1)
}

func TestImportWorkspace(t *testing.T) {

	var (
		workspaceRoot = "_tw"
		testDbPath    = "_testDb"

		fileManager *boltStore.FileManager
	)

	initTestDb(testDbPath)
	defer closeTestDb()

	fileManager = boltStore.NewFileManager(testDb)

	err := os.RemoveAll(workspaceRoot)
	assert.NoError(t, err, "Error on clean test workspace")

	err = writeTestWorkspace(workspaceRoot)
	assert.NoError(t, err, "Error on creating test workspace")

	err = ImportWorkspace(testDb, workspaceRoot)
	assert.NoError(t, err, "Error test import workspace in empty database")

	for filePath, fileMeta := range tw {

		arr := strings.Split(filePath, "/")
		ifile, err := fileManager.FindFileByName(arr[1], arr[2], interfaces.FullFile)
		assert.NoError(t, err, "Get must existing file from database %s/%s", arr[1], arr[2])

		assert.NoError(t, testFile(ifile, fileMeta.used, fileMeta.data), "file (%s) data test", filePath)
	}
}

func TestImportFsDataFile(t *testing.T) {
	initTestDb()
	defer closeTestDb()

	var (
		bucketName = "еtestbucketname"
		fileName   = "fileName"
		//dataName   = "dataName"
		data = []byte("data")

		uses = []struct {
			used interfaces.DataUsed
			name string
			data []byte
		}{
			{interfaces.MetaData, MetaDataFileName, []byte(`{"alo":"da"}`)},
			{interfaces.StructuralData, StructuralDataFileName, []byte(`{"str":"str"}`)},
			{interfaces.RawData, fileName, data},
			{interfaces.LuaScript, LuaScriptDataFileName, []byte(`-- script`)},
		}
	)

	fPath := filepath.Join(DefaultWorkSpaceName, bucketName, fileName)
	err := os.MkdirAll(fPath, 0777)
	assert.NoError(t, err, "mkdir for test workspace")
	for _, v := range uses {
		fP := filepath.Join(fPath, v.name)
		err = ioutil.WriteFile(fP, v.data, FilesPermission)
		assert.NoError(t, err, "write file (%s)", v.name)

		err = ImportFsDataFile(testDb, DefaultWorkSpaceName, bucketName, fileName, v.name)
		assert.NoError(t, err, "Import data file (%s)", v.name)

		err = os.Remove(fP)
		assert.NoError(t, err, "Remove data file (%s)", v.name)

		err = ImportFsDataFile(testDb, DefaultWorkSpaceName, bucketName, fileName, v.name)
		assert.NoError(t, err, "Import not existing file (%s)", v.name)
		if err != nil {
			panic(err)
		}
	}

}

func TestImportFsDataFileData(t *testing.T) {
	initTestDb()
	defer closeTestDb()

	var (
		bucketName = "bucketname"
		fileName   = "fileName"
		dataName   = "dataName"
		data       = []byte("data")
	)

	err := importFsDataFileData(testDb, bucketName, fileName, dataName, data)
	assert.NoError(t, err, "import fs data file")

}

// create file
func TestImportNewFile(t *testing.T) {
	initTestDb()
	defer closeTestDb()

	var (
		fileName   = "testFile"
		bucketName = "ex1"
		testdata   = "testdata"
	)

	fileFolder := filepath.Join(DefaultWorkSpaceName, bucketName, fileName)
	filePath := filepath.Join(fileFolder, fileName)

	os.Remove(filePath)
	os.MkdirAll(fileFolder, FilesPermission)

	f, e := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, FilesPermission)
	assert.Nil(t, e)

	_, e = f.WriteString(testdata)
	assert.Nil(t, e)
	f.Close()

	e = ImportFsVirtualFile(testDb, DefaultWorkSpaceName, bucketName, fileName)
	assert.NoError(t, e, "import virtual file")

	e = ImportFsDataFile(testDb, DefaultWorkSpaceName, bucketName, fileName, fileName)
	assert.Nil(t, e, eToStr(e))

	fm := boltStore.NewFileManager(testDb)
	file, e := fm.FindFileByName(bucketName, fileName, interfaces.FullFile)
	assert.Nil(t, e, eToStr(e))
	assert.Equal(t, testdata, string(file.RawData))

	e = fm.DeleteFile(file.FileID)
	assert.NoError(t, e)
}

func eToStr(e error) string {
	if e != nil {
		return e.Error()
	}
	return ""
}

// update
func TestImportExistingFile(t *testing.T) {
	initTestDb()
	defer closeTestDb()

	var (
		fileName = "profile.html"
	)

	fileFolder := filepath.Join(DefaultWorkSpaceName, "ex1", fileName)
	filePath := filepath.Join(fileFolder, fileName)

	os.Remove(filePath)
	os.Mkdir(fileFolder, FilesPermission)

	f, e := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, FilesPermission)
	assert.Nil(t, e)

	_, e = f.WriteString("testdata")
	assert.Nil(t, e)

	e = ImportFsDataFile(testDb, DefaultWorkSpaceName, "ex1", fileName, fileName)
	assert.Nil(t, e, eToStr(e))

	fm := boltStore.NewFileManager(testDb)
	file, e := fm.FindFileByName("ex1", fileName, interfaces.FullFile)
	assert.Nil(t, e, eToStr(e))
	assert.Equal(t, "testdata", string(file.RawData))

	file.RawData = []byte{}

	e = fm.UpdateFileFrom(file, interfaces.RawData)
	assert.NoError(t, e)
}
