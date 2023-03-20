#!/bin/bash

set -euo pipefail

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
ls -lh .
echo "pack_json=$pack_json" >>"$GITHUB_OUTPUT"

filename=$(echo "$pack_json" | jq -r '.[0].filename')
# NOTE: Get the absolute path of the file since we could be in a subdirectory.
resolved_filename=$(realpath -e "$filename")
echo "filename=$resolved_filename" >>"$GITHUB_OUTPUT"
