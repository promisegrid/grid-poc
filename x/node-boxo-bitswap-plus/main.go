package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/ipfs/go-datastore"
	dsync "github.com/ipfs/go-datastore/sync"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multicodec"

	"github.com/ipfs/boxo/blockservice"
	blockstore "github.com/ipfs/boxo/blockstore"
	chunker "github.com/ipfs/boxo/chunker"
	offline "github.com/ipfs/boxo/exchange/offline"
	"github.com/ipfs/boxo/ipld/merkledag"
	unixfile "github.com/ipfs/boxo/ipld/unixfs/file"
	"github.com/ipfs/boxo/ipld/unixfs/importer/balanced"
	uih "github.com/ipfs/boxo/ipld/unixfs/importer/helpers"

	bsclient "github.com/ipfs/boxo/bitswap/client"
	bsnet "github.com/ipfs/boxo/bitswap/network/bsnet"
	bsserver "github.com/ipfs/boxo/bitswap/server"
	"github.com/ipfs/boxo/files"
)

const exampleBinaryName = "bitswap-transfer"

// The CID of the file with the number 0 to 100k, built with the parameters:
// CIDv1 links, a 256bit sha2-256 hash function, raw-leaves, a balanced layout,
// 256kiB chunks, and 174 max links per block
const fileCid = "bafybeiecq2irw4fl5vunnxo6cegoutv4de63h7n27tekkjtak3jrvrzzhe"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse options from the command line
	targetF := flag.String("d", "", "target peer to dial")
	seedF := flag.Int64("seed", 0, "set random seed for id generation")
	gsF := flag.Bool("gs", false, "Run gossipsub demo")
	flag.Parse()

	// For this example we are going to be transferring data using Bitswap
	// over libp2p or demonstrating gossipsub. This means we need to create a
	// libp2p host first.

	// Make a host that listens on the given multiaddress
	h, err := makeHost(0, *seedF)
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	fullAddr := getHostAddress(h)
	log.Printf("I am %s\n", fullAddr)

	// If the gossipsub demo is requested, run it.
	if *gsF {
		if err := runGossipDemo(ctx, h, *targetF); err != nil {
			log.Fatal(err)
		}
		return
	}

	// Otherwise run the Bitswap demo.
	if err := runBitswapDemo(ctx, h, *targetF); err != nil {
		log.Fatal(err)
	}
}

// runGossipDemo runs a gossipsub demo that sends a message and waits for a
// response. If target is provided, this node acts as the sender, publishing
// "hello world" and waiting for a "hello back" from the responder.  If no target
// is provided, this node acts as the responder, waiting for a "hello world" and
// replying with "hello back".
func runGossipDemo(ctx context.Context, h host.Host, target string) error {
	// Enable flood publishing to ensure messages reach all peers
	ps, err := pubsub.NewGossipSub(ctx, h, pubsub.WithFloodPublish(true))
	if err != nil {
		return err
	}
	topic, err := ps.Join("gossip-demo")
	if err != nil {
		return err
	}
	sub, err := topic.Subscribe()
	if err != nil {
		return err
	}

	// If target peer is provided, act as the publisher.
	if target != "" {
		maddr, err := multiaddr.NewMultiaddr(target)
		if err != nil {
			return err
		}
		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return err
		}
		if err := h.Connect(ctx, *info); err != nil {
			return err
		}

		// Allow time for connection and mesh formation
		log.Println("Waiting 2 seconds for connection stabilization...")
		select {
		case <-time.After(2 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}

		// Publish with retries
		const maxRetries = 5
		for i := 0; i < maxRetries; i++ {
			if err := topic.Publish(ctx, []byte("hello world")); err != nil {
				log.Printf("Publish attempt %d failed: %v", i+1, err)
			} else {
				log.Printf("Published message: hello world (attempt %d)", i+1)
			}

			// Wait before next attempt
			select {
			case <-time.After(1 * time.Second):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		log.Println("Waiting for response on gossipsub (topic: gossip-demo)...")
		responseCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		for {
			msg, err := sub.Next(responseCtx)
			if err != nil {
				return fmt.Errorf("failed to receive response: %w", err)
			}
			if msg.ReceivedFrom == h.ID() {
				continue
			}
			if string(msg.Data) == "hello back" {
				log.Printf("Received response: %s", string(msg.Data))
				return nil
			}
		}
	} else {
		// Act as responder: wait for "hello world" then send "hello back".
		log.Println("Waiting for message on gossipsub (topic: gossip-demo)...")
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				return err
			}
			if msg.ReceivedFrom == h.ID() {
				continue
			}
			if string(msg.Data) == "hello world" {
				log.Println("Received 'hello world' message, sending response")
				
				// Send response with retries
				const maxRetries = 3
				for i := 0; i < maxRetries; i++ {
					if err := topic.Publish(ctx, []byte("hello back")); err != nil {
						log.Printf("Response publish attempt %d failed: %v", i+1, err)
					} else {
						log.Printf("Sent 'hello back' response (attempt %d)", i+1)
					}
					time.Sleep(500 * time.Millisecond)
				}
				return nil
			}
		}
	}
}

