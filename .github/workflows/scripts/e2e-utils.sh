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

source "./.github/workflows/scripts/e2e-assert.sh"

e2e_this_file_full_path() {
    gh api -H "Accept: application/vnd.github.v3+json" "/repos/$GITHUB_REPOSITORY/actions/runs/$GITHUB_RUN_ID" | jq -r '.path'
}

# e2e_this_file prints the name of the currently running workflow file.
# Note: this requires GH_TOKEN to be set in the workflows.
e2e_this_file() {
    e2e_this_file_full_path | cut -d '/' -f3
}

e2e_verify_predicate_subject_name() {
    _e2e_verify_query "$1" "$2" '.subject[0].name'
}

e2e_verify_predicate_builder_id() {
    _e2e_verify_query "$1" "$2" '.predicate.builder.id'
}

e2e_verify_predicate_buildType() {
    _e2e_verify_query "$1" "$2" '.predicate.buildType'
}

e2e_verify_predicate_invocation_configSource() {
    _e2e_verify_query "$1" "$2" '.predicate.invocation.configSource'
}

# e2e_verify_predicate_invocation_environment(attestation, env_key, expected)
e2e_verify_predicate_invocation_environment() {
    _e2e_verify_query "$1" "$3" '.predicate.invocation.environment.'"$2"
}

# $1: step number
# $2: the attestation content
# $3: expected value.
e2e_verify_predicate_buildConfig_step_command() {
    _e2e_verify_query "$2" "$3" ".predicate.buildConfig.steps[$1].command[1:]"
}

# $1: step number
# $2: the attestation content
# $3: expected value.
e2e_verify_predicate_buildConfig_step_env() {
    local attestation="$2"
    local expected
    expected="$(echo -n "$3" | jq -c '.| sort')"

    if [[ "${expected}" == "[]" ]]; then
        _e2e_verify_query "${attestation}" "null" ".predicate.buildConfig.steps[$1].env"
    else
        _e2e_verify_query "${attestation}" "${expected}" ".predicate.buildConfig.steps[$1].env | sort"
    fi
}

# $1: step number
# $2: the attestation content
# $3: expected value.
e2e_verify_predicate_buildConfig_step_workingDir() {
    _e2e_verify_query "$2" "$3" ".predicate.buildConfig.steps[$1].workingDir"
}

e2e_verify_predicate_metadata() {
    _e2e_verify_query "$1" "$2" '.predicate.metadata'
}

e2e_verify_predicate_materials() {
    _e2e_verify_query "$1" "$2" '.predicate.materials[0]'
}

# _e2e_verify_query verifies that the result of the given jq query is equal to
# the expected value.
_e2e_verify_query() {
    local attestation="$1"
    local expected="$2"
    local query="$3"
    name=$(echo -n "${attestation}" | jq -c -r "${query}")
    e2e_assert_eq "${name}" "${expected}" "${query} should be ${expected}"
}

# _e2e_verify_presence verifies that the result of the given jq query is present.
_e2e_verify_presence() {
    local attestation="$1"
    local query="$2"
    name=$(echo -n "${attestation}" | jq -c -r "${query} | select(type != \"null\")")
    assert_not_empty "${name}" "${query} should not be empty"
}

e2e_verify_predicate_v1_buildDefinition_externalParameters_workflow() {
    if [[ -z "${BUILDER_INTERFACE_TYPE:-}" ]]; then
        return 0
    fi
    if [[ "${BUILDER_INTERFACE_TYPE}" == "builder" ]]; then
        return 0
    fi

    _e2e_verify_query "$1" "$2" ".buildDefinition.externalParameters.workflow.path"
    _e2e_verify_query "$1" "$3" ".buildDefinition.externalParameters.workflow.ref"
    _e2e_verify_query "$1" "$4" ".buildDefinition.externalParameters.workflow.repository"
}

e2e_verify_predicate_v1_buildDefinition_externalParameters_source() {
    _e2e_verify_query "$1" "$2" '.buildDefinition.externalParameters.source'
}

e2e_verify_predicate_v1_buildDefinition_externalParameters_inputs() {
    _e2e_verify_query "$1" "$2" '.buildDefinition.externalParameters.inputs'
}

e2e_verify_predicate_v1_buildDefinition_buildType() {
    _e2e_verify_query "$1" "$2" '.buildDefinition.buildType'
}

e2e_verify_predicate_v1_buildDefinition_resolvedDependencies() {
    _e2e_verify_query "$1" "$2" '.buildDefinition.resolvedDependencies'
}

e2e_verify_predicate_v1_buildDefinition_internalParameters() {
    _e2e_verify_query "$1" "$3" '.buildDefinition.internalParameters.'"$2"
}

e2e_present_predicate_v1_buildDefinition_internalParameters() {
    _e2e_verify_presence "$1" '.buildDefinition.internalParameters.'"$2"
}

e2e_verify_predicate_v1_runDetails_builder_id() {
    _e2e_verify_query "$1" "$2" '.runDetails.builder.id'
}

e2e_verify_predicate_v1_runDetails_metadata_invocationId() {
    _e2e_verify_query "$1" "$2" '.runDetails.metadata.invocationId'
}

e2e_get_source_sha1() {
    local digest="$GITHUB_SHA"
    if [[ -n "${CHECKOUT_SHA1:-}" ]]; then
        # If the TRW provided a sha1 for checkout, the predicate should use it instead.
        digest="${CHECKOUT_SHA1}"
    fi
    echo "$digest"
}
