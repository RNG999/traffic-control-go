# Linux Traffic Control: A Comprehensive Technical Report

## 1. Introduction to Linux Traffic Control (TC)

Linux Traffic Control (TC) is a sophisticated and powerful framework integrated within the Linux kernel, designed to manage and manipulate network traffic. Its primary purpose is to provide mechanisms for Quality of Service (QoS), enabling administrators to control how packets are queued, scheduled, prioritized, and shaped. This control is essential in diverse networking environments, from individual hosts managing their bandwidth to complex enterprise networks and internet service providers (ISPs) enforcing service level agreements (SLAs).

The core of the TC subsystem resides in the kernel's networking stack, specifically within the `net/sched` component.[1, 2] It operates by intercepting outgoing IP packets before they are handed over to the network interface card (NIC) driver for transmission. For incoming traffic, TC's capabilities are more limited but still offer mechanisms for policing and classification.[3, 4] The userspace utility `tc`, part of the `iproute2` package, serves as the command-line interface for configuring all aspects of the TC framework.[5, 6]

The fundamental building blocks of Linux TC include:
*   **Queueing Disciplines (Qdiscs)**: Algorithms that govern how packets are stored, ordered, and transmitted. They are the primary schedulers in the TC system.[5, 4]
*   **Classes**: Logical subdivisions within "classful" qdiscs, allowing for different treatment of various types of traffic. Classes can be arranged hierarchically.[5, 4]
*   **Filters**: Rules that classify packets based on their headers (e.g., IP addresses, ports) or other metadata (e.g., firewall marks), directing them to appropriate classes or qdiscs.[4, 7]
*   **Actions**: Operations that can be performed on packets, often in conjunction with filters, such as dropping, policing, or modifying packets.[8]

The TC framework allows for the implementation of various QoS strategies, including bandwidth limiting (shaping), rate enforcement (policing), traffic prioritization, and ensuring fairness among different flows. For instance, critical applications like VoIP can be given higher priority and guaranteed bandwidth, while bulk data transfers can be rate-limited to prevent them from congesting the network for other users. The GREE labs articles, for example, discuss using TC to manage bandwidth for server daemons like Redis, particularly during data-intensive operations such as replication.[4, 7]

The development model of the Linux kernel, with its hierarchical maintainership and regular release cycles, ensures that the TC subsystem continues to evolve.[9] The networking subsystem, including `net/sched`, is typically maintained by dedicated maintainers who merge contributions from a wide range of developers.[9, 10] This collaborative approach has led to a rich set of qdiscs and features within TC, catering to a wide array of traffic management needs. More recent developments have focused on advanced AQM (Active Queue Management) qdiscs like FQ-CoDel and CAKE, which aim to combat bufferbloat and provide good default QoS with minimal configuration.[11, 12, 13]

Understanding Linux TC is crucial for network administrators and system engineers who need to optimize network performance, ensure fair resource allocation, and meet specific QoS requirements. While its complexity can be daunting, its flexibility and power are unparalleled in open-source networking.

### 1.1. Purpose: QoS, Bandwidth Management, and Traffic Shaping
Linux Traffic Control (TC) serves multiple critical purposes in network management, primarily revolving around implementing Quality of Service (QoS), managing bandwidth allocation, and shaping network traffic flows. These capabilities allow administrators to exert fine-grained control over how network packets are handled, ensuring that network resources are utilized efficiently and fairly, and that the performance requirements of different applications and users are met.[5, 4]

**Quality of Service (QoS)** is a broad term referring to the ability to provide different priorities to different applications, users, or data flows, or to guarantee a certain level of performance to a data flow.[12] In packet-switched networks, where resources are shared, QoS mechanisms are essential for preventing high-priority, latency-sensitive traffic (e.g., VoIP, online gaming, interactive sessions) from being negatively impacted by high-volume, less critical traffic (e.g., bulk downloads, backups). TC enables QoS by allowing administrators to classify traffic and apply different scheduling policies, ensuring that important packets are expedited.[14]

**Bandwidth Management** involves controlling the amount of network bandwidth consumed by specific types of traffic, users, or applications. This is crucial for preventing network congestion and ensuring that all users receive their allocated share of bandwidth. With TC, administrators can define guaranteed bandwidth rates for certain classes of traffic and also set upper limits (ceilings) on the bandwidth they can consume.[7] This is particularly important for ISPs managing customer subscriptions or enterprises allocating departmental bandwidth. The Hierarchical Token Bucket (HTB) qdisc is a prominent tool for such hierarchical bandwidth allocation, allowing for complex sharing and borrowing schemes.[4]

**Traffic Shaping** is the technique of delaying packets to meet a desired traffic profile, typically to control the rate at which traffic is sent. Shaping is an egress (outgoing) phenomenon; it smooths out traffic bursts and ensures that the traffic conforms to a configured rate, preventing sustained overloads on downstream network devices or links.[3, 4] Qdiscs like TBF and HTB employ token bucket mechanisms to achieve shaping.[4] By queuing excess packets and transmitting them later when capacity is available, shaping helps in avoiding packet loss that might occur if traffic were sent unthrottled. This is distinct from **policing**, which typically drops packets that exceed a configured rate and is often applied to ingress (incoming) traffic.[15]

The ability to perform these functions makes TC indispensable for:
*   **Preventing Bufferbloat**: Modern AQM qdiscs available in TC, like FQ_CODEL and CAKE, are specifically designed to reduce excessive buffering in network devices, which leads to high latency and poor interactivity.[13, 16]
*   **Enforcing Service Level Agreements (SLAs)**: ISPs and service providers use TC to ensure customers receive the bandwidth and service quality they pay for.[17]
*   **Optimizing Application Performance**: Prioritizing latency-sensitive applications over bulk traffic can significantly improve user experience.
*   **Fair Resource Sharing**: Ensuring that no single user or application can monopolize network resources, especially on shared links.

The TC framework's modular design, allowing for different qdiscs, classes, and filters to be combined, provides a highly flexible toolkit to address these diverse network management objectives.[5]

### 1.2. Scope: Egress and Ingress Traffic Control
The Linux Traffic Control (TC) framework offers distinct capabilities for managing both egress (outgoing) and ingress (incoming) network traffic, though the extent and nature of control differ significantly between the two.

**Egress Traffic Control:**
This is where TC provides its most comprehensive and powerful features. Egress control applies to packets that are leaving the Linux system. Because the system is the source of these packets or is forwarding them, it has full control over when and how they are transmitted. This allows for:
*   **Shaping**: Delaying packets to conform to a specific rate, smoothing out bursts, and preventing network congestion downstream.[4] Qdiscs like HTB, TBF, and HFSC are primarily egress shapers.
*   **Scheduling**: Deciding the order in which packets are sent. This includes prioritization (e.g., with PRIO qdisc) and fair queuing (e.g., with SFQ, FQ_CODEL, CAKE).[4, 12, 13, 18]
*   **Classification**: Sorting outgoing packets into different classes based on various criteria (IP addresses, ports, protocol, firewall marks) using filters, so that different scheduling or shaping policies can be applied.[4, 7]

Most classful qdiscs and complex scheduling algorithms are designed for egress traffic. The root qdisc attached to a network device's egress path is the entry point for all outgoing packets on that interface.[5, 19]

**Ingress Traffic Control:**
Controlling incoming traffic is inherently more challenging because by the time packets arrive at a network interface, they have already consumed bandwidth on the upstream link.[3, 4] True "shaping" of ingress traffic (i.e., telling the sender to slow down proactively) is not possible from the receiver's end using TC alone. Instead, ingress TC mechanisms are reactive.

Linux provides a special "ingress qdisc" (handle `ffff:0`) to which filters can be attached.[12, 15, 20] These filters can classify incoming packets and apply actions. The most common action for ingress control is **policing**.
*   **Policing**: This involves checking if incoming traffic conforms to a defined rate. Packets exceeding this rate are typically dropped, or they can be remarked (e.g., DSCP values) or reclassified.[15, 20] The `police` action is frequently used for this purpose. An example from [20] and [15] shows policing ingress traffic to 1 Mbit/s:
    ```bash
    # tc qdisc add dev eth0 handle ffff: ingress
    # tc filter add dev eth0 parent ffff: u32 \
         match u32 0 0 \
         police rate 1mbit burst 100k
    ```
    This configuration adds an ingress qdisc to `eth0` and attaches a filter that matches all packets (`match u32 0 0`), applying a police action to limit the rate to 1 Mbit/s with a burst of 100k. Packets exceeding this rate would typically be dropped (the default exceed action for police is often `reclassify`, but in practice, for ingress rate limiting, `drop` is implied or explicitly configured).

The limitations of direct ingress shaping led to the development of mechanisms like the **Intermediate Queuing device (IMQ)**.[3, 21] IMQ allows packets (including incoming ones marked by `iptables`) to be redirected to a virtual IMQ interface, where egress qdiscs can then be applied. This effectively allows for "shaping" of ingress traffic after it has been received by the kernel but before it's passed to local applications or forwarded. However, IMQ requires kernel patches and its usage has become less common with the advent of more sophisticated qdiscs like CAKE, which can be applied to an Intermediate Functional Block (IFB) device to process ingress traffic redirected via `mirred` actions.[13, 22]

In summary, while egress TC offers proactive shaping and scheduling, ingress TC is primarily reactive, focusing on policing and classification of already-arrived packets. Effective management of incoming bandwidth often relies on these reactive measures or more complex setups involving IFB devices or, historically, IMQ.

## 2. Fundamental TC Elements in Detail

The Linux Traffic Control framework is built upon several core components that work in concert to manage network traffic. These are queueing disciplines (qdiscs), classes, filters, and actions. Understanding each of these elements is crucial for effectively configuring and utilizing TC.

### 2.1. Queueing Disciplines (Qdiscs)

A Queueing Discipline (qdisc) is the heart of the Linux traffic scheduler. It is an algorithm that manages the queue of packets for a network interface and determines the sequence and timing of their transmission.[5] Every network interface in Linux has a qdisc associated with its egress path, and optionally, an ingress qdisc.[19, 12]

#### 2.1.1. The `enqueue` and `dequeue` Paradigm
Qdiscs fundamentally operate through two primary operations: `enqueue` and `dequeue`.[4]

*   **Enqueue**: This is the process where an outgoing IP packet, having traversed the network stack, is passed to the qdisc for storage.[4] The qdisc decides whether to accept the packet into its queue(s) or to drop it. An enqueue operation can fail if the qdisc is full or if its internal logic (e.g., an Active Queue Management algorithm) dictates that the packet should be dropped to signal congestion.[4] The kernel's `sch_api.c` defines `enqueue` as a major routine for every qdisc, returning a status code indicating success or the reason for a drop (e.g., `NET_XMIT_DROP`).[1, 23] The success or failure of an enqueue operation is a critical feedback mechanism. For instance, AQM qdiscs like CoDel (integral to FQ_CODEL and CAKE) intentionally drop packets during enqueue if persistent queue delay is detected, signaling TCP to reduce its sending rate before buffers become excessively large.[12, 13]

