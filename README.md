# Traffic Control Go

[![Test](https://github.com/rng999/traffic-control-go/actions/workflows/test.yml/badge.svg)](https://github.com/rng999/traffic-control-go/actions/workflows/test.yml)
[![Build](https://github.com/rng999/traffic-control-go/actions/workflows/build.yml/badge.svg)](https://github.com/rng999/traffic-control-go/actions/workflows/build.yml)
[![codecov](https://codecov.io/gh/rng999/traffic-control-go/branch/main/graph/badge.svg)](https://codecov.io/gh/rng999/traffic-control-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/rng999/traffic-control-go)](https://goreportcard.com/report/github.com/rng999/traffic-control-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/rng999/traffic-control-go.svg)](https://pkg.go.dev/github.com/rng999/traffic-control-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/release/rng999/traffic-control-go.svg)](https://github.com/rng999/traffic-control-go/releases/latest)

A human-readable Go library for Linux Traffic Control (TC) - **v0.1.0 Released! Date-based Releases Available**

## Overview

This library provides an intuitive API for managing Linux Traffic Control, making complex network traffic management accessible through simple, readable code.

## Features

- **Improved Human-Readable API**: Clean, intuitive method chaining without redundant calls
- **Dual API Design**: Classic API (v0.1.0) and Improved API (v0.2.0+) for better developer experience
- **Type-Safe**: Leverages Go's type system to prevent configuration errors
- **Event-Driven**: Built with CQRS and Event Sourcing for configuration history
- **Multiple Qdiscs**: HTB, TBF, PRIO, FQ_CODEL with complete CQRS integration
- **Event Sourcing**: SQLite-based persistent event store for configuration history
- **Statistics**: Real-time traffic monitoring and metrics collection
- **CLI Tool**: Standalone binary for command-line traffic control management
- **Well-Tested**: Extensive unit and integration tests with iperf3
- **CI/CD Pipeline**: Fully automated testing and release workflows

## Quick Start

### Improved API (Recommended)

```go
import "github.com/rng999/traffic-control-go/api"

// Create a traffic controller with the improved API
tc := api.NewImproved("eth0").
    TotalBandwidth("1Gbps")

// Configure traffic classes with clean, natural syntax
tc.Class("Database").
    Guaranteed("100Mbps").
    BurstTo("200Mbps").
    Priority(2).
    DestIPs("192.168.1.10")

tc.Apply()
```

### Classic API (v0.1.0 - Still Supported)

```go
import "github.com/rng999/traffic-control-go/api"

// Classic API with And() calls
controller := api.New("eth0").
    SetTotalBandwidth("1Gbps")

err := controller.
    CreateTrafficClass("database").
        WithGuaranteedBandwidth("100Mbps").
        WithMaxBandwidth("200Mbps").
        ForDestination("192.168.1.10").
    Apply()
```

**Benefits of the Improved API:**
- ✅ **No redundant And() calls** - cleaner method chaining
- ✅ **Shorter method names** - `Guaranteed()` vs `WithGuaranteedBandwidth()`
- ✅ **Natural flow** - configure controller first, then classes
- ✅ **Variadic parameters** - `Ports(80, 443, 8080)` in one call
- ✅ **Enhanced filtering** - IP ranges, protocols, multiple criteria
- ✅ **Class reuse** - configure the same class incrementally

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
- **Multiple API Styles**: Chain API for programmatic use and Structured Configuration API for YAML/JSON configs
- **Structured Logging**: Comprehensive logging system with context-aware, structured logging built on Zap

## Installation

```bash
go get github.com/rng999/traffic-control-go
```

## Examples

### Home Network Fair Sharing

**Improved API:**
```go
tc := api.NewImproved("eth0").
    TotalBandwidth("100Mbps")

tc.Class("Streaming").
    Guaranteed("40Mbps").
    BurstTo("60Mbps").
    Priority(4).
    SourceIPs("192.168.1.100") // Smart TV

tc.Class("Work").
    Guaranteed("30Mbps").
    Priority(1). // High priority
    SourceIPs("192.168.1.101") // Laptop

tc.Apply()
```

**Classic API:**
```go
controller := api.New("eth0").
    SetTotalBandwidth("100Mbps")

controller.CreateTrafficClass("streaming").
    WithGuaranteedBandwidth("40Mbps").
    WithBurstableTo("60Mbps").
    WithPriority(4).
    ForSource("192.168.1.100"). // Smart TV
    And().
CreateTrafficClass("work").
    WithGuaranteedBandwidth("30Mbps").
    WithPriority(1). // High priority
    ForSource("192.168.1.101"). // Laptop
    And().
Apply()
```

### Priority-Based Traffic Control

**Improved API:**
```go
tc := api.NewImproved("eth0").
    TotalBandwidth("1Gbps")

// Set priority values (0-7, where 0 is highest)
tc.Class("Critical Services").
    Guaranteed("200Mbps").
    BurstTo("400Mbps").
    Priority(0). // Highest priority
    Ports(5060, 5061). // VoIP
    Protocols("rtp", "sip")

tc.Class("Normal Traffic").
    Guaranteed("500Mbps").
    BurstTo("800Mbps").
    Priority(4). // Default priority
    Ports(80, 443, 8080, 8443) // HTTP/HTTPS

tc.Class("Background").
    Guaranteed("100Mbps").
    BurstTo("200Mbps").
    Priority(7). // Lowest priority
    Ports(873, 22). // rsync, SSH
    Protocols("rsync", "ssh")

tc.Apply()
```

**Classic API:**
```go
controller := api.New("eth0").
    SetTotalBandwidth("1Gbps")

// Set priority values (0-7, where 0 is highest)
controller.CreateTrafficClass("critical").
    WithGuaranteedBandwidth("200Mbps").
    WithPriority(0). // Highest priority
    ForPort(5060, 5061). // VoIP
    And().
CreateTrafficClass("normal").
    WithGuaranteedBandwidth("500Mbps").
    WithPriority(4). // Default priority
    ForPort(80, 443). // HTTP/HTTPS
    And().
CreateTrafficClass("background").
    WithGuaranteedBandwidth("100Mbps").
    WithPriority(7). // Lowest priority
    ForPort(873). // rsync
    And().
Apply()
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
tc := api.NewImproved("eth0").TotalBandwidth("1Gbps")
tc.Class("Critical").Guaranteed("400Mbps").BurstTo("600Mbps").Priority(0).Ports(22)
tc.Class("Standard").Guaranteed("600Mbps").BurstTo("800Mbps").Priority(4).Ports(80, 443)
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
- [x] Standalone CLI Binary (traffic-control command)
- [x] GoReleaser & Release Please for automated releases

### v0.2.0 (In Development)
- [x] **Improved API Design** - Clean method chaining without redundant And() calls
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
- **[Improved API Guide](docs/improved-api-guide.md) - New improved API without redundant And() calls**
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

### Quick Start
```bash
# Set up development environment
make dev-setup

# Run all tests and checks
make check

# Build for all platforms
make build-all
```

### Available Make Targets
```bash
make help           # Show all available targets
make build          # Build both binaries (traffic-control, tcctl)
make test           # Run all tests
make clean          # Clean build artifacts
make install        # Install binaries to system
make dev            # Set up development environment
make fmt            # Format code
make lint           # Run basic linting
make check          # Run all quality checks
make version        # Show current version
make release-simple # Simple release (manual)
make release-goreleaser # Release with GoReleaser
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

This project uses GitHub Actions for continuous integration and deployment:

### Pull Request Workflow
- **Test Suite**: Runs on Go 1.21, 1.22, and 1.23
- **Unit Tests**: Run on all platforms
- **Integration Tests**: Run with root privileges on Linux runners
- **Linting**: golangci-lint with comprehensive checks
- **Security**: Gosec security scanner
- **Coverage**: Codecov integration with detailed reports

### Release Workflow
- **Multi-platform Builds**: Linux, macOS, Windows (amd64/arm64) via GoReleaser
- **Automated Versioning**: Release Please for semantic versioning
- **GitHub Releases**: Automated releases with binary attachments
- **Version v0.1.0**: First feature-complete release ready!

### CLI Tool Usage
```bash
# Install the CLI tool
make install

# Basic traffic shaping with TBF
sudo traffic-control tbf eth0 1:0 100Mbps

# Priority scheduling with PRIO
sudo traffic-control prio eth0 1:0 3

# Fair queuing with FQ_CODEL
sudo traffic-control fq_codel eth0 1:0 --target 1000 --ecn

# Show statistics
sudo traffic-control stats eth0

# Show version
traffic-control --version
```