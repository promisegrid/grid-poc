package ipldexplore

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/multiformats/go-multihash"
)

// RunDemo creates and explores a simple hierarchical IPLD structure
func RunDemo() {
	// Initialize in-memory storage and link system
	store := NewMemStore()
	ls := cidlink.DefaultLinkSystem()
	ls.StorageReadOpener = store.OpenRead
	ls.StorageWriteOpener = store.OpenWrite

	// Create leaf nodes
	leaf1 := createNode(&ls, "Leaf 1", nil)
	leaf2 := createNode(&ls, "Leaf 2", nil)

	// Create parent node with children links
	parent := createNode(&ls, "Parent Node", []ipld.Link{leaf1, leaf2})

	// Begin recursive exploration from root
	fmt.Println("IPLD Node Structure:")
	exploreNode(ls, parent, 0)
}

// MemStore implements simple in-memory block storage
type MemStore struct {
	blocks map[cid.Cid][]byte
}

func NewMemStore() *MemStore {
	return &MemStore{blocks: make(map[cid.Cid][]byte)}
}

func (m *MemStore) OpenRead(_ linking.LinkContext, l ipld.Link) (io.Reader, error) {
	c, err := cid.Parse([]byte(l.Binary()))
	if err != nil {
		return nil, err
	}
	data, exists := m.blocks[c]
	if !exists {
		return nil, fmt.Errorf("block not found")
	}
	return bytes.NewReader(data), nil
}

func (m *MemStore) OpenWrite(_ linking.LinkContext) (io.Writer, linking.BlockWriteCommitter, error) {
	buf := bytes.NewBuffer(nil)
	return buf, func(lnk ipld.Link) error {
		c, err := cid.Parse([]byte(lnk.Binary()))
		if err != nil {
			return err
		}
		m.blocks[c] = buf.Bytes()
		return nil
	}, nil
}

// createNode constructs and stores a node with name and optional children links
func createNode(ls *linking.LinkSystem, name string, children []ipld.Link) ipld.Link {
	// Create new map node with name and optional children
	nodeBuilder := basicnode.Prototype.Map.NewBuilder()
	ma, _ := nodeBuilder.BeginMap(2)
	
	// Set name field
	ma.AssembleKey().AssignString("Name")
	ma.AssembleValue().AssignString(name)

	// Add children links if provided
	if len(children) > 0 {
		ma.AssembleKey().AssignString("Children")
		listBuilder := basicnode.Prototype.List.NewBuilder()
		listAssembler, _ := listBuilder.BeginList(int64(len(children)))
		for _, child := range children {
			listAssembler.AssembleValue().AssignLink(child)
		}
		listAssembler.Finish()
		ma.AssembleValue().AssignNode(listBuilder.Build())
	}

	ma.Finish()
	node := nodeBuilder.Build()

	// Store node using CIDv1 with SHA2-256 hash
	lnk, err := ls.Store(
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
	return lnk
}

// exploreNode recursively traverses and prints IPLD nodes
func exploreNode(ls linking.LinkSystem, link ipld.Link, depth int) {
	// Load node from link system
	node, err := ls.Load(linking.LinkContext{}, link, basicnode.Prototype.Any)
	if err != nil {
		panic("load failed: " + err.Error())
	}

	printNode(node, depth)

	// Process child links if present
	if children, err := node.LookupByString("Children"); err == nil {
		iter := children.ListIterator()
		for !iter.Done() {
			_, item, _ := iter.Next()
			if item.Kind() == ipld.Kind_Link {
				if childLink, err := item.AsLink(); err == nil {
					exploreNode(ls, childLink, depth+1)
				}
			}
		}
	}
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
