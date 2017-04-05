
test:
	GOPATH=${PWD} go test -v \
		-run= ./src/api/...
	GOPATH=${PWD} go test -v \
		-run= ./src/synchronizer/...
	GOPATH=${PWD} go test -v \
		-run= ./src/store/...


.PHONY: test

run:
	GOPATH=${PWD} \
		go run src/cmd/platform/*.go web -watch=true
.PHONY: test


import:
	GOPATH=${PWD} \
		go run src/cmd/platform/*.go import -input=file.zip
.PHONY: test

import64:
	GOPATH=${PWD} \
		go run src/cmd/platform/*.go import64 -input=_fader2.setup.txt
.PHONY: test

import-remote:
	GOPATH=${PWD} \
		go run src/cmd/platform/*.go import -input=https://github.com/zhuharev/fader-sceleton
.PHONY: test


test7:
	 docker run --rm -v "${PWD}":/usr/src/myapp -w /usr/src/myapp golang:1.7 bash -c "make test"
.PHONY: test7

test8:
	 docker run --rm -v "${PWD}":/usr/src/myapp -w /usr/src/myapp golang:1.8 bash -c "make test"
.PHONY: test8