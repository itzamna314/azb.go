package main

import (
	"europium.io/x/azb"
	"fmt"
	"github.com/docopt/docopt-go"
	"os"
)

const (
	ProgramVersion string = "azb version 0.1.0"
)

func main() {
	parseOpt()
	return

	// err = runCommand(cmd, cmdArgs)
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
}

func parseOpt() (err error) {
	res, err := cmdAzb(os.Args[1:])
	if err != nil {
		cmdAzb([]string{"azb", "--help"})
		os.Exit(1)
	}

	// load config
	configFile := res["-F"].(string)
	environment := res["-e"].(string)

	conf, err := azb.GetConfig(configFile, environment)
	if err != nil {
		panic(err)
	}

	// detect mode
	mode := "shell"
	if res["--json"].(bool) {
		mode = "json"
	}

	// detect ls
	if res["ls"].(bool) {
		cmd := &azb.SimpleCommand{
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
		if err == azb.ErrContainerNotFound {
			fmt.Println("azb ls: No such container")
			os.Exit(1)
		} else if err != nil {
			panic(err)
		}
	}

	if res["pull"].(bool) {
		cmd := &azb.SimpleCommand{
			Config:     conf,
			Command:    "pull",
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
			cmd.LocalPath = fmt.Sprintf("%s/%s", src.Container, src.Path)
		}

		err = cmd.Dispatch()
		if err == azb.ErrContainerOrBlobNotFound {
			fmt.Println("azb pull: No such container or blob")
			os.Exit(1)
		} else if err != nil {
			panic(err)
		}
	}

	return nil
}

func cmdAzb(argv []string) (map[string]interface{}, error) {
	usage := `azb - an uncomplicated azure blob storage client

Usage:
  azb [ -F configFile ] [ -e environment ] [ --json ] ls [ <blobspec> ] 
  azb [ -F configFile ] [ -e environment ] [ --json ] pull <blobpath> [ <dst> ]
  azb [ -F configFile ] [ -e environment ] [ --json ] push [ -R ] <blobpath> [ <src> ]
  azb [ -F configFile ] [ -e environment ] [ --json ] rm <blobpath>
  azb [ -F configFile ] [ -e environment ] [ --json ] cp <srcblobpath> <dstblobpath>
  azb [ -F configFile ] [ -e environment ] [ --json ] mv <srcblobpath> <dstblobpath>
  azb -h | --help
  azb --version

Arguments:
  container     	The name of the container to query
  blobspec      	A reference to one or more blobs (e.g. "mycontainer/foo*", "mycontainer/")
  blobpath			The path of a blob (e.g. "mycontainer/foo.txt")

Options:
  --add=NAME    	Creates a container
  -e environment    Specifies the Azure Storage Services account to use [default: default]
  -F configFile  	Specifies an alternative per-user configuration file [default: /etc/azb/config]
  -h, --help     	Show this screen.
  --list        	List all of the containers in the environment
  --rm=NAME     	Destroys a container
  --version     	Show version.

The most commonly used azb commands are:
   ls         	Lists blobs
   container	List, create, and destroy containers

See 'git help <command>' for more information on a specific command.
`

	dict, err := docopt.Parse(usage, argv, true, ProgramVersion, false)
	if err != nil {
		return nil, err
	}

	//fmt.Println("cmdRoot says:")
	//fmt.Printf("dict= %v\n", dict)

	return dict, err
}

func blobSpec(res map[string]interface{}, key string, pathPresent bool) (*azb.BlobSpec, error) {
	s, ok := res[key].(string)
	if !ok {
		s = ""
	}

	src, err := azb.ParseBlobSpec(s)
	if err != nil {
		return nil, err
	} else if pathPresent && !src.PathPresent {
		fmt.Println("azb: operation requires a fully-qualified path (e.g. foo/bar.txt)")
		os.Exit(1)
	}

	return src, nil
}
