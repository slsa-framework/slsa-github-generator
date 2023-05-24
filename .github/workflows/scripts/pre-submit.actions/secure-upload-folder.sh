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

compare_trees() {
    local upload_tree="$1"
    local download_tree="$2"
    if [[ "${upload_tree}" != "${download_tree}" ]]; then
        echo "Folder trees differ: ${upload_tree} != ${download_tree}"
        echo "Upload tree: ${upload_tree}"
        echo "Download tree: ${download_tree}"
        exit 1
    fi
}

original_wd="$PWD"

download_folder="$1"
upload_folder="$2"

tree "${download_folder}"
cd "${upload_folder}"
upload_tree=$(tree .)
cd "${original_wd}"
cd "${download_folder}/${upload_folder}"
download_tree=$(tree .)

compare_trees "${upload_tree}" "${download_tree}"
