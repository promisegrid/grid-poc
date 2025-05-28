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
- Single protocol CID used for both publishing and subscribing
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
All agents use the same protocol CID for both sending and receiving messages,
implementing a simple hello protocol:

- Every 1 second, each agent sends a "hello from <my name>" message.
- Upon receiving a message beginning with "hello from", an agent extracts the
  sender name and replies using the same connection with a message formatted as
  "hello back from <my name> to <sender name>".
- If a message does not match the "hello from" format (for example, a hello back
  reply), the message is simply printed to stdout.

- agent1: Dials a peer node and sends hello messages every second.
- agent2: Listens for incoming connections and sends a reply upon receiving a
  hello request; it also sends its own hello messages every second.
- agent3: Similar to agent1, dials its peer and sends hello messages every second.

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
   - Each agent sends a "hello from <agent_name>" message every second via its
     outbound connection.
   - Upon receiving a "hello from" message, the agent replies with a message
     "hello back from <agent_name> to <sender name>" using the same connection.
   - Agents print to stdout any message that does not trigger a reply.

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
Agent1 received: hello back from agentX to agent1
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
Agent3 received: hello back from agentX to agent3
```

## Notes

- A single protocol CID is used for the hello pub/sub protocol throughout:
  - Hello protocol:
    bafkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44ydelk6a2mhny
- The kernel automatically reconnects if a TCP connection drops.
- Agents are started by the kernel after being registered via the
  AddAgent() method.
- Press Ctrl+C in any terminal to gracefully shut down the node.
