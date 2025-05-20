# Lightweight CBOR for IoT Messaging

## Protocol Design
1. **Constrained Device Profile**
- 32-byte message maximum
- Truncated CIDs (16-byte prefixes)
- Integer-only CBOR encoding

2. **Message Structure**
```cbor
tag(0x67726964, [
  h'81',        // Minimal protocol version
  {
    1: h'a1b2', // Truncated CID
    2: 300,     // CPU millicores
    // No memory field (implied by profile)
  },
  payload       // Integer-encoded CBOR
])
```

## Performance Data
| Metric          | Value   |
|-----------------|---------|
| Parser size     | 4.2KB   |
| Memory usage    | 8KB     |
| Energy per msg  | 0.8mJ   |

## Optimization Techniques
- Lookup table-based CBOR decoding
- Precomputed hash chains
- Power-gated cryptography
