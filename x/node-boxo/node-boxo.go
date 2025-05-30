package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ipfs/boxo/bitswap/network/bsnet"
	bsserver "github.com/ipfs/boxo/bitswap/server"
	"github.com/ipfs/boxo/blockstore"
	"github.com/ipfs/boxo/ipns"
	"github.com/ipfs/boxo/provider"
	config "github.com/ipfs/go-ipfs-config"
	ds "github.com/ipfs/go-datastore"
	flatfs "github.com/ipfs/go-ds-flatfs"
	repo "github.com/ipfs/kubo/repo"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	dual "github.com/libp2p/go-libp2p-kad-dht/dual"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const repoPath = "~/.ipfs-boxo"

type BoxoNode struct {
	Host          host.Host
	DHT           *dual.DHT
	PubSub        *pubsub.PubSub
	Blockstore    blockstore.Blockstore
	Bitswap       *bsserver.Server
	IPNSPublisher *ipns.Publisher
}

func expandPath(p string) (string, error) {
	if len(p) > 0 && p[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, p[1:]), nil
	}
	return p, nil
}

func setupRepo(ctx context.Context, p string) (repo.Repo, error) {
	expandedPath, err := expandPath(p)
	if err != nil {
		return nil, fmt.Errorf("expanding repo path: %w", err)
	}
	if err := os.MkdirAll(expandedPath, 0755); err != nil {
		return nil, fmt.Errorf("creating repo directory: %w", err)
	}

	// Initialize configuration
	cfg, err := config.Init(os.Stdout, 2048)
	if err != nil {
		return nil, fmt.Errorf("initializing config: %w", err)
	}

	// Configure flatfs datastore
	cfg.Datastore.Spec = map[string]interface{}{
		"type":      "flatfs",
		"path":      "blocks",
		"sync":      true,
		"shardFunc": "/repo/flatfs/shard/v1/next-to-last/2",
	}

	// Open or create the flatfs datastore
	d, err := flatfs.CreateOrOpen(
		flatfs.ParseShardFunc(cfg.Datastore.Spec["shardFunc"].(string)),
		filepath.Join(expandedPath, "blocks"),
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("opening datastore: %w", err)
	}

	return &repo.Mock{
		D: d,
		C: *cfg,
	}, nil
}

func NewBoxoNode(ctx context.Context) (*BoxoNode, error) {
	// Initialize repository with disk storage
	r, err := setupRepo(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("repo setup failed: %w", err)
	}
	// Since we are using repo.Mock, extract the datastore
	dstore := r.(*repo.Mock).D

	// Create libp2p host with recommended options
	hst, err := libp2p.New(
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/4001",
			"/ip6/::/tcp/4001",
		),
		libp2p.NATPortMap(),
	)
	if err != nil {
		return nil, fmt.Errorf("creating host: %w", err)
	}

	// Initialize DHT in server mode
	dht, err := dual.New(ctx, hst, dual.DHTOption(
		dual.DHTMode(dual.ModeAuto),
	))
	if err != nil {
		return nil, fmt.Errorf("creating DHT: %w", err)
	}

	// Initialize Gossipsub router
	ps, err := pubsub.NewGossipSub(ctx, hst)
	if err != nil {
		return nil, fmt.Errorf("creating pubsub: %w", err)
	}

	// Initialize blockstore and Bitswap server
	bs := blockstore.NewBlockstore(dstore)
	network := bsnet.NewFromIpfsHost(hst)
	bswap := bsserver.New(ctx, network, bs)
	network.Start(bswap)

	// Retrieve the host's private key for IPNS
	priv, ok := hst.Peerstore().PrivKey(hst.ID())
	if !ok {
		return nil, fmt.Errorf("no private key found for host")
	}

	// Set up IPNS publisher using the DHT as the routing interface.
	ipnsPublisher := ipns.NewPublisher(dht, dstore, priv)

	return &BoxoNode{
		Host:          hst,
		DHT:           dht,
		PubSub:        ps,
		Blockstore:    bs,
		Bitswap:       bswap,
		IPNSPublisher: ipnsPublisher,
	}, nil
}

func (n *BoxoNode) Start(ctx context.Context) error {
	// Bootstrap DHT
	if err := n.DHT.Bootstrap(ctx); err != nil {
		return fmt.Errorf("dht bootstrap failed: %w", err)
	}

	// Connect to IPFS bootstrap peers using repo configuration.
	cfg := n.getRepoConfig()
	peers, err := cfg.BootstrapPeers()
	if err != nil {
		return fmt.Errorf("getting bootstrap peers: %w", err)
	}

	for _, p := range peers {
		if err := n.Host.Connect(ctx, p); err != nil {
			fmt.Printf("Failed to connect to bootstrap peer %s: %v\n", p.ID(), err)
		}
	}

	// Start content provider
	provider.NewProvider(n.Host, n.DHT, n.Blockstore)
	return nil
}

// getRepoConfig attempts to retrieve the repository configuration.
func (n *BoxoNode) getRepoConfig() *config.Config {
	// For this example, we assume the config is stored in the provider's repo.
	// In a complete implementation, you would store and retrieve the repo config.
	// Here we simply initialize a new config.
	cfg, _ := config.Init(os.Stdout, 2048)
	return cfg
}

func (n *BoxoNode) StartIPNSPublication(ctx context.Context) {
	// Example IPNS publication
	go func() {
		time.Sleep(5 * time.Second) // Wait for node initialization
		value := []byte("/ipfs/QmExampleContentHash")
		// Publish with 24h validity
		expiration := time.Now().Add(24 * time.Hour)
		priv, ok := n.Host.Peerstore().PrivKey(n.Host.ID())
		if !ok {
			fmt.Println("IPNS publication failed: no private key")
			return
		}
		err := n.IPNSPublisher.Publish(ctx, priv, value,
			ipns.WithEOL(expiration))
		if err != nil {
			fmt.Printf("IPNS publication failed: %v\n", err)
		} else {
			fmt.Println("Successfully published IPNS record")
		}
	}()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := NewBoxoNode(ctx)
	if err != nil {
		panic(fmt.Errorf("node creation failed: %w", err))
	}

	if err := node.Start(ctx); err != nil {
		panic(fmt.Errorf("node start failed: %w", err))
	}
	defer node.Host.Close()

	fmt.Printf("IPFS node running with ID %s\n", node.Host.ID())
	fmt.Printf("Listening on addresses:\n")
	for _, addr := range node.Host.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, node.Host.ID())
	}

	// Start IPNS publication example
	node.StartIPNSPublication(ctx)

	// Keep the node running
	select {}
}
