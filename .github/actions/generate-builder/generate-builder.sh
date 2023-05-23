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

if [[ "$COMPILE_BUILDER" == true ]]; then
    echo "Building the builder with ref: $BUILDER_REF"

    cd "$BUILDER_DIR"
    git checkout "$BUILDER_REF"

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
