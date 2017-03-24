// Copyright (c) Fader, IP. All Rights Reserved.
// See LICENSE for license information.

package fs

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	fsnotify "gopkg.in/fsnotify.v1"
)

var (
	DefaultWorkspace = "FaderWorkspace"
)

type Op uint32

const (
	ModifyOrCreateFile Op = 1 << iota
	CreateFileOrFolder
	RemoveFileOrFolder
	RenameFolder
	RenameFile
)

func (op Op) String() string {
	if op&ModifyOrCreateFile == ModifyOrCreateFile {
		return "modify or create file"
	}
	if op&CreateFileOrFolder == CreateFileOrFolder {
		return "create file or folder"
	}
	if op&RemoveFileOrFolder == RemoveFileOrFolder {
		return "remove file or folder"
	}
	if op&RenameFolder == RenameFolder {
		return "rename folder"
	}
	if op&RenameFile == RenameFile {
		return "rename file"
	}
	return "unknown"
}

type Hook func(opname Op, name string, oldname string)

func NewFSWatcher() *FSWatcher {
	return NewFSWatcherWithHook(
		func(op Op, name, oldname string) {
			if op&(RenameFile|RenameFolder) == RenameFile|RenameFolder {
				log.Println(op, oldname, "->", name)
			} else {
				log.Println(op, name)
			}
		},
	)
}

func NewFSWatcherWithHook(hook Hook) *FSWatcher {
	return &FSWatcher{
		errors: make(chan error, 5),
		hook:   hook,
	}
}

type FSWatcher struct {
	sync.Mutex
	lastEventTime time.Time
	pool          chan fsnotify.Event
	watcher       *fsnotify.Watcher
	errors        chan error
	hook          Hook
}

func (w FSWatcher) Watch(
	ctx context.Context,
	rootPath string,
) (
	err error,
) {
	fmt.Println("Starded eatcher", rootPath)

	w.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return
	}

	ctx, _ = context.WithCancel(ctx)

	if len(rootPath) == 0 {
		rootPath = DefaultWorkspace
	}

	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		if err := os.MkdirAll(rootPath, 0755); err != nil {
			log.Fatal("create workspace folder:", err)
		}
	}

	// root path
	w.watcher.Add(rootPath)

	filepath.Walk(rootPath, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			w.watcher.Add(path)
		}
		return nil
	})

	go func() {
		for {
			select {
			case event := <-w.watcher.Events:
				go w.notify(event)
			case err := <-w.watcher.Errors:
				w.errors <- err
			case <-ctx.Done():
				return
			}
		}
	}()

	return
}

func (w *FSWatcher) notify(event fsnotify.Event) {
	w.Lock()
	defer func() {
		w.lastEventTime = time.Now()
		w.Unlock()
	}()

	if time.Now().Sub(w.lastEventTime) > time.Millisecond*20 {
		go func() {
			ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*20)

			/*
				OSX:
					rename file:
						- RENAME for old file
						- CREATE for new file
						- CHMOD for new file
					rename folder:
						- CREATE for new folder
						- REMOVE|RENAME for old folder
					rename folder:
						- REMOVE|RENAME for old folder
						- CREATE for new folder
					create file:
						- CREATE
						- WRITE not required
					modify file:
						- CHMOD
						- WRITE
					remove file:
						- RENAME
					remove folder:
						- REMOVE|RENAME

			*/

			select {
			case <-ctx.Done():
				poolLen := len(w.pool)
				histogram := make(map[fsnotify.Op]int, poolLen)
				events := make([]fsnotify.Event, poolLen)
				seq := 0
				close(w.pool)
				for event := range w.pool {
					histogram[event.Op]++
					events[seq] = event
					seq++
				}

				// OSX modify or create file
				if (poolLen == 1 && histogram[fsnotify.Write] == 1) ||
					(poolLen == 2 && histogram[fsnotify.Write] == 1 && histogram[fsnotify.Chmod] == 1) {
					go w.hook(ModifyOrCreateFile, events[0].Name, events[0].Name)
					return
				}

				// OSX remove folder or file
				if poolLen == 1 && histogram[fsnotify.Remove|fsnotify.Rename] == 1 ||
					poolLen == 1 && histogram[fsnotify.Rename] == 1 ||
					poolLen == 1 && histogram[fsnotify.Remove] == 1 {
					w.watcher.Remove(events[0].Name)
					go w.hook(RemoveFileOrFolder, events[0].Name, events[0].Name)
					return
				}

				// OSX rename folder
				if poolLen == 2 && histogram[fsnotify.Remove|fsnotify.Rename] == 1 &&
					histogram[fsnotify.Create] == 1 {
					var oldFolder, newFolder string

					if events[1].Op&(fsnotify.Remove|fsnotify.Rename) == fsnotify.Remove|fsnotify.Rename {
						oldFolder = events[1].Name
						newFolder = events[0].Name
					} else if events[0].Op&(fsnotify.Remove|fsnotify.Rename) == fsnotify.Remove|fsnotify.Rename {
						oldFolder = events[0].Name
						newFolder = events[1].Name
					} else {
						w.errors <- fmt.Errorf("not expected order of events")
						return
					}
					go w.hook(RenameFolder, newFolder, oldFolder)
					w.watcher.Remove(oldFolder)
					w.watcher.Add(newFolder)
					return
				}

				// OSX rename file
				if (poolLen == 3 && histogram[fsnotify.Create] == 1 &&
					histogram[fsnotify.Rename] == 1 &&
					histogram[fsnotify.Chmod] == 1) ||
					(poolLen == 2 && histogram[fsnotify.Create] == 1 &&
						histogram[fsnotify.Rename] == 1) {

					var oldFile, newFile string

					if events[1].Op&fsnotify.Rename == fsnotify.Rename {
						oldFile = events[1].Name
						newFile = events[0].Name
					} else if events[0].Op&fsnotify.Rename == fsnotify.Rename {
						oldFile = events[0].Name
						newFile = events[1].Name
					} else {
						w.errors <- fmt.Errorf("not expected order of events")
						return
					}
					go w.hook(RenameFile, newFile, oldFile)
					return
				}

				// OSX create file or folder
				if (poolLen == 1 && histogram[fsnotify.Create] == 1) ||
					(poolLen == 2 &&
						histogram[fsnotify.Create] == 1 &&
						histogram[fsnotify.Write] == 1) {

					name := events[0].Name
					fi, err := os.Stat(name)
					if err != nil {
						w.errors <- fmt.Errorf("error open file %s (%s)", name, err)
						return
					}
					if fi.IsDir() {
						go w.hook(CreateFileOrFolder, name, name)
						w.watcher.Add(name)
					} else {
						go w.hook(CreateFileOrFolder, name, name)
					}

					return
				}
			}
		}()

		w.pool = make(chan fsnotify.Event, 10) // max 3 events per tick
		w.pool <- event
	} else {
		w.pool <- event
	}
}
