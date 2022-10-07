#!/bin/bash

# Verify that no internal Actions are using `actions/checkout`
# See reasoning in ./github/actions/README.md

set -euo pipefail

# Split the command to ignore the `1` error `grep` returns when there is no match.
results=$(
    grep -r \
        --include='*.yml' \
        --include='*.yaml' \
        --exclude-dir='node_modules' \
        --exclude-dir='checkout-go' \
        --exclude-dir='checkout-node' \
        --exclude-dir='secure-checkout' \
        --exclude-dir='generate-builder' \
        -e 'actions/checkout\|actions/secure-checkout\|actions/checkout-go\|actions/checkout-node' \
        .github/actions/* || true
)
if [[ "$results" != "" ]]; then
    echo "Some Actions are using 'actions/checkout'"
    echo "$results"
    exit 1
fi
