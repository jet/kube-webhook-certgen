#!/usr/bin/env bash
set -eox pipefail

. .cicd/env.sh

build() {
  ldflags="-X $mod/core.Version=$vers -X $mod/core.BuildTime=$buildTime -buildid= -w -s"
  docker run --rm \
    -v "$(pwd):/go/src/$mod" \
    -w "/go/src/$mod"  \
    -e CGO_ENABLED=0   \
    -e GOOS=linux      \
    -e GOARCH=$1       \
    golang:1.15-buster \
      go build -mod=vendor -o kube-webhook-certgen-$1 -trimpath -ldflags="$ldflags"
}

build amd64
build arm
build arm64
build s390x
build ppc64le
