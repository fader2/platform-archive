package synchronizer

import (
	"io"
)

type VersionChecker interface {
	Check(io.Reader) error
	FileName() string
	WritePackageInfo(io.Writer) error
	PackageFileName() string
}
