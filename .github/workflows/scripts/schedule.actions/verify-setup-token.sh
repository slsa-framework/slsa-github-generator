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

source "./.github/workflows/scripts/e2e-verify.common.sh"

# TODO(#1709): Add more token verification.
if [[ "$SLSA_TOKEN" == "" ]]; then
    echo "SLSA_TOKEN is empty."
    exit 1
fi

b64_token=$(echo -n "$SLSA_TOKEN" | cut -d '.' -f2)
decoded_token=$(echo "$b64_token" | base64 -d)

e2e_verify_decoded_token "$decoded_token"
