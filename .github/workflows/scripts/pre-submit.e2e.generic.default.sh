#!/usr/bin/env bash
#
# Copyright 2022 SLSA Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

source "./.github/workflows/scripts/e2e-verify.common.sh"

# TODO(github.com/slsa-framework/slsa-github-generator/issues/129): Address base64 output format.
ATTESTATION=$(cat "$PROVENANCE")

# Verify common provenance fields.
e2e_verify_common_all "$ATTESTATION"

e2e_verify_predicate_subject_name "$ATTESTATION" "$BINARY"
e2e_verify_predicate_buildType "$ATTESTATION" "https://github.com/slsa-framework/slsa-github-generator/generic@v1"
