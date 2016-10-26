package lib

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

type Command interface {
	Dispatch() error
	SetConfig(cfg *AzbConfig)
	Config() *AzbConfig
	AddSource(blob *BlobSpec)
	SetDst(blob *BlobSpec)
	SetLocalPath(path string)
	SetOutputMode(mode string)
	OutputMode() string
	SetDestructive(isDestructive bool)
	SetWorkers(n int)
	SetLogger(l Logger)
	Logger() Logger
}

type SimpleCommand struct {
	config      *AzbConfig
	Command     string
	source      *BlobSpec
	destination *BlobSpec
	localPath   string
	outputMode  string
	destructive bool
	workers     int
	logger      Logger
}

// Command interface
func (cmd *SimpleCommand) SetConfig(cfg *AzbConfig)  { cmd.config = cfg }
func (cmd *SimpleCommand) Config() *AzbConfig        { return cmd.config }
func (cmd *SimpleCommand) AddSource(blob *BlobSpec)  { cmd.source = blob }
func (cmd *SimpleCommand) SetDst(blob *BlobSpec)     { cmd.destination = blob }
func (cmd *SimpleCommand) SetLocalPath(path string)  { cmd.localPath = path }
func (cmd *SimpleCommand) SetOutputMode(mode string) { cmd.outputMode = mode }
func (cmd *SimpleCommand) OutputMode() string        { return cmd.outputMode }
func (cmd *SimpleCommand) SetDestructive(b bool)     { cmd.destructive = b }
func (cmd *SimpleCommand) SetWorkers(n int)          { cmd.workers = n }
func (cmd *SimpleCommand) SetLogger(l Logger)        { cmd.logger = l }
func (cmd *SimpleCommand) Logger() Logger            { return cmd.logger }

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
	if cmd.source == nil || cmd.destination != nil {
		return ErrUnrecognizedCommand
	}

	if cmd.source.PathPresent {
		return cmd.listBlobs()
	} else {
		return cmd.listContainers()
	}
}

func (cmd *SimpleCommand) rm() error {
	if cmd.source == nil || cmd.destination != nil {
		return ErrUnrecognizedCommand
	}

	return cmd.rmBlob()
}

func (cmd *SimpleCommand) tree() error {
	if cmd.source == nil || cmd.destination != nil {
		return ErrUnrecognizedCommand
	}

	if cmd.source.PathPresent {
		return ErrUnrecognizedCommand
	}

	return cmd.treeBlobs()
}

func (cmd *SimpleCommand) pull() error {
	if cmd.source == nil {
		return ErrUnrecognizedCommand
	}

	return cmd.pullBlob()
}

func (cmd *SimpleCommand) put() error {
	if cmd.destination == nil || cmd.localPath == "" {
		return ErrUnrecognizedCommand
	}

	return cmd.putBlob()
}

func (cfg *AzbConfig) getStorageService() (*storageservice.StorageServiceClient, error) {
	cli, err := management.NewClient(cfg.Name, cfg.ManagementCertificate)
	if err != nil {
		return nil, err
	}

	stor := storageservice.NewClient(cli)

	return &stor, nil
}

func (cfg *AzbConfig) getBlobStorageClient() (*storage.BlobStorageClient, error) {
	var res error
	for i := 0; i < 3; i++ {
		stor, err := storage.NewBasicClient(cfg.Name, cfg.AccessKey)
		if err != nil {
			res = err
			continue
		}
		c := stor.GetBlobService()
		return &c, nil
	}

	return nil, res
}

func parseLastModified(s string) time.Time {
	d, err := time.Parse("Mon, 02 Jan 2006 15:04:05 MST", s)
	if err != nil {
		return time.Time{}
	}

	return d
}
