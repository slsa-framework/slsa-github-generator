mkdir release-files-for-slsa
GRADLE_VERSION=$(./gradlew properties -q | grep "version:" | awk '{print $2}')

# Move artifacts from the user-supplied artifact list
ARTIFACT_LIST=\"${ARTIFACT_LIST}\"
for i in ${ARTIFACT_LIST//,/ }
do
    i=$(echo $i | sed 's/\"//')
    # Check if $i is empty
    if [ -z "$i" ]
    then
        continue
    fi
    
    # Replace 'GRADLE_VERSION' with $GRADLE_VERSION
    PATH_WITH_VERSION="${i//GRADLE_VERSION/"$GRADLE_VERSION"}"

    # Move the file
    bn=$(basename -- $PATH_WITH_VERSION)
    mv $PATH_WITH_VERSION release-files-for-slsa/$bn
done
