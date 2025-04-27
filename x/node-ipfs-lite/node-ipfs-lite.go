package main

// This example launches a persistent IPFS-Lite peer using FlatFS block storage
// and fetches a hello-world hash from the IPFS network.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	flatfs "github.com/ipfs/go-ds-flatfs"
	datastore "github.com/ipfs/go-datastore"
	mount "github.com/ipfs/go-datastore/mount"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiformats/go-multiaddr"
)

func main() {
	// Create a context that can cancel ongoing operations
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configure persistent storage location and FlatFS sharding
	repoPath := "/tmp/ipfs-lite"
	// Use the NextToLast sharding function for FlatFS
	shardFunc := flatfs.NextToLast(2)

	// Create or open FlatFS datastore for block storage
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		panic(err)
	}
	if err := flatfs.Create(repoPath, shardFunc); err != nil {
		// Ignore error if datastore already exists
		if err.Error() != "datastore already exists" &&
			!errors.Is(err, os.ErrExist) {
			panic(err)
		}
	}
	ds, err := flatfs.Open(repoPath, false)
	if err != nil {
		panic(err)
	}
	defer ds.Close()

	// Wrap the FlatFS datastore with a mount to handle block keys without
	// unsupported prefixes.  IPFS-Lite uses keys with a "/blocks/" prefix,
	// but FlatFS does not allow keys containing the "/" separator.
	// The mount datastore will only store the key in the FlatFS datastore if
	// the key begins with "blocks" (without a leading slash).
	blocksDs := mount.New([]mount.Mount{{
		Prefix:    datastore.NewKey("blocks"),
		Datastore: ds,
	}})

	// Generate RSA key pair for persistent peer identity
	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		panic(err)
	}

	// Configure network listener for long-running service
	listen, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/4005")

	// Set up libp2p host and DHT with persistent configuration
	h, dht, err := ipfslite.SetupLibp2p(
		ctx,
		priv,                          // Persistent cryptographic identity
		nil,                           // No existing peerstore
		[]multiaddr.Multiaddr{listen}, // Persistent listener config
		ds,                            // Shared datastore for network metadata
		ipfslite.Libp2pOptionsExtra..., // Default libp2p options
	)
	if err != nil {
		panic(err)
	}

	// Create IPFS-Lite node with persistent block store using the mounted
	// datastore.
	lite, err := ipfslite.New(ctx, blocksDs, nil, h, dht, nil)
	if err != nil {
		panic(err)
	}

	// Connect to IPFS network using bootstrap peers
	lite.Bootstrap(ipfslite.DefaultBootstrapPeers())

	// Fetch and display test content
	c, _ := cid.Decode("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8u")

	// Retrieve the file from the IPFS network using the Lite node
	rsc, err := lite.GetFile(ctx, c)
	if err != nil {
		panic(err)
	}
	defer rsc.Close()

	// Read the entire content of the fetched file
	content, err := io.ReadAll(rsc)
	if err != nil {
		panic(err)
	}

	// Print the content (should be "Hello World")
	fmt.Println(string(content))
}
