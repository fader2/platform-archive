package test

import (
	"fmt"
	"fs"
)

type Synchronizer struct {
	watcher  *fs.FSWatcher
	rootPath string
}

func NewSynchronizer(rootPath string) (*Synchronizer, error) {
	s := &Synchronizer{
		rootPath: rootPath,
	}
	s.watcher = fs.NewFSWatcherWithHook(s.MakeWatchFunc())
	return s, nil
}

func (s *Synchronizer) Start() error {
	s.watcher.Run(nil, s.rootPath)
	return nil
}

func (s *Synchronizer) MakeWatchFunc() func(opname fs.Op, name string, oldname string) {
	return func(opname fs.Op, name string, oldname string) {
		fmt.Println(opname, name, oldname)
	}
}