// runBitswapDemo runs the Bitswap demo. If target is empty, it runs in server
// mode hosting a UnixFS file and listening for Bitswap requests.  If target is
// provided, it runs in client mode and downloads the file from the target.
func runBitswapDemo(ctx context.Context, h host.Host, target string) error {
	if target == "" {
		c, bs, err := startDataServer(ctx, h)
		if err != nil {
			return err
		}
		defer bs.Close()
		log.Printf("hosting UnixFS file with CID: %s\n", c)
		log.Println("listening for inbound connections and Bitswap requests")
		log.Printf("Now run \"./%s -d %s\" on a different terminal\n",
			exampleBinaryName, getHostAddress(h))
		<-ctx.Done()
	} else {
		log.Printf("downloading UnixFS file with CID: %s\n", fileCid)
		fileData, err := runClient(ctx, h, cid.MustParse(fileCid), target)
		if err != nil {
			return err
		}
		log.Println("found the data")
		log.Println(string(fileData))
		log.Println("the file was all the numbers from 0 to 100k!")
	}
	return nil
}

// makeHost creates a libP2P host with a random peer ID listening on the given
// multiaddress.
func makeHost(listenPort int, randseed int64) (host.Host, error) {
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it at least to obtain a
	// valid host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	// Some basic libp2p options, see the go-libp2p docs for more details
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	}

	return libp2p.New(opts...)
}

func getHostAddress(h host.Host) string {
	// Build host multiaddress
	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", h.ID().String()))

	// Now we can build a full multiaddress to reach this host by
	// encapsulating both addresses:
	addr := h.Addrs()[0]
	return addr.Encapsulate(hostAddr).String()
}

// createFile0to100k creates a file with the number 0 to 100k
func createFile0to100k() ([]byte, error) {
	b := strings.Builder{}
	for i := 0; i <= 100000; i++ {
		s := strconv.Itoa(i)
		_, err := b.WriteString(s)
		if err != nil {
			return nil, err
		}
	}
	return []byte(b.String()), nil
}

func startDataServer(ctx context.Context, h host.Host) (cid.Cid,
	*bsserver.Server, error) {
	fileBytes, err := createFile0to100k()
	if err != nil {
		return cid.Undef, nil, err
	}
	fileReader := bytes.NewReader(fileBytes)

	ds := dsync.MutexWrap(datastore.NewMapDatastore())
	bs := blockstore.NewBlockstore(ds)
	bs = blockstore.NewIdStore(bs) // handle identity multihashes, these don't
	// do any actual lookups

	bsrv := blockservice.New(bs, offline.Exchange(bs))
	dsrv := merkledag.NewDAGService(bsrv)

	// Create a UnixFS graph from our file, parameters described here but
	// can be visualized at https://dag.ipfs.tech/
	ufsImportParams := uih.DagBuilderParams{
		Maxlinks:   uih.DefaultLinksPerBlock, // Default max of 174 links per block
		RawLeaves:  true,                     // Leave the actual file bytes untouched
		// instead of wrapping them in a dag-pb protobuf wrapper
		CidBuilder: cid.V1Builder{ // Use CIDv1 for all links
			Codec:    uint64(multicodec.DagPb),
			MhType:   uint64(multicodec.Sha2_256), // Use SHA2-256 as the hash function
			MhLength: -1, // Use the default hash length for the given hash function
		},
		Dagserv: dsrv,
		NoCopy:  false,
	}
	ufsBuilder, err := ufsImportParams.New(
		chunker.NewSizeSplitter(fileReader, chunker.DefaultBlockSize))
	if err != nil {
		return cid.Undef, nil, err
	}
	nd, err := balanced.Layout(ufsBuilder) // Arrange the graph with a balanced
	// layout
	if err != nil {
		return cid.Undef, nil, err
	}

	// Start listening on the Bitswap protocol. For this example we're not
	// leveraging any content routing (DHT, IPNI, delegated routing requests, etc.)
	// as we know the peer we are fetching from.
	n := bsnet.NewFromIpfsHost(h)
	bswap := bsserver.New(ctx, n, bs)
	n.Start(bswap)
	return nd.Cid(), bswap, nil
}

func runClient(ctx context.Context, h host.Host, c cid.Cid,
	targetPeer string) ([]byte, error) {
	n := bsnet.NewFromIpfsHost(h)
	bswap := bsclient.New(ctx, n, nil,
		blockstore.NewBlockstore(datastore.NewNullDatastore()))
	n.Start(bswap)
	defer bswap.Close()

	// Turn the targetPeer into a multiaddr.
	maddr, err := multiaddr.NewMultiaddr(targetPeer)
	if err != nil {
		return nil, err
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return nil, err
	}

	// Directly connect to the peer that we know has the content.
	if err := h.Connect(ctx, *info); err != nil {
		return nil, err
	}

	dserv := merkledag.NewReadOnlyDagService(merkledag.NewSession(ctx,
		merkledag.NewDAGService(blockservice.New(
			blockstore.NewBlockstore(datastore.NewNullDatastore()), bswap))))
	nd, err := dserv.Get(ctx, c)
	if err != nil {
		return nil, err
	}

	unixFSNode, err := unixfile.NewUnixfsFile(ctx, dserv, nd)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if f, ok := unixFSNode.(files.File); ok {
		if _, err := io.Copy(&buf, f); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}
