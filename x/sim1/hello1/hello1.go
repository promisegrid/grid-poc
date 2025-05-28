package hello1

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"sim1/kernel"
	"sim1/wire"

	"github.com/ipfs/go-cid"
)

const helloProtocolStr = "bafkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44ydelk6a2mhny"

// Agent represents a consolidated hello1 agent.
type Agent struct {
	k         *kernel.Kernel
	agentName string
	done      chan struct{}
}

// NewAgent creates a new instance of the hello1 agent using the
// provided kernel. The agent name is passed in to personalize its messages.
func NewAgent(k *kernel.Kernel, name string) *Agent {
	return &Agent{
		k:         k,
		agentName: name,
		done:      make(chan struct{}),
	}
}

// Run starts the hello1 agent, registering the hello protocol and sending
// hello messages every second.
func (a *Agent) Run(ctx context.Context) {
	helloCid, err := cid.Decode(helloProtocolStr)
	if err != nil {
		log.Printf("Agent %s: invalid hello protocol CID: %v",
			a.agentName, err)
		return
	}

	// Register the hello protocol to receive and respond to messages.
	a.k.Register(helloCid, func(msg wire.Message) {
		text := string(msg.Payload)
		fmt.Printf("Agent %s received: %s\n", a.agentName, text)
		if strings.HasPrefix(text, "hello from ") {
			// Extract sender name from the message.
			sender := text[len("hello from "):]
			// Avoid replying to self.
			if sender == a.agentName {
				return
			}
			reply := fmt.Sprintf("hello back from %s to %s",
				a.agentName, sender)
			err := a.k.Send(wire.Message{
				Protocol: helloCid.Bytes(),
				Payload:  []byte(reply),
			})
			if err != nil {
				log.Printf("Agent %s response send failed: %v",
					a.agentName, err)
			}
		}
	})

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-a.done:
			fmt.Printf("Agent %s stopping...\n", a.agentName)
			return
		case <-ticker.C:
			err := a.k.Send(wire.Message{
				Protocol: helloCid.Bytes(),
				Payload: []byte(fmt.Sprintf("hello from %s",
					a.agentName)),
			})
			if err != nil {
				log.Printf("Agent %s publish failed: %v",
					a.agentName, err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// Stop signals the hello1 agent to stop processing.
func (a *Agent) Stop() {
	close(a.done)
}
