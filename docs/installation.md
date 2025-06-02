# Installation Guide

## Prerequisites

- Linux system with kernel 2.6.32 or later
- Go 1.21 or later
- Root privileges or CAP_NET_ADMIN capability for traffic control operations
- `tc` utility from `iproute2` package (usually pre-installed)
- `iperf3` (optional, for running integration tests)

## System Requirements

### Required Kernel Modules
The following kernel modules must be available for traffic control operations:
- `sch_htb` - HTB (Hierarchical Token Bucket) qdisc
- `sch_tbf` - TBF (Token Bucket Filter) qdisc  
- `sch_prio` - PRIO (Priority) qdisc
- `sch_fq_codel` - FQ_CODEL (Fair Queue CoDel) qdisc
- `cls_u32` - U32 packet classifier

Check module availability:
```bash
# Check if modules are loaded
lsmod | grep -E "sch_htb|sch_tbf|sch_prio|sch_fq_codel|cls_u32"

# Load modules if not available
sudo modprobe sch_htb sch_tbf sch_prio sch_fq_codel cls_u32
```

### Required Capabilities
The library requires the following Linux capabilities when performing traffic control operations:
- `CAP_NET_ADMIN` - Network administration operations

## Library Installation

### Install as Go Module

```bash
# Add to your Go project
go get github.com/rng999/traffic-control-go

# Import in your code
import "github.com/rng999/traffic-control-go/api"
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/rng999/traffic-control-go.git
cd traffic-control-go

# Run tests to verify functionality
make test-unit          # Unit tests (no root required)
sudo make test-integration  # Integration tests (requires root)

# Install dependencies
go mod download
```

## Usage in Your Application

### Basic Example

```go
package main

import (
    "log"
    "github.com/rng999/traffic-control-go/api"
)

func main() {
    // Create traffic controller with Improved API
    tc := api.NewImproved("eth0").
        TotalBandwidth("1Gbps")
    
    // Configure traffic class
    tc.Class("Web Traffic").
        Guaranteed("100Mbps").
        BurstTo("200Mbps").
        Priority(4).
        Ports(80, 443)
    
    // Apply configuration (requires root/CAP_NET_ADMIN)
    if err := tc.Apply(); err != nil {
        log.Fatalf("Failed to apply traffic control: %v", err)
    }
}
```

### Running Your Application

```bash
# Run with root privileges
sudo go run main.go

# Or build and run
go build -o myapp
sudo ./myapp

# Or set capabilities (alternative to sudo)
sudo setcap cap_net_admin=+ep ./myapp
./myapp
```

## Verification

### Check Applied Configuration
```bash
# Show qdisc configuration
tc qdisc show dev eth0

# Show class configuration
tc class show dev eth0

# Show filter configuration
tc filter show dev eth0

# Monitor statistics in real-time
watch -n 1 'tc -s qdisc show dev eth0'
```

## Troubleshooting

### Common Issues

#### Permission Denied
```bash
Error: operation not permitted
```
**Solution**: Ensure your application runs with root privileges or CAP_NET_ADMIN capability.

#### Module Not Found
```bash
Error: Unknown qdisc "htb"
```
**Solution**: Load the required kernel module:
```bash
sudo modprobe sch_htb
```

#### Interface Not Found
```bash
Error: Cannot find device "eth0"
```
**Solution**: Check available interfaces:
```bash
ip link show
```

#### Configuration Conflicts
```bash
Error: RTNETLINK answers: File exists
```
**Solution**: The library handles cleanup automatically, but you can manually clear existing configuration:
```bash
sudo tc qdisc del dev eth0 root
```

### Debug Mode
Enable verbose logging in your application:
```go
import "github.com/rng999/traffic-control-go/pkg/logging"

// Initialize logging
logging.InitializeDevelopment()

// Or with custom configuration
logging.Initialize(&logging.Config{
    Level:  "debug",
    Format: "console",
    Outputs: []string{"stdout"},
})
```

### Integration with systemd
If running as a service:
```ini
[Unit]
Description=My Traffic Control Application
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/myapp
Restart=on-failure
RestartSec=5

# Grant network admin capability
AmbientCapabilities=CAP_NET_ADMIN
CapabilityBoundingSet=CAP_NET_ADMIN

[Install]
WantedBy=multi-user.target
```

## Security Considerations

### Network Access Control
- Only grant CAP_NET_ADMIN capability to trusted applications
- Consider using dedicated service accounts
- Monitor traffic control changes in production environments

### Best Practices
- Always validate configuration before applying
- Implement proper error handling and rollback mechanisms
- Test configurations in isolated environments first
- Use configuration management for reproducible setups

### Container Environments
When using in containers:
```bash
# Docker example
docker run --cap-add=NET_ADMIN --network=host myapp

# Kubernetes example
securityContext:
  capabilities:
    add:
    - NET_ADMIN
```

## Next Steps

- Review the [API Documentation](../memory-bank/api-design.md) for detailed API usage
- Check the [Examples](../examples/) directory for working code samples
- Read about [Traffic Control Basics](traffic-control.md) to understand TC concepts
- See [Priority Guide](priority-guide.md) for traffic prioritization strategies