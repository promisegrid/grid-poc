# PromiseGrid Simulation (sim1)

A demonstration of two agents communicating via persistent TCP connections
using the PromiseGrid protocol pattern.

## Overview

This simulation shows two Go processes (agent1 and agent2) exchanging CBOR-
encoded messages through a kernel layer that manages network connections and
message routing. Key features:

- Persistent TCP connections with automatic reconnection
- CID-based protocol subscriptions
- Asynchronous message handling
- Bi-directional communication over a single TCP connection

## Components

### wire Package
- Defines `Message` struct with protocol CID and payload
- CBOR serialization/deserialization with deterministic encoding
- Custom CBOR tag (0x67726964) for message structure

### kernel Package
- Manages network connections
- Handles message routing to protocol handlers
- Provides publish/subscribe interface to agents
- Single active TCP connection maintained between peers
- Automatic connection failover and reconnection

### Agents
- agent1: Initiates conversation by dialing agent2 and sending a request
- agent2: Listens for connections from agent1, processes the request and
  sends a response using the same TCP connection established by agent1

## How It Works

1. Agents start with specified peer addresses and ports.
2. Agent1 dials agent2 and agent2 accepts the connection.
3. The kernel maintains one persistent TCP connection between the peers.
4. Message flow:
   - agent1 sends a "hello world" request every second via the outbound
     connection.
   - agent2 receives the request on the accepted TCP connection and sends a
     "hello back" response using the same connection.
   - agent1 receives the response on the connection it used to send the
     request.
   - Both agents print received messages to stdout.

## Running the Simulation

### Setup
1. In terminal 1 (agent1):
```bash
cd agent1
go run agent1.go -peer localhost:7272
```

2. In terminal 2 (agent2):
```bash
cd agent2
go run agent2.go -port 7272 
```

### Expected Output

**agent1:**
```
Agent1 running. Press enter to exit...
Agent1 received: hello back
```

**agent2:**
```
Agent2 running. Press enter to exit...
Agent2 received: hello world
```

## Notes

- Protocol CIDs are hardcoded for demonstration:
  - Request protocol: `bafkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44ydelk6a2mhny`
  - Response protocol:
    `bafkreieq5jui4j25l3wpyw54my6fzdtcssgxhtd7wvb5klqnbawtgta5iu`
- The kernel automatically reconnects if the TCP connection drops.
- Press Enter in either terminal to gracefully shut down the agent.
