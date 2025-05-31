# Structured Configuration API

The Traffic Control library supports structured configuration through YAML and JSON files, allowing you to define complex traffic control setups declaratively.

## Overview

The Structured Configuration API provides:
- YAML and JSON configuration file support
- Hierarchical class definitions
- Traffic matching rules
- Default settings for all classes
- Validation before applying configurations

## Configuration Structure

### Basic Structure

```yaml
version: "1.0"
device: eth0
bandwidth: 1Gbps

defaults:
  burst_ratio: 1.5
  priority: 4      # Default priority value

classes:
  - name: high_priority
    guaranteed: 400Mbps
    maximum: 800Mbps
    priority: 1    # High priority
    
  - name: standard
    guaranteed: 600Mbps

rules:
  - name: ssh_traffic
    match:
      dest_port: [22]
    target: high_priority
    priority: 1
```

### Configuration Fields

#### Root Fields
- `version`: Configuration version (currently "1.0")
- `device`: Network interface name
- `bandwidth`: Total interface bandwidth
- `defaults`: Default settings for all classes (optional)
- `classes`: List of traffic classes
- `rules`: List of traffic matching rules (optional)

#### Default Settings
- `burst_ratio`: Maximum bandwidth multiplier (default: 1.5)
- `priority`: Default priority level (default: 4)

#### Class Configuration
- `name`: Unique class name
- `guaranteed`: Guaranteed bandwidth
- `maximum`: Maximum bandwidth (optional, uses burst_ratio if not set)
- `priority`: Priority level - numeric value 0-7 (where 0 = highest priority, default: 4)
- `children`: Nested child classes (optional)

#### Rule Configuration
- `name`: Rule name
- `match`: Match conditions
  - `source_ip`: Source IP address/CIDR
  - `destination_ip`: Destination IP address/CIDR
  - `source_port`: List of source ports
  - `dest_port`: List of destination ports
  - `protocol`: Protocol ("tcp", "udp", etc.)
  - `application`: List of application names
- `target`: Target class name (supports hierarchical names)
- `priority`: Rule priority (lower numbers = higher priority)

## Usage Examples

### Loading and Applying Configuration

```go
// Load and apply YAML configuration
err := api.LoadAndApplyYAML("config.yaml", "eth0")
if err != nil {
    log.Fatal(err)
}

// Load and apply JSON configuration
err = api.LoadAndApplyJSON("config.json", "eth0")
if err != nil {
    log.Fatal(err)
}
```

### Loading and Modifying Configuration

```go
// Load configuration
config, err := api.LoadConfigFromYAML("config.yaml")
if err != nil {
    log.Fatal(err)
}

// Modify configuration
config.Device = "eth1"
config.Classes = append(config.Classes, api.TrafficClassConfig{
    Name:       "emergency",
    Guaranteed: "100Mbps",
    Priority:   0, // Highest priority
})

// Apply modified configuration
controller := api.New(config.Device)
err = controller.ApplyConfig(config)
if err != nil {
    log.Fatal(err)
}
```

### Programmatic Configuration Creation

```go
config := &api.TrafficControlConfig{
    Version:   "1.0",
    Device:    "eth0",
    Bandwidth: "1Gbps",
    Defaults: &api.DefaultConfig{
        BurstRatio: 2.0,
        Priority:   4, // Default priority
    },
    Classes: []api.TrafficClassConfig{
        {
            Name:       "critical",
            Guaranteed: "400Mbps",
            Maximum:    "800Mbps",
            Priority:   1, // High priority
        },
        {
            Name:       "standard",
            Guaranteed: "600Mbps",
        },
    },
    Rules: []api.TrafficRuleConfig{
        {
            Name: "ssh_traffic",
            Match: api.MatchConfig{
                DestPort: []int{22},
                Protocol: "tcp",
            },
            Target:   "critical",
            Priority: 1,
        },
    },
}

controller := api.New(config.Device)
err := controller.ApplyConfig(config)
```

## Hierarchical Classes

The configuration API supports nested class hierarchies:

```yaml
classes:
  - name: business
    guaranteed: 800Mbps
    children:
      - name: critical
        guaranteed: 500Mbps
        children:
          - name: voip
            guaranteed: 200Mbps
          - name: database
            guaranteed: 300Mbps
      - name: standard
        guaranteed: 300Mbps

rules:
  - name: voip_rule
    match:
      dest_port: [5060, 5061]
    target: business.critical.voip  # Hierarchical target
```

## Validation

The configuration is validated before application:
- Device name must be specified
- Total bandwidth must be specified
- At least one class must be defined
- Class names must be unique (including hierarchical names)
- Rule targets must reference existing classes
- Total guaranteed bandwidth cannot exceed interface bandwidth

## Complete Example

See `examples/config-example.yaml` and `examples/config-example.json` for complete configuration examples demonstrating:
- Hierarchical class structures
- Multiple traffic matching rules
- Default settings
- Various priority levels

## Integration with Chain API

The Structured Configuration API integrates seamlessly with the existing chain API. You can:
1. Use configuration files for initial setup
2. Modify configurations programmatically using the chain API
3. Mix both approaches as needed

```go
// Start with configuration file
config, _ := api.LoadConfigFromYAML("base-config.yaml")
controller := api.New(config.Device)
controller.ApplyConfig(config)

// Add additional rules using chain API
controller.CreateTrafficClass("dynamic").
    WithGuaranteedBandwidth("50Mbps").
    ForPort(8080).
    Apply()
```