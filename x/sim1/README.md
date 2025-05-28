# PromiseGrid Simulation (sim1)

A demonstration of two agents communicating via persistent TCP
connections using the PromiseGrid protocol pattern.

## Overview

This simulation shows two Go processes (agent1 and agent2) exchanging
CBOR-encoded messages through a kernel layer that manages network
connections and message routing. Key features:

- Persistent TCP connections with automatic reconnection
- CID-based protocol subscriptions
- Asynchronous message handling
- Bi-directional communication

## Components

### wire Package
- Defines `Message` struct with Protocol CID and Payload
- CBOR serialization/deserialization with deterministic encoding
- Custom CBOR tag (0x67726964) for message structure

### kernel Package
- Manages network connections
- Handles message routing to protocol handlers
- Provides publish/subscribe interface to agents
- Automatic connection maintenance

### Agents
- agent1: Initiates conversation, listens on port 7271
- agent2: Responds to messages, listens on port 7272

## How It Works

1. Agents start with specified peer addresses and ports
2. Kernel maintains persistent TCP connection between peers
3. Message flow:
   - agent1 sends "hello world" every second
   - agent2 receives message and sends "hello back" response
   - Both agents print received messages to stdout

## Running the Simulation

### Setup
1. Build both agents:
```bash
go build -o agent1 ./agent1
go build -o agent2 ./agent2
```

2. In terminal 1 (agent1):
```bash
./agent1 -port 7271 -peer localhost:7272
```

3. In terminal 2 (agent2):
```bash
./agent2 -port 7272 -peer localhost:7271
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
  - Main protocol: `bafkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44ydelk6a2mhny`
  - Response protocol: `bafkreieq5jui4j25l3wpyw54my6fzdtcssgxhtd7wvb5klqnbawtgta5iu`
- The kernel automatically reconnects if the TCP connection drops
- Press Enter in either terminal to gracefully shut down the agent
