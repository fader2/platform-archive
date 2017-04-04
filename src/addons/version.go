package addons

import (
	"fmt"
	"github.com/hashicorp/go-version"
)

const (
	MIN_ADDON_VERSION = "0.1"
	MAX_ADDON_VERSION = "0.1"
)

func CheckVersion(stringVer string) (bool, error) {
	c, err := version.NewConstraint(fmt.Sprintf(">= %s, <= %s", MIN_ADDON_VERSION, MAX_ADDON_VERSION))
	if err != nil {
		return false, err
	}
	v, err := version.NewVersion(stringVer)
	if err != nil {
		return false, err
	}
	return c.Check(v), nil
}
