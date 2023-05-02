#!/bin/bash

set -euo pipefail

actual_diff=$(
    diff .github/workflows/delegator_generic_slsa3.yml .github/workflows/delegator_lowperms-generic_slsa3.yml \
    || true
)
expected_diff=$(cat .github/workflows/scripts/pre-submit.delegators/expected.diff)

if [[ "$expected_diff" != "$actual_diff" ]]; then
    echo "Unexpected differences between the delegator workflows"
    echo "$actual_diff"
    exit 1
fi
