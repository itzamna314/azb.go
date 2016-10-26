package lib

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/alecthomas/units"
)

type SizeCommand struct {
	config        *AzbConfig
	sources       []*BlobSpec
	outputMode    string
	workers       int
	workerTimeout time.Duration
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
	// default for now
	cmd.workerTimeout = 100 * time.Millisecond

	sourcesChan := make(chan *BlobSpec)
	blobChan := make(chan []*blob)
	// Keep track of sources that need to be expanded.
	// Once we've expanded all necessary sources, we can close sourcesChan.
	expandedChan := make(chan string)
	// Keep track of workers.  When all workers have exited, we can add up
	// the total blob size.
	exitChan := make(chan string)

	// Add an additional worker for expanding each source, to guarantee we have
	// enough workers
	numWorkers := cmd.workers + len(cmd.sources)
	for i := 0; i < numWorkers; i++ {
		id := fmt.Sprintf("%d", i)
		go (*cmd).sizeWorker(id, sourcesChan, blobChan, expandedChan, exitChan)
	}

	// First, send all input sources to our workers
	for _, src := range cmd.sources {
		fmt.Printf("Sending source '%s'\n", src)
		sourcesChan <- src
	}

	fmt.Printf("-------\nDone sending sources\n------\n\n")

	// Count up how many sources do not have a path. We need to expand all
	// such sources.  When we have done so, we can close the sources channel.
	numToExpand := 0
	for _, src := range cmd.sources {
		if !src.PathPresent {
			numToExpand++
		}
	}

	// Count up the size as we go
	var size int64 = 0
	// Wait for our workers to list out blobs.
waitLoop:
	for {
		select {
		case <-expandedChan:
			numToExpand--
			if numToExpand <= 0 {
				fmt.Printf("------\nDone expanding sources\n------\n\n")
				close(sourcesChan)
			}
		case id := <-exitChan:
			numWorkers--
			fmt.Printf("Worker %s exiting. %d still working\n", id, numWorkers)
			// Once all workers have exited, close off blob chan
			if numWorkers <= 0 {
				fmt.Printf("------\nDone walking blobs\n------\n\n")
				close(blobChan)
			}
		case blobs, ok := <-blobChan:
			// Once blob chan is closed and empty, stop receiving
			if !ok {
				break waitLoop
			}

			for _, b := range blobs {
				size += b.ContentLength
			}
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

func (cmd SizeCommand) sizeWorker(id string, sources chan *BlobSpec,
	blobs chan<- []*blob, expanded chan<- string, exited chan<- string) {

	client, err := cmd.config.getBlobStorageClient()
	if err != nil {
		panic("Failed to get blob storage client in worker.  Add retry logic")
	}

	for src := range sources {
		if !src.PathPresent {
			fmt.Printf("Started expanding source %s\n", src)
			// We need to break this source down into all containers it could refer to
			// List all containers for this source, and enqueue them back onto sources
			// Then get out
			sendContainersToChannel(client, sources, src)
			expanded <- id
			fmt.Printf("Finished expanding source %s\n", src)
			continue
		}

		// We have a path present, so we can list all matching blobs and count their
		// size.
		params := storage.ListBlobsParameters{Prefix: src.Path, MaxResults: 5000}
		fmt.Printf("Worker %s started enumerating container %s\n", id, src.Container)

		var curBlobs []*blob
		res := storage.BlobListResponse{}
		for firstTime := true; firstTime || res.NextMarker != ""; firstTime = false {
			res, err := client.ListBlobs(src.Container, params)
			if err != nil {
				panic("Failed to list blobs in worker.  Add retry logic")
			}

			// flatten results
			for _, u := range res.Blobs {
				curBlobs = append(curBlobs, newBlob(u))
			}

			params.Marker = res.NextMarker
		}

		fmt.Printf("Worker %s finished enumerating container %s\n", id, src.Container)
		blobs <- curBlobs
	}

	exited <- id
}

func sendContainersToChannel(client *storage.BlobStorageClient,
	outChan chan<- *BlobSpec, src *BlobSpec) error {

	containers, err := listContainersInternal(client, src.Container)
	if err != nil {
		return err
	}

	for _, c := range containers {
		bs := BlobSpec{
			Container:   c.Name,
			Path:        "",
			PathPresent: true,
		}
		fmt.Printf("Sent container %+v to sources channel\n", bs)
		outChan <- &bs
	}

	return nil
}

func (cmd *SimpleCommand) sizeBlobs(client *storage.BlobStorageClient) ([]*blob, error) {
	res, err := cmd.listBlobsInternal(client)
	if err != nil {
		return nil, err
	}

	return res, nil
}
