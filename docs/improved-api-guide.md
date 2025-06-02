# Improved API Guide

This document describes the improved API design for traffic-control-go that eliminates redundant `And()` calls and provides a more natural, readable configuration experience.

## Overview

The improved API removes the need for redundant `And()` method calls by allowing direct chaining of class configurations and providing cleaner method names.

## API Comparison

### Before (v0.1.0)
```go
controller := api.New("eth0").
    SetTotalBandwidth("100Mbps").
    CreateTrafficClass("Web Services").
        WithGuaranteedBandwidth("30Mbps").
        WithBurstableTo("60Mbps").
        WithPriority(4).
        ForPort(80, 443).
    And().  // ← Redundant!
    CreateTrafficClass("SSH Management").
        WithGuaranteedBandwidth("5Mbps").
        WithBurstableTo("10Mbps").
        WithPriority(1).
        ForPort(22).
    And().  // ← Redundant!
    Apply()
```

### After (Improved)
```go
tc := api.NewImproved("eth0").
    TotalBandwidth("100Mbps")

tc.Class("Web Services").
    Guaranteed("30Mbps").
    BurstTo("60Mbps").
    Priority(4).
    Ports(80, 443)

tc.Class("SSH Management").
    Guaranteed("5Mbps").
    BurstTo("10Mbps").
    Priority(1).
    Ports(22)

tc.Apply()
```

## Key Improvements

### 1. No More Redundant `And()` Calls
- Eliminated the need for `.And()` between class configurations
- Each class configuration is self-contained and natural

### 2. Cleaner Method Names
- `TotalBandwidth()` instead of `SetTotalBandwidth()`
- `Class()` instead of `CreateTrafficClass()`
- `Guaranteed()` instead of `WithGuaranteedBandwidth()`
- `BurstTo()` instead of `WithBurstableTo()`
- `Ports()` instead of `ForPort()` with variadic parameters

### 3. Natural Flow
- Configure the controller first: `tc := api.NewImproved("eth0").TotalBandwidth("1Gbps")`
- Then configure each class: `tc.Class("name").Guaranteed().Priority().Ports()`
- Finally apply: `tc.Apply()`

### 4. Enhanced Filtering Options
- `Ports(80, 443, 8080)` - Multiple ports in one call
- `SourceIPs("192.168.1.0/24", "10.0.1.100")` - Source IP filtering
- `DestIPs("10.0.2.100", "10.0.2.101")` - Destination IP filtering
- `Protocols("ssh", "http", "https")` - Protocol-based filtering

## Complete API Reference

### Controller Methods

#### `NewImproved(deviceName string) *ImprovedController`
Creates a new improved traffic controller for the specified network device.

```go
tc := api.NewImproved("eth0")
```

#### `TotalBandwidth(bandwidth string) *ImprovedController`
Sets the total bandwidth available on the interface.

```go
tc.TotalBandwidth("1Gbps")   // 1 Gigabit per second
tc.TotalBandwidth("100Mbps") // 100 Megabits per second
tc.TotalBandwidth("500mbit") // 500 megabits per second (alternative format)
```

#### `Class(name string) *ImprovedClass`
Creates or selects a traffic class for configuration. If a class with the same name already exists, it will be reused.

```go
webClass := tc.Class("Web Services")
dbClass := tc.Class("Database")
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

### Class Methods

#### `Guaranteed(bandwidth string) *ImprovedClass`
Sets the guaranteed bandwidth for the class.

```go
class.Guaranteed("100Mbps")
class.Guaranteed("50mbit")
```

#### `BurstTo(bandwidth string) *ImprovedClass`
Sets the maximum burstable bandwidth for the class.

```go
class.BurstTo("200Mbps")
```

#### `Priority(priority int) *ImprovedClass`
Sets the priority for the class (0-7, where 0 is highest priority).

```go
class.Priority(0)  // Highest priority
class.Priority(4)  // Normal priority
class.Priority(7)  // Lowest priority
```

#### `Ports(ports ...int) *ImprovedClass`
Adds port-based filtering to the class. Can be called multiple times to add more ports.

```go
class.Ports(80, 443)           // HTTP and HTTPS
class.Ports(22)                // SSH (adds to existing ports)
class.Ports(3306, 5432, 1521)  // Database ports
```

#### `SourceIPs(ips ...string) *ImprovedClass`
Adds source IP-based filtering to the class.

```go
class.SourceIPs("192.168.1.0/24")                    // Subnet
class.SourceIPs("10.0.1.100", "10.0.1.101")         // Specific IPs
class.SourceIPs("192.168.100.0/24", "172.16.0.0/16") // Multiple subnets
```

#### `DestIPs(ips ...string) *ImprovedClass`
Adds destination IP-based filtering to the class.

```go
class.DestIPs("192.168.1.10")        // Single server
class.DestIPs("10.0.2.0/24")         // Server subnet
```

#### `Protocols(protocols ...string) *ImprovedClass`
Adds protocol-based filtering to the class.

```go
class.Protocols("tcp", "udp")
class.Protocols("ssh", "http", "https")
class.Protocols("dns", "ntp")
```

## Usage Examples

### Basic Web Server Configuration
```go
tc := api.NewImproved("eth0").
    TotalBandwidth("1Gbps")

tc.Class("Web Traffic").
    Guaranteed("400Mbps").
    BurstTo("600Mbps").
    Priority(3).
    Ports(80, 443, 8080, 8443)

tc.Class("SSH Management").
    Guaranteed("10Mbps").
    BurstTo("50Mbps").
    Priority(1).
    Ports(22)

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