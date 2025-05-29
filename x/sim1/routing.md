# PromiseGrid Multi-Hop Routing Brainstorm

Below are alternative designs for multi-hop routing in
PromiseGrid. Each alternative builds upon ideas from ARP, routing
tables with hints, promises, CID-based routing, layered routing
assemblies, pub/sub systems, and additional novel strategies inspired
by OSI/TCP-IP, Usenet, UUCP, double-auction markets, and prediction
markets. The goal is to create a flexible, decentralized, and
efficient routing mechanism that can adapt to dynamic network
conditions while maintaining simplicity in the kernel.

1. ARP-Inspired Dynamic Discovery  
   In this design, each node maintains an ARP-like table mapping peer
   identifiers (CIDs or DIDs) to direct neighbor addresses. Nodes
   broadcast lightweight "who-has" messages to discover reachable hops
   and update their routing table dynamically, enabling rapid path
   discovery.

2. Extended Routing Table with Hints  
   Nodes maintain a routing table keyed by destination identifier (CID or
   DID) but each entry contains routing hints. The hints are a set of
   potential intermediate nodes or topics that suggest possible
   forwarding paths, allowing adaptable routing decisions based on hint
   sets.

3. Promise/Claim Chain with Source Routing  
   A hybrid approach where each intermediate node makes an explicit
   promise to forward messages for a given destination. A chain of these
   promises is carried within the message, combining source routing with a
   UUCPnet-style claim path that is verified cumulatively by each hop.

4. CID-Based Routing Algorithm  
   In this model the protocol CID is reinterpreted as a routing method
   CID. Different CIDs specify distinct routing behaviors, for example one
   for direct routing, another for multi-hop flooding, and yet another for
   promise-based routing, providing flexibility within a single protocol
   suite.

5. Layered Node and Agent Routing  
   This design separates concerns into two layers: a node routing layer and
   an agent routing layer. The node layer routes packets to the right host,
   while the agent layer functions similarly to port addressing, ensuring
   messages reach the correct process on every host.

6. Pub/Sub System with ARP-Like WhoHas/IHave Promises  
   In a pure publish/subscribe model, nodes publish interest topics and
   send out whohas/ihave message chains. An ARP-like mechanism informs
   nodes which peers are interested in which topics, so routing decisions
   are made on top of dynamic interest registration rather than static
   tables.

7. Gossip-Based Overlay Routing  
   Nodes propagate routing hints using a gossip protocol. Each node shares
   information about reachable destinations with a select group of peers.
   This information creates an overlay network in which each node
   dynamically selects the best multi-hop path via accumulated gossip data.

8. Flooding with Controlled Selective Repeat  
   Rather than maintaining extensive routing tables, intermediate nodes
   flood messages to all connected peers while marking duplicate messages.
   A selective repeat mechanism ensures that once a node recognizes a route
   it can prune redundant paths while ensuring eventual delivery.

9. Distributed Hash Table (DHT) Assisted Routing  
   Borrowing from BitTorrent and other P2P systems, nodes use a DHT to store
   and query routing hints. Each node can resolve a destination via a DHT
   lookup and then forward the message along the discovered path,
   enabling decentralized and scalable routing.

10. Multi-Path Bonding and Redundant Routing  
    This design uses multiple parallel routing paths. Each message is sent
    redundantly over different hops and then the destination selects the
    first correct arriving copy. Such redundancy ensures higher reliability
    in the presence of dynamic network partitions.

11. Store-and-Forward Delay Tolerant Networking  
    Nodes store messages temporarily when the destination is not immediately
    reachable and forward them when a suitable route becomes available.
    This is analogous to delay tolerant networking where intermediate nodes act
    as caches, ensuring eventual message delivery.

12. Hierarchical Routing with Proximity Awareness  
    A multi-tier grid is established where clusters of nodes form sub-grids.
    Within each sub-grid, routing is handled by local coordinators while a
    higher layer manages cross-grid routing. This hierarchy minimizes
    routing table sizes and speeds up local delivery.

