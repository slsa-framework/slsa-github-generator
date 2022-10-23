#!/bin/bash

# Verify that Action calls that are pinned (not locally called)
# use the same referemnce as the relealse tag.
# See reasoning in ./github/actions/README.md

set -euo pipefail

results=$(
    find .github/workflows/ -name '*.yaml' -o -name '*.yml' \
    | xargs grep -P "slsa-framework/slsa-github-generator/.github/actions/.*@(?!$RELEASE_TAG)" \
    || true
)
if [[ "$results" != "" ]]; then
    echo "Some Actions are not referenced via the correct release tag \"$RELEASE_TAG\""
    echo "$results"
    exit 1
fi

