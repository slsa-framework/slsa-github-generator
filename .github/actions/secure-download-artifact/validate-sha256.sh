#!/usr/bin/env bash
set -euo pipefail

if [[ ! -f "$UNTRUSTED_PATH" ]]; then
    echo "File $UNTRUSTED_PATH not present"
    exit 5
fi
    
echo "expected hash is $UNTRUSTED_EXPECTED_HASH"
echo "computed hash is $UNTRUSTED_COMPUTED_HASH"

if [[ "$UNTRUSTED_COMPUTED_HASH" != "$UNTRUSTED_EXPECTED_HASH" ]]; then
    echo "hashes do not match"
    exit -2
fi

echo "hashes match"

if [[ "$SET_EXECUTABLE" == "true" ]]; then
    echo "Setting $UNTRUSTED_PATH as executable"
    chmod u+x "$UNTRUSTED_PATH"
fi

if [[ -n "$UNTRUSTED_DST_PATH" ]]; then
    echo "Moving $UNTRUSTED_PATH to $UNTRUSTED_DST_PATH"
    mv "$UNTRUSTED_PATH" "$UNTRUSTED_DST_PATH"
fi
