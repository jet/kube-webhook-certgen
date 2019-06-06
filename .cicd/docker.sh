#!/usr/bin/env bash
set -eox pipefail

. .cicd/env.sh

rm -rf dockerbuild > /dev/null
mkdir dockerbuild
cp Dockerfile dockerbuild
mv kube-webhook-certgen dockerbuild

docker build -f dockerbuild/Dockerfile dockerbuild -t $dockerRepo:$vers
docker run --rm $dockerRepo:$vers version