package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"time"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	routing "github.com/libp2p/go-libp2p/core/routing"
	bsnet "github.com/ipfs/boxo/bitswap/network"
	"github.com/ipfs/boxo/bitswap"
)

func main() {
	if len(os.Args) < 2 {
		runParent()
	} else {
		runChild(os.Args[1])
	}
}

func runParent() {
	ctx := context.Background()

	// Create parent node components
	parentHost, parentDHT, parentBS := setupNode(ctx)
	defer parentHost.Close()
	defer parentDHT.Close()

	// Generate and add test file
	cid := addFile(ctx, parentBS)
	fmt.Printf("Parent generated CID: %s\n", cid.String())
	fmt.Printf("File contents: %s\n", getFileContents(ctx, parentBS, cid))

	// Advertise CID in DHT
	fmt.Println("Parent providing CID to DHT...")
	if err := parentDHT.Provide(ctx, cid, true); err != nil {
		log.Fatal("Provide failed:", err)
	}

	// Allow time for DHT propagation
	time.Sleep(5 * time.Second)

	// Fork child process
	cmd := exec.Command(os.Args[0], cid.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatal("Fork failed:", err)
	}

	// Wait for completion
	cmd.Wait()
}

func runChild(cidStr string) {
	ctx := context.Background()
	
	// Parse CID
	cid, err := cid.Parse(cidStr)
	if err != nil {
		log.Fatal("Invalid CID:", err)
	}

	// Create child node components
	childHost, childDHT, childBS := setupNode(ctx)
	defer childHost.Close()
	defer childDHT.Close()

	// Connect to parent (as bootstrap peer)
	connectToParent(ctx, childHost)

	// Find providers through DHT
	fmt.Println("Child searching for providers...")
	providers := findProviders(ctx, childDHT, cid)
	if len(providers) == 0 {
		log.Fatal("No providers found")
	}

	// Retrieve file through Bitswap
	fmt.Printf("Child retrieving from %s...\n", providers[0].ID)
	blk := retrieveBlock(ctx, childBS, cid)
	fmt.Printf("Child retrieved contents: %s\n", blk.RawData())
}

func setupNode(ctx context.Context) (host.Host, *dht.IpfsDHT, blockstore.Blockstore) {
	// Create libp2p host
	h, err := libp2p.New()
	if err != nil {
		log.Fatal(err)
	}

	// Create DHT
	dhtInst, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		log.Fatal(err)
	}

	// Bootstrap DHT
	if err := dhtInst.Bootstrap(ctx); err != nil {
		log.Fatal(err)
	}

	// Create blockstore and bitswap
	ds := dssync.MutexWrap(datastore.NewMapDatastore())
	bs := blockstore.NewBlockstore(ds)
	bsNet := bsnet.NewFromIpfsHost(h, dhtInst)
	bsExch := bitswap.New(ctx, bsNet, bs)

	// Start bitswap network
	bsNet.Start(bsExch)

	return h, dhtInst, bs
}

func addFile(ctx context.Context, bs blockstore.Blockstore) cid.Cid {
	// Generate random number 0-9999
	num, _ := rand.Int(rand.Reader, big.NewInt(10000))
	data := []byte(num.String())

	// Create CIDv1 raw block
	pref := cid.NewPrefixV1(cid.Raw, 0x00)
	cid, err := pref.Sum(data)
	if err != nil {
		log.Fatal(err)
	}

	// Create and store block
	blk, _ := blocks.NewBlockWithCid(data, cid)
	if err := bs.Put(ctx, blk); err != nil {
		log.Fatal(err)
	}

	return cid
}

func getFileContents(ctx context.Context, bs blockstore.Blockstore, c cid.Cid) []byte {
	blk, err := bs.Get(ctx, c)
	if err != nil {
		log.Fatal(err)
	}
	return blk.RawData()
}

func connectToParent(ctx context.Context, child host.Host) {
	// In real use we would parse multiaddr from command line
	// For demo, connect to default libp2p relay
	relayAddr := "/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN"
	ma, _ := multiaddr.NewMultiaddr(relayAddr)
	ai, _ := peer.AddrInfoFromP2pAddr(ma)
	
	if err := child.Connect(ctx, *ai); err != nil {
		log.Fatal("Parent connection failed:", err)
	}
}

func findProviders(ctx context.Context, dht *dht.IpfsDHT, c cid.Cid) []peer.AddrInfo {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	providers, err := dht.FindProviders(ctx, c)
	if err != nil {
		log.Fatal("FindProviders failed:", err)
	}
	return providers
}

func retrieveBlock(ctx context.Context, bs blockstore.Blockstore, c cid.Cid) blocks.Block {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	blk, err := bs.Get(ctx, c)
	if err != nil {
		log.Fatal("Block retrieval failed:", err)
	}
	return blk
}
