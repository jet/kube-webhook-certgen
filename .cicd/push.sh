#!/usr/bin/env bash
set -eo pipefail

. .cicd/env.sh

function exists() {
    curl --silent -f -lSL https://index.docker.io/v1/repositories/$1/tags/$2 > /dev/null
}

if exists $dockerRepo $vers; then
    echo $dockerRepo:$vers already exists, will not overwrite
    exit 1
else
    docker login -u jettech -p $jettechPassword
    docker push $dockerRepo:$vers
fi