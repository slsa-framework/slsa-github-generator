#!/bin/bash

# Verify that no internal Actions are using `actions/checkout`
# See reasoning in ./github/actions/README.md

set -euo pipefail

# NOTE: All actions and workflows should not use actions/checkout. The should
# use an action that uses secure-checkout such as checkout-go, checkout-node
# etc. or use secure-checkout directly.
# TODO(github.com/slsa-framework/slsa-github-generator/issues/626): Check workflows as well and not just actions.
# TODO(github.com/slsa-framework/slsa-github-generator/issues/626): Disallow checkouts for repos other than the repo that triggered the action(i.e. github.repository).
results=$(
    grep -r \
        --include='*.yml' \
        --include='*.yaml' \
        --exclude-dir='node_modules' \
        --exclude-dir='secure-checkout' \
        --exclude-dir='checkout-go' \
        --exclude-dir='checkout-node' \
        -e 'uses: *actions/checkout' \
        .github/actions/* || true
)
if [[ "$results" != "" ]]; then
    echo "Some Actions are using 'actions/checkout'"
    echo "$results"
    exit 1
fi
