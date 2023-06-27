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

echo "** Using the following npm version **"
npm version

if [ "${GITHUB_WORKSPACE}" == "" ]; then
    echo "\$GITHUB_WORKSPACE is empty."
    exit 1
fi

untrusted_realpath=$(realpath -e "${UNTRUSTED_DIRECTORY}")
echo "Directory '${UNTRUSTED_DIRECTORY}' resolved to '${untrusted_realpath}'"

github_workspace_realpath=$(realpath -e "${GITHUB_WORKSPACE}")
echo "GitHub workspace '${GITHUB_WORKSPACE}' resolved to '${github_workspace_realpath}'"

echo "Checking directory '${untrusted_realpath}' is a sub-directory of '${github_workspace_realpath}'"
if [[ "${untrusted_realpath}" != "${github_workspace_realpath}" ]] && [[ ${untrusted_realpath} != ${github_workspace_realpath}/* ]]; then
    echo "${UNTRUSTED_DIRECTORY} not a sub-directory of ${GITHUB_WORKSPACE}"
    exit 1
fi
# Directory was validated. Explicitly trust it.
directory="${UNTRUSTED_DIRECTORY}"

cd "${directory}"

# Remove tab, newlines, spaces from the scripts input.
run_scripts=$(echo "${UNTRUSTED_RUN_SCRIPTS//[$'\t\r\n ']/}" | tr "," "\n")

for script in $run_scripts; do
    echo "** Running 'npm run $script' **"
    npm run "$script"
    echo
done

echo "** Running 'npm pack' **"
pack_json=$(npm pack --json | tee pack.json | jq -c)
jq <pack.json

package_filename=$(echo "$pack_json" | jq -r '.[0].filename')
package_name=$(echo "${pack_json}" | jq -r '.[0].name')
package_version=$(echo "${pack_json}" | jq -r '.[0].version')
package_integrity=$(echo "${pack_json}" | jq -r '.[0].integrity')
if [ ! -f "${package_filename}" ]; then
    echo "** ${package_filename} not found. **"
    ls -lh
    # NOTE: Some versions of npm pack --json returns a filename that is incorrect
    # attempt to determine the name by converting the package name and version
    # into the filename '<namespace>-<name>-<version>.tgz'.
    package_name=$(cut -d "=" -f 2 <<<"$(npm run env | grep "npm_package_name")")
    package_filename="$(echo "${package_name}" | sed 's/^@//' | sed 's/\//-/g')-${package_version}.tgz"
    echo "** Trying ${package_filename}... **"
fi

# NOTE: Get the absolute path of the file since we could be in a subdirectory.
resolved_filename=$(realpath -e "${package_filename}")

{
    echo "file-path=${resolved_filename}"
    echo "package-filename=${package_filename}"
    echo "package-name=${package_name}"
    echo "package-version=${package_version}"
    echo "package-integrity=${package_integrity}"
}>>"$GITHUB_OUTPUT"
