package ipldpath

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
	"github.com/ipld/go-ipld-prime/traversal/patch"
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

// RunDemo executes a demonstration of multi-block navigation,
// data exploration, and patch application.
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

	// 1. Demonstrate cross-block navigation.
	fmt.Println("=== Initial Navigation ===")
	navigate(ls, rootLink, "Profile/Age")     // Should show 30
	navigate(ls, rootLink, "Settings/Active") // Should show true

	// 2. Explore data structure.
	fmt.Println("\n=== Data Exploration ===")
	exploreNode(ls, rootLink, 0)

	// 3. Apply patches.
	fmt.Println("\n=== Applying Patches ===")
	_ = applyPatches(ls, rootNode)

	// 4. Demonstrate post-patch navigation.
	fmt.Println("\n=== Post-Patch Navigation ===")
	navigate(ls, rootLink, "Profile/Age")      // Should show 31
	navigate(ls, rootLink, "Settings/Timeout") // New field

	// 5. Explore modified structure.
	fmt.Println("\n=== Patched Data Exploration ===")
	exploreNode(ls, rootLink, 0)
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

func exploreNode(ls linking.LinkSystem, link ipld.Link, indent int) {
	node, err := ls.Load(ipld.LinkContext{}, link, basicnode.Prototype.Any)
	if err != nil {
		panic(err)
	}
	explore(node, indent)
}

func explore(n ipld.Node, indent int) {
	switch n.Kind() {
	case ipld.Kind_Map:
		fmt.Printf("%sMap:\n", getIndent(indent))
		iter := n.MapIterator()
		for !iter.Done() {
			k, v, err := iter.Next()
			if err != nil {
				panic(err)
			}
			key, err := k.AsString()
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s- Key: %s\n", getIndent(indent+1), key)
			explore(v, indent+2)
		}
	case ipld.Kind_List:
		fmt.Printf("%sList:\n", getIndent(indent))
		for i := 0; ; i++ {
			v, err := n.LookupByIndex(int64(i))
			if err != nil {
				break
			}
			fmt.Printf("%s- Index: %d\n", getIndent(indent+1), i)
			explore(v, indent+2)
		}
	case ipld.Kind_Link:
		fmt.Printf("%sLink: %v\n", getIndent(indent), n)
	default:
		var buf bytes.Buffer
		if err := dagjson.Encode(n, &buf); err != nil {
			panic(err)
		}
		fmt.Printf("%sValue: %s\n", getIndent(indent), buf.String())
	}
}

func getIndent(n int) string {
	return fmt.Sprintf("%*s", n, "")
}

func applyPatches(ls linking.LinkSystem, rootNode ipld.Node) ipld.Node {
	ops := []patch.Operation{
		{
			Op:    patch.Op_Replace,
			Path:  datamodel.ParsePath("Profile/Age"),
			Value: fromJSONString("31"),
		},
		{
			Op:    patch.Op_Add,
			Path:  datamodel.ParsePath("Settings/Timeout"),
			Value: fromJSONString("30"),
		},
	}

	patchedNode, err := patch.Eval(rootNode, ops)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := dagjson.Encode(patchedNode, &buf); err != nil {
		panic(err)
	}
	
	nb := basicnode.Prototype.Any.NewBuilder()
	if err := dagjson.Decode(nb, &buf); err != nil {
		panic(err)
	}
	newNode := nb.Build()

	_, err = ls.Store(
		ipld.LinkContext{},
		cidlink.LinkPrototype{
			Prefix: cid.Prefix{
				Version:  1,
				Codec:    0x0129,
				MhType:   0x13,
				MhLength: 32,
			},
		},
		newNode,
	)
	if err != nil {
		panic(err)
	}

	return newNode
}

func fromJSONString(s string) ipld.Node {
	nb := basicnode.Prototype.Any.NewBuilder()
	if err := dagjson.Decode(nb, bytes.NewBufferString(s)); err != nil {
		panic(err)
	}
	return nb.Build()
}
