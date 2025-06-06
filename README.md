# Traffic Control Go

[![CI](https://github.com/RNG999/traffic-control-go/actions/workflows/ci.yml/badge.svg)](https://github.com/RNG999/traffic-control-go/actions/workflows/ci.yml)
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
import (
    "github.com/rng999/traffic-control-go/api"
    "github.com/rng999/traffic-control-go/pkg/tc"
    "github.com/rng999/traffic-control-go/pkg/types"
)

// Create a traffic controller
deviceName, _ := tc.NewDeviceName("eth0")
config := api.Config{
    Device: deviceName,
}

result := api.NewTrafficControlService(config)
if result.IsFailure() {
    log.Fatal(result.Error())
}

tcService := result.Value()

// Configure HTB qdisc with bandwidth control
bandwidth := tc.Mbps(1000)
handle := tc.NewHandle(1, 0)
defaultClass := tc.NewHandle(1, 10)

result = tcService.CreateHTBQdisc(handle, defaultClass, bandwidth)
if result.IsFailure() {
    log.Fatal(result.Error())
}
```

Compare this to traditional TC commands:
```bash
tc qdisc add dev eth0 root handle 1: htb default 10
tc class add dev eth0 parent 1: classid 1:1 htb rate 1000mbit ceil 1000mbit
tc class add dev eth0 parent 1:1 classid 1:10 htb rate 100mbit ceil 200mbit
tc filter add dev eth0 parent 1:0 protocol ip prio 1 u32 match ip dst 192.168.1.10/32 flowid 1:10
```

## Core Concepts

### Handles
Handles are unique identifiers for qdiscs and classes in Linux Traffic Control, represented as `major:minor` format:

```go
// Create a handle with major=1, minor=0 (represents "1:0")
rootHandle := tc.NewHandle(1, 0)     // Root qdisc handle
classHandle := tc.NewHandle(1, 10)   // Class handle "1:10"
defaultHandle := tc.NewHandle(1, 30) // Default class "1:30"

// Parse handle from string
handle, err := tc.ParseHandle("1:10") // Parses "1:10"
```

**Handle Rules:**
- **Root qdiscs**: `major:0` (e.g., `1:0`, `2:0`)
- **Classes**: `major:minor` where minor > 0 (e.g., `1:10`, `1:20`)
- **Parent-child relationships**: Classes with same major number belong to the same qdisc
- **Default class**: Special class that handles unclassified traffic

### Bandwidth
Type-safe bandwidth representations with multiple creation methods:

```go
// Direct creation with unit methods
bandwidth1 := tc.Mbps(100)     // 100 Mbps
bandwidth2 := tc.Kbps(500)     // 500 Kbps
bandwidth3 := tc.Gbps(1.5)     // 1.5 Gbps

// Parse from string
bandwidth4, err := tc.ParseBandwidth("100Mbps")
bandwidth5, err := tc.ParseBandwidth("1.5Gbps")
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
// Create service
deviceName, _ := tc.NewDeviceName("eth0")
config := api.Config{Device: deviceName}
tc := api.NewTrafficControlService(config).Value()

// Create root HTB qdisc
rootHandle := tc.NewHandle(1, 0)
defaultHandle := tc.NewHandle(1, 30)
totalBandwidth := tc.Mbps(100)

tc.CreateHTBQdisc(rootHandle, defaultHandle, totalBandwidth)

// Streaming class (40Mbps guaranteed, 60Mbps max)
streamingHandle := tc.NewHandle(1, 10)
streamingRate := tc.Mbps(40)
streamingCeil := tc.Mbps(60)

tc.CreateHTBClass(rootHandle, streamingHandle, "Streaming", streamingRate, streamingCeil)

// Work class (30Mbps guaranteed, can borrow more)
workHandle := tc.NewHandle(1, 20)
workRate := tc.Mbps(30)
workCeil := tc.Mbps(100)

tc.CreateHTBClass(rootHandle, workHandle, "Work", workRate, workCeil)
```

### Priority-Based Traffic Control

```go
// Create PRIO qdisc for priority-based scheduling
deviceName, _ := tc.NewDeviceName("eth0")
config := api.Config{Device: deviceName}
tc := api.NewTrafficControlService(config).Value()

// Create PRIO qdisc with 3 bands
prioHandle := tc.NewHandle(1, 0)
bands := uint8(3)
priomap := []uint8{1, 2, 2, 2, 1, 2, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1}

tc.CreatePRIOQdisc(prioHandle, bands, priomap)

// Create HTB qdiscs on each band for bandwidth control
// Band 0: Critical (highest priority)
criticalHandle := tc.NewHandle(2, 0)
criticalBandwidth := tc.Mbps(200)
tc.CreateHTBQdisc(criticalHandle, tc.NewHandle(2, 1), criticalBandwidth)

// Band 1: Normal traffic
normalHandle := tc.NewHandle(3, 0)
normalBandwidth := tc.Mbps(500)
tc.CreateHTBQdisc(normalHandle, tc.NewHandle(3, 1), normalBandwidth)

// Band 2: Background (lowest priority)
backgroundHandle := tc.NewHandle(4, 0)
backgroundBandwidth := tc.Mbps(100)
tc.CreateHTBQdisc(backgroundHandle, tc.NewHandle(4, 1), backgroundBandwidth)
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
// Load configuration from file
config, err := api.LoadConfigFromYAML("config.yaml")
if err != nil {
    log.Fatal(err)
}

// Create service from config
result := api.NewTrafficControlService(config)
if result.IsFailure() {
    log.Fatal(result.Error())
}

tc := result.Value()

// Or build configuration programmatically
deviceName, _ := tc.NewDeviceName("eth0")
config := api.Config{Device: deviceName}
tc := api.NewTrafficControlService(config).Value()

// Create HTB qdisc and classes
rootHandle := tc.NewHandle(1, 0)
bandwidth := tc.Mbps(1000)

tc.CreateHTBQdisc(rootHandle, tc.NewHandle(1, 30), bandwidth)

// Critical class
criticalHandle := tc.NewHandle(1, 10)
criticalRate := tc.Mbps(400)
criticalCeil := tc.Mbps(600)
tc.CreateHTBClass(rootHandle, criticalHandle, "Critical", criticalRate, criticalCeil)
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