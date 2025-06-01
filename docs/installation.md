# Installation Guide

## Prerequisites

- Linux system with kernel 2.6.32 or later
- Go 1.19 or later (for building from source)
- Root privileges for network interface modifications
- `tc` utility from `iproute2` package (usually pre-installed)

## System Requirements

### Required Kernel Modules
The following kernel modules must be available:
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
The application requires the following Linux capabilities:
- `CAP_NET_ADMIN` - Network administration operations

## Installation Methods

### Method 1: Download Pre-built Binary

```bash
# Download the latest release (currently v0.1.0)
wget https://github.com/YOUR_ORG/traffic-control-go/releases/latest/download/traffic-control-linux-amd64.tar.gz

# Extract
tar -xzf traffic-control-linux-amd64.tar.gz

# Install to system path
sudo cp traffic-control /usr/local/bin/

# Verify installation
traffic-control version
```

### Method 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/YOUR_ORG/traffic-control-go.git
cd traffic-control-go

# Build using the provided script
./scripts/build.sh

# Install to system path
sudo cp bin/traffic-control /usr/local/bin/

# Verify installation
traffic-control version
```

### Method 3: Install with Go

```bash
go install github.com/YOUR_ORG/traffic-control-go/cmd/traffic-control@latest
```

## Post-Installation Setup

### 1. Verify Permissions
Ensure the binary has proper permissions:
```bash
sudo chown root:root /usr/local/bin/traffic-control
sudo chmod 755 /usr/local/bin/traffic-control
```

### 2. Set up Sudoers (Optional)
To allow specific users to run traffic-control without full sudo access:

```bash
# Create a sudoers rule
echo "%netadmin ALL=(root) NOPASSWD: /usr/local/bin/traffic-control" | sudo tee /etc/sudoers.d/traffic-control

# Add users to the netadmin group
sudo groupadd netadmin
sudo usermod -a -G netadmin $USER
```

### 3. Bash Completion (Optional)
```bash
# Generate bash completion
traffic-control completion bash | sudo tee /etc/bash_completion.d/traffic-control
```

## Quick Start

### Basic Usage
```bash
# Show help
traffic-control help

# Configure simple rate limiting (requires sudo)
sudo traffic-control tbf eth0 1:0 100Mbps

# Show interface statistics
sudo traffic-control stats eth0

# Configure HTB with multiple classes
sudo traffic-control htb eth0 1:0 1:999 \
    --class 1:10,parent=1:,rate=60Mbps,ceil=80Mbps \
    --class 1:20,parent=1:,rate=30Mbps,ceil=50Mbps
```

### Verification
```bash
# Check if configuration was applied
tc qdisc show dev eth0
tc class show dev eth0
tc filter show dev eth0

# Monitor in real-time
watch -n 1 'tc -s qdisc show dev eth0'
```

## Troubleshooting

### Common Issues

#### Permission Denied
```bash
Error: operation not permitted
```
**Solution**: Run with sudo or ensure proper capabilities are set.

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
**Solution**: Remove existing qdisc configuration:
```bash
sudo tc qdisc del dev eth0 root
```

### Debug Mode
Enable verbose logging:
```bash
export TC_DEBUG=1
sudo traffic-control tbf eth0 1:0 100Mbps
```

### Log Files
Check system logs for detailed error information:
```bash
sudo journalctl -u traffic-control
dmesg | grep -i traffic
```

## Uninstallation

```bash
# Remove binary
sudo rm /usr/local/bin/traffic-control

# Remove sudoers rule (if created)
sudo rm /etc/sudoers.d/traffic-control

# Remove bash completion (if installed)
sudo rm /etc/bash_completion.d/traffic-control

# Clean up any remaining traffic control rules (if needed)
sudo tc qdisc del dev eth0 root 2>/dev/null || true
```

## Security Considerations

### Network Access Control
- Only grant network administration privileges to trusted users
- Consider using dedicated service accounts for automated scripts
- Monitor traffic control changes in production environments

### Audit Trail
- Enable audit logging for network configuration changes
- Use configuration management tools for reproducible setups
- Document all traffic control policies and their purposes

### Network Isolation
- Test configurations in isolated environments first
- Use network namespaces for development and testing
- Implement rollback procedures for production changes

## Next Steps

- Read the [User Guide](user-guide.md) for detailed usage instructions
- Check the [Examples](examples.md) for common configuration patterns
- Review the [API Documentation](api.md) for programmatic usage
- Join the community discussions for support and feature requests