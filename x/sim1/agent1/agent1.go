package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"sim1/kernel"
	"sim1/wire"

	"github.com/ipfs/go-cid"
)

var (
	protocolStr = "bafkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44ydelk6a2mhny"
	peerAddr    = flag.String("peer", "localhost:8081", "peer address")
	listenPort  = flag.Int("port", 8080, "listen port")
)

func main() {
	flag.Parse()

	protocolCid, err := cid.Decode(protocolStr)
	if err != nil {
		log.Fatal("invalid protocol CID:", err)
	}

	k := kernel.NewKernel()
	k.SetPeer(*peerAddr)

	err = k.Start(*listenPort)
	if err != nil {
		log.Fatal("kernel start failed:", err)
	}
	defer k.Stop()

	// Subscribe to response protocol
	respCid, _ := cid.Decode("bafkreieq5jui4j25l3wpyw54my6fzdtcssgxhtd7wvb5klqnbawtgta5iu")
	k.Subscribe(respCid, func(msg wire.Message) {
		fmt.Println("Agent1 received:", string(msg.Payload))
	})

	// Send initial message
	payload := []byte("hello world")
	msg, err := wire.NewMessage(protocolCid, payload)
	if err != nil {
		log.Fatal("create message failed:", err)
	}

	err = k.Publish(protocolCid, wire.Message{
		Protocol: protocolCid.Bytes(),
		Payload:  payload,
	})
	if err != nil {
		log.Fatal("publish failed:", err)
	}

	fmt.Fprintln(os.Stderr, "Agent1 running. Press enter to exit...")
	fmt.Scanln()
}
