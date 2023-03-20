#!/bin/bash

set -euo pipefail

package_scope=$(echo "${PACK_JSON}" | jq -r '.[0].name' | cut -d'/' -f1)
package_name=$(echo "${PACK_JSON}" | jq -r '.[0].name' | cut -d'/' -f2)
if [ "${package_name}" == "" ]; then
	package_name="${package_scope}"
	package_scope=""
fi
# NOTE: npm URI encodes package scope in the provenance.
package_scope=$(echo "\"${package_scope}\"" | jq -r '. | @uri')

package_version=$(echo "${PACK_JSON}" | jq -r '.[0].version')

package_id="${package_name}@${package_version}"
if [ "${package_scope}" != "" ]; then
	package_id="${package_scope}/${package_id}"
fi

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
attestation_name="attestation.intoto"
cat <<EOF | jq | tee "$SLSA_OUTPUTS_ARTIFACTS_FILE"
{
  "version": 1,
  "attestations":
  [
    {
      "name": "${attestation_name}",
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

echo "attestation-name=${attestation_name}" >>"$GITHUB_OUTPUT"
