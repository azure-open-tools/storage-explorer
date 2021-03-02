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

clean:
	@ go version
	@ echo -e "\ncleaning...\n" && pushd src/ && rm -f stgexplorer && popd

build: clean
	@ pushd src/ && go mod tidy && echo -e "\nbuilding...\n" && go build -o stgexplorer main.go && popd
	@ ls -lah src/stgexplorer