*   **Dequeue**: This is the process where the qdisc selects a packet from its queue(s) and passes it to the network device driver for actual transmission onto the wire.[4] A dequeue operation might not always yield a packet; it can "fail" if the qdisc is empty or if the qdisc's algorithm decides not to send a packet at that particular moment (e.g., a shaper waiting for tokens).[4] The `sch_generic.c` file contains generic scheduler routines, including `dequeue_skb`, which is responsible for extracting a packet from the qdisc.[24] The logic embedded within the dequeue operation embodies the "scheduling" aspect of the qdisc. It's at this point that decisions about packet ordering (FIFO, priority-based, fair-queuing) and timing (shaping) are implemented.[4]

The interplay between enqueue and dequeue defines the behavior of the qdisc and its impact on network traffic characteristics such as latency, throughput, and fairness.

#### 2.1.2. Classless Qdiscs: Simple Schedulers
Classless qdiscs are schedulers that do not have configurable internal subdivisions or classes. They treat all packets passing through them in a uniform manner, applying a single scheduling policy to the entire stream of traffic.[4] These are often simpler to configure and are suitable for basic traffic management tasks or as leaf qdiscs within more complex classful hierarchies.

*   **`pfifo_fast` (and `pfifo`, `bfifo`)**:
    *   **Mechanism**: `pfifo_fast` is the traditional default qdisc for network interfaces in Linux.[5, 4] It implements a First-In, First-Out (FIFO) queuing strategy but with a layer of prioritization. It consists of three internal queues, referred to as "bands" (band 0, band 1, band 2). Packets are classified into these bands based on the Type of Service (TOS) bits in their IP headers. Band 0 has the highest priority; packets from band 0 are always dequeued before packets from band 1, and band 1 before band 2. Within each band, packets are processed in FIFO order.[4]
        The `pfifo` qdisc is a simpler variant that implements a strict FIFO queue with a limit defined by the number of packets. `bfifo` is similar but its limit is defined in bytes.[4]
    *   **Parameters**: For `pfifo_fast`, the number of bands (typically 3) and the `priomap` (which defines how TOS values map to bands) are key, though often used with defaults.[12] For `pfifo` and `bfifo`, the main parameter is `limit`.
    *   **Use Cases**: `pfifo_fast` serves as a basic, low-overhead default scheduler. It provides rudimentary QoS if applications correctly set TOS bits. `pfifo` or `bfifo` can be used as simple buffer limiters.
    *   **Limitations**: `pfifo_fast` offers limited configurability and its effectiveness depends on consistent TOS marking by applications. It does not address issues like bufferbloat or ensure fairness between flows.

*   **`TBF` (Token Bucket Filter)**:
    *   **Mechanism**: TBF is designed for rate limiting.[4] It operates based on the token bucket model. Imagine a bucket that can hold a certain number of "tokens." Tokens are added to this bucket at a constant, configurable rate (the `rate` parameter). When a packet is to be transmitted, the TBF qdisc checks if there are enough tokens in the bucket to "pay" for the packet's size. If sufficient tokens are available, they are consumed, and the packet is dequeued immediately. If not, the packet is typically queued (up to a `limit`) or, if the queue is full or the delay excessive, dropped. This mechanism allows for short bursts of traffic exceeding the average rate, up to the size of the token bucket (`burst` parameter).[4, 25]
    *   **Parameters**:
        *   `rate`: The sustained rate at which tokens are generated (e.g., `1mbit`).
        *   `burst` (or `buffer` or `maxburst`): The size of the token bucket, in bytes. This determines the maximum burst size.
        *   `limit` (or `latency`): The maximum number of bytes that can be queued in TBF waiting for tokens. Packets arriving when this limit is exceeded are dropped.
        *   `peakrate`: An optional parameter to allow for a higher, short-term burst rate by specifying a second, smaller token bucket that refills quickly.
        *   `mtu` (or `minburst`): The minimum amount of tokens that must be available, often related to the MTU.
    *   **Use Cases**: Simple and effective for enforcing a hard cap on the bandwidth usage of an entire interface or, when used as a leaf qdisc in a classful system, for a specific category of traffic.[25] It is effective for preventing a link from being overwhelmed.
    *   **Characteristics**: TBF is relatively simple to configure for basic rate limiting. While it can introduce latency if its `limit` is set high, it's also capable of dropping packets, which can be beneficial for signaling congestion.[26]

*   **`SFQ` (Stochastic Fairness Queuing)**:
    *   **Mechanism**: SFQ aims to provide fair bandwidth allocation among a large number of concurrent flows.[4] It achieves this by hashing incoming packets into one of a configurable number of FIFO sub-queues. Each flow (typically identified by source/destination IP and ports for TCP/UDP) is ideally mapped to its own sub-queue. SFQ then services these sub-queues in a round-robin fashion, dequeuing a certain amount of data (related to `quantum`) from each active queue in turn before moving to the next.[4] This prevents any single aggressive flow from monopolizing the outgoing bandwidth and starving other flows. The "stochastic" nature refers to the hashing, which may result in occasional collisions (multiple flows in the same sub-queue), but these are managed, and the hash function can be perturbed periodically to redistribute flows.[4]
    *   **Parameters**:
        *   `perturb`: The interval (in seconds) at which the hashing algorithm is reconfigured to prevent long-term unfairness due to hash collisions (default is 0, meaning no perturbation, or often 10 seconds if enabled).
        *   `quantum`: The number of bytes a flow is allowed to dequeue in one round-robin turn before SFQ moves to the next flow. Defaults to the interface MTU. This is only used if GSO/TSO (Generic Segmentation Offload/TCP Segmentation Offload) is disabled or packets are not GSO.
        *   `flows`: The number of hash buckets (and thus potential sub-queues) to maintain.
        *   `depth`: The maximum number of packets allowed in an individual sub-queue.
    *   **Use Cases**: Commonly used as a leaf qdisc in classful systems (like HTB) to ensure fairness among multiple user connections or application streams that are grouped into a single traffic class.[7, 26] It's effective in environments with many competing flows, such as web servers or shared internet gateways.
    *   **Characteristics**: SFQ is good at preventing flow starvation and ensuring a degree of fairness. However, the hashing and round-robin processing can introduce a small amount of additional latency compared to a simple FIFO queue.[4]

These classless qdiscs, while relatively simple on their own, are fundamental. They not only serve for basic traffic control scenarios but also act as the terminal queuing mechanisms within the leaves of more sophisticated classful qdisc hierarchies, where they manage the actual packets belonging to a specific traffic category defined by the parent classful qdisc. The choice of such a leaf qdisc (e.g., `pfifo` for basic queuing within a rate-limited HTB class, or `sfq` for ensuring fairness among users within that class) can significantly influence the perceived performance for that traffic category.[7]

#### 2.1.3. Classful Qdiscs: Introduction to Hierarchical Structures
Classful qdiscs represent a more advanced tier of traffic scheduling in Linux, enabling the creation of intricate, hierarchical traffic management policies.[4] Unlike classless qdiscs that apply a single scheduling rule to all traffic, classful qdiscs can contain multiple "classes." Each class can, in turn, have its own specific bandwidth allocations, priorities, and even its own child qdisc.[5] This hierarchical structure allows for granular control over different types of network traffic.

The core idea is to build a tree-like structure where the root qdisc is attached to the network device. This root qdisc can then have several child classes. These child classes might further branch into more specific sub-classes, or they can be leaf classes that have a simpler (often classless) qdisc attached to them to handle the actual packet queuing.[4, 19]

To direct packets to the appropriate class within this hierarchy, **filters** are used.[4] Filters examine packet characteristics (like source/destination IP, port numbers, protocol, or firewall marks) and decide which class the packet belongs to. Once a packet is classified into a leaf class, it is typically enqueued into the qdisc attached to that leaf class. The classful qdisc then orchestrates how these leaf qdiscs are serviced, implementing the overarching QoS policy (e.g., guaranteed rates, borrowing, priorities).

The hierarchical nature of classful qdiscs is what provides their power and flexibility. It allows for:
*   **Aggregate Control**: Bandwidth limits or guarantees can be set at higher levels of the hierarchy, which are then shared or subdivided among lower-level classes.
*   **Differentiated Services**: Different types of traffic (e.g., interactive vs. bulk) can be segregated into different classes and receive different treatment.
*   **Resource Sharing**: Mechanisms like bandwidth borrowing (prominently in HTB) allow classes to utilize unused bandwidth from their parent or siblings, leading to efficient link utilization.[4]

Prominent examples of classful qdiscs include HTB (Hierarchical Token Bucket), PRIO (Priority), CBQ (Class Based Queuing), and HFSC (Hierarchical Fair Service Curve), each offering different mechanisms and trade-offs for traffic management. The design of classful qdiscs inherently relies on a robust classification mechanism (filters) to be effective; without filters, traffic would typically fall into a default class, largely negating the benefits of the hierarchical structure.[4, 7]

### 2.2. Classes within Classful Qdiscs

Classes are fundamental components of classful qdiscs, representing distinct categories or flows of traffic that can be managed with specific rules and parameters.[5, 4] They form the branches and leaves of the hierarchical tree structure that classful qdiscs enable.

#### 2.2.1. Defining Traffic Categories
Within a classful qdisc like HTB or HFSC, classes are created to segregate network traffic based on desired criteria. For example, an administrator might create separate classes for:
*   Interactive traffic (e.g., SSH, VoIP)
*   Bulk data transfers (e.g., FTP, backups)
*   Web browsing traffic
*   Traffic from specific departments or users

Each class can then be configured with parameters specific to the parent qdisc's capabilities. For HTB, this would include `rate` (guaranteed bandwidth) and `ceil` (maximum bandwidth).[7] For HFSC, this would involve defining service curves.[27] This allows for differentiated treatment, ensuring that high-priority traffic receives preferential handling or that certain types of traffic do not exceed their allocated bandwidth. The process of creating these categories is shown in configuration examples where `tc class add...` commands are used to define each specific class under a parent qdisc or another class.[7]

#### 2.2.2. Class Identification (`major:minor` handles)
Each class within the TC hierarchy is uniquely identified by a handle, commonly referred to as a `classid`. This identifier typically follows a `major:minor` format.[4, 19]
*   The `major` number usually corresponds to the handle of the parent qdisc to which the class is being added. For instance, if a root qdisc has the handle `1:`, its direct child classes will have classids like `1:1`, `1:10`, `1:20`, etc.
*   The `minor` number is a unique identifier for that class within the scope of its parent `major` number.

