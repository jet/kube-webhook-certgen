#!/usr/bin/env bash
set -eo pipefail

. .cicd/env.sh

function exists() {
    curl --silent -f -lSL https://index.docker.io/v1/repositories/$1/tags/$2 > /dev/null
}

dmtag() {
  docker manifest annotate $dockerRepo:$vers $dockerRepo:$vers-$1 --os linux --arch $1
}


if exists $dockerRepo $vers; then
    echo $dockerRepo:$vers already exists, will not overwrite
    exit 1
else
    docker login -u jettech -p $jettechPassword
    docker manifest create $dockerRepo:$vers \
      $dockerRepo:$vers-amd64 \
      $dockerRepo:$vers-arm   \
      $dockerRepo:$vers-arm64 --amend
    dmtag amd64
    dmtag arm
    dmtag arm64

    docker manifest push $dockerRepo:$vers
fi