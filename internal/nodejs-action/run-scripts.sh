#!/bin/bash

set -euo pipefail

run_scripts="${UNTRUSTED_RUN_SCRIPTS}"

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

for script in $run_scripts; do
    echo "** Running 'npm run $script' **"
    npm run "$script"
    echo
done
