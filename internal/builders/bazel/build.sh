#!/bin/bash -eu
#
# Copyright 2023 SLSA Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

mkdir binaries

# Transfer flags and targets to their respective arrays
IFS=' ' read -r -a BUILD_FLAGS <<< "${FLAGS}"
IFS=' ' read -r -a BUILD_TARGETS <<< "${TARGETS}"

# Build with respect to entire arrays of flags and targets
bazel build "${BUILD_FLAGS[@]}" "${BUILD_TARGETS[@]}"

# Using target string, copy artifact to binaries dir
for CURR_TARGET in "${BUILD_TARGETS[@]}"; do
  # Take out the first two // in CURR_TARGET
  # "//src/internal:fib" --> "src/internal:fib"
  CD_PATH=$(echo "$CURR_TARGET" | cut -d'/' -f3-)

  # Removes field after and including the colon
  # "src/internal:fib" --> "src/internal"
  CD_PATH=$(echo "$CD_PATH" | cut -d':' -f1)

  # Removes everything up to and including the first colon
  # "//src/internal:fib" --> "fib"
  BINARY_NAME=${CURR_TARGET#*:}

  #Copy the binary to artifact directory, binaries
  cp "bazel-bin/$CD_PATH/$BINARY_NAME" ./binaries
done
