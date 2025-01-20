# grid-poc

Experimental proof of concept for
[grid-cli](https://github.com/stevegt/grid-cli) and eventually
[promisegrid](https://github.com/promisegrid/promisegrid). Will
influence the content of [the
whitepaper](https://github.com/promisegrid/paper-ism) and vice versa.



## References:

- [CBOR](https://cbor.io/)
- [Radix Tree](https://en.wikipedia.org/wiki/Radix_tree)
- [Hypergraph](https://en.wikipedia.org/wiki/Hypergraph)
- [Merkle Tree](https://en.wikipedia.org/wiki/Merkle_tree)
- [Merkle DAG](https://docs.ipfs.tech/concepts/merkle-dag/)

## Possible POC tentative goals:

- Maybe POC that just takes data on input and spits out CBOR or vice versa
- Maybe POC that takes CBOR in, finds and executes method in transition function graph, spits out CBOR
- Maybe POC that takes an event, generates CBOR hyperedge, emits an event
- Maybe POC that takes an event, stores CBOR hyperedge, emits an event
- Maybe POC that emulates a VCS, with CBOR hyperedges as commits.
