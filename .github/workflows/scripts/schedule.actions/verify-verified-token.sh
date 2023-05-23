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

source "./.github/workflows/scripts/e2e-assert.sh"
source "./.github/workflows/scripts/e2e-verify.common.sh"

# Tool information.
echo "VERIFIED_TOKEN: $VERIFIED_TOKEN"
echo "TOOL_REPOSITORY: $TOOL_REPOSITORY"
echo "TOOL_REF: $TOOL_REF"
echo "PREDICATE: $PREDICATE"

e2e_verify_decoded_token "$VERIFIED_TOKEN"

e2e_assert_eq "$TOOL_REPOSITORY" "$GITHUB_REPOSITORY"
e2e_assert_eq "$TOOL_REF" "$GITHUB_REF"

PREDICATE_CONTENT=$(<"$PREDICATE")
echo "PREDICATE_CONTENT: $PREDICATE_CONTENT"

# Verify common predicate fields.
e2e_verify_common_all_v1 "$PREDICATE_CONTENT"
e2e_verify_predicate_v1_buildDefinition_buildType "$PREDICATE_CONTENT" "https://github.com/slsa-framework/slsa-github-generator/delegator-generic@v0"
e2e_verify_predicate_v1_runDetails_builder_id "$PREDICATE_CONTENT" "https://github.com/$GITHUB_REPOSITORY/.github/workflows/e2e.verify-token.schedule.yml@$GITHUB_REF"
e2e_verify_predicate_v1_buildDefinition_externalParameters_workflow_path "$PREDICATE_CONTENT" "$(e2e_this_file_full_path)"

# Verify external parameters source
e2e_verify_predicate_v1_buildDefinition_externalParameters_source "$PREDICATE_CONTENT" "{\"uri\":\"git+https://github.com/$GITHUB_REPOSITORY@$GITHUB_REF\",\"digest\":{\"sha1\":\"$GITHUB_SHA\"}}"
e2e_verify_predicate_v1_buildDefinition_externalParameters_inputs "$PREDICATE_CONTENT" '{"name1":"value1","name2":"***","name3":"value3","name4":"***","name5":"value5","name6":"***","private-repository":true}'
