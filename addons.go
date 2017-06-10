package main

import (
	"github.com/CloudyKit/jet/loaders/multi"
	// addons
	_ "github.com/fader2/platform/addons/bar"
	_ "github.com/fader2/platform/addons/foo"
)

var (
	assets *multi.Multi
)

// +Build ignore
//go:generate make -C addons/foo build
//go:generate make -C addons/bar build
