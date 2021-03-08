package main

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"log"
	"strings"
	"sync"

	"github.com/Azure/azure-storage-blob-go/azblob"
	az "github.com/Azure/azure-storage-blob-go/azblob"
)

type blob struct {
	Name       string            `json:"name"`
	Content    []byte            `json:"content"`
	Properties map[string]string `json:"Properties"`
	Metadata   map[string]string `json:"metadata"`
}

func parseBlobs(blobItems []az.BlobItemInternal, blobFilter string, showContent bool, containerURL azblob.ContainerURL, metadataFilter []Filter) []blob {
	var blobWg sync.WaitGroup
	bc := make(chan *blob)

	var blobs []blob

	for _, blobItem := range blobItems {
		if len(blobFilter) > 0 && !strings.Contains(blobItem.Name, blobFilter) {
			continue
		}
		blobWg.Add(1)
		go createBlobOutput(blobItem, &blobWg, bc, showContent, containerURL, metadataFilter)
	}

	go func() {
		blobWg.Wait()
		close(bc)
	}()

	for elem := range bc {
		blobs = append(blobs, *elem)
	}
	return blobs
}

func parseBlobProperties(properties az.BlobProperties) map[string]string {
	result := make(map[string]string)

	result["Blob Type"] = string(properties.BlobType)
	result["Content MD5"] = b64.StdEncoding.EncodeToString(properties.ContentMD5)
	result["Created at"] = properties.CreationTime.String()
	result["Last modified at"] = properties.LastModified.String()
	result["Lease Status"] = string(properties.LeaseStatus)
	result["Lease State"] = string(properties.LeaseState)
	result["Lease Duration"] = string(properties.LeaseDuration)

	return result
}

func createBlobOutput(blobItem az.BlobItemInternal, wg *sync.WaitGroup, c chan *blob, downloadContent bool, containerURL azblob.ContainerURL, metadataFilter []Filter) {
	defer wg.Done()

	if len(metadataFilter) == 0 || (len(metadataFilter) > 0 && containsMetadataMatch(blobItem.Metadata, metadataFilter)) {
		blob := new(blob)
		blob.Name = blobItem.Name
		blob.Properties = parseBlobProperties(blobItem.Properties)
		blob.Metadata = blobItem.Metadata

		if downloadContent {
			blob.Content = downloadBlob(blobItem.Name, containerURL)
		}

		c <- blob
	}
}

func downloadBlob(blobName string, containerUrl az.ContainerURL) []byte {
	blobURL := containerUrl.NewBlockBlobURL(blobName)
	downloadResponse, err := blobURL.Download(context.Background(), 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false)

	if err != nil {
		log.Fatalf("Error downloading blob %s", blobName)
	}

	bodyStream := downloadResponse.Body(azblob.RetryReaderOptions{MaxRetryRequests: 20})
	downloadedData := bytes.Buffer{}
	_, err = downloadedData.ReadFrom(bodyStream)

	if err != nil {
		log.Fatalf("Error reading blob %s", blobName)
	}

	return downloadedData.Bytes()
}
