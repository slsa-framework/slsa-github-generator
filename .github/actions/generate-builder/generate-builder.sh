#!/usr/bin/env bash
set -euo pipefail

if [[ "$COMPILE_BUILDER" = true ]]; then
    echo "Building the builder"

    #TODO(reproducible)
    go mod vendor

    # https://go.dev/ref/mod#build-commands.
    go build -mod=vendor -o "$BUILDER_BINARY" "github.com/slsa-framework/slsa-github-generator/$BUILDER_DIR"

else
    echo "Fetching the builder with ref: $BUILDER_REF"

    .github/actions/generate-builder/builder-fetch.sh

    mv "$BUILDER_RELEASE_BINARY" "$BUILDER_BINARY"
fi

chmod u+x "$BUILDER_BINARY"