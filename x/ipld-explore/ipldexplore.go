package ipldexplore

import (
	"bytes"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/storage/memstore"
)

// UserData represents a root node containing links to other data sections
type UserData struct {
	Profile  ipld.Link
	Settings ipld.Link
}

// Profile contains basic user information
type Profile struct {
	Name string
	Age  int64
}

// Settings holds application configuration
type Settings struct {
	Active bool
}

// RunDemo demonstrates IPLD path navigation and node exploration
func RunDemo() {
	// Initialize in-memory storage and linking system
	store := &memstore.Store{}
	ls := cidlink.DefaultLinkSystem()
	ls.SetReadStorage(store)
	ls.SetWriteStorage(store)

	// Create and store profile data
	profile := &Profile{Name: "Alice", Age: 30}
	profileLink := storeNode(&ls, profile)

	// Create and store settings data
	settings := &Settings{Active: true}
	settingsLink := storeNode(&ls, settings)

	// Create and store root node
	root := &UserData{Profile: profileLink, Settings: settingsLink}
	rootLink := storeNode(&ls, root)

	// Demonstrate path navigation
	fmt.Println("=== Basic Navigation ===")
	navigate(ls, rootLink, "Profile/Age")     // Expected: 30
	navigate(ls, rootLink, "Settings/Active") // Expected: true

	// Show full structure exploration
	fmt.Println("\n=== Structure Exploration ===")
	exploreStructure(ls, rootLink, 0)
}

// storeNode generic function to store any Go value as an IPLD node
func storeNode(ls *linking.LinkSystem, data interface{}) ipld.Link {
	node := bindnode.Wrap(data, nil)
	linkProto := cidlink.LinkPrototype{Prefix: cid.Prefix{
		Version:  1,
		Codec:    0x0129, // dag-json codec
		MhType:   0x13,   // sha2-256
		MhLength: 32,
	}}

	link, err := ls.Store(ipld.LinkContext{}, linkProto, node)
	if err != nil {
		panic(fmt.Sprintf("failed to store node: %v", err))
	}
	return link
}

// navigate demonstrates path traversal through linked nodes
func navigate(ls linking.LinkSystem, startLink ipld.Link, pathStr string) {
	fmt.Printf("Path: %s\n", pathStr)
	path := datamodel.ParsePath(pathStr)

	currentNode, err := ls.Load(ipld.LinkContext{}, startLink, basicnode.Prototype.Any)
	if err != nil {
		panic(fmt.Sprintf("loading failed: %v", err))
	}

	for !path.Empty() {
		seg, rest := path.Shift()
		path = rest

		// Resolve links automatically
		if currentNode.Kind() == ipld.Kind_Link {
			link, _ := currentNode.AsLink()
			currentNode, _ = ls.Load(ipld.LinkContext{}, link, basicnode.Prototype.Any)
		}

		var err error
		switch currentNode.Kind() {
		case ipld.Kind_Map:
			currentNode, err = currentNode.LookupByString(seg.String())
		case ipld.Kind_List:
			idx, _ := seg.Index()
			currentNode, err = currentNode.LookupByIndex(idx)
		default:
			panic("invalid path segment")
		}

		if err != nil {
			panic(fmt.Sprintf("traversal failed: %v", err))
		}
	}

	fmt.Print("Result: ")
	_ = dagjson.Encode(currentNode, &consoleWriter{})
	fmt.Println()
}

// exploreStructure recursively explores and prints node structure
func exploreStructure(ls linking.LinkSystem, link ipld.Link, indent int) {
	node, err := ls.Load(ipld.LinkContext{}, link, basicnode.Prototype.Any)
	if err != nil {
		panic(fmt.Sprintf("loading failed: %v", err))
	}

	switch node.Kind() {
	case ipld.Kind_Map:
		fmt.Printf("%sMap:\n", indentString(indent))
		iter := node.MapIterator()
		for !iter.Done() {
			key, value, _ := iter.Next()
			keyStr, _ := key.AsString()
			fmt.Printf("%s- Key: %s\n", indentString(indent+1), keyStr)
			exploreNodeContents(value, indent+2)
		}

	case ipld.Kind_List:
		fmt.Printf("%sList:\n", indentString(indent))
		for i := 0; ; i++ {
			value, err := node.LookupByIndex(int64(i))
			if err != nil {
				break
			}
			fmt.Printf("%s- Index: %d\n", indentString(indent+1), i)
			exploreNodeContents(value, indent+2)
		}

	case ipld.Kind_Link:
		link, _ := node.AsLink()
		fmt.Printf("%sLink: %s\n", indentString(indent), link)
		exploreStructure(ls, link, indent+1)

	default:
		printNodeValue(node, indent)
	}
}

// exploreNodeContents handles recursive exploration of node contents
func exploreNodeContents(n ipld.Node, indent int) {
	if n.Kind() == ipld.Kind_Link {
		link, _ := n.AsLink()
		exploreStructure(ls, link, indent)
	} else {
		printNodeValue(n, indent)
	}
}

// printNodeValue displays scalar values and links
func printNodeValue(n ipld.Node, indent int) {
	var buf bytes.Buffer
	_ = dagjson.Encode(n, &buf)
	fmt.Printf("%sValue: %s\n", indentString(indent), buf.String())
}

// indentString creates whitespace for visual hierarchy
func indentString(n int) string {
	return fmt.Sprintf("%*s", n, "")
}

// consoleWriter implements io.Writer for direct console output
type consoleWriter struct{}

func (w *consoleWriter) Write(p []byte) (int, error) {
	fmt.Print(string(p))
	return len(p), nil
}
