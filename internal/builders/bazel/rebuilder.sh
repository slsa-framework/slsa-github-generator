#!/bin/bash
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

# This directory is where the rebuilt artifacts will be stored. It is made upon
# running the rebuilder. The long name is to avoid potential collisions.
rebuilt_artifacts_dir="rebuilt_artifacts_0ffe97cd2693d6608f5a787151950ed8"
mkdir $rebuilt_artifacts_dir

################################################
#                                              #
#            Color Code Env Vars               #
#                                              #
################################################

RESET="\033[0m"
BOLD="\033[1m"
RED="\033[1;31m"
LIGHT_RED="\033[0;31m"
GREEN="\033[1;32m"
YELLOW="\033[1;33m"
BLUE="\033[1;34m"
CYAN="\033[1;36m"
MAGENTA="\033[1;35m"
PURPLE="\033[1;35m"
BOLD_RED_BG="\033[1;41m"
UNDERLINE="\033[4m"

################################################
#                                              #
#           To Output Styled Progress          #
#                                              #
################################################

type_speed=0.02
function type_writer {
  text="$1"

  for (( i=0; i<${#text}; i++ )); do
    echo -n "${text:$i:1}"
    sleep $type_speed
  done
  echo ""
}

################################################
#                                              #
#             Process Arguments                #
#                                              #
################################################

# For later - after rebuild - to cd into binaries folder to obtain artifacts.
# The long name is to avoid potential collisions.
binaries_dir="bazel_builder_binaries_to_upload_to_gh_7bc972367cb286b7f36ab4457f06e369"

# Boolean that decides whether to use the slsa-verifier in addition to rebuild.
verify=0

# Boolean to trigger verbose version of Rebuilder.
verbose=0

# Boolean to trigger cleanup upon completion or failure.
cleanup=0

# Outputs the usage of the Rebuilder script for the two modes:
# 1) Verify and Rebuild
# 2) Rebuild only
function usage() {
  if [[ $verify ]]; then
    echo -e "${RED}[ERROR] ${LIGHT_RED}Wrong usage. Usage to verify AND rebuild artifact:${RESET}"
    echo -e "${CYAN}Usage: $0 ${YELLOW}--artifact_path${RESET}=<path> ${YELLOW}--prov_path${RESET}=<path> ${YELLOW}--source_uri${RESET}=<uri> ${YELLOW}--builder_id${RESET}=<id> ${MAGENTA}[--env_image]${RESET}=<image> ${MAGENTA}[--verify]${RESET}"
    echo -e "${RED}[ERROR] ${LIGHT_RED}Wrong usage. Usage to ONLY rebuild the artifact:${RESET}"
    echo -e "${CYAN}Usage: $0 ${YELLOW}--artifact_path${RESET}=<path> ${YELLOW}--prov_path${RESET}=<path> ${MAGENTA}[--env_image]${RESET}=<image>"
  else
    echo -e "${RED}[ERROR] ${LIGHT_RED}Wrong usage. Usage to ONLY rebuild the artifact:${RESET}"
    echo -e "${CYAN}Usage: $0 ${YELLOW}--artifact_path${RESET}=<path> ${YELLOW}--prov_path${RESET}=<path> ${MAGENTA}[--env_image]${RESET}=<image>"
    echo -e "${RED}[ERROR] ${LIGHT_RED}Wrong usage. Usage to verify AND rebuild artifact:${RESET}"
    echo -e "${CYAN}Usage: $0 ${YELLOW}--artifact_path${RESET}=<path> ${YELLOW}--prov_path${RESET}=<path> ${YELLOW}--source_uri${RESET}=<uri> ${YELLOW}--builder_id${RESET}=<id> ${MAGENTA}[--env_image]${RESET}=<image> ${MAGENTA}[--verify]${RESET}"
  fi
}

# Processes an argument for the script. Returns 0 if the given argument
# was recognized as an argument for this script, and 1 if it was not.
function process_argument() {
  case "$1" in
    --artifact_path=*) artifact_path="${1#--artifact_path=}" ;;
    --prov_path=*) prov_path="${1#--prov_path=}" ;;
    --source_uri=*) source_uri="${1#--source_uri=}" ;;
    --builder_id=*) builder_id="${1#--builder_id=}" ;;
    --env_image=*) env_image="${1#--env_image=}" ;;
    --verify) verify=1 ;;
    --verbose) verbose=1 ;;
    --cleanup) cleanup=1 ;;

    *)
      return 1 ;;
  esac
  return 0
}

# This function is will clean up built directories off after error.
function cleanup() {
  # If the cleanup flag is specified, remove created directories.
  if [[ $cleanup -eq 1 ]]
  then
    type_writer "🧹---> Cleaning up $rebuilt_artifacts_dir..."
    rm -rf $rebuilt_artifacts_dir

    type_writer "🧹---> Cleaning up $repo_name..."
    sudo rm -rf "$repo_name"

    type_writer "🧹---> Cleaning up slsa-verifier..."
    sudo rm -rf slsa-verifier
  fi
}

