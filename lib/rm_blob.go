package lib

import "fmt"

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
	extraHeaders := map[string]string{}
	_, err = client.DeleteBlobIfExists(cmd.Source.Container, cmd.Source.Path, extraHeaders)
	if err != nil {
		return err
	}

	return nil
}
