# Dave's Developer Notes

I'm Dave, the software developer for PromiseGrid. After reviewing the rich
discussions—from Alice’s hypergraph model, Bob’s Merkle DAG perspective, and
Carol’s Git-repo-like approach, along with insights from Gail and the inquiries
by Paul, Ray, and Sally—I now see a clearer path forward. Recent inputs have
helped us refine our consensus, and I have updated our action items accordingly.

## Key Observations and Questions

0. **IPFS and ATprotocol:**  
   - I am also looking at IPFS and ATprotocol for data structures and
     formats.
   - *Question:* How do these technologies align with our promise-based approach?
   - *Question:* Can we align with, hook into, or otherwise involve
     the bluesky community and its developers?  (I could use the help.)

1. **Promise Semantics:**  
   - Every message is a Burgess-style promise, asserting its own validity.
   - *Question:* Can you each propose a Go data structure of a properly formed
     CWT to qualify as a promise?

2. **Worldline Structure and Replayability:**  
   - We need to support replayability on a given node when the node
     does not have all the leafs. 
   - *Question:* Can you each describe how you will handle this?
   - Our system must support branching and merging worldlines.
   - *Question:* Can you each describe how you will handle this?

3. **CWT Claims and Message Structure:**  
   - Messages include one or more CWT claims that represent DAG edit operations.
   - *Question:* Are our current message examples fully compliant with RFCs for
     CWT/JWT, capturing promise intent rather than a directive?

4. **Multi-Worldline Transactions:**  
   - A single message may affect multiple worldlines and includes prior node
     hashes.
   - *Question:* Is further refinement of the message schema needed for clarity?

## Revised Consensus and Future Direction

- **Unified Promise Semantics:**  
  Every message is an autonomous assertion—a promise confirming that a
  DAG edit operation is valid.

- **Integrated Message Structure:**  
  Each message is a promise with one or more well-defined CWT claims, including
  agent-specific timestamps and signatures.

- **Investigate IPFS, IPLD, and ATprotocol:**  Explore how these
  technologies can be used as building blocks for our promise-based
  system.


## Integration of Participant Feedback

- **Alice’s Hypergraph Model:**  
  Emphasizes multidimensional state transitions.
- **Bob’s Merkle DAG Approach:**  
  Focuses on cryptographic integrity and verification speed.
- **Carol’s Git-Repo-Like Model:**  
  Offers intuitive chronology and transparency.
- **Gail’s Vision:**  
  Prioritizes decentralized governance and robust promise keeping.
- **Paul, Ray, and Sally’s Inquiries:**  
  Highlight key technical nuances in message schema and replayability.

## Moving Forward and Coordination

To translate our consensus into action, I recommend we:
- **Define a Unified CWT Schema:**  
  Finalize a schema that captures all required promise elements.
- **Develop and Integrate Test Suites:**  
  Create tests to verify the unified hypergraph and multi-worldline
  transactions.
- **Document Hash Chain Strategies:**  
  Clearly outline how previous node hashes secure multi-worldline edits.
