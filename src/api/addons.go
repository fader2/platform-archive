package api

import (
	"addons"
	"addons/example"
	"fmt"
	"github.com/BurntSushi/toml"
	"io"
	"os"
)

func SetupAddons() error {
	addons.Addons[example.NAME] = example.NewAddon()
	addons.Addons[ADDONS_BASIC_NAME] = NewBasicAddon()

	return nil
}

func AddonList() (map[string]string, error) {
	res := make(map[string]string)
	for _, addon := range addons.Addons {
		res[addon.Name()] = addon.Version()
	}
	return res, nil
}

type addonsSetting struct {
	Addons map[string]string `toml:"addons"`
}

func CheckCompatibility(fpath string) error {
	f, err := os.OpenFile(fpath, os.O_RDONLY, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	err = CheckCompatibilityFromReader(f)
	if err != nil {
		return err
	}

	return nil
}

func CheckCompatibilityFromReader(rdr io.Reader) error {
	var as = addonsSetting{Addons: make(map[string]string)}
	_, err := toml.DecodeReader(rdr, &as)
	if err != nil {
		return err
	}

	for name, ver := range as.Addons {
		ok, err := addons.CheckAddonVersion(name, ver)
		if err != nil {
			return err
		} else if !ok {
			cur := addons.Addons[name]
			return fmt.Errorf("Addon version compare error: for addon %s expected version %s, got %s", name, cur.Version(), ver)
		}
	}

	return nil
}

type versionChecker struct {
}

func NewVersionChecker() *versionChecker {
	return new(versionChecker)
}

func (vc *versionChecker) Check(rdr io.Reader) error {
	return CheckCompatibilityFromReader(rdr)
}

func (vc *versionChecker) FileName() string {
	return DefaultPackageFileName
}

// todo move this method from here
func (vc *versionChecker) WritePackageInfo(w io.Writer) error {
	sttngs := addonsSetting{Addons: make(map[string]string)}
	for _, v := range addons.Addons {
		sttngs.Addons[v.Name()] = v.Version()
	}
	sttngs.Addons[DefaultAppName] = VERSION

	wr := toml.NewEncoder(w)
	err := wr.Encode(sttngs)
	if err != nil {
		return err
	}
	return nil
}

func (vc *versionChecker) PackageFileName() string {
	return DefaultPackageFileName
}
