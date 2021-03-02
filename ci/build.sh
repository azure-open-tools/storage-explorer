#!/usr/bin/env bash

name=$1

go version
echo "$OS"
echo "$targetarch"
echo "GitHubRef: $GITHUB_REF"

targetos="$OS"
targetarch="amd64"

# show whats there
ls -lh

version=$(go run . -v)

echo "Using $version to build"

cd src/

if [[ "$targetos" == *"Windows_NT"* ]];
then
	set GOARCH="$targetarch"
	set GOOS="$targetos"
	set GO111MODULE=on
	extension=".exe"
	go build -ldflags "-s -w" -o "$name-windows-""$targetarch"-"$version""$extension" .
	mv "$name-windows-""$targetarch"-"$version""$extension" ../
	ls -lah
else
  targetos=$(sw_vers | awk '{print $1$2$3}' | head -n 1)
  echo "Target OS: $targetos"
  if [[ "$targetos" == *"MacOSX"* ]];
  then
  	echo "$PWD"
    env GO111MODULE=on GOOS="darwin" GOARCH="$targetarch" go build -ldflags "-s -w" -o "$name-darwin-""$targetarch"-"$version" .
  	mv "$name-darwin-""$targetarch"-"$version" ../
  else
  	env GO111MODULE=on GOOS="linux" GOARCH="$targetarch" go build -ldflags "-s -w" -o "$name-linux-""$targetarch"-"$version" .
    mv "$name-linux-""$targetarch"-"$version" ../
  fi
fi
