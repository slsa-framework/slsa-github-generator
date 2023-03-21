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

ATTESTATION=$(cat "$PROVENANCE")

# Verify subject
e2e_verify_predicate_subject_name "$ATTESTATION" "$BINARY"

WORKFLOW_PATH=$(echo "${GITHUB_WORKFLOW_REF%@*}" | cut -d'/' -f3-)

# Verify all common provenance fields.
PREDICATE_CONTENT=$(echo "$ATTESTATION" | jq -r '.predicate')
e2e_verify_common_all_v1 "$PREDICATE_CONTENT"
e2e_verify_predicate_v1_runDetails_builder_id "$PREDICATE_CONTENT" "https://github.com/$GITHUB_HEAD_REPOSITORY/$WORKFLOW_PATH@$HEAD_SHA"
e2e_verify_predicate_v1_buildDefinition_buildType "$PREDICATE_CONTENT" "https://slsa.dev/container-based-build/v0.1?draft"

# Verify source: note that the ref is the head sha on pull request presubmits.
e2e_verify_predicate_v1_buildDefinition_externalParameters_source "$PREDICATE_CONTENT" "{\"uri\":\"git+https://github.com/$GITHUB_REPOSITORY@$GITHUB_SHA\",\"digest\":{\"sha1\":\"$GITHUB_SHA\"}}"
