package api

import (
	"addons"
	"addons/example"
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
