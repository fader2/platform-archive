package main

import (
	"github.com/CloudyKit/jet/loaders/multi"
	"github.com/fader2/platform/addons"
	"github.com/fader2/platform/config"

	// addons
	_ "github.com/fader2/platform/addons/foo"
)

var (
	assets *multi.Multi
)

func BootstrapAddons(cfg *config.Config) {
	addons.Bootstrap(cfg)
}
