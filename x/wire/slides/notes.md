* ensure IoT compatibility with grid messages
    * see 17 Apr for concepts 
    * discuss/document addresses -- what is the address of an agent?
        * might be a CID of its code
        * might be a bsky-style DID
        * might be an IPFS peer address XXX provide URL
            * note that IPFS addresses are multiaddr -- e.g. all of the above and future-proofed
            * [https://github.com/multiformats/multiaddr](https://github.com/multiformats/multiaddr) 
    * relationship between addresses and public keys
        * e.g. how to find the public key of an agent to be able to verify signatures on the agent's messages?
    * possible guinea pig is to put our decoder board on the grid -- what would that look like?
        * for now try using the scenario tree model to answer the following:
        * what does the encoder promise?
        * what do we promise the encoder?
        * ref the existing cable cutter machine: motor, encoder A and encoder B all make promises; when they don't agree, we halt
        * storage -- promises likely influence how long and where we store encoder data
        * timeliness -- promises likely influence how much we trust the encoder data to be current
* crazy idea:  could grid messages be the network or complex system equivalent of i2c for local data transfer?  i.e. could this become a decentralized IoT standard?
    * existing standards are e.g. MQTT, Particle ecosystem, offerings from AWS and Google, Arduino, Adafruit
        * these all tend to need some sort of translation between IoT and larger, full-OS device message formats, e.g. MQTT⇔ HTTP
        * these are generally all centralized, single points of failure
            * often rely on e.g. CSV file transfers between wild apricot and a raspberry pi for door or machine access
            * what we really want is for the infrastructure, the grid itself, to take care of data replication, rather than the app needing to handle that
                * e.g. door access dataset lives on both membership server and raspberry pi, not on one or the other
    * if grid messages can do this, then IoT devices become first-class members of a network


* ensure IoT compatibility with grid messages
    * concepts:
        * kernel types (e.g. monolithic and microkernel)
            * e.g. Mach is a message-passing microkernel
            * so is grid
        * message formats
            * future-proffing is important
            * kernels and other recipients need to know how to parse/route messages based on header information
            * grid message uses CBOR format containing:
                * array length (always 3)
                * 'grid'  to identify who we are
                * protocol identifier (e.g. are we using a scenario-tree model, or something else that we can't yet imagine, again it's about future-proofing)
                    * protocol identifier is a multihash CID of the document describing the protocol
                        * no need for a central registry of protocols
                    * [https://github.com/multiformats/multihash](https://github.com/multiformats/multihash) 
                    * [https://proto.school/anatomy-of-a-cid](https://proto.school/anatomy-of-a-cid) 
                * payload -- layout, encoding, and meaning are all specified in the protocol document that is referred to by the CID.
                    * the document is probably (and should be) accessible by CID via an IPFS get() request
    * ensure a small IoT device can be an agent
        * if we do the above, minimum message size is going to be something like 50 bytes XXX finish wire_test.go to get the exact number
            * extremely constrained devices will need to be fronted by others that can strip off header info and just send payload
        * properties of an agent:
            * has the ability to keep or break(revoke?) promises
            * has the ability to operate in a trust framework -- others will have an interest in building trust metrics based on promise keep/break record
            * does an agent have self-interest?
                * if an agent has at least one personal currency, then it would seem that it would have an interest in maintaining the currency's value at a particular target (not necessarily maximizing, the agent may have other goals for its own currency market)
            * problem:  Some agents might be too simple to manage their own self interest when interacting with others, might need proxies
                * problem: Burgess' definition of agent includes the concept that an agent can't make promises on behalf of another; a proxy would need to be able to do so XXX Steve ask Mark/revisit his papers
                * practical example:
                    * Is an incandescent light bulb an agent?
                    * Or is a microcontroller monitoring the filament status the agent?
            * has some form of address on the grid
                * can send and receive messages at that address
            * can sign messages
                * this would require proxies for simpler devices
                * in many cases, a kernel may be the only reasonable signer due to local security issues, data access to private keys, etc. 
                    * e.g. a WASM module may not have secure local storage, kernel may need to sign for its messages
                    * but that might be okay -- the kernel is in the best position to hash the module to verify it's the agent we think we're talking to anyway





* discussion of markets, models, scenario trees, simulations
    * PromiseGrid's low-level data model is likely a hypergraph of worldlines that represent a scenario tree
        * community consensus forms the probabilities, risk/reward numbers, etc. of a given worldline, possibly as a decision market:  
            * upper probability limit is ask, lower is bid 
            * currency is likely personal trust tokens
            * actual outcomes versus agents' bid/ask *can be used* in trust calculations by ***each*** agent (note that exchange rate is the relationship between two personal currencies and can be used in weighting trust calculations by an agent)
                * see e.g. Iowa Electronic Markets "unit portfolio" for math that could work:
                    * [https://www.perplexity.ai/search/where-is-the-documentaiton-of-iSVsDqRMSpiz4XqlSMKwxg?0=r](https://www.perplexity.ai/search/where-is-the-documentaiton-of-iSVsDqRMSpiz4XqlSMKwxg?0=r) 
    * The most primitive worldline type in the tree is probably an individual agent's timeline of actions the agent has taken, all expressed as a hash chain



* IPFS/IPLD ecosystem
    * like minds, healthy community
        * may be an even better fit than e.g. MIT decentralized AI roundtable folks
        * or DevOps
    * hash-based decentralized content-addressable storage
        * just like PromiseGrid
    * hashes are expressed as Content IDs (CIDs)
        * was already thinking to use CIDs for PromiseGrid
        * CIDs are also used by bluesky
            * [https://atproto.com/specs/data-model](https://atproto.com/specs/data-model) 
    * IPFS data structure is a decentralized directed acyclic graph (DAG)
        * was considering a DAG for PromiseGrid's basic structure
    * DAG links are expressed as Interplanetary Linked Data (IPLD) items
        * was already thinking to use IPLD for PromiseGrid DAG links
        * IPLD is also used by bluesky
            * [https://atproto.com/specs/data-model](https://atproto.com/specs/data-model) 
    * rich ecosystem
        * including IPFS-based WASM decentralized computation platforms, including AI bits:
            * [https://github.com/bacalhau-project/bacalhau](https://github.com/bacalhau-project/bacalhau)
            * [https://github.com/IceFireLabs/WASM-IPFS-SentientNet](https://github.com/IceFireLabs/WASM-IPFS-SentientNet) 
        * lots of overlap between IPFS and bluesky dev communities
            * hundreds of github code search hits for Go code that mentions both bsky and ipfs:
                * [https://github.com/search?q=language%3AGo+ipfs+bsky&type=Code](https://github.com/search?q=language%3AGo+ipfs+bsky&type=Code) 
        * hundreds of other related projects using IPFS libraries as of 2023:
            * [https://github.com/ipfs/boxo/wiki/Dependents-(consumers)-of-Boxo](https://github.com/ipfs/boxo/wiki/Dependents-(consumers)-of-Boxo) 
    * most IPFS ecosystem code is in Go
    * bluesky borrowed from IPFS:
        * CBOR
        * DAG-CBOR
        * [multicodecs](https://github.com/multiformats/multicodec)
        * [multihashes](https://multiformats.io/multihash/)
        * CIDs
        * …all of which I was already trending towards for PromiseGrid
        * Other developers likely to be learning and using these things already, because of the popularity of bluesky
* So what's missing?  Why don't we just use IPFS or one of the existing projects like bacalhau or SentientNet instead of developing PromiseGrid?
    * IPFS itself has issues -- lots of code, some very old, lots of API instability, still evolving, components still at v0.X version numbers
    * what's missing is accountability -- promises
        * need some sort of personal currencies, tokens, tickets, vouchers, or other trust metrics
        * most existing projects (e.g. [Filecoin](https://filecoin.io/)) are built on single-currency blockchains instead
    * still looking at existing projects to see what we can import from those projects
        * looks like there's a lot we can borrow
    * messaging (as opposed to data storage) looks thin -- still looking
        * between all agents, not just human
        * we may need to write this
        * we may be able to build on bluesky's infrastructure or their [AT protocol](https://atproto.com/)

 

