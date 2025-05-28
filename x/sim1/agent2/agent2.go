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
	// Request protocol expected on incoming messages from agent1.
	requestProtocolStr = "bafkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44ydelk6a2mhny"
	// Response protocol used to reply to agent1.
	responseProtocolStr = "bafkreieq5jui4j25l3wpyw54my6fzdtcssgxhtd7wvb5klqnbawtgta5iu"
	listenPort          = flag.Int("port", 7272, "listen port")
)

func main() {
	flag.Parse()

	reqCid, err := cid.Decode(requestProtocolStr)
	if err != nil {
		log.Fatal("invalid request protocol CID:", err)
	}

	respCid, err := cid.Decode(responseProtocolStr)
	if err != nil {
		log.Fatal("invalid response protocol CID:", err)
	}

	k := kernel.NewKernel()

	err = k.Start(*listenPort)
	if err != nil {
		log.Fatal("kernel start failed:", err)
	}
	defer k.Stop()

	// Subscribe to the request protocol. When a message is received from
	// agent1, send back a response using the same TCP connection.
	k.Subscribe(reqCid, func(msg wire.Message) {
		fmt.Println("Agent2 received:", string(msg.Payload))

		err := k.Publish(wire.Message{
			Protocol: respCid.Bytes(),
			Payload:  []byte("hello back"),
		})
		if err != nil {
			log.Print("response send failed:", err)
		}
	})

	fmt.Fprintln(os.Stderr, "Agent2 running. Press enter to exit...")
	fmt.Scanln()
}
