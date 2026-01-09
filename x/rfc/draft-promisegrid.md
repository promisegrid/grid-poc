Network Working Group                                        S. Traugott
Internet-Draft                                           March 1, 2025
Intended status: Informational                   Draft PromiseGrid Wire Protocol
Expires: July 2025

                PromiseGrid Wire Protocol
                Draft Version 00

Abstract

   This document specifies the wire protocol for PromiseGrid – a
   decentralized consensus-based computing, communications, and
   governance system.  The protocol defines the format and semantics of
   messages exchanged between PromiseGrid agents.  Messages are encoded in
   CBOR and wrapped in a "grid" tag whose value is an array with a
   protocol CID (pCID) as the first element.  The pCID defines the payload
   and signature formats, enabling evolution without changing the
   envelope.  Messages are transported over an underlying substrate
   (e.g., HTTP, NATS, libp2p, WebSocket, UDP, SCTP, file transfer) and MAY be linked
   using IPLD/DAG-CBOR.  This document also provides example
   pCID protocols, including DAG edit operations and Promise Theory
   primitives (promise, imposition, assessment).

Status of This Memo

   This Internet-Draft is submitted in full conformance with the
   provisions of BCP 78 and BCP 79.  This document is a work in progress
   and may be updated, replaced, or obsoleted by other documents at any
   time.  It is not appropriate to use this Internet-Draft as reference
   material.

Table of Contents

   1.  Introduction
   2.  Common Envelope (All pCIDs)
       2.1.  Minimal Envelope Rationale
       2.2.  Hypergraph Semantics
       2.3.  Canonical Encoding (Protocol-Defined)
       2.4.  Unknown pCID Handling
   3.  Example Protocol: DAG Edit Operations
       3.1.  Payload Fields (Example Protocol)
       3.2.  Operation Types (Example Protocol)
           3.2.1.  Insert Operation
           3.2.2.  Delete Operation
           3.2.3.  Reorder Operation
           3.2.4.  Query Operation
           3.2.5.  Subscription Operation
   4.  Example Protocol: Capability Call
   5.  Example Protocol: Scenario Tree
   6.  Example Protocol: Promise, Imposition, Assessment
   7.  Security Considerations
   8.  IANA Considerations
   9.  Acknowledgments
   10.  References

1.  Introduction

   PromiseGrid is designed as a decentralized system in which agents
   communicate via promises concerning modifications to shared state.
   In PromiseGrid, every message on the wire serves both as an instruction
   for a Directed Acyclic Graph (DAG) edit operation and as an irrevocable
   assertion of validity in a manner that is consistent with Promise Theory.
   This revision clarifies that each message explicitly describes an edit
   to the shared Merkle DAG – the ledger of global state – where operations
   include insertions, deletions, reordering, and other modifications.
   These DAG edits are interlinked via hash pointers using IPLD, enabling
   consistency checks, replayability, and audit trails over time.

   Example protocols in this document may leverage:
   
   o  CBOR for binary encoding of structured messages.
   o  COSE for digital signing and optional encryption.
   o  One or more transport bindings to carry encoded messages (e.g.,
      HTTP, NATS, libp2p, WebSocket, UDP, SCTP, file transfer).
   o  IPLD and DAG-CBOR for linking messages into a verifiable Merkle
      DAG.
   o  Explicit semantics wherein each message represents a promised edit to the
      DAG structure.

2.  Common Envelope (All pCIDs)

   This document has two major sections: Section 2 defines the common
   envelope that applies to all protocols identified by a pCID, and
   Sections 3 through 6 define example protocols that use this
   envelope.

   The PromiseGrid envelope is built on the following guiding principles:

   o  Each message is encoded in CBOR and wrapped in a "grid" tag whose
      value is an array with the protocol CID (pCID) as the first element.

   o  The pCID defines the payload and signature formats.  Protocols are
      free to define their own payload schemas and signature containers.

   o  Messages are transported over an underlying substrate.  This
      document specifies only the message encoding; transport selection
      and message framing (e.g., HTTP Content-Length or chunked bodies,
      NATS messages, WebSocket frames, UDP datagrams, SCTP records, or
      length-prefixing on stream transports such as TCP or QUIC) are
      defined by the chosen transport binding.

   o  The envelope is transport agnostic.  Although the wire encoding is
      based on CBOR, the underlying transport can vary.

