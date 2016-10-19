package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/docopt/docopt-go"
	"github.com/itzamna314/azb.go/lib"
)

const (
	ProgramVersion string = "azb version 1.0.1"
)

func main() {
	err := doit()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

func doit() (err error) {
	res, err := usage(os.Args[1:])
	if err != nil {
		usage([]string{"azb", "--help"})
		return
	}

	// load config
	configFile := res["-F"].(string)
	environment := res["-e"].(string)

	conf, err := lib.GetConfig(configFile, environment)
	if err != nil {
		return err
	}

	cmd, err := CreateSimpleCommand(conf, res)
	if err != nil {
		return err
	}

	return handleErr(cmd.Dispatch())
}

func handleErr(err error) error {
	if err == lib.ErrContainerOrBlobNotFound {
		fmt.Println("azb: No such container or blob")
		os.Exit(1)
	} else if err == lib.ErrContainerNotFound {
		fmt.Println("azb: No such container")
		os.Exit(1)
	} else if err == lib.ErrUnrecognizedCommand {
		fmt.Println("azb: unexpected arguments")
		os.Exit(1)
	}

	return err
}

func usage(argv []string) (map[string]interface{}, error) {
	dict, err := docopt.Parse(usageMsg, argv, true, ProgramVersion, false)
	if err != nil {
		fmt.Printf("parse error: %s\n", err)
		return nil, err
	}

	return dict, err
}

func blobSpec(path string, requirePath bool) (*lib.BlobSpec, error) {
	src, err := lib.ParseBlobSpec(path)
	if err != nil {
		return nil, err
	} else if requirePath && !src.PathPresent {
		return nil, fmt.Errorf("azb: operation requires a fully-qualified path (e.g. foo/bar.txt)")
	}

	return src, nil
}

func CreateSimpleCommand(cfg *lib.AzbConfig, res map[string]interface{}) (*lib.SimpleCommand, error) {
	// detect mode
	mode := "bare"
	if res["--json"].(bool) {
		mode = "json"
	}

	w, err := strconv.Atoi(res["-w"].(string))
	if err != nil {
		fmt.Printf("Usage: expected -w to be an int, was %s\n", res["-w"].(string))
		os.Exit(1)
	}

	cmd := &lib.SimpleCommand{
		Config:      cfg,
		OutputMode:  mode,
		Workers:     w,
		Destructive: res["-f"].(bool),
	}

	var blobSrc, blobDst, localPath *string
	requireBlobPath := false

	// dispatch ls
	switch {
	case res["ls"].(bool):
		cmd.Command = "ls"
		blobSrc = stringOrDefault("<blobspec>", res)
		break
	case res["tree"].(bool):
		cmd.Command = "tree"
		blobSrc = stringOrDefault("<container>", res)
		break
	case res["get"].(bool):
		cmd.Command = "get"
		blobSrc = stringOrDefault("<blobpath>", res)
		localPath = stringOrDefault("<dst>", res)
		requireBlobPath = true
		break
	case res["rm"].(bool):
		cmd.Command = "rm"
		blobSrc = stringOrDefault("<blobpath>", res)
		requireBlobPath = true
		break
	case res["put"].(bool):
		cmd.Command = "put"
		blobDst = stringOrDefault("<blobpath>", res)
		localPath = stringOrDefault("<src>", res)
		break
	case res["size"].(bool):
		cmd.Command = "size"
		blobSrc = stringOrDefault("<blobspec>", res)
		break
	}

	if blobSrc != nil {
		src, err := blobSpec(*blobSrc, requireBlobPath)
		if err != nil {
			return nil, err
		}

		cmd.Source = src
	}

	if blobDst != nil {
		dst, err := blobSpec(*blobDst, false)
		if err != nil {
			return nil, err
		}

		cmd.Destination = dst
	}

	if localPath != nil {
		cmd.LocalPath = *localPath
	}

	return cmd, nil
}

func stringOrDefault(key string, dict map[string]interface{}) (s *string) {
	s = new(string)
	if str, ok := dict[key].(string); ok {
		*s = str
	} else {
		*s = ""
	}
	return
}
