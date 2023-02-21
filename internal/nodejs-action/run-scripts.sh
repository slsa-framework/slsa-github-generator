#!/bin/bash

set -euo pipefail

run_scripts="$1"

for script in $run_scripts; do
    echo "** Running 'npm run $script' **"
    npm run "$script"
    echo
done
