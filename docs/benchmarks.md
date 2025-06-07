# Performance Benchmarks

This document describes the comprehensive benchmark suite for the Traffic Control Go library, designed to measure and monitor the performance of core operations.

## Overview

The benchmark suite covers the most performance-critical components of the library:

- **Value Objects** (`pkg/tc/`): Bandwidth, Handle, and Device operations
- **Event Store** (`internal/infrastructure/eventstore/`): Event sourcing performance
- **API Layer** (`api/`): Human-readable API performance
- **Command Handlers** (planned): CQRS command processing
- **Netlink Operations** (planned): Kernel communication overhead

## Running Benchmarks

### Quick Start

```bash
# Run all benchmarks
make bench

# Run benchmarks with memory allocation stats
make bench-report

# Compare performance across multiple runs
make bench-compare

# Generate CPU and memory profiles
make bench-profile
```

### Targeted Benchmarks

```bash
# Value objects only (fastest)
make bench-value-objects

# Event store operations
make bench-eventstore

# API layer performance
make bench-api
```

### Manual Execution

```bash
# Run specific benchmark patterns
go test -bench=BenchmarkBandwidthParsing ./pkg/tc/
go test -bench=BenchmarkEventStore ./internal/infrastructure/eventstore/
go test -bench=BenchmarkAPI ./api/

# With memory allocation tracking
go test -bench=. -benchmem ./pkg/tc/

# Multiple iterations for statistical significance
go test -bench=. -count=10 ./pkg/tc/

# Custom benchmark time
go test -bench=. -benchtime=10s ./pkg/tc/
```

## Benchmark Categories

### 1. Value Object Performance

**Location**: `pkg/tc/*_test.go`

#### Bandwidth Operations
- `BenchmarkBandwidthCreation`: Direct bandwidth object creation
- `BenchmarkBandwidthParsing`: String parsing performance
- `BenchmarkBandwidthArithmetic`: Mathematical operations
- `BenchmarkBandwidthFormatting`: String formatting performance
- `BenchmarkBandwidthComparisons`: Equality and comparison operations

#### Handle Operations
- `BenchmarkHandleCreation`: Handle object creation
- `BenchmarkHandleParsing`: String to handle conversion
- `BenchmarkHandleConversions`: Uint32 conversions
- `BenchmarkHandleFormatting`: String representation

#### Device Operations
- `BenchmarkDeviceNameCreation`: Device name validation
- `BenchmarkDeviceNameValidation`: Input validation performance
- `BenchmarkDeviceNameOperations`: String and equality operations

### 2. Event Store Performance

**Location**: `internal/infrastructure/eventstore/benchmark_test.go`

#### Core Operations
- `BenchmarkEventStoreSave`: Event persistence performance
- `BenchmarkEventStoreRetrieve`: Event retrieval by aggregate
- `BenchmarkEventStoreRetrieveFromVersion`: Partial event history
- `BenchmarkEventStoreGetAllEvents`: Full event history

#### Concurrency Tests
- `BenchmarkEventStoreConcurrentReads`: Parallel read operations
- `BenchmarkEventStoreConcurrentWrites`: Parallel write operations

#### Implementation Comparison
- `BenchmarkEventStoreComparison`: Memory vs SQLite performance
- `BenchmarkEventSerialization`: JSON marshaling performance

### 3. API Layer Performance

**Location**: `api/benchmark_test.go`

#### Core API Operations
- `BenchmarkNetworkInterfaceCreation`: Controller instantiation
- `BenchmarkTrafficControllerConfiguration`: Configuration methods
- `BenchmarkTrafficClassBuilder`: Fluent API performance

#### Complex Scenarios
- `BenchmarkCompleteTrafficSetup`: End-to-end configuration
- `BenchmarkConfigFromYAML`: Configuration file loading
- `BenchmarkTrafficControlService`: Service layer operations

#### Filter Performance
- `BenchmarkFilterGeneration`: Traffic filter creation
- `BenchmarkPriorityToHandle`: Priority mapping performance

## Performance Expectations

### Value Objects (Excellent Performance)
```
BenchmarkBandwidthCreation/Mbps-4         1000000000    0.22 ns/op
BenchmarkBandwidthParsingSimple-4              8208   14821 ns/op
BenchmarkHandleCreation/NewHandle-4       1000000000    0.22 ns/op
BenchmarkHandleParsing/ParseHandle-4           5000  250000 ns/op
```

**Key Insights:**
- Direct object creation is extremely fast (~0.2ns)
- String parsing is ~70,000x slower but still efficient for typical usage
- Prefer direct creation over parsing in performance-critical code

