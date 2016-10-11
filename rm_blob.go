package azb

import (
	"fmt"
	"github.com/MSOpenTech/azure-sdk-for-go/storage"
)

func (cmd *SimpleCommand) rmBlob() error {

	// get the client
	client, err := cmd.getBlobStorageClient()
	if err != nil {
		return err
	}

	if cmd.Destructive == false {
		fmt.Printf("Would remove %s\n", cmd.Source.Path)
		return nil
	}

	// query the endpoint
	_, err = client.DeleteBlobIfExists(cmd.Source.Container, cmd.Source.Path)
	if err != nil {
		if sse, ok := err.(storage.StorageServiceError); ok {
			switch sse.Code {
			case "ContainerNotFound":
				return ErrContainerOrBlobNotFound
			case "BlobNotFound":
				return ErrContainerOrBlobNotFound
			}
		}
		return err
	}

	return nil
}
