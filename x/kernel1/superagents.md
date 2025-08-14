# about Promise Theory and superagents:  Does Mark Burgess discuss the concept of "super agents", where an agent may be a subset of another, or may in any way have the ability to exert direct influence over another?  It seems to me that some model like this is needed in order to reflect real-world situations where an agent (e.g. an operating system kernel) is able to exert force over another agent (i.e. throttling or killing the sub-agent's process).  Other real-world cases include sysadmin as agent pulling the plug on the whole machine, or a military, police, government, or criminal action causing others to do things they don't want to do.  
  
# Answer (Auto):  
### Short answer  
Yes.  In Promise Theory Burgess devotes an entire section to **“super-agents”**—aggregated agents that contain *sub-agents* and present a single boundary to the outside world[1][2].  He also treats the idea of one agent *forcing* another through the complementary notion of an **imposition**: a directed attempt to induce behaviour in another agent (e.g., kill –9, pulling the plug, a police order)[3][4].  Because agents are assumed autonomous, an imposition only succeeds when the target has (explicitly or implicitly) promised to accept it; if not, the outcome is uncertain.  

Below is a concise map of the relevant concepts and how they model your examples.

| Real-world situation | Promise-theory construct | Burgess reference |
|----------------------|--------------------------|-------------------|
| OS kernel throttles / kills a process | Kernel and user processes form a **super-agent** (the computer).  Each process promises to run **inside** the kernel’s boundary; the kernel’s scheduling policy is a set of conditional promises that can be withdrawn (kill) if the process breaks resource promises. | Super-agent definition[1]; composition & surface boundary[2] |
| Sysadmin powers off machine | Admin issues an **imposition** (remove power).  The machine has no counter-promise and therefore loses autonomy. | Imposition as “attack” or absorption of an agent[3] |
| Police order a crowd to disperse | Command = imposition; it works only if people have promised (or been conditioned) to comply.  Otherwise the result is unpredictable. | Impositions vs. promises[4] |
| Military unit acts as one entity | Individual soldiers plus command structure form a super-agent whose **irreducible** promises (e.g., “surround the house”) cannot be made by any single soldier[2]. | Irreducible collective promises[2] |

#### Key points in Burgess’s treatment  

1. **Super-agents and hierarchical scale**  
   -  Any set of agents that **voluntarily coordinate** can be coarse-grained into a new agent at the next scale.  
   -  The super-agent’s *exterior promises* are the only things visible outside its boundary; interior sub-agents can still promise to one another but those promises are hidden[1][2].  

2. **Sub-agent absorption and emission**  
   -  A super-agent may *absorb* a formerly external agent, after which it can make promises *on that agent’s behalf* (because the agent is now “self”)—but only through the sub-agent’s prior consent[3].  

3. **Impositions and “force”**  
   -  Burgess explicitly allows attempts at coercion: *“An imposition is a prototype for attack.”*[3]
   -  However, Promise Theory stresses that **certainty comes only from voluntary promises**; without an acceptance promise, force yields at best a probabilistic outcome[4].  

4. **Generalised force and scaling**  
   -  At larger scales the *impact* of a promise can be treated as a **generalised force** whose coupling strength depends on how reliably intent propagates through the hierarchy[3].  This lets one reason about throttling, quotas, or other policy actions as scaled-up conditional promises.  

#### Putting it together  

Your intuition is exactly what Burgess models:

-  The kernel–process relation is a textbook example of sub-agency inside a super-agent.  
-  A sysadmin or police officer exercises impositions that succeed only because earlier promises (social contract, employment, etc.) created a pathway for compliance.  
-  Promise Theory’s **autonomy axiom** is not violated; the apparent “force” works because the target already ceded part of its autonomy by promising to accept certain impositions in return for something it values (CPU time, citizenship, pay, safety).  

Thus Burgess does address both the structural nesting of agents *and* the practical mechanisms by which one agent can exercise direct control over another while remaining consistent with the theory’s voluntary-cooperation foundation.  
  