This `major:minor` numbering scheme is not merely an identifier; it implicitly defines the hierarchical relationship. The `major` part links a class to its parent qdisc or parent class, forming the tree structure that is essential for hierarchical scheduling and resource allocation.[19] For example, a class `1:10` is a child of the entity identified by `1:`. If `1:10` itself is a classful entity (e.g., an intermediate HTB class), it could have children like `10:1`, `10:2` (though more commonly, if `1:10` is a child of `1:`, its children would be `1:100`, `1:101` if the parent `1:10` is an HTB class, where the major number of the child's `classid` matches the major number of the parent class's qdisc, and the minor number is extended, or it could follow a different scheme if `1:10` has a qdisc with a new major handle attached to it, e.g. `tc qdisc add... parent 1:10 handle 100:...`). The precise structure depends on the qdisc type and configuration.

Leaf classes in a hierarchy are particularly important. These are the classes at the terminal points of the tree that do not have further child classes of the same classful qdisc type. Instead, they typically have a simpler, often classless, qdisc (like `pfifo` or `sfq`) attached to them.[7] This attached qdisc is responsible for the actual queuing of packets that have been classified into that leaf class. The parent classful qdisc then makes scheduling decisions based on the state and configuration of these leaf classes and their attached qdiscs.[4]

### 2.3. Filters for Packet Classification

Filters are the mechanism by which Linux Traffic Control directs packets to their appropriate classes within a classful qdisc hierarchy, or triggers specific actions.[4, 7, 19] Without filters, a classful qdisc would typically send all traffic to a default class, thereby failing to leverage its hierarchical capabilities for differentiated traffic management.

#### 2.3.1. Role in Directing Packets to Classes
When a packet arrives at a classful qdisc (or a class that itself supports classification), the attached filters are evaluated sequentially, usually based on a `priority` parameter.[7] Each filter contains rules to match specific characteristics of the packet. If a packet matches a filter's criteria, the filter specifies a `flowid` which directs the packet to a particular class (identified by its `classid`) within the qdisc's hierarchy.[4] Alternatively, a filter can trigger an action directly. This classification process is fundamental for ensuring that, for example, VoIP packets are sent to a high-priority class and bulk download packets to a lower-priority, rate-limited class.

#### 2.3.2. Common Matching Criteria
Filters can inspect various parts of a packet or its associated metadata. Common criteria include [7]:
*   **Source and Destination IP Address**: Matching packets based on their origin or intended recipient IP address or subnet (e.g., `match ip src 192.168.1.5/32`, `match ip dst 192.168.1.0/24`).
*   **Source and Destination Port Number**: Matching packets based on TCP/UDP source or destination ports (e.g., `match ip sport 5555 0xffff`, `match ip dport 80 0xffff`). This is useful for identifying specific applications like HTTP (port 80), SSH (port 22), etc.
*   **Protocol**: Matching based on the transport protocol (e.g., TCP, UDP, ICMP) using selectors like `match ip protocol 6 0xff` for TCP.
*   **Firewall Marks (fwmark)**: Integrating with `iptables` or `nftables`, where packets can be marked by the firewall, and TC filters can then match on these marks (e.g., `handle <mark_value> fw`). This allows for complex classification logic to be handled by the firewall, with TC focusing on queuing and scheduling based on those marks.[6]
*   **Type of Service (TOS) / Differentiated Services Code Point (DSCP)**: Matching based on the TOS/DSCP field in the IP header, allowing for prioritization based on these standard QoS markings (e.g., `match ip tos 0x10 0xff`).
*   **Other IP header fields**: Such as TTL, fragment bits, etc.
*   **Interface**: Though less common for classification within a qdisc on a single interface, some filter types can consider the network interface.

#### 2.3.3. The `u32` Classifier: The Workhorse
The `u32` classifier is the most versatile and widely used filter type in Linux TC. It is capable of matching on almost any 32-bit aligned field within a packet header, and even non-aligned fields with careful construction.[28, 29]

*   **Syntax and Basic Operation**:
    The fundamental syntax for a `u32` match is `match u32 <value> <mask> at <offset>`.[28]
    *   `<value>`: The target value to compare against.
    *   `<mask>`: A bitmask applied (logical AND) to the packet data before comparison.
    *   `<offset>`: The byte offset from the beginning of the packet (or sometimes from the beginning of a specific header, like the network header if `protocol ip` is specified) where the 32-bit word for matching is located.
    For example, `match u32 0x00000016 0x0000ffff at nexthdr+0` could match on the first two bytes (destination port) of the next header (e.g., TCP/UDP) if it's 22 (SSH).[29] The kernel essentially extracts a 32-bit word, masks it, and compares it to the value. This is the underlying mechanism for most classifications; `tc` often compiles simpler filter syntaxes into these `u32` operations.[28]

*   **Convenience Selectors**:
    Because manually calculating offsets, values, and masks for common fields like IP addresses and ports is complex and error-prone, `tc` provides "syntactic sugar" – more intuitive selectors for these common cases.[28] Examples include:
    *   `match ip src 192.168.1.0/24`
    *   `match ip dst 10.0.0.5/32`
    *   `match ip sport 8080 0xffff` (matching source port 8080)
    *   `match ip dport 22 0xffff` (matching destination port 22)
    *   `match ip protocol tcp 0xff`
    These are translated by the `tc` utility into the appropriate raw `u32 <value> <mask> at <offset>` commands for the kernel.[7, 28]

*   **Logical ANDing of Multiple Matches**:
    A single `tc filter add... u32...` command can include multiple `match` statements. For the filter as a whole to be considered a match, all individual `match` conditions within that command line must evaluate to true (logical AND).[7, 28] For example:
    ```bash
    tc filter add dev eth0 parent 1:0 protocol ip prio 1 u32 \
      match ip dst 192.168.1.10/32 \
      match ip dport 6666 0xffff \
      flowid 1:20
    ```
    This filter matches traffic destined for IP 192.168.1.10 AND destination port 6666.

*   **Logical ORing of Multiple Matches**:
    To achieve a logical OR, multiple separate `tc filter add` commands are typically used, all directing traffic to the same `flowid` if any of them match. Each filter is evaluated independently based on its priority.[7] For example, to send traffic from source port 5555 OR 5556 to class `1:20`:
    ```bash
    tc filter add dev eth0 parent 1:0 protocol ip prio 1 u32 match ip sport 5555 0xffff flowid 1:20
    tc filter add dev eth0 parent 1:0 protocol ip prio 1 u32 match ip sport 5556 0xffff flowid 1:20
    ```

*   **Linking Filter Lists and Hash Tables for Complex Logic**:
    For scenarios with a very large number of rules or complex conditional logic, `u32` supports more advanced features like linking and hash tables.[28]
    *   **Linking**: A filter item, upon matching, can `link` to another filter list (identified by a handle). Processing then continues in the linked list. If no match in the linked list classifies the packet, control may return to the original list. This allows for creating nested decision trees. Up to 7 levels of nesting are generally supported.[28]
    *   **Hash Tables**: Filter lists are organized within hash tables. The root hash table (e.g., `800:`) is created automatically. Additional hash tables can be created (e.g., `handle 1: u32 divisor 256` creates table `1:` with 256 buckets). A `link` command combined with a `hashkey` can direct a packet to a specific bucket (filter list) within a target hash table based on data from the packet itself (e.g., hashing on source IP). This provides efficient classification when dealing with many distinct values, as hash table lookups can be faster than sequentially evaluating a long linear list of filters.[28]

The `prio` parameter in `tc filter add` commands is crucial. It dictates the order in which filters attached to the same parent are evaluated. Filters with lower priority numbers are checked first.[7] This is essential for creating an unambiguous classification hierarchy, especially when packet characteristics might match multiple filter rules. The first filter (by priority order) that matches a packet will typically determine its fate (e.g., its `flowid`).

The `u32` classifier's ability to match arbitrary 32-bit aligned fields, combined with its support for linking and hashing, grants it immense power. While its raw syntax can be daunting, the convenience selectors make it accessible for common tasks, solidifying its role as the cornerstone of packet classification in Linux TC.

### 2.4. Actions on Classified Packets

Beyond simply directing packets to a class via a `flowid`, filters can also trigger "actions." The `tc actions` framework allows for a modular way to define and apply various operations to packets that match filter criteria.[8]

#### 2.4.1. Overview of `tc actions`
The `tc actions` subsystem allows users to define packet manipulation or disposition routines independently of specific qdiscs or classifiers.[8] Once an action is defined, it can be associated with one or more filters. When a packet matches such a filter, the designated action (or sequence of actions) is executed on that packet.

Actions can perform a wide range of operations, including:
*   Dropping a packet (e.g., `action drop`).
*   Policing a packet stream (e.g., `action police`).
*   Mirroring a packet to another interface (e.g., `action mirred`).
*   Modifying packet headers or metadata (e.g., setting a DSCP value, although some modifications are more complex or handled by specific actions like `ctinfo` for conntrack marks [22]).
*   Passing the packet to a userspace application via netlink.

The `tc actions` command allows for adding, changing, deleting, listing, and getting action specifications.[8] Actions also have a `CONTROL` parameter that dictates how TC should proceed after the action is executed. Common controls include [8]:
*   `reclassify`: Restart classification from the first filter in the current list.
*   `pipe`: Continue to the next action in the current filter's action sequence (default).
*   `drop`: Drop the packet and stop further processing.
*   `continue`: Continue classification with the next filter in the parent's filter list.
*   `ok` or `pass`: Finish classification and return the packet to the calling qdisc for queuing or further processing.

This modular approach means a single, complex action (like a detailed policing rule) can be defined once and then referenced by multiple filters, simplifying configuration and maintenance.

#### 2.4.2. The `police` Action: Parameters and Use for Rate Enforcement
The `police` action is a common and powerful tool for enforcing rate limits on traffic that matches a filter. Unlike shaping qdiscs (like TBF or HTB) which primarily delay packets to conform to a rate, policing typically drops packets that exceed the specified rate, though other dispositions are possible.[15, 20]

*   **Algorithms**: The `police` action can use two main algorithms for byte-rate measurement:
    1.  An internal dual token bucket mechanism.
    2.  An in-kernel sampling mechanism (tuned with the `estimator` filter parameter).[15, 20]
    A similar token bucket approach is used for packet-rate policing.

