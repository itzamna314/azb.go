package azb

import (
	"encoding/json"
	"fmt"
	"github.com/MSOpenTech/azure-sdk-for-go/storage"
	"os"
	"sort"
	"strings"
)

type node struct {
	Name  string
	Nodes map[string]*node
}

func dirNode(name string) *node {
	return &node{name, map[string]*node{}}
}

func (n *node) Len() int {
	return len(n.Nodes)
}

func (n *node) AddFile(name string) *node {
	r := &node{name, nil}
	n.Nodes[r.Name] = r
	return r
}

type blobs []*blob

// Ensure it satisfies sort.Interface
func (d blobs) Len() int           { return len(d) }
func (d blobs) Less(i, j int) bool { return d[i].Name < d[j].Name }
func (d blobs) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }

func (cmd *SimpleCommand) treeBlobs() error {
	// get the client
	client, err := cmd.getBlobStorageClient()
	if err != nil {
		return err
	}

	// query the endpoint
	res, err := client.ListBlobs(cmd.Source.Container, storage.ListBlobsParameters{Prefix: cmd.Source.Path})
	if err != nil {
		if sse, ok := err.(storage.StorageServiceError); ok {
			switch sse.Code {
			case "ContainerNotFound":
				return ErrContainerNotFound
			}
		}
		return err
	}

	if res.Marker != "" || res.NextMarker != "" {
		fmt.Printf("\n---\nmarker: %s\nnext marker: %s\n---\n\n", res.Marker, res.NextMarker)
	}

	// flatten results
	arr := blobs{}
	for _, u := range res.Blobs {
		arr = append(arr, newBlob(u))
	}

	// sort results lexicographically
	sort.Sort(arr)

	root := buildTree(arr)

	cmd.treeBlobsReport(root)

	return nil
}

func buildTree(arr blobs) (root *node) {
	root = dirNode(".")

	for _, u := range arr {
		dirs := strings.Split(u.Name, "/")

		curr := root
		for i := 0; i < len(dirs)-1; i++ {
			name := dirs[i]

			next, ok := curr.Nodes[name]
			if !ok {
				next = dirNode(name)
				curr.Nodes[name] = next
			}

			curr = next
		}

		curr.AddFile(dirs[len(dirs)-1])
	}

	return
}

// .
// ├── Goopfile
// ├── Goopfile.lock
// ├── Makefile
// ├── README.md
// ├── TODO
// ├── VERSION
// ├── src
// │   ├── azb.go
// │   ├── blobspec.go
// │   ├── blobspec_test.go
// │   ├── cmd
// │   │   └── azb
// │   │       └── main.go
// │   ├── config.go
// │   ├── ls_blob.go
// │   ├── ls_container.go
// │   └── pull.go
// └── tmp
//     └── azb

// .
// B Goopfile
// B Goopfile.lock
// B Makefile
// B README.md
// B TODO
// B VERSION
// B src
// T B azb.go
// T B blobspec.go
// T B blobspec_test.go
// T B cmd
// T T L azb
// T T A L main.go
// T B config.go
// T B ls_blob.go
// T B ls_container.go
// T L pull.go
// L tmp
// A L azb

const (
	TRUNK  string = "│  "
	BRANCH string = "├──"
	LEAF   string = "└──"
	AIR    string = "   "
	STAR   string = "─x─"
)

func printRoot(node *node) (nd, nf int) {
	return printTree(node, []string{})
}

func pop(arr []string) (head []string, s string) {
	if len(arr) == 0 {
		return []string{}, ""
	} else if len(arr) == 1 {
		return []string{}, arr[0]
	} else {
		return arr[:len(arr)-2], arr[len(arr)-1]
	}
}

func push(arr []string, s ...string) []string {
	for _, u := range s {
		arr = append(arr, u)
	}
	return arr
}

func printTree(node *node, stack []string) (nd, nf int) {
	if len(stack) > 0 {
		fmt.Printf("%s ", strings.Join(stack, " "))
	}

	fmt.Println(node.Name)

	if node.Nodes == nil {
		return 0, 1
	}

	base, top := pop(stack)

	nd = 0
	nf = 0
	zz := node.Len() - 1
	z0 := 0

	for _, v := range node.Nodes {

		switch top {
		case "":
			if z0 < zz {
				stack = push(base, BRANCH)
			} else {
				stack = push(base, LEAF)
			}
			break
		case BRANCH:
			if z0 < zz {
				stack = push(base, AIR, TRUNK, BRANCH)
			} else {
				stack = push(base, AIR, TRUNK, LEAF)
			}
			break
		case LEAF:
			if z0 < zz {
				stack = push(base, AIR, AIR, BRANCH)
			} else {
				stack = push(base, AIR, AIR, LEAF)
			}
			break
		default:
			fmt.Println("\n---")
			fmt.Println("invalid stack: ")
			fmt.Printf("stack=%v, top=%s\n", base, top)
			os.Exit(9)
		}

		xd, xf := printTree(v, stack)

		nd = nd + xd
		nf = nf + xf

		z0++
	}

	return nd + 1, nf
}

func (cmd *SimpleCommand) treeBlobsReport(root *node) {
	if cmd.OutputMode == "json" {
		tmp := struct {
			StorageAccount string `json:"storageAccount"`
			Container      string `json:"container"`
			Root           *node  `json:"tree"`
		}{
			StorageAccount: cmd.Config.Name,
			Container:      cmd.Source.Container,
			Root:           root,
		}

		s, _ := json.Marshal(tmp)
		fmt.Printf("%s\n", s)
	} else {
		nd, nf := printRoot(root)

		fmt.Printf("\n%d directories, %d files\n", nd, nf)
	}
}
