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
   - **Question:**  I see some dispute about whether the data
     structure is a DAG, a hypergraph, or a Git-like repo. Can we at
     least agree on a single model for our implementation? 

3. **CWT Claims and Message Structure:**
   - Our messages are to be expressed as CWT claims, right?

4. **Handling Multiple Edits and Multi-Worldline Transactions:**
   - A single message may edit multiple worldlines.  Right?

6. **Terminology Consistency:**
   - I noticed that our teams sometimes use the terms “message” and “event” interchangeably. For clarity in our code, we need a consistent model—perhaps using “message” exclusively to denote a promise that creates or modifies one or more events.

## My Next Steps

Before I start writing any new modules, I need confirmation on the above points. My aim is to implement:
- A robust message parser/validator that processes CWT claims correctly.
- A thread-safe store to support rapid appends, deletions, and reordering.
- A replay mechanism that can accurately reconstruct the worldline(s) from a sequence of messages.
- A design that allows a single message to impact multiple worldlines without violating the promise model.

I look forward to your input on these questions. Once we have a consensus on these design decisions, I can begin to draft the low-level modules and integrate them into PromiseGrid.

Best regards,

Dave
