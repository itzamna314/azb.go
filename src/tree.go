package azb

import (
	"encoding/json"
	"fmt"
	"github.com/MSOpenTech/azure-sdk-for-go/storage"
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

func printTree(node *node, indent int) (nd, nf int) {
	fmt.Println(node.Name)

	if node.Nodes == nil {
		return 0, 1
	}

	i := 1
	nd = 0
	nf = 0

	for _, v := range node.Nodes {

		for j := 0; j < indent; j++ {
			fmt.Printf("│   ")
		}

		if i < node.Len() {
			fmt.Printf("├── ")
		} else {
			fmt.Printf("└── ")
		}

		xd, xf := printTree(v, indent+1)

		nd = nd + xd
		nf = nf + xf

		i++
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
		nd, nf := printTree(root, 0)

		fmt.Printf("\n%d directories, %d files\n", nd, nf)
	}
}
