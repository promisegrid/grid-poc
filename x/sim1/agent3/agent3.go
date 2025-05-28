package agent3

import (
	"context"
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

// Agent represents Agent3.
type Agent struct {
	k    *kernel.Kernel
	done chan struct{}
}

// NewAgent creates a new instance of Agent3 using the provided kernel.
func NewAgent(k *kernel.Kernel) *Agent {
	return &Agent{
		k:    k,
		done: make(chan struct{}),
	}
}

// Run starts Agent3, subscribing to responses and sending requests every
// second.
func (a *Agent) Run(ctx context.Context) {
	respCid, err := cid.Decode(responseProtocolStr)
	if err != nil {
		log.Printf("Agent3: invalid response protocol CID: %v", err)
		return
	}

	reqCid, err := cid.Decode(requestProtocolStr)
	if err != nil {
		log.Printf("Agent3: invalid request protocol CID: %v", err)
		return
	}

	// Subscribe to the response protocol to receive replies.
	a.k.Register(respCid, func(msg wire.Message) {
		fmt.Println("Agent3 received:", string(msg.Payload))
	})

	// Send messages every second.
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-a.done:
			fmt.Println("Agent3 stopping...")
			return
		case <-ticker.C:
			err := a.k.Send(wire.Message{
				Protocol: reqCid.Bytes(),
				Payload:  []byte("hello from agent3"),
			})
			if err != nil {
				log.Printf("Agent3 publish failed: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// Stop signals Agent3 to stop processing.
func (a *Agent) Stop() {
	close(a.done)
}
