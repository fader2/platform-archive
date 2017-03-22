package synchronizer

import (
	"encoding/json"
	"fmt"
	"fs"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
	//boltStore "store/boltdb"
)

const (
	StructuralDataFileName = "structural_data.json"
	MetaDataFileName       = "meta.json"
	LuaScriptDataFileName  = "script.lua"

	SyncInfoFileName = ".fader_index"
)

type Synchronizer struct {
	inited bool

	dbManager     DbManager
	workSpacePath string

	tree tree
}

func NewSynchronizer(workSpacePath string, dbManager DbManager) (*Synchronizer, error) {
	s := &Synchronizer{
		workSpacePath: workSpacePath,

		dbManager: dbManager,

		// tree will be initialized in init
	}
	return s, nil
}

// Init create workspace path if not exists
//    File system         DB
//
//    empty         <-    has data
//    has data      ->    empty
//    has data      ->    has data
//    empty               empty    // do nothing
//
func (s *Synchronizer) Init() error {
	if s.inited {
		return nil
	}

	var (
		workSpaceHasData         bool
		workSpaceHasSyncInfoFile bool
		dbHasBuckets             bool
	)

	if _, err := os.Stat(s.workSpacePath); err == nil {
		files, err := ioutil.ReadDir(s.workSpacePath)
		if err != nil {
			return err
		}
		if len(files) > 0 {
			workSpaceHasData = true
		}
	} else if os.IsNotExist(err) {
		err = os.MkdirAll(s.workSpacePath, FilesPermission)
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(filepath.Join(s.workSpacePath, SyncInfoFileName)); err == nil {
		workSpaceHasSyncInfoFile = true
	}

	if workSpaceHasSyncInfoFile {
		s.tree = make(tree)
		f, err := os.OpenFile(filepath.Join(s.workSpacePath, SyncInfoFileName), os.O_RDONLY, FilesPermission)
		if err != nil {
			return err
		}
		err = s.tree.Decode(f)
		if err != nil {
			return err
		}
		err = f.Close()
		if err != nil {
			return err
		}
	} else {
		s.tree = make(tree)
	}

	dbHasBuckets, err := dbHasData(s.dbManager)
	if err != nil {
		return err
	}

	// return if both data stores empty
	if !dbHasBuckets && !workSpaceHasData {
		fmt.Println("[Sync] both workspace and database empty")
		return nil
	}

	if dbHasBuckets && !workSpaceHasData {
		fmt.Println("[Sync] export data from db to fs")
		err := Export(s.dbManager, s.workSpacePath)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("[Sync] import data from fs to db")
		err := ImportWorkspace(s.dbManager, s.workSpacePath)
		if err != nil {
			return err
		}
	}

	s.inited = true
	return nil
}

func (s *Synchronizer) Sync(filePaths ...string) error {
	if err := s.Init(); err != nil {
		return err
	}

	return nil
}

// todo remove from this package and remove fs dependence
func (s *Synchronizer) MakeWatchFunc() func(opname fs.Op, name string, oldname string) {
	return func(opname fs.Op, name string, oldname string) {
		fmt.Println("[Watch Event]", opname, name, oldname)
		switch opname {
		case fs.CreateFileOrFolder,
			fs.ModifyOrCreateFile:
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
		err := deleteFileByName(s.dbManager, bucketName, fileName)
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

		err := ImportFsVirtualFile(s.dbManager, s.workSpacePath, bucketName, fileName)
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

		err := ImportFsDataFile(s.dbManager, s.workSpacePath, bucketName, fileName, dataName)
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

/* tree */

type tree map[string]fsItem

type fsItem struct {
	Path    string
	Size    int64
	ModTime time.Time
	Hash    string
}

func newTree() tree {
	return make(tree)
}

func (t tree) Encode(writer io.Writer) error {
	enc := json.NewEncoder(writer)
	err := enc.Encode(t)
	return err
}

func (t *tree) Decode(reader io.Reader) error {
	var tt = *t
	dec := json.NewDecoder(reader)
	err := dec.Decode(t)
	t = &tt
	return err
}
