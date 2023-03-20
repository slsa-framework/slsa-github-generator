#!/bin/bash

set -euo pipefail

if [ "${GITHUB_WORKSPACE}" == "" ]; then
    echo "\$GITHUB_WORKSPACE is empty."
    exit 1
fi

UNTRUSTED_REALPATH=$(realpath -e "${UNTRUSTED_DIRECTORY}")
echo "Directory '${UNTRUSTED_DIRECTORY}' resolved to '$UNTRUSTED_REALPATH'"

GITHUB_WORKSPACE_REALPATH=$(realpath -e "${GITHUB_WORKSPACE}")
echo "GitHub workspace '${GITHUB_WORKSPACE}' resolved to '${GITHUB_WORKSPACE_REALPATH}'"

echo "Checking directory '${UNTRUSTED_REALPATH}' is a sub-directory of '${GITHUB_WORKSPACE_REALPATH}'"
if [[ "${UNTRUSTED_REALPATH}" != "${GITHUB_WORKSPACE_REALPATH}" ]] && [[ ${UNTRUSTED_REALPATH} != ${GITHUB_WORKSPACE_REALPATH}/* ]]; then
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
