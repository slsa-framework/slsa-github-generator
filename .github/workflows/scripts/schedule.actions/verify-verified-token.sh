#!/bin/bash

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
e2e_verify_predicate_v1_buildDefinition_buildType "$PREDICATE_CONTENT" "https://github.com/slsa-framework/slsa-github-generator/delegator-generic@v1"
e2e_verify_predicate_v1_runDetails_builder_id "$PREDICATE_CONTENT" "https://github.com/$GITHUB_REPOSITORY/.github/workflows/e2e.verify-token.schedule.yml@$GITHUB_REF"
e2e_verify_predicate_v1_buildDefinition_externalParameters_workflow_path "$PREDICATE_CONTENT" "$(e2e_this_file_full_path)"
e2e_verify_predicate_v1_buildDefinition_externalParameters_inputs "$PREDICATE_CONTENT" '{"name1":"value1","name2":"***","name3":"value3","name4":"***","name5":"value5","name6":"***","private-repository":true}'

