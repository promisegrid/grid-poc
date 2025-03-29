package ipldpatch

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
	"github.com/ipld/go-ipld-prime/storage/memstore"
	"github.com/ipld/go-ipld-prime/traversal/patch"
	"github.com/multiformats/go-multihash"
)

// RunDemo demonstrates patching an IPLD node and showing before/after states
func RunDemo() {
	// Setup linking system with in-memory storage
	store := &memstore.Store{}
	ls := cidlink.DefaultLinkSystem()
	ls.SetReadStorage(store)
	ls.SetWriteStorage(store)

	// Create initial node with basic fields
	originalNode := createSimpleNode(&ls, "Alice", 30)
	fmt.Println("=== Original Node Structure ===")
	printNode(originalNode, 0)

	// Apply JSON patch operations
	fmt.Println("\n=== Applying Patches ===")
	patchedNode := applyPatches(ls, originalNode)

	fmt.Println("\n=== Patched Node Structure ===")
	printNode(patchedNode, 0)
}

// createSimpleNode constructs a basic map node with name and age fields
func createSimpleNode(ls *linking.LinkSystem, name string, age int64) ipld.Node {
	nodeBuilder := basicnode.Prototype.Map.NewBuilder()
	ma, _ := nodeBuilder.BeginMap(2)
	ma.AssembleKey().AssignString("name")
	ma.AssembleValue().AssignString(name)
	ma.AssembleKey().AssignString("age")
	ma.AssembleValue().AssignInt(age)
	ma.Finish()
	node := nodeBuilder.Build()

	// Store node using CIDv1 with SHA2-256 hash
	_, err := ls.Store(
		linking.LinkContext{},
		cidlink.LinkPrototype{
			Prefix: cid.Prefix{
				Version:  1,
				Codec:    cid.DagJSON,
				MhType:   multihash.SHA2_256,
				MhLength: 32,
			},
		},
		node,
	)
	if err != nil {
		panic("storage failed: " + err.Error())
	}
	return node
}

// printNode displays node content with proper indentation
func printNode(n ipld.Node, depth int) {
	indent := fmt.Sprintf("%*s", depth*2, "")
	
	switch n.Kind() {
	case ipld.Kind_Link:
		link, _ := n.AsLink()
		fmt.Printf("%s↳ %s\n", indent, link)
	default:
		var buf bytes.Buffer
		dagjson.Encode(n, &buf)
		fmt.Printf("%sNode (%s):\n%s%s\n", indent, n.Kind(), indent, buf.String())
	}
}

// applyPatches demonstrates modification of IPLD nodes using JSON Patch operations
func applyPatches(ls linking.LinkSystem, original ipld.Node) ipld.Node {
	// Define patch operations
	ops := []patch.Operation{
		{
			Op:    patch.Op_Replace,
			Path:  datamodel.ParsePath("age"),
			Value: fromJSONString("31"),
		},
		{
			Op:    patch.Op_Add,
			Path:  datamodel.ParsePath("city"),
			Value: fromJSONString(`"New York"`),
		},
	}

	// Apply patches to create new node
	patchedNode, err := patch.Eval(original, ops)
	if err != nil {
		panic(err)
	}

	// Store patched node
	_, err = ls.Store(
		linking.LinkContext{},
		cidlink.LinkPrototype{
			Prefix: cid.Prefix{
				Version:  1,
				Codec:    cid.DagJSON,
				MhType:   multihash.SHA2_256,
				MhLength: 32,
			},
		},
		patchedNode,
	)
	if err != nil {
		panic(err)
	}

	return patchedNode
}

// fromJSONString converts JSON string to IPLD Node
func fromJSONString(s string) ipld.Node {
	nb := basicnode.Prototype.Any.NewBuilder()
	if err := dagjson.Decode(nb, bytes.NewBufferString(s)); err != nil {
		panic(err)
	}
	return nb.Build()
}
