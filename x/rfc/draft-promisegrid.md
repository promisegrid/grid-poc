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
   as structured claims based on CBOR Web Tokens (CWTs) and secured using
   COSE, ensuring integrity, authenticity, and non-repudiation of promises.
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

   o  Messages include essential context such as timestamps, the target
      worldline or resource identifier, prior node hashes for replayability,
      and an operation code specifying the type of edit or query.

   o  The protocol is transport agnostic.  While the wire encoding
      relies on a compact binary format based on CBOR, underlying transport
      layers can vary (e.g., TCP, HTTP/2, QUIC).

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

3.1.  Common Fields

   op
      A string value that identifies the operation type (e.g., "insert",
      "delete", "reorder", "query", or "subscribe").

   agent
      Identifier of the agent issuing the promise.  This field may
      contain a globally unique name or cryptographic key identifier.

   timestamp
      A UTC timestamp (in ISO 8601 format) representing the creating agent’s
      local time when the message is issued.

   target
      A string or binary identifier denoting the target resource or
      worldline affected by the operation.

   payload
      An object containing one or more claims.  Claims are structured as key/value
      pairs that detail the intended edit operation (e.g., descriptions, details of
      changes, and additional parameters).

   prevHashes
      An array of one or more hash values representing the immediate previous
      internal node(s) in the DAG.  This field supports replayability and consistency
      checks.

   signature
      A COSE-encapsulated digital signature covering the fields of the message.
      This signature provides in-band security guaranteeing the authenticity and the
      promise integrity of the message.

3.2.  Operation Types

   PromiseGrid supports several operation types.  The following sub-sections describe
   each supported type along with an example JSON representation (for human readability)
   that is subsequently encoded in CBOR.

3.2.1.  Insert Operation

   The insert operation signals the addition of a new event into a worldline.
   The payload conveys a promise to insert data along with relevant metadata.

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
   Context is provided to enable validators to confirm the operation.

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
   certain criteria.  Rather than directly retrieving data, the query itself is a promise
   regarding the retrieval behavior.

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
   (e.g., insertions) on particular worldlines.  The promise here is to receive
   notifications when events meeting specified criteria occur.

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

   o  Message integrity – any tampering is detectable by a mismatch in the computed hash.
   o  Authenticity – the origin of the message is authenticated by verifying the signature
      against the issuing agent’s public key.
   o  Replay protection – inclusion of timestamps and previous node hashes enables detection
      of message replays.

   Implementers are advised to use secure cryptographic hash functions (e.g., SHA-256 or
   stronger) and follow best practices for key management and certificate validation.

5.  IANA Considerations

   This document does not specify any new registries.  However, future versions may
   define a registry for PromiseGrid operation codes and parameter names.

6.  Acknowledgments

   The authors gratefully acknowledge contributions from the PromiseGrid research
   community, as well as insights drawn from related work in IPFS, CBOR, COSE, and
   Promise Theory.

7.  References

   [RFC 8152]   Crocker, D., "CBOR Object Signing and Encryption (COSE)", RFC 8152,
                April 2017.

   [RFC 8392]   Bormann, C., "CBOR Web Token (CWT)", RFC 8392, November 2018.

   [Promise Theory]  Burgess, M., "Promise Theory: Principles and Applications",
                2005.

   [CBOR]   Jennings, C., "Concise Binary Object Representation (CBOR)", IETF Draft.

   [IPFS]   Benet, J., "IPFS - Content Addressed, Versioned, P2P File System",
                2014.

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
   Through its use of structured CWT claims, digital signatures using COSE,
   and content-addressable references, the protocol ensures that the promises
   made by agents can be audited and trusted across a network of autonomous nodes.
   Future extensions may further elaborate on advanced merging, conflict
   resolution mechanisms, and additional operation types.


                              End of Document
