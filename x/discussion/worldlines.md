# Worldline Completion Model in Messaging Systems

The worldline completion model offers an innovative approach to managing events in a messaging system by leveraging a Merkle DAG structure to create a complete, verifiable history of events. Every message is a Directed Acyclic Graph (DAG) edit operation—a promise (in Promise Theory style as described by Mark Burgess) that asserts the validity of the intended update to one or more worldline(s). This document explains the worldline data structure, its implementation considerations, and provides examples of DAG edit, query, and subscription messages expressed as CWT claims.

> Note on Promise Theory:  
> - Promises are Burgess-style: an agent can only make promises concerning its own behavior, never that of another agent.  
> - A promise is an assertion that the corresponding DAG edit operation is valid; it is not a request for action.  

---

## Overview of the Worldline Completion Model

Key aspects of the model include:

- **Multiple Worldlines with Branching and Merging:**  
  The architecture supports multiple concurrent worldlines that may branch or merge. A single event can contribute to several worldlines simultaneously.


- **Messages as DAG Edit Operations and Promises:**  
  Every message is both an instruction to modify the DAG and a promise. Each promise is expressed as one or more structured CWT claims. Importantly:
  - An agent promises only about its own behavior.
  - Promises assert that the corresponding DAG edit operation is valid within the current state of the system.
  - A promise is not a command but a self-contained assertion, forming the basis for auditability and trust.

- **Merkle DAG Architecture:**  
  - **Events in Leaves:**  
    Events, recorded as leaves in the DAG, represent immutable historical records. They do not include direct links to previous events, allowing flexibility in insertion, deletion, or reordering within different worldlines.
    
  - **Internal Nodes as Edit Logs:**  
    Internal nodes act as logs of messages (edit operations) and store cryptographic hashes of their child nodes, ensuring tamper resistance. Each message bundles one or more CWT claims that detail specific edits.
  
- **Replayability and Integrity:**  
  Each edit operation includes sufficient contextual data (like previous internal node hashes) to enable the entire message history to be replayed, ensuring that the DAG remains a verifiable and trustworthy representation of one or more worldlines.

- **Worldline Scoping:**  
  The model is sufficiently flexible to be implemented as either a single unified DAG representing all known worldlines or as distinct DAGs (one per worldline) that may eventually merge. This design decision depends on scalability, performance, and operational requirements.

- **Handling Inaccessible Nodes:**  
  The system anticipates situations where not all referenced leaves or inner nodes are currently accessible. By including prior internal node hashes within messages, the system can later validate and reconstruct the complete historical context when missing components become available.

- **CWT Claims as Containers for Promises:**  
  A CWT is used as a container that holds one or more promises expressed as claims. It does not represent a promise by itself but groups related promises, allowing compound operations to be securely bundled and verified.

---

## Relationships between Messages, DAG Nodes, Promises, and Events

- **Messages as Edit Operations and Promises:**  
  Each message instructs a modification to the DAG and simultaneously acts as a promise. The agent’s assertion within the message (expressed via CWT claims) commits to the validity of the change. By design, agents can only promise actions within their control.

- **DAG Nodes – Leaves and Internal Nodes:**  
  - **Leaf Nodes (Events):**  
    Capture finalized events as immutable records. Once inserted, these events serve as the cryptographic foundation of the historical record.
  
  - **Internal Nodes (Edit Logs):**  
    Maintain ordered logs of promise messages. They contain hashes of child nodes to ensure that the history remains tamper-proof and supports complete replayability.

- **Promises via CWT Claims:**  
  The promises are expressed within CWT claims, each declaring a specific edit operation. This mechanism ensures that all promised operations can later be independently validated for consistency and integrity.

- **Immutable Events:**  
  Once a message is applied, its effects become part of the immutable event record (leaf node), preserving the historical state against tampering.

- **Assuring a Valid Worldline Representation:**  
  The combination of promise-based edit messages and immutable event records ensures that the DAG faithfully represents the state of all worldlines at any given point. This design safeguards the verifiability and auditability of the entire system.

---

## Alternative Structural Considerations

While this document describes a Merkle DAG architecture as the basis for the worldline completion model, it is important to note that:

