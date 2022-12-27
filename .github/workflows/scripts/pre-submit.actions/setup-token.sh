#!/bin/bash

set -euo pipefail

source "./.github/workflows/scripts/assert.sh"

# NOTE: this is a pre-submit, so the signature is not generated and there is
# just a place holder for it.
echo "SLSA_TOKEN: $SLSA_TOKEN"
[[ "$SLSA_TOKEN" != "" ]]

b64_token=$(echo -n "$SLSA_TOKEN" | cut -d '.' -f2)
echo "b64_token:"
echo "$b64_token"

decoded_token=$(echo "$b64_token" | base64 -d)
echo "decoded_token:"
echo "$decoded_token"

# Non-GitHub's information.
audience=$(echo "$decoded_token" | jq -r '.builder.audience')
runner_label=$(echo "$decoded_token" | jq -r '.builder.runner_label')
private_repository=$(echo "$decoded_token" | jq -r '.builder.private_repository')
action_path=$(echo "$decoded_token" | jq -r '.tool.actions.build_artifacts.path')
inputs=$(echo "$decoded_token" | jq -rc '.tool.inputs')

# GitHub's information.
run_attempt=$(echo "$decoded_token" | jq -r '.github.run_attempt')
run_id=$(echo "$decoded_token" | jq -r '.github.run_id')
run_number=$(echo "$decoded_token" | jq -r '.github.run_number')
sha=$(echo "$decoded_token" | jq -r '.github.sha')
workflow=$(echo "$decoded_token" | jq -r '.github.workflow')
event_name=$(echo "$decoded_token" | jq -r '.github.event_name')
repository=$(echo "$decoded_token" | jq -r '.github.repository')
repository_owner=$(echo "$decoded_token" | jq -r '.github.repository_owner')
ref=$(echo "$decoded_token" | jq -r '.github.ref')
ref_type=$(echo "$decoded_token" | jq -r '.github.ref_type')
actor=$(echo "$decoded_token" | jq -r '.github.actor')

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