# Parse arguments sequentially to check for unrecognized arguments
for ARG in "$@"; do
  returnValue=$?
  process_argument "$ARG"
  if [[ ! ($returnValue) ]]; then
    my_arg="$ARG"
    echo -e "${RED}[ERROR] ${LIGHT_RED}$my_arg is unrecognized${RESET}"
    usage
    exit 1
  fi
done

################################################
#                                              #
#        Check Usage and Output Verbose        #
#                                              #
################################################

# Check if mandatory arguments for rebuild are not empty
if [ -z "${artifact_path:-}" ]; then
  echo -e "${RED}[ERROR] ${LIGHT_RED}Mandatory argument for rebuild, --artifact_path, is missing or empty${RESET}"
  usage
  exit 1
fi

if [ -z "${prov_path:-}" ]; then
  echo -e "${RED}[ERROR] ${LIGHT_RED}Mandatory argument for rebuild, --prov_path, is missing or empty${RESET}"
  usage
  exit 1
fi

if [ -z "${source_uri:-}" ]; then
  echo -e "${RED}[ERROR] ${LIGHT_RED}Mandatory argument for rebuild, --source_uri, is missing or empty${RESET}"
  usage
  exit 1
fi

# Check if mandatory arguments for verification are not empty
if [[ $verify -eq 1 && ( -z ${source_uri:-} || -z ${builder_id:-} ) ]]; then
  echo -e "${RED}[ERROR] ${LIGHT_RED}Mandatory arguments for verification missing or empty${RESET}"
  usage
  exit 1
fi

# Print received arguments (optional)
if [[ $verbose -eq 1 ]]; then
    echo -e "${BLUE}✔ Input Arguments Received:${RESET}"
    echo -e "${CYAN}artifact_path: ${GREEN}$artifact_path${RESET}"
    echo -e "${CYAN}prov_path: ${GREEN}$prov_path${RESET}"
    echo -e "${CYAN}source_uri: ${GREEN}$source_uri${RESET}"

  if [[ -n ${builder_id:-} ]]; then
      echo -e "${CYAN}builder_id: ${GREEN}$builder_id${RESET}"
  fi

  if [[ -n ${env_image:-} ]]; then
      echo -e "${CYAN}env_image: ${GREEN}$env_image${RESET}"
  fi

  echo -e "${CYAN}verify: ${GREEN}$verify${RESET}"
  echo -e "${CYAN}verbose: ${GREEN}$verbose${RESET}"
  echo -e "${CYAN}cleanup: ${GREEN}$cleanup${RESET}"
  echo ""
fi

################################################
#                                              #
#           Use Verifier (if --verify)         #
#                                              #
################################################

if [[ $verify -eq 1 ]]; then
  # Clone the slsa-verifier repository
  if [ -d "slsa-verifier" ]; then
    type_writer "📁---> The slsa-verifier repository is already cloned."
    type_writer "⚠️---> To verify please remove the collision and try again"
    exit 1
  else
    echo -e "${CYAN}====================================================${RESET}"
    type_writer "📥---> The slsa-verifier repository is not cloned. Cloning..."
    git clone https://github.com/slsa-framework/slsa-verifier
  fi

  # Change directory to the slsa-verifier directory
  cd slsa-verifier

  # Run SLSA Verifier on user inputs
  # write if builder id then this if not include builder id then other command
  # this is for once the non-compulsory feature gets merged.
  go run ./cli/slsa-verifier/ verify-artifact ../"$artifact_path" --provenance-path ../"$prov_path" --source-uri "$source_uri" --builder-id "$builder_id"

  cd ..
  echo -e "${CYAN}====================================================${RESET}"
  echo ""
fi

# Compute the original checksum of the artifact to compare with Rebuilt.
orig_checksum=$(sha256sum "$artifact_path" | awk '{ print $1 }')

################################################
#                                              #
#               Parse Provenance               #
#                                              #
################################################

# Associative Array to store the inputs to the GH workflow in key:value form
declare -A data

# Extract the inputs and put them in data map where the key is the workflow input,
# and the value is the value that the user inputted. Pipe value to @text, to deal with booleans.
while IFS='=' read -r key value; do
    data["$key"]="$value"
done < <(<"$prov_path" jq -r '.dsseEnvelope.payload' | base64 -d | jq -r '.predicate.buildDefinition.externalParameters.inputs | to_entries | .[] | .key + "=" + (.value | @text)')

# Todo: Style Env Vars Later

