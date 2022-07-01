#!/usr/bin/env bash

set -euo pipefail

# Verify that no internal Actions are using `actions/checkout`
# See reasoning in ./github/actions/README.md

results=$(grep -R --include='*.yml' --include='*.yaml' -e 'actions/checkout@\|actions/checkout-go@' .github/actions/* | grep -v 'checkout-go\|generate-builder')
if [[ "$results" != "" ]]; then
    echo "Some Actions are using 'actions/checkout'"
    echo "$results"
    exit -1
fi