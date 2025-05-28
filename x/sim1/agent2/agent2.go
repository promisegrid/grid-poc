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
	peerAddr    = flag.String("peer", "localhost:7271", "peer address")
	listenPort  = flag.Int("port", 7272, "listen port")
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

	// Subscribe to main protocol
	k.Subscribe(protocolCid, func(msg wire.Message) {
		fmt.Println("Agent2 received:", string(msg.Payload))

		// Send response
		respCid, _ := cid.Decode("bafkreieq5jui4j25l3wpyw54my6fzdtcssgxhtd7wvb5klqnbawtgta5iu")
		payload := []byte("hello back")
		err := k.Publish(wire.Message{
			Protocol: respCid.Bytes(),
			Payload:  payload,
		})
		if err != nil {
			log.Print("response send failed:", err)
		}
	})

	fmt.Fprintln(os.Stderr, "Agent2 running. Press enter to exit...")
	fmt.Scanln()
}
