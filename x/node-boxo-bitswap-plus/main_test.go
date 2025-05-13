package main

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/ipfs/go-cid"

	"github.com/libp2p/go-libp2p/core/peer"
)

func TestBitswapFetch(t *testing.T) {
	t.Run("direct", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		server, err := makeHost(0, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer server.Close()
		client, err := makeHost(0, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer client.Close()

		c, bs, err := startDataServer(ctx, server)
		if err != nil {
			t.Fatal(err)
		}
		defer bs.Close()

		expectedCid := cid.MustParse(fileCid)
		if !expectedCid.Equals(c) {
			t.Fatalf("expected CID %s, got %s", expectedCid, c)
		}
		multiaddrs, err := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{
			ID:    server.ID(),
			Addrs: server.Addrs(),
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(multiaddrs) != 1 {
			t.Fatalf("expected a single multiaddr")
		}
		out, err := runClient(ctx, client, c, multiaddrs[0].String(), false, nil)
		if err != nil {
			t.Fatal(err)
		}
		fileBytes, err := createFile0to100k()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(out, fileBytes) {
			t.Fatalf("retrieved bytes did not match sent bytes")
		}
	})

	t.Run("dht", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// Setup server and client hosts.
		server, err := makeHost(0, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer server.Close()
		client, err := makeHost(0, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer client.Close()

		// Setup DHT instances for both hosts.
		serverDht, err := setupDHT(ctx, server)
		if err != nil {
			t.Fatal(err)
		}
		defer serverDht.Close()
		clientDht, err := setupDHT(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		defer clientDht.Close()

		// Start the data server.
		c, bs, err := startDataServer(ctx, server)
		if err != nil {
			t.Fatal(err)
		}
		defer bs.Close()
		expectedCid := cid.MustParse(fileCid)
		if !expectedCid.Equals(c) {
			t.Fatalf("expected CID %s, got %s", expectedCid, c)
		}

		// Announce the file provider via the DHT.
		// Provide may take some time to propagate.
		if err := serverDht.Provide(ctx, c, true); err != nil {
			t.Fatal(err)
		}
		// Give the provider announcement a moment to propagate.
		time.Sleep(1 * time.Second)

		out, err := runClient(ctx, client, c, "", true, clientDht)
		if err != nil {
			t.Fatal(err)
		}
		fileBytes, err := createFile0to100k()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(out, fileBytes) {
			t.Fatalf("retrieved bytes did not match sent bytes")
		}
	})
}
