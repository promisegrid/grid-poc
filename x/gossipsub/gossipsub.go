package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	// mdns for local relay discovery
	mdns "github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// discoveryNotifee implements the mdns.Notifee interface to discover relays.
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound is called when new peer is found via mDNS.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	log.Println("Discovered relay candidate:", pi.ID)
	// Attempt to connect to the discovered peer.
	if err := n.h.Connect(context.Background(), pi); err != nil {
		log.Println("Error connecting to discovered peer:", err)
	} else {
		log.Println("Connected to relay:", pi.ID)
	}
}

func main() {
	// Create a libp2p host with relay enabled and auto relay enabled.
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/4001"),
		libp2p.EnableRelay(),     // Allow relay connections.
		libp2p.EnableAutoRelay(), // Automatically use discovered relays.
	)
	if err != nil {
		panic(err)
	}

	// Set up mDNS discovery to find relay candidates.
	ctx := context.Background()
	notifee := &discoveryNotifee{h: host}
	mdnsService, err := mdns.NewMdnsService(ctx, host, "relayDiscovery", notifee)
	if err != nil {
		log.Println("Failed to start mDNS relay discovery:", err)
	} else {
		log.Println("mDNS relay discovery service started.")
	}
	// Ensure the mDNS service is closed on shutdown.
	defer mdnsService.Close()

	// Create Gossipsub router with default parameters.
	gossipsub, err := pubsub.NewGossipSub(
		ctx,
		host,
		pubsub.WithMessageSignaturePolicy(pubsub.StrictSign),
	)
	if err != nil {
		panic(err)
	}

	// Subscribe to our topic.
	topic := "grid-demo"
	sub, err := gossipsub.Subscribe(topic)
	if err != nil {
		panic(err)
	}
	defer sub.Cancel()

	// Print our listening addresses.
	fmt.Println("Host ID:", host.ID())
	fmt.Println("Listening on:")
	for _, addr := range host.Addrs() {
		fmt.Println("  ", addr)
	}

	// Message handler.
	go func() {
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				log.Println("Subscription closed:", err)
				return
			}
			fmt.Printf("\n[%s] %s\n> ", msg.ReceivedFrom, string(msg.Data))
		}
	}()

	// Input loop.
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("> ")

	// Connect to other nodes if specified.
	if len(os.Args) > 1 {
		for _, addr := range os.Args[1:] {
			maddr, err := ma.NewMultiaddr(addr)
			if err != nil {
				log.Println("Invalid multiaddr:", err)
				continue
			}

			peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
			if err != nil {
				log.Println("Invalid peer address:", err)
				continue
			}

			err = host.Connect(ctx, *peerInfo)
			if err != nil {
				log.Println("Connection failed:", err)
			} else {
				log.Println("Connected to:", peerInfo.ID)
			}
		}
	}

	// Publish loop.
	go func() {
		for {
			text, _ := reader.ReadString('\n')
			if text == "\n" {
				continue
			}

			err := gossipsub.Publish(topic, []byte(text))
			if err != nil {
				log.Println("Publish error:", err)
			}
			fmt.Print("> ")
		}
	}()

	// Wait for termination signal.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down...")
	host.Close()
}
