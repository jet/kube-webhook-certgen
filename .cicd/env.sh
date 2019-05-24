#!/usr/bin/env bash
set -eo pipefail

export dockerRepo="jet/kube-webhook-certgen"
export mod="$(head -n 1 go.mod | cut -f 2 -d ' ')"

# Get version if there is a current git tag, otherwise use the commit
export rev=$(git rev-parse HEAD)
export tag=$(git tag --points-at HEAD)

# Get date
export buildTime=$(date -u +%FT%TZ)

if [ $tag ]; then
  export vers=$tag
  export dockerTag=$vers
else
  export vers=$rev
  export dockerTag=latest
fi

echo "##vso[task.setvariable variable=dockerRepo]$dockerRepo"
echo "##vso[task.setvariable variable=mod]$mod"
echo "##vso[task.setvariable variable=rev]$rev"
echo "##vso[task.setvariable variable=vers]$vers"
echo "##vso[task.setvariable variable=dockerTag]$dockerTag"