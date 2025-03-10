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
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/multiformats/go-multiaddr"
)

func main() {
	// Create a libp2p host
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/4001"),
		libp2p.EnableRelay(),
		// libp2p.EnableAutoRelay(), // Actively find and use relays
	)
	if err != nil {
		panic(err)
	}

	// Create Gossipsub router with default parameters
	gossipsub, err := pubsub.NewGossipSub(
		context.Background(),
		host,
		pubsub.WithMessageSignaturePolicy(pubsub.StrictSign),
	)
	if err != nil {
		panic(err)
	}

	// Subscribe to our topic
	topic := "grid-demo"
	sub, err := gossipsub.Subscribe(topic)
	if err != nil {
		panic(err)
	}
	defer sub.Cancel()

	// Print our listening addresses
	fmt.Println("Host ID:", host.ID())
	fmt.Println("Listening on:")
	for _, addr := range host.Addrs() {
		fmt.Println("  ", addr)
	}

	// Message handler
	go func() {
		for {
			msg, err := sub.Next(context.Background())
			if err != nil {
				log.Println("Subscription closed:", err)
				return
			}
			fmt.Printf("\n[%s] %s\n> ", msg.ReceivedFrom, string(msg.Data))
		}
	}()

	// Input loop
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("> ")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to other nodes if specified
	if len(os.Args) > 1 {
		for _, addr := range os.Args[1:] {
			maddr, err := multiaddr.NewMultiaddr(addr)
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

	// Publish loop
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

	// Wait for termination signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down...")
	host.Close()
}
