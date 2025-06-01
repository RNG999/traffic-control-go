//go:build integration
// +build integration

package integration_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/rng999/traffic-control-go/api"
)

// TestTrafficControlWithIperf3 tests actual bandwidth limiting using iperf3
func TestTrafficControlWithIperf3(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping iperf3 test in short mode")
	}

	// Check if running as root
	if os.Geteuid() != 0 {
		t.Skip("Test requires root privileges")
	}

	// Check if iperf3 is installed
	if _, err := exec.LookPath("iperf3"); err != nil {
		t.Skip("iperf3 not installed, skipping test")
	}

	// Skip test if no suitable network interface is available
	device := findTestInterface(t)
	if device == "" {
		t.Skip("No suitable network interface found for testing")
	}
	
	// Start iperf3 server in background for the entire test
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	serverCmd := exec.CommandContext(ctx, "iperf3", "-s", "-p", "5201")
	err := serverCmd.Start()
	require.NoError(t, err, "Failed to start iperf3 server")
	defer func() {
		cancel()
		_ = serverCmd.Wait()
	}()
	
	// Wait for server to start
	time.Sleep(1 * time.Second)
	
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
				CreateTrafficClass("test_limit").
				WithGuaranteedBandwidth(fmt.Sprintf("%dmbit", tc.limitMbps)).
				WithPriority(4). // Normal priority
				And().
				Apply()
			require.NoError(t, err, "Failed to apply traffic control")
			
			// Verify TC was applied
			tcOutput, _ := exec.Command("tc", "qdisc", "show", "dev", device).CombinedOutput()
			t.Logf("TC qdisc after apply: %s", tcOutput)
			tcOutput, _ = exec.Command("tc", "class", "show", "dev", device).CombinedOutput()
			t.Logf("TC class after apply: %s", tcOutput)
			
			// Wait for TC to take effect
			time.Sleep(1 * time.Second)
			
			// Run iperf client
			output, err := runIperfClient()
			if err != nil {
				t.Logf("iperf3 output:\n%s", output)
			}
			require.NoError(t, err, "Failed to run iperf client")
			
			// Parse bandwidth from iperf3 output
			actualMbps := parseIperf3Bandwidth(t, output)
			
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
		t.Skip("Skipping iperf3 test in short mode")
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
		CreateTrafficClass("high_priority").
		WithGuaranteedBandwidth("15mbit").
		WithPriority(1).
		And().
		CreateTrafficClass("low_priority").
		WithGuaranteedBandwidth("5mbit").
		WithPriority(7).
		And().
		Apply()
	require.NoError(t, err, "Failed to apply traffic control")
	
	// Note: Full priority testing would require marking packets with different
	// priorities and running multiple iperf3 streams simultaneously.
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