if [[ $verbose -eq 1 ]]; then
  echo -e "${PURPLE}✔ Arguments Parsed from Provenance:${RESET}"
  for key in "${!data[@]}"; do
      echo -e "${MAGENTA}$key: ${GREEN}${data[$key]}${RESET}"
  done
  echo ""
fi

################################################
#                                              #
#                 Setup ENV Vars               #
#                                              #
################################################

# The name map will convert and export the key strings of inputs to
# match with the environment variables of the Bazel Builder build.sh
declare -A name_mapping
name_mapping["targets"]="UNTRUSTED_TARGETS"
name_mapping["flags"]="UNTRUSTED_FLAGS"
name_mapping["env-image"]="UNTRUSTED_DOCKER_IMAGE"

# Note: These boolean inputs are now dealed with as strings
name_mapping["includes-java"]="UNTRUSTED_INCLUDES_JAVA"
name_mapping["needs-runfiles"]="UNTRUSTED_NEEDS_RUNFILES"

# Export the inputs for later use
for key in "${!data[@]}"; do
    # Check to see if the key is in name map before export as env var.
    if [[ ${name_mapping[$key]+_} ]]; then
        export "${name_mapping[$key]}"="${data[$key]}"
    fi
done

################################################
#                                              #
#            Clone Repo to Rebuild             #
#                                              #
################################################

repo_name=$(basename "$source_uri")
# Clone the source_uri repository to begin rebuild process
if [ -d "$repo_name" ]; then
  echo -e "${CYAN}====================================================${RESET}"
  type_writer "📁---> Source repository appears already."
  type_writer "⚠️---> To run rebuilder, fix collision by removing directory with name of $repo_name."
  exit 1
else
  echo -e "${CYAN}====================================================${RESET}"
  type_writer "🐑---> Cloning the source repository..."
  echo ""
  git clone https://"$source_uri"
  echo -e "${CYAN}====================================================${RESET}"
  echo ""
fi

# Enter the Repo
cd "$repo_name"

# Check to see if JAVA_HOME is set then empty to
# avoid triggering unbound variable error.
if [[ "${UNTRUSTED_INCLUDES_JAVA}" == "true" ]]
then
    if [[ ! -v JAVA_HOME || -z "${JAVA_HOME}" ]]; then
        # if JAVA_HOME is empty, set to jdk bin path from $(which java)
        if java_path=$(which java); then
            JAVA_HOME="$(dirname "$(dirname "${java_path}")")"
            export JAVA_HOME
        # JAVA_HOME cannot be set automatically
        else
            echo "JAVA_HOME cannot be set automatically. Check PATH."
        fi
    else
        echo "JAVA_HOME already set to ${JAVA_HOME}"
    fi
fi

################################################
#                                              #
#              Rebuild the Artifacts           #
#                                              #
################################################

echo ""
echo -e "${CYAN}======================================================${RESET}"
echo -e "${CYAN}|${RESET}${YELLOW}${UNDERLINE}        🔨  Starting the Rebuild Process  🔨        ${RESET}${CYAN}|${RESET}"
echo -e "${CYAN}======================================================${RESET}"

# Conditionals for environment images depend on if a Docker Image was use to build on Github.
# If a Docker Image was not used to build on Github, then build locally. This is done to
# ensure consistent build environment between both platforms.
if [[ -n ${UNTRUSTED_DOCKER_IMAGE:-} ]]; then
    cd -
    sudo docker pull "$UNTRUSTED_DOCKER_IMAGE"
    echo ""
    echo -e "${CYAN}======================================================${RESET}"
    type_writer "🔨---> Rebuilding with Docker Image Environment..."    # Mount environment image on this directory as workdir to gain access to script env
    echo -e "${CYAN}======================================================${RESET}"
    echo ""

    sudo docker run --env repo_name="$repo_name" --env UNTRUSTED_TARGETS="${UNTRUSTED_TARGETS}" --env UNTRUSTED_FLAGS="${UNTRUSTED_FLAGS}" --env UNTRUSTED_NEEDS_RUNFILES="${UNTRUSTED_NEEDS_RUNFILES}" --env UNTRUSTED_INCLUDES_JAVA="${UNTRUSTED_INCLUDES_JAVA}" --rm -v "$PWD":/workdir -w /workdir "$UNTRUSTED_DOCKER_IMAGE" /bin/sh -c "cd $repo_name && ./../build.sh"
    echo ""
    echo -e "${CYAN}======================================================${RESET}"
    echo -e "${CYAN}|${RESET}${YELLOW}${UNDERLINE}        ✅  Artifacts Rebuilt! ✅          ${RESET}${CYAN}|${RESET}"
    echo -e "${CYAN}======================================================${RESET}"
    echo ""