2.1.  Minimal Envelope Rationale

   PromiseGrid treats each message as a protocol-defined function call.  The
   first element identifies the protocol using a content-addressed identifier
   (pCID).  The remaining elements are interpreted by that protocol, typically
   as the call's arguments.  This mirrors the capability-token-plus-payload
   model described in the project README and language notes, and avoids fixed
   version fields ahead of the token.  Protocol evolution is achieved by
   introducing new pCIDs without changing the envelope, keeping parsing simple
   for constrained nodes.

   This separation between a stable envelope and evolvable protocol
   semantics is consistent with long-horizon systems design. CBOR
   itself was designed for long-term interoperability: a compact,
   self-describing binary format with an explicit extensibility
   mechanism (tags) and optional deterministic encoding rules for
   signing and hashing [CBOR].

   PromiseGrid's pCID-first envelope follows a similar philosophy: the
   envelope is fixed and easy to parse, while the pCID is an explicit
   extension point that defines the semantics of the remaining items.
   Like CBOR tags, new pCIDs can be introduced without changing the base
   decoding rules, aiming for long-term viability.

   Unlike CBOR tags, which benefit from a centralized registry to avoid
   collisions, pCIDs do not require prior coordination: a pCID is
   typically the content hash (CID) of the protocol's specification
   document (or a canonical representation of it), so independent
   publishers can define new protocols without a shared numbering
   authority.

2.2.  Hypergraph Semantics

   PromiseGrid models messages as hyperedges in a shared Merkle DAG.  Each
   message links one or more parent nodes to a new node representing the
   asserted state or event.  Protocols MAY include explicit input/output CID
   lists in the payload to make head and tail sets explicit.  The example
   protocol in Section 3 uses prevHashes as tails and target as the head.

2.3.  Canonical Encoding (Protocol-Defined)

   Canonical encoding rules are defined by each protocol (pCID).  The
   envelope does not mandate DAG-CBOR or any other canonical CBOR profile.
   Protocols that rely on stable byte encodings for signatures or content
   addressing MUST specify their canonicalization rules.

2.4.  Unknown pCID Handling

   Nodes that do not recognize a pCID SHOULD treat the message as opaque.
   They MAY route or store the message without verification.  Resource
   limits (size caps, rate limits, and retention policies) SHOULD be
   applied locally.

3.  Example Protocol: DAG Edit Operations

   This section defines an edit-operations protocol identified by a pCID.
   The protocol payload is a CBOR map with fields described below and uses
   a COSE-based signature container for integrity.

      PromiseGridEnvelope = [
         pCID,
         payload,
         signature
      ]

   The payload for this protocol conforms to the following logical model:

      PromiseGridPayload = {
         "op":          OperationCode,
         "agent":       AgentID,
         "timestamp":   Timestamp,
         "target":      WorldlineID,
         "payload":     ClaimsPayload,
         "prevHashes":  [ HashValue, ... ],
         "signature":   SignatureData
      }

   The “prevHashes” field enables linking of messages within a DAG, as defined
   by IPLD using DAG-CBOR, to express relationships and ordering between events,
   thereby solidifying the edit history of the shared DAG.

3.1.  Payload Fields (Example Protocol)

   op
      A string value identifying the operation type (e.g., "insert",
      "delete", "reorder", "query", or "subscribe").

   agent
      Identifier of the agent issuing the promise.  This field may
      contain a globally unique name or cryptographic key identifier.

   timestamp
      A UTC timestamp, in ISO 8601 format, representing the issuing agent’s
      local time when the message is created.

   target
      A string or binary identifier denoting the target resource or
      worldline affected by the operation.

   payload
      An object containing one or more claims.  Claims are structured as key/value
      pairs that detail the intended edit operation on the DAG (e.g., descriptions,
      details of changes, and additional parameters).

   prevHashes
      An array of one or more hash values representing the immediate previous
      internal node(s) in the DAG.  These hash pointers, encoded with DAG-CBOR,
      provide the basis for linking messages in an IPLD structure, ensuring verifiable
      history and replayability.

   signature
      A COSE-encapsulated digital signature covering the message’s fields.
      This signature guarantees both authenticity of the issuing agent and the
      integrity of the promise embedded in the message for this protocol.

