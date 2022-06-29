#!/usr/bin/env bash
set -euo pipefail

if [[ "$COMPILE_BUILDER" = true ]]; then
    echo "Building the builder"

    cd "$BUILDER_DIR"

    #TODO(reproducible)
    go mod vendor

    # https://go.dev/ref/mod#build-commands.
    go build -mod=vendor -o "$BUILDER_BINARY"

    cd -

    mv "$BUILDER_DIR/$BUILDER_BINARY" .

else
    echo "Fetching the builder with ref: $BUILDER_REF"

    .github/actions/generate-builder/builder-fetch.sh

    mv "$BUILDER_RELEASE_BINARY" "$BUILDER_BINARY"
fi

chmod u+x "$BUILDER_BINARY"