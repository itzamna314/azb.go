package lib

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/alecthomas/units"
)

func (cmd *SimpleCommand) size() error {
	// get the client
	client, err := cmd.getBlobStorageClient()
	if err != nil {
		return err
	}

	var blobs []*blob
	if !cmd.Source.PathPresent {
		blobs, err = cmd.sizeContainers(client)
		if err != nil {
			return err
		}
	} else {
		blobs, err = cmd.sizeBlobs(client)
	}

	var size int64 = 0
	for _, b := range blobs {
		size += b.ContentLength
	}

	unitMap := units.MakeUnitMap("B", "b", 1000)
	for k, v := range unitMap {
		szUnit := float64(size) / float64(v)
		if szUnit < 1000 && szUnit > 1 {
			fmt.Printf("Total size: %.2f %s\n", szUnit, k)
			break
		}
	}

	return nil
}

func (cmd *SimpleCommand) sizeContainers(client *storage.BlobStorageClient) ([]*blob, error) {
	var blobs []*blob
	containers, err := cmd.listContainersInternal(client)
	if err != nil {
		return nil, err
	}

	numContainers := len(containers)
	containerChan := make(chan string, numContainers)
	blobChan := make(chan []*blob, numContainers)

	for i := 0; i < cmd.Workers; i++ {
		go (*cmd).containerWorker(i, containerChan, blobChan)
	}

	for _, c := range containers {
		containerChan <- c.Name
	}
	close(containerChan)

	for i := 0; i < numContainers; i++ {
		curBlobs := <-blobChan
		for _, b := range curBlobs {
			blobs = append(blobs, b)
		}
	}

	return blobs, nil
}

func (cmd SimpleCommand) containerWorker(id int, containers <-chan string, blobs chan<- []*blob) {
	client, err := cmd.getBlobStorageClient()
	if err != nil {
		panic("Failed to get blob storage client in worker.  Everything is fucked")
	}

	for c := range containers {
		params := storage.ListBlobsParameters{Prefix: cmd.Source.Path, MaxResults: 5000}
		fmt.Printf("Worker %d enumerating container %s\n", id, c)
		var curBlobs []*blob
		res, err := client.ListBlobs(c, params)
		if err != nil {
			panic("Failed to list blobs in worker.  Everything is fucked")
		}

		// flatten results
		for _, u := range res.Blobs {
			curBlobs = append(curBlobs, newBlob(u))
		}

		for res.NextMarker != "" {
			params.Marker = res.NextMarker
			res, err = client.ListBlobs(c, params)
			if err != nil {
				panic("Failed to list blobs with marker in worker.  Everything is fucked")
			}

			for _, u := range res.Blobs {
				curBlobs = append(curBlobs, newBlob(u))
			}
		}

		blobs <- curBlobs
	}

	fmt.Printf("Worker %d exiting\n", id)
}

func (cmd *SimpleCommand) sizeBlobs(client *storage.BlobStorageClient) ([]*blob, error) {
	res, err := cmd.listBlobsInternal(client)
	if err != nil {
		return nil, err
	}

	return res, nil
}
