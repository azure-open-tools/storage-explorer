package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/spf13/cobra"
)

type myContainer struct {
	Name  string   `json:"name"`
	Blobs []myBlob `json:"blobs"`
}

type storageAccount struct {
	Name      string        `json:"name"`
	Container []myContainer `json:"container"`
}

type arguments struct {
	AccountName    string
	AccessKey      string
	MSI            string
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
	Version: getVersion(),
}

const (
	storageURLTemplate = "https://%s.blob.core.windows.net"
)

var foundContainer []myContainer

func init() {
	rootCmd.Flags().StringVarP(&largs.AccountName, "accountName", "n", "", "accountName of the Storage Account")
	rootCmd.Flags().StringVarP(&largs.AccessKey, "accessKey", "k", "", "accessKey for the Storage Account")
	rootCmd.Flags().StringVarP(&largs.MSI, "msi", "i", "", "user assigned managed Identity to Access the Storage Account")
	rootCmd.Flags().StringVarP(&largs.ContainerName, "container", "c", "", "filter for container name with substring match")
	rootCmd.Flags().StringVarP(&largs.BlobName, "blob", "b", "", "filter for blob name with substring match")
	rootCmd.Flags().BoolVar(&largs.ShowContent, "show-content", false, "downloads and prints content of blobs in addition to other logs")
	rootCmd.Flags().StringSliceVarP(&largs.MetadataFilter, "metadata-filter", "m", []string{}, "OR filter for blob metadata. Structure is <key>:<value>")
	rootCmd.MarkFlagRequired("accountName")
	rootCmd.SetVersionTemplate(getVersion())
}

func exec(args arguments) {
	ctx := context.Background()
	var clientError error
	var client *azblob.Client

	metadataFilter := createMetadataFilter(args.MetadataFilter)

	URL, _ := url.Parse(fmt.Sprintf(storageURLTemplate, args.AccountName))

	// if AccessKey is provided use them
	if len(args.AccessKey) > 0 {
		keyCredentials, authErr := azblob.NewSharedKeyCredential(args.AccountName, args.AccessKey)
		if authErr != nil {
			log.Fatal("Error while Authentication with AccessKey", authErr)
		}
		client, clientError = azblob.NewClientWithSharedKeyCredential(URL.String(), keyCredentials, nil)
	} else {
		var authErr error
		var credentials *azidentity.ManagedIdentityCredential

		// if user assigned managed identity is provided use them
		if len(args.MSI) > 0 {
			options := azidentity.ManagedIdentityCredentialOptions{}
			options.ID = azidentity.ClientID(args.MSI)
			credentials, authErr = azidentity.NewManagedIdentityCredential(&options)
		} else {
			// for system assigned managed identity we don't need to pass anything
			credentials, authErr = azidentity.NewManagedIdentityCredential(nil)
		}

		if authErr != nil {
			log.Fatal("Error while Authentication with DefaultCredentials", authErr)
		}

		client, clientError = azblob.NewClient(URL.String(), credentials, nil)
	}

	if clientError != nil {
		log.Fatal("Error while initializing client", clientError)
	}

	containerPager := client.NewListContainersPager(&azblob.ListContainersOptions{
		Include: azblob.ListContainersInclude{
			Metadata: true, Deleted: false,
		},
		Marker: new(string),
	})

	for containerPager.More() {
		page, pageErr := containerPager.NextPage(ctx)
		if pageErr != nil {
			log.Fatal(pageErr.Error())
		}
		for _, container := range page.ContainerItems {
			// TODO substring match? to match containers: ['test-1', 'test-2'], term: 'test, matches ['test-1', 'test-2']
			if len(args.ContainerName) == 0 || strings.Contains(*container.Name, args.ContainerName) {
				foundBlobs := queryBlobs(ctx, *container.Name, args.BlobName, args.ShowContent, client, metadataFilter)
				c := myContainer{
					Name:  args.ContainerName,
					Blobs: foundBlobs,
				}
				foundContainer = append(foundContainer, c)
			}
		}

		if page.NextMarker != nil && len(*page.NextMarker) != 0 {
			containerPager = client.NewListContainersPager(&azblob.ListContainersOptions{
				Include: azblob.ListContainersInclude{
					Metadata: true, Deleted: false,
				},
				Marker: page.NextMarker,
			})
		}
	}

	sa := storageAccount{
		Name:      args.AccountName,
		Container: foundContainer,
	}
	print(sa)
}

func print(sa storageAccount) {
	m, _ := json.Marshal(sa)
	fmt.Println(string(m))
}

func main() {
	rootCmd.Execute()
}
