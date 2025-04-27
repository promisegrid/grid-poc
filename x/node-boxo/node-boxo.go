package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ipfs/boxo/bitswap"
	bsnet "github.com/ipfs/boxo/bitswap/network"
	"github.com/ipfs/boxo/blockstore"
	"github.com/ipfs/boxo/exchange"
	"github.com/ipfs/boxo/namesys"
	"github.com/ipfs/boxo/provider"
	config "github.com/ipfs/kubo/config"
	"github.com/ipfs/kubo/repo"
	flatfs "github.com/ipfs/go-ds-flatfs"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multihash"
)

const repoPath = "/tmp/.ipfs-boxo"

type BoxoNode struct {
	Host          host.Host
	DHT           *dual.DHT
	PubSub        *pubsub.PubSub
	Blockstore    blockstore.Blockstore
	Bitswap       exchange.Interface
	IPNSPublisher *namesys.Publisher
	Repo          repo.Repo
	ProviderSys   provider.System
}

func setupRepo(ctx context.Context, path string) (repo.Repo, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("creating repo directory: %w", err)
	}

	cfg, err := config.Init(os.Stdout, 2048)
	if err != nil {
		return nil, fmt.Errorf("initializing config: %w", err)
	}

	cfg.Datastore.Spec = map[string]interface{}{
		"type":      "flatfs",
		"path":      "blocks",
		"sync":      true,
		"shardFunc": "/repo/flatfs/shard/v1/next-to-last/2",
	}

	shardFuncStr, ok := cfg.Datastore.Spec["shardFunc"].(string)
	if !ok {
		return nil, fmt.Errorf("shardFunc is not a string")
	}

	shardFunc, err := flatfs.ParseShardFunc(shardFuncStr)
	if err != nil {
		return nil, fmt.Errorf("parsing shard function: %w", err)
	}

	dsPath := filepath.Join(path, "blocks")
	ds, err := flatfs.CreateOrOpen(dsPath, shardFunc, true)
	if err != nil {
		return nil, fmt.Errorf("creating flatfs datastore: %w", err)
	}

	return &repo.Mock{
		D: ds,
		C: *cfg,
	}, nil
}

func NewBoxoNode(ctx context.Context) (*BoxoNode, error) {
	repo, err := setupRepo(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("repo setup failed: %w", err)
	}

	hst, err := libp2p.New(
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/4001",
			"/ip6/::/tcp/4001",
		),
		libp2p.NATPortMap(),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			return dual.New(ctx, h,
				dual.DHTOption(
					dht.Datastore(repo.Datastore()),
				),
				dual.DHTMode(dht.ModeAutoServer),
			)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("creating host: %w", err)
	}

	dht, ok := hst.(routing.Routing).(*dual.DHT)
	if !ok {
		return nil, fmt.Errorf("could not cast routing to dual.DHT")
	}

	ps, err := pubsub.NewGossipSub(ctx, hst)
	if err != nil {
		return nil, fmt.Errorf("creating pubsub: %w", err)
	}

	bs := blockstore.NewBlockstore(repo.Datastore())
	network := bsnet.NewFromIpfsHost(hst, dht)
	bswap := bitswap.New(ctx, network, bs,
		bitswap.EngineBlockstoreWorkerCount(3),
	)

	ipnsPublisher, err := namesys.NewPublisher(dht, repo.Datastore())
	if err != nil {
		return nil, fmt.Errorf("creating IPNS publisher: %w", err)
	}

	providerSys := provider.NewSystem(
		provider.NewProvider(bs),
		provider.Reprovider(time.Hour),
		provider.Online(dht),
	)

	return &BoxoNode{
		Host:          hst,
		DHT:           dht,
		PubSub:        ps,
		Blockstore:    bs,
		Bitswap:       bswap,
		IPNSPublisher: ipnsPublisher,
		Repo:          repo,
		ProviderSys:   providerSys,
	}, nil
}

func (n *BoxoNode) Start(ctx context.Context) error {
	if err := n.DHT.Bootstrap(ctx); err != nil {
		return fmt.Errorf("dht bootstrap failed: %w", err)
	}

	cfg, _ := config.Init(os.Stdout, 2048)
	bootstrapPeers, err := cfg.BootstrapPeers()
	if err != nil {
		return fmt.Errorf("getting bootstrap peers: %w", err)
	}

	for _, p := range bootstrapPeers {
		if err := n.Host.Connect(ctx, peer.AddrInfo{ID: p.ID}); err != nil {
			fmt.Printf("Failed to connect to bootstrap peer %s: %v\n", p.ID, err)
		}
	}

	if err := n.ProviderSys.Run(); err != nil {
		return fmt.Errorf("starting provider system: %w", err)
	}

	return nil
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

	go func() {
		time.Sleep(5 * time.Second)
		privKey := node.Host.Peerstore().PrivKey(node.Host.ID())

		mh, _ := multihash.Sum([]byte("example"), multihash.SHA2_256, -1)
		exampleCid := cid.NewCidV1(cid.Raw, mh)
		expiration := time.Now().Add(24 * time.Hour)

		err := node.IPNSPublisher.Publish(ctx, privKey, "/ipfs/"+exampleCid.String(), 0, time.Hour, namesys.WithEOL(expiration))
		if err != nil {
			fmt.Printf("IPNS publication failed: %v\n", err)
		} else {
			fmt.Println("Successfully published IPNS record")
		}
	}()

	select {}
}
