#!/usr/bin/env bash
set -eox pipefail

export dockerRepo="jet/kube-webhook-certgen"
export mod="$(head -n 1 go.mod | cut -f 2 -d ' ')"
export rev=$(git rev-parse HEAD)
export tag=$(git tag --points-at HEAD)
export buildTime=$(date -u +%FT%TZ)

# This will break if there are multiple tags set to the same commit, which is what we want
if [ $tag ]; then
  export vers=$tag
  export dockerTag=$vers
else
  export vers=$rev
  export dockerTag=latest
fi