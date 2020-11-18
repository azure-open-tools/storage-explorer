package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/spf13/cobra"
)

func printLine(level int, content string) {
	spacer := "___"
	prefix := "|"
	for i := 0; i < level; i++ {
		prefix = prefix + " |"
	}
	line := prefix + spacer + content
	fmt.Println(line)
}

type arguments struct {
	AccountName   string
	AccessKey     string
	ContainerName string
	BlobName      string
}

var largs = arguments{}

var rootCmd = &cobra.Command{
	Use:   "tbd",
	Short: "tbd shows containers and blobs of a azure storage account",
	Long: `tbd shows containers and blobs of a azure storage account.
Complete documentation is available at http://hugo.spf13.com`,
	Run: func(cmd *cobra.Command, args []string) {
		exec(largs)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&largs.AccountName, "accountName", "n", "", "accountName of the Storage Account")
	rootCmd.Flags().StringVarP(&largs.AccessKey, "accessKey", "k", "", "accessKey for the Storage Account")
	rootCmd.Flags().StringVarP(&largs.ContainerName, "container", "c", "", "filter for container name")
	rootCmd.Flags().StringVarP(&largs.BlobName, "blob", "b", "", "filter for blob name")
	rootCmd.MarkFlagRequired("accountName")
	rootCmd.MarkFlagRequired("accessKey")
}

func exec(args arguments) {
	ctx := context.Background()

	// Create a default request pipeline using your storage account name and account key
	credential, _ := azblob.NewSharedKeyCredential(args.AccountName, args.AccessKey)
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	// From the Azure portal, get your storage account blob service URL endpoint
	URL, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", args.AccountName))

	serviceURL := azblob.NewServiceURL(*URL, p)

	for marker := (azblob.Marker{}); marker.NotDone(); {
		listContainer, _ := serviceURL.ListContainersSegment(ctx, marker, azblob.ListContainersSegmentOptions{})

		fmt.Println(listContainer.ServiceEndpoint)
		for _, val := range listContainer.ContainerItems {
			containerName := val.Name

			if len(args.ContainerName) > 0 && containerName != args.ContainerName {
				continue
			}
			printLine(0, fmt.Sprintf("Container: %s", val.Name))

			containerURL, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", args.AccountName, containerName))
			containerServiceURL := azblob.NewContainerURL(*containerURL, p)

			for blobMarker := (azblob.Marker{}); blobMarker.NotDone(); {
				// Get a result segment starting with the blob indicated by the current Marker.
				listBlob, _ := containerServiceURL.ListBlobsFlatSegment(ctx, blobMarker, azblob.ListBlobsSegmentOptions{Details: azblob.BlobListingDetails{Metadata: true}})

				// ListBlobs returns the start of the next segment; you MUST use this to get
				// the next segment (after processing the current result segment).
				blobMarker = listBlob.NextMarker

				// Process the blobs returned in this result segment (if the segment is empty, the loop body won't execute)
				for _, blobItems := range listBlob.Segment.BlobItems {
					blobName := blobItems.Name

					if len(args.BlobName) > 0 && blobName != args.BlobName {
						continue
					}

					printLine(1, fmt.Sprintf("Blob: %s", blobName))
					// see here for Properties: https://github.com/Azure/azure-storage-blob-go/blob/456ab4777f89ceb54316ddf71d2acfd39bb86e1d/azblob/zz_generated_models.go#L2343
					printLine(2, fmt.Sprintf("Blob Type: %s", blobItems.Properties.BlobType))
					printLine(2, fmt.Sprintf("Created at: %s", blobItems.Properties.CreationTime))
					printLine(2, fmt.Sprintf("Last modified at: %s", blobItems.Properties.LastModified))
					printLine(2, fmt.Sprintf("Lease Status: %s", blobItems.Properties.LeaseStatus))
					printLine(2, fmt.Sprintf("Lease State: %s", blobItems.Properties.LeaseState))
					printLine(2, fmt.Sprintf("Lease Duration: %s", blobItems.Properties.LeaseDuration))

					for key, entry := range blobItems.Metadata {
						printLine(2, fmt.Sprintf("%s: %s", key, entry))
					}
				}
			}
		}
		marker = listContainer.NextMarker // Pagination
	}
}

//kudos to: https://github.com/Azure/azure-storage-blob-go/blob/456ab4777f89ceb54316ddf71d2acfd39bb86e1d/azblob/zt_examples_test.go
func main() {
	rootCmd.Execute()
}
