# API Guide - Human-Readable Traffic Control

This document describes the current API design for traffic-control-go that provides clear, human-readable method names and bandwidth concepts.

## Overview

The API uses descriptive method names and clear bandwidth concepts to make traffic control configuration intuitive for developers.

## Current API Design

### Bandwidth Concepts
- **Hard Limit**: Physical bandwidth limit (cannot be exceeded)
- **Soft Limit**: Policy bandwidth limit (can be exceeded through borrowing)
- **Guaranteed**: Minimum bandwidth reservation for a class

### API Example
```go
// Create controller for network interface
tc := api.NetworkInterface("eth0").
    WithHardLimitBandwidth("1Gbps")  // Physical interface limit

// Configure web services class
tc.CreateTrafficClass("Web Services").
    WithGuaranteedBandwidth("300Mbps").     // Minimum guarantee
    WithSoftLimitBandwidth("500Mbps").      // Policy limit (borrowing allowed)
    WithPriority(4).                        // Normal priority
    ForPort(80, 443).                       // HTTP/HTTPS traffic
    AddClass()                              // Complete class configuration

// Configure database class
tc.CreateTrafficClass("Database").
    WithGuaranteedBandwidth("200Mbps").
    WithSoftLimitBandwidth("400Mbps").
    WithPriority(2).                        // High priority
    ForPort(3306, 5432).                    // MySQL, PostgreSQL
    AddClass()

// Apply configuration
tc.Apply()
```

## Key Features

### 1. Human-Readable Method Names
- `NetworkInterface()` - Clear interface selection
- `WithHardLimitBandwidth()` - Physical interface limit
- `WithSoftLimitBandwidth()` - Policy-based limit with borrowing
- `WithGuaranteedBandwidth()` - Minimum bandwidth guarantee
- `AddClass()` - Complete class configuration

### 2. Clear Bandwidth Concepts
- **Hard Limit**: Absolute physical constraint (line rate, interface speed)
- **Soft Limit**: Policy constraint that allows borrowing from other classes
- **Guaranteed**: Minimum reserved bandwidth for each class

### 3. Natural Configuration Flow
- Configure the controller: `api.NetworkInterface("eth0").WithHardLimitBandwidth("1Gbps")`
- Configure each class: `tc.CreateTrafficClass("name").WithGuaranteedBandwidth().WithPriority().AddClass()`
- Apply configuration: `tc.Apply()`

### 4. Enhanced Filtering Options
- `ForPort(80, 443, 8080)` - Multiple ports in one call
- `ForSourceIPs("192.168.1.0/24", "10.0.1.100")` - Source IP filtering
- `ForDestinationIPs("10.0.2.100", "10.0.2.101")` - Destination IP filtering
- `ForProtocols("ssh", "http", "https")` - Protocol-based filtering

## Complete API Reference

### Controller Methods

#### `NetworkInterface(deviceName string) *TrafficController`
Creates a new traffic controller for the specified network device.

```go
tc := api.NetworkInterface("eth0")
```

#### `WithHardLimitBandwidth(bandwidth string) *TrafficController`
Sets the total physical bandwidth available on the interface (hard limit).

```go
tc.WithHardLimitBandwidth("1Gbps")   // 1 Gigabit per second
tc.WithHardLimitBandwidth("100Mbps") // 100 Megabits per second  
tc.WithHardLimitBandwidth("500mbit") // 500 megabits per second (alternative format)
```

#### `CreateTrafficClass(name string) *TrafficClassBuilder`
Creates a new traffic class for configuration.

```go
webClass := tc.CreateTrafficClass("Web Services")
dbClass := tc.CreateTrafficClass("Database")
```

#### `Apply() error`
Applies the configuration to the system. Validates all settings before application.

```go
if err := tc.Apply(); err != nil {
    log.Fatal(err)
}
```

#### `String() string`
Returns a human-readable representation of the configuration.

```go
fmt.Println(tc.String())
```

### Class Builder Methods

#### `WithGuaranteedBandwidth(bandwidth string) *TrafficClassBuilder`
Sets the guaranteed minimum bandwidth for the class.

```go
builder.WithGuaranteedBandwidth("100Mbps")
builder.WithGuaranteedBandwidth("50mbit")
```

#### `WithSoftLimitBandwidth(bandwidth string) *TrafficClassBuilder`
Sets the soft limit bandwidth for the class (borrowing allowed).

```go
builder.WithSoftLimitBandwidth("200Mbps")
```

#### `WithPriority(priority int) *TrafficClassBuilder`
Sets the priority for the class (0-7, where 0 is highest priority).

```go
builder.WithPriority(0)  // Highest priority
builder.WithPriority(4)  // Normal priority
builder.WithPriority(7)  // Lowest priority
```

#### `ForPort(ports ...int) *TrafficClassBuilder`
Adds port-based filtering to the class.

```go
builder.ForPort(80, 443)           // HTTP and HTTPS
builder.ForPort(22)                // SSH
builder.ForPort(3306, 5432, 1521)  // Database ports
```

#### `ForSourceIPs(ips ...string) *TrafficClassBuilder`
Adds source IP-based filtering to the class.

```go
builder.ForSourceIPs("192.168.1.0/24")                    // Subnet
builder.ForSourceIPs("10.0.1.100", "10.0.1.101")         // Specific IPs
builder.ForSourceIPs("192.168.100.0/24", "172.16.0.0/16") // Multiple subnets
```

