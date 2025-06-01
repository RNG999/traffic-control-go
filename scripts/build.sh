#!/bin/bash

# Build script for Traffic Control Go

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "=== Building Traffic Control Go ==="
echo

# Build information
VERSION="${VERSION:-0.1.1}"
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S UTC')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X 'main.version=$VERSION'"
LDFLAGS="$LDFLAGS -X 'main.buildTime=$BUILD_TIME'"
LDFLAGS="$LDFLAGS -X 'main.gitCommit=$GIT_COMMIT'"

echo "Version: $VERSION"
echo "Build Time: $BUILD_TIME"
echo "Git Commit: $GIT_COMMIT"
echo

# Create bin directory
mkdir -p bin

# Build the main binary
echo "Building traffic-control binary..."
go build -ldflags "$LDFLAGS" -o bin/traffic-control ./cmd/traffic-control/

# Build the CLI demo (tcctl)
echo "Building tcctl demo binary..."
go build -ldflags "$LDFLAGS" -o bin/tcctl ./cmd/tcctl/

echo
echo "=== Build completed successfully! ==="
echo
echo "Binaries created:"
echo "  - bin/traffic-control  (main binary)"
echo "  - bin/tcctl           (demo/testing tool)"
echo
echo "Usage examples:"
echo "  sudo ./bin/traffic-control htb eth0 1:0 1:999"
echo "  sudo ./bin/traffic-control tbf eth0 1:0 100Mbps"
echo "  sudo ./bin/traffic-control stats eth0"
echo
echo "Note: Root privileges (sudo) are required for network interface modifications."