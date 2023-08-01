#!/bin/bash
set -eo pipefail

# Notes:
#    Current constraint: make the users download and setup the specific version of Java
#                     that the workflow uses in the meantime of coming up with automation solution
#                     (if one is needed).
#
#                     Update: I think this can be solved with new docker image input ^

# This directory is where the rebuilt artifacts will be stored. It is made upon
# running the rebuilder. The long name is to avoid potential collisions.
rebuilt_artifacts_dir="rebuilt_artifacts_0ffe97cd2693d6608f5a787151950ed8"
mkdir $rebuilt_artifacts_dir

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

# Outputs the usage of the Rebuilder script for the two modes:
# 1) Verify and Rebuild
# 2) Rebuild only
function usage() {
  if [[ $verify ]]
  then
    printf "\033[1;31m[ERROR] \033[0;31mWrong usage. Usage to verify AND rebuild artifact:\033[0m\n"
    printf "\033[1;36mUsage: $0 \033[1;33m--artifact_path\033[0m <path> \033[1;33m--prov_path\033[0m <path> \033[1;33m--source_uri\033[0m <uri> \033[1;33m--builder_id\033[0m <id> \033[1;35m[--docker_image]\033[0m <image> \033[1;35m[--verify]\033[0m\n"
    printf "\033[1;31m[ERROR] \033[0;31mWrong usage. Usage to ONLY rebuild the artifact:\033[0m\n"
    printf "\033[1;36mUsage: $0 \033[1;33m--artifact_path\033[0m <path> \033[1;33m--prov_path\033[0m <path> \033[1;35m[--docker_image]\033[0m <image>\n"
  else
    printf "\033[1;31m[ERROR] \033[0;31mWrong usage. Usage to ONLY rebuild the artifact:\033[0m\n"
    printf "\033[1;36mUsage: $0 \033[1;33m--artifact_path\033[0m <path> \033[1;33m--prov_path\033[0m <path> \033[1;35m[--docker_image]\033[0m <image>\n"
    printf "\033[1;31m[ERROR] \033[0;31mWrong usage. Usage to verify AND rebuild artifact:\033[0m\n"
    printf "\033[1;36mUsage: $0 \033[1;33m--artifact_path\033[0m <path> \033[1;33m--prov_path\033[0m <path> \033[1;33m--source_uri\033[0m <uri> \033[1;33m--builder_id\033[0m <id> \033[1;35m[--docker_image]\033[0m <image> \033[1;35m[--verify]\033[0m\n"
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
    --verify) verify=1 ;;
    --verbose) verbose=1 ;;
    *)
      # echo "in"
      return 1 ;;
  esac
  # echo "not in"
  return 0
}

# This function is will clean up built directories off after error
function cleanup() {
  type_writer "🧹---> Cleaning up $rebuilt_artifacts_dir..."
  rm -rf $rebuilt_artifacts_dir
  
  if [[ -n $repo_name ]]
  then 
    type_writer "🧹---> Cleaning up $repo_name..."
    sudo rm -rf $repo_name
  fi
}

# Parse arguments sequentially to check for unrecognized arguments
for ARG in "$@"; do
  # echo $@
  # echo $ARG
  returnValue=$?
  process_argument $ARG
  # echo $returnValue
  if [[ !($returnValue) ]]
  then
    my_arg="$ARG"
    printf "\033[1;31m[ERROR] \033[0;31m$my_arg is unrecognized\033[0m\n"
    usage
    cleanup
    exit 1
  fi
done

################################################
#                                              #
#        Check Usage and Output Verbose        #
#                                              #
################################################

# Check if mandatory arguments for rebuild are not empty
if [ -z "$artifact_path" ]; then
  printf "\033[1;31m[ERROR] \033[0;31mMandatory argument for rebuild, --artifact_path, is missing or empty\033[0m\n"
  usage
  cleanup
  exit 1
fi

if [ -z "$prov_path" ]; then
  printf "\033[1;31m[ERROR] \033[0;31mMandatory argument for rebuild, --prov_path, is missing or empty\033[0m\n"
  usage
  cleanup
  exit 1
fi

