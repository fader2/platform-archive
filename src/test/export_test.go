package test

import (
	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestExport(t *testing.T) {
	db, err := bolt.Open("/home/god/go/src/github.com/inpime/fader/_app.db", FilesPermission, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	defer db.Close()
	assert.Nil(t, err)
	assert.NotNil(t, db)

	err = Export(db, DefaultWorkSpaceName)
	assert.Nil(t, err, err)

}

func TestGetFSFileName(t *testing.T) {
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
}
