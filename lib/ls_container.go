package lib

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
)

type container struct {
	Name          string    `json:"name"`
	LastModified  time.Time `json:"lastModified"`
	Etag          string    `json:"etag"`
	LeaseStatus   string    `json:"leaseStatus,omitempty"`
	LeaseState    string    `json:"leaseState,omitempty"`
	LeaseDuration string    `json:"leaseDuration,omitempty"`
}

func newContainer(c storage.Container) *container {
	return &container{
		Name:          c.Name,
		LastModified:  parseLastModified(c.Properties.LastModified),
		Etag:          c.Properties.Etag,
		LeaseStatus:   c.Properties.LeaseStatus,
		LeaseState:    c.Properties.LeaseState,
		LeaseDuration: c.Properties.LeaseDuration,
	}
}

func (cmd *SimpleCommand) listContainers() error {
	// get the client
	client, err := cmd.config.getBlobStorageClient()
	if err != nil {
		return err
	}

	arr, err := listContainersInternal(client, cmd.source.Container)
	if err != nil {
		return err
	}

	// list blobs if there was a direct match on the container
	if len(arr) == 1 && cmd.source.Container == arr[0].Name {
		return cmd.listBlobs()
	}

	listContainersReport(arr, cmd)

	return nil
}

func listContainersInternal(client *storage.BlobStorageClient, namePrefix string) ([]*container, error) {
	// query the endpoint
	params := storage.ListContainersParameters{}
	res, err := client.ListContainers(params)
	if err != nil {
		return nil, err
	}

	// flatten results
	arr := []*container{}
	for _, u := range res.Containers {
		if strings.HasPrefix(u.Name, namePrefix) {
			arr = append(arr, newContainer(u))
		}
	}

	for res.NextMarker != "" {
		params.Marker = res.NextMarker
		res, err = client.ListContainers(params)
		if err != nil {
			return nil, handleListError(err)
		}

		for _, u := range res.Containers {
			if strings.HasPrefix(u.Name, namePrefix) {
				arr = append(arr, newContainer(u))
			}
		}
	}

	return arr, nil
}

func listContainersReport(arr []*container, cmd Command) {
	if cmd.OutputMode() == "json" {
		tmp := struct {
			StorageAccount string       `json:"storageAccount"`
			Containers     []*container `json:"containers"`
		}{
			StorageAccount: cmd.Config().Name,
			Containers:     arr,
		}

		s, _ := json.Marshal(tmp)
		fmt.Printf("%s\n", s)
	} else {
		for _, u := range arr {
			fmt.Printf("%s\n", u.Name)
		}
		fmt.Printf("Found %d containers\n", len(arr))
	}
}
