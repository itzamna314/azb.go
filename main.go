package azb

import (
	"fmt"
	"os"

	"github.com/docopt/docopt-go"
)

const (
	ProgramVersion string = "azb version 0.2.0"
)

func main() {
	err := doit()
	if err != nil {
		panic(err)
	}
}

func doit() (err error) {
	res, err := usage(os.Args[1:])
	if err != nil {
		usage([]string{"azb", "--help"})
		os.Exit(1)
	}

	// load config
	configFile := res["-F"].(string)
	environment := res["-e"].(string)

	conf, err := GetConfig(configFile, environment)
	if err != nil {
		return err
	}

	// detect mode
	mode := "bare"
	if res["--json"].(bool) {
		mode = "json"
	}

	// dispatch ls
	if res["ls"].(bool) {
		cmd := &SimpleCommand{
			Config:     conf,
			Command:    "ls",
			OutputMode: mode,
		}

		src, err := blobSpec(res, "<blobspec>", false)
		if err != nil {
			return err
		}

		cmd.Source = src

		err = cmd.Dispatch()
		if err == ErrContainerNotFound {
			fmt.Println("azb ls: No such container")
			os.Exit(1)
		} else if err == ErrUnrecognizedCommand {
			fmt.Println("azb ls: unexpected arguments")
			os.Exit(1)
		} else if err != nil {
			return err
		}
	}

	// dispatch tree
	if res["tree"].(bool) {
		cmd := &SimpleCommand{
			Config:     conf,
			Command:    "tree",
			OutputMode: mode,
		}

		src, err := blobSpec(res, "<container>", false)
		if err != nil {
			return err
		}

		cmd.Source = src

		err = cmd.Dispatch()
		if err == ErrContainerNotFound {
			fmt.Println("azb tree: No such container")
			os.Exit(1)
		} else if err == ErrUnrecognizedCommand {
			fmt.Println("azb tree: unexpected arguments")
			os.Exit(1)
		} else if err != nil {
			return err
		}
	}

	// dispatch get
	if res["get"].(bool) {
		cmd := &SimpleCommand{
			Config:     conf,
			Command:    "get",
			OutputMode: mode,
		}

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

		err = cmd.Dispatch()
		if err == ErrContainerOrBlobNotFound {
			fmt.Println("azb get: No such container or blob")
			os.Exit(1)
		} else if err != nil {
			return err
		}
	}

	// dispatch rm
	if res["rm"].(bool) {
		cmd := &SimpleCommand{
			Config:     conf,
			Command:    "rm",
			OutputMode: mode,
		}

		if res["-f"].(bool) {
			cmd.Destructive = true
		}

		src, err := blobSpec(res, "<blobpath>", true)
		if err != nil {
			return err
		}

		cmd.Source = src

		err = cmd.Dispatch()
		if err == ErrContainerOrBlobNotFound {
			fmt.Println("azb rm: No such container or blob")
			os.Exit(1)
		} else if err != nil {
			return err
		}
	}

	if res["put"].(bool) {
		cmd := &SimpleCommand{
			Config:     conf,
			Command:    "put",
			OutputMode: mode,
		}

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

		err = cmd.Dispatch()
		if err == ErrContainerOrBlobNotFound {
			fmt.Println("azb put: No such container or blob")
			os.Exit(1)
		} else if err != nil {
			return err
		}
	}

	return nil
}

func usage(argv []string) (map[string]interface{}, error) {
	usage := `azb - an uncomplicated azure blob storage client

Usage:
  azb [ -F configFile ] [ -e environment ] [ --json ] ls [ <blobspec> ]
  azb [ -F configFile ] [ -e environment ] [ --json ] tree <container>
  azb [ -F configFile ] [ -e environment ] [ --json ] get <blobpath> [ <dst> ]
  azb [ -F configFile ] [ -e environment ] [ --json ] put <blobpath> [ <src> ]
  azb [ -F configFile ] [ -e environment ] [ --json ] rm [ -f ] <blobpath>
  azb -h | --help
  azb --version

Arguments:
  container     	The name of the container to query
  blobspec      	A reference to one or more blobs (e.g. "mycontainer/foo", "mycontainer/")
  blobpath			The path of a blob (e.g. "mycontainer/foo.txt")

Options:
  -e environment    Specifies the Azure Storage Services account to use [default: default]
  -F configFile  	Specifies an alternative per-user configuration file [default: /usr/local/etc/azb/config]
  -f                Forces a destructive operation
  -h, --help     	Show this screen.
  --version     	Show version.

The most commonly used commands are:
   ls         	Lists containers and blobs
   get          Downloads a blob
   put          Uploads a blob
   tree         Prints the contents of a container as a tree
   rm           Deletes a blob
`

	dict, err := docopt.Parse(usage, argv, true, ProgramVersion, false)
	if err != nil {
		return nil, err
	}

	//fmt.Println("cmdRoot says:")
	//fmt.Printf("dict= %v\n", dict)

	return dict, err
}

func blobSpec(res map[string]interface{}, key string, pathPresent bool) (*BlobSpec, error) {
	s, ok := res[key].(string)
	if !ok {
		s = ""
	}

	src, err := ParseBlobSpec(s)
	if err != nil {
		return nil, err
	} else if pathPresent && !src.PathPresent {
		fmt.Println("azb: operation requires a fully-qualified path (e.g. foo/bar.txt)")
		os.Exit(1)
	}

	return src, nil
}
