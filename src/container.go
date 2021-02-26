package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"

	pipeline "github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
	az "github.com/Azure/azure-storage-blob-go/azblob"
)

func parseContainer(azContainer az.ContainerItem, p pipeline.Pipeline, accountName string, containerFilter string, blobFilter string, showContent bool, c chan *container, wg *sync.WaitGroup, marker az.Marker, metadataFilter []Filter) {
	defer wg.Done()
	containerName := azContainer.Name

	// TODO substring match? to match containers: ['test-1', 'test-2'], term: 'test, matches ['test-1', 'test-2']
	if len(containerFilter) > 0 && !strings.Contains(containerName, containerFilter) {
		return
	}

	// new returns pointer to the container instance
	containerResult := new(container)
	containerResult.Name = containerName

	containerURL, _ := url.Parse(fmt.Sprintf(containerURLTemplate, accountName, containerName))
	containerServiceURL := azblob.NewContainerURL(*containerURL, p)

	ctx := context.Background()

	for blobMarker := (azblob.Marker{}); blobMarker.NotDone(); {
		listBlob, _ := containerServiceURL.ListBlobsFlatSegment(ctx, marker, azblob.ListBlobsSegmentOptions{Details: azblob.BlobListingDetails{Metadata: true}})
		blobMarker = listBlob.NextMarker
		blobItems := listBlob.Segment.BlobItems
		containerResult.Blobs = parseBlobs(blobItems, blobFilter, showContent, containerServiceURL, metadataFilter)
	}

	c <- containerResult
}
