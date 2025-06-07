# Traffic Control Go - Quick Start Guide

Get up and running with Traffic Control Go in 5 minutes.

## Installation

```bash
go get github.com/rng999/traffic-control-go
```

## Basic Example

```go
package main

import (
    "log"
    "github.com/rng999/traffic-control-go/api"
)

func main() {
    // Create controller for your network interface
    controller := api.NetworkInterface("eth0")
    
    // Set total bandwidth
    controller.WithHardLimitBandwidth("1gbps")
    
    // Create traffic class for web services
    controller.CreateTrafficClass("Web").
        WithGuaranteedBandwidth("100mbps").
        WithSoftLimitBandwidth("200mbps").
        WithPriority(1).
        ForPort(80, 443)
    
    // Apply configuration
    if err := controller.Apply(); err != nil {
        log.Fatal(err)
    }
}
```

## Key Concepts in 30 Seconds

### 1. Bandwidth Types
- **Hard Limit**: Total interface capacity (`WithHardLimitBandwidth`)
- **Guaranteed**: Minimum assured bandwidth (`WithGuaranteedBandwidth`)
- **Soft Limit**: Maximum with borrowing (`WithSoftLimitBandwidth`)

### 2. Priority System
```go
WithPriority(0)  // Highest (Handle 1:10)
WithPriority(3)  // Normal  (Handle 1:13)
WithPriority(7)  // Lowest  (Handle 1:17)
```

### 3. Traffic Matching
```go
ForPort(80, 443)                    // Match by port
ForSource("192.168.1.0/24")        // Match by source IP
ForDestination("10.0.0.100")       // Match by destination
ForProtocol("tcp")                 // Match by protocol
```

## Common Patterns

### Home Network QoS
```go
controller := api.NetworkInterface("eth0")
controller.WithHardLimitBandwidth("100mbps")

// High priority for video calls
controller.CreateTrafficClass("VideoCall").
    WithGuaranteedBandwidth("30mbps").
    WithPriority(0).
    ForPort(3478, 3479) // Common video call ports

// Normal priority for web browsing
controller.CreateTrafficClass("Web").
    WithGuaranteedBandwidth("40mbps").
    WithPriority(3).
    ForPort(80, 443)

// Low priority for downloads
controller.CreateTrafficClass("Downloads").
    WithGuaranteedBandwidth("20mbps").
    WithPriority(7).
    ForPort(8080) // Download manager port

controller.Apply()
```

### Server Traffic Shaping
```go
controller := api.NetworkInterface("eth0")
controller.WithHardLimitBandwidth("10gbps")

// Database replication (critical)
controller.CreateTrafficClass("Database").
    WithGuaranteedBandwidth("2gbps").
    WithSoftLimitBandwidth("3gbps").
    WithPriority(0).
    ForPort(3306).
    ForSource("10.0.1.0/24") // DB subnet

// API traffic (high priority)
controller.CreateTrafficClass("API").
    WithGuaranteedBandwidth("3gbps").
    WithSoftLimitBandwidth("5gbps").
    WithPriority(1).
    ForPort(8443)

// Static content (normal priority)
controller.CreateTrafficClass("Static").
    WithGuaranteedBandwidth("1gbps").
    WithSoftLimitBandwidth("4gbps").
    WithPriority(3).
    ForPort(443)

controller.Apply()
```

## Configuration Files

### YAML Configuration
```yaml
# traffic-config.yaml
version: "1.0"
device: eth0
bandwidth: 1Gbps
classes:
  - name: critical
    guaranteed: 400Mbps
    max: 600Mbps
    priority: 0
    ports: [22, 3306]
  
  - name: standard
    guaranteed: 300Mbps
    max: 500Mbps
    priority: 3
    ports: [80, 443]
```

### Loading Configuration
```go
config, err := api.LoadConfigFromYAML("traffic-config.yaml")
if err != nil {
    log.Fatal(err)
}

result := api.NewTrafficControlService(config)
if result.IsFailure() {
    log.Fatal(result.Error())
}

service := result.Value()
```

## Monitoring

```go
// Get current statistics
stats, err := controller.GetStatistics()
if err != nil {
    log.Fatal(err)
}

// Check each class
for _, class := range stats.Classes {
    fmt.Printf("Class %s: %d bytes sent, %d packets dropped\n",
        class.Name, class.BytesSent, class.PacketsDropped)
}
```

## Best Practices Checklist

✅ **Always set a default class** for unmatched traffic
```go
controller.CreateTrafficClass("Default").
    WithGuaranteedBandwidth("10mbps").
    WithPriority(7)
```

✅ **Leave bandwidth headroom** (~10% unallocated)
```go
// If interface is 1Gbps, allocate only 900Mbps total
```

✅ **Validate configuration** before applying
```go
if totalGuaranteed > hardLimit {
    log.Fatal("Oversubscribed bandwidth")
}
```

✅ **Monitor packet drops** after applying
```go
stats := controller.GetStatistics()
for _, class := range stats.Classes {
    if class.PacketsDropped > 0 {
        log.Printf("Warning: %s dropping packets", class.Name)
    }
}
```

## Troubleshooting

### Permission Denied
```bash
# Run with sudo or add CAP_NET_ADMIN capability
sudo ./your-program
```

### Changes Not Taking Effect
```go
// Clear existing rules first
// #nosec G204 - device name is hardcoded for example purposes
exec.Command("tc", "qdisc", "del", "dev", "eth0", "root").Run()
```

### High Packet Drops
```go
// Increase buffer sizes or reduce guaranteed bandwidth
controller.CreateTrafficClass("Problem").
    WithGuaranteedBandwidth("50mbps"). // Reduce from 100mbps
    WithSoftLimitBandwidth("100mbps")
```

## Next Steps

- 📖 Read the [Comprehensive API Guide](api-usage-guide.md)
- 🏭 Review [Production Best Practices](best-practices.md)
- 🔧 Explore [Advanced Examples](../examples/)
- 📊 Learn about [Traffic Monitoring](logging.md)

## Need Help?

- Check the [FAQ](faq.md)
- Browse [GitHub Issues](https://github.com/rng999/traffic-control-go/issues)
- Read the [Architecture Overview](../memory-bank/architecture-overview.md)