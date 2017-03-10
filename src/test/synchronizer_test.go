package test

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSync(t *testing.T) {
	initTestDb()
	defer closeTestDb()

	s, _ := NewSynchronizer(DefaultWorkSpaceName, testDb)
	go func() {
		err := s.Start()
		assert.NoError(t, err)
	}()

	time.Sleep(1 * time.Second)

	fileName := "testfiledata"
	folder := filepath.Join(DefaultWorkSpaceName, "ex1", fileName)
	filePath := filepath.Join(folder, fileName)
	os.MkdirAll(folder, FilesPermission)
	time.Sleep(1 * time.Second)
	err := ioutil.WriteFile(filePath, []byte("lol"), FilesPermission)
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)
}
