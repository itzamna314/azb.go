package lib

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/alecthomas/units"
)

const maxBlockSize = 4 * units.KB

func (cmd *SimpleCommand) putBlob() error {

	container := cmd.destination.Container
	remotePath := cmd.destination.Path

	if !cmd.destination.PathPresent {
		_, remotePath = filepath.Split(cmd.localPath)
	}

	// fmt.Printf("Would upload %s to %s/%s", cmd.LocalPath, container, remotePath)
	// os.Exit(2)

	// open the local file to be uploaded
	f, err := os.Open(cmd.localPath)
	if err != nil {
		return err
	}

	defer f.Close()

	// get the client
	client, err := cmd.config.getBlobStorageClient()
	if err != nil {
		return err
	}

	// create the blob, if it doesn't exists (Block Blob)
	err = client.CreateBlockBlob(container, remotePath)
	if err != nil {
		return err
	}

	buf := make([]byte, maxBlockSize, maxBlockSize)
	var rdErr error = nil
	var blocks []storage.Block

	// upload the blob one block at a time
	for i := 0; rdErr == nil && err == nil; i++ {
		_, rdErr = f.Read(buf)
		blockId := fmt.Sprintf("%10i", i)
		err = client.PutBlock(container, remotePath, blockId, buf[:])
		blocks = append(blocks, storage.Block{ID: blockId})
	}

	if rdErr != io.EOF {
		return rdErr
	}

	if err != nil {
		return err
	}

	if err = client.PutBlockList(container, remotePath, blocks); err != nil {
		return err
	}

	return nil
}
