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

# TODO(Issue #2331): switch copy to binaries to a temp dir
mkdir binaries

# Transfer flags and targets to their respective arrays
IFS=' ' read -r -a build_flags <<< "${FLAGS}"
IFS=' ' read -r -a build_targets <<< "${TARGETS}"

# If the targets includes Java targets, include Java build flag
# and add Github Runner Java rule to WORKSPACE
if [[ "${INCLUDES_JAVA}" == "true" ]]
then
  build_flags+=("--java_runtime_version=myjdk")

  java_rule="local_java_repository(
    name = \"myjdk\",
    java_home = \"$JAVA_HOME\",
  )"

  # Echo the configuration for the local Github Runner java into the root WORKSPACE file.
  echo "load(\"@bazel_tools//tools/jdk:local_java_repository.bzl\", \"local_java_repository\")" >> ./WORKSPACE
  echo "$java_rule" >> ./WORKSPACE
fi

# Use associative array as a set to increase efficency in avoiding double copying the target.
# Useful in generic glob patterns.
declare -A targets_set

################################################
#                                              #
#              Build Target Set                #
#                                              #
################################################

for input in "${build_targets[@]}"; do
  
  # Using bazel query extracts all targets from a glob pattern.
  # Thus we can change Java targets to their _deploy.jar target.
  for target in $(bazel query "$input"); do

    # Check to see if the target is a Java target. If it is the output is a Java target.
    # Note: targets that already have the _deploy.jar suffix will have no output from the query
    output=$(bazel query "kind(java_binary, $target)" 2>/dev/null)
    
    # If there is a Java target without _deploy.jar suffix, add suffix, build and add to target set.
    if [[ -n "$output" ]]
    then
      bazel build "${build_flags[@]}" "${target}_deploy.jar"
      targets_set["${target}_deploy.jar"]="1"    
    else
      # Build target regularly.
      bazel build "${build_flags[@]}" "$target"
      targets_set["$target"]="1" # A set of unique targets
    fi
  done
done

################################################
#                                              #
#    Copy Needed Artifacts To Binaries Dir     #
#                                              #
################################################

for curr_target in "${!targets_set[@]}"; do
  
  # Removes everything up to and including the first colon
  # "//src/internal:fib" --> "fib"
  binary_name=${curr_target#*:}

  ################################################
  #                                              #
  #            Logic for Java Targets            #
  #                                              #
  ################################################

  # If the target name includes _deploy.jar it is a Java target.
  if [[ "$binary_name" == *"_deploy.jar"* ]]
  then
    # Uses _deploy.jar as a field seperator and grabs the field before it.
    run_script_name=$(echo "$binary_name" | awk -F'_deploy.jar' '{print $1}')

    # Create dir for artifact and its runfiles
    mkdir "./binaries/$run_script_name"

    # Get the absolute path to output of Java JAR artifact.
    bazel_generated=$(bazel cquery --output=starlark --starlark:expr="'\n'.join([f.path for f in target.files.to_list()])" "$curr_target" 2>/dev/null)

    # Copy JAR to artifact-specific dir in ./binaries and remove symbolic links.
    file="$bazel_generated"
    cp -Lr "$file" "./binaries/$run_script_name"
    
    # Get the path the to run-script associated with the {$curr_target}_deploy.jar
    # If the user inputted the path to their local JAVABIN insert that into the run-script to define it.
    # Inputting a local path to JAVABIN is needed or else run-script will not work as it points to Github Runner JAVABIN
    run_script_path=$(echo "$file" | awk -F'_deploy.jar' '{print $1}')

    # This adds an additional flag to the the run-script for the Java target which sets the Java bin
    # to the user input. This allows users that download the binaries from the Github workflow to be able
    # to run the run-script themselves, which would not be possible as it is either set to the Github Runner VM Java bin path
    # if no flag to USER_LOCAL_JAVABIN is passed in their workflow or to the path passed in their flag.
    awk -v n=66 -v s='    --local_javabin=*) USER_JAVA_BIN="${1#--local_javabin}" ;;' 'NR == n {print s} {print}' "$run_script_path" > temp_file && mv -f temp_file "$run_script_path"
    
    # Updates Java Bin in run-script after the flags get proccessed
    awk -v n=127 -v s='' 'NR == n {print s} {print}' "$run_script_path" > temp_file && mv -f temp_file "$run_script_path"
    awk -v n=128 -v s='if [[ -n $USER_JAVA_BIN ]]; then JAVABIN=$USER_JAVA_BIN; fi' 'NR == n {print s} {print}' "$run_script_path" > temp_file && mv -f temp_file "$run_script_path"
    
    cp -L "$run_script_path" "./binaries/$run_script_name"

  ################################################
  #                                              #
  #          Logic for Non-Java Targets          #
  #                                              #
  ################################################
  
  else

    ################################################
    #                                              #
    #       Logic for Runfile-Needing Targets      #
    #                                              #
    ################################################

    if [[ "${NEEDS_RUNFILES}" == "true" ]]
    then
      # Get file(s) generated from build with respect to the target
      bazel_generated=$(bazel cquery --output=starlark --starlark:expr="'\n'.join([f.path for f in target.files.to_list()])" "$curr_target" 2>/dev/null)

      # Create dir for artifact and its runfiles
      mkdir "./binaries/$binary_name"

      # Uses a Starlark expression to pass new line seperated list of file(s) into the set of files
      while read -r path_to_artifact; do

        # Copy generated artifact from absolute path from bazel cquery
        cp -L "$path_to_artifact" "./binaries/$binary_name"

        # if runfiles dir exists, copy runfiles into artifact's dir
        if [[ -d "${path_to_artifact}.runfiles" ]]
        then
          path_to_target_runfiles="${path_to_artifact}.runfiles"
          cp -Lr "$path_to_target_runfiles" "./binaries/$binary_name"
          cd "./binaries/$binary_name/$binary_name.runfiles/"

          # Unneeded and can contain unwanted symbolic links
          rm -rf _main/external
          rm -rf MANIFEST
          rm -rf _repo_mapping

          # Go back to the old dir
          cd -
        fi
      done <<< "$bazel_generated"

    ################################################
    #                                              #
    #   Logic for NON-Runfile-Needing Targets      #
    #                                              #
    ################################################
    else
      # Get file(s) generated from build with respect to the target
      bazel_generated=$(bazel cquery --output=starlark --starlark:expr="'\n'.join([f.path for f in target.files.to_list()])" "$curr_target" 2>/dev/null)
      
      # Uses a Starlark expression to pass new line seperated list of file(s) into the set of files
      while read -r file; do
          cp -L "$file" ./binaries
      done <<< "$bazel_generated"
    fi
  fi
done
