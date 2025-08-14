# What's the correct term? Is it actually API or is it ABI?  
  
# Answer (turbo):  
The correct term is ABI, which stands for Application Binary Interface. The ABI defines the low-level interface between a program and the operating system or kernel, like system calls or how data structures are passed. An API, or Application Programming Interface, operates at a higher level and is more about function calls and libraries in source code. So, for interactions between a process and the kernel, ABI is the more accurate term.  
  
# Citations:  
  
---  


# What about between a WebAssembly module and the WebAssembly VM?  
  
# Answer (turbo):  
For WebAssembly, the term ABI is also commonly used to describe how a compiled WebAssembly module interacts with the WebAssembly VM. The WebAssembly ABI defines how functions are called, how data is passed, and how memory is managed between the module and the host environment. So just like with native binaries, ABI is the appropriate term for that low-level interaction.  
  
# Citations:  
  
---  


# Are WebAssembly host functions part of the ABI, or are they something else?  
  
# Answer (turbo):  
WebAssembly host functions are indeed part of the ABI. They define how the WebAssembly module can invoke functions provided by the host environment. The ABI covers how those functions are called, how arguments are passed, and how results are returned. So even though they might feel like higher-level functions, the way they interact with the module is still governed by the ABI.  
  
# Citations:  
  
---  


# type of ABI, standardized ABI for WebAssembly.  
  
# Answer (turbo):  
Exactly! WASI is a standardized ABI for WebAssembly modules that provides a set of system-like interfaces, such as file I/O, clocks, and random number generation. It's designed to be portable and safe, so WebAssembly code can run outside of web browsers and interact with the host system through these well-defined interfaces. So yes, WASI is an example of a standardized ABI for WebAssembly.  
  
# Citations:  
  
---  


# Thanks.  
  
# Answer (turbo):  
You're welcome! So from the projects that exist, WASI is currently one of the most prominent efforts to unify an ABI across platforms, from WebAssembly environments to general-purpose OSes. But there are other projects trying to create similar portability layers.

For example, POSIX is a classic one that spans general-purpose OSes, though it doesn't cover WebAssembly or IoT directly. Another interesting effort is the IoT-focused Zephyr Project, which aims to provide a common platform for embedded devices, but it’s more hardware focused.

Right now, there's no fully universal ABI spanning all those environments like you're envisioning, but WASI is the most active project extending across WebAssembly and other platforms. Your idea of having a unified ABI across such a wide range of environments is definitely ambitious and cutting-edge!  
  
# Citations:  
  
---  


# do you think it's actually possible?  
  
# Answer (turbo):  
It’s definitely possible, but it’s a big challenge. The diversity of platforms—ranging from IoT devices with limited resources, to general-purpose OSes, to WebAssembly VMs—means you’d need very careful abstraction and compromise.

But WebAssembly and WASI already show that it’s possible to create a portable, secure interface that works across many environments. Extending that approach to IoT devices and more powerful systems is technically complex, but not impossible.

It would require careful design to balance performance, portability, and security, but I think it’s an exciting frontier that’s definitely achievable with collaboration and standardization efforts.  
  
# Citations:  
  
---  


# or something that's more like function as a service, sort of like a decentralized Lambda service.  
  
# Answer (turbo):  
Given your goal of decentralization and portability across such different environments, a more message-oriented, microkernel-like model might be more flexible. Message passing tends to be easier to adapt across systems with different capabilities, especially when you’re dealing with limited or heterogeneous devices.

Message-passing systems also tend to scale better in decentralized and distributed structures, since they naturally encourage loose coupling between components. That fits well with your vision of PromiseGrid.

However, for certain low-level operations, like file I/O or direct hardware interactions, a syscall-like interface might still be necessary for efficiency.

You could end up with a hybrid approach: message-passing for general operations and coordination, and function-oriented calls for performance-critical tasks. It’s all about balancing flexibility with the realities of different hardware and software environments.  
  
# Citations:  
  
---  


# that last thing I mentioned about decentralized lambda as a service, functions as a service, decentralized.  
  
# Answer (turbo):  
That’s actually a really interesting model for what you’re doing. A decentralized "functions-as-a-service" could work really well for PromiseGrid. In that model, each node—whether it’s an IoT device, a VM, or a server—could execute small, portable functions on demand. Nodes would communicate through message passing to handle things like task delegation and results retrieval.

