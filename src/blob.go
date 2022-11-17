package main

import (
	"context"
	b64 "encoding/base64"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

type myBlob struct {
	Name       string             `json:"name"`
	Content    []byte             `json:"content"`
	Properties map[string]string  `json:"Properties"`
	Metadata   map[string]*string `json:"metadata"`
}

func queryBlobs(ctx context.Context, containerName string, blobNameFiler string, showContent bool, client *azblob.Client, metadataFilter []Filter) []myBlob {
	pager := client.NewListBlobsFlatPager(containerName, nil)
	var foundBlobs []myBlob
	// continue fetching pages until no more remain
	for pager.More() {
		// advance to the next page
		page, err := pager.NextPage(ctx)

		if err != nil {
			log.Fatal(err)
		}

		for _, blob := range page.Segment.BlobItems {
			// TODO substring match? to match blobs: ['test-1', 'test-2'], term: 'test, matches ['test-1', 'test-2']
			if len(blobNameFiler) == 0 || strings.Contains(*blob.Name, blobNameFiler) {
				b := createBlobOutput(ctx, client, containerName, *blob, showContent, metadataFilter)
				if b != nil {
					foundBlobs = append(foundBlobs, *b)
				}
			}
		}
	}
	return foundBlobs
}

func parseBlobProperties(properties *container.BlobProperties) map[string]string {
	result := make(map[string]string)

	result["Content MD5"] = b64.StdEncoding.EncodeToString(properties.ContentMD5)
	result["Created at"] = properties.CreationTime.String()
	result["Last modified at"] = properties.LastModified.String()

	if properties.BlobType != nil {
		result["Blob Type"] = string(*properties.BlobType)
	}

	if properties.LeaseStatus != nil {
		result["Lease Status"] = string(*properties.LeaseStatus)
	}

	if properties.LeaseState != nil {
		result["Lease State"] = string(*properties.LeaseState)
	}

	if properties.LeaseDuration != nil {
		result["Lease Duration"] = string(*properties.LeaseDuration)
	}

	return result
}

func downloadBlob(ctx context.Context, containerName string, blobName string, client *azblob.Client) []byte {
	var buffer []byte
	// TODO make it work
	// _, err := client.DownloadBuffer(ctx, containerName, blobName, buffer, nil)
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }
	return buffer
}

func createBlobOutput(ctx context.Context, client *azblob.Client, containerName string, blobItem container.BlobItem, downloadContent bool, metadataFilter []Filter) *myBlob {
	if len(metadataFilter) == 0 || (len(metadataFilter) > 0 && containsMetadataMatch(blobItem.Metadata, metadataFilter)) {
		blob := new(myBlob)
		blob.Name = *blobItem.Name
		blob.Properties = parseBlobProperties(blobItem.Properties)
		blob.Metadata = blobItem.Metadata

		if downloadContent {
			blob.Content = downloadBlob(ctx, containerName, *blobItem.Name, client)
		}
		return blob
	}
	return nil
}
