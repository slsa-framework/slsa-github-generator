#!/bin/bash -eu
#
# Copyright 2023 SLSA Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

# Caller sets the following:
#BUILDER_REPOSITORY="slsa-framework/slsa-github-generator"
#BUILDER_RELEASE_BINARY="builder-linux-amd64"
#
#VERIFIER_REPOSITORY="slsa-framework/slsa-verifier"
#VERIFIER_RELEASE="v13.0.10"
#VERIFIER_RELEASE_BINARY="slsa-verifier-linux-amd64"
#VERIFIER_RELEASE_BINARY_SHA256="89fbcba9aed67d5146ea99946c7e4e5a80e3767871f0e3ffcd0b582134efd010"

PREFIX="refs/tags/"

# Extract version.
if [[ "$BUILDER_REF" != "$PREFIX"* ]]; then
    echo "Invalid ref: $BUILDER_REF. Expected ref of the form refs/tags/vX.Y.Z"
    exit 2
fi

builder_tag="${BUILDER_REF#"$PREFIX"}"

if [[ "$builder_tag" == "$(echo -n "$builder_tag" | grep -P '^[a-f\d]{40}$')" ]]; then
    echo "Builder referenced by hash: $builder_tag"
    echo "Resolving..."

    release_tag=""

    # List the releases and find the corresponding hash.
    release_list=$(gh release -R "$BUILDER_REPOSITORY" -L 50 list)
    while read -r line; do
        tag=$(echo "$line" | cut -f1)
        branch=$(gh release -R "$BUILDER_REPOSITORY" view "$tag" --json targetCommitish --jq '.targetCommitish')
        if [[ "$branch" != "main" ]]; then
            continue
        fi
        commit=$(gh api /repos/"$BUILDER_REPOSITORY"/git/ref/tags/"$tag" | jq -r '.object.sha')
        if [[ "$commit" == "$builder_tag" ]]; then
            release_tag="$tag"
            echo "Found tag $builder_tag match at tag $tag and commit $commit"
            break
        fi
    done <<<"$release_list"

    if [[ -z "$release_tag" ]]; then
        echo "Tag not found for $builder_tag"
        exit 3
    fi

    builder_tag="$release_tag"
fi

if [[ "$builder_tag" != "$(echo -n "$builder_tag" | grep -oe '^v[1-9]\+\.[0-9]\+\.[0-9]\+\(-rc\.[0-9]\+\)\?$')" ]]; then
    echo "Invalid builder version: $builder_tag. Expected version of the form vX.Y.Z(-rc.A)"
    echo "For details see https://github.com/slsa-framework/slsa-github-generator/blob/main/README.md#referencing-slsa-builders-and-generators"
    exit 7
fi

echo "Builder version: $builder_tag"

echo "BUILDER_REPOSITORY: $BUILDER_REPOSITORY"

# Fetch the release binary and provenance.
gh release -R "$BUILDER_REPOSITORY" download "$builder_tag" -p "$BUILDER_RELEASE_BINARY*" || exit 10

# Fetch the verifier at the right hash.
gh release -R "$VERIFIER_REPOSITORY" download "$VERIFIER_RELEASE" -p "$VERIFIER_RELEASE_BINARY" || exit 11
computed_hash=$(sha256sum "$VERIFIER_RELEASE_BINARY" | awk '{print $1}')
echo "verifier hash computed is $computed_hash"
echo "$VERIFIER_RELEASE_BINARY_SHA256 $VERIFIER_RELEASE_BINARY" | sha256sum --strict --check --status || exit 4
echo "verifier hash verification has passed"

# If this is a pre-release, set SLSA_VERIFIER_TESTING
pre_release=$(echo "${builder_tag#"v"}" | cut -s -d '-' -f2)
if [ "${pre_release}" != "" ]; then
    export SLSA_VERIFIER_TESTING="true"
fi

# Verify the provenance of the builder.
chmod a+x "$VERIFIER_RELEASE_BINARY"
./"$VERIFIER_RELEASE_BINARY" verify-artifact \
    --source-branch "main" \
    --source-tag "$builder_tag" \
    --provenance-path "$BUILDER_RELEASE_BINARY.intoto.jsonl" \
    --source-uri "github.com/$BUILDER_REPOSITORY" \
    "$BUILDER_RELEASE_BINARY" || exit 6

builder_commit=$(gh api /repos/"$BUILDER_REPOSITORY"/git/ref/tags/"$builder_tag" | jq -r '.object.sha')
provenance_commit=$(jq -r '.payload' <"$BUILDER_RELEASE_BINARY.intoto.jsonl" | base64 -d | jq -r '.predicate.materials[0].digest.sha1')
if [[ "$builder_commit" != "$provenance_commit" ]]; then
    echo "Builder commit sha $builder_commit != provenance material $provenance_commit"
    exit 5
fi

#TODO: verify the command
echo "Builder provenance verified at tag $builder_tag and commit $builder_commit"
