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

# "version" and "attestations" fields:
echo -e -n "{\n  \"version\": 1,\n  \"attestations\": [" >> "$SLSA_OUTPUTS_ARTIFACTS_FILE"

num_binary_files=$(find ./binaries -type f | wc -l)
counter=1
        
# Add one attestation per binary:
find ./binaries -type f -print0 | while read -r -d $'\0' fname
do
          
    bn=$(basename -- "$fname")
    hash=$(sha256sum "$fname" | awk '{print $1}')
          
    echo -n "
        {
          \"name\": \"${bn}.intoto\",
          \"subjects\": [
            { \"name\": \"${bn}\",
              \"digest\": { \"sha256\": \"${hash}\"  }
            }
          ]
        }" >> "$SLSA_OUTPUTS_ARTIFACTS_FILE"
          
    # Add comma between attestations and not after the last
    if [[ "$counter" != "$num_binary_files" ]]; then
      echo -n "," >> "$SLSA_OUTPUTS_ARTIFACTS_FILE"
    fi
         
    counter="$((counter +1))"
done
        
# Close "attestations" and "version":
echo -e "\n  ]" >> "$SLSA_OUTPUTS_ARTIFACTS_FILE"
echo "}" >> "$SLSA_OUTPUTS_ARTIFACTS_FILE"
cat "$SLSA_OUTPUTS_ARTIFACTS_FILE"
