# PromiseGrid Simulation (sim1)

A demonstration of three agents communicating via persistent TCP
connections using the PromiseGrid protocol pattern.

## Overview

This simulation shows three Go processes (agent1, agent2, and agent3)
exchanging CBOR-encoded messages through a kernel layer that manages
network connections and message routing. Key features:

- Persistent TCP connections with automatic reconnection
- CID-based protocol subscriptions
- Asynchronous message handling
- Bi-directional communication over multiple TCP connections maintained
  in a map for dynamic routing

## Components

### wire Package
- Defines Message struct with protocol CID and payload
- CBOR serialization/deserialization with deterministic encoding
- Custom CBOR tag (0x67726964) for message structure

### kernel Package
- Manages network connections using a map of active connections
- Handles message routing to protocol handlers
- Provides publish/subscribe interface to agents
- Multiple active TCP connections maintained between peers
- Automatic connection failover and reconnection

### Agents
- agent1: Initiates conversation by dialing agent2 and sending a request
  with the payload "hello from agent1".
- agent2: Listens for connections from agent1 and agent3, processes the
  request and sends a response "hello back from agent2" using the same TCP
  connection that was used.
- agent3: Also initiates conversation by dialing agent2 and sending a
  request with the payload "hello from agent2".

## How It Works

1. Agents start with specified peer addresses and ports.
2. Agent1 and agent3 dial agent2 while agent2 accepts incoming
   connections.
3. The kernel maintains multiple persistent TCP connections between the
   peers.
4. Message flow:
   - agent1 sends a "hello from agent1" request every second via its
     outbound connection.
   - agent3 sends a "hello from agent2" request every second via its
     outbound connection.
   - agent2 receives requests on the accepted TCP connections and sends a
     "hello back from agent2" response using the same connection that was
     used to receive the request.
   - Agents receive responses on the connection they used to send the
     request and print them to stdout.

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

3. In terminal 3 (agent3):
```bash
cd agent2
go run agent3.go -peer localhost:7272
```

### Expected Output

**agent1:**
```
Agent1 running. Press enter to exit...
Agent1 received: hello back from agent2
```

**agent2:**
```
Agent2 running. Press enter to exit...
Agent2 received: hello from agent1
Agent2 received: hello from agent3
```

**agent3:**
```
Agent3 running. Press enter to exit...
Agent3 received: hello back from agent2
```

## Notes

- Protocol CIDs are hardcoded for demonstration:
  - Request protocol:
    bafkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44ydelk6a2mhny
  - Response protocol:
    bafkreieq5jui4j25l3wpyw54my6fzdtcssgxhtd7wvb5klqnbawtgta5iu
- The kernel automatically reconnects if a TCP connection drops.
- Press Enter in any terminal to gracefully shut down the agent.
