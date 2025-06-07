# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go library for Linux Traffic Control (TC) that provides an intuitive, human-readable API for managing network traffic. The library follows CQRS, Event Sourcing, and Domain-Driven Design principles.

## Important: Project Rules

**ALWAYS refer to and follow the rules defined in the `rules/` directory:**
- `@rules/coding.md` - Comprehensive coding guidelines (CQRS, DDD, FP, testing, etc.)
- `@rules/management.md` - Project management guidelines using GitHub

**When you receive a request for a coding task (e.g., implementing a feature, fixing a bug), you MUST consult `@rules/coding.md` and implement the solution strictly according to its guidelines.**

**When you receive a request for a management task (e.g., task management), you MUST consult `@rules/management.md` and execute the task strictly according to its defined procedures.**

These rules must be consulted and strictly followed for any coding or project management tasks.

## Issue and PR Workflow

The following rules govern the lifecycle of issues and Pull Requests. This workflow ensures that all work is tracked and properly linked.

- **1. Issue-First Principle**:
    - All work must start with an Issue. Before creating a Pull Request (PR), ensure a corresponding Issue exists.
    - If a relevant Issue does not exist, you **MUST** create one first that describes the feature or bug.

- **2. One-to-One Correspondence**:
    - Each Issue **MUST** correspond to a single Pull Request, and each PR **MUST** resolve a single Issue.
    - This one-to-one mapping is strict. Do not bundle fixes for multiple, unrelated Issues into a single PR.

- **3. Linking PRs to Issues**:
    - When you create a PR, you **MUST** reference its corresponding Issue number in the PR's description (e.g., `Fixes #123`, `Closes #123`). This creates a formal link.

- **4. Closing an Issue on Merge**:
    - An Issue should be closed once its single, linked PR has been merged.
    - To identify which Issue to close, you **MUST** find the Issue number referenced in the PR's description or comments.
    - The trigger for closing the Issue can be either your own confirmation of the merge or a notification from the user.

- **5. Re-opening an Issue on Failure**:
    - A previously closed Issue **MUST** be re-opened if any negative factors arise after the PR merge. This includes situations such as:
        - The CI/CD pipeline triggered by the merge fails.
        - Any other regressions or problems are discovered as a result of the merge.

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