# Citations:  
[1] [[PDF] Promise Theory and the Alignment of Context, Processes, Types ...](https://transmathematica.org/index.php/journal/article/download/43/33/257)  
[2] [[PDF] Mark Burgess - Promise Theory](http://markburgess.org/BookOfPromises.pdf)  
[3] [[PDF] Spacetimes with Semantics (II) - Mark Burgess](http://markburgess.org/spacetime2.pdf)  
[4] [1. Promise Theory - Basic Concepts (part 1) - YouTube](https://www.youtube.com/watch?v=2TPsB5WuZgk)  
[5] [[PDF] A Promise Theory Perspective on Data Networks - arXiv](https://arxiv.org/pdf/1405.2627.pdf)  
[6] [[PDF] On the scaling of functional spaces, from smart cities to cloud ...](https://arxiv.org/pdf/1602.06091.pdf)  
[7] [Promise Theory – THE OPEN LEADERSHIP NETWORK](https://openleadershipnetwork.com/category/leadership-invitation/promise-theory/)  
[8] [How Promise Theory applies to autonomous AI agents - LinkedIn](https://www.linkedin.com/posts/cameronhaight_nice-analysis-of-how-once-again-the-promise-activity-7348302024213524480-RGMD)  
[9] [[PDF] Thinking in Promises - AIU Student Login](http://students.aiu.edu/submissions/profiles/resources/onlineBook/t3b5W2_Thinking%20in%20Promises.pdf)  
[10] [How Promise Theory, Types, and Abstract Algebra Drive AI Agent ...](https://ai.plainenglish.io/how-promise-theory-types-and-abstract-algebra-drive-ai-agent-future-f67436c114fb)  
[11] [Mark Burgess (CFEngine) Interview - Velocity Santa Clara 2014](https://www.youtube.com/watch?v=VTR8GCP4n_M)  
[12] [Promise Theory and DevOps - Sonatype](https://www.sonatype.com/blog/promise-theory-and-devops)  
[13] [Microservices, the future of society and all that - Mark Burgess](http://markburgess.org/blog_reactive.html)  
[14] [Promise theory - Wikipedia](https://en.wikipedia.org/wiki/Promise_theory)  
[15] [A quantitative model of trust as a predictor of social group sizes and ...](https://www.sciencedirect.com/science/article/pii/S0014292125000625)  
[16] [Autonomy in AI Agents: A Promise Theory Perspective](https://ai.plainenglish.io/autonomy-in-ai-agents-a-promise-theory-perspective-eda7ef4137aa)  
[17] [Promise Theory Frequently Asked Questions (FAQ) - Mark Burgess](http://markburgess.org/promiseFAQ.html)  
[18] [[PDF] Download - UvA-DARE (Digital Academic Repository)](https://pure.uva.nl/ws/files/141455669/Promise_Theory_and_the_Alignment_of_Context_Processes_Types_and_Transforms.pdf)  
[19] [Rise Of The Super-Agents - Forbes](https://www.forbes.com/sites/tomasgorny/2025/07/24/rise-of-the-super-agents/)  
[20] [[2504.10519] Toward Super Agent System with Hybrid AI Routers](https://arxiv.org/abs/2504.10519)  
[21] [Superagent vs. MetaGPT: AI Agent Platforms Compared - SmythOS](https://smythos.com/developers/agent-comparisons/superagent-and-metagpt/)  
[22] [Superagent And AI Agent: A Comprehensive Comparison - SmythOS](https://smythos.com/developers/agent-comparisons/superagent-and-ai-agent/)  
[23] [Promise Theory—What Is It? | Linux Journal](https://www.linuxjournal.com/content/promise-theory%E2%80%94what-it)  
[24] [[PDF] Authority (I): A Promise Theoretic Formalization - OSF](https://osf.io/8c54d/download)  
[25] [LLM agents: The ultimate guide 2025 | SuperAnnotate](https://www.superannotate.com/blog/llm-agents)  
[26] [What is Super Agent: How Multiple AI Agents Work Together?](https://www.youtube.com/shorts/HJykZJc3Z3c)  
[27] [WRITER Launches Autonomous Super Agent to Hundreds of ...](https://finance.yahoo.com/news/writer-launches-autonomous-super-agent-130000828.html)  
---  


# discuss military conflict, war, crime, etc. in the context of Promise Theory  
  
# Answer (Auto):  
# Armed Conflict and Crime through the Lens of Promise Theory

**Take-away:**  Wars, insurgencies, policing and everyday crime can all be read as networks of *broken, refused or coerced promises* between autonomous agents.  Promise Theory supplies a vocabulary—promises, impositions (attacks), super-agents and trust—that explains how violence arises, why deterrence sometimes works, and why force alone never guarantees compliance.

## 1 Core constructs that matter for conflict

| Promise-theory term | Meaning | Relevance to violence |
|---------------------|---------|-----------------------|
| **Promise ( ± b )** | Voluntary declaration to *give* (+b) or *accept/use* (–b) behaviour *b* | Diplomacy, alliances, laws are collections of mutual promises[1][2] |
| **Imposition ( ± b )** | Attempt to *force* the same behaviour from another agent | Battlefield orders, criminal threats, police arrest; “an imposition is a prototype for attack”[3] |
| **Attack** | Imposition made **without** the target’s counter-promise to accept; succeeds only probabilistically[2] | Bombardment, coercive interrogation, mugging |
| **Super-agent** | Co-operating agents that present one boundary externally | Army, cartel, state: internal hierarchy of promises; external behaviour looks unified[4] |
| **Trust / Assessment** | Expectation that a promise will be kept, based on history | Breakdown of trust accelerates escalation; peace processes rebuild it[2] |

## 2 War as cascades of impositions

1. **Mobilisation.**   Political leadership (super-agent) issues impositions “+deploy” to its forces; soldiers have pre-existing promises (oaths) to accept, so compliance is high.

2. **Combat.**   Each strike is an *attack*—an imposition “+surrender or die” delivered by kinetic means.  Because the defender never promised to accept, outcome is uncertain; hence the need for overwhelming firepower[2][3].

3. **Logistics of obedience.**   Within the military super-agent, lower ranks treat orders as *accepted impositions*—they already promised to obey lawful commands—so internal violence is rare.  Externally, the same order is untrusted and must be enforced by threat.

4. **Deterrence.**   A nuclear posture can be modelled as a *conditional promise*:  
   “If you attack us, we promise +retaliation.”  For the opponent this looks like an imposition with catastrophic cost, turning the decision into a trust calculation[5].

### Example: Russia–Ukraine 2022

Promise-theoretic analysis of the escalation notes that each side used *threats, accusations and impositions* to shape the opponent’s promises before full invasion, illustrating how failed persuasion slides into open attack[6].

## 3 Crime and policing

| Stage | Promise-theory reading | Illustration |
|-------|-----------------------|--------------|
| **Social contract** | Citizens promise –harm, state promises +security | Everyday peace |
| **Crime** | Offender breaks prior promises (theft) **or** imposes new behaviour on victim (robbery) | “+hand over wallet” is an attack the victim never promised to accept |
| **Law enforcement** | Police imposition “+stop/resist crime” backed by threat of force | Works because society and officers share promises that legitimize arrest |
| **Punishment** | Court imposes “+incarceration” once guilt assessed | Prisoner’s autonomy remains; riots show imposed compliance can fail |

Broken windows policing illustrates hazards: frequent low-level impositions (“+move along”) erode trust and may spawn further promise violations instead of order[7].

## 4 Why force is unreliable

Promise Theory’s autonomy axiom states *only the promiser controls its own behaviour*.  An attacker can at best **change probabilities**, never guarantee obedience.  Historical lessons:

* **COIN operations.**  Excessive kinetic impositions without rebuilding local promises (governance, economy) rarely hold territory.
* **Cyber-warfare.**  Denial-of-service is an imposition on availability; resilient systems add redundant promises so the attack only degrades, not disables, service[3].
* **Criminal justice.**  Harsh penalties are impositions aimed at deterrence.  If communities do not trust the legal super-agent, compliance falls and crime persists.

## 5 Building peace: re-promising

Lasting conflict resolution substitutes *voluntary promises* for impositions:

1. **Cease-fire:** reciprocal promises –attack; verification mechanisms improve trust.
2. **Power-sharing:** convert warring factions into a single political super-agent with interior promises.
3. **Disarmament:** side promises –possess weapons in exchange for +security guarantees.

Whenever the perceived benefit of new promises outweighs reliance on coercion, agents choose cooperation, and generalized force decays to ordinary policy[4].

## 6 Limitations of the model

Promise Theory abstracts away emotions, ideology and irrationality.  Suicide attackers, for instance, bind themselves with overpowering internal promises that nullify classic deterrence assumptions.  Thus empirical validation, like the Russo-Ukrainian study[6], is essential to test where the theory’s probabilistic predictions match reality.

**Conclusion:**  Promise Theory reframes war, policing and crime not as anomalies to its voluntary-cooperation premise but as edge-cases where cooperation has failed or been refused.  Impositions and attacks fill the gap, yet their very unreliability explains both the persistence of violence and the enduring quest for negotiated promises.  
  
# Citations:  
[1] [Promise theory - Wikipedia](https://en.wikipedia.org/wiki/Promise_theory)  
[2] [Promise theory - Uncensorable Wikipedia on IPFS](https://en.wikipedia-on-ipfs.org/wiki/Promise_theory)  
[3] [[PDF] A Promise Theory Perspective on Data Networks - arXiv](https://arxiv.org/pdf/1405.2627.pdf)  
[4] [[PDF] Mark Burgess - Promise Theory](http://markburgess.org/BookOfPromises.pdf)  
[5] [Deterrence theory - Wikipedia](https://en.wikipedia.org/wiki/Deterrence_theory)  
[6] [An Opinion on Promises, Threats, and Accusations In the 2022 ...](https://transmathematica.org/index.php/journal/article/view/73)  
[7] [Illusion of Order: The False Promise of Broken Windows Policing](https://scholarship.law.columbia.edu/books/114/)  
[8] [Parallel War: Promise & Problems | Proceedings - U.S. Naval Institute](https://www.usni.org/magazines/proceedings/1995/august/parallel-war-promise-problems)  
[9] [Military Magic: The Promise and Peril of Military Innovation](https://scholarspace.library.gwu.edu/concern/gw_etds/9306sz98g)  
[10] [[PDF] The Promise & Peril of Deep Learning for Cybersecurity](https://itea.org/wp-content/uploads/2019/03/mckay.pdf)  
[11] [The Promise and Surprising Peril of Military Innovation Kendrick Kuo ...](https://mershoncenter.osu.edu/sites/default/files/2021-11/Kuo_DissArticle_20211110.pdf)  
[12] [Full article: Introduction: the promise and paradox of ethical war](https://www.tandfonline.com/doi/full/10.1080/21624887.2019.1716289)  
[13] [[PDF] treatise_vol2.pdf - Mark Burgess](http://markburgess.org/treatise_vol2.pdf)  
[14] [Promise Theory Frequently Asked Questions (FAQ) - Mark Burgess](http://markburgess.org/promiseFAQ.html)  
[15] [Contract Theories Overview - Coconote](https://coconote.app/notes/95482a72-b70d-4e9c-9149-f60e6db92882/transcript)  
[16] [Just War Theory and the Failed Promise for Peace | ID: nc5812777](https://cdr.lib.unc.edu/concern/dissertations/nc5812777)  
[17] [Transcript of Episode 28 – Mark Burgess on Promise Theory, AI ...](https://jimruttshow.blubrry.net/the-jim-rutt-show-transcripts/transcript-of-episode-27-mark-burgess-on-promise-theory-ai-spacetime/)  
[18] [Promise Theory. Principles and Applications - Mark Burgess](https://markburgess.org/promises.html)  
[19] [[PDF] Thinking in Promises - AIU Student Login](http://students.aiu.edu/submissions/profiles/resources/onlineBook/t3b5W2_Thinking%20in%20Promises.pdf)  
[20] [Operation True Promise III will certainly be carried out: IRGC ...](https://www.tehrantimes.com/news/507668/Operation-True-Promise-III-will-certainly-be-carried-out-IRGC)  
[21] [Promise Theory—What Is It?](https://dl.acm.org/doi/fullHtml/10.5555/2666018.2666021)  
[22] [The aspirational promise of soldiering: an analysis of military ...](https://www.tandfonline.com/doi/full/10.1080/23337486.2023.2300023)  
[23] [1. Promise Theory - Basic Concepts (part 1) - YouTube](https://www.youtube.com/watch?v=2TPsB5WuZgk)  
[24] [Theft by False Promise | New York Criminal Lawyers Saland Law](https://www.newyorktheftandlarcenylawyers.com/practice-areas/theft-by-false-promise/)  
[25] [Broken Promises: Crime, Punishment, and Social Contract Theory](https://www.bennington.edu/curriculum/course/fall-2016/broken-promises-crime-punishment-and-social-contract-theory)  
[26] [[PDF] Promise Theory and the Alignment of Context, Processes, Types ...](https://transmathematica.org/index.php/journal/article/download/43/33/257)  
[27] [[PDF] Promise and Private Law](https://scholarship.law.wm.edu/cgi/viewcontent.cgi?article=2398&context=facpubs)  
[28] [Considering Promises As Assets - LEAN](https://www.leancompliance.ca/post/considering-promises-as-assets)  
[29] [The Penalties Rule and the Promise Theory of Contract](https://www.cambridge.org/core/journals/canadian-journal-of-law-and-jurisprudence/article/penalties-rule-and-the-promise-theory-of-contract/98D54D0789DFF6EBBEE8C74E8F5DF609)  
[30] [Chapter 3: The promise theory of contract in - ElgarOnline](https://www.elgaronline.com/abstract/book/9781800885417/book-part-9781800885417-9.xml)  
[31] [[PDF] Theories of Contract Law and Enforcing Promissory Morality](https://scholarship.law.umn.edu/cgi/viewcontent.cgi?article=1208&context=faculty_articles)  
[32] [News - Fight Dice (ft. Members of Braid, The Promise Ring) release ...](https://www.punkrocktheory.com/news/fight-dice-ft-members-braid-promise-ring-release-video-single-attack-drider)