#!/usr/bin/env bash
set -eox pipefail

. .cicd/env.sh

docker run --rm \
  -v "$(pwd):/go/src/$mod" \
  -w "/go/src/$mod"  \
  -e GO111MODULE=on  \
  -e CGO_ENABLED=0   \
  -e GOOS=linux      \
  -e GOARCH=amd64    \
  golang:1.12-stretch \
    go build -mod=vendor -ldflags "-X $mod/core.Version=$vers -X $mod/core.BuildTime=$buildTime"