3.2.  Operation Types (Example Protocol)

   PromiseGrid supports several operation types.  The following sub-sections
   describe each supported type along with an example JSON representation (for human
   readability) that is subsequently encoded in CBOR.  Each operation represents a specific
   edit to the shared DAG of events.

3.2.1.  Insert Operation

   The insert operation signals the addition of a new event into a worldline.
   The payload conveys a promise to insert data into the DAG along with necessary metadata.

   Example:

      {
         "op": "insert",
         "agent": "Alice",
         "timestamp": "2023-10-01T10:05:00Z",
         "target": "worldline123",
         "payload": {
           "claims": [
             {
               "description": "Insert event 'Hello' as an initial greeting",
               "detail": "Welcome message"
             }
           ]
         },
         "prevHashes": [ "hash_internalNode1", "hash_internalNode2" ],
         "signature": "AliceSignatureABC123"
      }

3.2.2.  Delete Operation

   A delete operation signals the intention to remove an existing event or mark it as obsolete.
   Context is provided to enable validators to confirm the deletion from the DAG.

   Example:

      {
         "op": "delete",
         "agent": "Bob",
         "timestamp": "2023-10-01T10:20:00Z",
         "target": "worldline123",
         "payload": {
           "claims": [
             {
               "description": "Delete event 'Obsolete'",
               "detail": "Removing outdated information"
             }
           ]
         },
         "prevHashes": [ "hash_internalNode3" ],
         "signature": "BobSignatureXYZ789"
      }

3.2.3.  Reorder Operation

   Reordering events within a worldline may be necessary for logical or temporal
   restructuring.  The payload details the intended new order of events in the DAG.

   Example:

      {
         "op": "reorder",
         "agent": "Alice",
         "timestamp": "2023-10-01T10:30:00Z",
         "target": "worldline123",
         "payload": {
           "claims": [
             {
               "description": "Reorder events to prioritize recent updates",
               "detail": "Moving update event up"
             }
           ]
         },
         "prevHashes": [ "hash_internalNode4" ],
         "signature": "AliceSignatureDEF456"
      }

3.2.4.  Query Operation

   A query operation is used by an agent to request information about events matching
   specified criteria within the DAG.  The query itself is a promise regarding retrieval
   behavior over the DAG’s structure.

   Example:

      {
         "op": "query",
         "agent": "Bob",
         "timestamp": "2023-10-01T10:10:00Z",
         "target": "worldline123",
         "payload": {
           "criteria": {
             "since": "2023-10-01T10:00:00Z"
           },
           "claims": [
             {
               "description": "Retrieve all events after the provided timestamp"
             }
           ]
         },
         "signature": "BobQuerySignatureXYZ"
      }

3.2.5.  Subscription Operation

   Subscription operations allow agents to register interest in future events
   (e.g., insertions) on specified worldlines.  The promise is to receive
   notifications when events that meet the given criteria occur in the DAG.

   Example:

      {
         "op": "subscribe",
         "agent": "Bob",
         "timestamp": "2023-10-01T10:15:00Z",
         "target": "worldline123",
         "payload": {
           "criteria": {
             "filter": "insert"
           },
           "claims": [
             {
               "description": "Subscribe to insertion events on the worldline"
             }
           ]
         },
         "signature": "BobSubsSignature456"
      }

4.  Example Protocol: Capability Call

   This protocol uses a single pCID (capcall_pCID) for capability-call
   messages.  The payload is a CBOR array whose first element is a
   message type (mtype); remaining elements are defined by the protocol.

      payload = [ mtype, ... ]

   One minimal call request profile is:

      payload = [ 0, fCID, [ arg1, arg2 ] ]

   where:

      0     call request
      fCID  function or capability CID to invoke

   Additional message types (e.g., response, error) are protocol-defined.
   The signature container and signature input are also defined by
   capcall_pCID; a detached COSE_Sign1 signature over a canonical encoding
   of [pCID, payload] is one common profile.

      PromiseGridEnvelope = [
         capcall_pCID,     ; capability-call protocol
         [ 0, fCID, [ arg1, arg2 ] ],
         signature
      ]

