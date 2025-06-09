# Traffic Control Examples

This directory contains examples demonstrating the traffic control library's capabilities.

## Working Examples

### 1. Basic Examples
- **`basic/main.go`** - Basic functionality test and API usage
- **`production/main.go`** - Production-ready traffic management setup
- **`priority_demo.go`** - Priority-based traffic classification

### 2. Filter Management Examples
- **`filter_management_demo.go`** - Comprehensive filter configuration examples
  - Basic port-based filtering
  - Complex multi-criteria filters
  - Protocol-based traffic classification

### 3. Advanced HTB Examples  
- **`htb_advanced_demo.go`** - Enterprise-grade HTB configurations
  - Enterprise network setup
  - ISP customer management
  - Server farm traffic management

### 4. Configuration Files
- **`config-modern.json`** - Modern JSON configuration example
- **`config-example.yaml`** - YAML configuration example
- **`logging-*.json`** - Logging configuration examples

## Running Examples

### Go Programs
Most examples are marked with `//go:build ignore` to prevent them from being compiled during regular builds. To run them:

```bash
# Navigate to examples directory
cd examples/

# Run a specific demo (will show configuration without applying)
go run filter_management_demo.go

# Run HTB advanced examples
go run htb_advanced_demo.go

# Run working directories
cd basic/ && go run main.go
cd production/ && go run main.go
```

### Important Notes
- Examples demonstrate configuration but don't apply changes to avoid system modification
- For actual traffic control, ensure you have root privileges
- Test in isolated environments before production use

## Example Categories

### Filter Management
The `filter_management_demo.go` shows:
- **Basic Filtering**: Port-based traffic classification
- **Complex Filtering**: Multi-criteria filters with IP ranges
- **Protocol Filtering**: Protocol-specific traffic handling

### HTB Advanced Configuration
The `htb_advanced_demo.go` demonstrates:
- **Enterprise Networks**: Multi-tier priority systems
- **ISP Management**: Customer bandwidth allocation
- **Server Farms**: High-throughput traffic management

### Configuration Files
JSON and YAML examples show:
- Hierarchical class structures
- Filter definitions with multiple match criteria
- Priority-based traffic classification
- IP and port-based filtering rules

## API Usage Patterns

### Modern API (Current)
```go
tc, err := api.NewTrafficController("eth0", config)
tc.WithHardLimitBandwidth("1Gbps")

tc.CreateTrafficClass("Web Traffic").
    WithGuaranteedBandwidth("300Mbps").
    WithSoftLimitBandwidth("600Mbps").
    WithPriority(1).
    ForPort(80, 443)

tc.Apply()
```

### Configuration-Based Setup
```go
tc, err := api.NewTrafficController("eth0", config)
err = tc.ApplyConfigFile("config-modern.json")
```

## Requirements

- Linux system with Traffic Control support
- Root privileges for actual traffic shaping
- Go 1.19+ for building examples

## Development

When adding new examples:
1. Use `//go:build ignore` for demonstration scripts
2. Include comprehensive comments
3. Show realistic use cases
4. Test with mock adapters where possible
5. Document any system requirements