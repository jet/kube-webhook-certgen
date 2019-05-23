#!/usr/bin/env bash
set -eo pipefail

if [ $tag ]; then
  vers=$tag
  dockerTag=$vers
else
  vers=$rev
  dockerTag=latest
fi

docker run --rm \
  -v "$(pwd):/go/src/$mod" \
  -w "/go/src/$mod"  \
  -e GO111MODULE=on  \
  -e CGO_ENABLED=0   \
  -e GOOS=linux      \
  -e GOARCH=amd64    \
  golang:1.12-stretch \
    /bin/bash -c \
    "go get -u github.com/jstemmer/go-junit-report; go test -v ./... 2>&1 | go-junit-report > TEST-ALL.xml"