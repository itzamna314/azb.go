package lib

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
)

type blob struct {
	Name            string    `json:"name"`
	LastModified    time.Time `json:"lastModified"`
	Etag            string    `json:"etag"`
	ContentLength   int64     `json:"contentLength"`
	ContentType     string    `json:"contentType"`
	ContentEncoding string    `json:"contentEncoding"`
}

func newBlob(c storage.Blob) *blob {
	return &blob{
		Name:            c.Name,
		LastModified:    parseLastModified(c.Properties.LastModified),
		Etag:            c.Properties.Etag,
		ContentLength:   c.Properties.ContentLength,
		ContentType:     c.Properties.ContentType,
		ContentEncoding: c.Properties.ContentEncoding,
	}
}

func (cmd *SimpleCommand) listBlobs() error {
	// get the client
	client, err := cmd.config.getBlobStorageClient()
	if err != nil {
		return err
	}

	arr, err := cmd.listBlobsInternal(client)
	if err != nil {
		return err
	}

	cmd.listBlobsReport(arr)

	return nil
}

func (cmd *SimpleCommand) listBlobsReport(arr []*blob) {
	if cmd.outputMode == "json" {
		tmp := struct {
			StorageAccount string  `json:"storageAccount"`
			Container      string  `json:"container"`
			Blobs          []*blob `json:"blobs"`
		}{
			StorageAccount: cmd.config.Name,
			Container:      cmd.source.Container,
			Blobs:          arr,
		}

		s, _ := json.Marshal(tmp)
		fmt.Printf("%s\n", s)
	} else {
		for _, u := range arr {
			fmt.Printf("%s\n", u.Name)
		}
		fmt.Printf("Found %d blobs\n", len(arr))
	}
}

func (cmd *SimpleCommand) listBlobsInternal(client *storage.BlobStorageClient) ([]*blob, error) {
	// query the endpoint
	params := storage.ListBlobsParameters{Prefix: cmd.source.Path}
	res, err := client.ListBlobs(cmd.source.Container, params)
	if err != nil {
		return nil, handleListError(err)
	}

	arr := []*blob{}
	for _, u := range res.Blobs {
		arr = append(arr, newBlob(u))
	}

	for res.NextMarker != "" {
		params.Marker = res.NextMarker
		res, err = client.ListBlobs(cmd.source.Container, params)
		if err != nil {
			return nil, handleListError(err)
		}

		for _, u := range res.Blobs {
			arr = append(arr, newBlob(u))
		}
	}

	return arr, nil
}

func handleListError(err error) error {
	if err != nil {
		if sse, ok := err.(storage.AzureStorageServiceError); ok {
			switch sse.Code {
			case "ContainerNotFound":
				return ErrContainerNotFound
			}
		}
		return err
	}

	return nil
}