13. Hybrid Source and Hop Routing  
    Messages carry explicit source routes along with local hop hints. Each
    hop can augment the route with its own best forwarding hints based on
    local criteria. This hybrid approach combines the strengths of source
    routing and dynamic hop-level decision making.

14. Real-Time Interest-Based Routing  
    Nodes propagate subscriptions and interests in real time. When a node
    has a message for a given topic, it queries the network for nodes
    currently subscribed to that topic. The resulting dynamic routing paths
    take into account current node availability and interest levels.

15. Mesh Network Routing with Adaptive TTL  
    Inspired by mesh and sensor network protocols, each message is tagged
    with an adaptive Time-To-Live (TTL) value. Every node decrements the TTL
    and drops the message if it reaches zero. Hops dynamically adjust the
    TTL based on network density and reliability, promoting balanced
    routing with controlled message life spans.

16. Probabilistic Gossip Routing  
    In this variant, nodes forward messages with a probability that adapts
    based on network density and recent message traffic. This probabilistic
    approach limits redundant transmissions, while still ensuring wide
    dissemination via random neighbor selection.

17. Epidemic Propagation with Expiring Tokens  
    Here, messages are spread in an epidemic fashion across the network.
    Each message carries an expiring token that prevents repeated
    retransmissions. Nodes exchange these tokens to track message history,
    reducing unnecessary floods and ensuring timely message expiry.

18. Hybrid Gossip-Epidemic Routing  
    This design combines randomized gossip and controlled epidemic spread.
    Nodes initially employ gossip to selectively share routing hints and
    monitor acknowledgements. If delivery rates fall below a threshold, the
    protocol switches to an epidemic mode with controlled backoff to achieve
    robust broadcast in sparse network conditions.

19. Promise-Theory: Conditional Promise Routing  
    Nodes make explicit promises with conditions attached to each hop.
    If a promise is unfulfilled, nodes dynamically reroute the message to a
    different path. This protocol integrates trust scoring based on past
    promise fulfillment.

20. Promise-Theory: Cooperative Exchange Routing  
    Each node negotiates promises for message forwarding. Nodes exchange
    commitments and mutually verify promise exchanges, leading to a
    self-enforcing mesh of reliable routes.

21. Promise-Theory: Reputation-Grounded Routing  
    Nodes calculate routing decisions based on a reputation score derived
    from historical promise fulfillment. High-reputation nodes are more
    likely to be chosen for message forwarding, while low-reputation nodes
    are penalized in future routing assignments.

22. Market-Based: Incentivized Routing Auctions  
    Nodes bid for message forwarding opportunities in realtime auctions.
    Sample auctions allow nodes to offer competitive fees (or credits) for
    taking on message routing tasks, ensuring efficient allocation of
    network resources.

23. Market-Based: Credit-Based Routing Exchange  
    A credit system is implemented where nodes earn credits for
    successfully forwarding messages. These credits can be traded or used
    to prioritize packet delivery during network congestion, aligning supply
    and demand.

24. Market-Based: Dynamic Routing Pricing  
    Node forwarders set dynamic prices based on network load and available
    capacity. Senders select routes based not only on speed but also on cost,
    creating a market-driven optimization for message transit.

25. Prediction-Market: Forecasted Route Efficiency  
    Nodes use historical traffic and performance data to forecast the
    efficiency of potential routes. These predictions are shared in a
    prediction market where nodes vote on the most reliable paths.

26. Prediction-Market: Expectation-Driven Routing  
    Each node submits predicted delivery times and success rates for
    various routes. Routes with the highest consensus of efficient delivery
    expectations are automatically chosen for message forwarding.

