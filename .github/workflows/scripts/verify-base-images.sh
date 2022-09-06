#!/bin/bash

set -euo pipefail

# NOTE: Use read to avoid whitespace issues.
find . -name Dockerfile -print0 | while IFS= read -r -d '' f; do
	echo "Checking $f"
	grep FROM "$f" | while IFS= read -r line; do
		image_full=$(echo "$line" | awk '{ print $2 }')
		image_name=$(echo "$image_full" | cut -d '@' -f 1)
		image_sha=$(echo "$image_full" | cut -d '@' -f 2- | cut -d ':' -f 2-)

		echo "Verifying base image $image_full"

		# verify that the image contains a sha.
		if [ "$image_sha" == "" ]; then
			echo "\"$image_full\" should be referenced by digest."
			exit 2
		fi

		# verify distroless base images.
		if [[ "$image_name" == gcr.io/distroless/* ]]; then

			# verify the image signature.
			cosign verify --key .github/workflows/scripts/distroless.pub "$image_full"
		else
			# All other base images should be signed using Docker Content Trust.
			if ! (DOCKER_CONTENT_TRUST=1 docker trust inspect --pretty "$image_name" | grep "$image_sha"); then
				echo "$image_full: unable to verify Docker Content Trust."
				exit 2
			fi
		fi
	done
done