*   **Key Parameters** [15, 20]:
    *   `rate RATE`: The maximum sustained byte rate (e.g., `1mbit`).
    *   `burst BYTES`: The maximum allowed burst in bytes. An optional cell size (power of 2) can follow a slash.
    *   `pkt_rate RATE`: The maximum sustained packet rate (packets per second).
    *   `pkt_burst PACKETS`: The maximum allowed burst in packets.
    *   `mtu BYTES`: The maximum packet size the policer handles correctly; larger packets are treated as exceeding the rate. Setting this improves precision.
    *   `peakrate RATE`: Allows for a higher burst rate beyond the sustained `rate`.
    *   `overhead BYTES`: Accounts for link-layer overhead in rate calculations.
    *   `linklayer TYPE`: Specifies link layer type (e.g., `ethernet`, `atm`, `adsl`) for accurate overhead calculations.
    *   `conform-exceed EXCEEDACT`: Defines how to handle packets that conform to the rate and those that exceed it.
        *   `EXCEEDACT` / `NOTEXCEEDACT` can be:
            *   `continue`: Do nothing, proceed to next action.
            *   `drop` (or `shot`): Drop the packet. (Default for exceeding packets is often `reclassify`).
            *   `ok` (or `pass`): Accept the packet. (Default for conforming packets).
            *   `reclassify`: Treat as non-matching, try next filter.
            *   `pipe`: Pass to next action in sequence.
            *   `goto chain CHAIN_INDEX`: Jump to a different filter chain.

*   **Use Case: Ingress Policing**:
    A primary application of the `police` action is to enforce rate limits on *ingress* traffic, as true shaping is difficult here. By dropping packets that exceed a certain rate, it prevents the local system or network from being overwhelmed.[15, 20]
    Example for policing all ingress traffic on `eth0` to 1 Mbit/s [15, 20]:
    ```bash
    # Ensure the ingress qdisc is present on the interface
    tc qdisc add dev eth0 handle ffff: ingress

    # Add a filter to match all traffic and apply the police action
    tc filter add dev eth0 parent ffff: protocol all prio 1 u32 \
      match u32 0 0 \
      action police rate 1mbit burst 100k conform-exceed drop/ok
    ```
    In this example, `match u32 0 0` is a catch-all. `conform-exceed drop/ok` means drop exceeding packets and accept conforming ones.

The `police` action offers a more immediate form of rate control than shaping. While shapers queue and delay, potentially increasing latency to smooth traffic, policers tend to drop, which can be a stronger signal to TCP congestion control but might be perceived as more aggressive. The choice between shaping and policing depends on the specific goals and the location in the network path (egress vs. ingress).

---
**Table 2.1: Comparison of Basic Classless Qdiscs**

| Qdisc Name | Primary Mechanism | Key Parameters | Typical Use Case | Pros | Cons |
|---------------|-------------------------------------------------------|-------------------------------------------------|------------------------------------------------------|----------------------------------------------------------------------|-------------------------------------------------------------------------|
| `pfifo_fast` | FIFO with 3 TOS-based priority bands | `bands`, `priomap` (often defaults) | Default qdisc, simple priority based on TOS | Simple, low overhead, default behavior | Limited configurability, relies on correct TOS marking, no fairness |
| `pfifo`/`bfifo` | Simple FIFO queue | `limit` (packets for `pfifo`, bytes for `bfifo`) | Basic buffering with a fixed limit | Very simple, predictable | No prioritization, no fairness, can lead to bufferbloat if limit is high |
| `TBF` | Token Bucket Filter | `rate`, `burst`, `limit`, `peakrate` (opt.) | Simple rate limiting for an interface or class | Effective for capping bandwidth, allows bursts | Classless (if root), can increase latency if `limit` is large |
| `SFQ` | Stochastic hashing to sub-queues, round-robin service | `perturb`, `quantum`, `flows`, `depth` | Fair sharing among many flows (e.g., as leaf qdisc) | Prevents flow starvation, good per-flow fairness | Can add slight latency, hashing collisions possible, not a shaper |

*Data for this table sourced from.[5, 4, 7, 12, 25, 26]*

---

## 3. In-Depth Guide to Key Classful Qdiscs

Classful qdiscs are essential for implementing sophisticated Quality of Service policies, as they allow traffic to be categorized and managed hierarchically. Each classful qdisc offers different mechanisms for bandwidth allocation, prioritization, and sharing.

### 3.1. HTB (Hierarchical Token Bucket)

The Hierarchical Token Bucket (HTB) qdisc is one of the most versatile and widely used classful qdiscs in Linux for managing outgoing network traffic. It enables the creation of complex hierarchical bandwidth allocation schemes, making it suitable for a wide range of QoS scenarios, from simple rate limiting to intricate service level agreement (SLA) enforcement.[5, 4, 7]

#### 3.1.1. Mechanism, Advantages, and Use Cases
**Mechanism**: HTB allows users to define a tree-like structure of classes. Each class in this hierarchy can be assigned a guaranteed bandwidth (`rate`) and a maximum permissible bandwidth (`ceil`). HTB uses token buckets, similar to the TBF qdisc, to control the transmission rate for each class. A key feature is its ability to allow classes to "borrow" unused bandwidth from their parent class, up to their own `ceil` limit, and also from sibling classes if the parent has unallocated capacity. This promotes efficient utilization of the available link bandwidth.[4] Packets are classified into these classes using filters.

**Advantages**:
*   **Hierarchical Structure**: Allows for intuitive mapping of organizational structures or service tiers to bandwidth policies.
*   **Guaranteed Bandwidth**: The `rate` parameter provides a minimum bandwidth assurance for a class when there is demand.
*   **Bandwidth Limits**: The `ceil` parameter enforces a hard upper limit on the bandwidth a class can consume.
*   **Bandwidth Borrowing**: Enables efficient use of idle bandwidth, as classes can exceed their guaranteed rate if spare capacity is available from their parent or siblings, up to their `ceil`.[4]
*   **Flexibility**: Can emulate the behavior of simpler qdiscs like TBF and PRIO (in some respects), making it a general-purpose choice.[7]

**Use Cases**:
*   **ISPs**: Managing bandwidth for different customer subscription tiers, ensuring each customer gets their contracted speed while allowing bursts.
*   **Enterprises**: Allocating bandwidth to different departments or applications, prioritizing critical services (e.g., VoIP, video conferencing) over less critical ones (e.g., bulk data transfers).
*   **Shared Internet Connections**: Fairly distributing bandwidth among multiple users or services on a home or small office network.
*   **Server Traffic Management**: Limiting the bandwidth of specific daemons or services to prevent them from saturating the network, as exemplified by GREE's use for Redis replication traffic.[4]

Due to its comprehensive feature set, HTB is often recommended as the default choice for complex bandwidth management tasks if one is unsure which classful qdisc to use.[7]

#### 3.1.2. Building the Class Hierarchy
An HTB configuration starts with attaching the HTB qdisc to a network interface, typically as the root qdisc. Then, classes are added in a hierarchical manner.
1.  **Root Qdisc**: First, the HTB qdisc itself is attached to the device. This qdisc will have a handle (e.g., `1:`). A default class minor ID is specified, to which packets not matching any filter will be directed.
    ```bash
    # tc qdisc add dev eth0 root handle 1: htb default 10
    ```
    This command attaches an HTB qdisc to `eth0`, gives it the handle `1:`, and specifies that unclassified packets should go to the class with minor ID `10` (i.e., class `1:10`).

2.  **Top-Level Class (Root Class of HTB)**: Directly under the root qdisc, one or more top-level HTB classes are defined. These classes define the total bandwidth available to their children. There must be at least one such class that acts as the "root" of the HTB class hierarchy. Its `parent` will be the qdisc handle (e.g., `parent 1:`), and it will have a `classid` (e.g., `classid 1:1`). For this top-level class, `rate` should equal `ceil`, representing the total capacity managed by this HTB instance.[7]
    ```bash
    # tc class add dev eth0 parent 1: classid 1:1 htb rate 1000Mbit ceil 1000Mbit burst 10MB cburst 10MB
    ```

3.  **Intermediate and Leaf Classes**: Further classes can be added as children to existing HTB classes.
    *   **Intermediate Classes**: These classes can have their own `rate` and `ceil` and can have further child classes. They help in subdividing bandwidth.
        ```bash
        # tc class add dev eth0 parent 1:1 classid 1:10 htb rate 500Mbit ceil 800Mbit...
        ```
        Here, class `1:10` is a child of `1:1`. It's guaranteed 500 Mbit/s and can burst up to 800 Mbit/s.
    *   **Leaf Classes**: These are classes at the bottom of a branch in the hierarchy. They do not have HTB child classes but instead have a simpler qdisc (like `pfifo` or `sfq`) attached to them to manage the actual packet queue.[7] Filters direct packets to these leaf classes.
        ```bash
        # tc class add dev eth0 parent 1:10 classid 1:100 htb rate 100Mbit ceil 200Mbit...
        # tc qdisc add dev eth0 parent 1:100 handle 100: pfifo limit 1000
        ```
        Here, `1:100` is a leaf class under `1:10`, and a `pfifo` qdisc with handle `100:` is attached to it.

This structure allows for defining broad allocations at higher levels and more specific ones at lower levels.

#### 3.1.3. Parameters: `rate`, `ceil`, `burst`, `cburst`
These parameters are fundamental to controlling bandwidth in HTB classes. It's crucial to specify units correctly: use `bit`, `kbit`, `mbit`, `gbit` for bits per second, and `b`, `kb`, `mb` for Bytes (for burst values). The `tc` utility interprets `bps` as Bytes per second, which can lead to significant errors if bits per second was intended.[7] Multipliers K, M, G are decimal (10^3, 10^6, 10^9 respectively).

*   **`rate` (Guaranteed Bandwidth)**:
    *   **For the top-level HTB class (e.g., `1:1` directly under qdisc `1:`)**: The `rate` parameter acts as an overall maximum limit for this HTB instance, equivalent to its `ceil`. It is recommended to set `rate` equal to `ceil` for this class.[4, 7] This class defines the total resource pool.
    *   **For child/intermediate/leaf HTB classes (e.g., `1:10`, `1:100`)**: `rate` specifies the bandwidth that is guaranteed to this class if it has traffic to send and the parent has capacity. The sum of the `rate`s of sibling classes should ideally not exceed their parent's `rate` if strict guarantees are needed without relying on borrowing from unallocated parent capacity.[4, 7] A minimal rate (e.g., `1kbit`) must be specified; it cannot be zero for child classes that need to borrow.

*   **`ceil` (Maximum Bandwidth / Ceiling)**:
    *   This parameter defines the absolute maximum bandwidth a class can use, even if its parent has ample spare bandwidth to lend.[4, 7] A class can borrow bandwidth from its parent up to its own `ceil` or the parent's `ceil`, whichever is lower.
    *   The `ceil` of a child class cannot exceed the `ceil` of its parent class.
    *   If `ceil` is not specified for a class, it defaults to the class's `rate`.[4] This means the class cannot burst above its guaranteed rate.

