#!/usr/bin/env bash
#
# Note: not settinng set -e because grep return 1 when there is not match.

# Verify that no internal Actions are using `actions/checkout`
# See reasoning in ./github/actions/README.md

results=$(grep -r --include='*.yml' --include='*.yaml' -e 'actions/checkout@\|actions/checkout-go@' .github/actions/* | grep -v 'checkout-go\|generate-builder')
if [[ "$results" != "" ]]; then
    echo "Some Actions are using 'actions/checkout'"
    echo "$results"
    exit -1
fi