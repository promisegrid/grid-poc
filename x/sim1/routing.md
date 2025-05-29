# PromiseGrid Multi-Hop Routing Brainstorm

Below are alternative designs for multi-hop routing in PromiseGrid.
Each alternative builds upon ideas from ARP, routing tables with hints,
promises, CID-based routing, layered routing assemblies, pub/sub systems,
and additional novel strategies inspired by OSI/TCP-IP, Usenet, and UUCP
mechanisms.

1. ARP-Inspired Dynamic Discovery  
   In this design, each node maintains an ARP-like table mapping peer
   identifiers (CIDs or DIDs) to direct neighbor addresses. Nodes broadcast
   lightweight "who-has" messages to discover reachable hops and update their
   routing table dynamically, enabling rapid path discovery.

2. Extended Routing Table with Hints  
   Nodes maintain a routing table keyed by destination identifier (CID or DID)
   but each entry contains routing hints. The hints are a set of potential
   intermediate nodes or topics that suggest possible forwarding paths,
   allowing adaptable routing decisions based on hint sets.

3. Promise/Claim Chain with Source Routing  
   A hybrid approach where each intermediate node makes an explicit promise
   to forward messages for a given destination. A chain of these promises is
   carried within the message, combining source routing with a UUCPnet-style
   claim path that is verified cumulatively by each hop.

4. CID-Based Routing Algorithm  
   In this model the protocol CID is reinterpreted as a routing method CID.
   Different CIDs specify distinct routing behaviors, for example one for direct
   routing, another for multi-hop flooding, and yet another for promise-based
   routing, providing flexibility within a single protocol suite.

5. Layered Node and Agent Routing  
   This design separates concerns into two layers: a node routing layer and an
   agent routing layer. The node layer routes packets to the right host, while
   the agent layer functions similarly to port addressing, ensuring messages
   reach the correct process on every host.

6. Pub/Sub System with ARP-Like WhoHas/IHave Promises  
   In a pure publish/subscribe model, nodes publish interest topics and send
   out whohas/ihave message chains. An ARP-like mechanism informs nodes which
   peers are interested in which topics, so routing decisions are made on top of
   dynamic interest registration rather than static tables.

7. Gossip-Based Overlay Routing  
   Nodes propagate routing hints using a gossip protocol. Each node shares
   information about reachable destinations with a select group of peers.
   This information creates an overlay network in which each node dynamically
   selects the best multi-hop path via accumulated gossip data.

8. Flooding with Controlled Selective Repeat  
   Rather than maintaining extensive routing tables, intermediate nodes flood
   messages to all connected peers while marking duplicate messages. A
   selective repeat mechanism ensures that once a node recognizes a route it
   can prune redundant paths while ensuring eventual delivery.

9. Distributed Hash Table (DHT) Assisted Routing  
   Borrowing from BitTorrent and other P2P systems, nodes use a DHT to store
   and query routing hints. Each node can resolve a destination via a DHT lookup
   and then forward the message along the discovered path, enabling decentralized
   and scalable routing.

10. Multi-Path Bonding and Redundant Routing  
    This design uses multiple parallel routing paths. Each message is sent
    redundantly over different hops and then the destination selects the first
    correct arriving copy. Such redundancy ensures higher reliability in the
    presence of dynamic network partitions.

11. Store-and-Forward Delay Tolerant Networking  
    Nodes store messages temporarily when the destination is not immediately
    reachable and forward them when a suitable route becomes available. This is
    analogous to delay tolerant networking where intermediate nodes act as caches,
    ensuring eventual message delivery.

12. Hierarchical Routing with Proximity Awareness  
    A multi-tier grid is established where clusters of nodes form sub-grids.
    Within each sub-grid, routing is handled by local coordinators while a
    higher layer manages cross-grid routing. This hierarchy minimizes routing
    table sizes and speeds up local delivery.

13. Hybrid Source and Hop Routing  
    Messages carry explicit source routes along with local hop hints. Each hop
    can augment the route with its own best forwarding hints based on local
    criteria. This hybrid approach combines the strengths of source routing and
    dynamic hop-level decision making.

14. Real-Time Interest-Based Routing  
    Nodes propagate subscriptions and interests in real time. When a node has a
    message for a given topic, it queries the network for nodes currently
    subscribed to that topic. The resulting dynamic routing paths take into
    account current node availability and interest levels.

15. Mesh Network Routing with Adaptive TTL  
    Inspired by mesh and sensor network protocols, each message is tagged with an
    adaptive Time-To-Live (TTL) value. Every node decrements the TTL and drops the
    message if it reaches zero. Hops dynamically adjust the TTL based on network
    density and reliability, promoting balanced routing with controlled message
    life spans.

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
    from historical promise fulfillment. High-reputation nodes are more likely
    to be chosen for message forwarding, while low-reputation nodes are
    penalized in future routing assignments.

22. Market-Based: Incentivized Routing Auctions  
    Nodes bid for message forwarding opportunities in realtime auctions.
    Sample auctions allow nodes to offer competitive fees (or credits) for
    taking on message routing tasks, ensuring efficient allocation of network
    resources.

23. Market-Based: Credit-Based Routing Exchange  
    A credit system is implemented where nodes earn credits for successfully
    forwarding messages. These credits can be traded or used to prioritize
    packet delivery during network congestion, aligning supply and demand.

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
