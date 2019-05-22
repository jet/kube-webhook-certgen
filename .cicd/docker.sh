#!/usr/bin/env bash
set -eo pipefail

. .cicd/env.sh

docker build . -t $dockerRepo:$dockerTag
docker run --rm $dockerRepo:$dockerTag version