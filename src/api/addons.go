package api

import (
	"addons"
	"addons/example"
	//"fmt"
	"github.com/BurntSushi/toml"
	//"github.com/hashicorp/go-version"
	"io"
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

func CheckCompatibility(fpath string) (bool, error) {
	var as addonsSetting
	_, err := toml.DecodeFile(fpath, &as)
	if err != nil {
		return false, err
	}

	//for _, v := range as.Addons {
	//	version.NewConstraint(fmt.Sprintf(">= %d, < %d"))
	//}
	return false, nil
}

func CheckCompatibilityFromReader(rdr io.Reader) (bool, error) {
	var as addonsSetting
	_, err := toml.DecodeReader(rdr, &as)
	if err != nil {
		return false, err
	}
	return false, nil
}
