# Traffic Control Go - Standalone Binary

## Overview

The `traffic-control` binary is a complete, standalone command-line tool for managing Linux Traffic Control (tc) configurations. It provides a user-friendly interface to the powerful traffic-control-go library, supporting multiple queueing disciplines (qdiscs) including HTB, TBF, PRIO, and FQ_CODEL.

## Key Features

- **Multiple Qdisc Support**: HTB, TBF, PRIO, and FQ_CODEL
- **Intuitive CLI**: Simple, consistent command structure
- **Rich Statistics**: Real-time and historical traffic statistics
- **Cross-Platform**: Builds for Linux, macOS, and Windows
- **Event Sourcing**: Full configuration history and rollback capability
- **Production Ready**: Comprehensive error handling and validation

## Quick Start

### Installation

```bash
# Download and install
curl -L https://github.com/YOUR_ORG/traffic-control-go/releases/latest/download/traffic-control-linux-amd64.tar.gz | tar -xz
sudo cp traffic-control /usr/local/bin/

# Or build from source
git clone https://github.com/YOUR_ORG/traffic-control-go.git
cd traffic-control-go
make install
```

### Basic Usage

```bash
# Simple bandwidth limiting (requires sudo)
sudo traffic-control tbf eth0 1:0 100Mbps

# Show configuration and statistics
sudo traffic-control stats eth0

# Get help for any command
traffic-control help
traffic-control tbf --help
```

## Supported Queueing Disciplines

### 1. HTB (Hierarchical Token Bucket)

Best for: Complex hierarchical bandwidth allocation with borrowing

```bash
# Create HTB qdisc with multiple classes
sudo traffic-control htb eth0 1:0 1:999 \
    --class 1:10,parent=1:,rate=60Mbps,ceil=80Mbps \
    --class 1:20,parent=1:,rate=30Mbps,ceil=50Mbps \
    --class 1:30,parent=1:,rate=10Mbps,ceil=20Mbps
```

**Features:**
- Hierarchical bandwidth allocation
- Rate guarantees with ceiling limits
- Bandwidth borrowing between classes
- Priority-based scheduling

**Use Cases:**
- ISP customer management
- Enterprise network QoS
- Multi-tenant environments

### 2. TBF (Token Bucket Filter)

Best for: Simple, effective rate limiting

```bash
# Basic rate limiting
sudo traffic-control tbf eth0 1:0 100Mbps

# Advanced configuration
sudo traffic-control tbf eth0 1:0 50Mbps \
    --buffer 65536 \
    --limit 20000 \
    --burst 8192
```

**Features:**
- Token bucket algorithm
- Configurable burst handling
- Low latency for compliant traffic
- Simple configuration

**Use Cases:**
- WAN bandwidth limiting
- Simple rate limiting
- Ingress policing

### 3. PRIO (Priority Scheduler)

Best for: Simple priority-based traffic separation

```bash
# 3-band priority scheduler
sudo traffic-control prio eth0 1:0 3

# Custom priority mapping
sudo traffic-control prio eth0 1:0 4 \
    --priomap 0,1,1,1,2,2,3,3,1,1,1,1,1,1,1,1
```

**Features:**
- Multiple priority bands
- Strict priority scheduling
- Configurable priority mapping
- Low overhead

**Use Cases:**
- VoIP prioritization
- Real-time traffic separation
- Simple QoS implementations

### 4. FQ_CODEL (Fair Queue CoDel)

Best for: Fair queuing with active queue management

```bash
# Default fair queuing with CoDel AQM
sudo traffic-control fq_codel eth0 1:0

# Low-latency configuration
sudo traffic-control fq_codel eth0 1:0 \
    --limit 20480 \
    --flows 2048 \
    --target 1000 \
    --interval 50000 \
    --ecn
```

**Features:**
- Fair per-flow queuing
- CoDel active queue management
- ECN support
- Automatic latency control

**Use Cases:**
- Data center networking
- Low-latency applications
- Mixed traffic environments
- Modern internet protocols

## Statistics and Monitoring

### Device Statistics
```bash
# Overall device statistics
sudo traffic-control stats eth0

# Real-time statistics
sudo traffic-control stats eth0 --realtime

# Continuous monitoring (5-second intervals)
sudo traffic-control stats eth0 --monitor 5
```

### Component-Specific Statistics
```bash
# Qdisc statistics
sudo traffic-control stats eth0 --qdisc 1:0

# Class statistics  
sudo traffic-control stats eth0 --class 1:10

# Filter statistics
sudo traffic-control stats eth0 --filter 1:
```

## Configuration Examples

