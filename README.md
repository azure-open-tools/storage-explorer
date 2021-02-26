# go-storage-information

[![storage-information](https://github.com/azure-open-tools/storage-explorer/actions/workflows/storage-information.yml/badge.svg)](https://github.com/azure-open-tools/storage-explorer/actions/workflows/storage-information.yml)

This tool was build to get quickly and easy Information about Azure Blobs.
It prints basic attributes and set metadata on blobs.
It is also possible to download the content of blobs wether to stdout or to a file.
Furthermore it is possible to filter for a certain Storage Container Names or/and Blob names.

## Parameter

* `--accountName` or `-n`: (mandatory) provide the name of the Storage Account
* `--accessKey` or `-k`: (mandatory) provide the Access Key of Storage Account
* `--container` or `-c`: (optional) filter for a specific container by its name. Substring match
* `--blob` or `-b`: (optional) provide the key Storage Account. Substring match
* `--show-content`: (optional) prints additionally the content of the blobs
* `--store-content`: (optional) stores additionally the content to a file. One Line per each Blob Content
* `--filename` or `-f`: (optional) use together with `--store-content` to store blob contents to given file. Default is `blobcontent.txt`
* `--content-only`: (optional) prints only the content of blobs to stdout. No other information are printed
