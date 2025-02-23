# Worldline Completion Model in Messaging Systems

The worldline completion model offers an innovative approach to managing events in a messaging system. It emphasizes creating, editing, and verifying a complete history of events for any given conversation or entity via a Merkle DAG structure. This document reviews the pros and cons of this model, provides a design outline of a worldline data structure, presents concrete examples with agents like Alice and Bob, and discusses how operations (DAG edits, queries, and subscriptions) can be expressed as CWT claims. Finally, it reviews the document for any internal inconsistencies and provides recommendations for terminology clarity.

---

## Overview of the Worldline Completion Model

In this model, every message is treated as a DAG edit operation—a promise that the resulting worldline is valid. Key properties include:

- **Multiple Worldlines:**  
  Worldlines can branch and merge, meaning that multiple histories can evolve concurrently. A single event may appear in more than one worldline, reflecting different perspectives or states.

- **DAG-Based Organization:**  
  Events are not simply locked into a linear sequence; instead, they are organized within a Merkle DAG:
  - **Leaf Nodes (Events):**  
    Represent immutable records that contain event data (e.g., a text message), metadata (such as a timestamp and the agent’s identity), and a payload consisting of one or more PT-style promises (expressed as CWT claims). Importantly, leaves do not store a hash of the previous event.
  - **Internal Nodes:**  
    Serve as logs of the messages that edited the worldline. They contain the hashes of their children arranged in worldline order, effectively proving the history’s integrity. When editing the DAG (inserting, deleting, or reordering events), the message includes the hashes of the previous internal nodes being updated.

- **Unified Editing, Query, and Subscription Language:**  
  A single language is used to express DAG edit operations, query requests, and subscription requests. Each operation is interpreted as a promise (via CWT claims) that certain edits or validations hold true in the worldline.

- **Message Integrity:**  
  Each message (i.e., DAG edit) is signed by the agent who creates it, and timestamps are recorded from the perspective of that agent. This ensures that the historical record is both auditable and replayable. A message may affect multiple worldlines at once, and replayability is central to reconstructing the state accurately—even if an agent was offline for a period.

- **Terminology Clarification:**  
  There is a tendency in the documentation to interchange the terms "message" and "event."  
  - **Recommendation:**  
    Use **"event"** exclusively to refer to the immutable record stored in the Merkle DAG, and **"message"** for the DAG edit (or query/subscription) operation that conveys a promise or edit instruction.

---

## Worldline Data Structure

A conceptual Go-style structure for a worldline event might be:

```go
type WorldlineEvent struct {
    Timestamp time.Time // Timestamp from the creating agent's perspective.
    Agent     string    // Creator's identifier, e.g., "Alice" or "Bob"
    Payload   []byte    // A set of PT-style promises expressed as CWT claims.
}
```

**Merkle DAG Characteristics:**

- **Editable History:**  
  Each edit operation (insert, delete, reorder) creates a new message that includes the current hash pointers from internal nodes. These pointers validate that the change is consistent with the prior state.

- **No Explicit Previous Hash in Leaves:**  
  Instead of linking directly to a prior event, the integrity of the history is maintained by internal nodes, which carry ordered child hashes.

- **DAG Node as Log:**  
  Each internal node acts as a log of the messages (or edit operations) that have affected the worldline, allowing for complex operations like branching and merging.

- **Consistency and Replayability:**  
  Because every edit is a verifiable promise and the history is fully maintained, agents can replay events from any point, ensuring that state reconstruction is deterministic.

---

## Example Scenarios

### 1. DAG Edit Message Example

An agent (e.g., Alice) sends a message that performs a DAG edit operation by inserting a new event into a specified worldline. This message includes the hashes of the relevant previous Merkle DAG nodes, ensuring that the edit is applied to a known state.

**Example JSON Representation:**

```json
{
  "op": "insert",
  "agent": "Alice",
  "timestamp": "2023-10-01T10:05:00Z",
  "target": "worldline123",
  "payload": {
    "claims": [
      {"promise": "Event 'Hello' inserted", "detail": "Initial greeting"}
    ]
  },
  "prevHashes": ["hash_internalNode1", "hash_internalNode2"],
  "signature": "abc123signature"
}
```

In this message, Alice promises that a "Hello" event is inserted into the worldline `worldline123` and provides the previous Merkle DAG internal node hashes to validate the edit.

---

### 2. DAG Query Message Example

An agent (e.g., Bob) may query a worldline to retrieve events after a specific timestamp. Although this is a query operation, it is expressed as a promise (request) that certain events exist, thereby following the same expression rules as other DAG messages.

**Example JSON Representation:**

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
    XXX this is not a promise, it is a request.  needs to be a promise.
    {"promise": "Retrieve all events after the specified timestamp"}
  ],
  "signature": "bobSignatureXYZ"
}
```

Here, Bob issues a query promise to retrieve all events from `worldline123` that occurred after `2023-10-01T10:00:00Z`.

---

### 3. DAG Subscription Message Example

An agent can subscribe to changes in a worldline—thereby receiving real-time updates as the DAG is edited. In this case, the subscription request is also framed as a promise that initiates a persistent link to the worldline.

**Example JSON Representation:**

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
    XXX this is not a promise, it is a request.  needs to be a promise.
    {"promise": "Subscribe to insertion events in the worldline"}
  ],
  "signature": "bobSubsSignature456"
}
```

This message expresses Bob’s promise to be notified of any new insertion events on the worldline `worldline123`.

---

## Consistency and Final Review

- **Message vs. Event:**  
  The document initially used "message" and "event" interchangeably. We now reserve "event" for the immutable record in the Merkle DAG and "message" for the operational edit, query, or subscription instruction.

- **Unified Language:**  
  The same language is employed throughout to express DAG edit, query, and subscription operations, ensuring consistency across different use cases.

- **Multiple Worldlines and Editing Operations:**  
  The model clearly supports branching, merging, multi-worldline edits, and the inclusion of previous Merkle hashes in edit messages to ensure historical consistency.

- **Expressibility via CWT Claims:**  
  All operations (DAG edit, query, and subscription) are framed as promises carried in CWT claims, reinforcing security and contractual integrity between agents.

- **Potential Inconsistencies Reviewed:**  
  While the conceptual model uses a unified promise model, practical implementation details (like hash inclusion and proper claim differentiation) must be carefully managed to prevent ambiguity. No additional inconsistencies were discovered beyond the clarified nomenclature.

---

## Conclusion

The worldline completion model leverages a Merkle DAG structure to provide a reliable, verifiable, and replayable history of events, supporting complex operations like branching, merging, and flexible editing. By expressing all operations as promises (using CWT claims) and maintaining a clear distinction between operational messages and immutable events, the system promotes both robustness and clarity. Agents such as Alice and Bob can confidently edit, query, and subscribe to events while ensuring that every edit operation is cryptographically secured and historically consistent.
