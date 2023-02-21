#!/bin/bash

set -euo pipefail

package_id=$(echo "${PACK_JSON}" | jq '.[0] .id')
digest=$(echo "${PACK_JSON}" | jq '.[0] .shasum')

# NOTE: the name of the attestation should be configurable.
cat <<EOF | jq
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
            "sha256": "${digest}"
          }
        }
      ]
    }
  ]
}
EOF
