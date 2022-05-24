#!/usr/bin/env bash
# Caller sets the following:
#BUILDER_REPOSITORY="slsa-framework/example-package"
#BUILDER_REF="v13.0.10" or "6a1e642a8689671a2cec9287149eb50bd9fe5ef6"
#BUILDER_REF="6a1e642a8689671a2cec9287149eb50bd9fe5ef6" # v11.0.14
#BUILDER_RELEASE_BINARY="builder-linux-amd64"

# Verifier info.
#VERIFIER_REPOSITORY="slsa-framework/slsa-verifier"
#VERIFIER_REPOSITORY="slsa-framework/example-package"
#VERIFIER_RELEASE="v13.0.10"
#VERIFIER_RELEASE_BINARY="binary-linux-amd64"
#VERIFIER_RELEASE_BINARY_SHA256="89fbcba9aed67d5146ea99946c7e4e5a80e3767871f0e3ffcd0b582134efd010"

if [[ -z "$BUILDER_REF" ]]; then
    echo "empty hash"
    exit 2
fi

if [[ "$BUILDER_REF" = "$(echo -n "$BUILDER_REF" | grep -P '^[a-f\d]{40}$')" ]]; then
    echo "Builder referenced by hash: $BUILDER_REF"
    echo "Resolving..."
    
    RELEASE_TAG=""

    # List the releases and find the corepsonding hash.
    RELEASE_LIST=$(gh release -R "$BUILDER_REPOSITORY" -L 50 list)
    while read line; do
        TAG=$(echo "$line" | cut -f1)
        BRANCH=$(gh release -R "$BUILDER_REPOSITORY" view "$TAG" --json targetCommitish --jq '.targetCommitish')
        if [[ "$BRANCH" != "main" ]]; then
            continue
        fi
        COMMIT=$(gh api /repos/"$BUILDER_REPOSITORY"/git/ref/tags/"$TAG" | jq -r '.object.sha')
        if [[ "$COMMIT" == "$BUILDER_REF" ]]; then
            RELEASE_TAG="$TAG"
            echo "Found tag $BUILDER_REF match at tag $TAG and commit $COMMIT"
            break
        fi
    done <<< "$RELEASE_LIST"

    if [[ -z "$RELEASE_TAG" ]]; then 
        echo "Tag not found for $BUILDER_REF"
        exit 3
    fi

    BUILDER_REF="$RELEASE_TAG"
fi

if [[ "$BUILDER_REF" != "$(echo -n "$BUILDER_REF" | grep -P '^v\d*(\.([\d]{1,})){0,2}$')" ]]; then
    echo "Invalid ref: $BUILDER_REF"
    exit 0
fi

echo "Builder version: $BUILDER_REF"

# Fetch the release binary and provenance.
gh release -R "$BUILDER_REPOSITORY" download "$BUILDER_REF" -p "$BUILDER_RELEASE_BINARY*"
# Test
#mv $BUILDER_RELEASE_BINARY builder-binary
#mv $BUILDER_RELEASE_BINARY.intoto.jsonl builder-binary.intoto.jsonl

# Fetch the verifier at the right hash.
gh release -R "$VERIFIER_REPOSITORY" download "$VERIFIER_RELEASE" -p "$VERIFIER_RELEASE_BINARY"
COMPUTED_HASH=$(sha256sum "$VERIFIER_RELEASE_BINARY" | awk '{print $1}')
echo "verifier hash computed is $COMPUTED_HASH"
echo "$VERIFIER_RELEASE_BINARY_SHA256 $VERIFIER_RELEASE_BINARY" | sha256sum --strict --check --status || exit 4
echo "verifier hash verification has passed"
# Test
#mv $VERIFIER_RELEASE_BINARY verifier-binary


# Verify the provenance of the builder.
./"$VERIFIER_RELEASE_BINARY" --branch "main" \
                            --tag "$BUILDER_REF" \
                            --artifact-path "$BUILDER_RELEASE_BINARY" \
                            --provenance "$BUILDER_RELEASE_BINARY.intoto.jsonl" \
                            --source "github.com/$BUILDER_REPOSITORY"

#./verifier-binary --branch main --tag "$BUILDER_REF" --artifact-path builder-binary --provenance builder-binary.intoto.jsonl --source "github.com/$BUILDER_REPOSITORY" || exit 5
BUILDER_COMMIT=$(gh api /repos/"$BUILDER_REPOSITORY"/git/ref/tags/"$BUILDER_REF" | jq -r '.object.sha')
PROVENANCE_COMMIT=$(cat "$BUILDER_RELEASE_BINARY.intoto.jsonl" | jq -r '.payload' | base64 -d | jq -r '.predicate.materials[0].digest.sha1')
if [[ "$BUILDER_COMMIT" != "$PROVENANCE_COMMIT" ]]; then
    echo "Builder commit sha $BUILDER_COMMIT != provenance material $PROVENANCE_COMMIT"
    exit 5
fi

echo "Builder provenance verified at tag $BUILDER_REF and commit $BUILDER_COMMIT"
