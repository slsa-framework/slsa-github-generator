#!/bin/bash -eu
#
# Copyright 2023 SLSA Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

mkdir binaries

# Transfer flags and targets to their respective arrays
IFS=' ' read -r -a build_flags <<< "${FLAGS}"
IFS=' ' read -r -a build_targets <<< "${TARGETS}"

if [[ "${INCLUDES-JAVA}" ]]
then
  build_flags+="--java_runtime_version=myjdk"
  echo "$JAVA_HOME"
  echo "$PWD"
fi

# Build with respect to entire arrays of flags and targets
bazel build "${build_flags[@]}" "${build_targets[@]}"

# Use associative array as a set to increase efficency in avoiding double copying the target
declare -A files_set
declare -A targets_set

# Create folders for each artifact and include the runfiles in that dir
if [[ "${NEEDS-RUNFILES}" ]]
then
  for curr_target in "${build_targets[@]}"; do
    # Removes everything up to and including the first colon
    # "//src/internal:fib" --> "fib"
    binary_name=${curr_target#*:}
    
    # Logic for Java Targets
    if [[ "$binary_name" == *"_deploy.jar"* ]] 
    then
      # Create dir for artifact and its runfiles
      mkdir "./binaries/$binary_name"
    
      # Two outputs are binary_name.jar and binary_name.sh
      bazel_generated=$(bazel cquery --output=starlark --starlark:expr="'\n'.join([f.path for f in target.files.to_list()])" "$curr_target" 2>/dev/null)
      
      # Uses a Starlark expression to pass new line seperated list of file(s) into the set of files
      while read -r file; do
        # Key value is target path, value we do not care about and is set to constant "1"
        cp -L "$file" "./binaries/$binary_name"
      done <<< "$bazel_generated"

      #Add the artifact & runfile dir to set of files
      files_set["./binaries/$binary_name"]="1"
      
    else # Logic for other targets needing runfiles

      # Run cquery twice and transfer targets in case of generic glob patterns such as "//src/..." that intersect with other targets
      bazel_generated=$(bazel cquery --output=starlark --starlark:expr="'\n'.join([f.path for f in target.files.to_list()])" "$curr_target" 2>/dev/null)
      
      # Uses a Starlark expression to pass new line seperated list of targets into the set of targets
      while read -r target; do
        # Key value is target path, value we do not care about and is set to constant "1"
        targets_set["${target}"]="1"
      done <<< "$bazel_generated"  
    fi
  done
  
else # Upload the artfiacts as standalone pieces, no folders

  # Using target string, copy artifact(s) to binaries dir
  for curr_target in "${build_targets[@]}"; do
    # Get file(s) generated from build with respect to the target
    bazel_generated=$(bazel cquery --output=starlark --starlark:expr="'\n'.join([f.path for f in target.files.to_list()])" "$curr_target" 2>/dev/null)
    
    # Uses a Starlark expression to pass new line seperated list of file(s) into the set of files
    while read -r file; do
      # Key value is target path, value we do not care about and is set to constant "1"
      files_set["${file}"]="1"
    done <<< "$bazel_generated"
  done
  
fi

# From runfile loops --> unique targets and runfiles to copy
if [[ "${#targets_set[@]}" -ne 0 ]]
then
  for unique_target in "${!targets_set[@]}"; do
    # Removes everything up to and including the first colon
    # "//src/internal:fib" --> "fib"
    binary_name=${unique_target#*:}
    
    # Create dir for artifact and its runfiles
    mkdir "./binaries/$binary_name"

    # Copy the target and runfiles
    target_path=$(bazel cquery --output=starlark --starlark:expr="'\n'.join([f.path for f in target.files.to_list()])" "$curr_target" 2>/dev/null)
    cp -L target_path "./binaries/$binary_name"
    
    if [[ -d (target_path+=".runfiles") ]]
    then
      target_runfiles=$target_path+=".runfiles"
      cp -Lr "$target_runfiles" "./binaries/$binary_name"
      cd "./binaries/$binary_name/$binary_name.runfiles/"
      
      # Unneeded and can contain unwanted symbolic links
      rm -rf _main/external
      rm -rf MANIFEST
      rm -rf _repo_mapping  
      
      # Go back to the old dir
      cd -
  done
fi

# Copy set of unique targets to binaries. Without !, it would give values not keys
# TODO(Issue #2331): switch copy to binaries to a temp dir
for file in "${!files_set[@]}"; do
  # Remove the symbolic link and copy
  cp -L "$file" ./binaries
done
