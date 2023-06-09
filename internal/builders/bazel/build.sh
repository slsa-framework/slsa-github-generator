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

BUILD_FLAGS="${FLAGS}"
BUILD_TARGETS="${TARGETS}"

bazel build "${BUILD_FLAGS}" "${BUILD_TARGETS}"

IFS=' ' read -r -a ARR_TARGETS <<< "${BUILD_TARGETS}"
for CURR_TARGET in "${ARR_TARGETS[@]}"; do
  CD_PATH=${CURR_TARGET%:*}
  CD_PATH=${CD_PATH////}
  BINARY_NAME=${CURR_TARGET#*:}
  cp "bazel-bin/$CD_PATH/$BINARY_NAME" ./binaries
done
