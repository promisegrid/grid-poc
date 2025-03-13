package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	ipldcbor "github.com/ipfs/go-ipld-cbor"
	"github.com/ipfs/kubo/client/rpc"
	mh "github.com/multiformats/go-multihash"
)

func main() {
	// Connect to the local IPFS node via the RPC API.
	ipfs, err := rpc.NewLocalApi()
	if err != nil {
		log.Fatalf("failed to create local IPFS API: %v", err)
	}

	// Pin a given file by its CID
	ctx := context.Background()
	c, err := cid.Decode("bafkreidtuosuw37f5xmn65b3ksdiikajy7pwjjslzj2lxxz2vc4wdy3zku")
	if err != nil {
		fmt.Println(err)
		return
	}
	cidPath := path.FromCid(c)
	err = ipfs.Pin().Add(ctx, cidPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create a three-node DAG.
	rootCid, err := createThreeNodeDAG(ctx, ipfs)
	if err != nil {
		log.Fatalf("failed to create three node DAG: %v", err)
	}

	// Pin the root node for persistence.
	rootCidPath := path.FromCid(rootCid)
	if err := ipfs.Pin().Add(ctx, rootCidPath); err != nil {
		log.Fatalf("failed to pin the root node: %v", err)
	}

	// Publish the root CID to the pubsub topic "t7a".
	message := []byte(rootCid.String())
	if err := ipfs.PubSub().Publish(ctx, "t7a", message); err != nil {
		log.Fatalf("failed to publish on pubsub: %v", err)
	}
	fmt.Println("Published root CID to pubsub topic 't7a'")
}

// createThreeNodeDAG creates a three-node DAG consisting of two leaf nodes and
// one internal node. The internal node links the two leaf nodes.
func createThreeNodeDAG(ctx context.Context, ipfs *rpc.HttpApi) (cid.Cid, error) {
	// Create the first leaf node.
	leaf1Data := map[string]interface{}{
		"data": "leaf node 1",
	}
	// convert leaf1Data to format.Node
	leaf1, err := ipldcbor.WrapObject(leaf1Data, mh.SHA2_256, -1)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to wrap leaf 1: %w", err)
	}
	leaf1Cid := leaf1.Cid()
	// add leaf1 to IPFS
	err = ipfs.Dag().Add(ctx, leaf1)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to dag put leaf 1: %w", err)
	}
	fmt.Printf("Leaf 1 CID: %s\n", leaf1Cid.String())

	// Create the second leaf node.
	leaf2Data := map[string]interface{}{
		"data": "leaf node 2",
	}
	// convert leaf2Data to format.Node
	leaf2, err := ipldcbor.WrapObject(leaf2Data, mh.SHA2_256, -1)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to wrap leaf 2: %w", err)
	}
	leaf2Cid := leaf2.Cid()
	// add leaf2 to IPFS
	err = ipfs.Dag().Add(ctx, leaf2)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to dag put leaf 2: %w", err)
	}
	fmt.Printf("Leaf 2 CID: %s\n", leaf2Cid.String())

	// Create the internal node linking both leaf nodes.
	internalData := map[string]interface{}{
		"left":  map[string]interface{}{"/": leaf1Cid.String()},
		"right": map[string]interface{}{"/": leaf2Cid.String()},
	}
	// convert internalData to format.Node
	internal, err := ipldcbor.WrapObject(internalData, mh.SHA2_256, -1)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to wrap internal node: %w", err)
	}
	rootCid := internal.Cid()
	// add internal to IPFS
	err = ipfs.Dag().Add(ctx, internal)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to dag put internal node: %w", err)
	}
	fmt.Printf("Internal (root) CID: %s\n", rootCid.String())

	return rootCid, nil
}
