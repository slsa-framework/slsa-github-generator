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

actual_diff=$(
    diff .github/workflows/delegator_generic_slsa3.yml .github/workflows/delegator_lowperms-generic_slsa3.yml ||
        true
)
expected_diff=$(cat .github/workflows/scripts/pre-submit.delegators/expected.diff)

if [[ "$expected_diff" != "$actual_diff" ]]; then
    echo "Unexpected differences between the delegator workflows"
    echo "$actual_diff"
    exit 1
fi
