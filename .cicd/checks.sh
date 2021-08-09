#!/usr/bin/env bash
set -eox pipefail

. .cicd/env.sh

command="go fmt ./... && git diff --exit-code; go vet ./..."

docker run --rm \
  -v "$(pwd):/go/src/$mod" \
  -w "/go/src/$mod"  \
  -e CGO_ENABLED=0   \
  -e GOOS=linux      \
  -e GOARCH=amd64    \
  golang:1.16-buster \
    /bin/bash -c "$command"
