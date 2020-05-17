#!/usr/bin/env bash
set -eox pipefail

. .cicd/env.sh

command="go test -v ./... -coverprofile coverage.txt -covermode count 2>&1 > testresults.txt; \
  go get github.com/jstemmer/go-junit-report; \
  go get github.com/axw/gocov/gocov;          \
  go get github.com/AlekSi/gocov-xml;         \
  go get github.com/matm/gocov-html;          \
  go mod vendor;                              \
  cat testresults.txt | go-junit-report > TEST-ALL.xml; \
  gocov convert coverage.txt > coverage.json;           \
  gocov-xml < coverage.json > coverage.xml;             \
  mkdir coverage || true                                \
  gocov-html < coverage.json > coverage/index.html"

docker run --rm \
  -v "$(pwd):/go/src/$mod" \
  -w "/go/src/$mod"  \
  -e CGO_ENABLED=0   \
  -e GOOS=linux      \
  -e GOARCH=amd64    \
  golang:1.14-stretch \
    /bin/bash -c "$command"