package main

// This example launches an IPFS-Lite peer and fetches a hello-world
// hash from the IPFS network. IPFS-Lite provides a lightweight way to
// interact with IPFS content without running a full IPFS node.

import (
	"context"
	"fmt"
	"io"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiformats/go-multiaddr"
)

func main() {
	// Create a context that can cancel ongoing operations
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an in-memory datastore for the peer (ephemeral storage)
	ds := ipfslite.NewInMemoryDatastore()
	
	// Generate RSA key pair for the peer's identity
	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		panic(err)
	}

	// Configure the network listener to listen on all interfaces, TCP port 4005
	listen, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/4005")

	// Set up libp2p host and DHT (Distributed Hash Table)
	// This creates the network stack for the IPFS-Lite node
	h, dht, err := ipfslite.SetupLibp2p(
		ctx,
		priv,          // Cryptographic identity
		nil,           // No existing peerstore
		[]multiaddr.Multiaddr{listen},  // Listen addresses
		ds,            // Datastore for network metadata
		ipfslite.Libp2pOptionsExtra..., // Default libp2p options
	)

	if err != nil {
		panic(err)
	}

	// Create the IPFS-Lite node instance
	lite, err := ipfslite.New(ctx, ds, nil, h, dht, nil)
	if err != nil {
		panic(err)
	}

	// Connect to IPFS bootstrap peers to join the network
	lite.Bootstrap(ipfslite.DefaultBootstrapPeers())

	// Decode a known CID (Content Identifier) for "Hello World"
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
