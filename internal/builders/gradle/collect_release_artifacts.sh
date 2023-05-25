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

mkdir release-files-for-slsa
GRADLE_VERSION=$(./gradlew properties -q | grep "version:" | awk '{print $2}')

# Move artifacts from the user-supplied artifact list
IFS=',' read -ra artifact_array <<< "$ARTIFACT_LIST"
for i in "${artifact_array[@]}"
do
    i="${i#"${i%%[![:space:]]*}"}" # trim leading whitespace                                                                                                                                                                                                                                                                                                                                         
    i="${i%"${i##*[![:space:]]}"}" # trim trailing whitespace
    # shellcheck disable=SC2001
    i=$(echo "$i" | sed 's/\"//') 
    # Check if $i is empty
    if [ -z "$i" ]
    then
        continue
    fi
    
    # Replace 'GRADLE_VERSION' with $GRADLE_VERSION
    path_with_version="${i//GRADLE_VERSION/"$GRADLE_VERSION"}"

    # Move the file
    bn=$(basename -- "$path_with_version")
    mv "$path_with_version" release-files-for-slsa/"$bn"
done
