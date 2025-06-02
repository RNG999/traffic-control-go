# Traffic Control Go Library Documentation

Welcome to the Traffic Control Go library documentation. This directory contains comprehensive documentation for understanding and using the library.

## Project Status

🎉 **v0.1.0 - A human-readable Go library for Linux Traffic Control**
- Complete API for programmatic traffic control
- Both Classic and Improved API styles
- Comprehensive test coverage (unit and integration)
- Production-ready for library integration

## Table of Contents

### Getting Started
- [Installation Guide](installation.md) - Library installation and requirements
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

### Architecture
- [System Architecture](../memory-bank/architecture-overview.md) - Library design and patterns
- [SQLite Event Store](sqlite-event-store.md) - Persistent event storage implementation
- [Event Store Data](event-store-data.md) - Event store schema and data format
- [Design Alternatives](design_alternatives.md) - Alternative design considerations

### Examples
- [Basic Demo](../examples/basic_demo.go) - Complete working examples
- [Priority Demo](../examples/priority_demo.go) - Priority-based traffic shaping
- [Qdisc Types Demo](../examples/qdisc_types_demo.go) - TBF, PRIO, FQ_CODEL examples
- [Statistics Demo](../examples/statistics_demo.go) - Statistics collection examples
- [YAML Configuration](../examples/config-example.yaml) - Example YAML configuration
- [JSON Configuration](../examples/config-example.json) - Example JSON configuration
- [Logging Configurations](../examples/) - Development and production logging examples

## Quick Links

- **Repository**: https://github.com/rng999/traffic-control-go
- **Go Package**: https://pkg.go.dev/github.com/rng999/traffic-control-go
- **Issues**: https://github.com/rng999/traffic-control-go/issues
- **Installation**: `go get github.com/rng999/traffic-control-go`

## Development Quick Reference

```bash
# Run all tests (requires root for integration tests)
make test

# Run only unit tests
make test-unit

# Run integration tests (requires root and iperf3)
sudo make test-integration

# Format and lint code
make fmt
make lint

# Run tests with coverage
make test-coverage

# Full quality check
make check
```

## Contributing

See the main [README](../README.md) for contribution guidelines.