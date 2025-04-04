# PromiseGrid Wire Protocol Specification

This specification document defines the PromiseGrid wire protocol, a system that leverages CBOR, COSE, and IPLD/DAG-CBOR technologies to securely transmit signed assertions of state transitions in a distributed network. The protocol is designed to support an infinite state machine modeled as a decision tree in which each branch is probabilistically weighted. This document describes the technical background, message formats, and design considerations for implementing such a protocol.

## 1. Introduction

PromiseGrid aims to provide a robust mechanism for sharing conditional promises between agents over a decentralized network. Each message in the system is intended to be a signed assertion—expressed as a COSE-signed CWT—about subsets of state transitions. These transitions are captured by linking event and state content through content identifiers (CIDs) in a DAG-CBOR compliant format. The design targets both efficient routing (by exposing minimal header information) and secure end-to-end integrity verification.

## 2. Background

### 2.1 CBOR and DAG-CBOR

CBOR (Concise Binary Object Representation) is a binary data serialization format standardized in RFC 8949. DAG-CBOR is a variant that adheres to the IPLD Data Model requirements. Among other constraints, DAG-CBOR mandates that:
- Map keys must be text strings.
- Content links are encoded using CBOR tag 42.
- Certain floating-point values such as NaN or Infinity are disallowed to preserve deterministic encoding.

These constraints ensure that serialized data remains canonical, which is critical for content addressing based on cryptographic hashing.

### 2.2 Content Identifiers (CIDs) and Identity Multibase

CIDv1 is the prevalent format used in IPLD, representing a pointer to content in a content-addressed system. A CIDv1 is composed of a multibase prefix, a version (typically 1), a multicodec indicating the data format (e.g., dag-cbor), and the multihash (often using SHA2-256) of the content. The identity multibase utilizes a one-byte prefix to represent the content in its raw binary form, reducing encoding overhead. For use in maps within DAG-CBOR, however, full CID strings (e.g., base32 encoded) must be employed because native binary forms (such as those with tag 42) are not permissible as map keys.

### 2.3 COSE and CWT Integration

COSE (CBOR Object Signing and Encryption) provides a standardized mechanism for protecting and verifying data encoded in CBOR. In a COSE message:
- The **protected** header contains cryptographically bound parameters such as the signature algorithm (e.g., ES256, indicated by –7) and key identifiers. Additionally, it carries CWT (CBOR Web Token) claims that encapsulate metadata such as issuer identity, expected probabilities, and global trust weights.
- The **payload** comprises the message content—in this protocol, a decision tree representing state transitions and pinning agreements.
- The **external_aad** field (external additional authenticated data) may be provided for contextual data (e.g., routing hints) that is not included in the payload.
The signature is computed over both the protected headers and the payload (and any external_aad), ensuring that changes in either invalidate the signature.

## 3. Protocol Message Design

### 3.1 Overall Message Structure

A PromiseGrid message is constructed as follows:
1. An outer envelope marked by a custom protocol tag (for example, the IANA-registered “grid” tag with value 0x67726964) which encapsulates the protocol version (expressed as a protocol CID) and the signed message.
2. A COSE_Sign1 object that includes:
   - A protected header with algorithm identifier, key identifiers, and CWT claims.
   - A payload that contains a nested data structure representing the decision tree of state transitions.
   - A signature computed over the Sig_structure, which includes the protected headers, any external_aad, and the payload.

### 3.2 Decision Tree Model as Nested Data Structures

The state transition information is modeled as an infinite decision tree. Each node in the tree is a branch represented by a list that holds:
- The event identifier (expressed as a full CID string to comply with IPLD mapping rules).
- The probability of the event occurring.
- A weight multiplier that will affect all subsequent probability evaluations along that branch.
- A nested branch that follows the same structure.

For example, one branch of the decision tree is represented as:

```cbor
[
  "bafyrei...event1",  // Event CID (full string, non-truncated)
  0.65,                // Probability of this event
  1.2,                 // Weight multiplier for subsequent branches
  [
    [
      "bafkre...state1", 
      0.9,
      1.0,
      []
    ],
    [
      "bafybei...state2",
      0.1,
      0.8,
      [ /* further nested branches */ ]
    ]
  ]
]
```

This flat list structure allows agents to rapidly extract routing-relevant information (the decision tree header) without having to process the entire, possibly deeply nested, message.

### 3.3 COSE-Sign1 Encapsulation with CWT Claims

A representative message format employing COSE with CWT claims might be illustrated as:

