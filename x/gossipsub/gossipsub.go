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

	// Create a routing setup that uses the Kademlia DHT
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
		// Listen on TCP
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),

		// Enable relay features: supply a Peer Source that returns the default bootstrap peers.
		libp2p.EnableRelay(),
		libp2p.EnableAutoRelay(autorelay.WithPeerSource(
			func() []peer.AddrInfo {
				return convertBootstrapPeers(dht.DefaultBootstrapPeers)
			},
		)),

		// NAT traversal
		libp2p.NATPortMap(),
		libp2p.EnableHolePunching(),

		// DHT for peer discovery
		libp2p.Routing(newDHT),
	)
	if err != nil {
		panic(err)
	}

	// Bootstrap the DHT
	log.Println("Bootstrapping DHT...")
	if err = dhtNode.Bootstrap(ctx); err != nil {
		panic(err)
	}

	// Connect to bootstrap peers
	connectToBootstrapPeers(ctx, h, convertBootstrapPeers(dht.DefaultBootstrapPeers))

	// Set up Gossipsub with DHT-based peer discovery
	ps, err := pubsub.NewGossipSub(ctx, h,
		pubsub.WithDiscovery(drouting.NewRoutingDiscovery(dhtNode)),
		pubsub.WithPeerExchange(true),
	)
	if err != nil {
		panic(err)
	}

	// Join the topic
	topic, err := ps.Join(topicName)
	if err != nil {
		panic(err)
	}

	// Subscribe to topic
	sub, err := topic.Subscribe()
	if err != nil {
		panic(err)
	}
	defer sub.Cancel()

	// Set up mDNS for local discovery
	setupMdnsDiscovery(ctx, h, rendezvous)

	// Advertise our presence using DHT
	routingDiscovery := drouting.NewRoutingDiscovery(dhtNode)
	log.Println("Advertising ourselves...")
	if _, err := routingDiscovery.Advertise(ctx, rendezvous); err != nil {
		log.Printf("Failed to advertise: %v", err)
	}

	// Start message handler
	go handleMessages(ctx, sub)

	// Print node information
	log.Printf("Node ID: %s\n", h.ID())
	log.Printf("Connect using: /p2p-circuit/p2p/%s\n", h.ID())

	// Periodically discover peers automatically
	go discoverPeers(ctx, h, routingDiscovery)

	// Start input loop in a separate goroutine
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

	// Wait for exit signal: either OS signal or user input closure (Ctrl-D)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigCh:
	case <-inputDone:
	}

	log.Println("Shutting down...")
	cancel()
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

			// Collect peers
			var peerList []peer.AddrInfo
			for p := range peers {
				peerList = append(peerList, p)
			}
			log.Printf("Found %d peers for topic %s", len(peerList), rendezvous)

			// Connect to new peers
			var connected, unconnected []peer.AddrInfo
			for _, p := range peerList {
				if p.ID == h.ID() || h.Network().Connectedness(p.ID) == network.Connected {
					connected = append(connected, p)
					continue
				}

				h.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.PermanentAddrTTL)
				if err := h.Connect(ctx, p); err != nil {
					// log.Printf("Failed to connect to %s: %v", p.ID, err)
					// log.Printf("Failed to connect to %s", p.ID)
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
