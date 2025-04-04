# PromiseGrid Wire Protocol Chat Discussion

This document presents an exploratory discussion regarding the design of PromiseGrid's wire protocol. In this chat‐style document we discuss several technical topics including the interpretation of COSE message fields (protected headers, payload, and external AAD), the concept of identity multibase in the context of CIDv1, the challenges of using native binary CIDs as map keys in IPLD/DAG-CBOR, and the design of a decision tree model for representing an infinite state machine with probabilistic transitions.

## COSE Message Fields – Protected, Payload, and external_aad

In a COSE (CBOR Object Signing and Encryption) message the data that gets signed is derived from a structure known as the Sig_structure. This structure comprises several elements: first, the protected headers that include cryptographically bound parameters such as the algorithm identifier (for example, –7 to denote ES256), key identifiers, and additional claims (commonly expressed as a CBOR Web Token (CWT) claim set); second, an optional external additional authenticated data (external_aad) that is not part of the message itself but still factored into the cryptographic signature; and finally the payload which encapsulates the core content of the message. The signature itself is computed over all those concatenated parts, ensuring that both the protected header fields and the payload are tamper‐evident. External AAD may hold contextual information like network-layer headers; its inclusion provides additional binding without increasing the message payload size.

## Identity Multibase and CID Representation

The term "identity multibase" refers to the minimal encoding method used in CIDv1 to represent the content identifier in its raw binary form. A typical CIDv1 in IPLD uses CBOR tag 42 and starts with a one-byte identity multibase marker, followed by bytes that signify the CID version, the multicodec for the content type (for example, indicating dag-cbor), and finally the multihash (often SHA2-256). This binary representation is much more compact than the human‐readable base32 encoding and is well suited for constrained networking environments. However, within IPLD and DAG-CBOR, map keys must be strings rather than raw binary objects. This means that even though links can natively be represented as binary CIDs (tag 42), when used as keys in maps they must be transformed to full CID strings (for example, the base32-encoded version) without truncation.

## Designing a Decision Tree for an Infinite State Machine

The concept behind the PromiseGrid promise is to model a decision tree where each branch represents the outcome of an event in an infinite state machine describing the universe. In this design each branch carries three key pieces of information: (1) an event identifier represented by a full CID string; (2) the probability that the event occurs; and (3) a weight multiplier that should be applied to all subsequent probabilities along that branch. The branch structure is recursively nested, meaning that the tree is represented as a nested list of lists in which each inner list follows the pattern:

  [event CID, probability, weight multiplier, branch]

For example, one possible representation might be:

```cbor
[
  [
    "bafyrei...event1",  // Full CID string corresponding to an event schema
    0.65,                // Probability of this event
    1.2,                 // Weight multiplier for the subtree
    [                    // Nested branch representing subsequent state transitions
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
        [/* further nested branches */]
      ]
    ]
  ]
]
```

This nested list design captures the full decision tree where each node’s information is self-contained and allows routers or agents to read only the beginning of the message (for routing efficiency) before deciding how deeply to parse into the nested structure. Notice that negative values are allowed in the trust computation – these result from the difference between actual observed outcomes and predicted probabilities, and they serve as a meaningful indicator in the trust metric calculation.

## Trust Metric Calculation and Code Modification

In the context of PromiseGrid, an agent’s trustworthiness is evaluated by traversing the decision tree over time and comparing the agent’s predicted probabilities with what actually occurs. A simple Python function to compute trust has been modified as follows to meet the design requirements:

```python
def evaluate_trust(chain):
    trust = 1.0
    for claim in chain:
        actual = claim['actual']          # Observed outcome value
        predicted = claim['predicted']    # Claimed probability
        # Updated trust calculation multiplies by the weight and the difference (actual - predicted)
        trust *= claim['weight'] * (actual - predicted)
    return trust

# Example usage with full (non-truncated) CID strings as keys:
data_chain = [
    {
        'cid': 'bafyrei...fullcid',
        'weight': 1.2,
        'actual': 0.8,
        'predicted': 0.6
    }
]
print(evaluate_trust(data_chain))
```

This evaluation function applies a multiplier (weight) to the difference between actual outcomes and predictions, thereby adjusting the trust metric accordingly. Using full CID strings ensures compliance with IPLD requirements where map keys must remain as complete, non-truncated strings.

## Alternative Message Formats

Given the requirements of routing efficiency and secure integrity protection for long chains of probabilistic state decisions, there are several possible message formats:

1. **COSE-Sign1 Standard Format:**  
 A fully conformant COSE_Sign1 object where both protected headers (containing algorithm identifiers, CWT claims such as probabilities and global weights, and issuer information) and the payload (including the decision tree and pinning contract information) are signed together. This structure is ideal for ensuring that message integrity is cryptographically guaranteed before any payload parsing occurs.

2. **Signature-Last Variant:**  
 A variant where the outer envelope begins with the protocol tag (for example, a custom “grid” tag) followed by the protocol CID, with the COSE structure embedded such that the signature appears at the end of the message. This arrangement could improve routing by allowing early extraction of protocol indications while deferring the costlier signature verification until later stages.

3. **Nested State Transition Format:**  
 A layered approach in which a COSE_Sign1 object wraps a claim that itself is a nested structure describing state transitions. For example, the protected header may carry a CWT claim with an array of transitions (each including an event CID, associated probability, and weight), while the payload can hold additional details such as pinning agreements. An example alternative format is:

```cbor
COSE_Sign1(
  protected: <<{
    1: -7,                        // Algorithm: ES256
    15: {                         // CWT Claims
      "prb": {                    // Probability claims for events
        "event1": 0.65,
        "event2": 0.35
      },
      "wgt": 1.2,                 // Global trust weight multiplier
      "iss": "PromiseGridNode/42" // Issuer identifier
    }
  }>>,
  payload: <<{
    "tree": "bafyrei...decisionTreeCID",  // Decision tree root expressed as a full CID string
    "pinning": {
      "commit": "bafyrei...pinContractCID", // Pinning contract CID
      "peers": [
        "bafyrei...peer1",
        "bafyrei...peer2"
      ]
    }
  }>>,
  signature: h'3045...signatureBytes'
)
```

Each of these formats offers tradeoffs between simplicity, extensibility, and the ability to quickly extract routing-relevant data without full parsing. The inline inclusion of the decision tree (expressed with nested lists) allows the agent to verify promises and pinning agreements efficiently.

## Open Questions and Considerations

Several important open questions remain as this design evolves:
- How should negative trust values be interpreted and incorporated into broader trust metrics?
- Would it be more robust to encode probabilities as fixed-point integers (e.g., 6500 for 0.65) rather than floating-point values, given the deterministic encoding requirements of CBOR and DAG-CBOR?
- In practical deployments, particularly on IoT devices with constrained memory, what would be an acceptable limit for the nesting depth of a decision tree without compromising performance?
- How should versioning be managed for CID string formats as new multicodecs or hash algorithms are adopted, ensuring backward compatibility?

This discussion forms the basis for further prototyping and iterative refinement to ensure that the PromiseGrid wire protocol meets both security and efficiency goals.

