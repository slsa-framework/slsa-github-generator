#!/bin/bash

set -euo pipefail

# "version" and "attestations" fields:
echo -e -n "{\n  \"version\": 1,\n  \"attestations\": [" >> "$SLSA_OUTPUTS_ARTIFACTS_FILE"

num_jar_files=$(find ./release-files-for-slsa -type f | wc -l)
counter=1
        
# Add one attestation per .jar file:
find ./release-files-for-slsa -type f -print0 | while read -r -d $'\0' fname
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
    if [[ "$counter" != "$num_jar_files" ]]; then
      echo -n "," >> "$SLSA_OUTPUTS_ARTIFACTS_FILE"
    fi
         
    counter=$[$counter +1]
done
        
# Close "attestations" and "version":
echo -e "\n  ]" >> "$SLSA_OUTPUTS_ARTIFACTS_FILE"
echo "}" >> "$SLSA_OUTPUTS_ARTIFACTS_FILE"
cat "$SLSA_OUTPUTS_ARTIFACTS_FILE"
