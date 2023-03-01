#!/bin/bash

set -euo pipefail

source "./.github/workflows/scripts/e2e-utils.sh"

# TODO(#1709): Add more token verification.
if [[ "$SLSA_TOKEN" == "" ]]; then
    echo "SLSA_TOKEN is empty."
    exit 1
fi

b64_token=$(echo -n "$SLSA_TOKEN" | cut -d '.' -f2)
decoded_token=$(echo "$b64_token" | base64 -d)

e2e_verify_decoded_token "$decoded_token"