- The structure might evolve into something new and unconventional as requirements become clearer and system demands grow.
- Alternative data structures or hybrid approaches could offer improved performance or flexibility compared to a standard Merkle DAG, especially in dynamic, distributed environments.

---

## Worldline Data Structure Example in Go

A representative Go structure for an event within a worldline is provided below:

```go
type WorldlineEvent struct {
    Timestamp time.Time // The creating agent’s local timestamp.
    Agent     string    // Identifier of the agent (e.g., "Alice" or "Bob").
    Payload   []byte    // Contains one or more promise claims expressed as CWT claims.
    Signature []byte    // Digital signature validating the agent's promise.
}
```

### Merkle DAG Essentials

- **Leaf Nodes (Events):**  
  Contain event details such as payload, timestamp, and signature without directly linking to prior events.

- **Internal Nodes (Edit Logs):**  
  Record sequences of messages (edit operations) by storing ordered child node hashes. This approach mirrors version control systems (such as Git) to ensure tamper resistance and replayability.

---

## Example Scenarios

### 1. DAG Edit Message Examples

Edit messages act as promises to modify one or more worldlines.

#### a. Insert Operation

Alice sends a message to insert a new event at a specific position. The message includes previous internal node hashes to supply necessary context.

**Example JSON:**

```json
{
  "op": "insert",
  "agent": "Alice",
  "timestamp": "2023-10-01T10:05:00Z",
  "target": "worldline123",
  "payload": {
    "claims": [
      {"description": "Insert event 'Hello' as an initial greeting", "detail": "Welcome message"}
    ]
  },
  "prevHashes": ["hash_internalNode1", "hash_internalNode2"],
  "signature": "AliceSignatureABC123"
}
```

#### b. Delete Operation

Bob issues a deletion message to remove an outdated event. The message contains context for verifying the deletion.

**Example JSON:**

```json
{
  "op": "delete",
  "agent": "Bob",
  "timestamp": "2023-10-01T10:20:00Z",
  "target": "worldline123",
  "payload": {
    "claims": [
      {"description": "Delete event 'Obsolete'", "detail": "Removing outdated information"}
    ]
  },
  "prevHashes": ["hash_internalNode3"],
  "signature": "BobSignatureXYZ789"
}
```

#### c. Reorder Operation

Alice proposes a reorder of events within a worldline. The new order is verifiable through the inclusion of previous internal node hashes.

**Example JSON:**

```json
{
  "op": "reorder",
  "agent": "Alice",
  "timestamp": "2023-10-01T10:30:00Z",
  "target": "worldline123",
  "payload": {
    "claims": [
      {"description": "Reorder events to prioritize recent updates", "detail": "Moving update event up"}
    ]
  },
  "prevHashes": ["hash_internalNode4"],
  "signature": "AliceSignatureDEF456"
}
```

---

### 2. DAG Query Message Example

A query message promises the retrieval of events that match specified criteria rather than merely requesting data. For instance, Bob might query for all events recorded since a specific timestamp.

**Example JSON:**

```json
{
  "op": "query",
  "agent": "Bob",
  "timestamp": "2023-10-01T10:10:00Z",
  "target": "worldline123",
  "criteria": {
    "since": "2023-10-01T10:00:00Z"
  },
  "claims": [
    {"description": "Retrieve all events after the provided timestamp"}
  ],
  "signature": "BobQuerySignatureXYZ"
}
```

---

### 3. DAG Subscription Message Example

A subscription message is a promise by an agent to receive notifications for specific types of events. Here, Bob subscribes to insertion events on a worldline.

**Example JSON:**

```json
{
  "op": "subscribe",
  "agent": "Bob",
  "timestamp": "2023-10-01T10:15:00Z",
  "target": "worldline123",
  "criteria": {
    "filter": "insert"
  },
  "claims": [
    {"description": "Subscribe to insertion events on the worldline"}
  ],
  "signature": "BobSubsSignature456"
}
```

---

## Conclusion

The worldline completion model, by utilizing a Merkle DAG structure paired with Burgess-style promise-based messages, provides a robust framework for managing and verifying events in a messaging system. It ensures that every message acts as a verifiable promise, preserving the integrity and replayability of the DAG while accommodating complex operations such as branching, merging, and multi-worldline editing. Additionally, the model remains flexible enough to evolve into novel structures beyond conventional Merkle DAGs as system requirements mature.