27. Prediction-Market: Consensus-Based Route Selection  
    A decentralized market mechanism gathers predictions and betting pools
    on route success. The route with the highest consensus via market-based
    scoring is selected, rewarding nodes for accurate predictions and
    efficient forwarding.

28. Personal Currency: Routing Collateral Tokens  
    Nodes issue personal currency tokens as collateral for message
    forwarding. Each token represents a promise to provide routing or other
    services such as storage, CPU, or bandwidth. Upon successful delivery,
    the tokens can be redeemed with the issuer, creating an incentive
    structure that rewards reliable behavior.

29. Personal-Currency-Enabled Store-and-Forward  
    Agents use their self-issued IOUs or tokens to pay for temporary
    storage services. In this protocol, nodes holding messages collect personal
    currency as a fee, which can later be used to request services from
    peers, creating a decentralized market for delay tolerant networking.

30. Prediction Market with Personal Currency Hedging  
    In this protocol, agents back their route efficiency predictions with
    personal currency tokens. These tokens act as a hedge against route
    failures. Nodes that accurately forecast performance earn tokens that
    can be exchanged for premium routing services in future transactions.

31. Pub/Sub with Dynamic Personal Currency Markets  
    Subscribers assign personal currency values to topics of interest.
    Messages are routed along paths where nodes with higher currency stakes
    gain priority in delivery. This dynamic market mechanism aligns the
    distribution incentives with the subscribers' valuations ensuring that
    high-value messages are prioritized.

32. Hybrid Promise-Personal Currency Routing  
    Combining promise-based routing with personal currency tokens, nodes
    first exchange tokens as a commitment to forward messages. Failure to
    complete the promise results in penalties measured in lost tokens.
    This hybrid protocol reinforces accountability and encourages robust
    routing through mutually beneficial exchanges.

33. Common Resource Protection via Quota-Based Routing  
    In this protocol, nodes are allocated a fixed quota of forwarding credits
    in each time interval. Exceeding the quota results in deprioritized
    routing, ensuring that no single node exhausts the shared forwarding
    capacity and contributes to a tragedy of the commons.

34. Collaborative Resource Sharing through Social Contracts  
    Nodes agree to a social contract that limits excessive use of shared
    network resources. With reputation and penalty systems in place, nodes
    monitor each other to ensure fair usage, thus mitigating network
    congestion and resource depletion.

35. Incentivized Contribution with Fair Credit Allocation  
    This design rewards nodes that contribute fairly to multi-hop routing.
    Nodes garner extra credits for balanced participation while those over-
    consuming resources receive lower credits. This market-based approach
    aligns individual incentives with overall network health.

36. Adaptive Fairness in Multi-Hop Routing  
    Nodes dynamically adjust their routing costs based on an adaptive
    fairness metric. Routing decisions take into account both past
    contributions and current load, promoting equitable resource usage and
    preventing any node from monopolizing the network.

37. Cooperative Hub Rotation Routing  
    In this protocol, nodes are required to take on routing roles on a
    rotational basis. The network periodically elects new hubs based on
    local consensus, discouraging any node from permanently accumulating
    power or centrality. Hubs receive temporary token bonuses for their
    service before rotating out in favor of other peers.

38. Decentralized Trust-Based Routing with Dynamic
    Accountability  
    Nodes earn trust scores through decentralized peer reviews of their
    performance. While higher scores may confer minor routing advantages,
    the protocol favors balanced participation. Overcentralized nodes
    face dynamic penalties that encourage distribution of routing tasks.

39. Reputation-Weighted Distributed Routing  
    This approach uses a reputation system that rewards nodes for balanced
    participation. The algorithm factors in equal load distribution and
    consistent performance, elevating nodes that avoid becoming a central
    hub. Reputation scores influence routing decisions to deter emergent
    centralization.

40. Peer-Collaboration Incentive Routing  
    Nodes periodically share performance metrics and resource usage
    statistics. Routing decisions are based on a collaborative evaluation,
    awarding credits to nodes that assist less-central peers and penalizing
    those that become overloaded. This shifting reward system helps keep
    the network topology widely distributed.

