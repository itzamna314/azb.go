package lib

import (
	"encoding/json"
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
	client, err := cmd.config.getBlobStorageClient()
	if err != nil {
		return err
	}

	// query the endpoint
	res, err := cmd.listBlobsInternal(client)
	if err != nil {
		return err
	}

	arr := blobs(res)

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

func (cmd *SimpleCommand) printRoot(node *node) (nd, nf int) {
	return cmd.printTree(node, &Stack{})
}

func (cmd *SimpleCommand) printTree(node *node, stack *Stack) (nd, nf int) {
	//	fmt.Printf("stack=[%s] len=%d\n", stack.String(), stack.Len())

	if stack.Len() > 0 {
		cmd.logger.Info("%s ", strings.Join(stack.Reverse(), " "))
	}

	cmd.logger.Info("%s\n", node.Name)

	if node.Nodes == nil {
		return 0, 1
	}

	top, _ := stack.Pop()
	base := stack.Len()

	nd = 0
	nf = 0
	zz := node.Len() - 1
	z0 := 0

	for _, v := range node.Nodes {

		switch top {
		case "":
			if z0 < zz {
				stack.Push(BRANCH)
			} else {
				stack.Push(LEAF)
			}
			break
		case BRANCH:
			if z0 < zz {
				stack.Push(TRUNK, BRANCH)
			} else {
				stack.Push(TRUNK, LEAF)
			}
			break
		case LEAF:
			if z0 < zz {
				stack.Push(AIR, BRANCH)
			} else {
				stack.Push(AIR, LEAF)
			}
			break
		default:
			cmd.logger.Info("\n---")
			cmd.logger.Info("invalid stack: ")
			cmd.logger.Info("stack=%v, top=%s\n", base, top)
			os.Exit(9)
		}

		xd, xf := cmd.printTree(v, stack)

		for stack.Len() > base {
			stack.Pop()
		}

		nd = nd + xd
		nf = nf + xf

		z0++
	}

	return nd + 1, nf
}

func (cmd *SimpleCommand) treeBlobsReport(root *node) {
	if cmd.outputMode == "json" {
		tmp := struct {
			StorageAccount string `json:"storageAccount"`
			Container      string `json:"container"`
			Root           *node  `json:"tree"`
		}{
			StorageAccount: cmd.config.Name,
			Container:      cmd.source.Container,
			Root:           root,
		}

		s, _ := json.Marshal(tmp)
		cmd.logger.Info("%s\n", s)
	} else {
		nd, nf := cmd.printRoot(root)

		cmd.logger.Debug("\n%d directories, %d files\n", nd, nf)
	}
}