### Home Internet Gateway
```bash
# Simple bandwidth limiting for home network
sudo traffic-control tbf eth0 1:0 100Mbps --buffer 32768

# Priority for interactive traffic
sudo traffic-control prio eth1 1:0 3
```

### Enterprise Network
```bash
# Hierarchical QoS for department bandwidth allocation
sudo traffic-control htb eth0 1:0 1:999 \
    --class 1:1,parent=1:,rate=800Mbps,ceil=1000Mbps \
    --class 1:10,parent=1:1,rate=400Mbps,ceil=600Mbps \
    --class 1:20,parent=1:1,rate=200Mbps,ceil=300Mbps \
    --class 1:30,parent=1:1,rate=100Mbps,ceil=200Mbps
```

### Data Center
```bash
# Low-latency fair queuing for microservices
sudo traffic-control fq_codel eth0 1:0 \
    --limit 20480 \
    --flows 2048 \
    --target 1000 \
    --ecn
```

### ISP Customer Management
```bash
# Per-customer rate limiting
sudo traffic-control tbf eth0.100 1:0 50Mbps   # Basic plan
sudo traffic-control tbf eth0.200 1:0 100Mbps  # Standard plan  
sudo traffic-control tbf eth0.300 1:0 200Mbps  # Premium plan
```

## Advanced Features

### Event Sourcing
All configuration changes are stored as events, enabling:
- Complete configuration history
- Point-in-time recovery
- Audit trails
- Rollback capabilities

### CQRS Architecture
Separate command and query paths provide:
- Optimized read performance
- Scalable write operations
- Flexible query models
- Future extensibility

### Integration Options
- **REST API**: Programmatic configuration management
- **Event Streaming**: Real-time configuration monitoring
- **Metrics Export**: Prometheus/Grafana integration
- **Configuration Management**: Ansible/Terraform support

## Troubleshooting

### Common Issues

#### Permission Denied
```bash
Error: operation not permitted
```
**Solution**: Run with sudo or ensure CAP_NET_ADMIN capability

#### Interface Not Found
```bash
Error: Cannot find device "eth0"
```
**Solution**: Check available interfaces with `ip link show`

#### Kernel Module Missing
```bash
Error: Unknown qdisc "htb"
```
**Solution**: Load required kernel module: `sudo modprobe sch_htb`

#### Configuration Conflicts
```bash
Error: RTNETLINK answers: File exists
```
**Solution**: Remove existing configuration: `sudo tc qdisc del dev eth0 root`

### Debug Mode
```bash
# Enable detailed logging
export TC_DEBUG=1
sudo traffic-control tbf eth0 1:0 100Mbps
```

### Verification Commands
```bash
# Check applied configuration
tc qdisc show dev eth0
tc class show dev eth0
tc filter show dev eth0

# Monitor statistics
watch -n 1 'tc -s qdisc show dev eth0'
```

## Performance Considerations

### Resource Usage
- **Memory**: ~10-50MB depending on configuration size
- **CPU**: Minimal overhead for most configurations
- **Disk**: Event store grows with configuration changes

### Scaling Limits
- **Interfaces**: No practical limit
- **Classes**: Up to 65536 per qdisc
- **Filters**: Up to 65536 per qdisc
- **Rules**: Limited by available memory

### Optimization Tips
1. **Use appropriate qdisc for workload**
   - HTB for complex hierarchies
   - TBF for simple rate limiting
   - PRIO for strict priorities
   - FQ_CODEL for fair queuing

2. **Monitor performance impact**
   - Use statistics to verify effectiveness
   - Adjust parameters based on traffic patterns
   - Consider hardware offloading where available

3. **Plan capacity appropriately**
   - Account for burst traffic
   - Leave headroom for growth
   - Monitor utilization trends

## Security Considerations

### Network Access Control
- Requires root privileges or CAP_NET_ADMIN
- Consider dedicated service accounts
- Use sudo rules for specific users
- Monitor configuration changes

### Audit and Compliance
- All changes are logged in event store
- Configuration history is immutable
- Timestamps and user tracking available
- Integration with SIEM systems possible

## Next Steps

1. **Read the API Documentation** for programmatic usage
2. **Check Example Configurations** for common scenarios  
3. **Review Performance Tuning Guide** for optimization
4. **Join Community Discussions** for support and feedback

## Related Tools

- **tcctl**: Demo/testing CLI (included in distribution)
- **tc**: Traditional Linux traffic control utility
- **iptables**: Packet filtering and NAT
- **ipset**: Efficient IP set management
- **nftables**: Modern packet filtering framework