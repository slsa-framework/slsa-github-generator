#!/bin/bash

set -euo pipefail

source "./.github/workflows/scripts/e2e-utils.sh"

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
_e2e_verify_query "$decoded_token" "delegator_generic_slsa3.yml" '.builder.audience'
_e2e_verify_query "$decoded_token" "ubuntu-latest" '.builder.runner_label'
_e2e_verify_query "$decoded_token" "true" '.builder.rekor_log_public'
_e2e_verify_query "$decoded_token" "./actions/build-artifacts-composite" '.tool.actions.build_artifacts.path'
_e2e_verify_query "$decoded_token" '{"name1":"value1","name2":"value2","private-repository":true}' '.tool.inputs'

# GitHub's information.
_e2e_verify_query "$decoded_token" "$GITHUB_ACTOR_ID" '.github.actor_id'
_e2e_verify_query "$decoded_token" "$GITHUB_EVENT_NAME" '.github.event_name'
_e2e_verify_query "$decoded_token" "$GITHUB_JOB" '.github.job'
_e2e_verify_query "$decoded_token" "$GITHUB_REF" '.github.ref'
_e2e_verify_query "$decoded_token" "$GITHUB_REF_TYPE" '.github.ref_type'
_e2e_verify_query "$decoded_token" "$GITHUB_REPOSITORY" '.github.repository'
_e2e_verify_query "$decoded_token" "$GITHUB_REPOSITORY_ID" '.github.repository_id'
_e2e_verify_query "$decoded_token" "$GITHUB_REPOSITORY_OWNER_ID" '.github.repository_owner_id'
_e2e_verify_query "$decoded_token" "$GITHUB_RUN_ATTEMPT" '.github.run_attempt'
_e2e_verify_query "$decoded_token" "$GITHUB_RUN_ID" '.github.run_id'
_e2e_verify_query "$decoded_token" "$GITHUB_RUN_NUMBER" '.github.run_number'
_e2e_verify_query "$decoded_token" "$GITHUB_SHA" '.github.sha'
_e2e_verify_query "$decoded_token" "$GITHUB_WORKFLOW_REF" '.github.workflow_ref'
_e2e_verify_query "$decoded_token" "$GITHUB_WORKFLOW_SHA" '.github.workflow_sha'
