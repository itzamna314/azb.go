package lib

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/storage"
)

func (cmd *SimpleCommand) pullBlob() error {

	// get the client
	client, err := cmd.config.getBlobStorageClient()
	if err != nil {
		return err
	}

	// query the endpoint
	body, err := client.GetBlob(cmd.source.Container, cmd.source.Path)
	if err != nil {
		if sse, ok := err.(storage.AzureStorageServiceError); ok {
			switch sse.Code {
			case "ContainerNotFound":
				return ErrContainerOrBlobNotFound
			case "BlobNotFound":
				return ErrContainerOrBlobNotFound
			}
		}
		return err
	}

	if cmd.localPath == "" {
		// echo content to stdout
		_, err := io.Copy(os.Stdout, body)
		if err != nil {
			return err
		}
	} else {
		// prepare the download location
		dir := filepath.Dir(cmd.localPath)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		// put the file on disk
		f, err := os.Create(cmd.localPath)
		if err != nil {
			return err
		}

		defer f.Close()

		written, err := io.Copy(f, body)
		if err != nil {
			return err
		}

		// tell the world about it
		cmd.pullBlobReport(written)
	}

	return nil
}

func (cmd *SimpleCommand) pullBlobReport(written int64) {
	if cmd.outputMode == "json" {
		tmp := struct {
			StorageAccount string `json:"storageAccount"`
			Container      string `json:"container"`
			Blob           string `json:"blob"`
			BytesWritten   int64  `json:"bytesWritten"`
			Destination    string `json:"destination"`
		}{
			StorageAccount: cmd.config.Name,
			Container:      cmd.source.Container,
			Blob:           cmd.source.Path,
			BytesWritten:   written,
			Destination:    cmd.localPath,
		}

		s, _ := json.Marshal(tmp)
		cmd.logger.Info("%s\n", s)
	}
}
