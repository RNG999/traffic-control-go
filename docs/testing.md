# Testing Guide

This document describes the testing strategy and how to run tests for the Traffic Control Go project.

## Test Categories

### Unit Tests
Standard Go unit tests that don't require special privileges or external dependencies.

```bash
make test
```

### Integration Tests
Tests that verify Traffic Control functionality with real network interfaces. These tests require:
- Root privileges (sudo)
- iperf3 installed
- Linux environment with tc support

```bash
# Install iperf3 (Ubuntu/Debian)
sudo apt-get install iperf3

# Install iperf3 (RHEL/CentOS)
sudo yum install iperf3

# Run integration tests
make test-integration
```

## Integration Test Details

### Bandwidth Limiting Tests (`iperf_bandwidth_test.go`)
These tests verify that bandwidth limits are correctly applied:

1. **Basic Bandwidth Limits**: Tests various bandwidth limits (1Mbps, 5Mbps, 10Mbps)
2. **Priority Classes**: Tests traffic prioritization between classes
3. **Burst Traffic**: Tests token bucket burst behavior

The tests use iperf3 to measure actual bandwidth and verify it matches the configured limits within a reasonable tolerance.

### Virtual Ethernet Tests (`veth_iperf_test.go`)
These tests use virtual ethernet pairs (veth) to create isolated network environments:

1. **Veth Pair Testing**: Creates virtual interfaces to test TC in isolation
2. **Network Namespace**: Uses Linux network namespaces for complete isolation
3. **Realistic Scenarios**: Tests TC behavior in more realistic network conditions

## Running Specific Tests

```bash
# Run only bandwidth tests
sudo go test -v -tags=integration -run TestTrafficControlWithIperf ./test/integration/

# Run only veth tests
sudo go test -v -tags=integration -run TestTrafficControlWithVethPair ./test/integration/

# Run with verbose output
sudo go test -v -tags=integration ./test/integration/
```

## Test Environment Setup

### Docker Environment
For consistent testing, you can use a Docker container:

```dockerfile
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    iproute2 \
    iperf3 \
    golang-go \
    sudo \
    git

WORKDIR /app
COPY . .

# Run tests
CMD ["make", "test-integration"]
```

### CI/CD Integration
Integration tests can be run in CI with appropriate permissions:

```yaml
- name: Run Integration Tests
  run: |
    sudo apt-get update
    sudo apt-get install -y iperf3
    make test-integration
```

## Writing New Integration Tests

When writing new integration tests:

1. **Use Build Tags**: Add `//go:build integration` to exclude from regular test runs
2. **Check Prerequisites**: Skip tests if requirements aren't met (root, iperf3, etc.)
3. **Clean Up**: Always clean up TC rules and network interfaces
4. **Use Tolerances**: Network tests can be flaky, use reasonable tolerances
5. **Log Output**: Provide detailed logging for debugging failures

Example:
```go
//go:build integration
// +build integration

func TestNewFeature(t *testing.T) {
    if os.Geteuid() != 0 {
        t.Skip("Test requires root privileges")
    }
    
    // Clean up before and after
    defer cleanupTC(t, device)
    
    // Your test here...
}
```

## Troubleshooting

### Common Issues

1. **Permission Denied**: Ensure you're running with sudo
2. **iperf3 Not Found**: Install iperf3 package
3. **TC Command Failed**: Check if kernel has TC support
4. **Flaky Tests**: Increase tolerances or run on isolated system

### Debugging

```bash
# Check current TC configuration
tc qdisc show
tc class show
tc filter show

# Monitor network traffic
tcpdump -i lo -n

# Check iperf3 connectivity
iperf3 -c 127.0.0.1 -p 5201
```