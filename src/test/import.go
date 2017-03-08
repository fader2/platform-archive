package test

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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

func ImportFile(fileDir string) error {

	var (
		bucketPath, fileName = filepath.Split(fileDir)
		_, bucketName        = filepath.Split(bucketPath)
	)

	log.Println(bucketPath, fileName, bucketName)

	walkFileDir := func(path string, info os.FileInfo, err error) error {
		return nil
	}

	if err := filepath.Walk(fileDir, walkFileDir); err != nil {
		return err
	}

	return nil
}
