#!/usr/bin/env bash
set -euo pipefail

if [[ ! -f "$UNTRUSTED_PATH" ]]; then
    echo "File $UNTRUSTED_PATH not present"
    exit 5
fi

DIGEST=$(sha256sum "$UNTRUSTED_PATH" | awk '{print $1}')

echo "::set-output name=$OUTPUT_NAME::$DIGEST"