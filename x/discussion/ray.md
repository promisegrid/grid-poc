# Ray's Discussion Script

Hi everyone,

I’ve reviewed the available documentation and discussion papers for our messaging
system, and I’d like to clarify several points to ensure that the requirements are
correctly documented. Below are my observations and questions:

1. **Promise Semantics (Burgess-Style):**
   - Each DAG edit message is a promise following Promise Theory—an agent can only
     promise about its own behavior, and these promises are assertions that the
     corresponding DAG edit operation is valid.
   - **Question:** Can we confirm that every CWT claim included in a message is a
     promise (an assertion of validity) rather than a request for action?

2. **Replayability and DAG Integrity:**
   - The system is designed so that messages (edit operations) can be replayed to
     reconstruct the Merkle DAG representing the worldline(s) at any given time.
   - **Question:** Are we planning to maintain a separate DAG per worldline, or will
     there be a single unified DAG representing all known worldlines? Additionally,
     how should branching and merging be handled in either approach?

3. **Message Structure and CWT Claims:**
   - Messages, which are also referred to as events, are expressed as structured
     CWT (or JWT) claims. These messages include one or more promises that declare
     insert, delete, or reorder operations.
   - **Question:** Could someone confirm whether the current message examples fully
     comply with the relevant RFCs for CWT/JWT? Any additional details regarding
     the proper structure would be helpful.

4. **Signature, Timestamps, and Security:**
   - Each message is signed by the agent that created the event and includes a
     timestamp from that agent’s perspective.
   - **Question:** Are there specific cryptographic algorithms or additional security
     protocols mandated for signature verification that we should incorporate in our
     documentation and reference implementations?

5. **Payload Consistency and Terminology:**
   - The payload of each message is a set of one or more PT-style promises. However,
     there’s been an interchange in terminology between “message” and “event.”
   - **Observation:** I recommend that we standardize our terminology (for example,
     consistently use “message” to refer both to the DAG edit operation and the
     resulting event) to avoid confusion.
   - **Question:** Should we explicitly document that a single message may edit multiple
     worldlines and how these edits are recorded within the Merkle DAG?

6. **Handling of Inaccessible Nodes:**
   - Documentation indicates that not all referenced leaves or inner nodes are
     guaranteed to be accessible at runtime.
   - **Question:** What should the expected behavior be when certain DAG nodes (or
     their hashes) are unavailable during replay? Do we log these separately, or is
     there a recovery strategy in mind?

7. **Alternative Structures and Future Exploration:**
   - While a conventional Merkle DAG is the current model, there’s also exploration
     into alternative structures that might provide improved flexibility or
     efficiency.
   - **Question:** Are there any design alternatives already under consideration? If so,
     how do their advantages compare against the Merkle DAG approach, particularly in
     relation to auditability and replay ability?

## Clarification and Feedback

To help resolve these issues, I propose we:

- Initiate a joint review of both Merkle DAG and alternative proposals to
  determine the best path forward.

Best regards,  
Ray
