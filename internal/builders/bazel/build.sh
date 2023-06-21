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

# Use associative array as a set to increase efficency in avoiding double copying the target
declare -A FILES_SET

# Allows the modifications from sub-shell loops
export FILES_SET

# Using target string, copy artifact to binaries dir
for CURR_TARGET in "${BUILD_TARGETS[@]}"; do

  # Uses a Starlark expression to pass new line seperated list of files produced by targets into the set of files
  (bazel cquery --output=starlark --starlark:expr="'\n'.join([f.path for f in target.files.to_list()])" "$CURR_TARGET" 2>/dev/null) | 
    while read -r file; do
      echo "$file"
      # Key value is target path, value we do not care about and is set to constant "1"
      FILES_SET["${file}"]="1"
  done
done

echo "first loop complete"
echo ""
echo "below is output of file set:"
echo "${!FILES_SET[@]}"
echo ""
echo "starting copy loop"
echo ""

# Copy set of unique targets to binaries, without !, would give values not keys
for file in "${!FILES_SET[@]}"; do
  echo "$file"
  cp -L "$file" ./binaries
done