This model would inherently encourage loose coupling, scalability, and fault tolerance, since each node would only need to know about the messages it’s handling at that moment.

It also aligns with the idea of extreme portability, since each function could be deployed independently of the underlying hardware or OS. You’d just need to define the "contract"—basically the ABI—between the function and the host environment.

So yes, a decentralized Lambda-like service model could be a great fit for what you’re aiming to achieve. It’s very forward-thinking and aligns well with modern distributed computing trends.  
  
# Citations:  
  
---  


# Um...  
  
# Answer (turbo):  
the line between message passing and function calls can blur quite a bit, especially in distributed systems. Underneath, a message often triggers some function or callback on the receiving end. So function calls are the basic building blocks, but message passing adds that layer of abstraction and flexibility you need for decentralization.

If you have a system that’s event-driven or uses publish-subscribe patterns, then message passing fits perfectly. You can still optimize the actual low-level primitives—like function calls for send and receive—while keeping that higher-level message-passing abstraction for developers.

That way you get the best of both worlds: a flexible, portable architecture on the surface, and efficient, low-level operations underneath. It’s a good way to handle both the complexity and performance issues across different environments.  
  
# Citations:  
  
---  


# and you'll see more of what I'm thinking about there.  
  
# Answer (turbo):  
I took a look at your README, and it’s really clear how PromiseGrid’s architecture is designed around decentralized execution and capability-based security.

You’re absolutely right: with function calls, it’s natural to pass capability tokens along, which align really well with how PromiseGrid handles code and resource access. The idea of tokens representing permissions or resources (like personal currencies or compute credits) definitely fits into that model.

Because WebAssembly doesn’t have built-in messaging or streaming facilities, using function calls to host functions is the way to go. But layering on the capability tokens and message-passing abstraction on top of that still gives you the flexibility and security model you’re aiming for.

So, combining function calls as the lower-level primitive and capability-based tokens for security makes a lot of sense for PromiseGrid’s decentralized approach.  
  
