#!/usr/bin/env bash

name=$1

go version
echo "RUNNER_OS: $RUNNER_OS"
echo "RUNNER_Arch: $RUNNER_ARCH"
echo "GitHubRef: $GITHUB_REF"

targetos="$RUNNER_OS"
targetarch="amd64"

cd src/
version=$(go run . -v)

if [[ "$targetos" == *"Windows"* ]];
then
	set GOARCH="$targetarch"
	set GOOS="$targetos"
	set GO111MODULE=on
	extension=".exe"
	go build -ldflags "-s -w" -o "$name-windows-""$targetarch"-"$version""$extension" .
	mv "$name-windows-""$targetarch"-"$version""$extension" ../
	ls -lah

elif [[ "$targetos" == *"macOS"* ]];
then
  	echo "$PWD"
    env GO111MODULE=on GOOS="darwin" GOARCH="$targetarch" go build -ldflags "-s -w" -o "$name-darwin-""$targetarch"-"$version" .
  	mv "$name-darwin-""$targetarch"-"$version" ../

elif [[ "$targetos" == *"Linux"* ]];
then
  	env GO111MODULE=on GOOS="linux" GOARCH="$targetarch" go build -ldflags "-s -w" -o "$name-linux-""$targetarch"-"$version" .
    mv "$name-linux-""$targetarch"-"$version" ../
else
	echo "ERROR: RUNNER:OS=$targetos is not supported"
    exit 1
fi
