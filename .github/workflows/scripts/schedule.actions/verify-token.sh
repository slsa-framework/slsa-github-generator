#!/bin/bash

set -euo pipefail

source "./.github/workflows/scripts/assert.sh"

# Tool information.
echo "VERIFIED_TOKEN: $VERIFIED_TOKEN"
echo "TOOL_REPOSITORY: $TOOL_REPOSITORY"
echo "TOOL_REF: $TOOL_REF"
echo "TOOL_URI: $TOOL_URI"

# Non-GitHub's information.
audience=$(echo "$VERIFIED_TOKEN" | jq -r '.builder.audience')
runner_label=$(echo "$VERIFIED_TOKEN" | jq -r '.builder.runner_label')
private_repository=$(echo "$VERIFIED_TOKEN" | jq -r '.builder.private_repository')
action_path=$(echo "$VERIFIED_TOKEN" | jq -r '.tool.actions.build_artifacts.path')
inputs=$(echo "$VERIFIED_TOKEN" | jq -rc '.tool.inputs')

# GitHub's information.
run_attempt=$(echo "$VERIFIED_TOKEN" | jq -r '.github.run_attempt')
run_id=$(echo "$VERIFIED_TOKEN" | jq -r '.github.run_id')
run_number=$(echo "$VERIFIED_TOKEN" | jq -r '.github.run_number')
sha=$(echo "$VERIFIED_TOKEN" | jq -r '.github.sha')
workflow=$(echo "$VERIFIED_TOKEN" | jq -r '.github.workflow')
event_name=$(echo "$VERIFIED_TOKEN" | jq -r '.github.event_name')
repository=$(echo "$VERIFIED_TOKEN" | jq -r '.github.repository')
repository_owner=$(echo "$VERIFIED_TOKEN" | jq -r '.github.repository_owner')
ref=$(echo "$VERIFIED_TOKEN" | jq -r '.github.ref')
ref_type=$(echo "$VERIFIED_TOKEN" | jq -r '.github.ref_type')
actor=$(echo "$VERIFIED_TOKEN" | jq -r '.github.actor')

assert_eq "delegator_generic_slsa3.yml", "$audience"
assert_eq "$GITHUB_RUN_ATTEMPT", "$run_attempt"
assert_eq "$GITHUB_RUN_NUMBER", "$run_number"
assert_eq "$GITHUB_RUN_ID", "$run_id"
assert_eq "$GITHUB_SHA", "$sha"
assert_eq "$GITHUB_WORKFLOW", "$workflow"
assert_eq "ubuntu-latest", "$runner_label"
assert_eq "true", "$private_repository"
assert_eq "./actions/build-artifacts-composite", "$action_path"
assert_eq '{"name1":"value1","name2":"value2","private-repository":true}', "$inputs"
assert_eq "$GITHUB_EVENT_NAME", "$event_name"
assert_eq "$GITHUB_REPOSITORY", "$repository"
assert_eq "$GITHUB_REPOSITORY_OWNER", "$repository_owner"
assert_eq "$GITHUB_REF", "$ref"
assert_eq "$GITHUB_REF_TYPE", "$ref_type"
assert_eq "$GITHUB_ACTOR", "$actor"
assert_eq "$TOOL_REPOSITORY", "$GITHUB_REPOSITORY"
assert_eq "$TOOL_REF", "$GITHUB_REF"
assert_eq "$GITHUB_URI", "https://github.com/$GITHUB_REPOSITORY/.github/workflows/schedule.e2e.verify-token.default.yml@$GITHUB_REF"
