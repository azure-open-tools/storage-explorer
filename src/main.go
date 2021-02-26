package main

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	az "github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/spf13/cobra"
)

type arguments struct {
	AccountName   string
	AccessKey     string
	ContainerName string
	BlobName      string
	ShowContent   bool
	StoreContent  bool
	ContentOnly   bool
	FileName      string
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

var defaultFileName = "blobcontent.txt"

func init() {
	rootCmd.Flags().StringVarP(&largs.AccountName, "accountName", "n", "", "accountName of the Storage Account")
	rootCmd.Flags().StringVarP(&largs.AccessKey, "accessKey", "k", "", "accessKey for the Storage Account")
	rootCmd.Flags().StringVarP(&largs.ContainerName, "container", "c", "", "filter for container name with substring match")
	rootCmd.Flags().StringVarP(&largs.BlobName, "blob", "b", "", "filter for blob name with substring match")
	rootCmd.Flags().BoolVar(&largs.ShowContent, "show-content", false, "downloads and prints content of blobs in addition to other logs")
	rootCmd.Flags().BoolVar(&largs.StoreContent, "store-content", false, "downloads and stores content of blob in a file. Use --filename or -f to set a specific filename. Stores one line for each blob")
	rootCmd.Flags().StringVarP(&largs.FileName, "filename", "f", "", "in addtion")
	rootCmd.Flags().BoolVar(&largs.ContentOnly, "content-only", false, "prints only content of blob. overrules --show-content")
	rootCmd.MarkFlagRequired("accountName")
	rootCmd.MarkFlagRequired("accessKey")
}

func line(print bool, level int, content string) {
	if !print {
		return
	}

	if level == 0 {
		fmt.Println(content)
	} else {
		spacer := "___"
		prefix := "|"

		for i := 0; i < (level - 1); i++ {
			prefix = prefix + " |"
		}
		line := prefix + spacer + content
		fmt.Println(line)
	}
}

func downloadBlob(wg *sync.WaitGroup, blobName string, containerUrl az.ContainerURL) string {
	defer wg.Done()
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

	return downloadedData.String()
}

func storeBlobContent(f *os.File, content string) {
	f.WriteString(fmt.Sprintf("%s\n", content))
}

func exec(args arguments) {
	var f *os.File
	var err error

	if args.StoreContent {
		outputFile := defaultFileName
		if len(args.FileName) > 0 {
			outputFile = args.FileName
		}

		f, err = os.Create(outputFile)
		if err != nil {
			log.Fatalf("Could not create file %s", outputFile)
		}
		defer f.Close()
	}

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

	line(!args.ContentOnly, 0, URL.String())

	for marker := (azblob.Marker{}); marker.NotDone(); {
		listContainer, err := serviceURL.ListContainersSegment(ctx, marker, azblob.ListContainersSegmentOptions{})

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
			if !args.ContentOnly {
				line(!args.ContentOnly, 1, fmt.Sprintf("Container: %s", val.Name))
			}

			containerURL, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", args.AccountName, containerName))
			containerServiceURL := azblob.NewContainerURL(*containerURL, p)

			for blobMarker := (azblob.Marker{}); blobMarker.NotDone(); {
				// Get a result segment starting with the blob indicated by the current Marker.
				listBlob, _ := containerServiceURL.ListBlobsFlatSegment(ctx, blobMarker, azblob.ListBlobsSegmentOptions{Details: azblob.BlobListingDetails{Metadata: true}})

				// ListBlobs returns the start of the next segment; you MUST use this to get
				// the next segment (after processing the current result segment).
				blobMarker = listBlob.NextMarker

				// waitgroup to download all blobs in this segemnt before getting next segment
				var wg sync.WaitGroup

				// Process the blobs returned in this result segment (if the segment is empty, the loop body won't execute)
				for _, blobItem := range listBlob.Segment.BlobItems {
					blobName := blobItem.Name

					// TODO substring match?
					if len(args.BlobName) > 0 && !strings.Contains(blobName, args.BlobName) {
						continue
					}
					blobFound = true
					line(!args.ContentOnly, 2, fmt.Sprintf("Blob: %s", blobName))
					// see here for Properties: https://github.com/Azure/azure-storage-blob-go/blob/456ab4777f89ceb54316ddf71d2acfd39bb86e1d/azblob/zz_generated_models.go#L2343
					line(!args.ContentOnly, 3, fmt.Sprintf("Blob Type: %s", blobItem.Properties.BlobType))
					line(!args.ContentOnly, 3, fmt.Sprintf("Content MD5: %s", b64.StdEncoding.EncodeToString(blobItem.Properties.ContentMD5)))
					line(!args.ContentOnly, 3, fmt.Sprintf("Created at: %s", blobItem.Properties.CreationTime))
					line(!args.ContentOnly, 3, fmt.Sprintf("Last modified at: %s", blobItem.Properties.LastModified))
					line(!args.ContentOnly, 3, fmt.Sprintf("Lease Status: %s", blobItem.Properties.LeaseStatus))
					line(!args.ContentOnly, 3, fmt.Sprintf("Lease State: %s", blobItem.Properties.LeaseState))
					line(!args.ContentOnly, 3, fmt.Sprintf("Lease Duration: %s", blobItem.Properties.LeaseDuration))

					line(!args.ContentOnly, 3, "Metadata:")
					for key, entry := range blobItem.Metadata {
						line(!args.ContentOnly, 4, fmt.Sprintf("%s: %s", key, entry))
					}

					if args.ShowContent || args.StoreContent || args.ContentOnly {
						wg.Add(1)
						content := downloadBlob(&wg, blobName, containerServiceURL)

						if args.ContentOnly {
							line(args.ContentOnly, 0, content)
						} else if args.ShowContent {
							line(!args.ContentOnly, 3, fmt.Sprintf("Content: %s", content))
						}

						if args.StoreContent {
							storeBlobContent(f, content)
						}
					}
				}
				// wait for all go routines to finish before continue
				wg.Wait()
			}
		}
		marker = listContainer.NextMarker // Pagination
	}
	if !containerFound {
		line(true, 1, fmt.Sprintf("No Container found for Name %s", args.ContainerName))
	} else if !blobFound {
		line(true, 2, fmt.Sprintf("No Blob found for Name %s", args.BlobName))
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
