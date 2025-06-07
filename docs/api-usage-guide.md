# Traffic Control Go - Comprehensive API Usage Guide

This guide provides detailed examples, best practices, and advanced usage patterns for the Traffic Control Go library.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Core Concepts](#core-concepts)
3. [API Patterns](#api-patterns)
4. [Common Use Cases](#common-use-cases)
5. [Best Practices](#best-practices)
6. [Advanced Features](#advanced-features)
7. [Error Handling](#error-handling)
8. [Performance Tips](#performance-tips)
9. [Troubleshooting](#troubleshooting)

## Getting Started

### Installation

```bash
go get github.com/rng999/traffic-control-go
```

### Basic Setup

```go
package main

import (
    "log"
    "github.com/rng999/traffic-control-go/api"
)

func main() {
    // Create a controller for your network interface
    controller := api.NetworkInterface("eth0")
    
    // Set the total available bandwidth
    controller.WithHardLimitBandwidth("1gbps")
    
    // Apply basic configuration
    if err := controller.Apply(); err != nil {
        log.Fatal(err)
    }
}
```

## Core Concepts

### Traffic Control Hierarchy

```
Physical Interface (eth0)
    └── Root Qdisc (HTB)
        ├── Class 1 (Web Traffic)
        │   └── Filters (Port 80, 443)
        ├── Class 2 (Database)
        │   └── Filters (Port 3306, 5432)
        └── Class 3 (Default)
```

### Bandwidth Types Explained

```go
// 1. Hard Limit Bandwidth - Total interface capacity
controller.WithHardLimitBandwidth("1gbps")

// 2. Guaranteed Bandwidth - Minimum assured rate
class.WithGuaranteedBandwidth("100mbps")

// 3. Soft Limit Bandwidth - Maximum with borrowing
class.WithSoftLimitBandwidth("500mbps")
```

**Key Relationships:**
- Hard Limit ≥ Sum of all Soft Limits
- Soft Limit ≥ Guaranteed Bandwidth
- Guaranteed bandwidth is always available
- Soft limit allows borrowing unused bandwidth

### Priority System

```go
// Priority values (0-7) automatically map to handles
WithPriority(0)  // Highest priority → Handle 1:10
WithPriority(3)  // Normal priority  → Handle 1:13
WithPriority(7)  // Lowest priority  → Handle 1:17
```

## API Patterns

### Pattern 1: Fluent Builder Pattern

```go
controller := api.NetworkInterface("eth0").
    WithHardLimitBandwidth("1gbps")

controller.CreateTrafficClass("WebServices").
    WithGuaranteedBandwidth("200mbps").
    WithSoftLimitBandwidth("400mbps").
    WithPriority(1).
    ForPort(80, 443).
    ForProtocol("tcp")
```

### Pattern 2: Structured Configuration

```go
// Using configuration objects
config := api.Config{
    Device: "eth0",
    Bandwidth: "1gbps",
    Classes: []api.ClassConfig{
        {
            Name:       "critical",
            Guaranteed: "400mbps",
            Max:        "600mbps",
            Priority:   0,
        },
    },
}

service, err := api.NewTrafficControlService(config)
```

### Pattern 3: Error Handling with Result Type

```go
import "github.com/rng999/traffic-control-go/pkg/types"

// Functions return Result type for better error handling
result := api.NewTrafficControlService(config)
if result.IsFailure() {
    log.Fatal(result.Error())
}

service := result.Value()
```

## Common Use Cases

### Use Case 1: Web Server Traffic Shaping

```go
func setupWebServerTrafficControl() error {
    controller := api.NetworkInterface("eth0")
    controller.WithHardLimitBandwidth("10gbps")
    
    // Critical API traffic
    controller.CreateTrafficClass("API").
        WithGuaranteedBandwidth("2gbps").
        WithSoftLimitBandwidth("4gbps").
        WithPriority(0).
        ForPort(8080, 8443).
        ForDestination("10.0.1.0/24")
    
    // Static content delivery
    controller.CreateTrafficClass("Static").
        WithGuaranteedBandwidth("1gbps").
        WithSoftLimitBandwidth("3gbps").
        WithPriority(2).
        ForPort(80, 443).
        ForProtocol("tcp")
    
    // Database connections
    controller.CreateTrafficClass("Database").
        WithGuaranteedBandwidth("500mbps").
        WithSoftLimitBandwidth("1gbps").
        WithPriority(1).
        ForPort(3306, 5432)
    
    // Background tasks
    controller.CreateTrafficClass("Background").
        WithGuaranteedBandwidth("100mbps").
        WithSoftLimitBandwidth("500mbps").
        WithPriority(7).
        ForPort(9000) // Job queue port
    
    return controller.Apply()
}
```

### Use Case 2: Multi-Tenant Network Isolation

```go
func setupMultiTenantTrafficControl(tenants []Tenant) error {
    controller := api.NetworkInterface("eth0")
    controller.WithHardLimitBandwidth("100gbps")
    
    for i, tenant := range tenants {
        // Each tenant gets guaranteed bandwidth based on their tier
        guaranteed := calculateTenantBandwidth(tenant.Tier)
        maxBandwidth := guaranteed * 2 // Allow 2x burst
        
        controller.CreateTrafficClass(tenant.Name).
            WithGuaranteedBandwidth(guaranteed).
            WithSoftLimitBandwidth(maxBandwidth).
            WithPriority(uint8(i % 8)). // Distribute priorities
            ForSource(tenant.IPRange)
    }
    
    // Default class for unmatched traffic
    controller.CreateTrafficClass("Default").
        WithGuaranteedBandwidth("100mbps").
        WithSoftLimitBandwidth("1gbps").
        WithPriority(7)
    
    return controller.Apply()
}
```

### Use Case 3: QoS for Video Streaming

```go
func setupVideoStreamingQoS() error {
    controller := api.NetworkInterface("eth0")
    controller.WithHardLimitBandwidth("10gbps")
    
    // Live streaming (highest priority)
    controller.CreateTrafficClass("LiveStream").
        WithGuaranteedBandwidth("4gbps").
        WithSoftLimitBandwidth("6gbps").
        WithPriority(0).
        ForPort(1935, 8080). // RTMP, HLS
        ForProtocol("tcp", "udp")
    
    // VOD content
    controller.CreateTrafficClass("VOD").
        WithGuaranteedBandwidth("3gbps").
        WithSoftLimitBandwidth("5gbps").
        WithPriority(2).
        ForPort(443).
        ForDestination("cdn.example.com")
    
    // Control traffic (API, metrics)
    controller.CreateTrafficClass("Control").
        WithGuaranteedBandwidth("500mbps").
        WithSoftLimitBandwidth("1gbps").
        WithPriority(1).
        ForPort(8443, 9090)
    
    return controller.Apply()
}
```

### Use Case 4: Dynamic Traffic Control

```go
func dynamicTrafficControl(ctx context.Context) error {
    controller := api.NetworkInterface("eth0")
    controller.WithHardLimitBandwidth("10gbps")
    
    // Monitor and adjust based on conditions
    go func() {
        ticker := time.NewTicker(5 * time.Minute)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                stats := controller.GetStatistics()
                adjustTrafficClasses(controller, stats)
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return controller.Apply()
}

func adjustTrafficClasses(controller *api.TrafficController, stats *api.Statistics) {
    // Adjust bandwidth based on current usage
    if stats.Utilization > 0.9 {
        // Reduce non-critical traffic
        controller.UpdateTrafficClass("Background").
            WithSoftLimitBandwidth("100mbps")
    }
}
```

## Best Practices

### 1. Bandwidth Planning

```go
// Always leave headroom for system traffic
totalBandwidth := "1gbps"
allocatedBandwidth := "900mbps" // Keep 10% for system

// Ensure guaranteed bandwidth sum doesn't exceed total
guaranteedSum := 0
for _, class := range classes {
    guaranteedSum += class.Guaranteed
}
if guaranteedSum > totalBandwidth {
    return errors.New("oversubscribed guaranteed bandwidth")
}
```

### 2. Priority Assignment

```go
// Use constants for priority levels
const (
    PriorityCritical   = 0
    PriorityHigh       = 1
    PriorityNormal     = 3
    PriorityLow        = 5
    PriorityBackground = 7
)

// Assign priorities based on business importance
controller.CreateTrafficClass("PaymentAPI").
    WithPriority(PriorityCritical)
    
controller.CreateTrafficClass("UserAPI").
    WithPriority(PriorityHigh)
    
controller.CreateTrafficClass("Analytics").
    WithPriority(PriorityBackground)
```

### 3. Filter Specificity

```go
// More specific filters first
controller.CreateTrafficClass("CriticalHost").
    WithPriority(0).
    ForSource("10.0.1.100").     // Specific host
    ForPort(443)

controller.CreateTrafficClass("WebTraffic").
    WithPriority(3).
    ForPort(443)                  // General HTTPS

// Order matters - specific rules should be higher priority
```

### 4. Configuration Validation

```go
func validateTrafficConfig(controller *api.TrafficController) error {
    config := controller.GetConfiguration()
    
    // Check bandwidth relationships
    for _, class := range config.Classes {
        if class.Guaranteed > class.SoftLimit {
            return fmt.Errorf("class %s: guaranteed > soft limit", class.Name)
        }
    }
    
    // Check priority uniqueness
    priorities := make(map[uint8]string)
    for _, class := range config.Classes {
        if existing, ok := priorities[class.Priority]; ok {
            return fmt.Errorf("duplicate priority %d: %s and %s", 
                class.Priority, existing, class.Name)
        }
        priorities[class.Priority] = class.Name
    }
    
    return nil
}
```

## Advanced Features

### 1. Multiple Qdisc Types

```go
// Token Bucket Filter for rate limiting
func setupRateLimiting() error {
    deviceName, _ := tc.NewDeviceName("eth0")
    config := api.Config{Device: deviceName}
    service := api.NewTrafficControlService(config).Value()
    
    // Create TBF qdisc for simple rate limiting
    handle := tc.NewHandle(1, 0)
    rate := tc.Mbps(100)
    buffer := uint32(1600)
    limit := uint32(3000)
    
    return service.CreateTBFQdisc(handle, rate, buffer, limit)
}

// Priority Queueing
func setupPriorityQueueing() error {
    service := createService("eth0")
    
    handle := tc.NewHandle(1, 0)
    bands := uint8(3)
    priomap := []uint8{1, 2, 2, 2, 1, 2, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1}
    
    return service.CreatePRIOQdisc(handle, bands, priomap)
}

// Fair Queueing with Controlled Delay
func setupFairQueueing() error {
    service := createService("eth0")
    
    handle := tc.NewHandle(1, 0)
    limit := uint32(10000)
    flows := uint32(1024)
    target := uint32(5000)    // 5ms
    interval := uint32(100000) // 100ms
    quantum := uint32(1514)
    ecn := true
    
    return service.CreateFQCODELQdisc(handle, limit, flows, 
        target, interval, quantum, ecn)
}
```

### 2. Statistics Collection

```go
func monitorTrafficStatistics(ctx context.Context, service *api.TrafficControlService) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            stats, err := service.GetStatistics()
            if err != nil {
                log.Printf("Failed to get statistics: %v", err)
                continue
            }
            
            // Process statistics
            for _, classStats := range stats.Classes {
                log.Printf("Class %s: TX=%d bytes, Dropped=%d packets",
                    classStats.Name,
                    classStats.BytesSent,
                    classStats.PacketsDropped)
                
                // Alert on high drop rate
                if classStats.DropRate > 0.01 { // 1% drop rate
                    alertHighDropRate(classStats)
                }
            }
            
        case <-ctx.Done():
            return
        }
    }
}
```

### 3. Event-Driven Updates

```go
// Listen for configuration changes
func setupEventHandlers(service *api.TrafficControlService) {
    service.OnClassCreated(func(event ClassCreatedEvent) {
        log.Printf("Class created: %s with bandwidth %s",
            event.ClassName, event.Bandwidth)
    })
    
    service.OnClassModified(func(event ClassModifiedEvent) {
        log.Printf("Class modified: %s", event.ClassName)
    })
    
    service.OnThresholdExceeded(func(event ThresholdEvent) {
        log.Printf("Threshold exceeded for %s: %v", 
            event.ClassName, event.Metric)
        // Take corrective action
    })
}
```

### 4. Configuration Persistence

```go
// Save configuration for disaster recovery
func saveConfiguration(controller *api.TrafficController) error {
    config := controller.ExportConfiguration()
    
    data, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile("tc-config-backup.json", data, 0644)
}

// Restore configuration
func restoreConfiguration(filename string) error {
    data, err := os.ReadFile(filename)
    if err != nil {
        return err
    }
    
    var config api.Configuration
    if err := json.Unmarshal(data, &config); err != nil {
        return err
    }
    
    controller := api.NetworkInterface(config.Device)
    return controller.ImportConfiguration(config)
}
```

## Error Handling

### Using Result Types

```go
import "github.com/rng999/traffic-control-go/pkg/types"

func setupTrafficControl() types.Result[*api.TrafficController] {
    // Parse bandwidth
    bandwidthResult := tc.ParseBandwidth("1gbps")
    if bandwidthResult.IsFailure() {
        return types.Failure[*api.TrafficController](
            fmt.Errorf("invalid bandwidth: %w", bandwidthResult.Error()))
    }
    
    controller := api.NetworkInterface("eth0")
    controller.WithHardLimitBandwidth(bandwidthResult.Value().String())
    
    // Chain operations with Result type
    return types.Success(controller)
}

// Usage
result := setupTrafficControl()
result.
    Map(func(controller *api.TrafficController) *api.TrafficController {
        controller.CreateTrafficClass("Web").
            WithGuaranteedBandwidth("100mbps")
        return controller
    }).
    MapError(func(err error) error {
        return fmt.Errorf("traffic control setup failed: %w", err)
    })
```

### Common Error Scenarios

```go
func handleCommonErrors(err error) {
    switch {
    case errors.Is(err, api.ErrInvalidBandwidth):
        log.Println("Check bandwidth format (e.g., '100mbps', '1gbps')")
        
    case errors.Is(err, api.ErrInsufficientPrivileges):
        log.Println("Run with CAP_NET_ADMIN or as root")
        
    case errors.Is(err, api.ErrInterfaceNotFound):
        log.Println("Check interface name with 'ip link show'")
        
    case errors.Is(err, api.ErrBandwidthOversubscribed):
        log.Println("Total guaranteed bandwidth exceeds interface capacity")
        
    default:
        log.Printf("Unexpected error: %v", err)
    }
}
```

## Performance Tips

### 1. Batch Operations

```go
// Instead of applying after each change
for _, class := range classes {
    controller.CreateTrafficClass(class.Name).
        WithGuaranteedBandwidth(class.Bandwidth)
    controller.Apply() // Don't do this
}

// Apply once after all changes
for _, class := range classes {
    controller.CreateTrafficClass(class.Name).
        WithGuaranteedBandwidth(class.Bandwidth)
}
controller.Apply() // Do this once
```

### 2. Efficient Filtering

```go
// Use specific filters to reduce processing
controller.CreateTrafficClass("Database").
    ForSource("10.0.1.0/24").      // Subnet filter
    ForDestination("10.0.2.100").   // Specific host
    ForPort(3306).                  // Specific port
    ForProtocol("tcp")              // Protocol filter

// Avoid overly broad filters
controller.CreateTrafficClass("All").
    ForProtocol("ip") // Matches everything - inefficient
```

### 3. Statistics Caching

```go
type CachedStatistics struct {
    mu         sync.RWMutex
    stats      *api.Statistics
    lastUpdate time.Time
    ttl        time.Duration
}

func (c *CachedStatistics) GetStatistics(service *api.TrafficControlService) (*api.Statistics, error) {
    c.mu.RLock()
    if time.Since(c.lastUpdate) < c.ttl {
        defer c.mu.RUnlock()
        return c.stats, nil
    }
    c.mu.RUnlock()
    
    c.mu.Lock()
    defer c.mu.Unlock()
    
    stats, err := service.GetStatistics()
    if err != nil {
        return nil, err
    }
    
    c.stats = stats
    c.lastUpdate = time.Now()
    return stats, nil
}
```

## Troubleshooting

### 1. Verification Commands

```go
func verifyConfiguration(device string) error {
    // Check if configuration was applied
    cmd := exec.Command("tc", "qdisc", "show", "dev", device)
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("failed to verify: %w", err)
    }
    
    log.Printf("Current configuration:\n%s", output)
    return nil
}
```

### 2. Debug Logging

```go
import "github.com/rng999/traffic-control-go/pkg/logging"

// Enable debug logging
logging.Initialize(&logging.Config{
    Level:       "debug",
    Format:      "console",
    Development: true,
})

// Add context to logs
logger := logging.WithComponent("traffic-setup").
    WithDevice("eth0").
    WithOperation("apply-config")

logger.Debug("Applying configuration",
    logging.String("bandwidth", "1gbps"),
    logging.Int("class_count", 5))
```

### 3. Common Issues and Solutions

```go
// Issue: Changes not taking effect
func troubleshootChanges(device string) {
    // 1. Check for existing qdiscs
    existing := checkExistingQdiscs(device)
    if len(existing) > 0 {
        log.Println("Existing qdiscs found, clearing...")
        clearExistingQdiscs(device)
    }
    
    // 2. Verify interface is up
    if !isInterfaceUp(device) {
        log.Printf("Interface %s is down", device)
        enableInterface(device)
    }
    
    // 3. Check for permission issues
    if !hasNetAdminCapability() {
        log.Println("Missing CAP_NET_ADMIN capability")
    }
}

// Issue: Bandwidth not being limited
func troubleshootBandwidth(class string) {
    // Check actual vs configured bandwidth
    stats := getClassStatistics(class)
    config := getClassConfiguration(class)
    
    if stats.Rate > config.SoftLimit {
        log.Printf("Class %s exceeding soft limit: %v > %v",
            class, stats.Rate, config.SoftLimit)
        
        // Check for filter issues
        checkFilters(class)
    }
}
```

## Summary

The Traffic Control Go library provides a powerful, human-readable API for managing Linux traffic control. Key takeaways:

1. **Use the fluent API** for readable, maintainable code
2. **Understand bandwidth relationships** to avoid oversubscription
3. **Leverage priorities** for clear traffic hierarchy
4. **Monitor statistics** to ensure configuration effectiveness
5. **Handle errors gracefully** using the Result type pattern
6. **Follow best practices** for production deployments

For more examples, see the [examples directory](../examples/) in the repository.