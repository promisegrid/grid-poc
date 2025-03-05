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
   messages exchanged between PromiseGrid agents.  Messages are expressed
   as structured claims based on CBOR Web Tokens (CWTs), digitally signed
   using COSE, and transmitted via a libp2p overlay network.  In addition,
   messages are interlinked using IPLD with DAG-CBOR encoding, ensuring
   verifiable relationships and replayability of event sequences.
   This document is intended to be consistent with IETF RFC style and
   practices.

Status of This Memo

   This Internet-Draft is submitted in full conformance with the
   provisions of BCP 78 and BCP 79.  This document is a work in progress
   and may be updated, replaced, or obsoleted by other documents at any
   time.  It is not appropriate to use this Internet-Draft as reference
   material.

Table of Contents

   1.  Introduction
   2.  Protocol Overview
   3.  Message Format
       3.1.  Common Fields
       3.2.  Operation Types
           3.2.1.  Insert Operation
           3.2.2.  Delete Operation
           3.2.3.  Reorder Operation
           3.2.4.  Query Operation
           3.2.5.  Subscription Operation
   4.  Security Considerations
   5.  IANA Considerations
   6.  Acknowledgments
   7.  References

1.  Introduction

   PromiseGrid is designed as a decentralized system in which agents
   communicate via promises concerning modifications to shared state.
   In PromiseGrid, every message on the wire serves both as an instruction
   for a Directed Acyclic Graph (DAG) edit operation and as an irrevocable
   assertion of validity in a manner that is consistent with Promise Theory.
   This document describes the wire format and processing semantics for
   these messages, hereby called “PromiseGrid Messages.”

   Notably, PromiseGrid leverages:
   
   o  CBOR for binary encoding of structured messages.
   o  COSE for digital signing and optional encryption of messages.
   o  libp2p as the transport mechanism for robust peer-to-peer communication.
   o  IPLD and DAG-CBOR for linking messages into a verifiable Merkle
      Directed Acyclic Graph, enabling replayability and audit trails.

2.  Protocol Overview

   The PromiseGrid wire protocol is built on the following guiding
   principles:

   o  Each message is expressed as a set of claims encapsulated within a
      CWT (CBOR Web Token).  These claims represent promises about DAG
      edit operations such as insertions, deletions, reordering, queries,
      and subscription requests.

   o  Every message is signed using COSE (CBOR Object Signing and
      Encryption) to ensure authenticity and integrity.  The digital
      signature binds the message to the issuing agent, whose identity is
      expressed in the claims.

   o  Messages are transmitted over a libp2p network, which provides a
      decentralized, peer-to-peer transport layer that abstracts from
      underlying network protocols (e.g., TCP, QUIC).

   o  Relationships between messages are maintained using IPLD.  Each
      message may reference previous messages via hash pointers using
      DAG-CBOR encoding, thereby constructing a Merkle Directed Acyclic Graph
      that supports consistency checks and replayability.

   o  The protocol is transport agnostic beyond its reliance on libp2p.
      Although the wire encoding is based on the compact binary format of CBOR,
      underlying transport layers can vary.

3.  Message Format

   PromiseGrid Messages are encoded using CBOR and secured with COSE.
   The overall structure of a message conforms to the following logical
   model:

      PromiseGridMessage = {
         "op":          OperationCode,
         "agent":       AgentID,
         "timestamp":   Timestamp,
         "target":      WorldlineID,
         "payload":     ClaimsPayload,
         "prevHashes":  [ HashValue, ... ],
         "signature":   SignatureData
      }

   The “prevHashes” field enables linking of messages within a DAG, as defined
   by IPLD using DAG-CBOR, to express relationships and ordering between events.

3.1.  Common Fields

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
      pairs that detail the intended edit operation (e.g., descriptions, details of
      changes, and additional parameters).

   prevHashes
      An array of one or more hash values representing the immediate previous
      internal node(s) in the DAG.  These hash pointers, encoded with DAG-CBOR,
      provide the basis for linking messages in an IPLD structure, ensuring verifiable
      history and replayability.

   signature
      A COSE-encapsulated digital signature covering the message’s fields.
      This signature guarantees both authenticity of the issuing agent and the
      integrity of the promise embedded in the message.

3.2.  Operation Types

   PromiseGrid supports several operation types.  The following sub-sections
   describe each supported type along with an example JSON representation (for human
   readability) that is subsequently encoded in CBOR.

3.2.1.  Insert Operation

   The insert operation signals the addition of a new event into a worldline.
   The payload conveys a promise to insert data along with necessary metadata.

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
   Context is provided to enable validators to confirm the deletion.

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
   restructuring.  The payload details the intended new order.

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
   specified criteria.  Rather than directly retrieving data, the query itself is a promise
   regarding retrieval behavior.

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
   (e.g., insertions) on specified worldlines.  The promise here is to receive
   notifications when events that meet the given criteria occur.

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

4.  Security Considerations

   The security of the PromiseGrid wire protocol is achieved by employing COSE for
   digital signatures along with the compact token format of CWT.  Each message’s signature
   binds the agent’s identity to the promise of the DAG edit operation, enabling recipients
   to verify:
   
   o  Message integrity – any tampering is detectable via hash mismatches.
   o  Authenticity – the origin of the message is authenticated by verifying the signature
      against the issuing agent’s public key.
   o  Replay protection – timestamps and ‘prevHashes’ (as IPLD links) enable detection of
      message replays.

   Implementers are advised to use secure cryptographic hash functions (e.g., SHA-256 or
   stronger) and follow best practices for key management and certificate validation.

5.  IANA Considerations

   This document does not specify any new registries.  However, future versions may
   define a registry for PromiseGrid operation codes and parameter names.

6.  Acknowledgments

   The authors gratefully acknowledge contributions from the PromiseGrid research
   community, as well as insights drawn from related work in IPFS, IPLD, CBOR, COSE,
   and Promise Theory.

7.  References

   [RFC 8152]   Crocker, D., "CBOR Object Signing and Encryption (COSE)", RFC 8152,
                April 2017.

   [RFC 8392]   Bormann, C., "CBOR Web Token (CWT)", RFC 8392, November 2018.

   [Promise Theory]  Burgess, M., "Promise Theory: Principles and Applications",
                2005.

   [CBOR]   Jennings, C., "Concise Binary Object Representation (CBOR)", IETF Draft.

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

   The PromiseGrid Wire Protocol outlined in this document enables secure,
   verifiable, and replayable operations across a decentralized platform.
   By leveraging CBOR for encoding, COSE for digital signatures, libp2p for
   message transport, and IPLD with DAG-CBOR for linking messages, the protocol
   ensures that promises made by autonomous agents are auditable and trustworthy.
   Future extensions may further elaborate on advanced merging, conflict resolution
   mechanisms, and additional operation types.

                              End of Document
