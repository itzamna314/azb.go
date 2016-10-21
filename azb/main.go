package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/itzamna314/azb.go/lib"
)

const (
	ProgramVersion string = "azb version 1.1.0"
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

	cmd, err := CreateCommand(conf, res)
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

func CreateCommand(cfg *lib.AzbConfig, res map[string]interface{}) (lib.Command, error) {
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

	var cmd lib.Command

	var blobSrc, blobDst, localPath *string
	requireBlobPath := false

	// dispatch ls
	switch {
	case res["ls"].(bool):
		cmd = &lib.SimpleCommand{Command: "ls"}
		blobSrc = stringOrDefault("<blobspec>", res, true)
		break
	case res["tree"].(bool):
		cmd = &lib.SimpleCommand{Command: "tree"}
		blobSrc = stringOrDefault("<container>", res, true)
		break
	case res["get"].(bool):
		cmd = &lib.SimpleCommand{Command: "get"}
		blobSrc = stringOrDefault("<blobpath>", res, true)
		localPath = stringOrDefault("<dst>", res, false)
		requireBlobPath = true
		break
	case res["rm"].(bool):
		cmd = &lib.SimpleCommand{Command: "rm"}
		blobSrc = stringOrDefault("<blobpath>", res, true)
		requireBlobPath = true
		break
	case res["put"].(bool):
		cmd = &lib.SimpleCommand{Command: "put"}
		blobDst = stringOrDefault("<blobpath>", res, true)
		localPath = stringOrDefault("<src>", res, true)
		break
	case res["size"].(bool):
		cmd = &lib.SizeCommand{}
		// Special handling - size accepts a slice of blobspec
		blobSrcs := stringsOrDefault("<blobspecs>", res, true)
		for _, src := range blobSrcs {
			bs, err := blobSpec(src, requireBlobPath)
			if err != nil {
				return nil, err
			}
			cmd.AddSource(bs)
		}
		break
	}

	cmd.SetConfig(cfg)
	cmd.SetOutputMode(mode)
	cmd.SetWorkers(w)
	cmd.SetDestructive(res["-f"].(bool))

	if blobSrc != nil {
		src, err := blobSpec(*blobSrc, requireBlobPath)
		if err != nil {
			return nil, err
		}

		cmd.AddSource(src)
	}

	if blobDst != nil {
		dst, err := blobSpec(*blobDst, false)
		if err != nil {
			return nil, err
		}

		cmd.SetDst(dst)
	}

	if localPath != nil {
		cmd.SetLocalPath(*localPath)
	}

	return cmd, nil
}

func stringOrDefault(key string, dict map[string]interface{}, stdIn bool) (s *string) {
	s = new(string)

	if stdIn && dict["-"].(bool) {
		*s = readStdIn()
		return
	}

	if str, ok := dict[key].(string); ok {
		if stdIn && str == "-" {
			*s = readStdIn()
		} else {
			*s = str
		}

		return
	}

	*s = ""
	return
}

func stringsOrDefault(key string, dict map[string]interface{}, stdIn bool) (s []string) {
	if stdIn && dict["-"].(bool) {
		rawStr := readStdIn()
		s = trimSplit(rawStr)
		return
	}

	if strs, ok := dict[key].([]string); ok {
		if stdIn {
			if len(strs) == 1 && strs[0] == "-" {
				rawStr := readStdIn()
				s = trimSplit(rawStr)
				return
			}
		}

		s = strs
		return
	}

	return []string{}
}

func readStdIn() string {
	rdr := bufio.NewReader(os.Stdin)
	if in, err := rdr.ReadString('\n'); err == nil {
		return in[:len(in)-1]
	}

	return ""
}

func trimSplit(rawStr string) (s []string) {
	splits := strings.Split(rawStr, " ")
	for _, spl := range splits {
		if strings.Trim(spl, " \t\n\r") != "" {
			s = append(s, spl)
		}
	}
	return
}
