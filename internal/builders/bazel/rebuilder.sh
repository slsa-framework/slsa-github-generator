#!/bin/bash
set -euo pipefail

# Notes:
#    Current constraint: make the users download and setup the specific version of Java
#                     that the workflow uses in the meantime of coming up with automation solution
#                     (if one is needed).
#                      
#                     Update: I think this can be solved with new docker image input ^

# This directory is where the rebuilt artifacts will be stored. It is made upon
# running the rebuilder.
rebuilt_artifacts_dir="rebuilt_artifacts_0ffe97cd2693d6608f5a787151950ed8"
mkdir $rebuilt_artifacts_dir
################################################
#                                              #
#             Process Arguments                #
#                                              #
################################################

# For later - after rebuild - to cd into binaries folder to obtain artifacts.
binaries_dir="bazel_builder_binaries_to_upload_to_gh_7bc972367cb286b7f36ab4457f06e369"

# Boolean that decides whether to use the slsa-verifier in addition to rebuild.
verify=false

# Boolean to trigger verbose version of Rebuilder.
verbose=false

# Outputs the usage of the Rebuilder script for the two modes:
# 1) Verify and Rebuild
# 2) Rebuild only
function usage() {
  if [[ $verify ]] 
  then
    echo "Usage to verify AND rebuild artifact is the following:"
    echo "Usage: $0 --artifact_path <path> --prov_path <path> --source_uri <uri> --builder_id <id> [--docker_image <image>] [--verify]"
    echo "To ONLY rebuild use the following usage:"
    echo "Usage: $0 --artifact_path <path> --prov_path <path> [--docker_image <image>]"
  else
    echo "Usage to ONLY rebuild artifact is the following:"
    echo "Usage: $0 --artifact_path <path> --prov_path <path> [--docker_image <image>]"
    echo "To verify AND rebuild artifact use the following usage:"
    echo "Usage: $0 --artifact_path <path> --prov_path <path> --source_uri <uri> --builder_id <id> [--docker_image <image>] [--verify]"
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
    --docker_image=*) docker_image="${1#--docker_image=}" ;;
    --verify) verify=true ;;
    --verbose) verbose=true ;;
    *)
      return 1 ;;
  esac
  return 0
}

# Parse arguments sequentially to check for unrecognized arguments
for ARG in "$@"; do
  if ! process_argument "$ARG"; then
    echo "'$ARG' is unrecognized"
    echo "Usage: $0 --artifact_path <path> --prov_path <path> --source_uri <uri> [--builder_id <id>] [--docker_image <image>] [--verify]"
    exit 1
  fi
done

################################################
#                                              #
#        Check Usage and Output Verbose        #
#                                              #
################################################

# Check if mandatory arguments for rebuild are not empty
if [ -z "$artifact_path" ] || [ -z "$prov_path" ]; then
  echo "Mandatory arguments missing or empty"
  usage
  exit 1
fi

# Check if mandatory arguments for verification are not empty
if $verify && ([ -z "$source_uri" ] || [ -z "$builder_id" ]); then
  echo "Mandatory arguments for verification are missing or empty"
  usage
  exit 1
fi

# Print received arguments (optional)
if [[ $verbose ]]
then
  echo "Output Arguments:"
  echo "artifact_path: $artifact_path"
  echo "prov_path: $prov_path"
  echo "source_uri: $source_uri"
  
  if [ -n "$builder_id" ]; then
    echo "builder_id: $builder_id"
  fi
  
  if [ -n "$docker_image" ]; then
    echo "docker_image: $docker_image"
  fi

  if [ "$verify" = true ]; then
    echo "verify: $verify"
  fi

  if [ -n "$builder_id" ]; then
    echo "builder_id: $builder_id"
  fi

  if [ -n "$docker_image" ]; then
    echo "docker_image: $docker_image"
  fi
fi

################################################
#                                              #
#           Use Verifier (if --verify)         #
#                                              #
################################################

if [[ $verify ]]
then
  # Clone the slsa-verifier repository
  if [ -d "slsa-verifier" ]; then
    echo "The slsa-verifier repository is already cloned."
  else
    echo "The slsa-verifier repository is not cloned. Cloning..."
    git clone https://github.com/enteraga6/slsa-verifier
  fi

  # Change directory to the slsa-verifier directory
  cd slsa-verifier
  
  # Run SLSA Verifier on user inputs
  go run ./cli/slsa-verifier/ verify-artifact ../$artifact_path --provenance-path ../$prov_path --source-uri $source_uri --builder-id $builder_id
  
  cd ..
  rm -rf ./slsa-verifier
