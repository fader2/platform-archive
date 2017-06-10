package main

import (
	"github.com/CloudyKit/jet/loaders/multi"

	// addons
	_ "github.com/fader2/platform/addons/example"
)

var (
	assets *multi.Multi
)

// +Build ignore
//go:generate make -C addons/example build
