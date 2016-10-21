package lib

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/alecthomas/units"
)

type SizeCommand struct {
	config     *AzbConfig
	sources    []*BlobSpec
	outputMode string
	workers    int
}

// Command interface
func (cmd *SizeCommand) SetConfig(cfg *AzbConfig)  { cmd.config = cfg }
func (cmd *SizeCommand) Config() *AzbConfig        { return cmd.config }
func (cmd *SizeCommand) AddSource(blob *BlobSpec)  { cmd.sources = append(cmd.sources, blob) }
func (cmd *SizeCommand) SetDst(blob *BlobSpec)     {}
func (cmd *SizeCommand) SetLocalPath(path string)  {}
func (cmd *SizeCommand) SetOutputMode(mode string) { cmd.outputMode = mode }
func (cmd *SizeCommand) OutputMode() string        { return cmd.outputMode }
func (cmd *SizeCommand) SetDestructive(b bool)     {}
func (cmd *SizeCommand) SetWorkers(n int)          { cmd.workers = n }

func (cmd *SizeCommand) Dispatch() error {
	// get the client
	client, err := cmd.config.getBlobStorageClient()
	if err != nil {
		return err
	}

	var blobs []*blob
	var size int64 = 0
	for _, src := range cmd.sources {
		simple := SimpleCommand{
			config:     cmd.config,
			source:     src,
			outputMode: cmd.outputMode,
			workers:    cmd.workers,
		}

		if !src.PathPresent {
			blobs, err = simple.sizeContainers(client)
			if err != nil {
				return err
			}
		} else {
			blobs, err = simple.sizeBlobs(client)
		}

		for _, b := range blobs {
			size += b.ContentLength
		}
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
	var containers []*container

	fmt.Printf("Sizing container %s\n", cmd.source.Container)

	containers, err := listContainersInternal(client, cmd.source.Container)
	if err != nil {
		return nil, err
	}

	numContainers := len(containers)
	containerChan := make(chan string, numContainers)
	blobChan := make(chan []*blob, numContainers)

	for i := 0; i < cmd.workers; i++ {
		go (*cmd).containerWorker(fmt.Sprintf("%s-%d", cmd.source, i), containerChan, blobChan)
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

	fmt.Printf("Finished Sizing container %s\n\n", cmd.source.Container)

	return blobs, nil
}

func (cmd SimpleCommand) containerWorker(id string, containers <-chan string, blobs chan<- []*blob) {
	client, err := cmd.config.getBlobStorageClient()
	if err != nil {
		panic("Failed to get blob storage client in worker.  Everything is fucked")
	}

	for c := range containers {
		params := storage.ListBlobsParameters{Prefix: cmd.source.Path, MaxResults: 5000}
		fmt.Printf("Worker %s enumerating container %s\n", id, c)

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

	fmt.Printf("Worker %s exiting\n", id)
}

func (cmd *SimpleCommand) sizeBlobs(client *storage.BlobStorageClient) ([]*blob, error) {
	res, err := cmd.listBlobsInternal(client)
	if err != nil {
		return nil, err
	}

	return res, nil
}
