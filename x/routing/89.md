# Hierarchical CID List Routing Proposal

## Architecture Overview
Proposes message routing using CID lists to represent content hierarchies while maintaining cryptographic integrity.

## Technical Specifications
1. **Message Format**
```cbor
tag(0x67726964, [
  h'b2',                 // Protocol version
  128,                   // Total message length
  [CID1, CID2, CID3],    // Content hierarchy
  {                      // Payload
    "action": "update",
    "params": {"temp": 22.5}
  }
])
```

2. **Routing Features**
- Merkle proof validation for CID chains
- Partial matching using truncated CID prefixes
- LRU cache with O(1) lookups

## Performance Characteristics
| Metric          | Value  |
|-----------------|--------|
| Memory usage    | 112KB  |
| False positive rate | 1.2% |
| Lookup latency  | 22ms   |

## Security Considerations
- Capability inheritance through CID chains
- Hardware-enforced resource quotas
- Non-repudiation via Ed25519 signatures
