package test

import (
	"bytes"
	"encoding/json"
	"github.com/boltdb/bolt"
	"interfaces"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	boltStore "store/boltdb"
	"sync"

	uuid "github.com/satori/go.uuid"
)

var (
	pool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	mu sync.Mutex
)

func ImportWorkspace(workspacePath string) error {

	files, err := ioutil.ReadDir(workspacePath)
	if err != nil {
		return err
	}
	for _, v := range files {
		log.Println(v.Name())
	}
	return nil
}

func ImportFsFile(db *bolt.DB, workspaceRoot, bucketName, fileName, dataName string) (err error) {

	mu.Lock()

	var (
		fileManager   = boltStore.NewFileManager(db)
		bucketManager = boltStore.NewBucketManager(db)

		filePath = filepath.Join(workspaceRoot, bucketName, fileName, dataName)
		has      bool
		used     interfaces.DataUsed
		file     *interfaces.File
		bucket   *interfaces.Bucket

		buffer *bytes.Buffer = pool.Get().(*bytes.Buffer)
	)

	defer func() {
		mu.Unlock()
		pool.Put(buffer)
	}()

	// copy file data to buffer
	{
		// empty buffer
		buffer.Reset()

		f, err := os.OpenFile(filePath, os.O_RDONLY, FilesPermission)
		if err != nil {
			return err
		}

		// todo test file is no folder

		// put data in buffer here
		_, err = io.Copy(buffer, f)
		if err != nil {
			return err
		}
		err = f.Close()
		if err != nil {
			return err
		}
	}

	if file, err = fileManager.FindFileByName(bucketName, fileName, interfaces.FullFile); err != nil && err != interfaces.ErrNotFound {
		return
	} else if err == interfaces.ErrNotFound {
		has = false
		bucket, err = bucketManager.FindBucketByName(bucketName, interfaces.FullBucket)
		if err != nil {
			return err
		}
	} else {
		has = true
	}

	if file == nil {
		file = interfaces.NewFile()
	}

	// detect content type
	{
		switch dataName {
		case "script.lua":
			used = interfaces.LuaScript
			file.LuaScript = buffer.Bytes()
		case "meta.json":
			used = interfaces.MetaData
			m := map[string]interface{}{}
			err = json.Unmarshal(buffer.Bytes(), &m)
			file.MetaData = m
		case "structural_data.json":
			used = interfaces.StructuralData
			m := map[string]interface{}{}
			err = json.Unmarshal(buffer.Bytes(), &m)
			file.StructuralData = m
		default:
			used = interfaces.RawData
			file.RawData = buffer.Bytes()
		}

		if err != nil {
			return err
		}
	}

	// put data to database
	if has {
		err = fileManager.UpdateFileFrom(file, used)
	} else {
		file.FileID = uuid.NewV4()
		file.BucketID = bucket.BucketID
		file.FileName = fileName
		//file.LuaScript = []byte{}
		//file.ContentType = "text/plain"

		err = fileManager.CreateFile(file)
	}
	if err != nil {
		return err
	}

	return nil
}
