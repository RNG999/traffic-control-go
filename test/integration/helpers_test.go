//go:build integration
// +build integration

package integration_test

import (
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// parseIperf3Bandwidth parses bandwidth from iperf3 output
func parseIperf3Bandwidth(t *testing.T, output string) float64 {
	// Look for the sender summary line
	// Example: "[  5]   0.00-5.00   sec  59.0 MBytes  99.0 Mbits/sec                  sender"
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "sender") && strings.Contains(line, "Mbits/sec") {
			fields := strings.Fields(line)
			for i, field := range fields {
				if field == "Mbits/sec" && i > 0 {
					bandwidth, err := strconv.ParseFloat(fields[i-1], 64)
					require.NoError(t, err, "Failed to parse bandwidth from: %s", line)
					return bandwidth
				}
			}
		}
	}
	
	t.Fatalf("Could not find bandwidth in iperf3 output:\n%s", output)
	return 0
}

// findTestInterface finds a suitable network interface for testing
func findTestInterface(t *testing.T) string {
	// First try to create a dummy interface
	if err := exec.Command("ip", "link", "add", "dummy0", "type", "dummy").Run(); err == nil {
		// Set it up
		exec.Command("ip", "link", "set", "dummy0", "up").Run()
		exec.Command("ip", "addr", "add", "192.168.100.1/24", "dev", "dummy0").Run()
		t.Cleanup(func() {
			exec.Command("ip", "link", "del", "dummy0").Run()
		})
		return "dummy0"
	}
	
	// If we can't create dummy, skip for now
	// In CI environment, we'll use veth pairs
	return ""
}