fi

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
done < <(cat $prov_path | jq -r '.dsseEnvelope.payload' | base64 -d | jq -r '.predicate.buildDefinition.externalParameters.inputs | to_entries | .[] | .key + "=" + (.value | @text)')

for key in "${!data[@]}"
do
    echo "$key = ${data[$key]}"
done

################################################
#                                              #
#                 Setup ENV Vars               #
#                                              #
################################################

# The name map will convert and export the key strings of inputs to
# match with the environment variables of the Bazel Builder build.sh
declare -A name_mapping
name_mapping["targets"]="TARGETS"
name_mapping["flags"]="FLAGS"
name_mapping["docker-image"]="DOCKER_IMAGE"

# Note: These boolean inputs are now dealed with as strings
name_mapping["includes-java"]="INCLUDES_JAVA" 
name_mapping["needs-runfiles"]="NEEDS_RUNFILES"

# Export the inputs for later use
for key in "${!data[@]}"; do
    # Check to see if the key is in name map before export as env var.
    if [[ ${name_mapping[$key]+_} ]]; then
        export "${name_mapping[$key]}"="${data[$key]}"
    fi
done

# Clone the source_uri repository to begin rebuild process
if [ -d "$repo_name" ]; then
  echo "The source repo is already cloned."
else
  echo "The source repository is not cloned. Cloning..."
  git clone https://$source_uri
fi

###################
#
# Enter the Repo
cd $repo_name

# Check to see if JAVA_HOME is set then empty to
# avoid triggering unbound variable error.
if [[ $INCLUDES_JAVA == 'true' ]]
then
  if [[ ! -v JAVA_HOME || -z "${JAVA_HOME}" ]]
  then
      # if JAVA_HOME is empty, set to jdk bin path from $(which java)
      if java_path=$(which java); then
          JAVA_HOME="$(dirname $(dirname "${java_path}"))"
          export JAVA_HOME
      # JAVA_HOME cannot be set automatically
      else
          echo "Java is not installed or it is not in system PATH"
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

if [[ -n "$DOCKER_IMAGE" ]]
then
    docker pull $DOCKER_IMAGE

    # Mount docker image on this directory as workdir to gain access to script env
    # TODO: Check to see if env vars need to be passed in.
    docker run --rm -v $PWD:/workdir -w workdir $DOCKER_IMAGE /bin/sh -c "./build.sh"
else
    # Run the build script locally without a docker image
    source ../build.sh
fi

# TODO: with java jars. Investigate current behavior and see if it is expected.
#       There might need to be a special edge to handle _deploy.jar targets
#       or java targets in general since they get transformed to java jars.  

# Obtain the name of the artifact
if [[ $artifact_path == */* ]]
then
    artifact_name=$(basename $artifact_path)
else
    artifact_name=$artifact_path
fi

# Compute the original checksum of the artifact to compare with Rebuilt.
orig_checksum=$(sha256sum $artifact_path | awk '{ print $1 }')
rebuilt_checksum=""
unset rebuilt_checksum # Makes sure it is empty before assigning.

# IF there are runfiles, the directory structure will be different.
# The binaries folder contains different directories for the its artifacts and
# the artifacts runfiles. Obtain the rebuilt binaries and copy them to the
# path at root before cleaning up and deleting the repo.
if [[ "${NEEDS_RUNFILES}" == "true" ]]
then
    cd $binaries_dir/$artifact_name
    rebuilt_checksum=$(sha256sum $artifact_name | awk '{ print $1 }')
    cp $artifact_name ./../../../rebuilt_artifacts_dir
    cd ../../../ && rm -rf $repo_name
else
    cd $binaries_dir
    rebuilt_checksum=$(sha256sum $artifact_name | awk '{ print $1 }')
    cp $artifact_name ./../../rebuilt_artifacts_dir
    cd ../../ && rm -rf $repo_name
fi

################################################
#                                              #
#       Check Build for Reproducibility        #
#                                              #
################################################

if [[ "$orig_checksum" == "$rebuilt_checksum" ]]
then
    echo "Checksum is the same for the original and rebuilt artifact"
    echo "This build is reproducible."
    echo "$orig_checksum = Original Checksum"
    echo "$rebuilt_checksum = Rebuilt Checksum"
else
    echo "Checksum is NOT the same for the original and rebuilt artifact"
    echo "This build is NOT reproducible"
    echo "$orig_checksum = Original Checksum"
    echo "$rebuilt_checksum = Rebuilt Checksum"
fi

