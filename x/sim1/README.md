# PromiseGrid Simulation (sim1)

A demonstration of three agents hosted on separate nodes that communicate via
persistent TCP connections using the PromiseGrid protocol pattern.

## Overview

This simulation shows three Go nodes (node1, node2, and node3) each hosting a
hello1 agent. Each node starts a kernel that is responsible for managing
network connections, routing messages, and starting the agents. The kernel is
started by the node’s main() function. The main() function then registers an
agent via the kernel.AddAgent() method. The hello1 agent is a consolidated
implementation that accepts an agent name as a parameter, allowing multiple
nodes to use the same code.

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

### Agents (Library Package)
A single consolidated agent implementation, hello1, is used by all nodes.
The agent name is passed as an argument to the NewAgent function. It implements a
simple hello protocol:

- Every 1 second, the agent sends a "hello from <agent name>" message.
- Upon receiving a message beginning with "hello from", an agent extracts the
  sender name and replies using the same connection with a message formatted as
  "hello back from <agent name> to <sender name>".
- If a message does not match the "hello from" format, the message is simply
  printed to stdout.

### Nodes (Executable Binaries)
Each node hosts one hello1 agent instance. The main() functions reside in
the node packages. In each node, the following steps occur:
- A kernel instance is created and started.
- The kernel is configured for dialing or listening.
- The hello1 agent is instantiated with a unique agent name (for example,
  "agent1", "agent2", or "agent3") and registered with the kernel via
  kernel.AddAgent(), which in turn starts the agent.
- Agents run asynchronously and interact via the kernel.

## How It Works

1. Nodes are started. Node1 and Node3 dial node2 while node2 accepts incoming
   connections.
2. The kernel maintains multiple persistent TCP connections between the nodes.
3. Message flow:
   - Each hello1 agent sends a "hello from <agent name>" message every second
     via its outbound connection.
   - Upon receiving a "hello from" message, the agent replies with a message
     "hello back from <agent name> to <sender name>" using the same connection.
   - Agents print to stdout any message that does not trigger a reply.

## Running the Simulation

### Setup
1. In terminal 1 (node1 hosting hello1 with agent name "agent1"):
```bash
cd node1
go run node1.go -peer localhost:7272 -name agent1
```

2. In terminal 2 (node2 hosting hello1 with agent name "agent2"):
```bash
cd node2
go run node2.go -port 7272 -name agent2
```

3. In terminal 3 (node3 hosting hello1 with agent name "agent3"):
```bash
cd node3
go run node3.go -peer localhost:7272 -name agent3
```

### Expected Output

**Node1 (hosting agent "agent1"):**
```
Node1 (hosting hello1 agent with name agent1) running. Press Ctrl+C to
exit...
Agent agent1 received: hello back from <agent_name> to agent1
```

**Node2 (hosting agent "agent2"):**
```
Node2 (hosting hello1 agent with name agent2) running. Press Ctrl+C to
exit...
Agent agent2 received: hello from agent1
Agent agent2 received: hello from agent3
```

**Node3 (hosting agent "agent3"):**
```
Node3 (hosting hello1 agent with name agent3) running. Press Ctrl+C to
exit...
Agent agent3 received: hello back from <agent_name> to agent3
```

## Notes

- A single protocol CID is used for the hello pub/sub protocol throughout:
  - Hello protocol:
    bafkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44ydelk6a2mhny
- The kernel automatically reconnects if a TCP connection drops.
- The hello1 agent is started by the kernel after being registered via the
  AddAgent() method.
- Press Ctrl+C in any terminal to gracefully shut down the node.
