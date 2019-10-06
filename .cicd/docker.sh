#!/usr/bin/env bash
set -eox pipefail

. .cicd/env.sh

rm -rf dockerbuild > /dev/null
mkdir dockerbuild
cp Dockerfile dockerbuild

dbuild() {
  cp kube-webhook-certgen-$1 dockerbuild/kube-webhook-certgen
  docker build -f dockerbuild/Dockerfile dockerbuild -t $dockerRepo:$vers-$1
}

dbuild amd64
dbuild arm
dbuild arm64

