# Sally's Discussion Script: Crafting Testable User Stories

Hello Team,

I'm Sally, and I'm working on translating our complex technical requirements into
clear, testable user stories. I've reviewed the available documents—covering messaging,
subscriptions, worldlines, and various position papers—and I'd like to share my
observations, propose some user story drafts, and ask a few clarifying questions.

## Key Insights Gathered

1. **Promise as a Commitment (Burgess-Style):**
   - Each message in our system is a DAG edit operation that serves as a promise.
   - According to Promise Theory, an agent can only make a promise about its own
     behavior; such promises are assertions of validity, not requests for action.
   - *User Story Consideration:* Stories should capture this autonomy. For example, a
     story might state, “As an agent, I promise that the event I create is valid,
     verifiable, and replayable within the system.”

2. **Replayability and DAG Integrity:**
   - Our design ensures that messages (or events) can be replayed to reconstruct the
     Merkle DAG, which in turn represents the state of one or more worldlines.
   - The DAG allows for operations such as insertion, deletion, and reordering of
     events.
   - *User Story Example:* “As a system operator, I want to replay the history of
     messages so that I can verify the integrity and evolution of the worldline over
     time.”

3. **Message Structure and CWT Claims:**
   - Each message is structured with one or more CWT (or JWT) claims, where each claim
     represents an edit (insert, delete, reorder) on the DAG.
   - Messages include signatures, timestamps (from the agent’s perspective), and the
     hashes of previous internal DAG nodes to support validation and replay.
   - There is some ambiguity between the terms “message” and “event” that we need to
     resolve.
   - *User Story Example:* “As a developer, I need unambiguous terminology in our
     documentation to differentiate between a ‘message’ and an ‘event’ so that all edits
     and promises are consistently understood.”

4. **Worldline Structure and Branching/Merging:**
   - The system supports multiple worldlines which can branch and merge, with messages
     potentially editing more than one worldline.
   - The design choice of either a unified DAG versus separate DAGs for each worldline
     remains under discussion.
   - *User Story Example:* “As an administrator, I want the ability to manage and verify
     edits in worldlines that branch and merge, ensuring all changes are reflected correctly
     in the Merkle DAG.”

5. **Security, Signatures, and Testability:**
   - Every message/event is signed by its creating agent and marked with a timestamp,
     establishing a secure and verifiable history.
   - Testable assertions should include verifying that the right cryptographic protocols
     are used and that each promise is truly an assertion (not a directive).
   - *User Story Example:* “As a security auditor, I want each message’s signature and
     timestamp to be validated automatically, ensuring that only authenticated and
     authorized DAG edits are accepted.”

## Clarifying Questions

- **Promise vs. Request:**  
  Should our user stories explicitly mention that all DAG edit operations are promises
  (assertions of their validity) and not requests for action? How do we best capture this
  nuance for non-technical stakeholders?

- **Replay Mechanism:**  
  What level of detail is expected in a user story regarding the replayability of the
  system? Should acceptance criteria include replaying the entire DAG or just key portions
  necessary for verification?

- **Unified vs. Multiple DAGs:**  
  Is there a consensus on whether we use a single unified DAG for all worldlines or
  maintain separate DAGs for each? This decision affects stories related to system-wide
  verification and management.

- **Terminology Consistency:**  
  How strict should the differentiation be between “message” and “event”? Would a
  glossary be useful to ensure consistency?

## Proposed User Story Drafts

1. **Promise Assertion Story:**  
   "As an agent, I promise that the event I create is valid, uniquely signed, and
   replayable, ensuring that the corresponding DAG edit operation maintains the
   integrity of the system's worldline."

2. **Replayable History Story:**  
   "As a system operator, I want to replay the complete history of messages so that I
   can independently verify the assembly and integrity of the worldline in the Merkle
   DAG at any given point in time."

3. **Terminology Clarity Story:**  
   "As a developer, I need clear definitions for ‘message’ and ‘event’ so that every
   promise (expressed as a CWT claim) and corresponding DAG edit is consistently
   understood."

4. **Multi-Worldline Management Story:**  
   "As an administrator, I want messages to support editing operations that span
   multiple worldlines—even when they branch and merge—so that overall system
   integrity is maintained through verifiable DAG edits."

## Final Thoughts

I believe these questions and proposed user stories bridge our technical
requirements with actionable development goals. Your feedback on the clarity and
scope of these stories will be invaluable as we refine our approach further.

Best regards,  
Sally
