#!/usr/bin/env bash

name=$1

go version
echo "OS: $RUNNER_OS"
echo "Arch: $RUNNER_ARCH"
echo "GitHubRef: $GITHUB_REF"

targetos="$OS"
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
else
  if [[ "$targetos" == *"macOS"* ]];
  then
  	echo "$PWD"
    env GO111MODULE=on GOOS="darwin" GOARCH="$targetarch" go build -ldflags "-s -w" -o "$name-darwin-""$targetarch"-"$version" .
  	mv "$name-darwin-""$targetarch"-"$version" ../
  else
  	env GO111MODULE=on GOOS="linux" GOARCH="$targetarch" go build -ldflags "-s -w" -o "$name-linux-""$targetarch"-"$version" .
    mv "$name-linux-""$targetarch"-"$version" ../
  fi
fi
