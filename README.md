# Traffic Control Go

[![Test](https://github.com/rng999/traffic-control-go/actions/workflows/test.yml/badge.svg)](https://github.com/rng999/traffic-control-go/actions/workflows/test.yml)
[![Build](https://github.com/rng999/traffic-control-go/actions/workflows/build.yml/badge.svg)](https://github.com/rng999/traffic-control-go/actions/workflows/build.yml)
[![codecov](https://codecov.io/gh/rng999/traffic-control-go/branch/main/graph/badge.svg)](https://codecov.io/gh/rng999/traffic-control-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/rng999/traffic-control-go)](https://goreportcard.com/report/github.com/rng999/traffic-control-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/rng999/traffic-control-go.svg)](https://pkg.go.dev/github.com/rng999/traffic-control-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A human-readable Go library for Linux Traffic Control (TC).

## Overview

This library provides an intuitive API for managing Linux Traffic Control, making complex network traffic management accessible through simple, readable code.

## Features

- **Human-Readable API**: Intuitive method chaining instead of cryptic TC commands
- **Type-Safe**: Leverages Go's type system to prevent configuration errors
- **Event-Driven**: Built with CQRS and Event Sourcing for configuration history
- **Comprehensive**: Supports HTB, PRIO, FQ_CODEL, and other major qdiscs
- **Well-Tested**: Extensive unit and integration tests

## Quick Start

```go
import tc "github.com/example/traffic-control-go"

// Create a traffic controller
controller := tc.New("eth0").
    SetTotalBandwidth("1Gbps")

// Add traffic classes with intuitive methods
err := controller.
    CreateTrafficClass("database").
        WithGuaranteedBandwidth("100Mbps").
        WithMaxBandwidth("200Mbps").
        ForDestination("192.168.1.10").
    Apply()
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
- **Multiple API Styles**: Chain API for programmatic use and Structured Configuration API for YAML/JSON configs
- **Structured Logging**: Comprehensive logging system with context-aware, structured logging built on Zap

## Installation

```bash
go get github.com/example/traffic-control-go
```

## Examples

### Home Network Fair Sharing
```go
controller := tc.New("eth0").
    SetTotalBandwidth("100Mbps")

controller.CreateTrafficClass("streaming").
    WithGuaranteedBandwidth("40Mbps").
    WithBurstableTo("60Mbps").
    ForDevice("smart-tv")

controller.CreateTrafficClass("work").
    WithGuaranteedBandwidth("30Mbps").
    WithPriority(1). // High priority
    ForDevice("laptop")

controller.Apply()
```

### Priority-Based Traffic Control
```go
// Set priority values (0-7, where 0 is highest)
controller.CreateTrafficClass("critical").
    WithGuaranteedBandwidth("50Mbps").
    WithPriority(0). // Highest priority
    ForPort(5060, 5061) // VoIP

controller.CreateTrafficClass("normal").
    WithGuaranteedBandwidth("100Mbps").
    // Default priority is 4
    ForPort(80, 443) // HTTP/HTTPS

controller.CreateTrafficClass("background").
    WithGuaranteedBandwidth("50Mbps").
    WithPriority(7). // Lowest priority
    ForPort(873) // rsync
```

### Configuration File Support
```yaml
# config.yaml
version: "1.0"
device: eth0
bandwidth: 1Gbps
classes:
  - name: critical
    guaranteed: 400Mbps
    priority: high
  - name: standard
    guaranteed: 600Mbps
rules:
  - name: ssh_traffic
    match:
      dest_port: [22]
    target: critical
```

```go
// Apply configuration from file
err := api.LoadAndApplyYAML("config.yaml", "eth0")
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

- [x] Core library with human-readable API
- [x] Basic TC operations (HTB, filters)
- [x] **Structured Configuration API (YAML/JSON)**
- [x] **Numeric Priority System (0-7)**
- [x] **Comprehensive Logging System**
- [x] **CI/CD Pipeline with GitHub Actions**
- [ ] Complete netlink integration
- [ ] Full qdisc support (PRIO, CBQ, HFSC, FQ_CODEL, CAKE)
- [ ] Comprehensive filter types (u32, fw, route)
- [ ] Actions support (police, mirred, nat)
- [ ] Statistics and monitoring API
- [ ] Performance optimizations

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

### Feature Guides
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
make test           # Run all tests
make test-unit      # Run unit tests only
make test-integration # Run integration tests
make lint           # Run golangci-lint
make security       # Run security scanner
make build          # Build for current platform
make build-all      # Build for all platforms
make docker-build   # Build Docker image
make clean          # Clean build artifacts
```

## Testing

```bash
# Run all tests
make test

# Run specific test types
make test-unit
make test-integration
make test-examples

# Run with coverage
make test-coverage

# Run tests manually
go test ./...
go test -cover ./...
go test -v ./test/integration/...
```

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
- **Linting**: golangci-lint with comprehensive checks
- **Security**: Gosec security scanner
- **Coverage**: Codecov integration

### Main Branch Workflow
- **Multi-platform Builds**: Linux, macOS, Windows (amd64/arm64)
- **Docker Images**: Multi-arch container builds
- **Release Automation**: Automatic releases on version tags

### Local CI Testing
```bash
# Run the full CI pipeline locally
make ci

# Check release readiness
make release-check
```