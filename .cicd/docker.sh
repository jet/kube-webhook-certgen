#!/usr/bin/env bash
set -eox pipefail

. .cicd/env.sh

docker build . -t $dockerRepo:$vers
docker run --rm $dockerRepo:$vers version