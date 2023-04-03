#!/bin/bash

set -euo pipefail

source "./.github/workflows/scripts/e2e-assert.sh"
source "./.github/workflows/scripts/e2e-verify.common.sh"

echo "PREDICATE: $PREDICATE"
PREDICATE_CONTENT=$(<"$PREDICATE")
echo "PREDICATE_CONTENT: $PREDICATE_CONTENT"

# Verify common predicate fields.
e2e_verify_common_all_v1  "$PREDICATE_CONTENT"
e2e_verify_predicate_v1_buildDefinition_buildType "$PREDICATE_CONTENT" "https://slsa.dev/container-based-build/v0.1?draft"
e2e_verify_predicate_v1_runDetails_builder_id "$PREDICATE_CONTENT" "https://github.com/$GITHUB_REPOSITORY/.github/workflows/e2e.create-docker_based-predicate.schedule.yml@$GITHUB_REF"

# Verify source
e2e_verify_predicate_v1_buildDefinition_externalParameters_source "$PREDICATE_CONTENT" "{\"uri\":\"git+https://github.com/$GITHUB_REPOSITORY@$GITHUB_REF\",\"digest\":{\"sha1\":\"$GITHUB_SHA\"}}"
# Verify resolved dependencies.
e2e_verify_predicate_v1_buildDefinition_resolvedDependencies "$PREDICATE_CONTENT" "[{\"uri\":\"git+https://github.com/$GITHUB_REPOSITORY@$GITHUB_REF\",\"digest\":{\"sha256\":\"46b3ce0fbb2998880c5518225b41ddf49fc0850b9b9a25e1ac944bc587c03ea7\"}}]"