else
    if [[ -n ${env_image:-} ]]; then
      # Warning message for the users if their artifact was not built with a Docker Image, but a Docker Image was provided at command.
      echo -e "${RED}[Warning] ${LIGHT_RED}Docker Image, $env_image, provided, but artifact was not originally built on Docker Image${RESET}"
    else
      echo "" # This is just for style.
    fi

    # Run the build script locally without an environment image.
    echo -e "${CYAN}======================================================${RESET}"
    type_writer "💻---> Rebuilding with local environment..."
    echo -e "${CYAN}======================================================${RESET}"
    echo ""

    # To fix shell check
    # shellcheck source=/dev/null
    source ../build.sh
    echo ""
    echo -e "${CYAN}======================================================${RESET}"
    echo -e "${CYAN}|${RESET}${YELLOW}${UNDERLINE}        ✅  Artifacts Rebuilt! ✅          ${RESET}${CYAN}|${RESET}"
    echo -e "${CYAN}======================================================${RESET}"
    echo ""
fi

# If Docker Image was used to build on Github, we need to cd into repo
# to access the binaries directory.
if [[ -n ${UNTRUSTED_DOCKER_IMAGE:-} ]]; then
  cd "$repo_name"
fi

################################################
#                                              #
#               Copy the Artifact              #
#                                              #
################################################

# Obtain the name of the artifact
if [[ $artifact_path == */* ]]; then
    artifact_name=$(basename "$artifact_path")
else
    artifact_name=$artifact_path
fi

rebuilt_checksum=""
unset rebuilt_checksum # Makes sure it is empty before assigning.

# IF there are runfiles, the directory structure will be different.
# The binaries folder contains different directories for the its artifacts and
# the artifacts runfiles. Obtain the rebuilt binaries and copy them to the
# path at root before cleaning up and deleting the repo.
if [[ "$artifact_name" == *"_deploy.jar"* ]]; then
      # Uses _deploy.jar as a field seperator and grabs the field before it.
      # Directory of Java artifacts is same as run script name.
      run_script_name=$(echo "$artifact_name" | awk -F'_deploy.jar' '{print $1}')
      cd $binaries_dir/
      rebuilt_checksum=$(sha256sum ./"$run_script_name"/"$artifact_name" | awk '{ print $1 }')

      # Copy the entire directory, including the run script.
      cp -R ./"$run_script_name" ./../../"$rebuilt_artifacts_dir"/
else
    if [[ "${NEEDS_RUNFILES}" == "true" ]]; then
        # For non-java targets with runfiles.
        cd $binaries_dir/
        rebuilt_checksum=$(sha256sum ./"$artifact_name"/"$artifact_name" | awk '{ print $1 }')

        # Copy entire directory, including the runfiles.
        cp -R ./"$artifact_name" ./../../"$rebuilt_artifacts_dir"/
    else
        # For files withouts runfiles.
        cd $binaries_dir
        rebuilt_checksum=$(sha256sum "$artifact_name" | awk '{ print $1 }')

        cp "$artifact_name" ./../../$rebuilt_artifacts_dir/
    fi
fi

################################################
#                                              #
#       Check Build for Reproducibility        #
#                                              #
################################################

if [[ "$orig_checksum" == "$rebuilt_checksum" ]]; then
    echo -e "${GREEN}Checksum is the ${BOLD}${UNDERLINE}same${RESET}${GREEN} for the original and rebuilt artifact!${RESET}"
    echo -e "${GREEN}✅ This build is ${BOLD}${UNDERLINE}reproducible!${RESET} ✅"
    echo ""
    echo -e "${GREEN}$orig_checksum${RESET} = Original Checksum"
    echo -e "${GREEN}$rebuilt_checksum${RESET} = Rebuilt Checksum"
    echo ""
  else
    echo -e "${BOLD_RED_BG}Checksum is ${BOLD}${UNDERLINE}NOT${RESET}${BOLD_RED_BG} the same for the original and rebuilt artifact!${RESET}"
    echo -e "${BOLD_RED_BG}        ⚠️  This build was ${BOLD}${UNDERLINE}NOT${RESET}${BOLD_RED_BG} able to be reproduced! ⚠️         ${RESET}"
    echo ""
    echo -e "${RED}$orig_checksum${RESET} = Original Checksum"
    echo -e "${RED}$rebuilt_checksum${RESET} = Rebuilt Checksum"
    echo ""
  fi


if [[ cleanup -eq 1 ]]; then

    # If there are runfiles or if the artifacts are Java, then each artifact
    # has its own directory, so you need to exit out of it first.

    # TODO TEST FUNCTIONALITY OF CLEANUP

    # if [[ "${NEEDS_RUNFILES}" == "true" || "${INCLUDES_JAVA}" == "true" ]]; then
    #   cd ..
    # fi

    # Current position is bazel_builder_dir/$repo_name/$binaries_dir,
    # and to clean up need to be in /bazel.
    cd ../..

    # Now cleanup of verifier and cloned $repo_name.
    cleanup
fi
