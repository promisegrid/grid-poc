package ipldpath

import (
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

// UserData represents a root node containing links to other data blocks
type UserData struct {
	Profile  ipld.Link
	Settings ipld.Link
}

// Profile demonstrates a linked data structure
type Profile struct {
	Name string
	Age  int64
}

// Settings shows cross-block content addressing
type Settings struct {
	Active bool
}

// RunDemo sets up and demonstrates cross-block path traversal
func RunDemo() {
	// Initialize in-memory storage and linking system
	store := &memstore.Store{}
	ls := cidlink.DefaultLinkSystem()
	ls.SetReadStorage(store)
	ls.SetWriteStorage(store)

	// Create and store profile block
	profile := &Profile{Name: "Alice", Age: 30}
	profileLink := storeNode(ls, profile)

	// Create and store settings block
	settings := &Settings{Active: true}
	settingsLink := storeNode(ls, settings)

	// Create root block linking to both
	root := &UserData{
		Profile:  profileLink,
		Settings: settingsLink,
	}
	rootLink := storeNode(ls, root)

	// Demonstrate cross-block navigation
	fmt.Println("=== Cross-Block Path Traversal ===")
	navigate(ls, rootLink, "Profile/Age")     // From root → Profile → Age
	navigate(ls, rootLink, "Settings/Active") // From root → Settings → Active
}

// storeNode generic storage helper for IPLD nodes
func storeNode(ls linking.LinkSystem, data interface{}) ipld.Link {
	node := bindnode.Wrap(data, nil)
	link, err := ls.Store(
		ipld.LinkContext{},
		cidlink.LinkPrototype{
			Prefix: cid.Prefix{
				Version:  1,
				Codec:    0x0129, // dag-json multicodec
				MhType:   0x13,    // sha3-384
				MhLength: 32,
			},
		},
		node,
	)
	if err != nil {
		panic(err)
	}
	return link
}

// navigate demonstrates path traversal across linked blocks
func navigate(ls linking.LinkSystem, startLink ipld.Link, pathStr string) {
	fmt.Printf("Navigating: %s\n", pathStr)
	path := datamodel.ParsePath(pathStr)

	node, err := ls.Load(ipld.LinkContext{}, startLink, basicnode.Prototype.Any)
	if err != nil {
		panic(err)
	}

	for path.Len() > 0 {
		seg, remaining := path.Shift()
		path = remaining

		// Resolve links automatically during traversal
		if node.Kind() == ipld.Kind_Link {
			link, _ := node.AsLink()
			node, _ = ls.Load(ipld.LinkContext{}, link, basicnode.Prototype.Any)
		}

		switch node.Kind() {
		case ipld.Kind_Map:
			node, _ = node.LookupByString(seg.String())
		case ipld.Kind_List:
			idx, _ := seg.Index()
			node, _ = node.LookupByIndex(idx)
		}
	}

	fmt.Print("Found value: ")
	_ = dagjson.Encode(node, &consoleWriter{})
	fmt.Println()
}

// consoleWriter implements io.Writer for direct console output
type consoleWriter struct{}

func (w *consoleWriter) Write(p []byte) (int, error) {
	fmt.Print(string(p))
	return len(p), nil
}
