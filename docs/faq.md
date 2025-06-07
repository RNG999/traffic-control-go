# Traffic Control Go - Frequently Asked Questions

Common questions and answers about using the Traffic Control Go library.

## General Questions

### Q: What is Traffic Control Go?
A: Traffic Control Go is a human-readable Go library for managing Linux Traffic Control (TC). It provides an intuitive API to shape network traffic, set bandwidth limits, and prioritize different types of traffic.

### Q: Do I need root access to use this library?
A: Yes, traffic control operations require the `CAP_NET_ADMIN` capability, which typically means running as root or with elevated privileges. This is a Linux kernel requirement, not a library limitation.

### Q: Which Linux distributions are supported?
A: Any Linux distribution with:
- Kernel 3.10+ (most modern distributions)
- iproute2 package installed
- HTB (Hierarchical Token Bucket) support in the kernel

### Q: Can I use this library in containers?
A: Yes, but the container needs:
```bash
# Docker
docker run --cap-add=NET_ADMIN your-image

# Kubernetes
securityContext:
  capabilities:
    add: ["NET_ADMIN"]
```

## Configuration Questions

### Q: What's the difference between guaranteed and soft limit bandwidth?
A: 
- **Guaranteed bandwidth**: The minimum rate a traffic class will always get
- **Soft limit bandwidth**: The maximum rate a class can use (including borrowed bandwidth)
- **Hard limit bandwidth**: The total interface capacity

```go
// Example: Web traffic gets 100Mbps minimum, up to 300Mbps max
controller.CreateTrafficClass("Web").
    WithGuaranteedBandwidth("100mbps").  // Always available
    WithSoftLimitBandwidth("300mbps")    // Maximum with borrowing
```

### Q: How do priority values work?
A: Priority values (0-7) automatically map to handles:
- **0** = Highest priority → Handle `1:10`
- **3** = Normal priority → Handle `1:13`  
- **7** = Lowest priority → Handle `1:17`

Lower numbers = higher priority.

### Q: Do I need to specify filters for every traffic class?
A: Not necessarily. If you don't specify filters, traffic won't match that class. However, you should always have a default class:

```go
// Catches all unmatched traffic
controller.CreateTrafficClass("Default").
    WithPriority(7). // Lowest priority
    WithGuaranteedBandwidth("10mbps")
```

### Q: Can guaranteed bandwidth exceed the hard limit?
A: No, this will cause a validation error:

```go
// Wrong - will fail
controller.WithHardLimitBandwidth("1gbps")
controller.CreateTrafficClass("A").WithGuaranteedBandwidth("600mbps")
controller.CreateTrafficClass("B").WithGuaranteedBandwidth("600mbps") // Total: 1.2gbps > 1gbps

// Correct
controller.WithHardLimitBandwidth("1gbps")
controller.CreateTrafficClass("A").WithGuaranteedBandwidth("400mbps")
controller.CreateTrafficClass("B").WithGuaranteedBandwidth("400mbps") // Total: 800mbps < 1gbps
```

## API Questions

### Q: How do I match traffic by IP address?
A: Use `ForSource()` or `ForDestination()`:

```go
controller.CreateTrafficClass("Database").
    ForSource("10.0.1.0/24").         // From database subnet
    ForDestination("10.0.2.100")      // To specific server
```

### Q: Can I match traffic by multiple criteria?
A: Yes, all criteria are combined with AND logic:

```go
controller.CreateTrafficClass("WebAPI").
    ForPort(443).                     // AND port 443
    ForSource("10.0.1.0/24").        // AND from subnet
    ForProtocol("tcp")               // AND TCP protocol
```

### Q: How do I update configuration without downtime?
A: Create a new controller and apply atomically:

```go
// Current running configuration
oldController := getCurrentController()

// New configuration
newController := api.NetworkInterface("eth0")
newController.WithHardLimitBandwidth("2gbps") // Increased

// Apply new configuration
if err := newController.Apply(); err != nil {
    // Rollback to old on failure
    oldController.Apply()
    return err
}
```

### Q: What's the difference between the fluent API and structured configuration?
A: Two different patterns for the same functionality:

```go
// Fluent API - more readable in code
controller.CreateTrafficClass("Web").
    WithGuaranteedBandwidth("100mbps").
    WithPriority(1).
    ForPort(80, 443)

// Structured configuration - better for config files
config := api.Config{
    Device: "eth0",
    Classes: []api.ClassConfig{
        {
            Name:       "Web",
            Guaranteed: "100mbps",
            Priority:   1,
            Ports:      []int{80, 443},
        },
    },
}
```

## Performance Questions

### Q: How much overhead does the library add?
A: Minimal. The library:
- Uses efficient netlink communication
- Caches configuration state
- Only makes kernel calls when necessary

Most overhead comes from the kernel TC subsystem itself, not the library.

### Q: How often should I check statistics?
A: Depends on your needs:
- **Real-time monitoring**: 1-5 seconds
- **Alerting**: 10-30 seconds  
- **Trending**: 1-5 minutes

