#!/usr/bin/env bash

# To test:
# export GITHUB_SHA=6f3b6435f5a17a25ad6cf2704d0c192bcef8193f
# export GITHUB_RUN_ID=2272442563
# export GITHUB_ACTOR=laurentsimon
# export GITHUB_RUN_ATTEMPT=1
# export GITHUB_REF=refs/heads/branch-name or refs/tags/tag-name
# export GITHUB_REF_TYPE=branch or tag
# export GITHUB_REPOSITORY=slsa-framework/example-package
# export GITHUB_REF_NAME=v1.2.3
# export GITHUB_WORKFLOW=go schedule main SLSA3 config-noldflags
# export THIS_FILE=e2e.go.workflow_dispatch.main.config-noldflags.slsa3.yml
# export BINARY=binary-linux-amd64
# export PROVENANCE=example.intoto.jsonl
# export GITHUB_EVENT_NAME=pull_request

source "./.github/workflows/scripts/e2e-utils.sh"

BRANCH="main"

# Provenance content verification.
ATTESTATION=$(cat "$PROVENANCE" | base64 -d)
LDFLAGS=$(echo "$THIS_FILE" | cut -d '.' -f3 | grep -v noldflags)

e2e_verify_predicate_subject_name "$ATTESTATION" "$BINARY"
e2e_verify_predicate_builder_id "$ATTESTATION" "https://github.com/Attestations/GitHubHostedActions@v1"
e2e_verify_predicate_builderType "$ATTESTATION" "https://github.com/slsa-framework/slsa-github-generator-go@v1"

e2e_verify_predicate_invocation_configSource "$ATTESTATION" "{\"uri\":\"git+https://github.com/$GITHUB_REPOSITORY@$GITHUB_REF\",\"digest\":{\"sha1\":\"$GITHUB_SHA\"},\"entryPoint\":\"$GITHUB_WORKFLOW\"}"

e2e_verify_predicate_invocation_environment "$ATTESTATION" "github_actor" "$GITHUB_ACTOR"
e2e_verify_predicate_invocation_environment "$ATTESTATION" "github_sha1" "$GITHUB_SHA"
e2e_verify_predicate_invocation_environment "$ATTESTATION" "os" "ubuntu20"
e2e_verify_predicate_invocation_environment "$ATTESTATION" "arch" "X64"
e2e_verify_predicate_invocation_environment "$ATTESTATION" "github_event_name" "$GITHUB_EVENT_NAME"
e2e_verify_predicate_invocation_environment "$ATTESTATION" "github_ref" "$GITHUB_REF"
e2e_verify_predicate_invocation_environment "$ATTESTATION" "github_ref_type" "$GITHUB_REF_TYPE"

# First step is vendoring
e2e_verify_predicate_buildConfig_step_command "0" "$ATTESTATION" "[\"mod\",\"vendor\"]"
e2e_verify_predicate_buildConfig_step_env "0" "$ATTESTATION" "[]"
e2e_verify_predicate_buildConfig_step_workingDir "0" "$ATTESTATION" "$PWD/internal/builders/go/e2e-presubmits"

# Second step is the actual compilation.
e2e_verify_predicate_buildConfig_step_env "1" "$ATTESTATION" "[\"GOOS=linux\",\"GOARCH=amd64\",\"GO111MODULE=on\",\"CGO_ENABLED=0\"]"
e2e_verify_predicate_buildConfig_step_workingDir "1" "$ATTESTATION" "$PWD/internal/builders/go/e2e-presubmits"

if [[ -n "$LDFLAGS" ]]; then
    e2e_verify_predicate_buildConfig_step_command "1" "$ATTESTATION" "[\"build\",\"-mod=vendor\",\"-trimpath\",\"-tags=netgo\",\"-ldflags=-X main.gitVersion=v1.2.3 -X main.gitCommit=abcdef -X main.gitBranch=$BRANCH\",\"-o\",\"$BINARY\",\"main.go\"]"
    chmod a+x ./"$BINARY"
    V=$(./"$BINARY" | grep 'GitVersion: v1.2.3')
    C=$(./"$BINARY" | grep 'GitCommit: abcdef')
    B=$(./"$BINARY" | grep "GitBranch: main")
    e2e_assert_not_eq "$V" "" "GitVersion should not be empty"
    e2e_assert_not_eq "$C" "" "GitCommit should not be empty"
    e2e_assert_not_eq "$B" "" "GitBranch should not be empty"
else
    e2e_verify_predicate_buildConfig_step_command "1" "$ATTESTATION" "[\"build\",\"-mod=vendor\",\"-trimpath\",\"-tags=netgo\",\"-o\",\"$BINARY\",\"main.go\"]"
fi

e2e_verify_predicate_metadata "$ATTESTATION" "{\"buildInvocationID\":\"$GITHUB_RUN_ID-$GITHUB_RUN_ATTEMPT\",\"completeness\":{\"parameters\":true,\"environment\":false,\"materials\":false},\"reproducible\":false}"
e2e_verify_predicate_materials "$ATTESTATION" "{\"uri\":\"git+https://github.com/$GITHUB_REPOSITORY@$GITHUB_REF\",\"digest\":{\"sha1\":\"$GITHUB_SHA\"}}"
