package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	mrand "math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ipfs/go-datastore"
	dsync "github.com/ipfs/go-datastore/sync"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"

	// "github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"

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

	dht "github.com/libp2p/go-libp2p-kad-dht"

	. "github.com/stevegt/goadapt"
)

// list of public bootstrap peers as recommended for IPFS
var defaultBootstrapPeers = dht.DefaultBootstrapPeers

// libp2p host
var p2pHost host.Host

const exampleFn = "/tmp/boxo-example-peerid.txt"

// The CID of the file with the number 0 to 100k, built with the parameters:
// CIDv1 links, a 256bit sha2-256 hash function, raw-leaves, a balanced layout,
// 256kiB chunks, and 174 max links per block
// const fileCid = "bafybeiecq2irw4fl5vunnxo6cegoutv4de63h7n27tekkjtak3jrvrzzhe"

// dynamically generated fileCid
var fileCid string

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse options from the command line
	targetF := flag.String("d", "", "target peer to dial")
	// The -dht flag will cause the code to rely on DHT discovery instead of
	// proactively dialing a specific peer.
	targetDht := flag.Bool("dht", false, "use DHT to find the target peer instead of direct dialing")
	seedF := flag.Int64("seed", 0, "set random seed for id generation")
	pingF := flag.String("ping", "", "target peer to ping")
	flag.Parse()

	// Recommend increasing system UDP buffer sizes for better QUIC performance.
	configureUDPBuffers()

	// For this example we are going to be transferring data using Bitswap
	// over libp2p or demonstrating gossipsub. This means we need to create a
	// libp2p host first.

	// Make a host that listens on the given multiaddress
	h, err := makeHost(ctx, 0, *seedF)
	if err != nil {
		log.Fatal(err)
	}
	p2pHost = h // Set global after creation
	defer h.Close()

	// Set up the DHT to join the public IPFS network. Bootstrap peers are used
	// to help the DHT discover the public network.
	dht, err := setupDHT(ctx, h)
	if err != nil {
		log.Fatal(err)
	}
	defer dht.Close()

	fullAddr := getHostAddress(h)
	log.Printf("I am %s\n", fullAddr)

	// Start a goroutine to list connected peers every 10 seconds.
	go listConnectedPeers(ctx, h, false /* verbose */)

	// ping targetF or pingF
	for _, target := range []string{*targetF, *pingF} {
		if target == "" {
			continue
		}
		// convert target string to an AddInfo struct
		info, err := peer.AddrInfoFromString(target)
		if err != nil {
			log.Fatalf("Invalid target peer info: %v", err)
		}
		err = pingWait(ctx, h, *info)
		if err != nil {
			log.Fatalf("ping failed: %v: %v", err, target)
		}
	}

	// write the host's peer ID to a file for use in the demos
	if *targetF == "" && !*targetDht {
		// call WriteFile to write the peer ID to a file WriteFile is
		// in the std lib os package.  it returns an error if it fails
		err = os.WriteFile(exampleFn, []byte(fullAddr), 0644)
		Ck(err)
		log.Printf("Peer ID written to %s\n", exampleFn)
	}

	wg := sync.WaitGroup{}

	// run the Bitswap demo.
	go func() {
		wg.Add(1)
		if err := runBitswapDemo(ctx, h, *targetF, *targetDht, dht); err != nil {
			log.Print(fmt.Errorf("Bitswap demo failed: %v", err))
		}
		wg.Done()
	}()

	// run the gossipsub demo
	go func() {
		wg.Add(1)
		if err := runGossipDemo(ctx, h, *targetF, *targetDht, dht); err != nil {
			log.Print(fmt.Errorf("Gossipsub demo failed: %v", err))
		}
		wg.Done()
	}()

	time.Sleep(1 * time.Second)
	wg.Wait()
	return
}

// configureUDPBuffers logs recommended system commands to increase UDP buffer sizes.
func configureUDPBuffers() {
	// XXX read from sysctl first and only print if not set
	log.Println("For better QUIC performance, consider running:")
	log.Println("sudo sysctl -w net.core.rmem_max=7500000")
	log.Println("sudo sysctl -w net.core.wmem_max=7500000")
}

