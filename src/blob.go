package main

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

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

	layout := "2006-01-02T15:04:05.000Z"
	// 19 UTC is 20 localtime
	str := "2021-03-09T19:00:00.000Z"
	t, _ := time.Parse(layout, str)

	cTime := blobItem.Properties.CreationTime

	if cTime.After(t) {
		fmt.Println(fmt.Sprintf("%s: reset cause %v is after %v", blobItem.Name, cTime, t))
		resetProperty(blobItem.Name, containerURL, blobItem.Metadata)
	}

	/*if len(metadataFilter) == 0 || (len(metadataFilter) > 0 && containsMetadataMatch(blobItem.Metadata, metadataFilter)) {
		blob := new(blob)
		blob.Name = blobItem.Name
		blob.Properties = parseBlobProperties(blobItem.Properties)
		blob.Metadata = blobItem.Metadata

		if downloadContent {
			blob.Content = downloadBlob(blobItem.Name, containerURL)
		}

		c <- blob
	}*/
}

func resetProperty(blobName string, containerUrl az.ContainerURL, meta map[string]string) {
	blobURL := containerUrl.NewBlobURL(blobName)
	ctx := context.Background()

	_, ok := meta["processed"]
	if ok {
		delete(meta, "processed")
		_, err := blobURL.SetMetadata(ctx, meta, az.BlobAccessConditions{})
		if err != nil {
			fmt.Println(err)
		}
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
