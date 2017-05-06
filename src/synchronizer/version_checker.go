package synchronizer

import (
	"github.com/BurntSushi/toml"
	"io"
)

type VersionChecker interface {
	Check(io.Reader) error
	FileName() string
	WritePackageInfo(io.Writer) error
}

type DefaultVersionChecker struct {
	Addons  map[string]string     `toml:"addons"`
	checker func(io.Reader) error `toml:"-"`
}

func NewVersionChecker(addons map[string]string, checker func(io.Reader) error) *DefaultVersionChecker {
	return &DefaultVersionChecker{Addons: addons, checker: checker}
}

func (vc *DefaultVersionChecker) Check(rdr io.Reader) error {
	return vc.checker(rdr)
}

func (vc *DefaultVersionChecker) FileName() string {
	return "package.toml"
}

// todo move this method from here
func (vc *DefaultVersionChecker) WritePackageInfo(w io.Writer) error {
	wr := toml.NewEncoder(w)
	err := wr.Encode(vc)
	if err != nil {
		return err
	}
	return nil
}
