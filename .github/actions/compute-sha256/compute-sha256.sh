#!/usr/bin/env bash
set -euo pipefail

if [[ ! -f "$UNTRUSTED_PATH" ]]; then
    echo "File $UNTRUSTED_PATH not present"
    exit 5
fi

digest=$(sha256sum "$UNTRUSTED_PATH" | awk '{print $1}')

echo "::set-output name=sha256::$digest"
