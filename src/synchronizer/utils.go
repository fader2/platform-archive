package synchronizer

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"net/url"
	store "store/boltdb"
	"strings"
	//"github.com/boltdb/bolt"
	"interfaces"
	"io/ioutil"
	"net/http"
)

// https://gist.github.com/rayrutjes/db9b9ea8e02255d62ce2
func DetectContentType(buffer []byte) string {
	// Always returns a valid content-type and "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType
}

func dirContent(dirname string, onlyDirs bool) (names []string, err error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	for _, fileInfo := range files {
		if onlyDirs == fileInfo.IsDir() {
			names = append(names, fileInfo.Name())
		}
	}

	return
}

func dirDirs(dirname string) ([]string, error) {
	return dirContent(dirname, true)
}

func dirFiles(dirname string) ([]string, error) {
	return dirContent(dirname, false)
}

/* import */
func HasFile(fileManager interfaces.FileManager, bucketName, fileName string) (bool, error) {
	if _, err := fileManager.FindFileByName(bucketName, fileName, interfaces.FullFile); err == nil {
		return true, nil
	} else if err == interfaces.ErrNotFound {
		return false, nil
	} else {
		return false, err
	}
}

func setDataToFile(file *interfaces.File, used interfaces.DataUsed, data []byte) (err error) {
	switch used {
	case interfaces.LuaScript:
		file.LuaScript = data
	case interfaces.MetaData:
		m := map[string]interface{}{}
		if len(data) > 0 {
			err = json.Unmarshal(data, &m)
		}
		file.MetaData = m
	case interfaces.StructuralData:
		m := map[string]interface{}{}
		if len(data) > 0 {
			err = json.Unmarshal(data, &m)
		}
		file.StructuralData = m
	case interfaces.RawData:
		file.RawData = data
	default:
		return fmt.Errorf("Err used")
	}

	if err != nil {
		return err
	}
	return nil
}

func detectUsedType(fileName, dataName string) (interfaces.DataUsed, error) {
	switch dataName {
	case LuaScriptDataFileName:
		return interfaces.LuaScript, nil
	case MetaDataFileName:
		return interfaces.MetaData, nil
	case StructuralDataFileName:
		return interfaces.StructuralData, nil
	case fileName:
		return interfaces.RawData, nil
	default:
		return 0, fmt.Errorf("Unknown file name (file:%s, data: %s)", fileName, dataName)
	}
}

func newDbManager(db *bolt.DB) DbManager {
	fm := store.NewFileManager(db)
	bm := store.NewBucketManager(db)
	return struct {
		interfaces.FileManager
		interfaces.BucketManager
		interfaces.BucketImportManager
		interfaces.FileImportManager
	}{
		fm, bm, bm, fm,
	}
}

func dbHasData(db DbManager) (bool, error) {
	cnt := 0
	err := db.EachBucket(func(_ *interfaces.Bucket) error {
		cnt++
		return nil
	})
	return cnt != 0, err
}

func fileHasData(f *interfaces.File, used interfaces.DataUsed) bool {
	if f == nil {
		return false
	}
	switch used {
	case interfaces.RawData:
		return len(f.RawData) != 0
	case interfaces.LuaScript:
		return len(f.LuaScript) != 0
	case interfaces.StructuralData:
		return f.StructuralData != nil && len(f.StructuralData) != 0
	case interfaces.MetaData:
		return f.MetaData != nil && len(f.MetaData) != 0
	default:
		panic("error DataUsed called for fileHasData (allowed: RawData, LuaScript, StructuralData, MetaData), got " + f.FileName)
	}
}

func getDownloadLink(remote string) (string, error) {
	u, err := url.Parse(remote)
	if err != nil {
		return "", err
	}
	// todo
	// https://github.com/ZloDeeV/gpsgame-android/archive/master.zip
	if u.Hostname() == "github.com" {
		return remote + "/archive/master.zip", nil
	}
	return "", nil
}

func isGithubArchive(files []*zip.File) bool {
	for _, file := range files {
		if strings.HasSuffix(file.Name, "-master/") {
			fmt.Println("GITHUB ARRCHIBFE")
			return true
		}
	}
	return false
}
