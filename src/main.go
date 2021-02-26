package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/spf13/cobra"
)

type container struct {
	Name  string `json:"name"`
	Blobs []blob `json:"blobs"`
}

type storageAccount struct {
	Name      string      `json:"name"`
	Container []container `json:"container"`
}

type arguments struct {
	AccountName    string
	AccessKey      string
	ContainerName  string
	BlobName       string
	ShowContent    bool
	MetadataFilter []string
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

const (
	storageURLTemplate   = "https://%s.blob.core.windows.net"
	containerURLTemplate = "https://%s.blob.core.windows.net/%s"
)

func init() {
	rootCmd.Flags().StringVarP(&largs.AccountName, "accountName", "n", "", "accountName of the Storage Account")
	rootCmd.Flags().StringVarP(&largs.AccessKey, "accessKey", "k", "", "accessKey for the Storage Account")
	rootCmd.Flags().StringVarP(&largs.ContainerName, "container", "c", "", "filter for container name with substring match")
	rootCmd.Flags().StringVarP(&largs.BlobName, "blob", "b", "", "filter for blob name with substring match")
	rootCmd.Flags().BoolVar(&largs.ShowContent, "show-content", false, "downloads and prints content of blobs in addition to other logs")
	rootCmd.Flags().StringSliceVarP(&largs.MetadataFilter, "metadata-filter", "m", []string{}, "OR filter for blob metadata. Structure is <key>:<value>")
	rootCmd.MarkFlagRequired("accountName")
	rootCmd.MarkFlagRequired("accessKey")
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
	URL, _ := url.Parse(fmt.Sprintf(storageURLTemplate, args.AccountName))

	serviceURL := azblob.NewServiceURL(*URL, p)

	s := new(storageAccount)
	s.Name = URL.String()
	var foundContainer []container

	metadataFilter := createMetadataFilter(args.MetadataFilter)

	c := make(chan *container)
	var wg sync.WaitGroup
	for marker := (azblob.Marker{}); marker.NotDone(); {
		listContainer, err := serviceURL.ListContainersSegment(ctx, marker, azblob.ListContainersSegmentOptions{})

		if err != nil {
			log.Fatal("Error while getting Container")
		}

		for _, val := range listContainer.ContainerItems {
			wg.Add(1)
			go parseContainer(val, p, args.AccountName, args.ContainerName, args.BlobName, args.ShowContent, c, &wg, marker, metadataFilter)
		}
		// used for Pagination
		marker = listContainer.NextMarker
	}

	// wait for all entries in waitgroup and close then the channel
	go func() {
		wg.Wait()
		close(c)
	}()

	// channel to collect results
	for elem := range c {
		foundContainer = append(foundContainer, *elem)
	}

	s.Container = foundContainer
	print(*s)
}

func print(sa storageAccount) {
	m, _ := json.Marshal(sa)
	fmt.Println(string(m))
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
