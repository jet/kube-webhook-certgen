#!/usr/bin/env bash
set -eox pipefail

. .cicd/env.sh

command="go get -u github.com/jstemmer/go-junit-report; \
go test -mod=vendor -v ./... 2>&1 | go-junit-report > TEST-ALL.xml"

docker run --rm \
  -v "$(pwd):/go/src/$mod" \
  -w "/go/src/$mod"  \
  -e GO111MODULE=on  \
  -e CGO_ENABLED=0   \
  -e GOOS=linux      \
  -e GOARCH=amd64    \
  golang:1.12-stretch \
    /bin/bash -c "$command"