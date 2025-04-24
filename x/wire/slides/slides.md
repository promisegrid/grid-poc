class: center, middle

# PromiseGrid Wire Protocol Design 

---

## Protocol Overview
### Decentralized Message Passing Architecture
- Microkernel-inspired design for sandbox orchestration
- Content-addressable network using IPFS multihash/CID standards
- Two-phase message structure: protocol hash and payload

A 'CID' (Content Identifier) is a unique identifier for content in
IPFS and compatible systems, including Bluesky.  

The protocol hash serves as a versioning mechanism, enabling future
updates without breaking existing implementations.

---

## Key Design Goals
- Future-proof extensibility through hash of protocol documents
- IoT compatibility with minimal message overhead
- Cross-platform execution (WASM, CLI, containers, bare metal)

---

## Core Protocol Components
### Message Structure
- Nested CBOR format
  - `grid` tag (0x67726964)
    - Protocol hash CID
    - Message payload

Tag 0x67726964 ("grid") will be registered with IANA.

The format of the message payload is described in the document(s)
referred to by the protocol hash CID. 

---

## Content Addressing
- CIDv1 implementation using multiformats standards:
  - Multihash (SHA2-256 default)
  - Multicodec (raw bytes)
  - Multibase encoding
- Protocol documents would typically be stored in IPFS DAGs for ease
  of retrieval by developers.

---

## Cryptographic Foundations
- Public key infrastructure integrated with CIDs
- Multihash-derived addresses for agent identification
  - TBD whether we use Bluesky-style DIDs or IPFS-style CIDs

---

## What is an agent?

An agent is defined as an entity that can:

- cryptographically sign and send messages
- receive and verify the cryptographic signatures of messages

Agents can be humans, machines, or software processes.

This definition of 'agent' is consistent with that of Promise Theory
(PT) in that "agents cannot make promises on behalf of other agents".

---

## PromiseGrid Agents are a subset of Promise Theory Agents

As PromiseGrid (PG) is a practical implementation of Promise Theory,
the engineering constraints involved in building network and security
protocols mean that we need to choose a reasonable subset of Promise
Theory's definition of an agent.  In particular:

- A simple machine or subatomic particle does not have the mechanisms
  required to cryptographically sign or verify messages.
  - So a PG agent cannot be a simple machine or subatomic particle.

This diverges from Promise Theory's definition of an agent.  From the
book "Thinking in Promises", page 9:

  Any agent (person, object, or machine) can harbour intentions. An
  intention might be something like “be red” for a [simple] light
  [...] When an intention is publicly declared to an audience (called
  its scope) it then becomes a promise.

---

## IoT Integration Strategy
### Constrained Device Support
- Example minimum viable agent requirements:
  - Aruino UNO
    - 32KB RAM
    - ability to run sha256 hash algorithm
    - ability to run ed25519 signature algorithm

---

## Decentralized IoT Standard

Is is possible that the grid could serve as a generic IoT network
fabric similar in spirit to the local I2C bus?

- Comparative analysis with MQTT/HTTP bridges
- Grid-as-backbone architecture vs traditional IoT hubs

---

## Next steps

A reasonable next step is to implement a 'hello world' equivalent; an
example of remote execution of a simple function in a CLI demo:

- sender writes a simple WASM module that returns a 'hello world' string
- sender compiles the module and stores it in IPFS, retaining the CID
  of  the module
- sender includes the CID in a message to a recipient
- recipient retrieves the module from IPFS using the CID
- recipient executes the module in a WASM runtime, seeing 'hello world'

---

## Standardization Roadmap

Register 'grid' tag with IANA.

Create protocol documents that can be stored in IPFS and referenced by
the protocol hash CID.  For example:

- Scenario tree modeling
- Personal currencies
- Advisor/Executor model
- ...

Start a draft RFC.
