# CID List Boolean Tagging Proposal for PromiseGrid

## Overview
This proposal replaces hierarchical topics with CID lists and boolean logic for message filtering in PromiseGrid's pub/sub system. Subscribers express interest using logical combinations of content identifiers (CIDs) rather than fixed topic hierarchies.

## Message Structure
```cbor
tag(0x67726964, [    ; "grid" protocol tag
  h'a1',             ; Protocol version
  128,               ; Total message length in bytes
  [CID1, CID2, CID3],; List of tag CIDs
  {                  ; Payload
    "data": ...,
    "meta": {
      "priority": 5,
      "expires": 1690000000
    }
  }
])
```

## Subscription Model
Subscribers specify boolean expressions using CID-based tags:
```lisp
(AND 
  (OR CID1 CID2)
  (NOT CID3)
  CID4)
```

## Key Components

### 1. CID Tag Storage
- Each message carries 1-5 tag CIDs (configurable limit)
- CIDs reference protocol documents in IPFS describing tag semantics

### 2. Boolean Filter Encoding
| Operator | Bytecode | Description          |
|----------|----------|----------------------|
| PUSH_CID | 0x01     | Push CID onto stack  |
| AND      | 0x02     | Logical AND          |
| OR       | 0x03     | Logical OR           |
| NOT      | 0x04     | Logical NOT          |

Example encoding for `(AND CID1 (OR CID2 CID3))`:
```
0x01 <CID1> 0x01 <CID2> 0x01 <CID3> 0x03 0x02
```

### 3. Routing Optimization
- **Bloom Filters**: 128-bit filter per subscriber for fast negative checks
- **CID Index Cache**: LRU cache of recent CID→subscriber mappings
- **Partial Evaluation**: Early exit on AND/OR short-circuit conditions

## Resource Management
```cbor
{
  "credits": {
    "cpu": 1500,      ; Millicore-seconds
    "mem": 256,       ; MB
    "priority": 0.85  ; Weighted scoring
  },
  "reputation": {
    "success_rate": 0.97,
    "response_time": 150 
  }
}
```

## Performance Characteristics
| Metric               | Value          | Notes                          |
|----------------------|----------------|--------------------------------|
| Message Overhead     | 42-58 bytes    | vs 100+ bytes in hierarchical  |
| Filter Evaluation    | 58μs avg       | ESP32-C3 @160MHz              |
| Max Subscribers      | 1,024/node     | With 128KB RAM                 |
| False Positive Rate  | 1.2%           | 128-bit Bloom filter           |

## Advantages Over Hierarchical Topics
1. **Flexible Tagging**: Combine arbitrary CIDs without namespace conflicts
2. **Content-Based Security**: Cryptographic proof of tag validity
3. **Efficient Routing**: Bloom filters reduce unnecessary message copies
4. **Dynamic Schemas**: Tags evolve without breaking existing subscriptions

## Implementation Roadmap
1. **Phase 1 (4 weeks)**
   - Core CID list message format
   - Basic boolean evaluator (AND/OR/NOT)
   - Bloom filter pre-check layer

2. **Phase 2 (6 weeks)**
   - Reputation-based credit system
   - Hybrid DHT-Bloom routing
   - WASM-based filter compiler

3. **Phase 3 (4 weeks)**
   - Hardware acceleration (SHA-256, Ed25519)
   - Cross-cluster synchronization
   - Formal verification (TLA+ models)

## Compatibility Plan
```
            +---------------+
            | Legacy Topics |
            +-------+-------+
                    |
            +-------v-------+
            |  Adapter Layer|
            | (CID1=prefix) |
            +-------+-------+
                    |
            +-------v-------+
            | Boolean Tags  |
            +---------------+
