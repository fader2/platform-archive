package test

import (
	"context"
	"fmt"
	"fs"
	"github.com/boltdb/bolt"
	"strings"
)

type Synchronizer struct {
	watcher  *fs.FSWatcher
	conn     *bolt.DB
	rootPath string
}

func NewSynchronizer(rootPath string, conn *bolt.DB) (*Synchronizer, error) {
	s := &Synchronizer{
		rootPath: rootPath,

		conn: conn,
	}
	s.watcher = fs.NewFSWatcherWithHook(s.MakeWatchFunc())
	return s, nil
}

func (s *Synchronizer) Start() error {
	s.watcher.Run(context.TODO(), s.rootPath)
	return nil
}

func (s *Synchronizer) MakeWatchFunc() func(opname fs.Op, name string, oldname string) {
	return func(opname fs.Op, name string, oldname string) {
		fmt.Println(opname, name, oldname)
		// todo test abs path
		arr := strings.SplitN(name, "/", 4)
		if len(arr) == 2 {
			fmt.Println("Path is bucket, skip")
		} else if len(arr) == 3 {
			fmt.Println("Path is file folder, skip")
		} else if len(arr) == 4 {
			fmt.Printf("we need update file, with %s filename, %s bucketname\n", arr[2], arr[1])

			bucketName := arr[1]
			fileName := arr[2]
			dataName := arr[3]

			err := ImportFsFile(s.conn, s.rootPath, bucketName, fileName, dataName)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