// listConnectedPeers periodically lists the connected peers of the host.
// This helps in monitoring the connectivity of the node.
func listConnectedPeers(ctx context.Context, h host.Host, verbose bool) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			peers := h.Network().Peers()
			if len(peers) == 0 {
				log.Println("No connected peers")
			} else {
				var ids []string
				for _, p := range peers {
					ids = append(ids, p.String())
				}
				if verbose {
					log.Println("Connected peers:")
				}
				peerCount := 0
				addrCount := 0
				for _, p := range peers {
					peerCount++
					if verbose {
						log.Printf("  %s:\n", p)
					}
					addrs := h.Peerstore().Addrs(p)
					for _, addr := range addrs {
						addrCount++
						if verbose {
							log.Printf("    %s\n", addr)
						}
					}
				}
				log.Printf("Connected to %d peers, %d addresses\n", peerCount, addrCount)
			}
		case <-ctx.Done():
			return
		}
	}
}

// setupDHT initializes a Kademlia DHT instance for the host and connects to
// a set of public bootstrap peers so that our node can join the IPFS network.
func setupDHT(ctx context.Context, h host.Host) (*dht.IpfsDHT, error) {
	// d, err := dht.New(ctx, h, dht.Mode(dht.ModeAuto))
	d, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		return nil, err
	}
	// Connect to each bootstrap peer.
	var ok, nok int
	for _, maddr := range defaultBootstrapPeers {
		// maddr is a multiaddr with the peer ID, so we can use it to get the
		// peer ID and address info.
		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			log.Printf("Invalid bootstrap peer info for %s: %v", maddr, err)
			continue
		}
		addrStr := maddr.String()
		if err := h.Connect(ctx, *info); err != nil {
			log.Printf("Error connecting to bootstrap peer %s: %v", addrStr,
				err)
			nok++
		} else {
			log.Printf("Connected to bootstrap peer: %s", addrStr)
			ok++
		}
	}
	// Bootstrap the DHT so that it starts the routing process.
	if err := d.Bootstrap(ctx); err != nil {
		return nil, err
	}
	log.Printf("DHT bootstrapped successfully: %d/%d peers connected",
		ok, ok+nok)
	return d, nil
}

// runGossipDemo runs a gossipsub demo that sends a message and waits for a
// response. If target is provided, this node acts as the sender, publishing
// "hello world" and waiting for a "hello back" from the responder. If no target
// is provided, this node acts as the responder, waiting for a "hello world" and
// replying with "hello back". The demo exits after a successful message exchange.
func runGossipDemo(ctx context.Context, h host.Host, target string, useDHT bool, dht *dht.IpfsDHT) error {

	// Create pubsub instance
	ps, err := pubsub.NewGossipSub(ctx, h,
		// use dht for discovery
		pubsub.WithDiscovery(routing.NewRoutingDiscovery(dht)), // Add DHT-based discovery
		// enable flood publishing to ensure messages reach all peers
		pubsub.WithFloodPublish(true),
	)

	if err != nil {
		return fmt.Errorf("failed to create pubsub: %w", err)
	}

	topic, err := ps.Join("gossip-demo")
	if err != nil {
		return err
	}
	sub, err := topic.Subscribe()
	if err != nil {
		return err
	}

	// If target peer is provided or DHT mode is enabled, act as sender
	if target != "" || useDHT {
		if !useDHT {
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
		} else {
			log.Println("Using DHT discovery for gossipsub; not dialing a specific peer")
		}

		// Allow time for connection and mesh formation
		log.Println("Waiting 2 seconds for connection stabilization...")
		select {
		case <-time.After(2 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}

		log.Println("Waiting for response on gossipsub (topic: gossip-demo)...")
		responseCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		done := make(chan error, 1)

		// start goroutine to handle incoming messages
		go func() {
			for {
				msg, err := sub.Next(responseCtx)
				if err != nil {
					log.Printf("Failed to receive message: %v", err)
					done <- err
					return
				}
				if msg.ReceivedFrom == h.ID() {
					continue
				}
				if strings.HasPrefix(string(msg.Data), "hello back") {
					log.Printf("Received %s", string(msg.Data))
					parts := strings.Split(string(msg.Data), " ")
					if len(parts) < 3 {
						log.Println("Invalid message format, ignoring...")
						continue
					}
					fileCid = parts[3]
					log.Printf("Parsed fileCid: %s", fileCid)
					done <- nil
					return
				}
			}
		}()

		// Publish with retries
		// create a context that times out after 60 seconds
		ctx60, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		const maxRetries = math.MaxInt
		for i := 0; i < maxRetries; i++ {
			msg := Spf("hello world %d", i+1)
			if err := topic.Publish(ctx60, []byte(msg)); err != nil {
				log.Printf("Publish attempt %d failed: %v", i+1, err)
			} else {
				log.Printf("Published message: hello world (attempt %d)", i+1)
			}

			// Wait before next attempt or until response is received
			select {
			case <-time.After(1 * time.Second):
			case err := <-done:
				if err != nil {
					return err
				}
				log.Println("Received response, exiting...")
				return nil
			case <-ctx60.Done():
				return ctx60.Err()
			}
		}
		return fmt.Errorf("did not receive a valid 'hello back' response after %d attempts", maxRetries)
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
			if strings.HasPrefix(string(msg.Data), "hello world") {
				log.Printf("Received %s, sending response...", string(msg.Data))
				// get the integer from the message
				parts := strings.Split(string(msg.Data), " ")
				if len(parts) < 3 {
					log.Println("Invalid message format, ignoring...")
					continue
				}
				numStr := parts[2]
				num, err := strconv.Atoi(numStr)
				if err != nil {
					log.Printf("Failed to parse number from message: %v", err)
					continue
				}
				log.Printf("Parsed number: %d", num)
				// Send response
				if fileCid == "" {
					log.Println("fileCid is empty, can't send response yet")
					continue
				}
				// Send fileCid in the response
				ack := Spf("hello back %d %s", num, fileCid)
				err = topic.Publish(ctx, []byte(ack))
				if err != nil {
					log.Printf("Response publish attempt failed: %v", err)
				} else {
					log.Printf("Response published: %s", ack)
				}
				time.Sleep(1 * time.Second)
			}
		}
	}
}

