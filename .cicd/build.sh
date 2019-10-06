#!/usr/bin/env bash
set -eox pipefail

. .cicd/env.sh

build() {
  docker run --rm \
    -v "$(pwd):/go/src/$mod" \
    -w "/go/src/$mod"  \
    -e GO111MODULE=on  \
    -e CGO_ENABLED=0   \
    -e GOOS=linux      \
    -e GOARCH=$1       \
    golang:1.13-stretch \
      go build -mod=vendor -o kube-webhook-certgen-$1 -ldflags "-X $mod/core.Version=$vers -X $mod/core.BuildTime=$buildTime"
}

build amd64
build arm
build arm64