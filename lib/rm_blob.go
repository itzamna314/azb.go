package lib

func (cmd *SimpleCommand) rmBlob() error {

	// get the client
	client, err := cmd.config.getBlobStorageClient()
	if err != nil {
		return err
	}

	if cmd.destructive == false {
		cmd.logger.Info("Would remove %s\n", cmd.source.Path)
		return nil
	}

	// query the endpoint
	extraHeaders := map[string]string{}
	_, err = client.DeleteBlobIfExists(cmd.source.Container, cmd.source.Path, extraHeaders)
	if err != nil {
		return err
	}

	return nil
}