### Event Store Performance
```
BenchmarkEventStoreSave/MemoryEventStore-4      500000   3000 ns/op
BenchmarkEventStoreSave/SQLiteEventStore-4        1000 1200000 ns/op
BenchmarkEventStoreRetrieve/Memory-4           100000  15000 ns/op
BenchmarkEventStoreRetrieve/SQLite-4             5000 280000 ns/op
```

**Key Insights:**
- Memory store is ~400x faster for writes, ~20x faster for reads
- SQLite provides persistence at significant performance cost
- Consider memory store for high-throughput scenarios with periodic snapshots

### API Layer Performance
```
BenchmarkNetworkInterfaceCreation-4           100000  12000 ns/op
BenchmarkCompleteTrafficSetup/Basic-4           5000 250000 ns/op
BenchmarkTrafficClassBuilder-4                50000  30000 ns/op
```

**Key Insights:**
- API layer adds reasonable overhead for developer convenience
- Complex configurations scale linearly with traffic class count
- Fluent API performance is acceptable for typical usage patterns

## Performance Monitoring

### Regression Detection

Run benchmarks before and after changes:

```bash
# Before changes
go test -bench=. -count=5 ./... > before.txt

# After changes  
go test -bench=. -count=5 ./... > after.txt

# Compare results
benchcmp before.txt after.txt
```

### Continuous Integration

The benchmark suite is integrated into CI/CD pipelines:

```yaml
# .github/workflows/benchmark.yml
- name: Run benchmarks
  run: |
    go test -bench=. -benchmem ./... > benchmark-results.txt
    # Upload results for performance tracking
```

### Memory Profiling

Identify memory allocation hotspots:

```bash
# Generate memory profile
go test -bench=BenchmarkEventStore -memprofile=mem.prof ./internal/infrastructure/eventstore/

# Analyze profile
go tool pprof mem.prof
```

### CPU Profiling

Identify CPU bottlenecks:

```bash
# Generate CPU profile
go test -bench=BenchmarkBandwidthParsing -cpuprofile=cpu.prof ./pkg/tc/

# Analyze profile
go tool pprof cpu.prof
```

## Optimization Guidelines

### Value Objects
1. **Use direct creation** over string parsing when possible
2. **Cache parsed values** for repeated operations
3. **Prefer immutable operations** over mutable state

### Event Store
1. **Use Memory store** for high-throughput scenarios
2. **Batch operations** when possible for SQLite
3. **Consider periodic snapshots** to reduce event history size

### API Layer
1. **Build configurations once** and reuse when possible
2. **Avoid repeated parsing** of bandwidth strings
3. **Use batch operations** for multiple traffic classes

## Benchmark Maintenance

### Adding New Benchmarks

1. **Follow naming convention**: `BenchmarkComponentOperation`
2. **Use table-driven tests** for multiple scenarios
3. **Include memory benchmarks** with `-benchmem`
4. **Reset timer** after setup: `b.ResetTimer()`

Example:
```go
func BenchmarkNewOperation(b *testing.B) {
    // Setup code here
    testData := setupTestData()
    
    b.ResetTimer() // Reset timer after setup
    for i := 0; i < b.N; i++ {
        // Operation being benchmarked
        result := performOperation(testData)
        _ = result // Prevent compiler optimization
    }
}
```

### Benchmark Best Practices

1. **Avoid timer manipulation** except for setup/teardown
2. **Prevent compiler optimization** by using results
3. **Use realistic test data** representative of actual usage
4. **Include both single and batch operations**
5. **Test with various input sizes** to identify scaling behavior

## Integration with Development Workflow

### Pre-commit Hooks
```bash
# Run quick benchmarks before commit
go test -bench=BenchmarkBandwidth -run=^$ ./pkg/tc/ -benchtime=100ms
```

### Performance Reviews
- Include benchmark results in pull requests for performance-sensitive changes
- Compare results against baseline measurements
- Document any intentional performance trade-offs

### Release Criteria
- No significant performance regressions (>10% slowdown)
- New features include appropriate benchmarks
- Performance improvements documented with before/after results

## Future Enhancements

### Planned Benchmark Coverage
1. **Command Handler Performance**: CQRS command processing benchmarks
2. **Netlink Operations**: Kernel communication overhead measurement
3. **Statistics Collection**: Real-time monitoring performance
4. **Concurrent Operations**: Multi-threaded traffic control scenarios

### Advanced Profiling
1. **Flame graphs** for visual performance analysis
2. **Trace analysis** for understanding execution flow
3. **Memory allocation tracking** over time
4. **Performance regression alerts** in CI/CD

The benchmark suite provides comprehensive performance visibility into the Traffic Control Go library, enabling data-driven optimization decisions and preventing performance regressions.