```cbor
COSE_Sign1(
  protected: <<{
    1: -7,                        // Algorithm: ES256
    15: {                         // CWT Claims
      "prb": {                    // Encoded probability claims (e.g., event probabilities)
        "event1": 0.65,
        "event2": 0.35
      },
      "wgt": 1.2,                 // Global trust weight multiplier
      "iss": "PromiseGridNode/42" // Issuer identifier as a string or CID
    }
  }>>,
  payload: <<{
    "tree": "bafyrei...decisionTreeCID",  // Root CID of the decision tree (full string)
    "pinning": {
      "commit": "bafyrei...pinContractCID", // CID identifying the pinning agreement or contract
      "peers": [
        "bafyrei...peer1",
        "bafyrei...peer2"
      ]
    }
  }>>,
  signature: h'3045...signatureBytes'
)
```

In this configuration, the protected header holds crucial metadata; the payload carries the decision tree and pinning instructions; and the signature, computed over a Sig_structure that typically includes `[protected_headers, external_aad, payload]`, ensures overall message integrity.

## 4. Detailed Design Considerations

### 4.1 Map Keys and CID Strings

Because IPLD and DAG-CBOR enforce that map keys be text strings (and not binary objects such as those with tag 42), all CIDs used as keys must be the full, non-truncated base32-encoded strings. This may impose a small overhead in terms of data size, but it ensures interoperability with tools that expect IPLD maps to have string keys.

### 4.2 Handling of Floating-Point Values

CBOR supports floating-point values, but both DAG-CBOR and the IPLD data model prohibit the use of non-deterministic float representations such as NaN or Infinity. In this protocol, probabilistic values (e.g., 0.65 for a probability) are better represented as fixed-point numbers or constrained floats that do not compromise deterministic encoding. Users may consider multiplying probabilities by an appropriate factor (for instance, representing 0.65 as 6500 if using a thousand-based fixed-point system) to avoid issues with floating-point variations.

### 4.3 Trust Metric Computation

The trust metric for an agent is updated as the product of each branch’s weight and the difference between actual outcomes and the predicted probability. A sample implementation in Python is as follows:

```python
def evaluate_trust(chain):
    trust = 1.0
    for claim in chain:
        actual = claim['actual']          # Observed outcome
        predicted = claim['predicted']    # Claimed probability
        # Incorporate weight multiplier and the outcome difference
        trust *= claim['weight'] * (actual - predicted)
    return trust
```

This formulation permits negative trust values when the actual outcome is below the predicted probability, thereby providing a quantifiable measure of deviation over time.

### 4.4 Routing and Depth Considerations

There is no inherent depth limit imposed by IPLD or DAG-CBOR on nested data structures. However, practical considerations on constrained devices (such as IoT endpoints) may necessitate a limit on the recursion depth to prevent excessive resource consumption. Protocol designers must balance the need for expressiveness in state transitions with the limitations of the hardware.

### 4.5 CID Versioning and Future-Proofing

CID strings in PromiseGrid are currently expressed as CIDv1. This format—comprising a multibase prefix, version byte, multicodec, and multihash—supports forward compatibility due to its self-describing nature. Future protocol updates (for example, adopting CIDv2 or new hash algorithms) will require careful version negotiation. In the meantime, using full CIDv1 strings (rather than truncated or simplified representations) ensures that each identifier remains robust and unambiguous.

## 5. Open Questions and Future Work

Several areas require further investigation and prototyping:

- **Negative Trust Interpretation:** How should the system handle agents that yield net-negative trust scores? What are the policy implications for such results?
- **Probability Representation:** Would fixed-point integer representation offer greater determinism and efficiency over conventional floating-point values?
- **Depth Limits:** In highly constrained environments, what is the optimal maximum depth for the decision tree without compromising protocol functionality or risking stack overflow?
- **Versioning Strategies:** As CID formats evolve, what strategies should be adopted to ensure backward compatibility while enabling new features?
- **Signature Positioning:** Does repositioning the signature (for example, moving it to the end of the message) provide measurable routing performance benefits without affecting security?

## 6. Conclusion

This specification lays out the design for a PromiseGrid wire protocol that leverages modern technologies—CBOR, COSE, and IPLD—to securely and efficiently represent an infinite state machine as a decision tree with probabilistic transitions. By enforcing strict adherence to IPLD’s rules (such as requiring string map keys for CIDs), and by integrating COSE-signed CWT claims, this protocol aims to promote both secure communication and robust state consistency across distributed agents. The open questions articulated in this specification serve as a roadmap for subsequent iterations of the protocol and prototype implementations.

