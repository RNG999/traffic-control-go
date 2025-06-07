# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go library for Linux Traffic Control (TC) that provides an intuitive, human-readable API for managing network traffic. The library follows CQRS, Event Sourcing, and Domain-Driven Design principles.

## Important: Project Rules

**ALWAYS refer to and follow the rules defined in the `rules/` directory:**
- `rules/coding.md` - Comprehensive coding guidelines (CQRS, DDD, FP, testing, etc.)
- `rules/management.md` - Project management guidelines using GitHub

**When you receive a request for a coding task (e.g., implementing a feature, fixing a bug), you MUST consult `rules/coding.md` and implement the solution strictly according to its guidelines.**

**When you receive a request for a management task (e.g., task management), you MUST consult `rules/management.md` and execute the task strictly according to its defined procedures.**

These rules must be consulted and strictly followed for any coding or project management tasks.

## Development Commands

Common commands for this project (from Makefile):

```bash
# Run tests
make test                   # Run all unit tests
make test-coverage          # Run tests with coverage report
make test-integration       # Run integration tests (requires root and iperf3)

# Code quality
make fmt                    # Format code
make lint                   # Run linting (golangci-lint)
make security              # Run security scanner (gosec)
make check                 # Run all quality checks (fmt, lint, test)

# Development
make dev                   # Set up development environment
make examples              # Build and test examples
make docs                  # Generate documentation

# Other
make version               # Show current version
make info                  # Show project info
make help                  # Show all available commands
```

Manual commands:

```bash
# Run specific tests
go test -v ./internal/domain/...
go test -run TestSpecificName ./...

# Run benchmarks
go test -bench=. ./...

# Generate test coverage for specific packages
go test -cover ./internal/application/...
```

## High-Level Architecture

This library implements Linux Traffic Control management with the following architecture:

### Core Architecture Patterns

1. **CQRS (Command Query Responsibility Segregation)**
   - Commands: Located in `internal/commands/` - handle write operations (creating qdiscs, classes, filters)
   - Queries: Located in `internal/queries/` - handle read operations (statistics, current configuration)
   - Clear separation between write and read models for optimal performance

2. **Event Sourcing**
   - All configuration changes are stored as immutable events in `internal/domain/events/`
   - SQLite-based event store in `internal/infrastructure/eventstore/`
   - Aggregate state reconstructed by replaying events
   - Enables configuration history and rollback capabilities

3. **Domain-Driven Design (DDD)**
   - Value Objects: `Bandwidth`, `Handle`, `Device` in `internal/domain/valueobjects/`
   - Entities: `Qdisc`, `Class`, `Filter` in `internal/domain/entities/`
   - Aggregates: `TrafficControl` in `internal/domain/aggregates/`
   - Rich domain model with business logic encapsulated in domain objects

### Key Components

- **API Layer** (`api/`): Human-readable fluent API for traffic control operations
- **Application Layer** (`internal/application/`): Orchestrates commands, queries, and events
- **Infrastructure Layer** (`internal/infrastructure/`):
  - `netlink/`: Direct kernel communication for TC operations
  - `eventstore/`: Event persistence (SQLite and in-memory implementations)
- **Projections** (`internal/projections/`): Read models built from events

### Traffic Control Concepts Mapping

- **Hard Limit Bandwidth**: Physical interface capacity (HTB root rate)
- **Soft Limit Bandwidth**: Policy-based maximum (HTB ceil rate)
- **Guaranteed Bandwidth**: Minimum guaranteed rate (HTB rate)
- **Priority**: Numeric value 0-7 (0 = highest priority)

## Code Structure Guidelines

When implementing features in this traffic control system:

### Directory Structure
```
traffic-control-go/
├── api/                     # Human-readable API layer
│   ├── api.go              # Main API implementation
│   └── config.go           # Configuration file support
├── internal/                # Private application code
│   ├── domain/             # Domain models and business logic
│   │   ├── aggregates/     # TrafficControl aggregate
│   │   ├── entities/       # Qdisc, Class, Filter entities
│   │   ├── events/         # Domain events for all operations
│   │   └── valueobjects/   # Bandwidth, Handle, Device
│   ├── commands/           # CQRS Commands
│   │   ├── handlers/       # Command handlers
│   │   └── models/         # Command DTOs
│   ├── queries/            # CQRS Queries
│   │   ├── handlers/       # Query handlers
│   │   └── models/         # Views and query models
│   ├── infrastructure/     # External concerns
│   │   ├── netlink/        # Kernel communication
│   │   └── eventstore/     # Event persistence
│   ├── application/        # Application services
│   └── projections/        # Read model projections
├── pkg/                    # Public packages
│   ├── types/              # Result type for error handling
│   └── logging/            # Structured logging
├── test/                   # Integration tests
│   ├── integration/        # Tests with real interfaces
│   └── unit/               # API unit tests
└── examples/               # Usage examples
```

### Testing Strategy

- **TDD Required**: Write tests first (Red/Green/Refactor cycle)
- **Table-driven tests**: Use for comprehensive test coverage
- **Unit tests**: Alongside code (`*_test.go`)
- **Integration tests**: In `test/integration/` using veth pairs
- **iperf3 tests**: For realistic bandwidth validation
- Use `github.com/stretchr/testify/assert` for assertions

### Important Project-Specific Notes

1. **Always use the Result type** from `pkg/types/` for error handling:
   ```go
   func DoSomething() types.Result[Output] {
       // Implementation
   }
   ```

2. **Value Objects for all domain concepts**:
   ```go
   type Bandwidth struct {
       value uint64
       unit  BandwidthUnit
   }
   ```

3. **Events are immutable** - use past tense naming:
   ```go
   type QdiscCreatedEvent struct {
       Device    valueobjects.Device
       Handle    valueobjects.Handle
       Type      string
       Timestamp time.Time
   }
   ```

4. **Netlink operations** require root for integration tests

5. **Follow the rules** in `rules/coding.md` and `rules/management.md`