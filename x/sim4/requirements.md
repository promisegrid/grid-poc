
- a security capability token is a promise that an agent will perform
  a specific action, such as sending a message or performing a
  computation
- a token is like a reservation for a resource or service, where the
  token issuer is promising to provide the resource or service at a
  later time
- issuing a token creates a liability for the issuer, as they are
  promising to perform an action or computation in the future
- receiving a token creates an asset for the recipient, as they are
  promised the future action or computation
- the issuing of a token via a kernel creates a Mach-like receive
  port, which allows the agent to receive messages or requests for
  computation
- tokens include data fields that uniquely identify the token issuer
  and the function or action that the token represents
- presenting a token to a kernel is similar to sending a message to a
  Mach-like receive port, where the kernel routes the token to the
  appropriate agent based on the issuer field
- a token is similar in concept to a CWT
- tokens are issued or transferred by appending a hyperdege (a
  transaction) to the hypergraph
- the hypergraph is a directed acyclic graph (DAG) that represents the
  state of the universe
- the hypergraph is never complete
- no agent ever holds a complete copy of the hypergraph
- the hypergraph supports speculative execution, where agents can
  attempt to spend the same token in multiple transactions, each on a
  different branch of the hypergraph
- each node or edge in the hypergraph includes a hash of the previous
  node or edge
- the economy of the system is based on trading tokens, which are
  promises to perform actions or computations
- the trading of tokens establishes exchange-rate-like relative values
  for the actions or computations represented by the tokens, and
  relative reputations of the agents who issue the tokens
- every hyperedge is a balanced transaction that represents a
  bidirectional transfer of value between two agents
- every hyperedge is a swap of tokens along with any associated data
- nodes do not contain results -- a node is a state symbol
  representing the current state of part of the universe at a specific
  point in time; it is similar to the state symbol in the rules of a
  Turing machine; the Turing state symbol is not the state itself --
  the state lives on tape; the difference between a Turing machine and
  the PromiseGrid hypergraph is that the hypergraph has unlimited
  states, while a Turing machine has a finite number of states
- when an agent wants to redeem a token, they include it in a
  hyperedge (transaction) that is appended to the hypergraph, along
  with any additional data needed to perform the action or
  computation; they also include a reply token (port) that the executor will
  use to return the result of the action or computation
- it is in the best interests of a token issuer to ensure that the
  action or computation is performed, as they have a liability to the
  token holder; they will want to spend the reply token in order to
  return the result of the action or computation to the token holder,
  because that is how they discharge their liability
