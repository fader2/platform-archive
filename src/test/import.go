package test

import (
	"bytes"
	"github.com/boltdb/bolt"
	"interfaces"
	"io"
	"io/ioutil"
	//"log"
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

func ImportWorkspace(db *bolt.DB, workspaceRoot string) error {

	buckets, err := ioutil.ReadDir(workspaceRoot)
	if err != nil {
		return err
	}
	for _, bucket := range buckets {
		if err = ImportBucket(db, workspaceRoot, bucket.Name()); err != nil {
			return err
		}
	}
	return nil
}

func ImportBucket(db *bolt.DB, workspaceRoot, bucketName string) (err error) {
	var (
		bucketPath = filepath.Join(workspaceRoot, bucketName)
	)

	if _, _, err = createOrGetBucket(db, bucketName); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(bucketPath)
	if err != nil {
		return err
	}
	for _, file := range files {
		if err = ImportFsVirtualFile(db, workspaceRoot, bucketName, file.Name()); err != nil {
			return err
		}
	}

	return nil
}

// ImportFsVirtualFile set folder items (file_name,config.json,meta.json,structural.json) as file in boltdb
// it return erro if file not exists en db
func ImportFsVirtualFile(db *bolt.DB, workspaceRoot, bucketName, fileName string) (err error) {

	// todo check file folder contains max 4 files

	var (
		fm   = boltStore.NewFileManager(db)
		file *interfaces.File

		buffer *bytes.Buffer = pool.Get().(*bytes.Buffer)
	)

	{
		// return buffer to pool buffer
		defer func() {
			pool.Put(buffer)
		}()
	}

	if file, _, err = createOrGetFile(db, bucketName, fileName); err != nil {
		return
	}

	{
		// empty file data
		// it need for deleting some data case
		// all values will be overwire from fs
		file.LuaScript = nil
		file.RawData = nil
		file.MetaData = make(map[string]interface{})
		file.StructuralData = make(map[string]interface{})
	}

	filePath := filepath.Join(workspaceRoot, bucketName, fileName)
	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		return
	}
	for _, dataFile := range files {
		buffer.Reset()
		f, err := os.OpenFile(filepath.Join(filePath, dataFile.Name()), os.O_RDONLY, FilesPermission)
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

		tused := detectUsedType(dataFile.Name())
		err = setDataToFile(file, tused, buffer.Bytes())
		if err != nil {
			return err
		}
	}

	err = fm.UpdateFileFrom(file, interfaces.DataFile)
	return
}

func ImportFsDataFile(db *bolt.DB, workspaceRoot, bucketName, fileName, dataName string) (err error) {

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
		used = detectUsedType(dataName)
		err = setDataToFile(file, used, buffer.Bytes())
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

func createOrGetBucket(db *bolt.DB, bucketName string) (bucket *interfaces.Bucket, isNew bool, err error) {
	var (
		bucketManager = boltStore.NewBucketManager(db)
	)

	if bucket, err = bucketManager.FindBucketByName(bucketName, interfaces.FullBucket); err != nil {
		if err != interfaces.ErrNotFound {
			return
		}
	} else {
		return
	}

	isNew = true

	bucket.BucketID = uuid.NewV4()
	bucket.BucketName = bucketName

	err = bucketManager.CreateBucket(bucket)
	return
}

func createOrGetFile(db *bolt.DB, bucketName, fileName string) (file *interfaces.File, isNew bool, err error) {

	var (
		fileManager   = boltStore.NewFileManager(db)
		bucketManager = boltStore.NewBucketManager(db)

		bucket *interfaces.Bucket
	)

	if file, err = fileManager.FindFileByName(bucketName, fileName, interfaces.FullFile); err != nil && err != interfaces.ErrNotFound {
		return
	} else if err == interfaces.ErrNotFound {
		isNew = true
		bucket, err = bucketManager.FindBucketByName(bucketName, interfaces.FullBucket)
		if err != nil {
			return
		}
	} else {
		isNew = false
		return
	}

	if file == nil {
		file = interfaces.NewFile()
	}
	file.FileID = uuid.NewV4()
	file.BucketID = bucket.BucketID
	file.FileName = fileName

	err = fileManager.CreateFile(file)
	return
}
