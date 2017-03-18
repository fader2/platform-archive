package synchronizer

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"time"
	//"io/ioutil"
	//"os"
	//"path/filepath"
	"strings"
	"testing"
	//"time"
)

/*func TestSync(t *testing.T) {
	initTestDb()
	defer closeTestDb()

	s, _ := NewSynchronizer(DefaultWorkSpaceName, newDbManager(testDb))

	fileName := "testfiledata"
	folder := filepath.Join(DefaultWorkSpaceName, "ex1", fileName)
	filePath := filepath.Join(folder, fileName)
	os.MkdirAll(folder, FilesPermission)
	time.Sleep(1 * time.Second)
	err := ioutil.WriteFile(filePath, []byte("lol"), FilesPermission)
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)
}*/

func TestTree(t *testing.T) {
	tr := newTree()
	item := fsItem{Path: "a", ModTime: time.Now(), Size: 2, Hash: "aaa"}
	tr["a"] = item

	buf := bytes.NewBuffer(nil)
	err := tr.Encode(buf)
	assert.NoError(t, err)

	tr = newTree()
	err = tr.Decode(strings.NewReader(buf.String()))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(tr))
	assert.Equal(t, tr["a"], item)
}
