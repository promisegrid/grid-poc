package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/linking"
	"github.com/ipld/go-ipld-prime/linking/cidlink"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/storage/memstore"
	"github.com/ipld/go-ipld-prime/traversal/patch"
)

type UserData struct {
	Profile  ipld.Link
	Settings ipld.Link
}

type Profile struct {
	Name string
	Age  int64
}

type Settings struct {
	Active bool
}

func main() {
	// Setup linking system with in-memory store
	store := &memstore.Store{}
	linkSystem := cidlink.DefaultLinkSystem()
	linkSystem.SetReadStorage(store)
	linkSystem.SetWriteStorage(store)

	// Create profile block
	profile := &Profile{Name: "Alice", Age: 30}
	profileNode := bindnode.Wrap(profile, nil)
	profileLink, _ := linkSystem.Store(
		ipld.LinkContext{},
		cidlink.LinkPrototype{Prefix: cid.Prefix{
			Version:  1,
			Codec:    0x71,
			MhType:   0x13,
			MhLength: 32,
		}},
		profileNode,
	)

	// Create settings block
	settings := &Settings{Active: true}
	settingsNode := bindnode.Wrap(settings, nil)
	settingsLink, _ := linkSystem.Store(
		ipld.LinkContext{},
		cidlink.LinkPrototype{Prefix: cid.Prefix{
			Version:  1,
			Codec:    0x71,
			MhType:   0x13,
			MhLength: 32,
		}},
		settingsNode,
	)

	// Create root block
	root := &UserData{
		Profile:  profileLink,
		Settings: settingsLink,
	}
	rootNode := bindnode.Wrap(root, nil)
	rootLink, _ := linkSystem.Store(
		ipld.LinkContext{},
		cidlink.LinkPrototype{Prefix: cid.Prefix{
			Version:  1,
			Codec:    0x71,
			MhType:   0x13,
			MhLength: 32,
		}},
		rootNode,
	)

	// 1. Demonstrate cross-block navigation
	fmt.Println("=== Initial Navigation ===")
	navigate(linkSystem, rootLink, "Profile/Age")     // Should show 30
	navigate(linkSystem, rootLink, "Settings/Active") // Should show true

	// 2. Explore data structure
	fmt.Println("\n=== Data Exploration ===")
	exploreNode(linkSystem, rootLink, 0)

	// 3. Apply patches
	fmt.Println("\n=== Applying Patches ===")
	patchedRoot := applyPatches(linkSystem, rootNode)

	// 4. Demonstrate post-patch navigation
	fmt.Println("\n=== Post-Patch Navigation ===")
	navigate(linkSystem, rootLink, "Profile/Age")      // Should show 31
	navigate(linkSystem, rootLink, "Settings/Timeout") // New field

	// 5. Explore modified structure
	fmt.Println("\n=== Patched Data Exploration ===")
	exploreNode(linkSystem, rootLink, 0)
}

func navigate(ls linking.LinkSystem, startLink ipld.Link, pathStr string) {
	fmt.Printf("Navigating: %s\n", pathStr)
	path := ipld.ParsePath(pathStr)

	node, _ := ls.Load(ipld.LinkContext{}, startLink, nil)

	for !path.IsEmpty() {
		seg, remaining := path.PopSegment()
		path = remaining

		if node.Kind() == ipld.Kind_Link {
			link, _ := node.AsLink()
			node, _ = ls.Load(ipld.LinkContext{}, link, nil)
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
	dagjson.Encode(node, os.Stdout)
	fmt.Println()
}

func exploreNode(ls linking.LinkSystem, link ipld.Link, indent int) {
	node, _ := ls.Load(ipld.LinkContext{}, link, nil)
	explore(node, indent)
}

func explore(n ipld.Node, indent int) {
	switch n.Kind() {
	case ipld.Kind_Map:
		fmt.Printf("%sMap:\n", getIndent(indent))
		iter := n.MapIterator()
		for !iter.Done() {
			k, v, _ := iter.Next()
			key, _ := k.AsString()
			fmt.Printf("%s- Key: %s\n", getIndent(indent+1), key)
			explore(v, indent+2)
		}

	case ipld.Kind_List:
		fmt.Printf("%sList:\n", getIndent(indent))
		for i := 0; ; i++ {
			v, err := n.LookupByIndex(i)
			if err != nil {
				break
			}
			fmt.Printf("%s- Index: %d\n", getIndent(indent+1), i)
			explore(v, indent+2)
		}

	case ipld.Kind_Link:
		fmt.Printf("%sLink: %s\n", getIndent(indent), n)

	default:
		val, _ := n.AsString()
		fmt.Printf("%sValue: %s\n", getIndent(indent), val)
	}
}

func getIndent(n int) string {
	return fmt.Sprintf("%*s", n, "")
}

func applyPatches(ls linking.LinkSystem, rootNode ipld.Node) ipld.Node {
	// Define patch operations
	ops := []patch.Operation{
		{
			Op:   patch.Op_Replace,
			Path: ipld.ParsePath("Profile/Age"),
			Value: func() ipld.Node {
				n, _ := ipld.FromJSONString(`31`)
				return n
			}(),
		},
		{
			Op:   patch.Op_Add,
			Path: ipld.ParsePath("Settings/Timeout"),
			Value: func() ipld.Node {
				n, _ := ipld.FromJSONString(`30`)
				return n
			}(),
		},
	}

	// Apply patches
	patchedNode, err := patch.Eval(rootNode, ops)
	if err != nil {
		panic(err)
	}

	// Store updated root
	buf := bytes.NewBuffer(nil)
	dagjson.Encode(patchedNode, buf)
	newNode, _ := ipld.FromJSONBytes(buf.Bytes())

	_, err = ls.Store(
		ipld.LinkContext{},
		cidlink.LinkPrototype{Prefix: cid.Prefix{
			Version:  1,
			Codec:    0x71,
			MhType:   0x13,
			MhLength: 32,
		}},
		newNode,
	)
	if err != nil {
		panic(err)
	}

	return newNode
}
