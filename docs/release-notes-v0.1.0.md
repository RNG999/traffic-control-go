# Release Notes - v0.1.0

## 🎉 First Feature-Complete Release!

We're excited to announce the first official release of traffic-control-go, a human-readable Go library for Linux Traffic Control.

### ✨ Features

#### Core Library
- **Human-Readable API**: Intuitive method chaining for traffic control operations
- **Type-Safe Design**: Leverages Go's type system to prevent configuration errors  
- **Event-Driven Architecture**: Built with CQRS and Event Sourcing patterns
- **Multiple Qdisc Support**: HTB, TBF, PRIO, and FQ_CODEL fully implemented
- **Statistics Collection**: Real-time traffic monitoring and metrics

#### Storage & Persistence
- **SQLite Event Store**: Persistent event storage with full history
- **Memory Store**: In-memory option for testing and development
- **Event Replay**: Reconstruct state from event history

#### Command-Line Tools
- **traffic-control**: Production-ready standalone binary
- **tcctl**: Demonstration and testing CLI tool
- Both tools support all qdisc types and configuration options

#### Testing & Quality
- **Comprehensive Unit Tests**: All components thoroughly tested
- **Integration Tests**: Real traffic control testing with iperf3
- **CI/CD Pipeline**: Automated testing on Go 1.21, 1.22, and 1.23
- **Code Coverage**: Detailed coverage reports with Codecov

#### Developer Experience
- **Structured Logging**: Built on Uber's Zap for high-performance logging
- **Configuration Support**: YAML and JSON configuration files
- **Priority System**: Numeric priority system (0-7) for traffic classes
- **Makefile-based Build**: No external build scripts required

### 📦 Installation

#### Download Pre-built Binaries
```bash
# Linux amd64
curl -L https://github.com/rng999/traffic-control-go/releases/download/v0.1.0/traffic-control-linux-amd64.tar.gz | tar -xz

# macOS amd64
curl -L https://github.com/rng999/traffic-control-go/releases/download/v0.1.0/traffic-control-darwin-amd64.tar.gz | tar -xz
```

#### Build from Source
```bash
git clone https://github.com/rng999/traffic-control-go.git
cd traffic-control-go
make build
sudo make install
```

#### Go Install
```bash
go install github.com/rng999/traffic-control-go/cmd/traffic-control@v0.1.0
go install github.com/rng999/traffic-control-go/cmd/tcctl@v0.1.0
```

### 🔧 Requirements

- Linux kernel 2.6.32 or later with TC support
- Go 1.21 or later (for building from source)
- Root privileges or CAP_NET_ADMIN capability
- iperf3 (for running integration tests)

### 📖 Documentation

- [Installation Guide](installation.md)
- [API Documentation](../memory-bank/api-design.md)
- [Testing Guide](testing.md)
- [Standalone Binary Usage](standalone-binary.md)
- [Examples](../examples/)

### 🚀 Quick Start

```go
// Library usage
import tc "github.com/rng999/traffic-control-go"

controller := tc.New("eth0").
    SetTotalBandwidth("1Gbps")

err := controller.
    CreateTrafficClass("database").
        WithGuaranteedBandwidth("100Mbps").
        WithMaxBandwidth("200Mbps").
        ForDestination("192.168.1.10").
    Apply()
```

```bash
# CLI usage
sudo traffic-control tbf eth0 1:0 100Mbps
sudo traffic-control stats eth0
```

### 📝 Notes

- This is a pre-1.0 release following semantic versioning
- API is stabilizing but may have minor changes before 1.0
- Production use is encouraged with appropriate testing
- Feedback and bug reports are welcome!

### 🙏 Acknowledgments

Thanks to all contributors who helped make this release possible!

### 📄 License

Apache License 2.0

---

For more information, visit the [project repository](https://github.com/rng999/traffic-control-go).