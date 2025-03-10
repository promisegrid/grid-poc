package main

import (
	"context"
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
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

const (
	rendezvous = "promisegrid-demo"
	topicName  = "grid-network"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create libp2p host with DHT and relay capabilities
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.EnableRelay(),
		libp2p.EnableAutoRelay(),
		libp2p.NATPortMap(),
	)
	if err != nil {
		panic(err)
	}

	// Set up DHT with default IPFS bootstrap nodes
	bootstrapPeers := dht.GetDefaultBootstrapPeerAddrInfos()
	dhtMode := dht.Mode(
		dht.ModeAuto,
	)
	kademliaDHT, err := dht.New(ctx, h,
		dhtMode,
		dht.BootstrapPeers(bootstrapPeers...),
	)
	if err != nil {
		panic(err)
	}

	// Bootstrap the DHT
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		panic(err)
	}

	// Connect to bootstrap peers
	connectToBootstrapPeers(ctx, h, bootstrapPeers)

	// Set up Gossipsub with DHT-based peer discovery
	ps, err := pubsub.NewGossipSub(ctx, h,
		pubsub.WithDiscovery(drouting.NewRoutingDiscovery(kademliaDHT)),
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
	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	routingDiscovery.Advertise(ctx, rendezvous)

	// Start message handler
	go handleMessages(ctx, sub)

	// Print node information
	log.Printf("Node ID: %s\n", h.ID())
	log.Printf("Connect using: /p2p-circuit/p2p/%s\n", h.ID())

	// Maintain network connectivity
	go discoverPeers(ctx, h, routingDiscovery)

	// Wait for exit signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
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

			for p := range peers {
				if p.ID == h.ID() || h.Network().Connectedness(p.ID) == network.Connected {
					continue
				}

				h.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.PermanentAddrTTL)
				if err := h.Connect(ctx, p); err != nil {
					log.Printf("Failed to connect to %s: %v", p.ID, err)
				} else {
					log.Printf("Connected to new peer: %s", p.ID)
				}
			}
		}
	}
}

func connectToBootstrapPeers(ctx context.Context, h host.Host, peers []peer.AddrInfo) {
	for _, p := range peers {
		h.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.PermanentAddrTTL)
		if err := h.Connect(ctx, p); err != nil {
			log.Printf("Failed to connect to bootstrap peer %s: %v", p.ID, err)
		} else {
			log.Printf("Connected to bootstrap peer: %s", p.ID)
		}
	}
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