*   **`burst` (Token Bucket for `rate`) and `cburst` (Token Bucket for `ceil`)**:
    *   These parameters define the size of the token buckets, in Bytes, associated with the `rate` and `ceil` speeds, respectively.[7] They determine how much data can be sent in a single burst at a speed higher than the configured `rate` or `ceil` before the token bucket is exhausted.
    *   **Importance**: Properly configured `burst` and `cburst` values are critical for achieving the desired `rate` and `ceil`, especially on links with some latency or for applications that send data in bursts. Kernel timer granularity and packet arrival patterns mean that without sufficient burst capacity, the effective throughput can be much lower than configured.[4, 7]
    *   **Calculation**: A theoretical minimum `burst` (in Bytes) can be estimated as: $burst = \frac{\text{bitrate}_{\text{bits/s}}}{8} \times \text{timer\_resolution}_{\text{s}}$. A common timer resolution assumed for initial calculation is 10ms ($0.01s$).[4, 7] For example, for 100 Mbit/s: $burst = \frac{100 \times 10^6}{8} \times 0.01 = 125000 \text{ Bytes} = 125 \text{kB}$.
    *   **Practical Values**: Practical `burst` and `cburst` values often need to be significantly larger than this theoretical minimum, typically 3 to 20 times, and should be tuned by testing with actual traffic. [7]/[7] provide an example where a calculated 125kB burst for 100Mbit/s only yielded 50Mbit/s, and 1000kB was needed to achieve the full 100Mbit/s. If unsure, especially for high bandwidths, setting a generously large value [7] for all burst parameters can be a starting point, provided sufficient system memory.
    *   `cburst` should generally be at least as large as `burst`.

**Table 3.1: HTB Key Parameter Guide**

| Parameter | Applies To (Context) | Description | Unit | Typical Usage/Calculation Notes |
|-----------|---------------------------|-----------------------------------------------------------------------------|--------------------------|--------------------------------------------------------------------------------------------------------------------------------|
| `rate` | Top-Level HTB Class | Acts as `ceil`; total capacity for this HTB instance. | `bit`, `kbit`, `mbit`, etc. | Should be equal to `ceil`. Defines the overall link speed HTB manages. |
| `rate` | Child/Leaf HTB Class | Guaranteed bandwidth for the class. | `bit`, `kbit`, `mbit`, etc. | Sum of child `rate`s ideally $\le$ parent `rate`. Cannot be 0 if borrowing is expected. |
| `ceil` | Top-Level/Child/Leaf Class | Maximum bandwidth the class can use, including borrowed bandwidth. | `bit`, `kbit`, `mbit`, etc. | Child `ceil` $\le$ parent `ceil`. Defaults to class `rate` if unspecified. |
| `burst` | Top-Level/Child/Leaf Class | Size of token bucket for `rate`, in Bytes. Allows bursting above `rate`. | `b`, `kb`, `mb` (Bytes) | Min calc: $\frac{\text{rate}_{\text{bps}}}{8} \times 0.01s$. Practically 3-20x higher; requires testing. Crucial for achieving `rate`. |
| `cburst` | Top-Level/Child/Leaf Class | Size of token bucket for `ceil`, in Bytes. Allows bursting up to `ceil`. | `b`, `kb`, `mb` (Bytes) | Min calc: $\frac{\text{ceil}_{\text{bps}}}{8} \times 0.01s$. Practically 3-20x higher. Should be $\ge$ `burst`. |
| `quantum` | Top-Level/Child/Leaf Class | Amount of data (Bytes) a class can dequeue per round relative to its `rate`. | Bytes | Default: $\frac{\text{rate}_{\text{Bytes/s}}}{\text{r2q}}$ (`r2q` defaults to 10). Usually defaults are fine. |

*Data for this table sourced from.[4, 7]*

The careful configuration of these parameters, especially `burst` and `cburst`, is paramount. Underestimated burst values are a common reason for HTB configurations failing to deliver the expected throughput, as the system cannot send data fast enough to fill the token buckets adequately due to inherent latencies and discrete processing intervals in the kernel.[7]

#### 3.1.4. Understanding Bandwidth Borrowing and Lending
A core strength of HTB is its sophisticated mechanism for bandwidth borrowing and lending within the defined class hierarchy.[4] This allows for efficient use of the total available bandwidth.

*   **Lending (from Parent to Child)**: A parent class allocates bandwidth to its children based on their `rate` settings. If a child class needs to send traffic and has tokens in its `rate` bucket, it can send. If its `rate` bucket is empty but it needs to send more (and is below its `ceil`), it can "borrow" tokens from its parent class, provided the parent has unutilized capacity (i.e., its own `rate` tokens are available or it can borrow from its parent, and so on up the tree).[4]
*   **Sharing among Siblings**: When a parent class has more bandwidth available than the sum of the `rate`s of its active children, this excess capacity can be shared among sibling classes that wish to send more than their guaranteed `rate`, up to their individual `ceil` limits. The distribution of this excess bandwidth among competing siblings is typically proportional to their configured `rate`s, though `quantum` also plays a role.
*   **`ceil` as the Ultimate Limit**: No matter how much bandwidth is available from parents or siblings, a class can never exceed its own `ceil`.[7] This ensures that even opportunistic bursting is capped.

This borrowing mechanism means that guaranteed rates (`rate`) act as minimums when there is contention, while ceilings (`ceil`) allow classes to opportunistically use available bandwidth, leading to better overall link utilization. For example, if a class for VoIP is guaranteed 2 Mbit/s but is currently idle, that 2 Mbit/s becomes available for other classes (e.g., web browsing, file downloads) to borrow, up to their respective `ceil`s.

#### 3.1.5. `quantum` and `r2q` Parameters
The `quantum` and `r2q` parameters influence how HTB distributes bandwidth among competing classes that are ready to send data, particularly when they are borrowing bandwidth or when the system needs to decide which class to service next in a round-robin like fashion.

*   **`quantum`**: This parameter specifies, in bytes, the amount of data a class should be allowed to dequeue in one go when it's its turn to send, relative to its `rate`.[4] Larger `quantum` values allow a class to send more data per activation, which can be more efficient for high-rate classes but might increase latency for other classes if too large. Smaller `quantum` values lead to finer-grained sharing but increase overhead due to more frequent scheduler interventions.
*   **`r2q` (Rate To Quantum)**: If `quantum` is not explicitly set for a class, HTB calculates it using the class's `rate` and the `r2q` factor: $quantum = \frac{\text{rate}_{\text{Bytes/s}}}{\text{r2q}}$.[4] The default `r2q` value is typically 10.[7]
    *   For most scenarios, the default `r2q` value (and thus the automatically calculated `quantum`) is sufficient. Manually setting `quantum` is usually only necessary for specific tuning requirements or to avoid kernel warnings that can sometimes occur with certain rate combinations if `quantum` becomes too small.[7] The documentation in [7]/[7] suggests that these settings are "usually not necessary."

The `quantum` primarily affects the fairness and responsiveness when multiple classes are competing for bandwidth, especially borrowed bandwidth. It ensures that a class with a higher `rate` gets proportionally more transmission opportunities or larger transmission chunks.

#### 3.1.6. Attaching Internal Qdiscs (e.g., `pfifo`, `sfq`) to Leaf Classes
HTB classes themselves are mechanisms for scheduling and rate control; they do not inherently queue packets for extended periods. The actual queuing of packets occurs in qdiscs attached to the **leaf classes** of the HTB hierarchy.[4, 7] A leaf class is one that has no further HTB child classes.

*   **Necessity**: Every HTB leaf class *must* have a qdisc attached to it to hold packets destined for that class before they are dequeued by HTB's scheduling logic.
*   **Implicit Attachment**: If no qdisc is explicitly attached to an HTB leaf class, the kernel will often implicitly attach a default `pfifo` qdisc, whose size might be tied to the interface's `txqueuelen`.[4] This can be problematic because the parameters of this implicit qdisc (especially its `limit`) are not directly controlled and can lead to unexpected behavior (e.g., excessive packet drops if the implicit limit is too small, or increased latency if too large). It also makes debugging with `tc -s qdisc show` more difficult as the parameters are unknown.[7]
*   **Explicit Attachment (Best Practice)**: It is strongly recommended to explicitly attach a qdisc to every HTB leaf class.[7] This allows for control over its type and parameters. Common choices include:
    *   **`pfifo` or `bfifo`**: A simple FIFO queue with a configurable `limit` (in packets for `pfifo`, in bytes for `bfifo`). Example: `tc qdisc add dev eth0 parent 1:100 handle 100: pfifo limit 1000`. The `limit` parameter is crucial: too small can cause unnecessary drops and underutilization of the HTB class rate; too large can contribute to bufferbloat if the class is consistently congested.[7]
    *   **`sfq` (Stochastic Fairness Queuing)**: If a leaf class serves multiple independent flows (e.g., different users or applications mapped to the same HTB class), attaching an `sfq` qdisc can ensure fair sharing of that class's allocated bandwidth among those flows.[7] However, SFQ can add some processing overhead and potentially slightly increase latency for individual flows compared to `pfifo`.[7]
    *   **`fq_codel`**: For modern AQM benefits (reducing bufferbloat and ensuring fairness) within a leaf class, `fq_codel` can be attached.

The choice of internal qdisc and its parameters (especially `limit`) directly impacts the behavior of the traffic within that HTB leaf class.

#### 3.1.7. Comprehensive `tc` Command Examples for HTB Scenarios

The following examples illustrate common HTB configurations. It's good practice to clear existing root qdiscs with `tc qdisc del dev <iface> root 2>/dev/null | true` before applying a new top-level configuration.

*   **Limiting All Traffic on an Interface** [7]:
    This scenario uses a minimal HTB structure to cap the total bandwidth for `eth0` to 500 Mbit/s.
    ```bash
    # Delete any existing qdisc on eth0 root
    tc qdisc del dev eth0 root 2>/dev/null | true

    # Add HTB as the root qdisc, default traffic to class 1:1
    tc qdisc add dev eth0 root handle 1: htb default 1

    # Create the top-level class defining the total bandwidth
    # For the top-level class, rate must equal ceil
    tc class add dev eth0 parent 1: classid 1:1 htb rate 500Mbit ceil 500Mbit burst 10MB cburst 10MB

    # Attach a pfifo qdisc to the leaf class 1:1
    tc qdisc add dev eth0 parent 1:1 handle 10: pfifo limit 1000
    ```

