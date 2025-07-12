# Sim2 Design Document: Double-Auction Routing Simulation

This document describes the design and protocol for Sim2, a
simulation of a decentralized, double-auction based routing mechanism
for PromiseGrid. The goal is to fairly allocate resources and prevent
spam using a double-auction market. In this protocol, agents exchange
bids and asks using personal currencies, and every message carries a
currency symbol and amount.

## Protocol Overview

Sim2 employs a peer-to-peer message passing system in which all
communication is local. Agents send messages directly to their
neighbors as defined by their peer connections. In our simulation the
agents are Alice, Bob, Carol, and Dave, with the following network:
- Alice, Bob, and Carol are direct peers.
- Bob, Carol, and Dave are direct peers.
- Alice and Dave cannot communicate directly but route messages via
  intermediate peers.

Each message represents either a bid or an ask. The buyer (Alice)
initiates a BID, and the seller (Dave) responds with an ASK. 

## Personal Currencies and Economic Incentives

Every agent has its own personal currency: "ALICE", "BOB", "CAROL", and
"DAVE". These currencies are used in bid and ask messages. An agent’s
balance is adjusted when a trade is executed. 

Agents trade each others' currencies as a means of valuing relative
trustworthiness; for example, if Alice trusts Bob, she may buy some of
Bob's currency to use in future transactions.  Unrustworthy or abusive
behavior would cause other agents to sell that agent's currency,
reducing its value.

Intermediate agents don't forward messages; they instead use
arbitrage. They create their own mid or ask orders based on received
messages, effectively acting as market makers or arbitrageurs. This
allows them to profit from the difference between the bid and ask
prices, incentivizing their participation in the network.

Here's and example of how this works in practice:
- Alice, Bob, Carol, and Dave all start with zero balances in all
  currencies.
- Alice wants to pay 10 ALICE for the results of a function call that
  only Dave can perform. She sends a BID message to Bob, who is her
  direct peer.  The BID message contains the amount (10 ALICE), the
  currency symbol ("ALICE"), and the function call f(2, 3).
- Bob receives the BID and thinks that he might profit by acting as an
  intermediary.  He sends a new BID message to Carol, promising to pay
  her 9 BOB for the same function call f(2, 3). The BID message
  contains the amount (9 BOB), the currency symbol ("BOB"), and the
  function call f(2, 3).
- Carol receives Bob's BID and decides to act as an intermediary as
  well. She sends a new BID message to Dave, promising to pay him 8
  CAROL for the same function call f(2, 3). The BID message contains
  the amount (8 CAROL), the currency symbol ("CAROL"), and the
  function call f(2, 3).
- Dave receives Carol's BID and decides to accept it.  He performs the
  function call f(2, 3) and calculates the result, which is 5. He then
  sends a CONFIRM message back to Carol, including the result (5).
- Carol receives Dave's CONFIRM and sends a CONFIRM message back to
  Bob, including the result (5).
- Bob receives Carol's CONFIRM and sends a CONFIRM message back to
  Alice, including the result (5).
- Alice receives Bob's CONFIRM and now knows the result of the
  function call f(2, 3) is 5.
- Alice now owes 10 ALICE to Bob, who owes 9 BOB to Carol, who owes
  8 CAROL to Dave.  Each agent maintains double-entry accounting to
  track these debits and credits in the proper asset and liability
  accounts.

## Simulation Details

The simulation is implemented in Go and encapsulated in a single file,
sim2.go, with testing in sim2_test.go. Key design aspects include:

- **Peer-to-Peer Exchange:** There is no central authority. Instead,
  messages propagate through a network of direct connections.

## Implementation Considerations

The design emphasizes minimalism and dependency-free code that can be
deployed even on resource-constrained devices such as IoT platforms.
The simulation uses in-memory data structures and simple string slicing
for routing. Message histories prevent loops, and the forwarding
mechanism ensures that each agent processes a message only once.

## Conclusion

Sim2 demonstrates a market-based approach to decentralized routing using
a double-auction mechanism. It shows how economic principles like
price discovery and auction matching can be applied to networking to
achieve fair resource allocation and spam prevention. This design is
intended to serve as a proof-of-concept for integrating economic models
into multi-hop, peer-to-peer communication protocols in PromiseGrid.
