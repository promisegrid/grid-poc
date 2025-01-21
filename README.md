# grid-poc

Experimental proof of concept for
[grid-cli](https://github.com/stevegt/grid-cli) and eventually
[promisegrid](https://github.com/promisegrid/promisegrid). Will
influence the content of [the
whitepaper](https://github.com/promisegrid/paper-ism) and vice versa.

## Concepts

- Suppose all data in the universe, past and future, has already been
  computed.  How might it be stored and accessed in a decentralized
  system?
- Assume that some data has yet to be computed, but computation
  results can be returned by the same access method that is used to
  retrieve existing data.  How might that be implemented? 
- Assume that a 256-bit number is approximately large enough to
  uniquely address every atom in the observable universe, or that a
  512-bit number is approximately large enough to uniquely address
  every atom in 256 universes.
- Assume that large cryptographic hash functions can reasonably be
  expected to be able to generate a unique identifier for any data in
  the Solar System if we are able to safely handle a small number of
  collisions.
- We can increase collision safety by using an access method that
  incorporates a path of identifiers to reach a target identifier.
  This path might describe a sequence in space, time, or semantics.
- The path may include literals, hashes, or other identifiers.
- We can use the path as an access method for either existing or
  computed data.  For example, we might use a path of one or more
  identifiers that describe a function, followed by zero or more
  parameters that are inputs to the function.
- The result of a function can be appended to a path; the entire path
  can be used to describe a computation including function,
  parameters, and result.  
- The set of all paths that describe all data in the universe can be
  modeled as a hypergraph, where each node is the state of some part
  of the universe, and each edge is a computation that takes one or
  more nodes as input (edge tails) and produces one or more nodes as
  output (edge heads).
- The hypergraph can be thought of as an N-dimensional space where one
  dimension is time and points in the remaining dimensions are
  addressable by paths.  
- The hypergraph can be projected into a 3-dimensional space where the
  X/Y plane is addressable by paths and the Z axis is time.
- Paths can be stored and accessed using a variety of methods.
  Candidates include radix trees, Merkle trees, and Merkle DAGs.
- If the function is pure and observes referential transparency, then
  a sequence can be stored and subsequently accessed without
  recomputation.
- If the function is impure, then subsequent access may generate new
  results.  Each new result creates a branch in the hypergraph.
- If an agent receives multiple results for the same path, it needs to
  be able to choose which result to use.  In some cases, the agent may
  have no ability to inspect the results, and needs some other metric
  to choose between them.
- Assume a decentralized open market for exchange of "personal
  currencies" -- these are tokens or points that any agent can issue
  in any quantity.
- Any agent can buy or sell tokens in the open market, causing
  exchange rates to move according to supply and demand.
- Agents promise to perform data retrieval or computation in exchange
  for tokens in one or more given denominations.
- An agent that fulfills its promises to provide data or computation
  services in return for payment in its own currency can expect the
  value of its currency to rise.
- An agent that fails to fulfill its promises can expect the value of
  its currency to fall.
- We postulate that agents will act in their own self-interest, and
  that the value of an agent's currency will be a measure of the
  agent's trustworthiness.

## References

- [CBOR](https://cbor.io/)
- [Radix Tree](https://en.wikipedia.org/wiki/Radix_tree)
- [Hypergraph](https://en.wikipedia.org/wiki/Hypergraph)
- [Merkle Tree](https://en.wikipedia.org/wiki/Merkle_tree)
- [Merkle DAG](https://docs.ipfs.tech/concepts/merkle-dag/)

## Possible POC tentative goals

- Maybe POC that just takes data on input and spits out CBOR or vice versa
- Maybe POC that takes CBOR in, finds and executes method in transition function graph, spits out CBOR
- Maybe POC that takes an event, generates CBOR hyperedge, emits an event
- Maybe POC that takes an event, stores CBOR hyperedge, emits an event
- Maybe POC that emulates a VCS, with CBOR hyperedges as commits.
