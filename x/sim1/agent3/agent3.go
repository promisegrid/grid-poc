package agent3

import (
	"fmt"
	"log"
	"time"

	"sim1/kernel"
	"sim1/wire"

	"github.com/ipfs/go-cid"
)

// Request protocol used to send messages to agent2.
var requestProtocolStr = "bafkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44ydelk6a2mhny"

// Response protocol expected from agent2.
var responseProtocolStr = "bafkreieq5jui4j25l3wpyw54my6fzdtcssgxhtd7wvb5klqnbawtgta5iu"

// Start initializes Agent3 with the given peer address and begins sending
// messages every second. It returns a stop function to gracefully shut down the
// agent.
func Start(peer string) (func(), error) {
	reqCid, err := cid.Decode(requestProtocolStr)
	if err != nil {
		return nil, fmt.Errorf("invalid request protocol CID: %v", err)
	}

	respCid, err := cid.Decode(responseProtocolStr)
	if err != nil {
		return nil, fmt.Errorf("invalid response protocol CID: %v", err)
	}

	k := kernel.NewKernel()
	k.SetPeer(peer)

	err = k.Start(0)
	if err != nil {
		return nil, fmt.Errorf("kernel start failed: %v", err)
	}

	// Subscribe to response protocol to receive replies.
	k.Subscribe(respCid, func(msg wire.Message) {
		fmt.Println("Agent3 received:", string(msg.Payload))
	})

	done := make(chan struct{})

	go func() {
		// Send messages every second.
		payload := []byte("hello from agent3")
		timer := time.NewTicker(1 * time.Second)
		defer timer.Stop()
		for {
			select {
			case <-done:
				fmt.Println("Agent3 stopping...")
				return
			case <-timer.C:
				err = k.Publish(wire.Message{
					Protocol: reqCid.Bytes(),
					Payload:  payload,
				})
				if err != nil {
					log.Printf("Agent3 publish failed: %v", err)
				}
			}
		}
	}()

	stop := func() {
		close(done)
		time.Sleep(2 * time.Second)
		k.Stop()
	}

	return stop, nil
}