```go
// Example monitoring
ticker := time.NewTicker(10 * time.Second)
for range ticker.C {
    stats := controller.GetStatistics()
    checkForProblems(stats)
}
```

### Q: Can I apply multiple configurations simultaneously?
A: No, each network interface can only have one traffic control configuration. However, you can:

```go
// Multiple interfaces
eth0 := api.NetworkInterface("eth0")
eth1 := api.NetworkInterface("eth1")

// Apply different configs to each
eth0.Apply() // Config A
eth1.Apply() // Config B
```

## Troubleshooting Questions

### Q: Why am I getting "permission denied" errors?
A: Traffic control requires `CAP_NET_ADMIN` capability:

```bash
# Check if you have the capability
sudo capsh --print | grep net_admin

# Run with sudo
sudo your-program

# Or add capability to binary (advanced)
sudo setcap cap_net_admin+ep your-program
```

### Q: Why aren't my bandwidth limits working?
A: Common causes:

1. **No matching filters**: Traffic isn't matching your class
```go
// Add a default class to catch unmatched traffic
controller.CreateTrafficClass("Default").WithPriority(7)
```

2. **Soft limit too high**: Class can borrow more than expected
```go
// Reduce soft limit
controller.CreateTrafficClass("Limited").
    WithGuaranteedBandwidth("10mbps").
    WithSoftLimitBandwidth("15mbps") // Was 100mbps
```

3. **Interface not actually limited**: Verify with tools
```bash
# Check applied configuration
tc qdisc show dev eth0
tc class show dev eth0
```

### Q: Why are packets being dropped?
A: Several possible reasons:

1. **Oversubscribed bandwidth**: Sum of guaranteed > total
2. **Buffer overflow**: Increase buffer sizes or reduce rates
3. **Bursty traffic**: Adjust burst parameters

```go
// Check statistics for drop patterns
stats := controller.GetStatistics()
for _, class := range stats.Classes {
    if class.PacketsDropped > 0 {
        log.Printf("Class %s: %d packets dropped", 
            class.Name, class.PacketsDropped)
    }
}
```

### Q: How do I debug filter matching?
A: Use TC commands to inspect:

```bash
# Show all filters
tc filter show dev eth0

# Show statistics with verbose output
tc -s class show dev eth0

# Monitor packets in real-time
watch -n 1 'tc -s class show dev eth0'
```

### Q: Configuration applies but doesn't work - why?
A: Verify each step:

```go
// 1. Check interface exists
if !interfaceExists("eth0") {
    log.Fatal("Interface eth0 not found")
}

// 2. Apply configuration
if err := controller.Apply(); err != nil {
    log.Fatal("Apply failed:", err)
}

// 3. Verify configuration applied
time.Sleep(1 * time.Second)
stats := controller.GetStatistics()
if len(stats.Classes) == 0 {
    log.Fatal("No classes found after apply")
}

// 4. Generate test traffic and monitor
```

## Error Messages

### Q: "RTNETLINK answers: Operation not permitted"
A: Missing `CAP_NET_ADMIN` capability. Run with `sudo` or add the capability.

### Q: "RTNETLINK answers: No such device"
A: Interface name is wrong or interface doesn't exist.
```bash
# List available interfaces
ip link show
```

### Q: "RTNETLINK answers: File exists"
A: Trying to create something that already exists. Clear existing rules:
```bash
tc qdisc del dev eth0 root
```

### Q: "Cannot find device 'eth0'"
A: Interface name is incorrect. Check with:
```bash
ip addr show
```

## Integration Questions

### Q: How do I integrate with Prometheus monitoring?
A: Create custom metrics:

```go
import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
    bytesTransmitted *prometheus.CounterVec
    packetsDropped   *prometheus.CounterVec
}

func (m *Metrics) UpdateFromStats(stats *api.Statistics) {
    for _, class := range stats.Classes {
        m.bytesTransmitted.WithLabelValues(stats.Device, class.Name).
            Add(float64(class.BytesSent))
        m.packetsDropped.WithLabelValues(stats.Device, class.Name).
            Add(float64(class.PacketsDropped))
    }
}
```

### Q: Can I use this with Docker networks?
A: Yes, but apply traffic control to the host interface, not container interfaces:

```go
// Shape traffic on the host interface where containers connect
controller := api.NetworkInterface("docker0")
// or the actual physical interface
controller := api.NetworkInterface("eth0")
```

### Q: How do I handle configuration in Kubernetes?
A: Use a DaemonSet with elevated privileges:

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: traffic-controller
spec:
  template:
    spec:
      containers:
      - name: traffic-controller
        image: your-image
        securityContext:
          capabilities:
            add: ["NET_ADMIN"]
        volumeMounts:
        - name: config
          mountPath: /config
      volumes:
      - name: config
        configMap:
          name: traffic-config
```

## Still Need Help?

- 📖 Check the [API Usage Guide](api-usage-guide.md)
- 🏭 Review [Best Practices](best-practices.md)
- 🔧 Browse [Examples](../examples/)
- 🐛 Report issues on [GitHub](https://github.com/rng999/traffic-control-go/issues)