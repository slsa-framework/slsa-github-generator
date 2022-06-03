#!/usr/bin/env bash
#
# Copyright 2022 SLSA Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

source "./.github/workflows/scripts/e2e-utils.sh"

# TODO(github.com/slsa-framework/slsa-github-generator/issues/152): Create
# separate script file for common checks.

# TODO(github.com/slsa-framework/slsa-github-generator/issues/129): Address base64 output format.
# disable pipefail and use cat in order to swallow base64 errors.
# set +o pipefail
# ATTESTATION=$(base64 -d "$PROVENANCE" | cat)
# set -o pipefail
# NOTE: Provenance is not-signed and is in simple in-toto format, not DSSE.
ATTESTATION=$(cat "$PROVENANCE")

e2e_verify_predicate_subject_name "$ATTESTATION" "binary-name"
e2e_verify_predicate_builder_id "$ATTESTATION" "https://github.com/Attestations/GitHubHostedActions@v1"
e2e_verify_predicate_builderType "$ATTESTATION" "https://github.com/slsa-framework/slsa-github-generator@v1"

# NOTE: We set GITHUB_WORKFLOW to the entryPoint for pull_requests.
# TODO(github.com/slsa-framework/slsa-github-generator/issues/131): support retrieving entryPoint in pull requests.
e2e_verify_predicate_invocation_configSource "$ATTESTATION" "{\"uri\":\"git+https://github.com/$GITHUB_REPOSITORY@$GITHUB_REF\",\"digest\":{\"sha1\":\"$GITHUB_SHA\"},\"entryPoint\":\"$GITHUB_WORKFLOW\"}"

e2e_verify_predicate_invocation_environment "$ATTESTATION" "github_actor" "$GITHUB_ACTOR"
e2e_verify_predicate_invocation_environment "$ATTESTATION" "github_sha1" "$GITHUB_SHA"
e2e_verify_predicate_invocation_environment "$ATTESTATION" "github_event_name" "$GITHUB_EVENT_NAME"
e2e_verify_predicate_invocation_environment "$ATTESTATION" "github_ref" "$GITHUB_REF"
e2e_verify_predicate_invocation_environment "$ATTESTATION" "github_ref_type" "$GITHUB_REF_TYPE"

e2e_verify_predicate_metadata "$ATTESTATION" "{\"buildInvocationID\":\"$GITHUB_RUN_ID-$GITHUB_RUN_ATTEMPT\",\"completeness\":{\"parameters\":true,\"environment\":false,\"materials\":false},\"reproducible\":false}"
e2e_verify_predicate_materials "$ATTESTATION" "{\"uri\":\"git+https://github.com/$GITHUB_REPOSITORY@$GITHUB_REF\",\"digest\":{\"sha1\":\"$GITHUB_SHA\"}}"
