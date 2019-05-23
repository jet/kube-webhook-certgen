#!/usr/bin/env bash
set -eo pipefail

if [ $tag ]; then
  vers=$tag
  dockerTag=$vers
else
  vers=$rev
  dockerTag=latest
fi

docker build . -t jettech/kube-webhook-certgen:$dockerTag
docker run --rm $dockerRepo:$dockerTag version
echo "Created image with $dockerRepo:$dockerTag"