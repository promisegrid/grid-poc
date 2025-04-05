# PromiseGrid Wire Protocol Compliance Test Cases

### CID Format Compliance
1. Are all CIDs used as map keys in DAG-CBOR structures represented as
   full base32-encoded strings rather than binary forms with CBOR tag
   42? 
2. Does the document prohibit use of truncated CID representations in
   map keys while allowing native binary CIDs (CBOR tag 42) in non-key
   positions? 

### Data Structure Validation
3. Does the decision tree structure exclusively use nested lists
   instead of maps for state transitions to avoid needing to use 
   string map keys?
4. Are all event and state identifiers in decision trees represented
   as binary forms? 
5. Does the protocol avoid using floating-point NaN/Infinity values in
   probability fields per DAG-CBOR constraints? 

### Trust Metric Implementation
9. Is weight multiplier application limited to subsequent branches
   rather than parent nodes in the decision tree? 

### Message Envelope Structure
10. Do the leading bytes of the message contain protocol tag, 
    protocol version CID, and the sequence of events and states?
11. Are all CIDs represented as binary encodings?

### Probability Handling
12. Does the document specify fixed-point integer encoding, rather than float, to ensure deterministic hashing? 
13. Are all probability values constrained to 0.0-1.0 inclusive with validation mechanisms? 

### Routing Efficiency
14. Can routing nodes extract protocol CID, and the beginning of the
    sequence of events and states without fully parsing the message?
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
