#!/bin/bash

# "version" and "attestations" fields:
echo -e -n "{\n  \"version\": 1,\n  \"attestations\": [" >> "$SLSA_OUTPUTS_ARTIFACTS_FILE"

NUM_JAR_FILES=$(find ./release-files-for-slsa -type f | wc -l)
COUNTER=1
        
# Add one attestation per .jar file:
find ./release-files-for-slsa -type f -print0 | while read -d $'\0' fname
do
          
    bn=$(basename -- $fname)
    hash=$(sha256sum $fname | awk '{print $1}')
          
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
    if [[ "$COUNTER" != "$NUM_JAR_FILES" ]]; then
      echo -n "," >> "$SLSA_OUTPUTS_ARTIFACTS_FILE"
    fi
         
    COUNTER=$[$COUNTER +1]
done
        
# Close "attestations" and "version":
echo -e "\n  ]" >> "$SLSA_OUTPUTS_ARTIFACTS_FILE"
echo "}" >> "$SLSA_OUTPUTS_ARTIFACTS_FILE"
cat "$SLSA_OUTPUTS_ARTIFACTS_FILE"
