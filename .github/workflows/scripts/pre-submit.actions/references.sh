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
# use the same referemnce as the relealse tag.
# See reasoning in ./github/actions/README.md

set -euo pipefail

cd __THIS_REPO__

# Verify our Actions are referenced at main in workflows.
results=$(
    find .github/workflows/ -maxdepth 1 -name '*.yaml' -o -name '*.yml' -type f -print0 |
        xargs -0 grep -P "slsa-framework/slsa-github-generator/.*@(?!main)" ||
        true
)
if [[ "$results" != "" ]]; then
    echo "Some Actions are not referenced at main"
    echo "$results"
    exit 1
fi

# Verify our Actions are referenced at main in internal actions.
results=$(
    find .github/actions/ -maxdepth 2 -name '*.yaml' -o -name '*.yml' -type f -print0 \
        xargs -0 grep -P "slsa-framework/slsa-github-generator/.*@(?!main)" ||
        true
)
if [[ "$results" != "" ]]; then
    echo "Some Actions are not referenced at main in internal Actions"
    echo "$results"
    exit 1
fi

# Verify our Actions are referenced at main in external actions.
results=$(
    find actions/ -maxdepth 3 -name '*.yaml' -o -name '*.yml' -type f -print0 |
        xargs -0 grep -Pn "slsa-framework/slsa-github-generator/.*@(?!main)" ||
        true
)
if [[ "$results" != "" ]]; then
    echo "Some Actions are not referenced at main in external Actions"
    echo "$results"
    exit 1
fi
