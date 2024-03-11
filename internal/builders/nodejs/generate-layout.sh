#!/bin/bash -eu
#
# Copyright 2023 SLSA Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

# We will encode the subject name as an npm package url (purl).
# https://github.com/package-url/purl-spec/blob/master/PURL-SPECIFICATION.rst
#
# The npm package's scope is considered a purl "namespace" and not part of the
# package name. So the subject will take the form of:
#
# With scope:
#   pkg:npm/<scope>/<name>@<version>
#
# Without scope:
#   pkg:npm/<name>@<version>
#
# Each of scope, name, and version are URL(percent) encoded.

# Get the raw package name and scope from the output of `npm pack --json`
# This name is of the form '<scope>/<package name>'
# NOTE: `cut -s` will suppress output for all fields if the delimiter is
# not present.
raw_package_scope=$(echo "${PACKAGE_NAME:-}" | cut -s -d'/' -f1)
raw_package_name=$(echo "${PACKAGE_NAME:-}" | cut -s -d'/' -f2)
if [ "${raw_package_name}" == "" ]; then
    # This is a non-scoped package.
    raw_package_name="${PACKAGE_NAME:-}"
    raw_package_scope=""
fi
# package scope (namespace) is URL(percent) encoded.
package_scope=$(echo "\"${raw_package_scope}\"" | jq -r '. | @uri')
# package name is URL(percent) encoded.
package_name=$(echo "\"${raw_package_name}\"" | jq -r '. | @uri')
# version is URL(percent) encoded. This is the version from the project's
# package.json and could be a commit, or any string by the user. It does not
# actually have to be a version number and is not validated as such by npm.
package_version=$(echo "\"${PACKAGE_VERSION:-}\"" | jq -r '. | @uri')

package_id="${package_name}@${package_version}"
if [ "${package_scope}" != "" ]; then
    package_id="${package_scope}/${package_id}"
fi
subject_name="pkg:npm/${package_id}"

# The integrity digest is formatted as follows:
#
# <hash alg>-<base64 encoded checksum>
#
# For example:
#   sha512-geEornsf879/Ygi9byQq/mpYboMcIKiGUxJ+RgHM3DCxqnOx15ttF5FparP/ZSITHTLM39MWVhW9qPa4XxtuSg==
integrity_digest="${PACKAGE_INTEGRITY:-}"

# We will parse out the checksum hash algorithm used.
# NOTE: ensure lowercase just to make sure.
alg=$(echo "${integrity_digest}" | cut -d'-' -f1 | tr '[:upper:]' '[:lower:]')
# Here we parse out the checksum and convert it from base64 to hex. 'od' seems
# to be the standard tool to do this kind conversion on Linux.
digest=$(echo "${integrity_digest}" | cut -d'-' -f2- | base64 -d | od -A n -v -t x1 | tr -d ' \n')

# NOTE: the name of the attestation should be configurable.
filename=${PACKAGE_FILENAME:-}
attestation_name="${filename%.*}"
cat <<EOF | jq | tee "${SLSA_OUTPUTS_ARTIFACTS_FILE}"
{
  "version": 1,
  "attestations":
  [
    {
      "name": "${attestation_name}",
      "subjects":
      [
        {
          "name": "${subject_name}",
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

echo "attestation-name=${attestation_name}.build.slsa" >>"${GITHUB_OUTPUT}"
