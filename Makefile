include utils.Makefile
include addons.Makefile

run:
	go run platform.go addons.go
.PHONY: run

build: build_addons
	go generate
	go build -tags=deploy_build -o bin/platform platform.go addons.go
.PHONY: build