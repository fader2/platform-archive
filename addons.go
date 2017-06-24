package main

import (
	"github.com/CloudyKit/jet"
	"github.com/CloudyKit/jet/loaders/multi"
	"github.com/fader2/platform/addons"
	"github.com/fader2/platform/config"

	// addons
	_ "github.com/fader2/platform/addons/boltdb"
	_ "github.com/fader2/platform/addons/foo"
	_ "github.com/fader2/platform/addons/tpls"
)

var (
	assets *multi.Multi
)

func BootstrapAddons(cfg *config.Config, tpls *jet.Set) error {
	return addons.Bootstrap(cfg, tpls)
}
