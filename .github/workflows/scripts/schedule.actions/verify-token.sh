#!/bin/bash

set -euo pipefail

source "./.github/workflows/scripts/e2e-assert.sh"
source "./.github/workflows/scripts/e2e-verify.common.sh"

# Tool information.
echo "VERIFIED_TOKEN: $VERIFIED_TOKEN"
echo "TOOL_REPOSITORY: $TOOL_REPOSITORY"
echo "TOOL_REF: $TOOL_REF"
echo "PREDICATE: $PREDICATE"

# Non-GitHub's information.
audience=$(echo "$VERIFIED_TOKEN" | jq -r '.builder.audience')
runner_label=$(echo "$VERIFIED_TOKEN" | jq -r '.builder.runner_label')
rekor_log_public=$(echo "$VERIFIED_TOKEN" | jq -r '.builder.rekor_log_public')
action_path=$(echo "$VERIFIED_TOKEN" | jq -r '.tool.actions.build_artifacts.path')
inputs=$(echo "$VERIFIED_TOKEN" | jq -rc '.tool.inputs')

# GitHub's information.
run_attempt=$(echo "$VERIFIED_TOKEN" | jq -r '.github.run_attempt')
run_id=$(echo "$VERIFIED_TOKEN" | jq -r '.github.run_id')
run_number=$(echo "$VERIFIED_TOKEN" | jq -r '.github.run_number')
sha=$(echo "$VERIFIED_TOKEN" | jq -r '.github.sha')
workflow=$(echo "$VERIFIED_TOKEN" | jq -r '.github.workflow')
workflow_ref=$(echo "$VERIFIED_TOKEN" | jq -r '.github.workflow_ref')
workflow_sha=$(echo "$VERIFIED_TOKEN" | jq -r '.github.workflow_sha')
event_name=$(echo "$VERIFIED_TOKEN" | jq -r '.github.event_name')
repository=$(echo "$VERIFIED_TOKEN" | jq -r '.github.repository')
repository_owner_id=$(echo "$VERIFIED_TOKEN" | jq -r '.github.repository_owner_id')
ref=$(echo "$VERIFIED_TOKEN" | jq -r '.github.ref')
ref_type=$(echo "$VERIFIED_TOKEN" | jq -r '.github.ref_type')
actor_id=$(echo "$VERIFIED_TOKEN" | jq -r '.github.actor_id')

e2e_assert_eq "delegator_generic_slsa3.yml" "$audience"
e2e_assert_eq "$GITHUB_RUN_ATTEMPT" "$run_attempt"
e2e_assert_eq "$GITHUB_RUN_NUMBER" "$run_number"
e2e_assert_eq "$GITHUB_RUN_ID" "$run_id"
e2e_assert_eq "$GITHUB_SHA" "$sha"
e2e_assert_eq "$GITHUB_WORKFLOW" "$workflow"
e2e_assert_eq "$GITHUB_WORKFLOW_REF" "$workflow_ref"
e2e_assert_eq "$GITHUB_WORKFLOW_SHA" "$workflow_sha"
e2e_assert_eq "ubuntu-latest" "$runner_label"
e2e_assert_eq "true" "$rekor_log_public"
e2e_assert_eq "./actions/build-artifacts-composite" "$action_path"
e2e_assert_eq '{"name1":"value1","name2":"value2","private-repository":true}' "$inputs"
e2e_assert_eq "$GITHUB_EVENT_NAME" "$event_name"
e2e_assert_eq "$GITHUB_REPOSITORY" "$repository"
e2e_assert_eq "$GITHUB_REPOSITORY_OWNER_ID" "$repository_owner_id"
e2e_assert_eq "$GITHUB_REF" "$ref"
e2e_assert_eq "$GITHUB_REF_TYPE" "$ref_type"
e2e_assert_eq "$GITHUB_ACTOR_ID" "$actor_id"
e2e_assert_eq "$TOOL_REPOSITORY" "$GITHUB_REPOSITORY"
e2e_assert_eq "$TOOL_REF" "$GITHUB_REF"

PREDICATE_CONTENT=$(<"$PREDICATE")
echo "PREDICATE_CONTENT: $PREDICATE_CONTENT"

# Verify common predicate fields.
e2e_verify_common_all_v1  "$PREDICATE_CONTENT"
e2e_verify_predicate_v1_buildDefinition_buildType "$PREDICATE_CONTENT" "https://github.com/slsa-framework/slsa-github-generator/delegator-generic@v0"
e2e_verify_predicate_v1_runDetails_builder_id "$PREDICATE_CONTENT" "https://github.com/$GITHUB_REPOSITORY/.github/workflows/e2e.verify-token.schedule.yml@$GITHUB_REF"
