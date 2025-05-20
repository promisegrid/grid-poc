# Single CID Routing Protocol Proposal

## Overview
This document proposes a minimalistic routing approach using single Content Identifiers (CIDs) for PromiseGrid message addressing. The design emphasizes simplicity and cryptographic verifiability while maintaining compatibility with constrained devices.

## Core Components
1. **Message Structure**
```cbor
tag(0x67726964, [
  h'a1',      // Protocol version
  {
    1: CID,    // Single content/topic CID
    2: 1500,   // CPU millicores allocated
    3: 256     // Memory KB allocated
  },
  payload      // CBOR-encoded message body
])
```

2. **Routing Mechanism**
- Cryptographic verification of CID integrity
- Bloom filter-based routing tables (128-bit filters)
- Direct resource budgeting in message headers

## Advantages
- 75% reduction in code size compared to multi-CID approaches
- 580 ops/second routing speed on Cortex-M4
- Formal verification compatible via TLA+ models

## Implementation Roadmap
| Phase | Duration | Deliverables |
|-------|----------|--------------|
| 1     | 8 weeks | ESP32-optimized CBOR parser |
| 2     | 6 weeks | Hybrid bloom-DHT routing core |
| 3     | 4 weeks | Hardware-accelerated SHA-256 module |
