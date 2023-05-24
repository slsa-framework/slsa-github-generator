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

# Verify that no internal Actions are using `actions/checkout`
# See reasoning in ./github/actions/README.md

set -euo pipefail

# NOTE: All actions and workflows should not use actions/checkout. They should
# use an action that uses secure-checkout such as checkout-go, checkout-node
# etc. or use secure-checkout directly.
# TODO(github.com/slsa-framework/slsa-github-generator/issues/626): Check workflows as well and not just actions.
# TODO(github.com/slsa-framework/slsa-github-generator/issues/626): Disallow checkouts for repos other than the repo that triggered the action(i.e. github.repository).
results=$(
    grep -r \
        --include='*.yml' \
        --include='*.yaml' \
        --exclude-dir='node_modules' \
        --exclude-dir='secure-project-checkout' \
        --exclude-dir='secure-builder-checkout' \
        -e 'uses: *actions/checkout' \
        .github/actions/* || true
)
if [[ "$results" != "" ]]; then
    echo "Some Actions are using 'actions/checkout'"
    echo "$results"
    exit 1
fi
