package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/storage"
)

func (cmd *SimpleCommand) pullBlob() error {

	// get the client
	client, err := cmd.getBlobStorageClient()
	if err != nil {
		return err
	}

	// query the endpoint
	body, err := client.GetBlob(cmd.Source.Container, cmd.Source.Path)
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

	if cmd.LocalPath == "" {
		// echo content to stdout
		_, err := io.Copy(os.Stdout, body)
		if err != nil {
			return err
		}
	} else {
		// prepare the download location
		dir := filepath.Dir(cmd.LocalPath)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}

		// put the file on disk
		f, err := os.Create(cmd.LocalPath)
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
	if cmd.OutputMode == "json" {
		tmp := struct {
			StorageAccount string `json:"storageAccount"`
			Container      string `json:"container"`
			Blob           string `json:"blob"`
			BytesWritten   int64  `json:"bytesWritten"`
			Destination    string `json:"destination"`
		}{
			StorageAccount: cmd.Config.Name,
			Container:      cmd.Source.Container,
			Blob:           cmd.Source.Path,
			BytesWritten:   written,
			Destination:    cmd.LocalPath,
		}

		s, _ := json.Marshal(tmp)
		fmt.Printf("%s\n", s)
	}
}
