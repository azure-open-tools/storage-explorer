SHELL:=/bin/bash

build-release:
	@ chmod +x ci/build.sh
	@ ci/build.sh "asi"

release:
	@ chmod +x ./ci/release.sh
	@ ./ci/release.sh

release-binaries:
	@ chmod +x ./ci/add-release-assets.sh
	@ ./ci/add-release-assets.sh