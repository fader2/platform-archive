package test

import (
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
