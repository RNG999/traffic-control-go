//go:build integration
// +build integration

package integration_test

import (
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