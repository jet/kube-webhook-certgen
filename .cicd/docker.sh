#!/usr/bin/env bash
set -eox pipefail

docker build . -t $dockerRepo:$dockerTag
docker run --rm $dockerRepo:$dockerTag version
echo "Created image with $dockerRepo:$dockerTag"