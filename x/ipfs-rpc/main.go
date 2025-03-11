package main

import (
	"context"
	"fmt"

	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/kubo/client/rpc"
)

func main() {
	// "Connect" to local node
	node, err := rpc.NewLocalApi()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Pin a given file by its CID
	ctx := context.Background()
	c, err := cid.Decode("bafkreidtuosuw37f5xmn65b3ksdiikajy7pwjjslzj2lxxz2vc4wdy3zku")
	if err != nil {
		fmt.Println(err)
		return
	}
	p := path.FromCid(c)
	err = node.Pin().Add(ctx, p)
	if err != nil {
		fmt.Println(err)
		return
	}

	// publish a message to a pubsub topic
	err = node.PubSub().Publish(ctx, "t7a", []byte("hello big world!\n"))
	if err != nil {
		fmt.Println(err)
		return
	}
}
