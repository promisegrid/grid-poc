package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"sim1/kernel"
	"sim1/wire"

	"github.com/ipfs/go-cid"
)

var (
	// Request protocol used to send messages to agent2.
	requestProtocolStr = "bafkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44ydelk6a2mhny"
	// Response protocol expected from agent2.
	responseProtocolStr = "bafkreieq5jui4j25l3wpyw54my6fzdtcssgxhtd7wvb5klqnbawtgta5iu"
	peerAddr            = flag.String("peer", "localhost:7272", "peer address")
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
	k.SetPeer(*peerAddr)

	err = k.Start(0)
	if err != nil {
		log.Fatal("kernel start failed:", err)
	}
	defer k.Stop()

	// Subscribe to response protocol to receive replies on the same TCP
	// connection that was used to send the request.
	k.Subscribe(respCid, func(msg wire.Message) {
		fmt.Println("Agent3 received:", string(msg.Payload))
	})

	done := make(chan bool)

	go func() {
		// Send messages every second.
		payload := []byte("hello from agent3")

		timer := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-done:
				fmt.Println("Kernel stopped, exiting...")
				return
			case <-timer.C:
				err = k.Publish(wire.Message{
					Protocol: reqCid.Bytes(),
					Payload:  payload,
				})
				if err != nil {
					log.Printf("publish failed: %v", err)
				}
			}
		}
	}()

	fmt.Fprintln(os.Stderr, "Agent3 running. Press enter to exit...")
	fmt.Scanln()
	done <- true
	time.Sleep(2 * time.Second) // Allow time for processing pending messages.
}
