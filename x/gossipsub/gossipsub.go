package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
)

const (
	rendezvous = "promisegrid-demo"
	topicName  = "grid-network"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a routing setup that uses the Kademlia DHT.
	var dhtNode *dht.IpfsDHT
	newDHT := func(h host.Host) (routing.PeerRouting, error) {
		var err error
		dhtNode, err = dht.New(ctx, h, dht.Mode(dht.ModeAutoServer))
		return dhtNode, err
	}

	// Create libp2p host with relay and NAT traversal enabled.
	// To avoid the panic ("Can not create a new relayFinder. Need a Peer Source fn or a list of static relays"),
	// we supply a peer source function to EnableAutoRelay.
	h, err := libp2p.New(
		// Listen on TCP.
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),

		// Enable relay features: supply a Peer Source that returns the default bootstrap peers.
		libp2p.EnableRelay(),
		libp2p.EnableAutoRelay(autorelay.WithPeerSource(
			func() []peer.AddrInfo {
				return convertBootstrapPeers(dht.DefaultBootstrapPeers)
			},
		)),

		// NAT traversal.
		libp2p.NATPortMap(),
		libp2p.EnableHolePunching(),

		// DHT for peer discovery.
		libp2p.Routing(newDHT),
	)
	if err != nil {
		panic(err)
	}

	// Bootstrap the DHT.
	log.Println("Bootstrapping DHT...")
	if err = dhtNode.Bootstrap(ctx); err != nil {
		panic(err)
	}

	// Connect to bootstrap peers.
	connectToBootstrapPeers(ctx, h, convertBootstrapPeers(dht.DefaultBootstrapPeers))

	// Set up Gossipsub with DHT-based peer discovery.
	ps, err := pubsub.NewGossipSub(ctx, h,
		pubsub.WithDiscovery(drouting.NewRoutingDiscovery(dhtNode)),
		pubsub.WithPeerExchange(true),
	)
	if err != nil {
		panic(err)
	}

	// Join the primary topic (for chat/input messages).
	topic, err := ps.Join(topicName)
	if err != nil {
		panic(err)
	}

	// Subscribe to the primary topic.
	sub, err := topic.Subscribe()
	if err != nil {
		panic(err)
	}
	defer sub.Cancel()

	// Set up mDNS for local discovery.
	setupMdnsDiscovery(ctx, h, rendezvous)

	// Advertise our presence using DHT.
	routingDiscovery := drouting.NewRoutingDiscovery(dhtNode)
	log.Println("Advertising ourselves...")
	if _, err := routingDiscovery.Advertise(ctx, rendezvous); err != nil {
		log.Printf("Failed to advertise: %v", err)
	}

	// Create a three-node DAG using the IPFS RPC API and publish its root CID
	// to the 't7a' pubsub topic.
	if err = createAndPublishDAG(ctx, ps); err != nil {
		log.Printf("Error creating & publishing DAG: %v", err)
	}

	// Start message handler for the primary topic.
	go handleMessages(ctx, sub)

	// Print node information.
	log.Printf("Node ID: %s\n", h.ID())
	log.Printf("Connect using: /p2p-circuit/p2p/%s\n", h.ID())

	// Periodically discover peers automatically.
	go discoverPeers(ctx, h, routingDiscovery)

	// Start input loop in a separate goroutine.
	inputDone := make(chan struct{})
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Enter text to send to the network. Press Ctrl-D to exit.")
		for scanner.Scan() {
			text := scanner.Text()
			if err := topic.Publish(ctx, []byte(text)); err != nil {
				log.Printf("Failed to publish message: %v", err)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Error reading input: %v", err)
		}
		log.Println("Input closed. Exiting...")
		close(inputDone)
	}()

	// Wait for exit signal: either OS signal or user input closure (Ctrl-D).
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigCh:
	case <-inputDone:
	}

	log.Println("Shutting down...")
	cancel()
}

