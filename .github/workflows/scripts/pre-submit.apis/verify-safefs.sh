#!/bin/bash

set -euo pipefail

# Verify that only safe file functions are used.
results=$(
    find .github/actions/ actions/ -maxdepth 3 -name '*.ts' -not -path "*/tscommon/*" -not -path "*/__tests__/*" -type f -print0 |
    xargs -0 grep '"fs"' ||
    true
)
if [[ "$results" != "" ]]; then
    echo "Some Actions are using unsafe file system functions"
    echo "$results"
    exit 1
fi
