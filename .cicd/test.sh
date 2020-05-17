#!/usr/bin/env bash
set -eox pipefail

. .cicd/env.sh

command="mkdir .cover || true; \
  go test -v ./... -coverprofile .cover/coverage.txt -covermode count 2>&1 > .cover/testresults.txt; \
  go get github.com/jstemmer/go-junit-report; \
  go get github.com/axw/gocov/gocov;          \
  go get github.com/AlekSi/gocov-xml;         \
  go mod vendor;                              \
  cat .cover/testresults.txt | go-junit-report > .cover/TEST-ALL.xml; \
  gocov convert .cover/coverage.txt > .cover/coverage.json;           \
  gocov-xml < .cover/coverage.json > .cover/coverage.xml;"

# go get github.com/matm/gocov-html;          \
# gocov-html < .cover/coverage.json > .cover/index.html

docker run --rm \
  -v "$(pwd):/go/src/$mod" \
  -w "/go/src/$mod"  \
  -e CGO_ENABLED=0   \
  -e GOOS=linux      \
  -e GOARCH=amd64    \
  golang:1.14-stretch \
    /bin/bash -c "$command"