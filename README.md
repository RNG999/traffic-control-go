# Traffic Control Go

[![CI](https://github.com/RNG999/traffic-control-go/actions/workflows/ci.yml/badge.svg)](https://github.com/RNG999/traffic-control-go/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/RNG999/traffic-control-go/branch/main/graph/badge.svg)](https://codecov.io/gh/RNG999/traffic-control-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/RNG999/traffic-control-go)](https://goreportcard.com/report/github.com/RNG999/traffic-control-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/RNG999/traffic-control-go.svg)](https://pkg.go.dev/github.com/RNG999/traffic-control-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/release/RNG999/traffic-control-go.svg)](https://github.com/RNG999/traffic-control-go/releases/latest)

A human-readable Go library for Linux Traffic Control (TC)

## Overview

This library provides an intuitive API for managing Linux Traffic Control, making complex network traffic management accessible through simple, readable code.

## Features

- **Human-Readable API**: Clean, intuitive method chaining without redundant calls
- **Type-Safe**: Leverages Go's type system to prevent configuration errors
- **Event-Driven**: Built with CQRS and Event Sourcing for configuration history
- **Multiple Qdiscs**: HTB, TBF, PRIO, FQ_CODEL with complete CQRS integration
- **Event Sourcing**: SQLite-based persistent event store for configuration history
- **Statistics**: Real-time traffic monitoring and metrics collection
- **Well-Tested**: Extensive unit and integration tests with iperf3
- **Production Ready**: Battle-tested API for enterprise applications

## Quick Start

```go
import "github.com/rng999/traffic-control-go/api"

// Create a traffic controller with human-readable API
tc := api.NetworkInterface("eth0").
    WithHardLimitBandwidth("1Gbps")  // Physical interface limit

// Configure traffic classes with clear bandwidth concepts
tc.CreateTrafficClass("Database").
    WithGuaranteedBandwidth("100Mbps").     // Minimum guarantee
    WithSoftLimitBandwidth("200Mbps").      // Policy limit (borrowing allowed)
    WithPriority(2).                        // High priority
    ForDestinationIPs("192.168.1.10").     // Target specific server
    AddClass()                              // Complete class configuration

tc.Apply()
```

Compare this to traditional TC commands:
```bash
tc qdisc add dev eth0 root handle 1: htb default 10
tc class add dev eth0 parent 1: classid 1:1 htb rate 1000mbit ceil 1000mbit
tc class add dev eth0 parent 1:1 classid 1:10 htb rate 100mbit ceil 200mbit
tc filter add dev eth0 parent 1:0 protocol ip prio 1 u32 match ip dst 192.168.1.10/32 flowid 1:10
```

## Library Design

This library focuses on providing a clean, intuitive API for Linux Traffic Control operations:

- **Value Objects**: Type-safe representations of bandwidth, handles, etc.
- **Domain Entities**: Qdiscs, Classes, Filters as first-class objects
- **Event Sourcing**: Track all configuration changes
- **Netlink Integration**: Direct kernel communication for TC operations
- **Flexible Configuration**: Chain API for programmatic use and Structured Configuration API for YAML/JSON configs
- **Structured Logging**: Comprehensive logging system with context-aware, structured logging built on Zap

## Installation

```bash
go get github.com/rng999/traffic-control-go
```

## Examples

### Home Network Fair Sharing

```go
tc := api.NetworkInterface("eth0").
    WithHardLimitBandwidth("100Mbps")

tc.CreateTrafficClass("Streaming").
    WithGuaranteedBandwidth("40Mbps").
    WithSoftLimitBandwidth("60Mbps").
    WithPriority(4).
    ForSourceIPs("192.168.1.100").  // Smart TV
    AddClass()

tc.CreateTrafficClass("Work").
    WithGuaranteedBandwidth("30Mbps").
    WithPriority(1).  // High priority
    ForSourceIPs("192.168.1.101").  // Laptop
    AddClass()

tc.Apply()
```

### Priority-Based Traffic Control

```go
tc := api.NetworkInterface("eth0").
    WithHardLimitBandwidth("1Gbps")

// Set priority values (0-7, where 0 is highest)
tc.CreateTrafficClass("Critical Services").
    WithGuaranteedBandwidth("200Mbps").
    WithSoftLimitBandwidth("400Mbps").
    WithPriority(0).  // Highest priority
    ForPort(5060, 5061).  // VoIP
    ForProtocols("rtp", "sip").
    AddClass()

tc.CreateTrafficClass("Normal Traffic").
    WithGuaranteedBandwidth("500Mbps").
    WithSoftLimitBandwidth("800Mbps").
    WithPriority(4).  // Default priority
    ForPort(80, 443, 8080, 8443).  // HTTP/HTTPS
    AddClass()

tc.CreateTrafficClass("Background").
    WithGuaranteedBandwidth("100Mbps").
    WithSoftLimitBandwidth("200Mbps").
    WithPriority(7).  // Lowest priority
    ForPort(873, 22).  // rsync, SSH
    ForProtocols("rsync", "ssh").
    AddClass()

tc.Apply()
```

### Configuration File Support

**YAML Configuration:**
```yaml
# config.yaml
version: "1.0"
device: eth0
bandwidth: 1Gbps
classes:
  - name: critical
    guaranteed: 400Mbps
    max: 600Mbps
    priority: 0 # Highest priority
  - name: standard
    guaranteed: 600Mbps
    max: 800Mbps
    priority: 4 # Normal priority
rules:
  - name: ssh_traffic
    match:
      dest_port: [22]
    target: critical
  - name: web_traffic
    match:
      dest_port: [80, 443]
    target: standard
```

**Programmatic Configuration:**
```go
// Apply configuration from file
err := api.LoadAndApplyYAML("config.yaml", "eth0")

// Or build configuration programmatically
tc := api.NetworkInterface("eth0").WithHardLimitBandwidth("1Gbps")

tc.CreateTrafficClass("Critical").
    WithGuaranteedBandwidth("400Mbps").
    WithSoftLimitBandwidth("600Mbps").
    WithPriority(0).
    ForPort(22).
    AddClass()
    
tc.CreateTrafficClass("Standard").
    WithGuaranteedBandwidth("600Mbps").
    WithSoftLimitBandwidth("800Mbps").
    WithPriority(4).
    ForPort(80, 443).
    AddClass()
    
tc.Apply()
```

### Logging Configuration
```go
import "github.com/rng999/traffic-control-go/pkg/logging"

// Initialize logging (call once at startup)
logging.Initialize(&logging.Config{
    Level:  "info",
    Format: "json",
    Outputs: []string{"stdout"},
})

// Use structured logging in your application
logger := logging.WithComponent(logging.ComponentAPI).
    WithDevice("eth0")
    
logger.Info("Traffic control operation started")
```

## Roadmap

### v0.1.0 (Released! 🎉)
- [x] Core library with human-readable API
- [x] Basic TC operations (HTB, filters)
- [x] Structured Configuration API (YAML/JSON)
- [x] Numeric Priority System (0-7)
- [x] Comprehensive Logging System
- [x] CI/CD Pipeline with GitHub Actions
- [x] Extended Qdisc Support (HTB, TBF, PRIO, FQ_CODEL)
- [x] SQLite Event Store for persistent storage
- [x] Statistics Collection and monitoring

### v0.2.0 (In Development)
- [x] **Human-Readable API Naming** - Clear method names and bandwidth concepts
- [x] **Hard vs Soft Bandwidth Limits** - Distinct concepts for physical and policy limits
- [ ] CQRS query handler interface compatibility fixes
- [ ] Enhanced statistics collection with detailed metrics
- [ ] Mark-based filtering (fw filter)

### Future Releases
- [ ] NETEM qdisc (network emulation)
- [ ] Flower filter types and police actions
- [ ] Performance optimization and benchmarks
- [ ] REST API server mode
- [ ] Kubernetes integration
- [ ] Web UI dashboard

## Requirements

- Linux kernel 3.10+
- Go 1.21+
- CAP_NET_ADMIN capability or root access
- iproute2 package installed

## Documentation

### Core Documentation
- [API Design](memory-bank/api-design.md) - Human-readable API examples
- [Architecture](memory-bank/architecture-overview.md) - System architecture
- [Traffic Control Basics](docs/traffic-control.md) - Linux TC fundamentals

### API Guides
- **[API Guide](docs/improved-api-guide.md) - Human-readable API with clear bandwidth concepts**
- [Structured Configuration API](docs/structured-config-api.md) - YAML/JSON configuration support
- [Priority System Guide](docs/priority-guide.md) - Numeric priority system (0-7)
- [Logging System](docs/logging.md) - Comprehensive structured logging

### Reference
- [TC Feature Coverage](docs/tc-feature-coverage.md) - Current implementation status
- [Documentation Hub](docs/README.md) - Complete documentation index

## Contributing

We welcome contributions! Please see our contributing guidelines for details.

## License

Apache License 2.0

## Development

### Prerequisites
- Go 1.21 or higher
- Linux system with Traffic Control support
- Root privileges for integration tests
- iperf3 installed for integration tests

### Development Workflow
```bash
# Clone the repository
git clone https://github.com/rng999/traffic-control-go
cd traffic-control-go

# Run tests
make test-unit          # Unit tests (no root required)
sudo make test-integration  # Integration tests (requires root)

# Format and lint
make fmt
make lint

# Run all quality checks
make check
```

## Testing

```bash
# Run all tests
make test

# Run specific test types
make test-unit
make test-integration  # Requires root privileges and iperf3
make test-examples

# Run with coverage
make test-coverage

# Run tests manually
go test ./...
go test -cover ./...
sudo go test -v ./test/integration/...  # Integration tests require root
```

### Integration Test Requirements
- Root privileges (uses network namespaces and virtual interfaces)
- iperf3 installed (`sudo apt-get install iperf3` or equivalent)
- Linux kernel with veth support

## Logging

The project includes a comprehensive structured logging system built on Uber's Zap library:

### Features
- **Structured Logging**: JSON and console output formats
- **Context-Aware**: Automatic context for devices, classes, operations
- **Configurable**: Environment variables, config files, programmatic setup
- **High Performance**: Built on Zap for minimal overhead

### Quick Start

```go
import "github.com/rng999/traffic-control-go/pkg/logging"

// Initialize logging
logging.InitializeDevelopment()

// Context-aware logging
logger := logging.WithComponent(logging.ComponentAPI).
    WithDevice("eth0").
    WithOperation(logging.OperationCreateClass)

logger.Info("Creating traffic class", 
    logging.String("class_name", "web-traffic"),
    logging.String("bandwidth", "100Mbps"),
)
```

### Environment Configuration

```bash
export TC_LOG_LEVEL=info          # debug, info, warn, error, fatal
export TC_LOG_FORMAT=json         # json, console
export TC_LOG_OUTPUTS=stdout      # stdout, stderr, or file paths
export TC_LOG_DEVELOPMENT=false   # true/false
```

For detailed logging documentation, see [docs/logging.md](docs/logging.md).

## CI/CD

This project uses GitHub Actions for continuous integration:

### Pull Request Workflow
- **Test Suite**: Runs on Go 1.21, 1.22, and 1.23
- **Unit Tests**: Run on all platforms
- **Integration Tests**: Run with root privileges on Linux runners
- **Linting**: golangci-lint with comprehensive checks
- **Security**: Gosec security scanner
- **Coverage**: Codecov integration with detailed reports

### Quality Assurance
- Comprehensive unit tests for all components
- Integration tests with real network interfaces (veth pairs)
- Bandwidth validation tests using iperf3
- Event sourcing and CQRS pattern validation
- Statistics collection accuracy tests