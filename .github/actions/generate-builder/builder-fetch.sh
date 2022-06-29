#!/usr/bin/env bash

set -euo pipefail

# Caller sets the following:
#BUILDER_REPOSITORY="slsa-framework/slsa-github-generator"
#BUILDER_TAG="v13.0.10" or "6a1e642a8689671a2cec9287149eb50bd9fe5ef6"
#BUILDER_RELEASE_BINARY="builder-linux-amd64"
#VERIFIER_REPOSITORY="slsa-framework/slsa-verifier"
#VERIFIER_RELEASE="v13.0.10"
#VERIFIER_RELEASE_BINARY="slsa-verifier-linux-amd64"
#VERIFIER_RELEASE_BINARY_SHA256="89fbcba9aed67d5146ea99946c7e4e5a80e3767871f0e3ffcd0b582134efd010"

PREFIX="refs/tags/"

# Extract version.
if [[ "$BUILDER_REF" != "$PREFIX"* ]]; then
    echo "Invalid ref: $BUILDER_REF"
    exit 2
fi

BUILDER_TAG="${BUILDER_REF#"$PREFIX"}"

if [[ "$BUILDER_TAG" == "$(echo -n "$BUILDER_TAG" | grep -P '^[a-f\d]{40}$')" ]]; then
    echo "Builder referenced by hash: $BUILDER_TAG"
    echo "Resolving..."

    RELEASE_TAG=""

    # List the releases and find the corepsonding hash.
    RELEASE_LIST=$(gh release -R "$BUILDER_REPOSITORY" -L 50 list)
    while read -r line; do
        TAG=$(echo "$line" | cut -f1)
        BRANCH=$(gh release -R "$BUILDER_REPOSITORY" view "$TAG" --json targetCommitish --jq '.targetCommitish')
        if [[ "$BRANCH" != "main" ]]; then
            continue
        fi
        COMMIT=$(gh api /repos/"$BUILDER_REPOSITORY"/git/ref/tags/"$TAG" | jq -r '.object.sha')
        if [[ "$COMMIT" == "$BUILDER_TAG" ]]; then
            RELEASE_TAG="$TAG"
            echo "Found tag $BUILDER_TAG match at tag $TAG and commit $COMMIT"
            break
        fi
    done <<<"$RELEASE_LIST"

    if [[ -z "$RELEASE_TAG" ]]; then
        echo "Tag not found for $BUILDER_TAG"
        exit 3
    fi

    BUILDER_TAG="$RELEASE_TAG"
fi

if [[ "$BUILDER_TAG" != "$(echo -n "$BUILDER_TAG" | grep -P '^v\d*(\.([\d]{1,})){0,2}$')" ]]; then
    echo "Invalid ref: $BUILDER_TAG"
    exit 7
fi

echo "Builder version: $BUILDER_TAG"

echo "BUILDER_REPOSITORY: $BUILDER_REPOSITORY"

# Fetch the release binary and provenance.
gh release -R "$BUILDER_REPOSITORY" download "$BUILDER_TAG" -p "$BUILDER_RELEASE_BINARY*" || exit 10

# Fetch the verifier at the right hash.
gh release -R "$VERIFIER_REPOSITORY" download "$VERIFIER_RELEASE" -p "$VERIFIER_RELEASE_BINARY" || exit 11
COMPUTED_HASH=$(sha256sum "$VERIFIER_RELEASE_BINARY" | awk '{print $1}')
echo "verifier hash computed is $COMPUTED_HASH"
echo "$VERIFIER_RELEASE_BINARY_SHA256 $VERIFIER_RELEASE_BINARY" | sha256sum --strict --check --status || exit 4
echo "verifier hash verification has passed"

# Verify the provenance of the builder.
chmod a+x "$VERIFIER_RELEASE_BINARY"
./"$VERIFIER_RELEASE_BINARY" --branch "main" \
    --tag "$BUILDER_TAG" \
    --artifact-path "$BUILDER_RELEASE_BINARY" \
    --provenance "$BUILDER_RELEASE_BINARY.intoto.jsonl" \
    --source "github.com/$BUILDER_REPOSITORY" || exit 6

BUILDER_COMMIT=$(gh api /repos/"$BUILDER_REPOSITORY"/git/ref/tags/"$BUILDER_TAG" | jq -r '.object.sha')
PROVENANCE_COMMIT=$(jq -r '.payload' <"$BUILDER_RELEASE_BINARY.intoto.jsonl" | base64 -d | jq -r '.predicate.materials[0].digest.sha1')
if [[ "$BUILDER_COMMIT" != "$PROVENANCE_COMMIT" ]]; then
    echo "Builder commit sha $BUILDER_COMMIT != provenance material $PROVENANCE_COMMIT"
    exit 5
fi

#TODO: verify the command
echo "Builder provenance verified at tag $BUILDER_TAG and commit $BUILDER_COMMIT"
