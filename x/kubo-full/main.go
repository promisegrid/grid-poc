package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ipfs/kubo/config"
	"github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/core/coreapi"
	"github.com/ipfs/kubo/core/corehttp"
	"github.com/ipfs/kubo/core/node/libp2p"
	"github.com/ipfs/kubo/plugin/loader"
	"github.com/ipfs/kubo/repo/fsrepo"
)

func main() {
	// Set up context and signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize configuration
	cfg, err := config.Init(os.Stdout, 2048)
	if err != nil {
		panic(fmt.Errorf("failed creating config: %w", err))
	}

	// Set API and Gateway addresses
	cfg.Addresses.API = []string{"/ip4/127.0.0.1/tcp/5001"}
	cfg.Addresses.Gateway = []string{"/ip4/0.0.0.0/tcp/8080"}

	// Initialize plugins
	plugins, err := loader.NewPluginLoader("")
	if err != nil {
		panic(fmt.Errorf("failed loading plugins: %w", err))
	}

	if err := plugins.Initialize(); err != nil {
		panic(fmt.Errorf("failed initializing plugins: %w", err))
	}

	if err := plugins.Inject(); err != nil {
		panic(fmt.Errorf("failed injecting plugins: %w", err))
	}

	repoPath := "/tmp/.ipfs" // Change for custom repo location
	if !fsrepo.IsInitialized(repoPath) {
		// Set up repository
		if err := fsrepo.Init(repoPath, cfg); err != nil {
			panic(fmt.Errorf("failed initing ipfs repo: %w", err))
		}
	}

	// Open the repository
	repo, err := fsrepo.Open(repoPath)
	if err != nil {
		panic(fmt.Errorf("failed opening repo: %w", err))
	}

	// Create full node configuration
	node, err := core.NewNode(ctx, &core.BuildCfg{
		Online:    true,
		Permanent: true,
		Routing:   libp2p.DHTOption,
		Repo:      repo,
		ExtraOpts: map[string]bool{
			"pubsub": true, // Enable pubsub
			"ipnsps": true, // Enable IPNS over pubsub
			"mplex":  true, // Enable mplex stream muxer
		},
	})
	if err != nil {
		panic(fmt.Errorf("failed creating node: %w", err))
	}

	// Get CoreAPI instance
	api, err := coreapi.NewCoreAPI(node)
	if err != nil {
		panic(fmt.Errorf("failed getting coreapi: %w", err))
	}

	// Start HTTP API server
	err = corehttp.ListenAndServe(node, "/ip4/127.0.0.1/tcp/5001", corehttp.CommandsOption(ctx))
	if err != nil {
		panic(fmt.Errorf("failed starting API server: %w", err))
	}

	// Print node information
	id, err := api.Key().Self(ctx)
	if err != nil {
		panic(fmt.Errorf("failed getting node ID: %w", err))
	}
	fmt.Printf("IPFS node %s is running\n", id.ID())

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until shutdown signal
	select {
	case <-sigChan:
		fmt.Println("\nShutting down node...")
	case <-ctx.Done():
	}

	// Close the node
	if err := node.Close(); err != nil {
		panic(fmt.Errorf("failed closing node: %w", err))
	}
}
