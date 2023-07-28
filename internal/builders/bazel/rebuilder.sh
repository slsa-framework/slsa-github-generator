#!/bin/bash
set -euo pipefail

# Current constraint: make the users download and setup the specific version of Java
#                     that the workflow uses in the meantime of coming up with automation solution
#                     (if one is needed)

# TODO: add flags that come before it. Example: --prov_path path/to/prov
# Get the inputs from the user
artifact_path=$1
prov_path=$2
source_uri=$3
builder_id="https://github.com/enteraga6/slsa-github-generator/.github/workflows/builder_bazel_slsa3.yml"

# Compute the original checksum of the artifact to compare
orig_checksum=$(sha256sum $artifact_path | awk '{ print $1 }')

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

declare -A data

# Extract the inputs and put them in associative array where the key is the workflow input,
# and the value is the value that the user inputted. Pipe value to @text, to deal with booleans.
while IFS='=' read -r key value; do
    data["$key"]="$value"
done < <(cat $prov_path | jq -r '.dsseEnvelope.payload' | base64 -d | jq -r '.predicate.buildDefinition.externalParameters.inputs | to_entries | .[] | .key + "=" + (.value | @text)')

echo ""
echo "loop"
for key in "${!data[@]}"
do
    echo "$key = ${data[$key]}"
done

# The name map will convert and export the key strings of inputs to
# match with the environment variables of the Bazel Builder build.sh
declare -A name_mapping
name_mapping["targets"]="TARGETS"
name_mapping["flags"]="FLAGS"
name_mapping["includes-java"]="INCLUDES_JAVA"
name_mapping["needs-runfiles"]="NEEDS_RUNFILES"
name_mapping["docker-image"]="DOCKER_IMAGE"

for key in "${!data[@]}"; do

    # Check to see if the key is in name map before export as env var.
    if [[ ${name_mapping[$key]+_} ]]; then
        export "${name_mapping[$key]}"="${data[$key]}"
    fi
done

echo $source_uri
repo_name=$(echo $source_uri | awk -F/ '{print $NF}')
echo "repo_name"
echo $repo_name
echo "repo named printed"
# Clone the source_uri repository
if [ -d "$repo_name" ]; then
  echo "The source repo is already cloned."
else
  echo "The source repository is not cloned. Cloning..."
  git clone https://$source_uri
fi

cd $repo_name

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

if [[ $artifact_path == */* ]]
then
    artifact_name=$(basename $artifact_path)
else
    artifact_name=$artifact_path
fi

if [[ -n "$DOCKER_IMAGE" ]]
then
    docker pull $DOCKER_IMAGE
    docker run --rm -v $PWD:/workdir -w workdir $DOCKER_IMAGE /bin/sh -c "./build.sh"
else
    source ../build.sh
fi

if [[ "${NEEDS_RUNFILES}" == "true" ]]
then
    cd bazel_builder_binaries_to_upload_to_gh/$artifact_name
else
    cd bazel_builder_binaries_to_upload_to_gh
fi

rebuilt_checksum=$(sha256sum $artifact_name | awk '{ print $1 }')

if [[ "$orig_checksum" == "$rebuilt_checksum" ]]
then
    echo "Checksum is the same for the original and rebuilt artifact"
    echo "$orig_checksum = Original Checksum"
    echo "$rebuilt_checksum = Rebuilt Checksum"
else
    echo "Checksum is NOT the same for the original and rebuilt artifact"
    echo "$orig_checksum = Original Checksum"
    echo "$rebuilt_checksum = Rebuilt Checksum"
fi

