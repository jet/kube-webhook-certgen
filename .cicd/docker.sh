#!/usr/bin/env bash
set -eox pipefail

if [ $tag ]; then
  vers=$tag
  dockerTag=$vers
else
  vers=$rev
  dockerTag=latest
fi

docker build . -t $dockerRepo:$dockerTag
docker run --rm $dockerRepo:$dockerTag version
echo "Created image with $dockerRepo:$dockerTag"