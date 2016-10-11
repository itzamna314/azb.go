package azb

import (
	// "fmt"
	"github.com/MSOpenTech/azure-sdk-for-go/storage"
	"os"
	"path/filepath"
)

func (cmd *SimpleCommand) putBlob() error {

	container := cmd.Destination.Container
	remotePath := cmd.Destination.Path

	if !cmd.Destination.PathPresent {
		_, remotePath = filepath.Split(cmd.LocalPath)
	}

	// fmt.Printf("Would upload %s to %s/%s", cmd.LocalPath, container, remotePath)
	// os.Exit(2)

	// open the local file to be uploaded
	f, err := os.Open(cmd.LocalPath)
	if err != nil {
		return err
	}

	defer f.Close()

	// get the client
	client, err := cmd.getBlobStorageClient()
	if err != nil {
		return err
	}

	// create the blob, if it doesn't exists (Block Blob)
	err = client.CreateBlockBlob(container, remotePath)
	if err != nil && err != storage.ErrNotCreated {
		return err
	}

	// upload the blob (Block Blob)
	err = client.PutBlockBlob(container, remotePath, f)
	if err != nil {
		return err
	}

	return nil
}
