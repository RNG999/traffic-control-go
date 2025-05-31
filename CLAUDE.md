# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a traffic control system implementation in Go. The project is currently in its initial setup phase with foundational documentation and development rules established.

## Development Commands

Since this is a Go project in early development, common commands would include:

```bash
# Initialize Go module (if not already done)
go mod init github.com/[username]/traffic-control-go

# Run the application
go run main.go

# Build the application
go build -o traffic-control

# Run tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Format code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run

# Run tests with coverage
go test -cover ./...

# Run tests with detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -run TestName ./...

# Benchmark tests
go test -bench=. ./...
```

## Project Rules

This project has specific rules defined in the `rules/` directory that must be followed:

### 1. General Coding Rules (`rules/001_general/`)
- Follow comprehensive software engineering guidelines including:
  - CQRS (Command Query Responsibility Segregation) principles
  - Event Sourcing fundamentals
  - Domain-Driven Design (DDD) patterns
  - Functional programming concepts
  - Test-Driven Development (TDD) practices
  - Immutability and defensive programming
  - Type-driven development workflow

### 2. Memory Management Rules (`rules/002_memory/`)
- Maintain a memory bank structure for project continuity
- Required core files in memory bank:
  - `projectbrief.md` - Foundation document
  - `productContext.md` - Why the project exists
  - `activeContext.md` - Current work focus
  - `systemPatterns.md` - Architecture and patterns
  - `techContext.md` - Technologies used
  - `progress.md` - Current status
- Update `.cursor/rules/*.mdc` with project-specific patterns

### 3. Documentation Rules (`rules/003_document/`)
- Keep root `README.md` updated with project overview
- Maintain `docs/README.md` as navigation hub for documentation
- Create feature-specific documentation in `docs/[feature]/README.md`
- Always update documentation when making significant changes

## Architecture Guidelines

Based on the coding rules, this traffic control system should follow:

1. **CQRS Pattern**: Separate command (write) and query (read) models
   - Commands in `internal/commands/` directory
   - Queries in `internal/queries/` directory
   - Shared domain events in `internal/domain/events/`

2. **Event Sourcing**: Store state changes as immutable events
   - All traffic control state changes recorded as events
   - Aggregate state reconstructed by replaying events
   - Event store for persistent storage

3. **DDD Principles**: Use Value Objects, Entities, and Aggregates
   - Value Objects: TrafficLightState, VehicleID, IntersectionID
   - Entities: Vehicle, TrafficLight
   - Aggregates: Intersection, TrafficFlow

4. **Functional Approach**: Prefer immutability and pure functions
   - Use `Result[T, E]` or similar for error handling
   - Avoid nil returns - use explicit optional types
   - No mutable shared state

5. **Type Safety**: Use Go's type system effectively with custom types
   - Define domain-specific types (not just strings/ints)
   - Use interfaces for abstractions
   - Leverage Go's type embedding where appropriate

## Key Reminders

1. **Initialize Messages**: When rules are loaded, acknowledge them in Japanese:
   - "001_generalを読み込みました！"
   - "002_memoryを読み込みました！" 
   - "003_documentを読み込みました！"

2. **Documentation First**: Always update relevant documentation when making changes

3. **Memory Bank**: If implementing features, create and maintain memory bank files

4. **Test Coverage**: Follow TDD practices - write tests before implementation

## Code Structure Guidelines

When implementing features in this traffic control system:

### Directory Structure
```
traffic-control-go/
├── cmd/                      # Application entry points
│   └── traffic-control/      # Main application
├── internal/                 # Private application code
│   ├── domain/              # Domain models and business logic
│   │   ├── aggregates/      # DDD Aggregates
│   │   ├── entities/        # DDD Entities
│   │   ├── events/          # Domain events
│   │   └── valueobjects/    # DDD Value Objects
│   ├── commands/            # CQRS Commands
│   │   ├── handlers/        # Command handlers
│   │   └── models/          # Command DTOs
│   ├── queries/             # CQRS Queries
│   │   ├── handlers/        # Query handlers
│   │   └── models/          # Query DTOs/Read models
│   ├── infrastructure/      # External concerns
│   │   ├── eventstore/      # Event persistence
│   │   └── persistence/     # Database adapters
│   └── application/         # Application services
├── pkg/                     # Public packages
│   └── types/               # Shared types (Result, Maybe, etc.)
├── memory-bank/             # Project memory files
└── test/                    # Integration tests
```

### Testing Requirements
- **ALWAYS** write tests first (TDD Red/Green/Refactor)
- Use table-driven tests for comprehensive coverage
- Test file naming: `*_test.go` in same package
- Use testify/assert for clearer test assertions

### Key Implementation Notes

1. **Never use primitive types directly for domain concepts**
   ```go
   // Bad
   type Vehicle struct {
       ID string
   }
   
   // Good
   type VehicleID string
   type Vehicle struct {
       ID VehicleID
   }
   ```

2. **Use Result types for operations that can fail**
   ```go
   type Result[T any] struct {
       value T
       err   error
   }
   ```

3. **Events must be immutable**
   ```go
   type TrafficLightChangedEvent struct {
       IntersectionID IntersectionID
       LightID        TrafficLightID
       OldState       TrafficLightState
       NewState       TrafficLightState
       Timestamp      time.Time
   }
   ```

## Next Steps

When starting development on this project:

1. Create the initial Go module structure
2. Set up the memory bank directory and core files
3. Define the domain model following DDD principles
4. Implement CQRS infrastructure
5. Set up event sourcing capabilities
6. Create comprehensive tests following TDD

## Important Warnings

- **DO NOT** create getters/setters for every field
- **DO NOT** use mutable shared state
- **DO NOT** use `interface{}` or `any` without strong justification
- **ALWAYS** handle errors explicitly (no ignored errors)
- **ALWAYS** validate inputs at boundaries
- **NEVER** expose internal state unnecessarily