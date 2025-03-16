package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	ipldcbor "github.com/ipfs/go-ipld-cbor"
	"github.com/ipfs/kubo/client/rpc"
	caopts "github.com/ipfs/kubo/core/coreiface/options"
	mh "github.com/multiformats/go-multihash"
)

func main() {
	// Connect to the local IPFS node via the RPC API.
	ipfs, err := rpc.NewLocalApi()
	if err != nil {
		log.Fatalf("failed to create local IPFS API: %v", err)
	}

	ctx := context.Background()

	// Pin a given file by its CID.
	c, err := cid.Decode("bafkreidtuosuw37f5xmn65b3ksdiikajy7pwjjslzj2lxxz2vc4wdy3zku")
	if err != nil {
		log.Fatalf("failed to decode cid: %v", err)
	}
	cidPath := path.FromCid(c)
	if err := ipfs.Pin().Add(ctx, cidPath); err != nil {
		log.Fatalf("failed to pin file: %v", err)
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

	// Publish an IPNS record for the root CID.
	if err := publishIPNSRecord(ctx, ipfs, rootCid); err != nil {
		log.Fatalf("failed to publish IPNS record: %v", err)
	}
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

// publishIPNSRecord publishes an IPNS record for the given root CID.
// It uses the private key "three-node-dag-test". If the key does not exist, it is created.
func publishIPNSRecord(ctx context.Context, ipfs *rpc.HttpApi, root cid.Cid) error {
	// Check if the key "three-node-dag-test" exists.
	keys, err := ipfs.Key().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list keys: %w", err)
	}
	var keyExists bool
	for _, k := range keys {
		if k.Name() == "three-node-dag-test" {
			keyExists = true
			break
		}
	}
	// If the key does not exist, create it.
	if !keyExists {
		// Creating a new key with RSA algorithm and a default size.
		_, err := ipfs.Key().Generate(ctx, "three-node-dag-test")
		if err != nil {
			return fmt.Errorf("failed to generate key 'three-node-dag-test': %w", err)
		}
		fmt.Println("Created private key 'three-node-dag-test'")
	}

	// Publish the IPNS record for the root CID.
	rootPath := path.FromCid(root)
	// Set a lifetime of 24 hours; adjust as needed.
	published, err := ipfs.Name().Publish(ctx, rootPath,
		NamePublishKey("three-node-dag-test"),
		NamePublishLifetime(24*time.Hour))
	if err != nil {
		return fmt.Errorf("failed to publish IPNS record: %w", err)
	}
	fmt.Printf("Published IPNS record with name: %s\n", published)
	return nil
}

// NamePublishKey returns a NamePublishOption that sets the key field.
func NamePublishKey(key string) caopts.NamePublishOption {
	return func(s *caopts.NamePublishSettings) error {
		s.Key = key
		return nil
	}
}

// NamePublishLifetime returns a NamePublishOption that sets the valid lifetime.
func NamePublishLifetime(lifetime time.Duration) caopts.NamePublishOption {
	return func(s *caopts.NamePublishSettings) error {
		s.ValidTime = lifetime
		return nil
	}
}
