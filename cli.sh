#!/bin/sh

newaddon() {
	NAME=$1
	ADDONPATH=addons/$1

	mkdir -p $ADDONPATH
	mkdir -p $ADDONPATH/views
	mkdir -p $ADDONPATH/assets
	mkdir -p $ADDONPATH/assets/templates

	cat > $ADDONPATH/Makefile << EOF
build:
	go run assets/generate.go
.PHONY: build
EOF

	cat > $ADDONPATH/views/addons.${NAME}___bootstrap.lua << EOF
print("bootstrap $NAME")
cfg():Set("$NAME", "init OK")
EOF

	cat > $ADDONPATH/assets/generate.go << EOF
// +build ignore

package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/shurcooL/vfsgen"
)

func main() {
	var cwd, _ = os.Getwd()

	templates := http.Dir(filepath.Join(cwd, "views"))

	if err := vfsgen.Generate(templates, vfsgen.Options{
		Filename:     "assets/templates/templates_vfsdata.go",
		PackageName:  "templates",
		BuildTags:    "deploy_build",
		VariableName: "Assets",
	}); err != nil {
		log.Fatalln(err)
	}
}
EOF

	cat > $ADDONPATH/assets/templates/templates.go << EOF
// +build !deploy_build

package templates

import "net/http"

var Assets http.FileSystem
EOF

	cat > $ADDONPATH/addon.go << EOF
package $NAME

import (
	"bytes"
	"io"
	"os"

	"github.com/CloudyKit/jet"
	"github.com/CloudyKit/jet/loaders/httpfs"
	"github.com/fader2/platform/addons"
	"github.com/fader2/platform/config"
	"github.com/fader2/platform/addons/$NAME/assets/templates"
	lua "github.com/yuin/gopher-lua"

	
)

const NAME = "$NAME"

func init() {
	addons.Register(&Addon{})
}

type Addon struct {
}

func (a *Addon) Name() string {
	return NAME
}

func (a *Addon) Bootstrap(cfg *config.Config) error {
	// TODO: bootstrap
	return nil
}

func (a *Addon) LuaModule() lua.LGFunction {
	return func(L *lua.LState) int {
		mod := L.SetFuncs(L.NewTable(), exports)
		L.SetField(mod, "name", lua.LString(a.Name()))

		L.Push(mod)
		return 1
	}
}

func (a *Addon) AssetsLoader() jet.Loader {
	return httpfs.NewLoader(templates.Assets)
}

var exports = map[string]lua.LGFunction{
	"Init": func(L *lua.LState) int {
		f, err := templates.Assets.Open(
			"addons." + NAME + "___bootstrap.lua",
		)
		if os.IsNotExist(err) {
			return 0
		}
		if err != nil {
			L.RaiseError("bootstrap %s: %s", NAME, err)
			return 0
		}
		defer f.Close()

		bootstrap := new(bytes.Buffer)
		io.Copy(bootstrap, f)
		if err := L.DoString(bootstrap.String()); err != nil {
			L.RaiseError("bootstrap %s: load cfg: %s", NAME, err)
		}
		return 0
	},
}

EOF
}


bye() {
	result=$?
	if [ "$result" != "0" ]; then
		echo "Fail"
	fi
	exit $result
}

trap "bye" EXIT

case $1 in
    newaddon) newaddon $2;;
esac