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

# Verify that only safe file functions are used.
results=$(
    find .github/actions/ actions/ -maxdepth 3 -name '*.ts' -not -path "*/tscommon/*" -not -path "*/__tests__/*" -type f -print0 |
        xargs -0 grep -P '(import.*|require.*)"fs"' ||
        true
)
if [[ "$results" != "" ]]; then
    echo "Some Actions are using unsafe file system functions"
    echo "$results"
    exit 1
fi