// createAndPublishDAG creates a three-node DAG using the IPFS RPC API.
// It creates two leaf nodes and one internal node that links to both leaves.
// The root CID is then published to the pubsub topic 't7a'.
// This example utilizes updated IPFS RPC client APIs from 2024.
func createAndPublishDAG(ctx context.Context, ps *pubsub.PubSub) error {
	// Connect to the local IPFS node via the RPC API.
	sh := shell.NewShell("localhost:5001")
	if sh == nil {
		return fmt.Errorf("failed to create IPFS shell")
	}

	// Create the first leaf node.
	leaf1 := map[string]interface{}{
		"data": "leaf node 1 content",
	}
	leaf1Cid, err := sh.DagPut(ctx, leaf1, "dag-cbor", "json")
	if err != nil {
		return fmt.Errorf("failed to create leaf1: %w", err)
	}

	// Create the second leaf node.
	leaf2 := map[string]interface{}{
		"data": "leaf node 2 content",
	}
	leaf2Cid, err := sh.DagPut(ctx, leaf2, "dag-cbor", "json")
	if err != nil {
		return fmt.Errorf("failed to create leaf2: %w", err)
	}

	// Create the internal node, linking to the two leaf nodes.
	// Using the IPFS DAG link notation: { "/": "<cid>" }.
	internal := map[string]interface{}{
		"child1": map[string]interface{}{"/": leaf1Cid.String()},
		"child2": map[string]interface{}{"/": leaf2Cid.String()},
	}
	rootCid, err := sh.DagPut(ctx, internal, "dag-cbor", "json")
	if err != nil {
		return fmt.Errorf("failed to create internal node: %w", err)
	}

	log.Printf("Created three-node DAG with root CID: %s", rootCid.String())

	// Join the pubsub topic 't7a' for publishing the root CID.
	t7aTopic, err := ps.Join("t7a")
	if err != nil {
		return fmt.Errorf("failed to join pubsub topic 't7a': %w", err)
	}
	// Note: We close the topic after publishing the DAG.
	defer t7aTopic.Close()

	// Publish the root CID to the topic.
	if err := t7aTopic.Publish(ctx, []byte(rootCid.String())); err != nil {
		return fmt.Errorf("failed to publish root CID to topic 't7a': %w", err)
	}
	log.Printf("Published root CID %s to pubsub topic 't7a'", rootCid.String())
	return nil
}

func handleMessages(ctx context.Context, sub *pubsub.Subscription) {
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			if ctx.Err() == nil {
				log.Printf("Subscription error: %v", err)
			}
			return
		}
		log.Printf("[%s] %s", msg.GetFrom(), string(msg.Data))
	}
}

func discoverPeers(ctx context.Context, h host.Host, discovery *drouting.RoutingDiscovery) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			peers, err := discovery.FindPeers(ctx, rendezvous)
			if err != nil {
				log.Printf("Peer discovery failed: %v", err)
				continue
			}

			// Collect peers.
			var peerList []peer.AddrInfo
			for p := range peers {
				peerList = append(peerList, p)
			}
			log.Printf("Found %d peers for topic %s", len(peerList), rendezvous)

			// Connect to new peers.
			var connected, unconnected []peer.AddrInfo
			for _, p := range peerList {
				if p.ID == h.ID() || h.Network().Connectedness(p.ID) == network.Connected {
					connected = append(connected, p)
					continue
				}

				h.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.PermanentAddrTTL)
				if err := h.Connect(ctx, p); err != nil {
					unconnected = append(unconnected, p)
				} else {
					connected = append(connected, p)
					log.Printf("Connected to new peer: %s", p.ID)
				}
			}

			for _, p := range connected {
				fmt.Printf(" - %s (connected)\n", p.ID)
			}

			for _, p := range unconnected {
				fmt.Printf(" - %s (unconnected)\n", p.ID)
			}
		}
	}
}

func connectToBootstrapPeers(ctx context.Context, h host.Host, peers []peer.AddrInfo) {
	for _, p := range peers {
		if p.ID == h.ID() {
			continue
		}
		h.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.PermanentAddrTTL)
		if err := h.Connect(ctx, p); err != nil {
			log.Printf("Failed to connect to bootstrap peer %s: %v", p.ID, err)
		} else {
			log.Printf("Connected to bootstrap peer: %s", p.ID)
		}
	}
}

func convertBootstrapPeers(addrs []multiaddr.Multiaddr) []peer.AddrInfo {
	var peers []peer.AddrInfo
	for _, addr := range addrs {
		pinfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			continue
		}
		peers = append(peers, *pinfo)
	}
	return peers
}

type mdnsNotifee struct {
	host host.Host
}

func (m *mdnsNotifee) HandlePeerFound(p peer.AddrInfo) {
	log.Printf("Discovered peer via mDNS: %s\n", p.ID)
	m.host.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.PermanentAddrTTL)
	if err := m.host.Connect(context.Background(), p); err != nil {
		log.Printf("Failed to connect to mDNS peer: %v", err)
	}
}

func setupMdnsDiscovery(ctx context.Context, h host.Host, serviceTag string) {
	mdnsService := mdns.NewMdnsService(h, serviceTag, &mdnsNotifee{h})
	if err := mdnsService.Start(); err != nil {
		log.Printf("Failed to start mDNS service: %v", err)
	}
}
