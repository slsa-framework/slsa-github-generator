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

set -euo pipefail

source "./.github/workflows/scripts/e2e-assert.sh"
source "./.github/workflows/scripts/e2e-verify.common.sh"

echo "PREDICATE: $PREDICATE"
PREDICATE_CONTENT=$(<"$PREDICATE")
echo "PREDICATE_CONTENT: $PREDICATE_CONTENT"

# Verify common predicate fields.
e2e_verify_common_all_v1 "$PREDICATE_CONTENT"
e2e_verify_predicate_v1_buildDefinition_buildType "$PREDICATE_CONTENT" "https://slsa.dev/container-based-build/v0.1?draft"
e2e_verify_predicate_v1_runDetails_builder_id "$PREDICATE_CONTENT" "https://github.com/$GITHUB_REPOSITORY/.github/workflows/e2e.create-container_based-predicate.schedule.yml@$GITHUB_REF"

# Verify source
e2e_verify_predicate_v1_buildDefinition_externalParameters_source "$PREDICATE_CONTENT" "{\"uri\":\"git+https://github.com/$GITHUB_REPOSITORY@$GITHUB_REF\",\"digest\":{\"sha1\":\"$GITHUB_SHA\"}}"
# Verify resolved dependencies.
e2e_verify_predicate_v1_buildDefinition_resolvedDependencies "$PREDICATE_CONTENT" "[{\"uri\":\"git+https://github.com/$GITHUB_REPOSITORY@$GITHUB_REF\",\"digest\":{\"sha256\":\"46b3ce0fbb2998880c5518225b41ddf49fc0850b9b9a25e1ac944bc587c03ea7\"}}]"
