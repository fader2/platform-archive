
test:
	GOPATH=${PWD} go test -v \
		-run= ./src/store/...
#	GOPATH=${PWD} go test -v \
		-run= ./src/fs/...
.PHONY: test

run:
	GOPATH=${PWD} \
		go run src/cmd/platform/main.go --watch=true
.PHONY: test

test7:
	 docker run --rm -v "${PWD}":/usr/src/myapp -w /usr/src/myapp golang:1.7 bash -c "make test"
.PHONY: test7

test8:
	 docker run --rm -v "${PWD}":/usr/src/myapp -w /usr/src/myapp golang:1.8 bash -c "make test"
.PHONY: test8