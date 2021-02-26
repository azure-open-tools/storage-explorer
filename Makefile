SHELL:=/bin/bash

build-release:
	@ chmod +x ci/build.sh
	@ ci/build.sh ${PWD}/src/version.go "asi"

release:
	@ chmod +x ./ci/release.sh
	@ ./ci/release.sh ${PWD}/src/version.go

release-binaries:
	@ chmod +x ./ci/add-release-assets.sh
	@ ./ci/add-release-assets.sh