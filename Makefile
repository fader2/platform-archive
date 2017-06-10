run:
	go run platform.go addons.go
.PHONY: run

build:
	go generate
	go build -tags=deploy_build -o bin/platform platform.go addons.go
.PHONY: build