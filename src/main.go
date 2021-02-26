package main

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	az "github.com/Azure/azure-storage-blob-go/azblob"
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
	ShowContent   bool
}

var largs = arguments{}

var rootCmd = &cobra.Command{
	Use:   "go-storage-explorer",
	Short: "go-storage-explorer shows containers and blobs of a azure storage account",
	Long: `go-storage-explorer shows containers and blobs of a azure storage account.
Complete documentation is available at http://hugo.spf13.com`,
	Run: func(cmd *cobra.Command, args []string) {
		exec(largs)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&largs.AccountName, "accountName", "n", "", "accountName of the Storage Account")
	rootCmd.Flags().StringVarP(&largs.AccessKey, "accessKey", "k", "", "accessKey for the Storage Account")
	rootCmd.Flags().StringVarP(&largs.ContainerName, "container", "c", "", "filter for container name with substring match")
	rootCmd.Flags().StringVarP(&largs.BlobName, "blob", "b", "", "filter for blob name with substring match")
	rootCmd.Flags().BoolVar(&largs.ShowContent, "show-content", false, "downloads and prints content of blob")
	rootCmd.MarkFlagRequired("accountName")
	rootCmd.MarkFlagRequired("accessKey")
}

func downloadBlob(fileName string, containerUrl az.ContainerURL) string {
	blobURL := containerUrl.NewBlockBlobURL(fileName)
	downloadResponse, err := blobURL.Download(context.Background(), 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false)

	if err != nil {
		log.Fatalf("Error downloading blob %s", fileName)
	}

	bodyStream := downloadResponse.Body(azblob.RetryReaderOptions{MaxRetryRequests: 20})
	downloadedData := bytes.Buffer{}
	_, err = downloadedData.ReadFrom(bodyStream)

	if err != nil {
		log.Fatalf("Error reading blob %s", fileName)
	}

	return downloadedData.String()
}

func exec(args arguments) {
	ctx := context.Background()

	// Create a default request pipeline using your storage account name and account key
	credential, authErr := azblob.NewSharedKeyCredential(args.AccountName, args.AccessKey)
	if authErr != nil {
		log.Fatal("Error while Authentication")
	}
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	// From the Azure portal, get your storage account blob service URL endpoint
	URL, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", args.AccountName))

	serviceURL := azblob.NewServiceURL(*URL, p)

	containerFound := false
	blobFound := false
	for marker := (azblob.Marker{}); marker.NotDone(); {
		listContainer, err := serviceURL.ListContainersSegment(ctx, marker, azblob.ListContainersSegmentOptions{})
		fmt.Println(listContainer.ServiceEndpoint)

		if err != nil {
			log.Fatal("Error while getting Container")
		}

		for _, val := range listContainer.ContainerItems {
			containerName := val.Name

			// TODO substring match? to match containers: ['test-1', 'test-2'], term: 'test, matches ['test-1', 'test-2']
			if len(args.ContainerName) > 0 && !strings.Contains(containerName, args.ContainerName) {
				continue
			}
			containerFound = true
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
				for _, blobItem := range listBlob.Segment.BlobItems {
					blobName := blobItem.Name

					// TODO substring match?
					if len(args.BlobName) > 0 && !strings.Contains(blobName, args.BlobName) {
						continue
					}
					blobFound = true
					printLine(1, fmt.Sprintf("Blob: %s", blobName))
					// see here for Properties: https://github.com/Azure/azure-storage-blob-go/blob/456ab4777f89ceb54316ddf71d2acfd39bb86e1d/azblob/zz_generated_models.go#L2343
					printLine(2, fmt.Sprintf("Blob Type: %s", blobItem.Properties.BlobType))
					printLine(2, fmt.Sprintf("Content MD5: %s", b64.StdEncoding.EncodeToString(blobItem.Properties.ContentMD5)))
					printLine(2, fmt.Sprintf("Created at: %s", blobItem.Properties.CreationTime))
					printLine(2, fmt.Sprintf("Last modified at: %s", blobItem.Properties.LastModified))
					printLine(2, fmt.Sprintf("Lease Status: %s", blobItem.Properties.LeaseStatus))
					printLine(2, fmt.Sprintf("Lease State: %s", blobItem.Properties.LeaseState))
					printLine(2, fmt.Sprintf("Lease Duration: %s", blobItem.Properties.LeaseDuration))

					printLine(2, "Metadata:")
					for key, entry := range blobItem.Metadata {
						printLine(3, fmt.Sprintf("%s: %s", key, entry))
					}

					if args.ShowContent {
						content := downloadBlob(blobName, containerServiceURL)
						printLine(2, fmt.Sprintf("Content: %s", content))
					}
				}
			}
		}
		marker = listContainer.NextMarker // Pagination
	}
	if !containerFound {
		printLine(0, fmt.Sprintf("No Container not found for Name %s", args.ContainerName))
	} else if !blobFound {
		printLine(1, fmt.Sprintf("No Blob not found for Name %s", args.BlobName))
	}
}

// kudos to:
// https://github.com/Azure/azure-storage-blob-go/blob/456ab4777f89ceb54316ddf71d2acfd39bb86e1d/azblob/zt_examples_test.go
// and
// https://github.com/Azure-Samples/storage-blobs-go-quickstart/blob/master/storage-quickstart.go
func main() {
	start := time.Now()
	// Code to measure

	rootCmd.Execute()

	duration := time.Since(start)

	// Formatted string, such as "2h3m0.5s" or "4.503μs"
	fmt.Println(duration)
}
