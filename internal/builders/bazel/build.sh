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

  # Uses a Starlark expression to pass new line seperated list of files produced by targets into the array files
  mapfile -t files < <(bazel cquery --output=starlark --starlark:expr="'\n'.join([f.path for f in target.files.to_list()])" "$CURR_TARGET" 2>/dev/null)
  echo "$CURR_TARGET"
  echo "start"
  # Copy files into downloadable artifact directory
  for file in "${files[@]}"; do
    
    # Double copying file leads to error in
    # if [[ ! -f ./binaries/"$file" ]]
    # then
    
    ls ./binaries
    cp "$file" ./binaries
    ls ./binaries
    echo " "
    echo " "
    # else
      # file="$file"
  done
done
