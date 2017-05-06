package addons

import (
	"fmt"
	"github.com/hashicorp/go-version"
)

const (
	MIN_ADDON_VERSION = "0.1"
	MAX_ADDON_VERSION = "0.1"
)

func CheckAddonVersion(addonName, stringVer string) (bool, error) {
	addon, has := Addons[addonName]
	if !has {
		return false, fmt.Errorf("Addon not found %s", addonName)
	}
	current, err := version.NewVersion(stringVer)
	if err != nil {
		return false, err
	}
	expected, err := version.NewVersion(addon.Version())
	if err != nil {
		return false, err
	}
	return expected.Equal(current), nil
}
