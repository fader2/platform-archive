package test

import (
	"github.com/stretchr/testify/assert"
	"interfaces"
	"os"
	"path/filepath"
	boltStore "store/boltdb"
	"testing"
	"time"

	"github.com/boltdb/bolt"
)

var (
	testDb *bolt.DB
)

func initTestDb() {
	var e error
	testDb, e = bolt.Open("/home/god/go/src/github.com/inpime/fader/_app.db", FilesPermission, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	if e != nil {
		panic(e)
	}
}

func closeTestDb() {
	testDb.Close()
}

// create file
func TestImportNewFile(t *testing.T) {
	initTestDb()
	defer closeTestDb()

	var (
		fileName = "testFile"
	)

	fileFolder := filepath.Join(DefaultWorkSpaceName, "ex1", fileName)
	filePath := filepath.Join(fileFolder, fileName)

	os.Remove(filePath)
	os.Mkdir(fileFolder, FilesPermission)

	f, e := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, FilesPermission)
	assert.Nil(t, e)

	_, e = f.WriteString("testdata")
	assert.Nil(t, e)

	e = ImportFsDataFile(testDb, DefaultWorkSpaceName, "ex1", "testFile", "testFile")
	assert.Nil(t, e, eToStr(e))

	fm := boltStore.NewFileManager(testDb)
	file, e := fm.FindFileByName("ex1", "testFile", interfaces.FullFile)
	assert.Nil(t, e, eToStr(e))
	assert.Equal(t, "testdata", string(file.RawData))

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