5.  Example Protocol: Scenario Tree

   The scenario-tree protocol encodes probabilistic state transitions as
   nested lists of branches with binary CIDs, fixed-point probabilities,
   and weights.  It uses the common envelope and defines its payload and
   signature profile separately.  See `x/wire/wire.md` for the full
   example protocol.

6.  Example Protocol: Promise, Imposition, Assessment

   Promise Theory distinguishes promises (declared intentions about self),
   impositions (e.g., requests), and independent assessments made by
   observers [Promise Theory].  This example protocol uses a single pCID and distinguishes
   message types within its payload.

   Bergstra and Burgess apply this framing to ownership and money, treating
   them as networks of promises and assessments [MoneyBook].

   Payload format:

      payload = [ mtype, body ]

   where mtype is:

      0  promise     (offer about self)
      1  imposition  (request)
      2  assessment  (receipt by an observer)

   Example (illustrative):

      PromiseGridEnvelope = [ pt_pCID, [ 0, [ intent, terms ] ], signature ]
      PromiseGridEnvelope = [ pt_pCID, [ 1, [ intent, args  ] ], signature ]
      PromiseGridEnvelope = [ pt_pCID, [ 2, [ about_cid, outcome ] ], signature ]

7.  Security Considerations

   PromiseGrid does not mandate a specific cryptographic container at the
   envelope level; security is defined by each protocol (pCID).  Protocols
   that require authenticity and integrity SHOULD specify how signatures
   are represented, how signing keys are identified, and whether any
   replay protection is required.

   Nodes that route or store messages for unknown pCIDs without
   verification SHOULD apply local resource limits (see Section 2.4).

   Implementers are advised to use secure cryptographic hash functions (e.g., SHA-256 or
   stronger) and follow best practices for key management and certificate validation.

8.  IANA Considerations

   This document does not specify any new registries.  However, future versions may
   define a registry for PromiseGrid operation codes and parameter names.

9.  Acknowledgments

   The authors gratefully acknowledge contributions from the PromiseGrid research
   community, as well as insights drawn from related work in IPFS, IPLD, CBOR, COSE,
   and Promise Theory.

10.  References

   [RFC 8152]   Crocker, D., "CBOR Object Signing and Encryption (COSE)", RFC 8152,
                April 2017.

   [RFC 8392]   Bormann, C., "CBOR Web Token (CWT)", RFC 8392, November 2018.

   [Promise Theory]  Bergstra, J. and Burgess, M., "Promise Theory: Principles and
                Applications", Book of Promises, 2nd ed., 2019.

   [MoneyBook]  Bergstra, J. and Burgess, M., "Money, Ownership, and Agency:
                As an Application of Promise Theory", χtAxis press, 2019.

   [CBOR]   Bormann, C. and Hoffman, P., "Concise Binary Object Representation (CBOR)",
                RFC 8949, December 2020.

   [IPFS]   Benet, J., "IPFS - Content Addressed, Versioned, P2P File System",
                2014.

   [IPLD]   Benet, J., "InterPlanetary Linked Data (IPLD)", https://ipld.io

   [DAG-CBOR]  Technical documentation on DAG-CBOR, https://github.com/ipld/dag-cbor

   [libp2p]  Protocol and documentation available at https://libp2p.io

   [IETF RFC Styles]   "Guidelines for Writing an Internet-Draft", IETF RFC 7322.

   This document is an Internet-Draft and is provided for discussion purposes only.

                              Authors' Addresses

   D. Traugott
   Email: stevegt@example.com

Disclaimer

   This document is provided on an "AS IS" basis and the authors DISCLAIM
   any and all warranties, express or implied, including without limitation
   any warranty related to fitness for a particular purpose.

Conclusion

   The PromiseGrid envelope outlined in this document enables evolvable
   messaging across a decentralized platform.  By using a pCID-first CBOR
   envelope, new protocols can be introduced without changing the wire
   format.  Protocols MAY use signatures and IPLD/DAG-CBOR links to support
   verification and replayability where needed.  Future extensions may
   elaborate on merge-as-consensus workflows and additional pCID
   protocols.

                              End of Document