if [ -z "$source_uri" ]; then
  printf "\033[1;31m[ERROR] \033[0;31mMandatory argument for rebuild, --source_uri, is missing or empty\033[0m\n"
  usage
  cleanup
  exit 1
fi

# Check if mandatory arguments for verification are not empty
echo $verify
if [[ $verify -eq 1 && ( -z "$source_uri" || -z "$builder_id" ) ]]
then
  printf "\033[1;31m[ERROR] \033[0;31mMandatory arguments for verification missing or empty\033[0m\n"
  usage
  cleanup
  exit 1
fi

# artifact_path="./fib"
# prov_path="./fib.build.slsa"
# source_uri="github.com/enteraga6/slsa-lvl3-provenance"
# Print received arguments (optional)
if [[ $verbose ]]
then
  printf "\033[1;34m✔ Output Arguments:\033[0m\n"
  printf "\033[1;36martifact_path: \033[0m\033[1;32m$artifact_path\033[0m\n"
  printf "\033[1;36mprov_path: \033[0m\033[1;32m$prov_path\033[0m\n"
  printf "\033[1;36msource_uri: \033[0m\033[1;32m$source_uri\033[0m\n"

  if [ -n "$builder_id" ]; then
    printf "\033[1;36mbuilder_id: \033[0m\033[1;32m$builder_id\033[0m\n"
  fi

  if [ -n "$docker_image" ]; then
    printf "\033[1;36mdocker_image: \033[0m\033[1;32m$docker_image\033[0m\n"
  fi

  if [ $verify -eq 1 ]; then
    printf "\033[1;36mverify: \033[0m\033[1;32m$verify\033[0m\n"
  fi
fi

################################################
#                                              #
#           To Output Styled Progress          #
#                                              #
################################################

