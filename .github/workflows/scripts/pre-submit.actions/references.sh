#!/bin/bash

# Verify that Action calls that are pinned (not locally called)
# use the same referemnce as the relealse tag.
# See reasoning in ./github/actions/README.md

set -euo pipefail

cd __THIS_REPO__

results=$(
    find .github/workflows/ -maxdepth 1 -name '*.yaml' -o -name '*.yml' -type f -print0 \
    | xargs -0 grep -P "slsa-framework/slsa-github-generator/.github/actions/.*@(?!main)" \
    || true
)
if [[ "$results" != "" ]]; then
    echo "Some Actions are not referenced at main"
    echo "$results"
    exit 1
fi

results=$(
    find .github/workflows/ -maxdepth 1 -name 'builder_*.yaml' -o -name 'builder_*.yml' -type f -print0 \
    | xargs -0 grep -P "slsa-framework/slsa-github-generator/.github/workflows/.*@(?!main)" \
    || true
)
if [[ "$results" != "" ]]; then
    echo "Some builder workflows are not referenced at main"
    echo "$results"
    exit 1
fi

