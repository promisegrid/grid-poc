package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-cid"
	datastore "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	"github.com/ipfs/go-datastore/query"
	badgerds "github.com/ipfs/go-ds-badger"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiformats/go-multiaddr"
)

type safeDS struct {
	ds datastore.Batching
}

func (s *safeDS) Put(ctx context.Context, key datastore.Key, value []byte) error {
	return s.ds.Put(ctx, safeKey(key), value)
}

func (s *safeDS) Get(ctx context.Context, key datastore.Key) ([]byte, error) {
	return s.ds.Get(ctx, safeKey(key))
}

func (s *safeDS) Has(ctx context.Context, key datastore.Key) (bool, error) {
	return s.ds.Has(ctx, safeKey(key))
}

func (s *safeDS) Delete(ctx context.Context, key datastore.Key) error {
	return s.ds.Delete(ctx, safeKey(key))
}

func (s *safeDS) Query(ctx context.Context, q query.Query) (query.Results, error) {
	return s.ds.Query(ctx, q)
}

func (s *safeDS) Batch(ctx context.Context) (datastore.Batch, error) {
	return s.ds.Batch(ctx)
}

func (s *safeDS) GetSize(ctx context.Context, key datastore.Key) (int, error) {
	if getter, ok := s.ds.(interface {
		GetSize(context.Context, datastore.Key) (int, error)
	}); ok {
		return getter.GetSize(ctx, safeKey(key))
	}
	return 0, datastore.ErrNotFound
}

func (s *safeDS) Sync(ctx context.Context, key datastore.Key) error {
	if dsSync, ok := s.ds.(interface {
		Sync(context.Context, datastore.Key) error
	}); ok {
		return dsSync.Sync(ctx, safeKey(key))
	}
	return nil
}

func (s *safeDS) Close() error {
	if c, ok := s.ds.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func safeKey(k datastore.Key) datastore.Key {
	s := k.String()
	if len(s) > 0 && s[0] == '/' {
		s = s[1:]
	}
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ToUpper(s)
	allowed := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+-_="
	var sb strings.Builder
	for _, r := range s {
		if strings.ContainsRune(allowed, r) {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}
	return datastore.NewKey(sb.String())
}

func Main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repoPath := "/tmp/ipfs-lite-badger"

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		panic(err)
	}

	ds, err := badgerds.NewDatastore(repoPath, nil)
	if err != nil {
		panic(err)
	}
	defer ds.Close()

	blocksDs := ds

	metadataDs := namespace.Wrap(ds, datastore.NewKey("metadata"))

	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		panic(err)
	}

	listen, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/4005")
	if err != nil {
		panic(err)
	}

	h, dht, err := ipfslite.SetupLibp2p(
		ctx,
		priv,
		nil,
		[]multiaddr.Multiaddr{listen},
		metadataDs,
		ipfslite.Libp2pOptionsExtra...,
	)
	if err != nil {
		panic(err)
	}

	bs := blockstore.NewBlockstore(blocksDs)
	lite, err := ipfslite.New(ctx, metadataDs, bs, h, dht, nil)
	if err != nil {
		panic(err)
	}

	lite.Bootstrap(ipfslite.DefaultBootstrapPeers())

	c, err := cid.Decode("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8u")
	if err != nil {
		panic(err)
	}

	// Convert to CIDv1 for base32 encoding
	c = cid.NewCidV1(c.Type(), c.Hash())

	// we get the file a few times, timing each operation to see if caching works
	for i := 0; i < 4; i++ {
		start := time.Now()
		content, err := getContent(ctx, lite, c)
		elapsed := time.Since(start)
		if err != nil {
			panic(err)
		}
		fmt.Printf("File content:\n%s\n", string(content))
		fmt.Printf("Elapsed time: %s\n", elapsed)
	}

	/*
		if false {
			fileKey := datastore.NewKey("files-" + c.String())
			if err := blocksDs.Put(ctx, fileKey, content); err != nil {
				panic(fmt.Errorf("when putting %q: %w", fileKey.String(), err))
			}
			fmt.Println("File stored in local datastore under key", fileKey.String())
		}
	*/
}

// getcontent returns the entire content of a file
func getContent(ctx context.Context, lite *ipfslite.Peer, c cid.Cid) (content []byte, err error) {

	rsc, err := lite.GetFile(ctx, c)
	if err != nil {
		panic(err)
	}
	defer rsc.Close()

	content, err = io.ReadAll(rsc)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func main() {
	Main()
}
