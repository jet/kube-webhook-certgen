#!/usr/bin/env bash
set -eox pipefail

. .cicd/env.sh

command="mkdir -p .cover; \
  go install github.com/jstemmer/go-junit-report@latest; \
  go install github.com/axw/gocov/gocov@latest;          \
  go install github.com/AlekSi/gocov-xml@latest;         \
  go mod vendor;        \
  git config --global --add safe.directory /go/src/github.com/jet/kube-webhook-certgen; \
  go test -v ./... -coverprofile .cover/coverage.txt -covermode count 2>&1 > .cover/testresults.txt; \
  cat .cover/testresults.txt | go-junit-report > .cover/TEST-ALL.xml; \
  gocov convert .cover/coverage.txt > .cover/coverage.json;           \
  gocov-xml < .cover/coverage.json > .cover/coverage.xml;             \
  git reset --hard HEAD; git clean -fdX"

docker run --rm \
  -v "$(pwd):/go/src/$mod" \
  -w "/go/src/$mod"  \
  -e CGO_ENABLED=0   \
  -e GOOS=linux      \
  -e GOARCH=amd64    \
  golang:1.20-buster \
    /bin/bash -c "$command"