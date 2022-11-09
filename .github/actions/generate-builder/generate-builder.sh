#!/usr/bin/env bash
set -euo pipefail

if [[ "$COMPILE_BUILDER" == true ]]; then
    echo "Building the builder"

    cd "$BUILDER_DIR"

    #TODO(reproducible)
    go mod vendor

    # https://go.dev/ref/mod#build-commands.
    go build -mod=vendor -o "$BUILDER_RELEASE_BINARY"

    cd -

    mv "$BUILDER_DIR/$BUILDER_RELEASE_BINARY" .

else
    echo "Fetching the builder with ref: $BUILDER_REF"

    ./__BUILDER_CHECKOUT_DIR__/.github/actions/generate-builder/builder-fetch.sh
fi

chmod u+x "$BUILDER_RELEASE_BINARY"
