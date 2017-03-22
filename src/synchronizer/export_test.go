package synchronizer

import (
	"archive/zip"
	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	//"os"
	"testing"
	"time"
)

func TestZipExport(t *testing.T) {
	initTestDb("_ex.db")
	defer closeTestDb()

	db, err := bolt.Open("_app.db", FilesPermission, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	defer db.Close()
	assert.Nil(t, err)
	assert.NotNil(t, db)

	err = Export(newDbManager(db), DefaultWorkSpaceName)
	assert.Nil(t, err, err)

	var (
		exWorkSpace = "_ex"
		target      = "./_ex.zip"
	)
	err = writeTestWorkspace(exWorkSpace)
	assert.NoError(t, err, "Create test workspace")

	manager := newDbManager(testDb)

	err = ImportWorkspace(manager, exWorkSpace)
	assert.NoError(t, err, "Import test workspace")

	err = Export(manager, target)
	assert.NoError(t, err, "Export test workspace to zip file")

	zRdr, err := zip.OpenReader(target)
	assert.NoError(t, err, "Open exported zip file")

	for _, file := range zRdr.File {

		_, has := tw["/"+file.Name]
		assert.Equal(t, true, has, "has file in zip file (%s)", file.Name)
	}
	t.Log(tw)

}

func TestExport(t *testing.T) {
	initTestDb("_ex.db")
	defer closeTestDb()

	db, err := bolt.Open("_app.db", FilesPermission, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	defer db.Close()
	assert.Nil(t, err)
	assert.NotNil(t, db)

	err = Export(newDbManager(db), DefaultWorkSpaceName)
	assert.Nil(t, err, err)

	var (
		exWorkSpace = "_ex"
		target      = "_ex1"
	)
	err = writeTestWorkspace(exWorkSpace)
	assert.NoError(t, err, "Create test workspace")

	manager := newDbManager(testDb)

	err = ImportWorkspace(manager, exWorkSpace)
	assert.NoError(t, err, "Import test workspace")

	err = Export(manager, target)
	assert.NoError(t, err, "Export test workspace to zip file")

	zRdr, err := ioutil.ReadDir(target)
	assert.NoError(t, err, "Open exported workspace")

	for _, file := range zRdr {
		if !file.IsDir() {
			t.Log("/" + file.Name())
			_, has := tw[file.Name()]
			assert.Equal(t, true, has, "has file in zip file (%s)", file.Name)
		}
	}

}

/*func TestGetFSFileName(t *testing.T) {
	table := [][]string{
		// filename, filetype, expected name
		[]string{"fname", "lua", "fname.lua"},
		[]string{"fname.lua", "lua", "fname.lua"},

		[]string{"fname.meta", "meta.json", "fname.meta.meta.json"},
		[]string{"fname.meta.json", "meta.json", "fname.meta.json"},
	}

	for _, arr := range table {
		got := getFileName(arr[0], arr[1])
		assert.Equal(t, got, arr[2])
	}
}*/
