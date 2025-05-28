package agent2

import (
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

// Start initializes Agent2 to listen on the given port. It returns a stop
// function to gracefully shut down the agent.
func Start(port int) (func(), error) {
	reqCid, err := cid.Decode(requestProtocolStr)
	if err != nil {
		return nil, fmt.Errorf("invalid request protocol CID: %v", err)
	}

	respCid, err := cid.Decode(responseProtocolStr)
	if err != nil {
		return nil, fmt.Errorf("invalid response protocol CID: %v", err)
	}

	k := kernel.NewKernel()

	err = k.Start(port)
	if err != nil {
		return nil, fmt.Errorf("kernel start failed: %v", err)
	}

	// Subscribe to the request protocol. Upon receiving a message, send a reply
	// using the same TCP connection.
	k.Subscribe(reqCid, func(msg wire.Message) {
		fmt.Println("Agent2 received:", string(msg.Payload))
		err := k.Publish(wire.Message{
			Protocol: respCid.Bytes(),
			Payload:  []byte("hello back from agent2"),
		})
		if err != nil {
			log.Printf("Agent2 response send failed: %v", err)
		}
	})

	stop := func() {
		k.Stop()
	}

	return stop, nil
}
