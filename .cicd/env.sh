#!/usr/bin/env bash
set -eox pipefail

export dockerRepo="jettech/kube-webhook-certgen"
export mod="$(head -n 1 go.mod | cut -f 2 -d ' ')"
export rev=$(git rev-parse HEAD)
export tag=$(git tag --points-at HEAD)
export buildTime=$(date -u +%FT%TZ)

# This will break if there are multiple tags set to the same commit, which is what we want
if [ $tag ]; then
  export vers=$tag
  export isTag=true
else
  export vers=$rev
  export isTag=false
fi

# Azure pipelines requires this invocation to set variables to be available in later steps
# And then you have to retrieve them in a similar interpolation fashion - i.e. they are _not_
# environment variables
echo "##vso[task.setvariable variable=dockerRepo]$dockerRepo"
echo "##vso[task.setvariable variable=mod]$mod"
echo "##vso[task.setvariable variable=rev]$rev"
echo "##vso[task.setvariable variable=vers]$vers"
echo "##vso[task.setvariable variable=isTag]$isTag"
