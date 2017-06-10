package main

import (
	// addons
	"github.com/CloudyKit/jet/loaders/multi"
	_ "github.com/fader2/platform/addons/example"
)

var (
	assets *multi.Multi
)

// +Build ignore
//go:generate go run addons/example/assets/generate.go
