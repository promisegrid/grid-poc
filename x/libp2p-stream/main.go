package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	resourcemanager "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/multiformats/go-multiaddr"
)

const (
	ProtocolName   = "/promisegrid/1.0.0"
	ProtocolAnswer = "ACK"
)

func main() {
	// Create resource manager with default limits
	limiter := resourcemanager.NewFixedLimiter(resourcemanager.DefaultLimits)
	rcmgr, err := resourcemanager.NewResourceManager(limiter)
	if err != nil {
		panic(err)
	}

	// Generate ECDSA P-256 key pair
	priv, _, err := crypto.GenerateECDSAKeyPair(rand.Reader)
	if err != nil {
		panic(err)
	}

	// Create libp2p host with resource management
	host, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ResourceManager(rcmgr),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/0",
			"/ip4/0.0.0.0/udp/0/quic-v1",
		),
	)
	if err != nil {
		panic(err)
	}
	defer host.Close()

	// Set stream handler for our protocol
	host.SetStreamHandler(ProtocolName, func(s network.Stream) {
		defer s.Close()
		fmt.Printf("\n[%s] New connection from %s\n", time.Now().Format("15:04:05"), s.Conn().RemotePeer())

		buf := make([]byte, 1024)
		n, err := s.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Printf("Read error: %v\n", err)
			return
		}

		msg := string(buf[:n])
		fmt.Printf("Received message: %s\n", msg)

		// Send acknowledgment
		_, err = s.Write([]byte(ProtocolAnswer))
		if err != nil {
			fmt.Printf("Write error: %v\n", err)
		}
	})

	// Print node information
	fmt.Printf("Peer ID: %s\n", host.ID())
	fmt.Println("Listening addresses:")
	for _, addr := range host.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, host.ID())
	}

	// If peer address provided, connect and send message
	if len(os.Args) > 1 {
		go connectAndSend(host, os.Args[1])
	}

	// Wait for exit signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	fmt.Println("\nShutting down...")
}

func connectAndSend(h host.Host, targetAddr string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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
	fmt.Printf("Connecting to %s...\n", info.ID)
	if err := h.Connect(ctx, *info); err != nil {
		panic(err)
	}

	// Open stream
	fmt.Printf("Opening stream to %s...\n", info.ID)
	s, err := h.NewStream(ctx, info.ID, ProtocolName)
	if err != nil {
		panic(err)
	}
	defer s.Close()

	// Send message
	message := fmt.Sprintf("Hello from %s @ %s", h.ID(), time.Now().Format("15:04:05"))
	_, err = s.Write([]byte(message))
	if err != nil {
		panic(err)
	}

	// Read response
	buf := make([]byte, len(ProtocolAnswer))
	_, err = s.Read(buf)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Received acknowledgment: %s\n", string(buf))
}
