# PromiseGrid Wire Protocol Compliance Test Cases

### CID Format Compliance
1. Are all CIDs used as map keys in DAG-CBOR structures represented as full base32-encoded strings rather than binary forms with CBOR tag 42? [wire.md §2.2][wire-chat.md §Identity Multibase]
2. Does the document prohibit use of truncated CID representations in map keys while allowing native binary CIDs (CBOR tag 42) in non-key positions? [wire.md §2.2][wire-chat.md §Identity Multibase]

### Data Structure Validation
3. Does the decision tree structure exclusively use nested lists instead of maps for state transitions to comply with DAG-CBOR requirements? [wire.md §3.2][wire-chat.md §Decision Tree]
4. Are all event and state identifiers in decision trees represented as complete CID strings rather than binary forms? [wire.md §3.2][wire-chat.md §Alternative Formats]
5. Does the protocol avoid using floating-point NaN/Infinity values in probability fields per DAG-CBOR constraints? [wire.md §2.1]

### Cryptographic Integrity
6. Is the COSE_Sign1 signature computed over the concatenated Sig_structure containing protected headers, external_aad (if present), and payload? [wire.md §3.3][wire-chat.md §COSE Message Fields]
7. Does the protected header include both algorithm identifiers (-7 for ES256) and CWT claims in its CBOR structure? [wire.md §3.3][wire-chat.md §COSE-Sign1]

### Trust Metric Implementation
8. Does the trust calculation formula prevent multiplication by negative values through use of absolute differences or signed weighting? [wire.md §4.3][wire-chat.md §Trust Metric]
9. Is weight multiplier application limited to subsequent branches rather than parent nodes in the decision tree? [wire.md §3.2][wire-chat.md §Decision Tree]

### Message Envelope Structure
10. Does the outer message envelope contain both protocol tag (0x67726964) and protocol version CID? [wire.md §3.1]
11. Are all message-layer CIDs represented as full binary forms using identity multibase when not acting as map keys? [wire.md §2.2][wire-chat.md §CID Representation]

### Probability Handling
12. Does the document specify fixed-point integer encoding for probabilities to ensure deterministic hashing? [wire-chat.md §Open Questions]
13. Are all probability values constrained to 0.0-1.0 inclusive with validation mechanisms? [wire.md §3.2][wire-chat.md §Decision Tree]

### Routing Efficiency
14. Can routing nodes extract protocol tag and version CID without fully parsing nested decision trees? [wire.md §3.1][wire-chat.md §Alternative Formats]
15. Does the message structure enable partial parsing of decision tree headers before signature verification? [wire-chat.md §Signature-Last Variant]

### Pinning Agreements
16. Do pinning contract references in payloads use full CID strings rather than binary forms? [wire.md §3.3]
17. Are pinned peer identifiers represented as complete base32-encoded CID strings? [wire.md §3.3]

### Versioning & Compatibility
18. Does the protocol document specify CIDv1 as mandatory minimum version? [wire.md §4.5]
19. Is there a defined strategy for backward compatibility with future CID versions? [wire.md §4.5]

### Constrained Devices
20. Does the specification recommend maximum recursion depth limits for decision trees? [wire.md §4.4][wire-chat.md §Open Questions]
21. Are there size constraints on external_aad fields to prevent memory exhaustion? [wire.md §2.3]

### Negative Trust Handling
22. Does the document define policy for agents with negative net trust values? [wire-chat.md §Open Questions]
23. Are there safeguards against trust metric underflow/overflow in continuous operation? [wire.md §4.3]

### Signature Positioning
24. Is the signature placement optimized for routing efficiency while maintaining cryptographic integrity? [wire-chat.md §Alternative Formats]
25. Does the envelope structure allow late-binding signature verification without payload parsing? [wire-chat.md §Signature-Last]
