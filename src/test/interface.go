package test

import (
	"interfaces"
)

type DbManager interface {
	interfaces.FileManager
	interfaces.BucketManager
}

type FSManager interface {
	DbManager

	WorkspaceRoot() string

	// if targetPath has ".zip extension"
	// it will be exported as zip archive
	ExportWorkspace(targetPath string) error

	// if file has extension ".zip", it will be unzipped as workspace
	Import(pathToFile string) error

	// todo if fsmanager no has Watch method
	// MakeWatchHook() sf.Hook
}
