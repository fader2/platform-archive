package synchronizer

import (
	"archive/zip"
	"fmt"
	"interfaces"
	"io/ioutil"
	"strings"
	//"log"
	"os"
	"path/filepath"

	uuid "github.com/satori/go.uuid"
)

func ImportWorkspace(db DbManager, workspaceRoot string) error {

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

func ImportWorkspaceZip(db DbManager, workspaceRoot string) error {

	var (
		buckets = make(map[string]*interfaces.Bucket)
	)

	r, err := zip.OpenReader(workspaceRoot)
	if err != nil {
		return err
	}

	defer func() {
		r.Close()
	}()

	for _, f := range r.File {
		arr := strings.SplitN(f.Name, string(os.PathSeparator), 3)
		if len(arr) != 3 {
			fmt.Printf("Skip file %s:\n", f.Name)
			continue
		}

		var (
			bucketName string = arr[0]
			fileName   string = arr[1]
			dataName   string = arr[2]
		)

		// todo. create other way to kreatin buckets
		if bucket, ok := buckets[bucketName]; !ok {
			bucket, _, err = createOrGetBucket(db, bucketName)
			if err != nil {
				return err
			}
			buckets[bucket.BucketName] = bucket
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		bts, err := ioutil.ReadAll(rc)
		if err != nil {
			return err
		}
		rc.Close()

		err = importFsDataFileData(db, bucketName, fileName, dataName, bts)
		if err != nil {
			return err
		}
	}
	return nil
}

func ImportBucket(db DbManager, workspaceRoot, bucketName string) (err error) {
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
func ImportFsVirtualFile(db DbManager, workspaceRoot, bucketName, fileName string) (err error) {

	// todo check file folder contains max 4 files

	var (
		file *interfaces.File

		filePath = filepath.Join(workspaceRoot, bucketName, fileName)
	)

	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		return
	}
	//todo ?? need delete file if it empty folder?
	//if len(files) == 0 {
	//	return deleteFileByName(db, bucketName, fileName)
	//}

	// ioutil.ReadDir(dirname)
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

	for _, dataFile := range files {
		f, err := os.OpenFile(filepath.Join(filePath, dataFile.Name()), os.O_RDONLY, FilesPermission)
		if err != nil {
			return err
		}

		// todo test file is no folder
		bts, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		err = f.Close()
		if err != nil {
			return err
		}

		tused, err := detectUsedType(fileName, dataFile.Name())
		if err != nil {
			return err
		}
		err = setDataToFile(file, tused, bts)
		if err != nil {
			return err
		}
	}

	err = db.UpdateFileFrom(file, interfaces.DataFile)
	return
}

func ImportFsDataFile(db DbManager, workspaceRoot, bucketName, fileName, dataName string) (err error) {

	var (
		filePath = filepath.Join(workspaceRoot, bucketName, fileName, dataName)
		bts      []byte
	)

	// copy file data to buffer
	{
		f, err := os.OpenFile(filePath, os.O_RDONLY, FilesPermission)
		if err != nil {
			if !strings.HasSuffix(err.Error(), "no such file or directory") {
				return err
			}
		} else {

			// todo test file is no folder

			// put data in buffer here
			bts, err = ioutil.ReadAll(f)
			if err != nil {
				return err
			}
			err = f.Close()
			if err != nil {
				return err
			}
		}
	}

	return importFsDataFileData(db, bucketName, fileName, dataName, bts)
}

func importFsDataFileData(db DbManager, bucketName, fileName, dataName string, data []byte) (err error) {
	var (
		has    bool
		used   interfaces.DataUsed
		file   *interfaces.File
		bucket *interfaces.Bucket
	)

	if file, err = db.FindFileByName(bucketName, fileName, interfaces.FullFile); err != nil && err != interfaces.ErrNotFound {
		return
	} else if err == interfaces.ErrNotFound {
		has = false
		bucket, _, err = createOrGetBucket(db, bucketName)
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
		used, err = detectUsedType(fileName, dataName)
		if err != nil {
			return err
		}
		err = setDataToFile(file, used, data)
		if err != nil {
			return err
		}
	}

	// put data to database
	if has {
		err = db.UpdateFileFrom(file, used)
	} else {
		file.FileID = uuid.NewV4()
		file.BucketID = bucket.BucketID
		file.FileName = fileName
		//file.LuaScript = []byte{}
		//file.ContentType = "text/plain"

		err = db.CreateFile(file)
		if err != nil {
			return err
		}
	}

	return nil
}

func createOrGetBucket(db DbManager, bucketName string) (bucket *interfaces.Bucket, isNew bool, err error) {
	if bucket, err = db.FindBucketByName(bucketName, interfaces.FullBucket); err != nil {
		if err != interfaces.ErrNotFound {
			return
		}
	} else {
		return
	}

	isNew = true

	bucket = interfaces.NewBucket()
	bucket.BucketID = uuid.NewV4()
	bucket.BucketName = bucketName

	err = db.CreateBucket(bucket)
	return
}

func createOrGetFile(db DbManager, bucketName, fileName string) (file *interfaces.File, isNew bool, err error) {

	var (
		bucket *interfaces.Bucket
	)

	if file, err = db.FindFileByName(bucketName, fileName, interfaces.FullFile); err != nil && err != interfaces.ErrNotFound {
		return
	} else if err == interfaces.ErrNotFound {
		isNew = true
		bucket, _, err = createOrGetBucket(db, bucketName)
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

	err = db.CreateFile(file)
	return
}

func deleteFileByName(db DbManager, bucketName, fileName string) error {
	file, err := db.FindFileByName(bucketName, fileName, 0)
	if err != nil {
		return err
	}
	err = db.DeleteFile(file.FileID)
	if err != nil {
		return err
	}
	return nil
}
