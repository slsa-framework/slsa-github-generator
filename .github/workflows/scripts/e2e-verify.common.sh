#!/usr/bin/env bash
#
# This file contains tests for common fields of Github Actions provenance.

source "./.github/workflows/scripts/e2e-utils.sh"

# Runs all generic SLSA checks that shouldn't change on a per-builder basis.
# $1: the attestation content
e2e_verify_common_all() {
    e2e_verify_common_builder "$1"
    e2e_verify_common_invocation "$1"
    e2e_verify_common_metadata "$1"
    e2e_verify_common_materials "$1"
}

# Verifies the builder for generic provenance.
# $1: the attestation content
e2e_verify_common_builder() {
    e2e_verify_predicate_builder_id "$1" "https://github.com/Attestations/GitHubHostedActions@v1"
}

# Verifies the invocation for generic provenance.
# $1: the attestation content
e2e_verify_common_invocation() {
    # NOTE: We set GITHUB_WORKFLOW to the entryPoint for pull_requests.
    # TODO(github.com/slsa-framework/slsa-github-generator/issues/131): support retrieving entryPoint in pull requests.
    e2e_verify_predicate_invocation_configSource "$1" "{\"uri\":\"git+https://github.com/$GITHUB_REPOSITORY@$GITHUB_REF\",\"digest\":{\"sha1\":\"$GITHUB_SHA\"},\"entryPoint\":\"$GITHUB_WORKFLOW\"}"

    e2e_verify_predicate_invocation_environment "$1" "github_actor" "$GITHUB_ACTOR"
    e2e_verify_predicate_invocation_environment "$1" "github_sha1" "$GITHUB_SHA"
    # e2e_verify_predicate_invocation_environment "$1" "os" "ubuntu20"
    # e2e_verify_predicate_invocation_environment "$1" "arch" "X64"
    e2e_verify_predicate_invocation_environment "$1" "github_event_name" "$GITHUB_EVENT_NAME"
    e2e_verify_predicate_invocation_environment "$1" "github_ref" "$GITHUB_REF"
    e2e_verify_predicate_invocation_environment "$1" "github_ref_type" "$GITHUB_REF_TYPE"
    e2e_verify_predicate_invocation_environment "$1" "github_run_id" "$GITHUB_RUN_ID"
    e2e_verify_predicate_invocation_environment "$1" "github_run_number" "$GITHUB_RUN_NUMBER"
    e2e_verify_predicate_invocation_environment "$1" "github_run_attempt" "$GITHUB_RUN_ATTEMPT"
    # The checks below are commented out because they are populated via the OIDC token, which is not available in PRs.
    #ACTOR_ID=$(gh api -H "Accept: application/vnd.github.v3+json"   /users/"$GITHUB_ACTOR" | jq -r '.id')
    #OWNER_ID=$(gh api -H "Accept: application/vnd.github.v3+json"   /users/"$GITHUB_REPOSITORY_OWNER" | jq -r '.id')
    #REPO_ID=$(gh api -H "Accept: application/vnd.github.v3+json"   /repos/"$GITHUB_REPOSITORY" | jq -r '.id')
    #e2e_verify_predicate_invocation_environment "$1" "github_actor_id" "$ACTOR_ID"
    #e2e_verify_predicate_invocation_environment "$1" "github_repository_owner_id" "$OWNER_ID"
    #e2e_verify_predicate_invocation_environment "$1" "github_repository_id" "$REPO_ID"
}

# Verifies the expected metadata.
# $1: the attestation content
e2e_verify_common_metadata() {
    e2e_verify_predicate_metadata "$1" "{\"buildInvocationID\":\"$GITHUB_RUN_ID-$GITHUB_RUN_ATTEMPT\",\"completeness\":{\"parameters\":true,\"environment\":false,\"materials\":false},\"reproducible\":false}"
}

# Verifies the materials include the GitHub repository.
# $1: the attestation content
e2e_verify_common_materials() {
    e2e_verify_predicate_materials "$1" "{\"uri\":\"git+https://github.com/$GITHUB_REPOSITORY@$GITHUB_REF\",\"digest\":{\"sha1\":\"$GITHUB_SHA\"}}"
}

# Verifies common fields of the SLSA v1.0 predicate.
# $1: the predicate content
e2e_verify_common_all_v1() {
    e2e_verify_common_buildDefinition_v1 "$1"
    e2e_verify_common_runDetails_v1 "$1"
}

# Verifies common fields of the SLSA v1.0 predicate buildDefinition.
# $1: the predicate content
e2e_verify_common_buildDefinition_v1() {
    # This does not include buildType since it is not common to all.
    e2e_verify_predicate_v1_buildDefinition_externalParameters_workflowPath "$1" "$(e2e_this_file)"
    e2e_verify_predicate_v1_buildDefinition_externalParameters_source "$1" "{\"artifact\":{\"uri\":\"git+https://github.com/$GITHUB_REPOSITORY@$GITHUB_REF\",\"digest\":{\"sha1\":\"$GITHUB_SHA\"}}}"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "GITHUB_EVENT_NAME" "$GITHUB_EVENT_NAME"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "GITHUB_JOB" "$GITHUB_JOB"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "GITHUB_REF" "$GITHUB_REF"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "GITHUB_REF_TYPE" "$GITHUB_REF_TYPE"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "GITHUB_REPOSITORY" "$GITHUB_REPOSITORY"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "GITHUB_RUN_ATTEMPT" "$GITHUB_RUN_ATTEMPT"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "GITHUB_RUN_ID" "$GITHUB_RUN_ID"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "GITHUB_RUN_NUMBER" "$GITHUB_RUN_NUMBER"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "GITHUB_SHA" "$GITHUB_SHA"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "GITHUB_WORKFLOW" "$GITHUB_WORKFLOW"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "GITHUB_ACTOR_ID" "$GITHUB_ACTOR_ID"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "GITHUB_REPOSITORY_ID" "$GITHUB_REPOSITORY_ID"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "GITHUB_REPSITORY_OWNER_ID" "$GITHUB_REPSITORY_OWNER_ID"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "GITHUB_WORKFLOW_REF" "$GITHUB_WORKFLOW_REF"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "GITHUB_WORKFLOW_SHA" "$GITHUB_WORKFLOW_SHA"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "IMAGE_OS" "$IMAGE_OS"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "IMAGE_VERSION" "$IMAGE_VERSION"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "RUNNER_ARCH" "$RUNNER_ARCH"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "RUNNER_NAME" "$RUNNER_NAME"
    e2e_verify_predicate_v1_buildDefinition_systemParameters "$1" "RUNNER_OS" "$RUNNER_OS"
}

# Verifies common fields of the SLSA v1.0 predicate runDetails.
# $1: the predicate content
e2e_verify_common_runDetails_v1() {
    # This does not include the builder ID since it is not common to all.
    e2e_verify_predicate_v1_runDetails_metadata_invocationId "$1" "https://github.com/$GITHUB_REPOSITORY/actions/runs/$GITHUB_RUN_ID/attempts/$GITHUB_RUN_ATTEMPT"
}