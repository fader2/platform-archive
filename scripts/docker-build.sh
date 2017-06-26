#!/bin/bash

docker run --rm \
    -v "$PWD":/go/src/github.com/fader2/platform \
    -w /go/src/github.com/fader2/platform \
    golang:1.8-onbuild \
    make build