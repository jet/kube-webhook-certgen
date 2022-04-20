#!/usr/bin/env bash
set -eox pipefail

. .cicd/env.sh

rm -rf dockerbuild > /dev/null
mkdir dockerbuild
cp Dockerfile dockerbuild

dbuild() {
  cp kube-webhook-certgen-$1 dockerbuild/kube-webhook-certgen
  docker build -f dockerbuild/Dockerfile dockerbuild -t docker.io/$dockerRepo:$1-$vers
}

dbuild amd64
dbuild arm
dbuild arm64
dbuild s390x
dbuild ppc64le

docker run $dockerRepo:amd64-$vers version
