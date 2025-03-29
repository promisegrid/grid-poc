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

// UserData represents a root node linking to Profile and Settings.
type UserData struct {
	Profile  ipld.Link
	Settings ipld.Link
}

// Profile holds a user's profile information.
type Profile struct {
	Name string
	Age  int64
}

// Settings holds configuration settings.
type Settings struct {
	Active bool
}

// RunDemo executes a demonstration of basic IPLD path traversal.
func RunDemo() {
	// Setup linking system with in-memory store.
	store := &memstore.Store{}
	ls := cidlink.DefaultLinkSystem()
	ls.SetReadStorage(store)
	ls.SetWriteStorage(store)

	// Create profile block.
	profile := &Profile{Name: "Alice", Age: 30}
	profileNode := bindnode.Wrap(profile, nil)
	profileLink, err := ls.Store(
		ipld.LinkContext{},
		cidlink.LinkPrototype{
			Prefix: cid.Prefix{
				Version:  1,
				Codec:    0x0129, // DAG-JSON multicodec
				MhType:   0x13,
				MhLength: 32,
			},
		},
		profileNode,
	)
	if err != nil {
		panic(err)
	}

	// Create settings block.
	settings := &Settings{Active: true}
	settingsNode := bindnode.Wrap(settings, nil)
	settingsLink, err := ls.Store(
		ipld.LinkContext{},
		cidlink.LinkPrototype{
			Prefix: cid.Prefix{
				Version:  1,
				Codec:    0x0129, // DAG-JSON multicodec
				MhType:   0x13,
				MhLength: 32,
			},
		},
		settingsNode,
	)
	if err != nil {
		panic(err)
	}

	// Create root block.
	root := &UserData{
		Profile:  profileLink,
		Settings: settingsLink,
	}
	rootNode := bindnode.Wrap(root, nil)
	rootLink, err := ls.Store(
		ipld.LinkContext{},
		cidlink.LinkPrototype{
			Prefix: cid.Prefix{
				Version:  1,
				Codec:    0x0129, // DAG-JSON multicodec
				MhType:   0x13,
				MhLength: 32,
			},
		},
		rootNode,
	)
	if err != nil {
		panic(err)
	}

	// Demonstrate basic path navigation
	fmt.Println("=== Basic Path Navigation ===")
	navigate(ls, rootLink, "Profile/Age")     // Shows 30
	navigate(ls, rootLink, "Settings/Active") // Shows true
}

func navigate(ls linking.LinkSystem, startLink ipld.Link, pathStr string) {
	fmt.Printf("Navigating: %s\n", pathStr)
	path := datamodel.ParsePath(pathStr)

	// Load initial node with prototype
	node, err := ls.Load(ipld.LinkContext{}, startLink, basicnode.Prototype.Any)
	if err != nil {
		panic(err)
	}

	for path.Len() > 0 {
		seg, remaining := path.Shift()
		path = remaining

		// Handle links before processing segments
		if node.Kind() == ipld.Kind_Link {
			link, err := node.AsLink()
			if err != nil {
				panic(err)
			}
			node, err = ls.Load(ipld.LinkContext{}, link, basicnode.Prototype.Any)
			if err != nil {
				panic(err)
			}
		}

		switch node.Kind() {
		case ipld.Kind_Map:
			node, err = node.LookupByString(seg.String())
		case ipld.Kind_List:
			idx, err := seg.Index()
			if err != nil {
				panic(err)
			}
			node, err = node.LookupByIndex(idx)
		}
		if err != nil {
			panic(err)
		}
	}

	fmt.Print("Found value: ")
	_ = dagjson.Encode(node, &noCloseWriter{})
	fmt.Println()
}

type noCloseWriter struct{}

func (w *noCloseWriter) Write(p []byte) (int, error) {
	fmt.Print(string(p))
	return len(p), nil
}
