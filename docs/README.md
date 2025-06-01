# Traffic Control Go Documentation

Welcome to the Traffic Control Go library documentation. This directory contains comprehensive documentation for understanding and using the library.

## Table of Contents

### Getting Started
- [Traffic Control Basics](traffic-control.md) - Understanding Linux Traffic Control fundamentals
- [API Overview](../memory-bank/api-design.md) - Introduction to the human-readable API

### API Documentation
- [Chain API](../memory-bank/api-design.md) - Programmatic fluent API for traffic control
- [Structured Configuration API](structured-config-api.md) - YAML/JSON configuration file support

### Feature Guides
- [Structured Configuration API](structured-config-api.md) - YAML/JSON configuration support
- [Priority System Guide](priority-guide.md) - Numeric priority system (0-7)
- [Logging System](logging.md) - Comprehensive structured logging
- [Testing Guide](testing.md) - Unit and integration testing with iperf3
- [TC Feature Coverage](tc-feature-coverage.md) - Current implementation status and roadmap
- [Standalone Binary](standalone-binary.md) - CLI tool usage guide

### Architecture
- [System Architecture](../memory-bank/architecture-overview.md) - Library design and patterns
- [Domain Model](../memory-bank/domain-model.md) - Core entities and value objects

### Examples
- [Basic Demo](../examples/basic_demo.go) - Complete working examples
- [Qdisc Types Demo](../examples/qdisc_types_demo.go) - TBF, PRIO, FQ_CODEL examples
- [Statistics Demo](../examples/statistics_demo.go) - Statistics collection examples
- [YAML Configuration](../examples/config-example.yaml) - Example YAML configuration
- [JSON Configuration](../examples/config-example.json) - Example JSON configuration

## Quick Links

- **Repository**: https://github.com/rng999/traffic-control-go
- **Issues**: https://github.com/rng999/traffic-control-go/issues
- **Go Package**: `go get github.com/rng999/traffic-control-go`
- **CLI Tool**: `make install` to install traffic-control binary

## Contributing

See the main [README](../README.md) for contribution guidelines.