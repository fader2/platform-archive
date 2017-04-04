package synchronizer

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	//"interfaces"
	"os"
	"path/filepath"
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
	tr := NewFSTree()
	item := fsItem{Path: "a", ModTime: time.Now(), Size: 2, Hash: "aaa"}
	tr.items["a"] = item

	buf := bytes.NewBuffer(nil)
	err := tr.Encode(buf)
	assert.NoError(t, err)

	tr = NewFSTree()
	err = tr.Decode(strings.NewReader(buf.String()))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(tr.items))
	assert.Equal(t, tr.items["a"], item)
}

func TestWatcher(t *testing.T) {
	var (
		testPath = ".watchertest"
		testFile = "testFile"
	)
	defer os.RemoveAll(testPath)
	err := os.Mkdir(testPath, FilesPermission)
	assert.NoError(t, err, "mkdir test path")
	tr, err := NewFSTreeFromFs(testPath)
	assert.NoError(t, err, "new fs tree")

	_, err = os.Create(filepath.Join(testPath, testFile))
	assert.NoError(t, err, "create file")

	tr2, err := NewFSTreeFromFs(testPath)
	assert.NoError(t, err, "new fs tree")

	ops := tr.Calculate(tr2)
	assert.Equal(t, 1, len(ops), "calculate")
	assert.Equal(t, int(create), int(ops[0].Op))
}
