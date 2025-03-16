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

	// Pin a given file by its CID.
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

// createNode creates a single node with the given payload.
// It wraps the payload in an IPLD CBOR node and adds it to IPFS.
// It returns the node's CID.
func createNode(ctx context.Context, ipfs *rpc.HttpApi, payload map[string]interface{}) (cid.Cid, error) {
	node, err := ipldcbor.WrapObject(payload, mh.SHA2_256, -1)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to wrap node: %w", err)
	}
	nodeCid := node.Cid()
	if err := ipfs.Dag().Add(ctx, node); err != nil {
		return cid.Undef, fmt.Errorf("failed to dag put node: %w", err)
	}
	fmt.Printf("Created node CID: %s\n", nodeCid.String())
	return nodeCid, nil
}

// createThreeNodeDAG creates a three-node DAG consisting of two leaf nodes and
// one internal node that links the two leaf nodes.
func createThreeNodeDAG(ctx context.Context, ipfs *rpc.HttpApi) (cid.Cid, error) {
	// Create the first leaf node.
	leaf1Data := map[string]interface{}{
		"data": "leaf node 1",
	}
	leaf1Cid, err := createNode(ctx, ipfs, leaf1Data)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to create leaf 1: %w", err)
	}

	// Create the second leaf node.
	leaf2Data := map[string]interface{}{
		"data": "leaf node 2",
	}
	leaf2Cid, err := createNode(ctx, ipfs, leaf2Data)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to create leaf 2: %w", err)
	}

	// Create the internal node linking both leaf nodes.
	internalData := map[string]interface{}{
		"left":  map[string]interface{}{"/": leaf1Cid.String()},
		"right": map[string]interface{}{"/": leaf2Cid.String()},
	}
	internalCid, err := createNode(ctx, ipfs, internalData)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to create internal node: %w", err)
	}

	return internalCid, nil
}
