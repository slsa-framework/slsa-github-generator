#!/usr/bin/env bash

source "./.github/workflows/scripts/e2e-assert.sh"

e2e_verify_predicate_subject_name() {
    _e2e_verify_query "$1" "$2" '.subject[0].name'
}

e2e_verify_predicate_builder_id() {
    _e2e_verify_query "$1" "$2" '.predicate.builder.id'
}

e2e_verify_predicate_builderType() {
    _e2e_verify_query "$1" "$2" '.predicate.buildType'
}

e2e_verify_predicate_invocation_configSource() {
    _e2e_verify_query "$1" "$2" '.predicate.invocation.configSource'
}

# e2e_verify_predicate_invocation_environment(attestation, env_key, expected)
e2e_verify_predicate_invocation_environment() {
    _e2e_verify_query "$1" "$3" '.predicate.invocation.environment.'"$2"
}

e2e_verify_predicate_buildConfig_command() {
    _e2e_verify_query "$1" "$2" '.predicate.buildConfig.steps[0].command[1:]'
}

e2e_verify_predicate_buildConfig_env() {
    local attestation="$1"
    local expected
    expected=$(echo -n "$2" | jq -c '.| sort')
    _e2e_verify_query "${attestation}" "${expected}" '.predicate.buildConfig.steps[0].env | sort'
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