#### `ForDestinationIPs(ips ...string) *TrafficClassBuilder`
Adds destination IP-based filtering to the class.

```go
builder.ForDestinationIPs("192.168.1.10")        // Single server
builder.ForDestinationIPs("10.0.2.0/24")         // Server subnet
```

#### `ForProtocols(protocols ...string) *TrafficClassBuilder`
Adds protocol-based filtering to the class.

```go
builder.ForProtocols("tcp", "udp")
builder.ForProtocols("ssh", "http", "https")
builder.ForProtocols("dns", "ntp")
```

#### `AddClass() *TrafficController`
Completes the class configuration and adds it to the controller.

```go
builder.AddClass()  // Finalizes and adds the class
```

## Usage Examples

### Basic Web Server Configuration
```go
tc := api.NetworkInterface("eth0").
    WithHardLimitBandwidth("1Gbps")

tc.CreateTrafficClass("Web Traffic").
    WithGuaranteedBandwidth("400Mbps").
    WithSoftLimitBandwidth("600Mbps").
    WithPriority(3).
    ForPort(80, 443, 8080, 8443).
    AddClass()

tc.CreateTrafficClass("SSH Management").
    WithGuaranteedBandwidth("10Mbps").
    WithSoftLimitBandwidth("50Mbps").
    WithPriority(1).
    ForPort(22).
    AddClass()

tc.Apply()
```

### Complex Server Infrastructure
```go
tc := api.NewImproved("bond0").
    TotalBandwidth("40Gbps")

// Web tier
tc.Class("Web Tier").
    Guaranteed("15Gbps").
    BurstTo("25Gbps").
    Priority(2).
    Ports(80, 443).
    SourceIPs("10.0.1.0/24")

// Application tier
tc.Class("App Tier").
    Guaranteed("10Gbps").
    BurstTo("20Gbps").
    Priority(1).
    Ports(8080, 8443, 9000).
    SourceIPs("10.0.2.0/24", "10.0.3.0/24")

// Database tier
tc.Class("DB Tier").
    Guaranteed("8Gbps").
    BurstTo("15Gbps").
    Priority(0).
    Ports(3306, 5432, 27017, 6379).
    SourceIPs("10.0.4.0/24")

// Management
tc.Class("Management").
    Guaranteed("2Gbps").
    BurstTo("5Gbps").
    Priority(3).
    Ports(22, 161, 162).
    SourceIPs("10.0.100.0/24")

tc.Apply()
```

### Class Reuse and Incremental Configuration
```go
tc := api.NewImproved("eth0").
    TotalBandwidth("1Gbps")

// Initial configuration
tc.Class("Web Services").
    Guaranteed("300Mbps").
    Priority(4).
    Ports(80, 443)

// Later, add more ports to the same class
tc.Class("Web Services").
    Ports(8080, 8443)  // Adds to existing ports

// Add SSL/TLS specific traffic
tc.Class("Web Services").
    Protocols("https", "tls")

tc.Apply()
```

### Error Handling
```go
tc := api.NewImproved("eth0").
    TotalBandwidth("1Gbps")

tc.Class("Test").
    Guaranteed("500Mbps").
    Priority(4)

if err := tc.Apply(); err != nil {
    switch {
    case strings.Contains(err.Error(), "total bandwidth"):
        log.Fatal("Total bandwidth not set")
    case strings.Contains(err.Error(), "guaranteed bandwidth"):
        log.Fatal("Class missing guaranteed bandwidth")
    case strings.Contains(err.Error(), "priority"):
        log.Fatal("Class missing priority")
    default:
        log.Fatalf("Configuration error: %v", err)
    }
}
```

## Migration from v0.1.0 API

To migrate from the old API to the improved API:

1. **Change the constructor**: `api.New()` → `api.NewImproved()`
2. **Remove `And()` calls**: Delete all `.And()` method calls
3. **Update method names**:
   - `SetTotalBandwidth()` → `TotalBandwidth()`
   - `CreateTrafficClass()` → `Class()`
   - `WithGuaranteedBandwidth()` → `Guaranteed()`
   - `WithBurstableTo()` → `BurstTo()`
   - `ForPort()` → `Ports()`
4. **Update structure**: Separate class configurations instead of chaining
5. **Use variadic parameters**: `Ports(80, 443)` instead of multiple `ForPort()` calls

## Benefits

1. **More Readable**: Natural flow without redundant method calls
2. **Less Verbose**: Shorter method names and fewer required calls
3. **More Flexible**: Can configure classes incrementally
4. **Better Error Handling**: Clear validation messages
5. **Enhanced Filtering**: More filtering options with cleaner syntax
6. **Backward Compatible**: Existing v0.1.0 API continues to work

## Validation Rules

The improved API enforces the same validation rules:

- Total bandwidth must be set
- At least one traffic class must be defined
- Each class must have guaranteed bandwidth set
- Each class must have priority set (0-7)
- Priority 0 is highest, 7 is lowest
- Invalid bandwidth formats are rejected
- Invalid priority values are rejected

## Performance

The improved API has the same performance characteristics as the original API, with slightly better memory usage due to reduced intermediate objects.

## Next Steps

- The improved API will become the default in v0.2.0
- The original API will be deprecated but remain available for backward compatibility
- Migration tools may be provided to automatically convert configurations