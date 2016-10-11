package azb

import (
	"errors"
	"time"

	"github.com/Azure/azure-sdk-for-go/management"
	"github.com/Azure/azure-sdk-for-go/management/storageservice"
	"github.com/Azure/azure-sdk-for-go/storage"
)

var (
	ErrUnrecognizedCommand     = errors.New("unrecognized command")
	ErrContainerNotFound       = errors.New("container not found")
	ErrContainerOrBlobNotFound = errors.New("container or blob not found")
)

type SimpleCommand struct {
	Config      *AzbConfig
	Command     string
	Source      *BlobSpec
	Destination *BlobSpec
	LocalPath   string
	OutputMode  string
	Destructive bool
}

func (cmd *SimpleCommand) Dispatch() error {
	switch cmd.Command {
	case "ls":
		return cmd.ls()
	case "tree":
		return cmd.tree()
	case "get":
		return cmd.pull()
	case "rm":
		return cmd.rm()
	case "put":
		return cmd.put()
	default:
		return ErrUnrecognizedCommand
	}
}

func (cmd *SimpleCommand) ls() error {
	if cmd.Source == nil || cmd.Destination != nil {
		return ErrUnrecognizedCommand
	}

	if cmd.Source.PathPresent {
		return cmd.listBlobs()
	} else {
		return cmd.listContainers()
	}
}

func (cmd *SimpleCommand) rm() error {
	if cmd.Source == nil || cmd.Destination != nil {
		return ErrUnrecognizedCommand
	}

	return cmd.rmBlob()
}

func (cmd *SimpleCommand) tree() error {
	if cmd.Source == nil || cmd.Destination != nil {
		return ErrUnrecognizedCommand
	}

	if cmd.Source.PathPresent {
		return ErrUnrecognizedCommand
	}

	return cmd.treeBlobs()
}

func (cmd *SimpleCommand) pull() error {
	if cmd.Source == nil {
		return ErrUnrecognizedCommand
	}

	return cmd.pullBlob()
}

func (cmd *SimpleCommand) put() error {
	if cmd.Destination == nil || cmd.LocalPath == "" {
		return ErrUnrecognizedCommand
	}

	return cmd.putBlob()
}

func (cmd *SimpleCommand) getStorageService() (*storageservice.StorageServiceClient, error) {
	cli, err := management.NewClient(cmd.Config.Name, cmd.Config.ManagementCertificate)
	if err != nil {
		return nil, err
	}

	stor := storageservice.NewClient(cli)

	return &stor, nil
}

func (cmd *SimpleCommand) getBlobStorageClient() (*storage.BlobStorageClient, error) {
	stor, err := storage.NewBasicClient(cmd.Config.Name, cmd.Config.AccessKey)
	if err != nil {
		return nil, err
	}

	c := stor.GetBlobService()
	return &c, nil
}

func parseLastModified(s string) time.Time {
	d, err := time.Parse("Mon, 02 Jan 2006 15:04:05 MST", s)
	if err != nil {
		return time.Time{}
	}

	return d
}