# Citations:  
[1] [PromiseGrid - GitHub](https://github.com/promisegrid/promisegrid)  
[2] [github.com/stevegt/grokker v2.1.9-0.20250425020153 ... - Libraries.io](https://libraries.io/go/github.com%2Fstevegt%2Fgrokker)  
[3] [Test toolbox for go-libp2p modules - GitHub](https://github.com/libp2p/go-libp2p-testing)  
[4] [Activity · promisegrid/promisegrid · GitHub](https://github.com/promisegrid/promisegrid/activity)  
[5] [promisegrid/README.md at main · promisegrid/promisegrid](https://github.com/promisegrid/promisegrid/blob/main/README.md)  
[6] [Syncing Docs via CLI / GitHub - ReadMe Docs](https://docs.readme.com/main/docs/rdme)  
[7] [README.md - jparkie/Promise - GitHub](https://github.com/jparkie/Promise/blob/master/README.md)  
[8] [ipfs/go-ds-pebble: A datastore implementation backed by ... - GitHub](https://github.com/ipfs/go-ds-pebble)  
[9] [hsanjuan/ipfs-lite - GitHub](https://github.com/hsanjuan/ipfs-lite)  
[10] [README.md - nodejs/promises - GitHub](https://github.com/nodejs/promises/blob/master/README.md)  
---  


# Now go back and revisit the IoT environment. Does everything that you just now said still hold true for IoT devices?  
  
# Answer (turbo):  
For IoT devices, most of what we discussed still holds true, but with a few important considerations.

IoT devices usually have much more limited processing power, memory, and power availability. That means efficiency is even more critical. The message-passing and capability-based model still works well, but you’d have to be very mindful of how much overhead each function call or message adds.

The capability tokens also remain relevant in IoT environments, especially for secure access control. In fact, they might be even more crucial for IoT, where security is a big concern due to their constrained nature and potential exposure.

So the core concepts—function calls as a base layer, message-passing for flexibility, and capability tokens for security—still make sense for IoT. You just have to optimize them carefully for resource constraints.  
  
# Citations:  
  
---  


# I don't want to use an RPC, a strict RPC model. I'd like something that's more flexible and asynchronous than that.  
  
# Answer (turbo):  
That makes a lot of sense. Using content-addressable storage like IPFS-style CIDs for generic functions and code is a great approach for decentralization and redundancy. It ensures that any node can fetch and run the code it needs, no matter where it lives in the network.

For the protocol between kernels on different hardware pieces, if you want to avoid strict RPC models and prefer something flexible and asynchronous, then a message-passing or event-driven protocol would be a natural fit. That way, each kernel can post tasks as messages and receive results asynchronously, allowing for more fluid interaction between peers.

It sounds like you're envisioning a model where functions can be invoked across nodes without tight coupling—something like pushing a function call into a queue and letting any node with the right capabilities pick it up. That’s definitely more flexible and scales well in a decentralized system.

So yes, combining content-addressable code with a flexible, asynchronous messaging protocol between kernels sounds like the right path forward for PromiseGrid’s architecture.  
  
# Citations:  
  
---  


# Are you sure that that actually makes sense in this situation?  
  
# Answer (turbo):  
I get what you mean—it does sound a bit like flipping the abstraction levels around. Usually, we think of message-passing as a higher-level abstraction over lower-level function calls.

But in a decentralized system like the one you’re envisioning, it can make sense for messaging to effectively become the foundation of communication, especially across nodes. That’s because message-passing is inherently more flexible and asynchronous, which suits a distributed architecture better than tightly coupled RPC-style function calls.

So in that sense, the “inversion” might actually be an advantage. You’re simplifying the system’s overall complexity by having a consistent communication model at all levels. Function calls would still exist inside nodes for performance-critical tasks, but communication between nodes leans on messaging because it’s more adaptable to a distributed, asynchronous world.

So yes, it does make sense, even if it feels a bit unusual at first glance.  
  
# Citations:  
  
---  


# turns those system calls into low-level message passing to different ports? Is that what's actually going on?  
  
# Answer (turbo):  
Yes, that's exactly what's happening in a microkernel like Mach. The system calls that processes make appear to be traditional function calls, but under the hood, those syscalls are actually translated into low-level message-passing between the user processes and the kernel.

So even though applications see a conventional "syscall" layer, what’s really going on underneath is message-passing between different components of the system, such as drivers, filesystems, and other services. That microkernel architecture is more modular and flexible, but it relies heavily on this message-passing mechanism.
So yeah, you're spot on—Mach is doing something pretty similar to what we’re talking about, just at a different scale and context.  
  
# Citations:  
  
---  


# its purpose.  
  
# Answer (turbo):  
All right, let's break it down with a minimal and efficient set of core function calls for the ABI between a process and the kernel, designed around message passing and capability-based access.

Let’s say we keep it lean with around six core function calls:

1. Send Message: This sends a message to another process or kernel service, either locally or across the network.

2. Receive Message: This lets a process receive incoming messages, blocking or non-blocking depending on configuration.

3. Grant Capability: This grants a capability token to a process, letting it access resources or services.

4. Revoke Capability: This revokes a previously granted capability from a process.

5. Allocate Resource: This requests a resource allocation from the kernel, like memory or compute time.

6. Free Resource: This frees a previously allocated resource back to the kernel.

These cover the essentials: messaging, capabilities, and resource management. That keeps your ABI lean and adaptable across different environments.  
  
# Citations:  
  
---  


# Give me the pros and cons of that design.  
  
# Answer (turbo):  
If you reduce it down to just “Send Message” and “Receive Message,” you’re going full minimalist. Here are the pros and cons:

Pros:
1. Simplicity: The ABI is super simple, making it easier to implement across different environments.
2. Flexibility: You can build any higher-level protocol you want on top of just those two calls.
3. Uniformity: Everything is a message, so the system is highly consistent.

Cons:
1. Performance: Every operation—like allocating memory or checking capabilities—has to go through message passing, which adds overhead.
2. Complexity at higher levels: All the higher-level functionality has to be implemented as messages, which adds complexity to application development.
3. Security overhead: You’d likely need more robust message validation and parsing, which can introduce risks or additional overhead.

So, it’s a trade-off between simplicity and performance. The two-call model is very pure and flexible, but you pay a price in terms of efficiency and developer convenience. It depends on your priorities and how much overhead you can tolerate.  
  
# Citations:  
  
---  


# I'm not sure if we would need a third thing related to standard error, I'm not sure how that would even fit into this kind of an ABI.  
  
# Answer (turbo):  
That’s a good analogy. The two-call model is similar to standard input and output—just generalized for any kind of communication between processes or kernels.

If you wanted to handle errors or control messages separately, you could introduce a third stream—like standard error in Unix systems. That could be a separate kind of message or a special flag in your “Send Message” call to differentiate errors from regular responses.

It all depends on how cleanly you want to keep your messaging model. You could keep everything in just two calls and handle error signaling within the message structure, or add a third call if you want stricter separation. It’s a design choice about simplicity versus expressiveness.  
  
# Citations:  
  
---  


# function calls the same way that mock does it by emulating the syscall interface but doing message passing under the covers.  
  
# Answer (turbo):  
In practice, there’s very little difference. Whether you provide a direct binding to a syscall-like function or you wrap message sending and receiving to emulate higher-level calls, the end result is similar.

It’s mostly about abstraction. Providing direct bindings to specific functions might feel more intuitive to developers familiar with syscall-style APIs. On the other hand, wrapping everything in message handling gives you more flexibility to modify or extend the protocol later without changing the ABI surface.

So, it’s more a design philosophy question than a functional difference. Both approaches end up doing similar work under the hood, just with different levels of abstraction exposed to the developer.  
  
# Citations:  
  
---  


# function ABI that you talked about versus the two functions, just send and receive ABI that we're talking about now.  
  
# Answer (turbo):  
Let’s break it down into two parts to highlight the performance considerations:

First, function calls and stack usage: In a language like Go or C, function calls push arguments and return addresses onto the stack. This is extremely fast and usually just a few instructions at the processor level. Return values are also passed by registers or on the stack, which is highly efficient.

When you compare that to message passing—where you have to construct a message, serialize it (with something like CBOR), send it, and then parse it on the receiving side—there’s naturally more overhead.

Second, the difference between the six-call ABI and the two-call ABI: With the six-call ABI, each specific operation (like capability granting or resource allocation) is a direct function call, which is fast. With the two-call ABI, each operation is a message, so you pay the cost of serializing, sending, receiving, and deserializing. CBOR helps a lot because it’s compact and self-describing, but even then, it’s still higher overhead than a simple function call.

In real terms, the performance difference can be significant if you’re doing a lot of small, frequent operations. Function calls are typically nanoseconds-level, while message passing can be microseconds-level, depending on complexity. But if your operations are large or infrequent, the overhead becomes less noticeable. Ultimately, it depends on the workload and whether the flexibility of messages outweighs the small performance hit for your use case.  
  
# Citations:  
  
---  


# and its local grid node.  
  
# Answer (turbo):  
Exactly, once you’re running everything in isolated memory spaces or sandboxed environments like WebAssembly, you can’t rely on direct function calls across those boundaries. Even if it’s local, it basically becomes message passing—whether that’s through the operating system’s IPC mechanisms, or through WebAssembly host calls.

So yes, the performance difference between function calls and message passing is reduced or eliminated when you’re prioritizing isolation and security. In that case, having a unified message interface makes even more sense, because the performance gap is no longer the main factor. Instead, you’re focusing on security, modularity, and consistency across different environments.  
  
# Citations:  
  
---  


# standard in and standard out of the Docker container itself.  
  
# Answer (turbo):  
Exactly. In a Docker container scenario, the messaging could just be CBOR-encoded messages flowing through the container’s standard input and output. That’s a super clean and portable approach because it works the same way whether you’re running in a container, a VM, or on bare metal.

That also reinforces the idea that having a simple, consistent messaging interface can scale across all these different scenarios without much customization. And it gives you that secure sandboxing by default.  
  
# Citations:  
  
---  


# write a detailed report based on the entire conversation, inferring questions, detailing answers, and reaching conclusions  
  
# Answer (Auto):  
# API vs ABI: A Comprehensive Analysis of Interface Design for Distributed Computing

Based on a comprehensive technical discussion between an AI assistant and Steve Traugott, this report explores the fundamental differences between APIs and ABIs, their application in distributed computing architectures, and design considerations for next-generation systems like PromiseGrid.

## **Key Technical Distinctions**

### **API (Application Programming Interface) vs ABI (Application Binary Interface)**

The conversation established that **ABI (Application Binary Interface)** is the more precise term for low-level system interactions, particularly between compiled modules and kernels. While APIs operate at the source code level and define function signatures and protocols for developers, ABIs govern the binary-level interactions including calling conventions, data structure layouts, and system call mechanisms.[1][2][3]

For WebAssembly modules interacting with their runtime environment, the ABI terminology is appropriate because these interactions occur at the compiled binary level. **WASI (WebAssembly System Interface)** serves as a prime example of a standardized ABI that provides portable system interfaces across different host environments.[4][5][6][7][8]

## **WebAssembly Host Functions and ABI Design**

WebAssembly host functions are indeed part of the ABI, defining how compiled WebAssembly modules invoke functions provided by the host environment. These functions operate through a well-defined interface that specifies parameter passing, memory management, and return value handling at the binary level.[6][4]

The discussion revealed that **capability-based security tokens** can be naturally integrated with function calls in this model. In PromiseGrid's architecture, these tokens would represent permissions or resources (such as compute credits or personal currencies) that can be passed alongside function parameters, enabling fine-grained access control in a decentralized environment.[9][10]

## **Microkernel Architecture and Message Passing**

The conversation explored how microkernel systems like **Mach** implement message passing as the fundamental communication primitive. In Mach, traditional system calls are converted into low-level message passing between user processes and kernel services, creating a more modular and flexible architecture.[11][12][13]

This design pattern offers several advantages for distributed systems:
- **Modularity**: Services can be isolated and independently managed
- **Security**: Message passing naturally supports capability-based access control
- **Scalability**: The same interface works locally and across network boundaries
- **Fault tolerance**: Service failures can be contained and recovered from

## **Decentralized Function Execution Models**

The discussion identified two primary approaches for implementing distributed computing:

### **1. Traditional RPC-Style Function Calls**
- Direct function invocation with parameters and return values
- Higher performance for simple operations
- More tightly coupled architecture
- Familiar to developers

### **2. Message-Passing Architecture**
- Asynchronous, event-driven communication
- Better suited for decentralized environments
- More flexible but potentially higher overhead
- Natural support for capability tokens

## **Performance Considerations in Isolated Environments**

A critical insight emerged regarding performance trade-offs in sandboxed environments. When processes run in isolated memory spaces (such as WebAssembly modules, Docker containers, or separate processes), the performance difference between function calls and message passing diminishes significantly. This is because:[14][15]

- Cross-boundary communication already requires marshaling/unmarshaling
- Security isolation prevents direct memory access
- The overhead of message formatting becomes relatively small

This finding suggests that **message passing can be the optimal choice for distributed systems prioritizing security and modularity**.

## **Content-Addressable Code Distribution**

The conversation explored using **content-addressable storage** (like IPFS) for distributing executable code in a decentralized system. This approach offers several benefits:[16][17][18]

- **Deduplication**: Identical code is stored only once across the network
- **Integrity**: Content hashing ensures code hasn't been tampered with
- **Availability**: Code can be retrieved from any node that has it cached
- **Versioning**: Different versions naturally have different content hashes

## **Proposed ABI Design for Distributed Systems**

The discussion converged on a minimalist ABI design consisting of just two core functions:
1. **Send Message** - Transmits structured data to another process/service
2. **Receive Message** - Receives incoming structured data

### **Advantages of the Two-Call Model:**
- **Extreme Simplicity**: Minimal interface reduces implementation complexity
- **Maximum Flexibility**: All higher-level protocols can be built on top
- **Uniform Semantics**: Same interface works locally and remotely
- **Natural Capability Integration**: Tokens can be embedded in messages

### **Challenges:**
- **Performance Overhead**: Every operation requires message serialization
- **Development Complexity**: Higher-level abstractions must be built in userspace
- **Error Handling**: Must be implemented within the message protocol

## **Docker Container Integration**

The conversation identified that this message-passing approach maps naturally to **Docker container architectures**, where CBOR-encoded messages could flow through standard input/output streams. This provides:[15][14]

- **Universal Interface**: Works with any containerization technology
- **Security**: Natural sandboxing through container isolation  
- **Portability**: Same interface across different host environments
- **Simplicity**: No special IPC mechanisms required

## **Conclusion and Recommendations**

The analysis suggests that for distributed computing systems like PromiseGrid, a **message-passing ABI built on capability-based security** offers the best balance of flexibility, security, and implementation simplicity. While there are performance trade-offs compared to direct function calls, these become negligible in isolated execution environments.

The proposed two-function ABI (Send/Receive Message) provides a solid foundation for building decentralized computing platforms, with the capability to layer higher-level abstractions as needed while maintaining the core benefits of modularity, security, and scalability inherent in message-passing architectures.

This approach aligns well with emerging trends in **serverless computing** and **decentralized function execution**, positioning such systems to take advantage of both centralized cloud infrastructure and distributed peer-to-peer networks as requirements and opportunities demand.[19][20][21]  
  
# Citations:  
[1] [API vs ABI: Why Both Matter in Software Development - Learn Steps](https://www.learnsteps.com/api-vs-abi-why-both-matter-in-software-development/)  
[2] [Application binary interface - Wikipedia](https://en.wikipedia.org/wiki/Application_binary_interface)  
[3] [Difference between API and ABI - Stack Overflow](https://stackoverflow.com/questions/3784389/difference-between-api-and-abi)  
[4] [HTTP Handler Application Binary Interface (ABI) - http-wasm](https://http-wasm.io/http-handler-abi/)  
[5] [What is WASI? - Fastly](https://www.fastly.com/learning/serverless/what-is-wasi)  
[6] [How to Extend WebAssembly with Host Functions - SecondState.io](https://www.secondstate.io/articles/extend-webassembly/)  
[7] [WebAssembly System Interface](https://wasi.dev)  
[8] [Standardizing WASI: A system interface to run WebAssembly ...](https://hacks.mozilla.org/2019/03/standardizing-wasi-a-webassembly-system-interface/)  
[9] [Capability-based security - Wikipedia](https://en.wikipedia.org/wiki/Capability-based_security)  
[10] [Moving to Capability-Based Security with Flow: A Critical Evolution ...](https://hackernoon.com/moving-to-capability-based-security-with-flow-a-critical-evolution-in-blockchain-security)  
[11] [[PDF] OS DESIGN PATTERNS II](http://users.cms.caltech.edu/~donnie/cs124/lectures/CS124Lec04.pdf)  
[12] [[PDF] Lessons And Insights From Reimplementing the Mach Microkernel.](https://fosdem.org/2025/events/attachments/fosdem-2025-5490-machina-lessons-and-insights-from-reimplementing-the-mach-microkernel/slides/238162/MACHINA_yuDPxDS.pdf)  
[13] [[PDF] Microkernel Operating System Architecture and Mach](https://people.eecs.berkeley.edu/~kubitron/courses/cs262a-F18/handouts/papers/microkernel-mach.pdf)  
[14] [Communicate between two Docker containers using RTI Connext ...](https://community.rti.com/kb/communicate-between-two-docker-containers-using-rti-connext-dds-and-shared-memory)  
[15] [IPC communication between Docker containers - Stack Overflow](https://stackoverflow.com/questions/44029035/ipc-communication-between-docker-containers)  
[16] [How IPFS works - IPFS Docs](https://docs.ipfs.tech/concepts/how-ipfs-works/)  
[17] [InterPlanetary File System - Wikipedia](https://en.wikipedia.org/wiki/InterPlanetary_File_System)  
[18] [What is Content Addressability? - Pinata](https://pinata.cloud/blog/what-is-content-addressability/)  
[19] [Deep Dive to Orbs Lambda](https://www.orbs.com/Deep-Dive-to-Orbs-Lambda/)  
[20] [Serverless Web3: Building Decentralized Applications with Agility ...](https://tokenminds.co/blog/web3-development/serverless-web3)  
[21] [How to use Web3 Functions to build serverless web3 apps](https://gelato.network/blog/serverless-web3-functions)  
[22] [Understanding API and ABI in Linux OS and Applications](https://www.domainindia.com/login/knowledgebase/443/Understanding-API-and-ABI-in-Linux-OS-and-Applications.html)  
[23] [http-wasm - GitHub](https://github.com/http-wasm)  
[24] [The Complexities of ABI and API in Software Development - CacheFly](https://www.cachefly.com/news/the-complexities-of-abi-and-api-in-software-development/)  
[25] [http-wasm](https://http-wasm.io)  
[26] [What's The State of WASI? - Fermyon](https://www.fermyon.com/blog/whats-the-state-of-wasi)  
[27] [What's the distinction between API and ABI in Linux programming?](https://community.latenode.com/t/whats-the-distinction-between-api-and-abi-in-linux-programming/17347)  
[28] [Module ABI Reference - Docs | SpacetimeDB](https://spacetimedb.com/docs/webassembly-abi)  
[29] [Wasm, WASI, Wagi: What are they? - Fermyon](https://www.fermyon.com/blog/wasm-wasi-wagi)  
[30] [proxy-wasm/spec: WebAssembly for Proxies (ABI specification)](https://github.com/proxy-wasm/spec)  
[31] [WebAssembly/WASI: WebAssembly System Interface - GitHub](https://github.com/WebAssembly/WASI)  
[32] [Differences Between APIs and ABIs | Baeldung on Computer Science](https://www.baeldung.com/cs/apis-vs-abis)  
[33] [WebAssembly - Open Policy Agent](https://openpolicyagent.org/docs/wasm)  
[34] [Serverless Architecture For Decentralized Apps](https://www.meegle.com/en_us/topics/serverless-architecture/serverless-architecture-for-decentralized-apps)  
[35] [Capability-Based Security: Fine-Grained Control and Resilient ...](https://startup-house.com/glossary/capability-based-security)  
[36] [The Serverless Functions Beginner's Guide - Splunk](https://www.splunk.com/en_us/blog/learn/serverless-functions.html)  
[37] [Capability-based Security - ISMS.online](https://www.isms.online/glossary/capability-based-security/)  
[38] [Mach (kernel) - Wikipedia](https://en.wikipedia.org/wiki/Mach_(kernel))  
[39] [[PDF] Message Passing and Microkernels - Jed Crandall](https://jedcrandall.github.io/courses/cse536spring2024/messagepassingmicrokernels.pdf)  
[40] [Capability-based security and RBAC are completely different things ...](https://news.ycombinator.com/item?id=10687104)  
[41] [Mach Overview - Kernel Programming Guide - Apple Developer](https://developer.apple.com/library/archive/documentation/Darwin/Conceptual/KernelProgramming/Mach/Mach.html)  
[42] [How can I replicate a serverless functions in a decentralized way?](https://www.reddit.com/r/ethdev/comments/zbypz6/how_can_i_replicate_a_serverless_functions_in_a/)  
[43] [Combining capability-based security and effects - Reddit](https://www.reddit.com/r/ProgrammingLanguages/comments/v09vit/combining_capabilitybased_security_and_effects/)  
[44] [Microkernels were an immensely popular idea in academia (Mach ...](https://news.ycombinator.com/item?id=4960920)  
[45] [Serverless edge functions for blockchain data - QuickNode](https://www.quicknode.com/functions)  
[46] [[PDF] The Computing and Data Grid Approach: Infrastructure for ...](https://ntrs.nasa.gov/api/citations/20030003699/downloads/20030003699.pdf)  
[47] [Grid Computing's Promises And Perils](https://www.networkcomputing.com/network-infrastructure/grid-computing-s-promises-and-perils)  
[48] [Add support for `ipc: service` syntax · Issue #7220 · docker/compose](https://github.com/docker/compose/issues/7220)  
[49] [The Promise Of Grid Computing | IndustryWeek](https://www.industryweek.com/leadership/companies-executives/article/21945013/the-promise-of-grid-computing)  
[50] [Shared memory : r/docker - Reddit](https://www.reddit.com/r/docker/comments/vc9k6i/shared_memory/)  
[51] [Grid Computing - EDUCAUSE Review](https://er.educause.edu/articles/2005/11/grid-computing)  
[52] [Breaking the chains: Exploring the decentralized power of IPFS](https://blogs.manageengine.com/corporate/general/2024/12/23/breaking-the-chains-exploring-the-decentralized-power-of-ipfs.html)  
[53] [Effective IPC Communication Between Docker Containers - YouTube](https://www.youtube.com/watch?v=KGp8-3piWxY)  
[54] [The Promise of Platform-based Grids - RMI](https://rmi.org/promise-platform-based-grids/)  
[55] [Content Identifiers (CIDs) - IPFS Docs](https://docs.ipfs.tech/concepts/content-addressing/)  
[56] [Kubernetes Basics Tutorial: Host IPC Should Not Be Configured](https://www.fairwinds.com/blog/kubernetes-basics-tutorial-host-ipc-should-not-be-configured)  
[57] [Grid computing: Will it fly? - Hart Energy](https://www.hartenergy.com/ep/exclusives/grid-computing-will-it-fly-22581)  
[58] [The Power of Content-addressing · Decentralized Web Primer](https://flyingzumwalt.gitbooks.io/decentralized-web-primer/content/avenues-for-access/lessons/power-of-content-addressing.html)  
[59] [Docker Socket File for IPC - Manish Pandit's Blog](https://lobster1234.github.io/2019/04/05/docker-socket-file-for-ipc/)  
[60] [Microgrids are making a comeback - here's why. - Promise Energy](https://promiseenergy.com/blogs/microgrids-are-making-a-comeback-heres-why)