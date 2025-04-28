package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ipfs/go-cid"
	flatfs "github.com/ipfs/go-ds-flatfs"

	// ds "github.com/ipfs/go-datastore"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	ds_sync "github.com/ipfs/go-datastore/sync"
	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

func main() {
	ctx := context.Background()

	// 1. Create or open a FlatFS datastore directory
	flatfsDir := "/tmp/ipfs-lite-flatfs"
	if err := os.MkdirAll(flatfsDir, 0755); err != nil {
		panic(err)
	}

	// 2. Choose a sharding function (use 2-level next-to-last as in IPFS)
	shard, err := flatfs.ParseShardFunc("/repo/flatfs/shard/v1/next-to-last/2")
	if err != nil {
		panic(err)
	}

	// 3. Open the FlatFS datastore using CreateOrOpen to handle shard configuration
	fs, err := flatfs.CreateOrOpen(flatfsDir, shard, false)
	if err != nil {
		panic(err)
	}
	// Wrap with sync to ensure thread safety
	datastore := ds_sync.MutexWrap(fs)

	// 4. Set up libp2p host and DHT (required by ipfs-lite)
	host, err := libp2p.New()
	if err != nil {
		panic(err)
	}
	kaddht, err := dht.New(ctx, host)
	if err != nil {
		panic(err)
	}

	// 5. Create the ipfs-lite peer with FlatFS as the datastore
	peer, err := ipfslite.New(ctx, datastore, nil, host, kaddht, nil)
	if err != nil {
		panic(err)
	}

	// bootstrap the peers list
	peer.Bootstrap(ipfslite.DefaultBootstrapPeers())

	fmt.Println("IPFS-Lite peer with FlatFS is ready:")
	// spew.Dump(peer)

	// get a file from IPFS
	c, err := cid.Decode("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8u")
	if err != nil {
		panic(err)
	}
	// Convert to CIDv1 for base32 encoding
	c = cid.NewCidV1(c.Type(), c.Hash())

	rd, err := peer.GetFile(ctx, c)
	if err != nil {
		panic(err)
	}
	defer rd.Close()
	// Read the data from the reader
	content, err := io.ReadAll(rd)
	if err != nil {
		panic(err)
	}
	fmt.Println("File content:", string(content))
}
