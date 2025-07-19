# PromiseGrid Minimal Viable Implementation

This document describes the simplest functional implementation of PromiseGrid
that satisfies core requirements while supporting containerized agents. The
design prioritizes kernel mediation, message-based interaction, and localized
matching to enable cooperative work between autonomous agents.

## Kernel Architecture

The kernel functions as universal mediator with three responsibilities: message
routing between agents, transaction escrow services, and double-entry ledger
management. Agents communicate exclusively through the kernel using CBOR-
encoded messages, maintaining no direct peer awareness. This client-server model
ensures all promises are kernel-recorded with cryptographic non-repudiation.
Containerized agents interact via Unix domain sockets, allowing Dockerized
implementations while preserving isolation[1][5].

## Message Structure

All interactions use Concise Binary Object Representation (CBOR) with standard
schema:
```cbor
{
  "type": "BID/ASK",       // Message classification
  "source": "agent_hash",   // Sender identifier
  "target": "service_hash", // Service endpoint identifier
  "promise": {              // Commitment details
    "compensation": 10,     // Offered currency amount
    "currency": "AGENT_A"   // Personal currency type
  },
  "signature": "crypto_sig" // Sender authentication
}
```
ASK messages establish service endpoints while BID messages initiate
transactions, following Mach port semantics where ASK hashes function as
receiving mailboxes[1][4][7].

## Matching Algorithm

The kernel implements localized price-time priority matching:

### Order Processing Sequence
1. **Order Receipt**: Kernel validates message signatures
2. **Order Classification**: ASK orders create service endpoints, BID orders
   queue as requests
3. **Priority Sorting**: BID orders sort by descending offer value then
   ascending arrival time
4. **Matching**: Highest value BID matched against earliest compatible ASK
5. **Escrow Activation**: Compensation reserved from bidder account
6. **Service Activation**: Provider notified to execute service
7. **Settlement**: Results verified → compensation released[2][6]

## Economic Model

The kernel maintains double-entry bookkeeping for all agents:

### Accounting Entries
```
BID Submission:
  Debit:  Bidder::Escrow
  Credit: Bidder::Liability

ASK Fulfillment:
  Debit:  Bidder::Liability
  Credit: Provider::Revenue
  Debit:  Provider::Asset
  Credit: Bidder::Escrow
```
Broken promises trigger reputation debits while successful completions
enhance creditworthiness[3][8].

## Agent Interaction Flow

### Service Publication
1. Alice creates ASK message: "Perform X for 10 ALICE points"
2. Kernel hashes service description → service_hash
3. ASK stored as endpoint (Mach port)

### Service Request
1. Bob creates BID targeting service_hash with 10 ALICE points
2. Kernel matches highest priority BID to ASK
3. Alice performs service X, returns result to kernel

### Settlement
1. Kernel verifies result against service description
2. Ledger updated:
   - Alice: +10 ALICE points (Revenue)
   - Bob: -10 ALICE points (Liability)
3. Result delivered to Bob

## Container Integration

Dockerized agents access kernel via:
```bash
docker run -v /var/promisegrid/socket:/socket agent_image
```
The kernel socket exposed in host-mounted volume enables communication while
preserving container isolation. Agent implementations require only CBOR message
construction capability and cryptographic signing functions[1][5].

## Limitations and Future Work

This minimal implementation assumes single-node deployment. Distributed
operation would require kernel-to-kernel synchronization protocols.
Additionally, the reputation system and currency exchange mechanics remain
simplified. Future versions should integrate content-addressable storage
for enhanced security and verifiable computation[1][8].
