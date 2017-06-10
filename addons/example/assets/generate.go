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
	
	templates := http.Dir(filepath.Join(cwd, "addons/example/views"))

	if err := vfsgen.Generate(templates, vfsgen.Options{
		Filename:     "addons/example/assets/templates/templates_vfsdata.go",
		PackageName:  "templates",
		BuildTags:    "deploy_build",
		VariableName: "Assets",
	}); err != nil {
		log.Fatalln(err)
	}
}