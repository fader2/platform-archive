package test

import (
	"encoding/json"
	"fmt"
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
		err = json.Unmarshal(data, &m)
		file.MetaData = m
	case interfaces.StructuralData:
		m := map[string]interface{}{}
		err = json.Unmarshal(data, &m)
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

func detectUsedType(dataName string) interfaces.DataUsed {
	switch dataName {
	case "script.lua":
		return interfaces.LuaScript
	case "meta.json":
		return interfaces.MetaData
	case "structural_data.json":
		return interfaces.StructuralData
	default:
		return interfaces.RawData
	}
}
