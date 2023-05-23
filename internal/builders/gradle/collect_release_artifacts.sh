#!/bin/bash

mkdir release-files-for-slsa
GRADLE_VERSION=$(./gradlew properties -q | grep "version:" | awk '{print $2}')

# Move artifacts from the user-supplied artifact list
IFS=',' read -ra ARTIFACT_ARRAY <<< "$ARTIFACT_LIST"
for i in "${ARTIFACT_ARRAY[@]}"
do
    i="${i#"${i%%[![:space:]]*}"}" # trim leading whitespace                                                                                                                                                                                                                                                                                                                                         
    i="${i%"${i##*[![:space:]]}"}" # trim trailing whitespace
    i=$(echo $i | sed 's/\"//')
    # Check if $i is empty
    if [ -z "$i" ]
    then
        continue
    fi
    
    # Replace 'GRADLE_VERSION' with $GRADLE_VERSION
    path_with_version="${i//GRADLE_VERSION/"$GRADLE_VERSION"}"

    # Move the file
    bn=$(basename -- $path_with_version)
    mv $path_with_version release-files-for-slsa/$bn
done
