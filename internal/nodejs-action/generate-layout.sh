#!/bin/bash

set -euo pipefail

package_id=$(echo "${PACK_JSON}" | jq -r '.[0].id')

# The integrity digest is formatted as follows:
#
# <hash alg>-<base64 encoded checksum>
#
# For example:
#   sha512-geEornsf879/Ygi9byQq/mpYboMcIKiGUxJ+RgHM3DCxqnOx15ttF5FparP/ZSITHTLM39MWVhW9qPa4XxtuSg==
integrity_digest=$(echo "${PACK_JSON}" | jq -r '.[0].integrity')

# We will parse out the checksum hash algorithm used.
# NOTE: ensure lowercase just to make sure.
alg=$(echo "${integrity_digest}" | cut -d'-' -f1 | tr '[:upper:]' '[:lower:]')
# Here we parse out the checksum and convert it to hex. 'od' seems to be the
# standard tool to do this kind conversion on Linux.
digest=$(echo "${integrity_digest}" | cut -d'-' -f2- | base64 -d | od -A n -v -t x1 | tr -d ' \n')

# NOTE: the name of the attestation should be configurable.
cat <<EOF | jq | tee "$SLSA_OUTPUTS_ARTIFACTS_FILE"
{
  "version": 1,
  "attestations":
  [
    {
      "name": "attestation.intoto",
      "subjects":
      [
        {
          "name": "pkg:npm/${package_id}",
          "digest":
          {
            "${alg}": "${digest}"
          }
        }
      ]
    }
  ]
}
EOF
