//go:build integration
// +build integration

package integration_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/rng999/traffic-control-go/api"
)

// TestTrafficControlWithIperf tests actual bandwidth limiting using iperf3
func TestTrafficControlWithIperf(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping iperf test in short mode")
	}

	// Check if running as root
	if os.Geteuid() != 0 {
		t.Skip("Test requires root privileges")
	}

	// Check if iperf3 is installed
	if _, err := exec.LookPath("iperf3"); err != nil {
		t.Skip("iperf3 not installed, skipping test")
	}

	// Use loopback interface for testing
	device := "lo"
	
	// Start iperf3 server in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	serverCmd := exec.CommandContext(ctx, "iperf3", "-s", "-p", "5201", "-1") // -1 for one connection then exit
	serverErr := make(chan error, 1)
	go func() {
		err := serverCmd.Run()
		if err != nil && ctx.Err() == nil {
			serverErr <- err
		}
	}()
	
	// Wait for server to start
	time.Sleep(2 * time.Second)
	
	// Test different bandwidth limits
	testCases := []struct {
		name          string
		limitMbps     int
		expectedMbps  float64
		tolerance     float64 // percentage tolerance
	}{
		{
			name:         "10Mbps limit",
			limitMbps:    10,
			expectedMbps: 10.0,
			tolerance:    20.0, // 20% tolerance
		},
		{
			name:         "5Mbps limit",
			limitMbps:    5,
			expectedMbps: 5.0,
			tolerance:    20.0,
		},
		{
			name:         "1Mbps limit",
			limitMbps:    1,
			expectedMbps: 1.0,
			tolerance:    30.0, // Higher tolerance for lower speeds
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clean up any existing tc rules
			cleanupTC(t, device)
			
			// Apply traffic control
			tcController := api.New(device)
			err := tcController.
				SetTotalBandwidth(fmt.Sprintf("%dmbit", tc.limitMbps)).
				AddClass("test_limit", fmt.Sprintf("%dmbit", tc.limitMbps)).
				Apply()
			require.NoError(t, err, "Failed to apply traffic control")
			
			// Wait for TC to take effect
			time.Sleep(1 * time.Second)
			
			// Run iperf client
			output, err := runIperfClient()
			require.NoError(t, err, "Failed to run iperf client")
			
			// Parse bandwidth from iperf output
			actualMbps := parseIperfBandwidth(t, output)
			
			// Check if actual bandwidth is within tolerance
			lowerBound := tc.expectedMbps * (1 - tc.tolerance/100)
			upperBound := tc.expectedMbps * (1 + tc.tolerance/100)
			
			t.Logf("Expected: %.1f Mbps, Actual: %.1f Mbps (tolerance: %.0f%%)", 
				tc.expectedMbps, actualMbps, tc.tolerance)
			
			require.GreaterOrEqual(t, actualMbps, lowerBound, 
				"Bandwidth %.1f Mbps is below lower bound %.1f Mbps", actualMbps, lowerBound)
			require.LessOrEqual(t, actualMbps, upperBound,
				"Bandwidth %.1f Mbps is above upper bound %.1f Mbps", actualMbps, upperBound)
			
			// Clean up
			cleanupTC(t, device)
		})
	}
}

// TestTrafficControlPriority tests priority-based traffic shaping
func TestTrafficControlPriority(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping iperf test in short mode")
	}

	if os.Geteuid() != 0 {
		t.Skip("Test requires root privileges")
	}

	if _, err := exec.LookPath("iperf3"); err != nil {
		t.Skip("iperf3 not installed, skipping test")
	}

	device := "lo"
	
	// Clean up any existing tc rules
	cleanupTC(t, device)
	
	// Apply traffic control with priority classes
	tcController := api.New(device)
	err := tcController.
		SetTotalBandwidth("20mbit").
		AddClass("high_priority", "15mbit").
		AddClass("low_priority", "5mbit").
		Apply()
	require.NoError(t, err, "Failed to apply traffic control")
	
	// Note: Full priority testing would require marking packets with different
	// priorities and running multiple iperf streams simultaneously.
	// This is a simplified test that verifies the configuration applies successfully.
	
	// Clean up
	cleanupTC(t, device)
}

// Helper functions

func cleanupTC(t *testing.T, device string) {
	// Delete existing qdisc (this removes all TC configuration)
	cmd := exec.Command("tc", "qdisc", "del", "dev", device, "root")
	_ = cmd.Run() // Ignore error if no qdisc exists
}

func runIperfClient() (string, error) {
	cmd := exec.Command("iperf3", "-c", "127.0.0.1", "-p", "5201", "-t", "5", "-f", "m")
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func parseIperfBandwidth(t *testing.T, output string) float64 {
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
	
	t.Fatalf("Could not find bandwidth in iperf output:\n%s", output)
	return 0
}