TYPE_SPEED=0.03
function type_writer {
  text="$1"

  for (( i=0; i<${#text}; i++ )); do
    echo -n "${text:$i:1}"
    sleep $TYPE_SPEED
  done
  echo ""
}

################################################
#                                              #
#           Use Verifier (if --verify)         #
#                                              #
################################################

printf "\033[1;36m====================================================\033[0m\n"
printf "\033[1;36m|\033[0m\033[1;33m\033[4m        🔨  Starting the Rebuild Process  🔨        \033[0m\033[1;36m|\033[0m\n"
printf "\033[1;36m====================================================\033[0m\n"

if [[ $verify -eq 1 ]]
then
  # Clone the slsa-verifier repository
  if [ -d "slsa-verifier" ]; then
    type_writer "📁---> The slsa-verifier repository is already cloned."
    type_writer "⚠️---> To verify please remove the collision and try again"
    cleanup
    exit 1
  else
    printf "\033[1;36m====================================================\033[0m\n"
    type_writer "📥---> The slsa-verifier repository is not cloned. Cloning..."
    git clone https://github.com/enteraga6/slsa-verifier
  fi

  # Change directory to the slsa-verifier directory
  cd slsa-verifier
  
  # Run SLSA Verifier on user inputs
  # write if builder id then this if not include builder id then other command
  go run ./cli/slsa-verifier/ verify-artifact ../$artifact_path --provenance-path ../$prov_path --source-uri $source_uri --builder-id $builder_id

  cd ..
  printf "\033[1;36m====================================================\033[0m\n"
  type_writer "🧹---> Cleaning up slsa-verifier..."
  rm -rf ./slsa-verifier
  echo ""
fi

#################################################
# compute og sum

# Compute the original checksum of the artifact to compare with Rebuilt.
orig_checksum=$(sha256sum $artifact_path | awk '{ print $1 }')


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

# Todo: Style Env Vars Later

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

echo $source_uri
repo_name=$(basename "$source_uri")
echo $repo_name
# Clone the source_uri repository to begin rebuild process
if [ -d "$repo_name" ]; then
  printf "\033[1;36m====================================================\033[0m\n"
  type_writer "📁---> Source repository appears already."
  type_writer "⚠️---> To run rebuilder, fix collision by removing directory with name of \$repo_name."
  cleanup
  exit 1
else
  type_writer "📥---> Cloning the source repository..."
  git clone https://$source_uri
fi

###################
#
# Enter the Repo
cd $repo_name

# Check to see if JAVA_HOME is set then empty to
# avoid triggering unbound variable error.
# TODO: Only care for java_home if INCLUDES_JAVA is set
# Check to see if JAVA_HOME is set then empty to
# avoid triggering unbound variable error.
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

################################################
#                                              #
#              Rebuild the Artifacts           #
#                                              #
################################################

if [[ -n "$DOCKER_IMAGE" ]]
then
    cd -
    sudo docker pull $DOCKER_IMAGE
    printf "\033[1;36m====================================================\033[0m\n"
    type_writer "🔨---> Rebuilding with Docker Image Environment..."    # Mount docker image on this directory as workdir to gain access to script env
    # TODO: Check to see if env vars need to be passed in.
    sudo docker run --env repo_name=$repo_name --env TARGETS=${TARGETS} --env FLAGS=${FLAGS} --env NEEDS_RUNFILES=${NEEDS_RUNFILES} --env INCLUDES_JAVA=${INCLUDES_JAVA} --rm -v $PWD:/workdir -w /workdir $DOCKER_IMAGE /bin/sh -c "cd $repo_name && ./../build.sh"
    printf "\033[1;42m✅ Artifacts rebuilt!\033[0m\n"
    echo ""
else
    # Run the build script locally without a docker image
    type_writer "💻---> Rebuilding with local environment..."
    source ../build.sh
    printf "\033[1;42m✅ Artifacts rebuilt!\033[0m\n"
    echo ""
fi

echo $(pwd)
cd $repo_name

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

rebuilt_checksum=""
unset rebuilt_checksum # Makes sure it is empty before assigning.

# IF there are runfiles, the directory structure will be different.
# The binaries folder contains different directories for the its artifacts and
# the artifacts runfiles. Obtain the rebuilt binaries and copy them to the
# path at root before cleaning up and deleting the repo.
if [[ "${NEEDS_RUNFILES}" == "true" || "${INCLUDES_JAVA}" == "true" ]]
then
    ### WRITE CONDITINOAL WHERE if includes _deploy.jar go to x_deploy go to x before it
    ## some logic that takes x out and cds to it if deploy.jar
    cd $binaries_dir/$artifact_name
    rebuilt_checksum=$(sha256sum $artifact_name | awk '{ print $1 }')
    cp $artifact_name ./../../../rebuilt_artifacts_dir
    printf "\033[1;36m====================================================\033[0m\n"
    type_writer "🧹---> Cleaning up $repo_name..."        
    cd ../../../ && rm -rf $repo_name
    echo ""
else
    ls
    cd $binaries_dir
    rebuilt_checksum=$(sha256sum $artifact_name | awk '{ print $1 }')
    ## WHY SUDO
    sudo cp $artifact_name ./../../rebuilt_artifacts_dir
    printf "\033[1;36m====================================================\033[0m\n"
    type_writer "🧹---> Cleaning up $repo_name..."    
    ## WHY DO I have to sudo
    cd ../../ && sudo rm -rf $repo_name
    echo ""
fi

ls

################################################
#                                              #
#       Check Build for Reproducibility        #
#                                              #
################################################

if [[ "$orig_checksum" == "$rebuilt_checksum" ]]
then
    printf "\033[1;42mChecksum is the \033[1m\033[4msame\033[0m\033[1;42m for the original and rebuilt artifact!\033[0m\n"
    printf "\033[1;42m✅ This build is \033[1m\033[4mreproducible\033[0m!\033[0m\n"
    echo "$orig_checksum = Original Checksum"
    echo "$rebuilt_checksum = Rebuilt Checksum"
else
    printf "\033[1;41mChecksum is \033[1m\033[4mNOT\033[0m\033[1;41m the same for the original and rebuilt artifact!\033[0m\n"
    printf "\033[1;41m⚠️ This build was \033[1m\033[4mNOT\033[0m\033[1;41m able to be reproduced!\033[0m\n"
    echo "$orig_checksum = Original Checksum"
    echo "$rebuilt_checksum = Rebuilt Checksum"
fi

cleanup