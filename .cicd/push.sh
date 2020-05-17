#!/usr/bin/env bash
set -eo pipefail

. .cicd/env.sh

function exists() {
    curl --silent -f -lSL https://index.docker.io/v1/repositories/$1/tags/$2 > /dev/null
}

dmtag() {
  docker manifest annotate $dockerRepo:$vers $dockerRepo:$1-$vers --os linux --arch $1
}

dpush() {
  docker push $dockerRepo:$1-$vers
}

if exists $dockerRepo $vers; then
    echo $dockerRepo:$vers already exists, will not overwrite
    exit 0
else
    docker login -u jettech -p $jettechPassword
    dpush amd64
    dpush arm
    dpush arm64
    docker manifest create $dockerRepo:$vers \
      $dockerRepo:amd64-$vers \
      $dockerRepo:arm-$vers   \
      $dockerRepo:arm64-$vers
    dmtag amd64
    dmtag arm
    dmtag arm64
    docker manifest push $dockerRepo:$vers
fi