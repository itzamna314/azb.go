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
		Config:     conf,
		OutputMode: mode,
		Workers:    w,
	}

	// dispatch ls
	if res["ls"].(bool) {
		cmd.Command = "ls"

		src, err := blobSpec(res, "<blobspec>", false)
		if err != nil {
			return err
		}

		cmd.Source = src
	}

	// dispatch tree
	if res["tree"].(bool) {
		cmd.Command = "tree"

		src, err := blobSpec(res, "<container>", false)
		if err != nil {
			return err
		}

		cmd.Source = src
	}

	// dispatch get
	if res["get"].(bool) {
		cmd.Command = "get"

		src, err := blobSpec(res, "<blobpath>", true)
		if err != nil {
			return err
		}

		cmd.Source = src

		if dst, ok := res["<dst>"].(string); ok {
			cmd.LocalPath = dst
		} else {
			cmd.LocalPath = ""
		}
	}

	// dispatch rm
	if res["rm"].(bool) {
		cmd.Command = "rm"

		if res["-f"].(bool) {
			cmd.Destructive = true
		}

		src, err := blobSpec(res, "<blobpath>", true)
		if err != nil {
			return err
		}

		cmd.Source = src
	}

	if res["put"].(bool) {
		cmd.Command = "put"

		dst, err := blobSpec(res, "<blobpath>", false)
		if err != nil {
			return err
		}

		cmd.Destination = dst

		if path, ok := res["<src>"].(string); ok {
			cmd.LocalPath = path
		} else {
			cmd.LocalPath = fmt.Sprintf("%s/%s", dst.Container, dst.Path)
		}
	}

	if res["size"].(bool) {
		cmd.Command = "size"

		src, err := blobSpec(res, "<blobspec>", false)
		if err != nil {
			return err
		}

		cmd.Source = src
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

	//fmt.Println("cmdRoot says:")
	//fmt.Printf("dict= %v\n", dict)

	return dict, err
}

func blobSpec(res map[string]interface{}, key string, pathPresent bool) (*lib.BlobSpec, error) {
	s, ok := res[key].(string)
	if !ok {
		s = ""
	}

	src, err := lib.ParseBlobSpec(s)
	if err != nil {
		return nil, err
	} else if pathPresent && !src.PathPresent {
		fmt.Println("azb: operation requires a fully-qualified path (e.g. foo/bar.txt)")
		os.Exit(1)
	}

	return src, nil
}
