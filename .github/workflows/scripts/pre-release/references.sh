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

# Verify that Action calls that are pinned (not locally called)
# use the same reference as the release tag.
# See reasoning in ./github/actions/README.md

set -euo pipefail

# Verify the tag has semver format.
cd __EXAMPLE_PACKAGE__
# shellcheck source=/dev/null
source "./.github/workflows/scripts/e2e-utils.sh"
major=$(version_major "$RELEASE_TAG")
minor=$(version_minor "$RELEASE_TAG")
patch=$(version_patch "$RELEASE_TAG")
rc=$(version_rc "$RELEASE_TAG")
echo "\"$RELEASE_TAG\" follows semver:"
echo "major: $major"
echo "minor: $minor"
echo "patch: $patch"
echo "rc: $rc"
cd -

if [ "$RELEASE_TAG" == "" ]; then
    echo "Release tag is empty: \"$RELEASE_TAG\""
    exit 1
fi

cd __THIS_REPO__

# Verify our Actions are referenced by the release tag in workflows.
results=$(
    find .github/workflows/ -maxdepth 1 -name '*.yaml' -o -name '*.yml' -type f -print0 |
        xargs -0 grep -Pn "slsa-framework/slsa-github-generator/.*@(?!$RELEASE_TAG)" |
        sed 's/\(.*:\) *uses:.*\(\/.*\)/\1 [...]\2/' ||
        true
)
if [[ "$results" != "" ]]; then
    echo "Some Actions are not referenced via the correct release tag \"$RELEASE_TAG\" in workflows"
    echo "$results"
    exit 1
fi

# Verify our Actions are referenced by the release tag in internal actions.
results=$(
    find .github/actions/ -maxdepth 2 -name '*.yaml' -o -name '*.yml' -type f -print0 |
        xargs -0 grep -Pn "slsa-framework/slsa-github-generator/.*@(?!$RELEASE_TAG)" |
        sed 's/\(.*:\) *uses:.*\(\/.*\)/\1 [...]\2/' ||
        true
)
if [[ "$results" != "" ]]; then
    echo "Some Actions are not referenced via the correct release tag \"$RELEASE_TAG\" in internal actions"
    echo "$results"
    exit 1
fi

# Verify our Actions are referenced by the release tag in external actions.
results=$(
    find actions/ -maxdepth 3 -name '*.yaml' -o -name '*.yml' -type f -print0 |
        xargs -0 grep -Pn "slsa-framework/slsa-github-generator/.*@(?!$RELEASE_TAG)" |
        sed 's/\(.*:\) *uses:.*\(\/.*\)/\1 [...]\2/' ||
        true
)
if [[ "$results" != "" ]]; then
    echo "Some Actions are not referenced via the correct release tag \"$RELEASE_TAG\" in external actions"
    echo "$results"
    exit 1
fi

if [[ "$RELEASE_TAG" =~ .*-rc\.[0-9]*$ ]]; then
    # don't check documentation for release candidates
    exit 0
fi

# Verify documentation refers to the most recent release tag
results=$(
    find . -name "*.md" -print0 |
        xargs -0 grep -Pn "uses: slsa-framework/slsa-github-generator/.*@(?!<|$RELEASE_TAG)" |
        sed "s/\(.*:\) *uses:.*\(\/.*\)/\1 [...]\2/" ||
        true
)

if [[ "$results" != "" ]]; then
    echo "Some documentation refers to an incorrect release tag"
    echo "Allowed tags are \"<pseudo_tags>\" or \"$RELEASE_TAG\""
    echo "$results"
    exit 1
fi
