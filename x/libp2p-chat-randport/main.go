package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
)

const Protocol = "/promisegrid/1.0.0"

func main() {
	// Create base context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize libp2p host with default settings
	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
	)
	if err != nil {
		panic(err)
	}
	defer node.Close()

	// Set stream handler for our protocol
	node.SetStreamHandler(Protocol, func(s network.Stream) {
		defer s.Close()
		fmt.Printf("\nReceived message from %s:\n", s.Conn().RemotePeer())
		buf := make([]byte, 1024)
		n, _ := s.Read(buf)
		fmt.Println(string(buf[:n]))
	})

	// Print node information
	fmt.Printf("Peer ID: %s\n", node.ID())
	fmt.Println("Listening addresses:")
	for _, addr := range node.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, node.ID())
	}

	// If peer address provided, connect and send message
	if len(os.Args) > 1 {
		sendMessage(ctx, node, os.Args[1])
	}

	// Wait for exit signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	fmt.Println("\nShutting down...")
}

func sendMessage(ctx context.Context, h host.Host, targetAddr string) {
	// Parse target multiaddress
	maddr, err := multiaddr.NewMultiaddr(targetAddr)
	if err != nil {
		panic(err)
	}

	// Extract peer ID from multiaddress
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		panic(err)
	}

	// Add peer to peerstore
	h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	// Connect to peer
	if err := h.Connect(ctx, *info); err != nil {
		panic(err)
	}

	// Open stream
	s, err := h.NewStream(ctx, info.ID, Protocol)
	if err != nil {
		panic(err)
	}
	defer s.Close()

	// Send message
	message := fmt.Sprintf("Hello from %s!", h.ID())
	_, err = s.Write([]byte(message))
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("Sent message to %s: %s\n", info.ID, message)
}