// runBitswapDemo runs the Bitswap demo. If target is empty and DHT mode is not
// specified, it runs in server mode hosting a UnixFS file and listening for
// Bitswap requests. If target is provided or DHT mode is enabled, it runs in
// client mode and downloads the file from a peer discovered either via a direct
// dial or via DHT.
func runBitswapDemo(ctx context.Context, h host.Host, target string,
	useDHT bool, dht *dht.IpfsDHT) error {
	if target == "" && !useDHT {
		// Pass DHT instance to startDataServer
		c, bs, err := startDataServer(ctx, h, dht)
		if err != nil {
			return err
		}
		defer bs.Close()
		log.Printf("hosting UnixFS file with CID: %s\n", c)
		log.Println("listening for inbound connections and Bitswap requests")
		log.Printf("Now run on a different terminal:\ngo run main.go -d $(cat %s)\n",
			exampleFn)
		<-ctx.Done()
	} else {
		for fileCid == "" {
			log.Println("Waiting for fileCid to be set...")
			time.Sleep(1 * time.Second)
		}
		log.Printf("downloading UnixFS file with CID: %s\n", fileCid)
		fileData, err := runClient(ctx, h, cid.MustParse(fileCid), target,
			useDHT, dht)
		if err != nil {
			return err
		}
		log.Println("found the data")
		// verify the data
		err = verifyFile0to100k(fileData)
		if err != nil {
			log.Println("the file was NOT all the numbers from 0 to 100k!")
			return err
		}
		log.Println("the file was all the numbers from 0 to 100k!")
	}
	return nil
}

// makeHost creates a libP2P host with a random peer ID listening on the given
// multiaddress.
func makeHost(ctx context.Context, listenPort int, randseed int64) (host.Host, error) {
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

	peerSource, err := autoRelayPeerSource(ctx)
	if err != nil {
		return nil, err
	}

	// Some basic libp2p options, see the go-libp2p docs for more details
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d",
			listenPort)),
		libp2p.Identity(priv),
		libp2p.EnableNATService(),
		libp2p.EnableRelayService(),
		libp2p.EnableAutoRelayWithPeerSource(peerSource),
	}

	libp2pHost, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	return libp2pHost, nil
}

func getHostAddress(h host.Host) string {
	// Build host multiaddress
	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s",
		h.ID().String()))

	// Now we can build a full multiaddress to reach this host by
	// encapsulating both addresses:
	addr := h.Addrs()[0]
	return addr.Encapsulate(hostAddr).String()
}