*   **Limiting Specific Traffic (e.g., by Destination IP)** [7]:
    This example limits traffic destined for IP `192.168.1.10` to 100 Mbit/s, while other traffic can use up to 1000 Mbit/s. The interface itself is assumed to be 1000 Mbit/s.
    ```bash
    tc qdisc del dev eth0 root 2>/dev/null | true
    tc qdisc add dev eth0 root handle 1: htb default 10 # Default traffic to class 1:10

    # Top-level class for the interface
    tc class add dev eth0 parent 1: classid 1:1 htb rate 1000Mbit ceil 1000Mbit burst 10MB cburst 10MB

    # Class for general/other traffic (child of 1:1)
    # Rate is minimal as it's not guaranteed, ceil allows full use
    tc class add dev eth0 parent 1:1 classid 1:10 htb rate 1Mbit ceil 1000Mbit burst 10MB cburst 10MB
    tc qdisc add dev eth0 parent 1:10 handle 100: pfifo limit 1000

    # Class for traffic to be limited (child of 1:1)
    # Rate is minimal, ceil is the desired limit
    tc class add dev eth0 parent 1:1 classid 1:20 htb rate 1Mbit ceil 100Mbit burst 2MB cburst 2MB # Smaller burst for lower rate
    tc qdisc add dev eth0 parent 1:20 handle 200: pfifo limit 1000

    # Filter to direct traffic for 192.168.1.10 to class 1:20
    # parent 1:0 refers to the root qdisc 1:
    tc filter add dev eth0 protocol ip parent 1:0 prio 1 u32 match ip dst 192.168.1.10/32 flowid 1:20
    ```

*   **Limiting and Guaranteeing Specific Traffic (e.g., by Source Port)** [7]:
    This example guarantees 200 Mbit/s to traffic from source ports 5555 and 5556, also capping it at 200 Mbit/s. Other traffic is also guaranteed 200 Mbit/s but can burst up to 1000 Mbit/s if available.
    ```bash
    tc qdisc del dev eth0 root 2>/dev/null | true
    tc qdisc add dev eth0 root handle 1: htb default 20 # Default traffic to class 1:20

    # Top-level class for the interface
    tc class add dev eth0 parent 1: classid 1:1 htb rate 1000Mbit ceil 1000Mbit burst 10MB cburst 10MB

    # Class for prioritized/guaranteed traffic (ports 5555, 5556)
    tc class add dev eth0 parent 1:1 classid 1:10 htb rate 200Mbit ceil 200Mbit burst 4MB cburst 4MB
    tc qdisc add dev eth0 parent 1:10 handle 100: pfifo limit 1000

    # Class for other traffic
    tc class add dev eth0 parent 1:1 classid 1:20 htb rate 200Mbit ceil 1000Mbit burst 10MB cburst 10MB
    tc qdisc add dev eth0 parent 1:20 handle 200: pfifo limit 1000

    # Filters to direct traffic from specified source ports to class 1:10
    tc filter add dev eth0 protocol ip parent 1:0 prio 1 u32 match ip sport 5555 0xffff flowid 1:10
    tc filter add dev eth0 protocol ip parent 1:0 prio 1 u32 match ip sport 5556 0xffff flowid 1:10
    ```
    A more complex version from [7]/[7] involves an intermediate class to manage an overall line guarantee (e.g., 800 Mbit/s) before splitting into specific and other traffic classes. This demonstrates deeper hierarchy.

#### 3.1.8. Monitoring and Verification with `tc -s class|qdisc show dev <iface>`
Verifying that an HTB configuration works as expected is crucial. The `tc` utility provides statistics that are invaluable for this purpose.[7]

*   **`tc -s class show dev <iface>`**: This command displays detailed statistics for each HTB class on the specified interface. Key metrics include [7]:
    *   `Sent <bytes> bytes <packets> pkt`: Cumulative bytes and packets processed by this class. Essential for confirming traffic is flowing through the intended class.
    *   `(dropped <num_dropped>, overlimits <num_overlimits> requeues <num_requeues>)`:
        *   `dropped`: Packets dropped by the leaf qdisc attached to this class (e.g., pfifo `limit` exceeded). If high, the leaf qdisc limit might be too small.
        *   `overlimits`: Number of times the class was prevented from dequeuing packets because it would have violated its `rate` or `ceil` constraints (or those of its parent). This indicates that the shaping/limiting is active.
    *   `rate <curr_rate_bps>bit <curr_pps>pps`: Current actual transmission rate and packets per second for the class.
    *   `backlog <bytes>b <packets>p`: Current number of bytes/packets queued in the leaf qdisc for this class. A persistent backlog indicates congestion for this class.
    *   `lended <pkts>`: Packets this class lent to its children.
    *   `borrowed <pkts>`: Packets this class borrowed from its parent.
    *   `tokens <val>`, `ctokens <val>`: Current state of the token buckets for `rate` and `ceil`.

*   **`tc -s qdisc show dev <iface>`**: This command displays statistics for all qdiscs on the interface, including the root HTB qdisc and any internal qdiscs attached to its leaf classes. For internal qdiscs like `pfifo`, it will show `Sent`, `dropped`, and `backlog`, which helps diagnose if drops are occurring within a specific leaf class's queue.[7]

By observing these statistics while generating relevant network traffic, administrators can confirm:
1.  Filters are correctly classifying traffic into the intended classes (by checking `Sent bytes/pkt` for each class).
2.  Rate limits (`ceil`) are being enforced (indicated by `overlimits` incrementing when the limit is pushed).
3.  Guarantees (`rate`) are being met (by observing the actual `rate` and ensuring no undue `dropped` packets if capacity should be available).
4.  Queues are not excessively dropping packets (low `dropped` count on leaf qdiscs).

Continuous monitoring and iterative tuning, especially of `burst`/`cburst` values and leaf qdisc `limit`s, are often necessary to achieve optimal performance with HTB.[7] The checklist provided in [7]/[7] emphasizes explicitly setting all key parameters (`rate`, `ceil`, `burst`, `cburst` for HTB classes; internal qdiscs and their limits for leaf classes) and verifying behavior in a real environment.

### 3.2. PRIO (Priority Qdisc)

The PRIO qdisc is a classful scheduler designed for strict priority-based traffic management. It is simpler in concept and operation compared to HTB or HFSC, offering a fixed number of priority bands where higher-priority bands are always serviced before lower-priority ones.[18]

#### 3.2.1. Mechanism: Strict Priority Bands
The PRIO qdisc creates a set number of internal classes, referred to as "bands." By default, it creates three bands, but this number can be changed during qdisc creation (e.g., `bands 5`).[18] When packets are ready to be dequeued for transmission:
1.  PRIO attempts to dequeue a packet from the highest priority band (band 0).
2.  If band 0 is empty or chooses not to send a packet, PRIO then attempts to dequeue from band 1.
3.  This process continues sequentially through the bands in descending order of priority (increasing band number).[18]

A lower band number signifies higher priority. For a PRIO qdisc with handle `X:`, its bands are typically addressed as classes `X:1` (highest priority, band 0), `X:2` (band 1), `X:3` (band 2), and so on.[18]

The PRIO qdisc itself is "work-conserving," meaning it does not intentionally delay packets if it has packets to send and the device is ready. Any delay or shaping would be implemented by qdiscs attached as children to its bands.[18] The primary risk with PRIO is **starvation**: if there is a continuous flow of high-priority traffic, lower-priority bands may receive little to no service.[18]

#### 3.2.2. Classification Methods
PRIO uses one of three methods to determine which band a packet should be enqueued into [18]:
1.  **Socket-set Priority**: Applications can set a priority for their sockets, which the kernel can use.
2.  **`priomap` (TOS-based)**: This is a mapping specific to the PRIO qdisc. It uses the Type of Service (TOS) bits from the IP header to assign a packet to a band. The kernel has an internal mapping of TOS values to Linux priority levels (0-15 or more). The `priomap` then maps these Linux priorities to the PRIO bands.[18] For example, the default `priomap` for 3 bands might look like `1 2 2 2 1 2 0 0 1 1 1 1 1 1 1 1`. This array maps Linux priorities 0 through 15 (and higher if defined) to bands; here, Linux priority 0 maps to band 1 (which is PRIO class `X:2`), priority 6 maps to band 0 (PRIO class `X:1`), etc..[18] The `priomap` can be customized when adding the PRIO qdisc.
3.  **`tc filter`**: The most explicit method is to use `tc filter` commands to direct specific traffic flows to particular PRIO bands (classes).[18, 30] This overrides the `priomap` for matched packets. The `flowid` specified in the filter command corresponds to the target PRIO class (e.g., `flowid 1:1` for the highest priority band if the PRIO qdisc handle is `1:`).

#### 3.2.3. Configuration Examples
Setting up PRIO involves adding the `prio` qdisc and then configuring filters if the default `priomap` behavior is not sufficient.

**Basic PRIO setup with 3 default bands:**
```bash
# Delete any existing qdisc on eth0 root
tc qdisc del dev eth0 root 2>/dev/null | true

# Add PRIO qdisc with handle 1: (creates classes 1:1, 1:2, 1:3 by default)
tc qdisc add dev eth0 root handle 1: prio
```
By default, this uses the standard `priomap`. To customize the number of bands and the `priomap`:
```bash
# PRIO with 4 bands and a custom priomap
tc qdisc add dev eth0 root handle 1: prio bands 4 priomap 0 1 1 1 2 2 2 2 3 3 3 3 3 3 3 3
```

**Using filters to direct traffic** [30]:
This example prioritizes SSH (port 22) and MySQL (port 3306) traffic into the highest priority band (`1:1`), and ICMP (protocol 1) also into the highest priority band. Other traffic would fall into lower bands based on the `priomap` or a catch-all filter.
```bash
tc qdisc del dev eth0 root 2>/dev/null | true
tc qdisc add dev eth0 root handle 1: prio

# Send SSH traffic to band 1:1 (highest priority)
tc filter add dev eth0 protocol ip parent 1:0 prio 1 u32 match ip dport 22 0xffff flowid 1:1

# Send MySQL traffic to band 1:1
tc filter add dev eth0 protocol ip parent 1:0 prio 1 u32 match ip dport 3306 0xffff flowid 1:1

# Send ICMP traffic to band 1:1
tc filter add dev eth0 protocol ip parent 1:0 prio 1 u32 match ip protocol 1 0xff flowid 1:1

# Optional: Send all other IP traffic to band 1:2 (medium priority)
# tc filter add dev eth0 protocol ip parent 1:0 prio 10 u32 match ip src 0.0.0.0/0 flowid 1:2
# If no catch-all filter, traffic not matching explicit filters uses the priomap,
# or if it doesn't map, may go to the lowest band by default.
```
The `parent 1:0` in the filter command refers to the PRIO qdisc itself (handle `1:`). The `prio` parameter on the filter command determines the filter's evaluation order, not the traffic priority band it assigns to.

**Attaching shapers to PRIO bands to prevent starvation** [18]:
To mitigate the risk of lower-priority bands being starved, one can attach a shaper like TBF to the higher-priority bands.
```bash
tc qdisc del dev eth0 root 2>/dev/null | true
tc qdisc add dev eth0 root handle 1: prio

# Attach TBF to band 1:1 (highest priority) to limit its rate
tc qdisc add dev eth0 parent 1:1 handle 10: tbf rate 1mbit burst 32kbit limit 3000

# Filters would then direct traffic to flowid 1:1 as before
tc filter add dev eth0 protocol ip parent 1:0 prio 1 u32 match ip dport 22 0xffff flowid 1:1
```
Now, traffic in band `1:1` is still prioritized but cannot exceed 1 Mbit/s, allowing lower bands a chance to transmit.

