// Copyright (c) Fader, IP. All Rights Reserved.
// See LICENSE for license information.
package fs

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"sync/atomic"
	"testing"

	"os"

	"time"

	"github.com/stretchr/testify/assert"
)

func TestWatcher_simple(t *testing.T) {
	dir, err := ioutil.TempDir("", "fader__workspace")
	assert.NoError(t, err)
	time.Sleep(150 * time.Millisecond)

	var counterCreate, counterModify, counterRemove int32

	w := NewFSWatcherWithHook(
		func(op Op, name, oldname string) {
			if op&CreateFileOrFolder == CreateFileOrFolder {
				atomic.AddInt32(&counterCreate, 1)
			}

			if op&ModifyOrCreateFile == ModifyOrCreateFile {
				atomic.AddInt32(&counterModify, 1)
			}

			if op&RemoveFileOrFolder == RemoveFileOrFolder {
				atomic.AddInt32(&counterRemove, 1)
			}

			t.Log(op, oldname, name)
		},
	)
	err = w.Run(
		context.TODO(),
		dir,
	)
	assert.NoError(t, err, "run watcher")

	// create folder
	subdir := filepath.Join(dir, "subdir")
	os.MkdirAll(subdir, 0700)
	time.Sleep(150 * time.Millisecond)

	// create file
	file := filepath.Join(subdir, "1")
	ofile, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0666)
	assert.NoError(t, err, "create file")
	ofile.Write([]byte("text"))
	ofile.Sync()
	ofile.Close()
	time.Sleep(150 * time.Millisecond)

	// change file
	ofile, err = os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0666)
	assert.NoError(t, err, "create file")
	ofile.Write([]byte("new text"))
	ofile.Sync()
	ofile.Close()
	time.Sleep(150 * time.Millisecond)

	// remove file
	os.RemoveAll(file)
	time.Sleep(150 * time.Millisecond)

	// remove folder
	RemoveContents(t, dir)

	time.Sleep(500 * time.Millisecond)

	assert.EqualValues(t, 2, counterCreate)
	assert.EqualValues(t, 1, counterModify)
	assert.EqualValues(t, 1, counterRemove)
}

// http://stackoverflow.com/questions/33450980/golang-remove-all-contents-of-a-directory?answertab=votes#tab-top
func RemoveContents(t *testing.T, dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		t.Logf("Remone %s", name)
	}
	return nil
}