// autoRelayPeerSource provides a dynamic Peer Source function for AutoRelay.
// A Peer Source is a function that returns a channel of potential relay nodes,
// which are peers that may help relay traffic when direct connections are not
// possible. Here we use the host's connected peers as candidates. Note that this
// implementation does not use static relays.
//
// Returns a function with this signature:
// type PeerSource func(ctx context.Context, num int) <-chan peer.AddrInfo
func autoRelayPeerSource(ctx context.Context) (fn autorelay.PeerSource, err error) {
	fn = func(ctx context.Context, num int) <-chan peer.AddrInfo {
		ch := make(chan peer.AddrInfo)
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				if p2pHost == nil {
					log.Println("Host is nil, waiting for it to be set...")
					time.Sleep(1 * time.Second)
					continue
				}
				select {
				case <-ticker.C:
					// Iterate over connected peers and send them as candidate relays.
					for _, pid := range p2pHost.Network().Peers() {
						// Skip self.
						if pid == p2pHost.ID() {
							continue
						}
						addrs := p2pHost.Peerstore().Addrs(pid)
						if len(addrs) == 0 {
							continue
						}
						select {
						case ch <- peer.AddrInfo{ID: pid, Addrs: addrs}:
						case <-ctx.Done():
							close(ch)
							return
						}
					}
				case <-ctx.Done():
					close(ch)
					return
				}
			}
		}()
		return ch
	}
	return
}

// createFile0to100k creates a file with the number 0 to 100k
func createFile0to100k() ([]byte, error) {
	b := strings.Builder{}
	for i := 0; i <= 100000; i++ {
		s := strconv.Itoa(i)
		_, err := b.WriteString(s + "\n")
		if err != nil {
			return nil, err
		}
	}
	return []byte(b.String()), nil
}

// verifyFile0to100k verifies that the file contains the number 0 to 100k
func verifyFile0to100k(fileData []byte) error {
	i := 0
	// create scanner
	scanner := bufio.NewScanner(bytes.NewReader(fileData))
	// read each line
	for scanner.Scan() {
		line := scanner.Text()
		// check if the line is equal to the number
		s := strconv.Itoa(i)
		if line != s {
			// dump fileData to stdout
			fmt.Printf("fileData: %s\n", fileData)
			log.Printf("file does not contain the number %d: %s", i, line)
			return fmt.Errorf("file does not contain the number %d", i)
		}
		i++
	}
	return nil
}

