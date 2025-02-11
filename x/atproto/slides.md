# Bluesky, ATProto, CWT, IPFS, IPLD, and PromiseGrid
## Steve Traugott

---

## CBOR Web Tokens (CWTs)

### Basic Concepts
- CWTs are compact tokens for representing claims, encoded using CBOR.
- Similar to JWTs but use binary encoding for efficiency.
- Can be secured using CBOR Object Signing and Encryption (COSE).

---

## CBOR Web Tokens (CWTs)

### Structure and Flexibility
- Flexible claim structure allows representation of various data types.
- Can include standard claims (e.g., "iss", "sub", "exp") and custom claims.
- Supports nesting of CWTs for added security or combining claim sets.

---

## CBOR Web Tokens (CWTs)

### Encoding and Key Ordering
- Uses CBOR encoding, which is inherently deterministic.
- Key ordering in CWTs is not mandated by RFC 8392.
- Deterministic encoding rules, including key sorting, are defined in RFC 8949 (CBOR specification).
- RFC-compliant CWTs can have unsorted keys, though sorting is recommended for interoperability.

---

## CBOR Web Tokens (CWTs)

### Use Cases
- Authentication and authorization in constrained environments (e.g., IoT).
- Potential use as capability tokens (explored by IETF ACE working group).
- Possible representation of promises in Promise Theory context (theoretical).

---

## IPLD (InterPlanetary Linked Data)

XXX IPFS history

### Relationship with CWTs
- No direct relationship between IPLD and CWTs/JWTs.
- IPLD is a data model for content-addressable systems.
- CWTs/JWTs are token formats for representing claims.

---

## IPLD (InterPlanetary Linked Data)

### Use in atproto
- Atproto uses a variant of IPLD for its underlying data model.
- Custom JSON encoding used instead of standard DAG-JSON.
- CBOR used for cryptographic operations in atproto.

---

## Promise Theory and CWTs

### Potential Representation
- CWTs could theoretically represent promises from Promise Theory.
- Flexible claim structure allows encoding of promise-related information.
- No inherent support for Mark Burgess' graphical promise notation.

---

## Promise Theory and CWTs

### Considerations
- Promising agent should be represented as the issuer (iss) in CWT.
- Subject (sub) claim could represent the entity about which the promise is made.
- No known widespread use of CWTs/JWTs for Promise Theory representation.

---

## Post-Quantum Cryptography

### Signatures and Hashes
- Hash-based signatures: XMSS, LMS, SPHINCS+.
- Lattice-based signatures: CRYSTALS-Dilithium (ML-DSA), FALCON.
- Multivariate-based signatures: Rainbow (no longer recommended).
- Code-based cryptography.
- Hash functions: SHA-2 and SHA-3 families (with increased output sizes).

---

## IETF ACE and CWTs for Capability Tokens

### Overview
- Authentication and Authorization for Constrained Environments (ACE) working group.
- Exploring CWTs for authorization in IoT scenarios.

---

## IETF ACE and CWTs for Capability Tokens

### Key Features
- Use of CWTs as Proof-of-Possession (PoP) tokens.
- Flexible representation of capabilities and access rights.
- Secured using COSE for integrity and optional encryption.

---

## Addressing Hash Collisions in Decentralized Systems

### Challenges
- Hash collisions are a concern, especially for long-term systems.
- Quantum computing poses potential threats to current cryptographic methods.

---

## Addressing Hash Collisions in Decentralized Systems

### Proposed Solutions
- Multi-hash approach using multiple algorithms simultaneously.
- Hash agility mechanism for algorithm upgrades.
- Merkle trees with multiple hash algorithms.
- Challenge-response verification with metadata CWTs.
- Post-quantum cryptographic signatures.
- Distributed consensus for content verification.
- Regular content validation and migration.

---

## CWTs in Decentralized Computing Systems

### Potential Use
- Could serve as a message format for decentralized systems with message-passing IPC.
- Offers compact representation and cryptographic protections.

---

## CWTs in Decentralized Computing Systems

### Considerations
- Performance overhead for encoding/decoding and cryptographic operations.
- Fixed structure might limit flexibility for complex IPC scenarios.
- Trust model may not align perfectly with all IPC security requirements.

---

## Conclusion

CWTs offer a flexible and efficient token format with potential applications in various domains, from IoT authorization to theoretical representations of promises. While they present interesting possibilities for decentralized systems and long-term data integrity, careful consideration of performance, security, and interoperability is necessary when implementing CWT-based solutions.

---

## Q&A


