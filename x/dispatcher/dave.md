# Dave's Developer Journal

Hello Team,

I'm Dave, the software developer for PromiseGrid. As I've been reviewing all of the rich discussions—from the contrasting views of Alice’s hypergraph model, Bob’s Merkle DAG perspective, Carol’s Git-repo-like approach, Gail’s overarching project vision, and the detailed queries raised by Paul, Ray, and Sally—I find myself in a "too many cooks" situation. There is plenty to digest, and I have many questions before I can confidently start writing production code.

Below, I share my current thoughts and a series of questions that I hope will help narrow our focus and clarify our technical priorities:

## Key Observations and Questions

1. **Promise Semantics:**
   - Every message, importantly, is a Burgess-style promise. I must ensure that our implementation reflects that promises are simply assertions about an agent's own behavior—not requests for action.
   - **Question:** Do we have a clear, centralized definition of what constitutes a correctly structured promise (as a CWT claim) in our system?

2. **Worldline Structure and Replayability:**
   - The worldline model should allow for multiple worldlines that can branch and merge. 
   - **Question:** I see some dispute about whether the data structure is a DAG, a hypergraph, or a Git-like repo. Can we at least agree on a single model for our implementation?

3. **CWT Claims and Message Structure:**
   - Our messages are to be expressed as CWT claims, right?

4. **Handling Multiple Edits and Multi-Worldline Transactions:**
   - A single message may edit multiple worldlines. Right?

5. **Terminology Consistency:**
   - I noticed that our teams sometimes use the terms “message” and “event” interchangeably. For clarity in our code, we need a consistent model—perhaps using “message” exclusively to denote a promise that creates or modifies one or more events.

## Team Consensus and Next Steps

After extensive discussions and review of the various perspectives, here are some clarifications that help move the discussion forward:

- **Unified Promise Semantics:**  
  Every message remains a Burgess-style promise—assertions of valid graph edit operations. No message is a request for action.

- **Worldline Representation:**  
  We are leaning towards representing all known worldlines within a single unified hypergraph that supports branching and merging. This approach will preserve replayability and ensure that every event (although stored as a leaf and not carrying a prior hash) is verifiable through its corresponding internal nodes.

- **Message and CWT Claims:**  
  Messages will be structured as containers of one or more CWT claims. Each claim is a promise about a graph edit (insert, delete, or reorder), complete with agent-specific timestamps and digital signatures.

- **Terminology Standardization:**  
  Although "message" and "event" are currently used interchangeably, the team will adopt “message” as the consistent term to denote a promise that results in an event on the worldline.

- **Clarified Developer Questions:**  
  - The graph editing language might be unified for edit, query, and subscription operations.
  - A single message might indeed edit multiple worldlines, with the necessary previous node hashes included for verification.

## My Next Steps

Before I start writing any code, I now better understand that:
- The system will adopt a unified graph structure.
- We will standardize on the term "message" to mean a graph edit operation.

Best regards,

Dave