The PRIO qdisc is a straightforward solution for strict prioritization needs. Its simplicity is an advantage, but the potential for starvation requires careful consideration, especially if high-priority traffic can be voluminous or unconstrained. When combined with filters, it offers precise control over which traffic receives preferential treatment.

### 3.3. CBQ (Class Based Queueing)

Class Based Queueing (CBQ) is one of the older classful qdiscs in Linux, designed with the aim of emulating link-sharing on a fixed-bandwidth link and providing weighted fair queuing capabilities.[6] While powerful, CBQ is known for its complexity in configuration and has, in many common use cases, been superseded by HTB due to HTB's more intuitive model and often better performance characteristics.

#### 3.3.1. Overview of CBQ for Constant Bitrate Control
CBQ attempts to manage traffic such that classes do not exceed their allocated bandwidth when the link is congested. It works by estimating the idle time of the link and using this information to regulate the sending rate of classes. The goal is to ensure that each class receives its configured share of the bandwidth over time.[6] It can be used to define pipes with constant bit rates.

Key concepts in CBQ include:
*   **`avgpkt`**: The average packet size, used in calculations. Typically around 1000 bytes.[6]
*   **`bandwidth`**: The total bandwidth of the physical interface this CBQ instance is managing.[6]
*   **`allot`**: The number of bytes a class is allowed to send each time it is selected for transmission. This is related to its share of the bandwidth.
*   **`rate`**: The desired rate for a class.
*   **`bounded`**: If set, the class cannot borrow bandwidth from its parent, even if available. It is strictly limited to its own `rate`.[6]
*   **`isolated`**: Similar to `bounded`, but also prevents the class from lending its unused bandwidth.

#### 3.3.2. Class Definition and Parameters
A CBQ hierarchy starts with a root CBQ qdisc, and then classes are added beneath it.

**Root CBQ Qdisc Definition** [6]:
```bash
# Define root CBQ qdisc on eth1 with 1000 Mbit/s total bandwidth
tc qdisc add dev eth1 root handle 1: cbq avpkt 1000 bandwidth 1000Mbit
```

**CBQ Class Definition** [6]:
Classes are defined with parameters like `rate`, `allot`, `priority`, and sharing controls like `bounded` or `isolated`.
```bash
# Define class 1:1 under parent 1: (the root qdisc) with a rate of 1 Mbit/s
# 'allot 1500' is a common value, related to MTU size.
# 'bounded' means this class cannot borrow bandwidth.
tc class add dev eth1 parent 1: classid 1:1 cbq rate 1Mbit allot 1500 priority 1 bounded

# Define class 1:2 with a rate of 50 Mbit/s, higher priority (lower number)
tc class add dev eth1 parent 1: classid 1:2 cbq rate 50Mbit allot 1500 priority 2 bounded
```
The `priority` parameter within CBQ classes influences the weighted round-robin (WRR) selection process when multiple classes are ready to send. Lower priority numbers are generally serviced first or get more weight.

#### 3.3.3. Filter Configuration Examples
Filters, typically `u32`, are used to direct traffic to the appropriate CBQ classes based on packet characteristics.

**Filter Examples** [6]:
```bash
# Direct IPv4 TCP traffic to destination port 5001 into class 1:1
tc filter add dev eth1 parent 1:0 protocol ip prio 1 u32 \
  match ip protocol 6 0xff \
  match ip dport 5001 0xffff \
  flowid 1:1

# Direct IPv6 TCP traffic to destination port 5001 into class 1:2
tc filter add dev eth1 parent 1:0 protocol ipv6 prio 2 u32 \
  match ip6 protocol 6 0xff \
  match ip6 dport 5001 0xffff \
  flowid 1:2

# Direct IPv6 traffic with iptables mark 32 into another class (e.g., 1:4)
# This requires a corresponding iptables rule:
# ip6tables -t mangle -A POSTROUTING -p tcp --dport 5003 -j MARK --set-mark 32
tc filter add dev eth1 parent 1:0 protocol ipv6 prio 3 handle 32 fw flowid 1:4
```
The `parent 1:0` refers to the root CBQ qdisc with handle `1:`.

CBQ's reliance on accurate idle time estimation and the intricate interplay of its parameters (`allot`, `weight`, `priority`, sharing settings) make it challenging to configure optimally. While it was a pioneering qdisc for complex QoS, the emergence of HTB provided a more intuitive and often more performant alternative for hierarchical bandwidth management. Consequently, CBQ is less frequently recommended or used in new configurations compared to HTB or more modern qdiscs like CAKE.

### 3.4. HFSC (Hierarchical Fair Service Curve)

The Hierarchical Fair Service Curve (HFSC) qdisc is a classful scheduler designed to provide precise bandwidth and delay guarantees, making it particularly suitable for real-time and latency-sensitive applications.[26, 27]

#### 3.4.1. Core Concept: Service Curves for Precise Guarantees
A service curve is a mathematical function that defines the minimum amount of service (i.e., bandwidth or work) a class should have received by any given point in time `t`.[27] HFSC's primary goal is to ensure that the actual service received by a class does not fall below this curve. This allows for strong guarantees on throughput and delay.

HFSC typically uses either:
*   **Linear service curves**: Defined by a single constant rate (e.g., `rate 1Mbit`).
*   **Two-piece linear service curves**: Defined by an initial rate (`m1`) for a certain duration (`d`), followed by a different rate (`m2`) thereafter. This allows for specifying, for example, an initial burst of bandwidth or a low-latency service for a short period, followed by a sustained rate. Example: `m1 2Mbit d 100ms m2 7Mbit` means 2 Mbit/s for the first 100ms, then 7 Mbit/s.[27]

The service curve concept allows HFSC to offer more deterministic performance compared to token bucket-based systems like HTB, especially for delay bounds.

#### 3.4.2. Realtime (RT) and Link-sharing (LS) Criteria
HFSC employs two distinct criteria for scheduling packets, which can be defined per class [27]:

*   **Realtime (RT) Service Curve**:
    *   This criterion is used to provide strict guarantees on bandwidth and, more importantly, delay. It is primarily applied to leaf classes (those with actual packet queues).
    *   The RT criterion effectively ignores the class hierarchy when making its decisions, focusing solely on meeting the service curve for the specific class.
    *   It calculates an "eligible time" (E()) for each packet, indicating the earliest time it can be sent without violating the curve, and a "deadline time" (D()), indicating the latest time it must be sent. Packets are selected based on the earliest deadline among eligible packets.[27]
    *   This is ideal for applications like VoIP or interactive gaming where low and predictable latency is paramount.

*   **Link-sharing (LS) Service Curve**:
    *   This criterion is used to distribute bandwidth fairly among classes according to the defined hierarchy, similar to how other hierarchical schedulers operate.
    *   Decisions are based on a "virtual time" (V()) associated with each class. The class with the smallest virtual time among active siblings is typically chosen next for transmission.[27]
    *   The LS curve defines a class's share of the bandwidth when competing with siblings. Absolute values of LS curves matter less than their ratios to each other for fair sharing.[27]

*   **Upperlimit (UL) Service Curve (Optional)**:
    *   An optional upper-limit service curve can also be defined for a class. This acts as a cap on the bandwidth a class can consume, even if it could borrow more under the LS criterion. It uses a "fit-time" (F()) calculation.[27]

A class can have both RT and LS (and UL) curves defined. The RT criterion generally takes precedence if a real-time guarantee is at risk of being violated; otherwise, the LS criterion governs scheduling.

#### 3.4.3. Application for Low-Latency Services
The precise delay and bandwidth guarantees offered by HFSC's RT service curves make it exceptionally well-suited for latency-sensitive applications.[26, 27] By defining an appropriate RT curve (e.g., a two-piece curve specifying a low initial delay and sufficient bandwidth for the initial part of a flow), HFSC can ensure that interactive sessions remain responsive even when the link is heavily utilized by other traffic. [26] describes HFSC as the "holy grail of traffic shaping" for its ability to saturate a link while maintaining excellent interactivity for non-bulk sessions. This is often achieved by combining HFSC with a fair queuing qdisc like SFQ on its leaf classes.[26]

Compared to HTB, which primarily manages rates and allows borrowing, HFSC's focus on service curves provides a more direct way to control delay. While HTB tends to queue packets when rates are exceeded (potentially increasing latency), HFSC, when configured for low latency, might be more inclined to drop packets if necessary to meet its service curve guarantees, similar in effect to TBF in that regard.[26]

#### 3.4.4. Configuration Examples
Configuring HFSC involves adding the `hfsc` qdisc and then defining classes with their respective service curves.

**Root HFSC Qdisc:**
```bash
# Delete any existing qdisc on eth0 root
tc qdisc del dev eth0 root 2>/dev/null | true

# Add HFSC as the root qdisc, default traffic to class 1:11 (example)
tc qdisc add dev $WAN_INTERFACE root handle 1: hfsc default 11
```

**Class Definition with Service Curves** [26]:
The script in [26] provides practical examples. A class definition involves specifying `sc` (for a single combined service curve if RT/LS are not separated), or separate `rt`, `ls`, and `ul` curves.
*   `sc rate <rate>`: Defines a simple linear service curve for both RT and LS.
*   `sc umax <bytes> dmax <time_ms> rate <rate>`: Defines a two-part curve often used for initial burst/delay control. `<bytes>` is the max packet size for the initial segment, `<time_ms>` is its max delay, and `<rate>` is the subsequent rate.
*   `ls rate <rate>`: Linear link-sharing curve.
*   `rt m1 <rate1> d <duration_ms> m2 <rate2>`: Two-piece real-time curve.
*   `ul rate <rate>`: Linear upper-limit curve.

Example structure from [26] (conceptual, actual rates are variables in the script):
```bash
# Parent class defining overall link characteristics
tc class add dev $WAN_INTERFACE parent 1: classid 1:1 hfsc sc rate $NEAR_MAX_UPRATE ul rate $NEAR_MAX_UPRATE

# Prioritized traffic class (e.g., ACKs, interactive services)
# Uses 'sc umax <bytes> dmax <delay_ms> rate <guaranteed_rate>' for initial low latency
# and 'ul rate <overall_max_rate>' to allow bursting up to link capacity
tc class add dev $WAN_INTERFACE parent 1:1 classid 1:10 hfsc \
  sc umax 1540 dmax 5ms rate $PRIO_UP_RATE \
  ul rate $NEAR_MAX_UPRATE

# Default/bulk traffic class
tc class add dev $WAN_INTERFACE parent 1:1 classid 1:11 hfsc \
  sc umax 1540 dmax 5ms rate $BULK_UP_RATE \
  ul rate $BULK_UP_RATE # Can also be capped lower than NEAR_MAX_UPRATE

# Attach SFQ to leaf classes for fairness within the class
tc qdisc add dev $WAN_INTERFACE parent 1:10 sfq perturb 10
tc qdisc add dev $WAN_INTERFACE parent 1:11 sfq perturb 10
```
Filters would then be used to direct specific traffic (e.g., ACK packets, SSH traffic, gaming ports) to class `1:10`, with other traffic going to `1:11`.

