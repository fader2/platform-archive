include utils.Makefile
include addons.Makefile

run:
	go run \
		-tags=deploy_build \
		-race \
		platform.go addons.go
.PHONY: run

prebuild: build_addons
	go generate
.PHONY: prebuild

build:
	go build -tags=deploy_build -o bin/platform platform.go addons.go
.PHONY: build

genproto:
	cd ./objects && protoc -I=. -I=${GOPATH}/src --gogoslick_out=. *.proto
	cd ./objects && protoc -I=. -I=${GOPATH}/src --gogoslick_out=. *.proto
.PHONY: genproto

test:
	go test -v \
		 ./objects
	go test -v \
		-timeout 1s \
		 ./addons/boltdb
.PHONY: test