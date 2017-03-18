package synchronizer

import (
	"encoding/json"
	"fmt"
	"interfaces"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultWorkSpaceName = "./FaderWorkspace"

	LuaExt  = ".lua"
	JSONExt = ".json"

	FilesPermission = 0777
)

func Export(
	db DbManager,
	targetPath string) error {

	if err := cleanWorkspace(targetPath); err != nil {
		return err
	}

	var (
		buckets = map[string]*interfaces.Bucket{}

		err error
	)

	if buckets, err = getBuckets(db); err != nil {
		return err
	}

	if err = db.EachFile(makeExportFileFunc(db, buckets, targetPath)); err != nil {
		return err
	}

	return nil
}

func cleanWorkspace(fpath ...string) error {

	var (
		targetPath = DefaultWorkSpaceName
	)

	if len(fpath) > 0 {
		targetPath = fpath[0]
	}

	if err := os.RemoveAll(targetPath); err != nil {
		return err
	}

	return os.Mkdir(targetPath, FilesPermission)
}

func getBuckets(bm DbManager) (map[string]*interfaces.Bucket, error) {
	var (
		buckets = map[string]*interfaces.Bucket{}
	)

	err := bm.EachBucket(func(b *interfaces.Bucket) error {
		if b == nil {
			return fmt.Errorf("bucket is nil")
		}
		buckets[b.BucketID.String()] = b
		return nil
	})

	return buckets, err
}

func makeExportFileFunc(fileManager DbManager, buckets map[string]*interfaces.Bucket, targetWorkspace string) func(*interfaces.File) error {
	return func(file *interfaces.File) error {
		if file == nil {
			return fmt.Errorf("file is nil")
		}

		var (
			bucket *interfaces.Bucket
			ok     bool
		)

		if bucket, ok = buckets[file.BucketID.String()]; !ok || bucket == nil {
			return fmt.Errorf("Bucket not found for file %s", file.FileName)
		}

		targetPath := filepath.Join(targetWorkspace, bucket.BucketName, file.FileName)

		if err := os.MkdirAll(targetPath, FilesPermission); err != nil && err != os.ErrExist {
			return err
		}

		// write lua script
		if data, has := getFileLuaScript(file); has {
			if err := ioutil.WriteFile(filepath.Join(targetPath, getFileName(file.FileName, "lua")), data, FilesPermission); err != nil {
				return err
			}
		}

		// write raw data
		if data, has := getFileRawData(file); has {
			if err := ioutil.WriteFile(filepath.Join(targetPath, file.FileName), data, FilesPermission); err != nil {
				return err
			}
		}

		// write json data
		if m, has := getFileStructuralData(file); has {
			data, err := json.Marshal(m)
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(filepath.Join(targetPath, getFileName(file.FileName, "json")), data, FilesPermission); err != nil {
				return err
			}
		}

		// write meta data
		if m, has := getFileMetaData(file); has {
			data, err := json.Marshal(m)
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(filepath.Join(targetPath, getFileName(file.FileName, "meta.json")), data, FilesPermission); err != nil {
				return err
			}
		}

		return nil
	}
}

func getFileName(currentName, typ string) string {
	typ = "." + typ
	if typ != "." && strings.HasSuffix(currentName, typ) {
		return currentName
	}

	return currentName + typ
}

func getFileLuaScript(f *interfaces.File) ([]byte, bool) {
	if len(f.LuaScript) > 0 {
		return f.LuaScript, true
	}

	return nil, false
}

func getFileRawData(f *interfaces.File) ([]byte, bool) {
	if len(f.RawData) > 0 {
		return f.RawData, true
	}

	return nil, false
}

func getFileStructuralData(f *interfaces.File) (map[string]interface{}, bool) {
	if len(f.StructuralData) > 0 {
		return f.StructuralData, true
	}
	return nil, false
}

func getFileMetaData(f *interfaces.File) (map[string]interface{}, bool) {
	if len(f.MetaData) > 0 {
		return f.MetaData, true
	}
	return nil, false
}