The complexity of defining appropriate service curves (especially two-piece ones) means HFSC configuration requires a deeper understanding of the traffic characteristics and desired performance than HTB. However, for scenarios demanding stringent latency and jitter control, HFSC offers a powerful solution. The separation of RT and LS criteria allows it to cater simultaneously to traffic with strict real-time needs and traffic that requires fair sharing of remaining bandwidth.

---
**Table 3.2: Comparing Key Classful Qdiscs**

| Qdisc Name | Primary Scheduling Goal | Key Mechanisms | Bandwidth Allocation | Latency Characteristics | Complexity | Typical Use Cases |
|------------|-------------------------------------------------------|--------------------------------------------------------|----------------------------------------------------------|----------------------------------------------------------------------------------------|------------|-----------------------------------------------------------------------------------|
| **HTB** | Hierarchical bandwidth sharing, guarantees & limits | Token buckets, class hierarchy, bandwidth borrowing | `rate` (guarantee), `ceil` (limit), borrowing | Moderate; can increase under load if queues are deep. | Moderate | ISPs, enterprise QoS, complex SLAs, general purpose hierarchical shaping. |
| **PRIO** | Strict priority scheduling | Fixed priority bands | No bandwidth sharing between bands; strict ordering | Low for highest priority; potential starvation for lower priorities. | Low | Simple prioritization of critical traffic (e.g., control packets, interactive SSH). |
| **CBQ** | Emulate link sharing, weighted fair queuing | Idle time estimation, `allot`, `avgpkt`, `bounded` | `rate`, complex sharing rules | Can be good if tuned perfectly, but sensitive to parameters. | High | Older systems, specific constant bitrate emulation (largely superseded by HTB). |
| **HFSC** | Precise delay & bandwidth guarantees, fair sharing | Service curves (Realtime, Link-sharing, Upperlimit) | Defined by service curves; can be very precise | Excellent for low latency (RT curve); good fairness (LS curve). Can drop packets. | High | VoIP, online gaming, real-time services requiring strict delay/jitter control. |

*Data for this table sourced from.[6, 4, 7, 18, 26, 27, 30]*

---

## 4. Modern AQM and Advanced Qdiscs

Traditional queuing disciplines often lead to overly large buffers in network devices, a phenomenon known as bufferbloat, which results in high latency and poor network responsiveness. Modern Active Queue Management (AQM) qdiscs aim to mitigate this by intelligently managing queue lengths and signaling congestion proactively. FQ_CODEL and CAKE are two prominent examples that have gained traction for their effectiveness and ease of use.

### 4.1. FQ_CODEL (Fair Queuing Controlled Delay)

FQ_CODEL is a qdisc that combines Fair Queuing (FQ) with the CoDel (Controlled Delay) AQM algorithm. Its goal is to reduce bufferbloat and ensure fairness among network flows with minimal configuration.[12, 31, 32]

#### 4.1.1. Mechanism: Combining Fair Queuing with CoDel AQM
*   **Fair Queuing (FQ)**: FQ_CODEL stochastically classifies incoming packets into a number of internal queues (representing different flows). Typically, a hash function based on source/destination IP addresses and ports is used for this classification.[12, 31, 32] It then attempts to provide a fair share of the bandwidth to each active flow by servicing these queues in a fair manner (often round-robin or a variant). This prevents a single high-bandwidth flow from monopolizing the link and starving other flows.
*   **CoDel (Controlled Delay) AQM**: Each of these individual flow-queues is managed by the CoDel algorithm. CoDel's primary objective is to keep the *sojourn time* (the time a packet spends in the queue) of packets low and stable. It monitors the minimum sojourn time for packets in each queue over a sliding time window (`interval`). If the sojourn time consistently exceeds a `target` value, CoDel infers persistent congestion and starts dropping packets (or marking them with ECN if enabled) from the head of that queue.[12] This proactive dropping signals TCP to reduce its sending rate, thus preventing the queue from growing excessively large. Reordering within a flow is avoided as CoDel internally uses a FIFO queue for each flow.[31, 32]

This combination ensures that not only is delay kept low (thanks to CoDel), but also that bandwidth is shared fairly among competing flows (thanks to FQ).

#### 4.1.2. Key Parameters
While FQ_CODEL is designed to work well with defaults, several parameters can be tuned [12, 31, 32]:

*   **`limit`**: The hard limit on the total number of packets the FQ_CODEL instance can hold across all its flow queues. Default: 10240 packets.
*   **`memory_limit`**: The hard limit on the total memory (in bytes) that can be consumed by packets queued in this FQ_CODEL instance. The effective limit is the lower of `limit` (converted to bytes) and `memory_limit`. Default: 32 MB.
*   **`flows`**: The number of internal queues (hash buckets) used for flow separation. Default: 1024.
*   **`target`**: CoDel's target for minimum persistent queue delay. If queue delay stays above this target, CoDel starts managing the queue. Default: 5ms.
*   **`interval`**: CoDel's time window over which the minimum delay is tracked and must be experienced. It should be set on the order of the Round Trip Time (RTT) of the path. Default: 100ms.
*   **`quantum`**: The number of bytes a flow is allowed to dequeue in one round of the fair queuing scheduler. Default: 1514 bytes (typical Ethernet MTU + header).
*   **`ecn | noecn`**: Enables or disables Explicit Congestion Notification. Unlike the standalone `codel` qdisc, FQ_CODEL enables ECN by default.[31, 32] If ECN is enabled, CoDel will mark packets with Congestion Experienced (CE) instead of dropping them, if the sender supports ECN.
*   **`ce_threshold`**: A specific delay threshold above which all packets are marked with CE. This is useful for DCTCP-style congestion control algorithms that rely on early ECN signals from very shallow queues.[31, 32]
*   `drop_batch_size` [32]: The maximum number of packets to drop at once when `limit` or `memory_limit` is exceeded. Default: 64.

**Table 4.1: FQ_CODEL Key Parameter Guide**

| Parameter | Description | Default Value | Unit | Snippet Reference |
|------------------|-----------------------------------------------------------------------------|-----------------------|---------|---------------------------|
| `limit` | Hard limit on total packets in the qdisc. | 10240 | packets | [31, 32] |
| `memory_limit` | Hard limit on total memory used by queued packets. | 32 | MB | [31, 32] |
| `flows` | Number of internal queues for flow separation. | 1024 | number | [31, 32] |
| `target` | CoDel's acceptable minimum persistent queue delay. | 5 | ms | [31, 32] |
| `interval` | CoDel's time window for tracking minimum delay. | 100 | ms | [31, 32] |
| `quantum` | Bytes per flow dequeued per fair queuing round. | 1514 (or MTU) | Bytes | [31, 32] |
| `ecn` | Enable Explicit Congestion Notification (marks packets instead of dropping). | Enabled | N/A | [31, 32] |
| `noecn` | Disable Explicit Congestion Notification. | Disabled (if ecn is on) | N/A | [31, 32] |
| `ce_threshold` | Delay threshold for aggressive ECN CE marking (for DCTCP-like algorithms). | (not set by default) | ms | [31, 32] |
| `drop_batch_size`| Max packets to drop at once when limit is exceeded. | 64 | packets | [32] |

#### 4.1.3. Benefits for Reducing Bufferbloat and Ensuring Fairness
FQ_CODEL offers significant advantages:
*   **Reduced Bufferbloat**: By actively managing queue delay through CoDel, it prevents queues from becoming excessively long, which is a primary cause of high latency and jitter on the internet.[12, 16]
*   **Flow Fairness**: The Fair Queuing component ensures that available bandwidth is distributed equitably among concurrent flows, preventing any single flow from dominating the link.[12] This improves responsiveness for interactive applications even when bulk transfers are active.
*   **Simplicity**: FQ_CODEL is designed to work well "out of the box" with minimal tuning for many common scenarios, making advanced AQM accessible without requiring deep expertise in queuing theory.[12] Its adoption as a default qdisc in some Linux distributions and systems like OpenWrt underscores its effectiveness and ease of use.[12]

The stochastic nature of its flow classification means it doesn't require manual filter configuration for per-flow fairness, which would be impractical for devices handling numerous unpredictable flows (like an internet gateway). While hash collisions can occur, leading to multiple flows sharing a single CoDel queue, the overall behavior is generally robust and provides substantial improvements over older, unmanaged FIFO queues.[31, 32]

#### 4.1.4. `tc` Command Examples
Here are common `tc` commands for FQ_CODEL:

*   **Add FQ_CODEL as the root qdisc with default settings** [12, 32]:
    ```bash
    sudo tc qdisc add dev eth0 root fq_codel
    ```

*   **Add FQ_CODEL with some custom parameters** [31]:
    ```bash
    sudo tc qdisc add dev eth0 root fq_codel limit 2000 target 3ms interval 40ms noecn flows 2048 quantum 300
    ```

*   **Show FQ_CODEL statistics** [12, 32]:
    ```bash
    tc -s qdisc show dev eth0
    ```
    This will display configured parameters and runtime statistics, including `Sent bytes/pkt`, `dropped`, `overlimits`, `requeues`, `backlog`, `maxpacket`, `drop_overlimit`, `new_flow_count`, `ecn_mark`, etc..[32]

*   **Delete FQ_CODEL and revert to default qdisc** [12]:
    ```bash
    sudo tc qdisc del dev eth0 root
    ```

FQ_CODEL represents a significant advancement in default queuing behavior, providing a good balance of low latency, fairness, and ease of use.

### 4.2. CAKE (Common Applications Kept Enhanced)

CAKE (Common Applications Kept Enhanced) is a comprehensive, shaping-capable queue discipline designed to provide excellent Quality of Service with minimal configuration. It integrates several advanced features, including a sophisticated AQM, flow isolation, and a precise shaper, aiming to be an "all-in-one" solution for managing network traffic, particularly on internet access links.[11, 13]

#### 4.2.1. Comprehensive Features: Integrated AQM (COBALT), Flow Queuing, Shaper
CAKE bundles multiple functionalities that traditionally required combining several qdiscs and `iptables` rules:
*   **Advanced AQM (COBALT)**: CAKE uses an AQM algorithm called COBALT, which is a hybrid of Codel and BLUE. This algorithm is effective at controlling latency by managing queue buildup

