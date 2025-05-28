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
	protocolStr = "bafkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44ydelk6a2mhny"
	peerAddr    = flag.String("peer", "localhost:7272", "peer address")
	listenPort  = flag.Int("port", 7271, "listen port")
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

	done := make(chan bool)

	go func() {
		// Send messages
		payload := []byte("hello world")

		timer := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-done:
				fmt.Println("Kernel stopped, exiting...")
				return
			case <-timer.C:
				// send a message every second
				err = k.Publish(wire.Message{
					Protocol: protocolCid.Bytes(),
					Payload:  payload,
				})
				if err != nil {
					log.Fatal("publish failed:", err)
				}
			}
		}
	}()

	fmt.Fprintln(os.Stderr, "Agent1 running. Press enter to exit...")
	fmt.Scanln()
	done <- true
	time.Sleep(2 * time.Second) // Give time for any pending messages to be processed
}
