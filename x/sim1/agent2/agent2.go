package agent2

import (
	"context"
	"fmt"
	"log"

	"sim1/kernel"
	"sim1/wire"

	"github.com/ipfs/go-cid"
)

// Request protocol expected on incoming messages from agent1 and agent3.
var requestProtocolStr = "bafkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44ydelk6a2mhny"

// Response protocol used to reply to agent1 and agent3.
var responseProtocolStr = "bafkreieq5jui4j25l3wpyw54my6fzdtcssgxhtd7wvb5klqnbawtgta5iu"

// Agent represents Agent2.
type Agent struct {
	k *kernel.Kernel
}

// NewAgent creates a new instance of Agent2 using the provided kernel.
func NewAgent(k *kernel.Kernel) *Agent {
	return &Agent{
		k: k,
	}
}

// Run starts Agent2, subscribing to incoming requests and replying using the
// same connection.
func (a *Agent) Run(ctx context.Context) {
	reqCid, err := cid.Decode(requestProtocolStr)
	if err != nil {
		log.Printf("Agent2: invalid request protocol CID: %v", err)
		return
	}

	respCid, err := cid.Decode(responseProtocolStr)
	if err != nil {
		log.Printf("Agent2: invalid response protocol CID: %v", err)
		return
	}

	// Subscribe to the request protocol. Upon receiving a message, send a reply.
	a.k.Subscribe(reqCid, func(msg wire.Message) {
		fmt.Println("Agent2 received:", string(msg.Payload))
		err := a.k.Publish(wire.Message{
			Protocol: respCid.Bytes(),
			Payload:  []byte("hello back from agent2"),
		})
		if err != nil {
			log.Printf("Agent2 response send failed: %v", err)
		}
	})

	// Block until context is cancelled.
	<-ctx.Done()
}

// Stop stops Agent2. No additional cleanup is required.
func (a *Agent) Stop() {
	// No op.
}
