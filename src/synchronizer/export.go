package synchronizer

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"interfaces"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultWorkSpaceName = "FaderWorkspace"

	LuaExt  = ".lua"
	JSONExt = ".json"

	FilesPermission = 0777
)

func Export(
	db DbManager,
	targetPath string, checkers ...VersionChecker) (err error) {

	var (
		isZip   = strings.HasSuffix(targetPath, ".zip")
		zipFile *zip.Writer
		buckets = map[string]*interfaces.Bucket{}
	)

	if err := cleanWorkspace(targetPath); err != nil {
		return err
	}

	if buckets, err = getBuckets(db); err != nil {
		return err
	}

	if isZip {
		if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
			fmt.Println(err)
			return err
		}

		zfile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, FilesPermission)
		if err != nil {
			return err
		}
		zipFile = zip.NewWriter(zfile)
		defer func() {
			// close file
			zfile.Close()
		}()
	}

	if err = db.EachFile(makeExportFileFunc(db, buckets, targetPath, isZip, zipFile)); err != nil {
		return err
	}

	/* write package.toml */
	if len(checkers) > 0 && checkers[0] != nil {
		var wr io.Writer
		if isZip {
			zwr, err := zipFile.Create(checkers[0].FileName())
			if err != nil {
				return err
			}
			wr = zwr
		} else {
			f, err := os.OpenFile(filepath.Join(targetPath, checkers[0].FileName()), os.O_CREATE|os.O_WRONLY, FilesPermission)
			if err != nil {
				return err
			}
			defer f.Close()
			wr = f
		}
		err = checkers[0].WritePackageInfo(wr)
		if err != nil {
			return err
		}
	}

	if isZip {
		err = zipFile.Close()
		if err != nil {
			return err
		}
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

func makeExportFileFunc(fileManager DbManager, buckets map[string]*interfaces.Bucket, targetWorkspace string, isZip bool, zipFile *zip.Writer) func(*interfaces.File) error {
	return func(file *interfaces.File) (err error) {
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

		targetPath := bucket.BucketName
		if !isZip {
			targetPath = filepath.Dir(filepath.Join(targetWorkspace, targetPath, file.FileName))
			log.Println(targetPath)
			if err := os.MkdirAll(targetPath, FilesPermission); err != nil && !os.IsExist(err) {
				return err
			}
		}

		us := []interfaces.DataUsed{
			interfaces.RawData,
			interfaces.LuaScript,
			interfaces.StructuralData,
			interfaces.MetaData,
		}

		for _, used := range us {
			if !fileHasData(file, used) {
				continue
			}
			var (
				wr       io.Writer
				_, fName = filepath.Split(getFileName(file.FileName, used))
				fname    = filepath.Join(targetPath, fName)
			)
			log.Println("Write", fname, file.FileName)
			if isZip {
				wr, err = zipFile.Create(filepath.Join(targetPath, getFileName(file.FileName, used)))
				if err != nil {
					return err
				}
			} else {
				wr, err = os.OpenFile(fname, os.O_CREATE|os.O_WRONLY, FilesPermission)
				if err != nil {
					return err
				}
			}
			err = writeFile(file, used, wr)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func writeFile(f *interfaces.File, used interfaces.DataUsed, w io.Writer) (err error) {
	// write lua script
	switch used {
	case interfaces.RawData:
		_, err = w.Write(f.RawData)
		return err
	case interfaces.LuaScript:
		_, err = w.Write(f.LuaScript)
		return err
	case interfaces.MetaData:
		bts, err := json.Marshal(f.MetaData)
		if err != nil {
			return err
		}
		_, err = w.Write(bts)
		return err
	case interfaces.StructuralData:
		bts, err := json.Marshal(f.StructuralData)
		if err != nil {
			return err
		}
		_, err = w.Write(bts)
		return err
	default:
		return fmt.Errorf("Mad used type: %s", used)
	}
}

func getFileName(currentName string, typ interfaces.DataUsed) string {
	switch typ {
	case interfaces.RawData:
		return currentName
	case interfaces.LuaScript:
		return currentName + ".lua"
	case interfaces.StructuralData:
		return currentName + ".json"
	case interfaces.MetaData:
		return currentName + ".meta.json"
	default:
		panic("error DataUsed called for getFileName (allowed: RawData, LuaScript, StructuralData, MetaData), got " + currentName)
	}
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
