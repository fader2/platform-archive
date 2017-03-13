package test

import (
	"context"
	"fmt"
	"fs"
	"github.com/boltdb/bolt"
	"strings"

	boltStore "store/boltdb"
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
		switch opname {
		case fs.ModifyOrCreateFile:
		case fs.CreateFileOrFolder:
			err := s.handleUpdateOrCreate(name, oldname)
			if err != nil {
				//todo handle it
				fmt.Println(err)
			}
		case fs.RemoveFileOrFolder:
			err := s.handleRemoveFile(name)
			if err != nil {
				//todo handle it
				fmt.Println(err)
			}
		case fs.RenameFolder:
		case fs.RenameFile:
		}

	}
}

func (s *Synchronizer) handleRemoveFile(name string) error {
	var (
		arr = strings.SplitN(name, "/", 4)

		//workspace  = arr[0]
		bucketName string
		fileName   string
		//dataName   string
	)

	if len(arr) == 1 {
		return fmt.Errorf("%s", "Workspace deleted")
	} else if len(arr) == 2 {
		return fmt.Errorf("Bucket deleted")
	} else if len(arr) == 3 {
		bucketName = arr[1]
		fileName = arr[2]

		// file deleted
		fileManager := boltStore.NewFileManager(s.conn)
		file, err := fileManager.FindFileByName(bucketName, fileName, 0)
		if err != nil {
			return err
		}
		err = fileManager.DeleteFile(file.FileID)
		if err != nil {
			return err
		}
	} else if len(arr) == 4 {
		// single change, but import all data in file
		// todo handle single data update
		bucketName = arr[1]
		fileName = arr[2]
		//dataName = arr[3]

		fmt.Printf("we need update file, with %s filename, %s bucketname\n", arr[2], arr[1])

		bucketName = arr[1]
		fileName = arr[2]
		//dataName = arr[3]

		err := ImportFsVirtualFile(s.conn, s.rootPath, bucketName, fileName)
		if err != nil {
			return err
		}

	}
	return nil
}

func (s *Synchronizer) handleUpdateOrCreate(name, oldname string) error {
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

		err := ImportFsDataFile(s.conn, s.rootPath, bucketName, fileName, dataName)
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}