func startDataServer(ctx context.Context, h host.Host, dht *dht.IpfsDHT) (cid.Cid,
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
		Maxlinks:  uih.DefaultLinksPerBlock, // Default max of 174 links per block
		RawLeaves: true,                     // Leave the actual file bytes untouched
		// instead of wrapping them in a dag-pb protobuf wrapper
		CidBuilder: cid.V1Builder{ // Use CIDv1 for all links
			Codec:    uint64(multicodec.DagPb),
			MhType:   uint64(multicodec.Sha2_256), // Use SHA2-256 as the hash
			MhLength: -1,                          // Use the default hash length for the
			// given hash function
		},
		Dagserv: dsrv,
		NoCopy:  false,
	}
	ufsBuilder, err := ufsImportParams.New(
		chunker.NewSizeSplitter(fileReader, chunker.DefaultBlockSize))
	if err != nil {
		return cid.Undef, nil, err
	}
	// Arrange the graph with a balanced layout
	nd, err := balanced.Layout(ufsBuilder)
	if err != nil {
		return cid.Undef, nil, err
	}
	rootCid := nd.Cid()

	// hang onto the fileCid so we can respond with it in pubsub
	fileCid = rootCid.String()

	// Advertise CID through DHT if available
	log.Printf("DHT: %p\n", dht)
	if dht != nil {
		if err := dht.Provide(ctx, rootCid, true); err != nil {
			return cid.Undef, nil, fmt.Errorf("failed to announce CID via DHT: %v", err)
		}
		// Start a goroutine to periodically reprovide the CID every 10 seconds.
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					if err := dht.Provide(ctx, rootCid, true); err != nil {
						log.Printf("failed to reprovide CID: %v", err)
					} else {
						log.Printf("reprovided CID: %s", rootCid)
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Start Bitswap server
	n := bsnet.NewFromIpfsHost(h)
	bswap := bsserver.New(ctx, n, bs)
	n.Start(bswap)

	return rootCid, bswap, nil
}

func runClient(ctx context.Context, h host.Host, c cid.Cid,
	target string, useDHT bool, dht *dht.IpfsDHT) ([]byte, error) {
	n := bsnet.NewFromIpfsHost(h)
	bswap := bsclient.New(ctx, n, nil,
		blockstore.NewBlockstore(datastore.NewNullDatastore()))
	n.Start(bswap)
	defer bswap.Close()

	if !useDHT {
		// Turn the targetPeer into a multiaddr.
		maddr, err := multiaddr.NewMultiaddr(target)
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
	} else {
		log.Println("Searching for providers via DHT...")
		// try providers until file is fetched
		tried := 0
		var provChan <-chan peer.AddrInfo
		open := false
		for {
			time.Sleep(1 * time.Second)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			if !open {
				Pl("Starting DHT FindProvidersAsync")
				// dht.FindProvidersAsync starts a goroutine to find providers
				provChan = dht.FindProvidersAsync(ctx, c, 999)
				open = true
			}
			log.Printf("tried %d providers, waiting for next...\n", tried)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case prov, ok := <-provChan:
				if !ok {
					log.Println("Provider channel closed")
					open = false
					continue
				}
				log.Printf("got candidate provider %s\n", prov.ID)
				// Skip self.
				if prov.ID == h.ID() {
					log.Println("Skipping self as provider")
					continue
				}

				tried++

				// Ping the provider to verify availability.  This
				// also adds the provider to the peerstore.
				err := pingWait(ctx, h, prov)
				if err != nil {
					log.Printf("Ping failed on provider %s: %v", prov.ID, err)
					continue
				}

				log.Printf("Attempting connection to provider %s:", prov.ID)
				for _, addr := range prov.Addrs {
					log.Printf("  %s\n", addr)
				}

				ctxConn, cancelConn := context.WithTimeout(ctx, 10*time.Second)
				defer cancelConn()
				err = h.Connect(ctxConn, prov)
				if err != nil {
					log.Printf("Connection to provider %s failed: %v", prov.ID, err)
					cancelConn()
					continue
				}

				log.Printf("Connected to provider %s", prov.ID)

				if h.Network().Connectedness(prov.ID) != network.Connected {
					log.Printf("Not fully connected to provider %s", prov.ID)
					cancelConn()
					continue
				}

				// create a context that times out after 60 seconds for fetching the file
				ctx60, cancel := context.WithTimeout(ctx, 60*time.Second)
				defer cancel()
				log.Printf("Fetching file from provider %s...\n", prov.ID)

				// Replace dserv creation with session-based approach
				sess := bswap.NewSession(ctx)
				// defer sess.Shutdown()

				dserv := merkledag.NewReadOnlyDagService(
					blockservice.New(
						blockstore.NewBlockstore(datastore.NewNullDatastore()),
						sess,
					),
				)

				// try to fetch the file
				nd, err := dserv.Get(ctx60, c)
				cancel()
				if err != nil {
					log.Printf("Error getting file from provider %s: %v", prov.ID, err)
					cancelConn()
					cancel()
					continue
				}
				log.Printf("Got file from provider %s", prov.ID)

				unixFSNode, err := unixfile.NewUnixfsFile(ctx, dserv, nd)
				if err != nil {
					return nil, err
				}
				log.Printf("Created UnixFS file from node")

				var buf bytes.Buffer
				if f, ok := unixFSNode.(files.File); ok {
					if _, err := io.Copy(&buf, f); err != nil {
						return nil, err
					}
				}
				log.Printf("Copied file data to buffer")
				return buf.Bytes(), nil
			}
		}
	}
	return nil, nil
}

// pingWait pings the target peer using the libp2p ping protocol and
// returns a nil err when the ping is successful.
func pingWait(ctx context.Context, h host.Host, target peer.AddrInfo) error {
	// Try to parse target as a multiaddr.
	log.Printf("Pinging peer ID %s...", target.ID)
	// Add each address to the peerstore.
	for _, a := range target.Addrs {
		h.Peerstore().AddAddr(target.ID, a, time.Hour)
	}
	for i := 0; i < 3; i++ {
		pingCh := ping.Ping(ctx, h, target.ID)
		result := <-pingCh
		if result.Error != nil {
			log.Printf("Ping error: %s", result.Error)
		} else {
			log.Printf("Ping RTT: %s", result.RTT)
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("ping timeout")
}

// isRelayAddr checks whether a multiaddress is a relay address via the "p2p-circuit" protocol.
func isRelayAddr(maddr multiaddr.Multiaddr) bool {
	for _, proto := range maddr.Protocols() {
		if proto.Name == "p2p-circuit" {
			return true
		}
	}
	return false
}

// isLoopbackAddr checks if the address is a local host address.
func isLoopbackAddr(maddr multiaddr.Multiaddr) bool {
	// Check if the address is a loopback address
	for _, loopPat := range []string{"127.0.0.1", "::1"} {
		if strings.Contains(maddr.String(), loopPat) {
			return true
		}
	}
	return false
}
