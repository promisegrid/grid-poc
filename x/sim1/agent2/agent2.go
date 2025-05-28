package agent2

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

// Hello protocol used for both sending and receiving messages.
var helloProtocolStr = "bafkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44ydelk6a2mhny"

// Agent represents Agent2.
type Agent struct {
	k    *kernel.Kernel
	done chan struct{}
}

// NewAgent creates a new instance of Agent2 using the provided kernel.
func NewAgent(k *kernel.Kernel) *Agent {
	return &Agent{
		k:    k,
		done: make(chan struct{}),
	}
}

// Run starts Agent2, subscribing to the hello protocol, replying when a
// "hello from" message is received, and sending its own hello messages every
// second.
func (a *Agent) Run(ctx context.Context) {
	helloCid, err := cid.Decode(helloProtocolStr)
	if err != nil {
		log.Printf("Agent2: invalid hello protocol CID: %v", err)
		return
	}

	// Subscribe to the hello protocol to process incoming messages.
	a.k.Register(helloCid, func(msg wire.Message) {
		text := string(msg.Payload)
		// If the message starts with "hello from", reply with a hello back message.
		if strings.HasPrefix(text, "hello from ") {
			sender := text[len("hello from "):]
			if sender == "agent2" {
				return
			}
			reply := fmt.Sprintf("hello back from agent2 to %s", sender)
			err := a.k.Send(wire.Message{
				Protocol: helloCid.Bytes(),
				Payload:  []byte(reply),
			})
			if err != nil {
				log.Printf("Agent2 response send failed: %v", err)
			}
		} else {
			fmt.Println("Agent2 received:", text)
		}
	})

	// Send hello messages every second.
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-a.done:
			fmt.Println("Agent2 stopping...")
			return
		case <-ticker.C:
			err := a.k.Send(wire.Message{
				Protocol: helloCid.Bytes(),
				Payload:  []byte("hello from agent2"),
			})
			if err != nil {
				log.Printf("Agent2 publish failed: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// Stop signals Agent2 to stop processing.
func (a *Agent) Stop() {
	close(a.done)
}
