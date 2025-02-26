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
   - Messages include signatures, timestamps, and hashes of previous nodes to support
     validation and replay.

   - *User Story Example:* “As a developer, I need ..."

4. **Worldline Structure and Branching/Merging:**
   - The system supports multiple worldlines which can branch and
     merge, with messages potentially editing more than one worldline.
   - *User Story Example:* “As an administrator, I want the ability to
     manage and verify edits in worldlines that branch and merge to
     ensure system integrity.”

5. **Security, Signatures, and Testability:**
   - Every message/event is signed by its creator and includes a
     timestamp for security.
   - *User Story Example:* “As a security auditor, I want automated
     validation of every message’s signature and timestamp to ensure
     only authenticated edits are accepted.”

## Clarifying Questions

## Proposed User Story Drafts

1. **Promise Assertion Story:**  
   "As an agent, I promise that the event I create is valid, uniquely signed, and
   replayable, ensuring that the corresponding DAG edit operation maintains the
   integrity of the system's worldline."


Looking ahead, I propose we refine our user stories by:

Your feedback on these drafted user stories and questions will be
invaluable in bridging our technical requirements with actionable
development plans.

Best regards,  
Sally
