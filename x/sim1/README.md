# PromiseGrid Simulation (sim1)

A demonstration of three agents hosted on separate nodes that communicate via
persistent TCP connections using the PromiseGrid protocol pattern.

## Overview

This simulation shows three Go nodes (node1, node2, and node3) each hosting a
single agent (agent1, agent2, and agent3 respectively). Each node starts a
kernel that is responsible for managing network connections, routing messages,
and starting the agents. The kernel is started by the node’s main() function.
The main() function then registers an agent via the kernel.AddAgent() method.

Key features:

- Persistent TCP connections with automatic reconnection
- CID-based protocol subscriptions
- Asynchronous message handling
- Bi-directional communication over multiple TCP connections maintained in
  a map for dynamic routing

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
- Starts registered agents via the AddAgent() method

### Agents (Library Packages)
Each agent is implemented as a structure that exposes a Run() method to
start its internal processing and a Stop() method for graceful shutdown.
- agent1: Dials its peer (node2) and sends a "hello from agent1"
  request every second.
- agent2: Listens for incoming connections on its node, and upon receiving a
  request from agent1 or agent3, sends a "hello back from agent2" response
  using the same TCP connection.
- agent3: Similar to agent1, dials its peer (node2) and sends a "hello from
  agent3" request every second.

### Nodes (Executable Binaries)
Each node hosts one agent instance. The main() functions now reside in the
node packages. In each node, the following steps occur:
- A kernel instance is created and started.
- The kernel is configured for dialing or listening.
- The agent is instantiated and registered with the kernel via
  kernel.AddAgent(), which in turn starts the agent.
- Agents run asynchronously and interact via the kernel.

## How It Works

1. Nodes are started. Node1 and Node3 dial node2 while node2 accepts incoming
   connections.
2. The kernel maintains multiple persistent TCP connections between the nodes.
3. Message flow:
   - Agent1 sends a "hello from agent1" request every second via its outbound
     connection.
   - Agent3 sends a "hello from agent3" request every second via its outbound
     connection.
   - Agent2 receives requests on the accepted TCP connections and sends a
     "hello back from agent2" response using the same connection that was used
     to receive the request.
   - Agents receive responses on the connection they used to send the request
     and print them to stdout.

## Running the Simulation

### Setup
1. In terminal 1 (node1 hosting agent1):
```bash
cd node1
go run node1.go -peer localhost:7272
```

2. In terminal 2 (node2 hosting agent2):
```bash
cd node2
go run node2.go -port 7272
```

3. In terminal 3 (node3 hosting agent3):
```bash
cd node3
go run node3.go -peer localhost:7272
```

### Expected Output

**Node1 (hosting Agent1):**
```
Node1 (hosting Agent1) running. Press Ctrl+C to exit...
Agent1 received: hello back from agent2
```

**Node2 (hosting Agent2):**
```
Node2 (hosting Agent2) running. Press Ctrl+C to exit...
Agent2 received: hello from agent1
Agent2 received: hello from agent3
```

**Node3 (hosting Agent3):**
```
Node3 (hosting Agent3) running. Press Ctrl+C to exit...
Agent3 received: hello back from agent2
```

## Notes

- Protocol CIDs are hardcoded for demonstration:
  - Request protocol:
    bafkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44ydelk6a2mhny
  - Response protocol:
    bafkreieq5jui4j25l3wpyw54my6fzdtcssgxhtd7wvb5klqnbawtgta5iu
- The kernel automatically reconnects if a TCP connection drops.
- Agents are started by the kernel after being registered via the
  AddAgent() method.
- Press Ctrl+C in any terminal to gracefully shut down the node.