41. Dynamic Forking for Load Decentralization  
    If a node is overwhelmed with routing requests, it may offload part
    of the workload to nearby neighbors. The protocol redistributes credits
    evenly among participating nodes, incentivizing load delegation and
    reducing the likelihood that any single node will evolve into a choke
    point.

42. Capability Token Enhanced Source Routing  
    In this protocol, nodes exchange routing messages secured by explicit
    capability tokens. A sender attaches a token to each message as a grant
    for intermediate hops to verify their authorization before forwarding.
    This mechanism ensures that only nodes with valid tokens engage in the
    multi-hop routing process.

43. Promise-Capability Negotiation Routing  
    Combining promise theory with capability tokens, nodes negotiate routing
    commitments by exchanging promises statements that act as routing
    capability tokens.

44. PromiseGrid Best-of-Breed Routing  
    This design integrates promise-like capability tokens, personal
    currencies, dynamic exchange rates for reputation, and pub/sub
    semantics. In this protocol, every message carries a promise token
    that authorizes hops for message forwarding while simultaneously
    including a personal currency value. The protocol uses dynamic
    exchange rates between personal currency units as reputation
    scores, enabling nodes to evaluate routing priority on real-time
    pub/sub topics. This approach drives both accountability and
    efficiency.
    
    Nodes negotiate forwarding commitments using capability tokens.
    A sender offers a token bundled with a tentative currency amount.
    As the message passes through routers, each hop validates the token,
    checks against its own reputation exchange rate, and agrees to forward
    the message provided the promise holds. On delivery, acknowledgements
    are routed back along the reverse path, updating the reputation metrics
    and settling currency balances in a decentralized market fashion.

    The following sequence diagram illustrates the interaction between a
    sender, intermediate routers, and a destination in this protocol.

    ```mermaid
    sequenceDiagram
      participant S as Sender
      participant R1 as Router1
      participant R2 as Router2
      participant D as Destination
      S->>R1: Send Message with Capability Token and Currency
      R1->>R2: Forward Promise, Token, Currency Bid Attached
      R2->>D: Deliver Message for Pub/Sub Event
      D-->>R2: Acknowledge Delivery with Reputation Update
      R2-->>R1: Confirm Forwarding, Adjust Currency Balance
      R1-->>S: Completion Acknowledgement and Reputation Gain
    ```

45. Ant Colony Inspired Routing  
    In this protocol, nodes simulate virtual pheromone trails to mark
    frequently successful paths. Over time, the accumulation of virtual
    pheromones helps nodes converge on optimal routes, as repeated
    traffic reinforces the trails that yield reliable delivery.

46. Reactive Ant Routing with Pheromone Decay  
    Nodes initially broadcast messages in a manner similar to ant
    foraging. On successful message delivery, a virtual pheromone is
    deposited along the route. This pheromone decays over time, so only
    recently successful paths maintain strong attractiveness.

47. Hybrid Ant-Promise Cooperative Routing  
    Combining ant-inspired signaling with promise-based commitments,
    nodes evaluate paths using both accumulated pheromone levels and
    explicit routing promises. This hybrid approach reinforces reliable
    message forwarding while dynamically adapting to changing network
    conditions.

48. Floating-Exchange Double Auction Routing  
    Agents use personal currencies with floating exchange rates to pay
    for promises and routing services. An open double-auction market
    dynamically sets the exchange rates between currency pairs. Nodes
    bid for reliable routes, and payments are automatically settled upon
    promise fulfillment.

49. Real-Time Market Negotiation Routing  
    In this protocol, nodes engage in real-time double auctions to set
    rates for routing services. Each agent maintains a personal currency,
    and current bids and asks determine the exchange rates. Dynamic
    pricing encourages fair competition and rewards nodes that offer
    consistently reliable message forwarding.

