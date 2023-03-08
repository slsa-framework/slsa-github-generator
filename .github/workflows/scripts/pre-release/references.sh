#!/bin/bash

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

# Verify internal Actions are referenced by the release tag.
cd __THIS_REPO__
results=$(
    find .github/workflows/ -maxdepth 1 -type f -print0 -name '*.yaml' -o -name '*.yml' |
        xargs -0 grep -Pn "slsa-framework/slsa-github-generator/.github/actions/.*@(?!$RELEASE_TAG)" |
        sed 's/\(.*:\) *uses:.*\(\/.*\)/\1 [...]\2/' ||
        true
)
if [[ "$results" != "" ]]; then
    echo "Some Actions are not referenced via the correct release tag \"$RELEASE_TAG\""
    echo "$results"
    exit 1
fi

# Verify the internal calls to BYOB workflows are referenced by the release tag.
results=$(
    find .github/workflows/ -maxdepth 1 -type f -print0 -name '*.yaml' -o -name '*.yml' |
        xargs -0 grep -Pn "slsa-framework/slsa-github-generator/.github/workflows/delegator_generic_slsa3.yml@(?!$RELEASE_TAG)" |
        sed 's/\(.*:\) *uses:.*\(\/.*\)/\1 [...]\2/' ||
        true
)
if [[ "$results" != "" ]]; then
    echo "Some delegator workflows are not referenced via the correct release tag \"$RELEASE_TAG\""
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
