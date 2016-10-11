package azb

import (
	"encoding/json"
	"fmt"
	"github.com/MSOpenTech/azure-sdk-for-go/storage"
	"time"
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

	arr, err := cmd.listBlobsInternal()
	if err != nil {
		return err
	}

	cmd.listBlobsReport(arr)

	return nil
}

func (cmd *SimpleCommand) listBlobsReport(arr []*blob) {
	if cmd.OutputMode == "json" {
		tmp := struct {
			StorageAccount string  `json:"storageAccount"`
			Container      string  `json:"container"`
			Blobs          []*blob `json:"blobs"`
		}{
			StorageAccount: cmd.Config.Name,
			Container:      cmd.Source.Container,
			Blobs:          arr,
		}

		s, _ := json.Marshal(tmp)
		fmt.Printf("%s\n", s)
	} else {
		fmt.Printf("total %d\n", len(arr))
		for _, u := range arr {
			fmt.Printf("%s\n", u.Name)
		}
	}
}

func (cmd *SimpleCommand) listBlobsInternal() ([]*blob, error) {
	// get the client
	client, err := cmd.getBlobStorageClient()
	if err != nil {
		return nil, err
	}

	// query the endpoint
	res, err := client.ListBlobs(cmd.Source.Container, storage.ListBlobsParameters{Prefix: cmd.Source.Path})
	if err != nil {
		if sse, ok := err.(storage.StorageServiceError); ok {
			switch sse.Code {
			case "ContainerNotFound":
				return nil, ErrContainerNotFound
			}
		}
		return nil, err
	}

	if res.Marker != "" || res.NextMarker != "" {
		fmt.Printf("\n---\nmarker: %s\nnext marker: %s\n---\n\n", res.Marker, res.NextMarker)
	}

	// flatten results
	arr := []*blob{}
	for _, u := range res.Blobs {
		arr = append(arr, newBlob(u))
	}

	return arr, nil
}