50. Floating Credit Exchange Routing  
    Nodes trade credits representing their personal currencies on an open
    double-auction platform. Messages require a deposit of credits that
    adjusts in real time. Local supply and demand fluctuations yield an
    adaptive pricing model, ensuring that only the best-informed nodes
    act as reliable routers.

51. Floating-Exchange Personal Currency Protocol  
    Agents operate with personal currencies and pay for message forwarding
    and promise fulfillment via an open double-auction market. Each agent
    uses a personal currency whose value against other currencies
    fluctuates in real-time based on supply and demand.
    
    Core features:
    - Matching Engine: A system component that pairs buy and sell orders
      in real time, ensuring efficient trade execution.
    - Order Book: A public ledger of current bids and asks for routing
      services and promise fulfillment.
    - Double-Entry Accounting Journal: A ledger that records every credit
      and debit transaction, ensuring accountability and transparency.
    - Digital Signatures & Smart Contracts: Mechanisms to verify
      transactions and conditional forwarding promises.
    - Reputation & Trust Metrics: A system that tracks historical performance,
      rewarding reliable nodes and penalizing nodes with unmet promises.
    - Discovery Protocol: Techniques (gossip, DHT, or distributed
      registries) for agents to locate potential routing partners.
    
    Protocol Workflow:
    1. Discovery: Agents publish their service advertisements along with
       their currency denomination and current exchange rate information.
    2. Promise Exchange: Nodes request promise services from peers,
       attaching a bid in their personal currency. Interested nodes respond
       by placing an ask on the order book.
    3. Matching and Auction: The matching engine continuously matches bids
       and asks via a real-time double auction, dynamically adjusting
       personal currency exchange rates.
    4. Execution of Promises: Once matched, nodes execute the agreed promise,
       with smart contracts ensuring that the token or promise is only
       redeemed upon successful message forwarding.
    5. Settlement: The double-entry accounting journal records the credit
       and debit transactions, adjusting each node's balance and updating
       their reputation scores.

52. Hypergraph of Worldlines Routing  
    In this protocol, the goal is to build a hypergraph where each
    message is an addition of one or more nodes or edges representing
    worldlines. Worldlines may include data or code, capture transactions
    in personal currencies, or append actions to an agent's work queue.
    Whether the recipient accepts an appended worldline is at the discretion
    of that agent. Every message is signed by its author to ensure
    authenticity and non-repudiation.

53. Scenario Hypertree of Worldlines Routing  
    This protocol focuses on constructing a scenario hypertree composed
    of past and future worldlines. Each message either adds nodes or edges
    to the tree, or functions as a bid or ask in a prediction-market
    mechanism that adjusts the probability of an edge, or the cost or
    benefit of a node. Worldlines here can encapsulate data or code,
    model transactions with personal currencies, or request actions to be
    appended to another agent's work queue. As with other protocols, the
    message is cryptographically signed by its author to maintain integrity.

54. Hypergraph-Based Multi-Hop Routing with CID
    Dependencies  
    In this protocol, each message carries the CID(s) of one or more
    previous nodes or edges that serve as dependencies, along with the
    CID(s) of one or more new nodes or edges being introduced into the
    hypergraph. The hypergraph is structured as a directed acyclic graph
    representing a set of worldlines.
    
    The routing challenge is to forward messages to the correct agent(s)
    based on the CID references from earlier entries. Upon receiving a
    message, an agent compares the referenced CIDs against its local state.
    If a match is found—indicating that the agent is connected to a prior
    node or edge—the message is processed and forwarded accordingly.
    
    Message propagation is facilitated by an implicit subscription model,
    where agents express interest in specific CIDs. Secure message signing,
    chaining, and verification ensure that only authenticated modifications
    are accepted. This design enables an efficient construction of a network-
    wide hypergraph of worldlines while preserving the acyclic properties
    of the graph.
