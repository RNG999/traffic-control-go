//go:build integration
// +build integration

package integration_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rng999/traffic-control-go/api"
	"github.com/stretchr/testify/require"
)

// TestTrafficControlWithIperf3 tests actual bandwidth limiting using iperf3
func TestTrafficControlWithIperf3(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping iperf3 test in short mode")
	}

	// Skip iperf3 tests in CI environment due to virtualization limitations
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping iperf3 bandwidth tests in CI environment")
	}

	// Check if running as root (skip this check in CI)
	if os.Getenv("CI") != "true" && os.Geteuid() != 0 {
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
		name         string
		limitMbps    int
		expectedMbps float64
		tolerance    float64 // percentage tolerance
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
			tcController := api.NetworkInterface(device)
			tcController.WithHardLimitBandwidth(fmt.Sprintf("%dmbit", tc.limitMbps))
			tcController.
				CreateTrafficClass("test_limit").
				WithGuaranteedBandwidth(fmt.Sprintf("%dmbit", tc.limitMbps)).
				WithPriority(4) // Normal priority

			err := tcController.Apply()
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

	if os.Getenv("CI") != "true" && os.Geteuid() != 0 {
		t.Skip("Test requires root privileges")
	}

	if _, err := exec.LookPath("iperf3"); err != nil {
		t.Skip("iperf3 not installed, skipping test")
	}

	device := "lo"

	// Clean up any existing tc rules
	cleanupTC(t, device)

	// Apply traffic control with priority classes
	tcController := api.NetworkInterface(device)
	tcController.WithHardLimitBandwidth("20mbit")
	tcController.
		CreateTrafficClass("high_priority").
		WithGuaranteedBandwidth("15mbit").
		WithPriority(1)
	tcController.
		CreateTrafficClass("low_priority").
		WithGuaranteedBandwidth("5mbit").
		WithPriority(7)

	err := tcController.Apply()
	require.NoError(t, err, "Failed to apply traffic control")

	// Note: Full priority testing would require marking packets with different
	// priorities and running multiple iperf3 streams simultaneously
	// This is a simplified test that verifies the configuration applies successfully

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

func findTestInterface(t *testing.T) string {
	// Try to find a suitable network interface for testing
	// Prefer loopback for isolation, but can use others if needed

	// First try loopback
	if _, err := exec.Command("ip", "link", "show", "lo").Output(); err == nil {
		return "lo"
	}

	// Try to find another interface
	output, err := exec.Command("ip", "link", "show").Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, ": ") && strings.Contains(line, "state UP") {
			parts := strings.Split(line, ": ")
			if len(parts) >= 2 {
				iface := strings.TrimSpace(parts[1])
				if iface != "" && !strings.Contains(iface, "@") {
					return iface
				}
			}
		}
	}

	return ""
}

// TestMultipleClassesConcurrent tests multiple traffic classes running concurrently
func TestMultipleClassesConcurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	// Skip in CI environment - complex multi-stream traffic classification unreliable in virtualized CI
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping complex multi-stream traffic classification test in CI environment")
	}

	if os.Geteuid() != 0 {
		t.Skip("Test requires root privileges")
	}

	if _, err := exec.LookPath("iperf3"); err != nil {
		t.Skip("iperf3 not installed, skipping test")
	}

	// Create veth pair for proper network testing
	_, cleanup := setupIperfVethPair(t, "concurrent")
	defer cleanup()
	
	device := "concurrent"

	// Start iperf3 servers on different ports
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Server 1 on port 5201 (bind to veth peer IP)
	server1 := exec.CommandContext(ctx, "iperf3", "-s", "-B", "10.0.1.2", "-p", "5201", "-1")
	go func() { _ = server1.Run() }()

	// Server 2 on port 5202 (bind to veth peer IP)
	server2 := exec.CommandContext(ctx, "iperf3", "-s", "-B", "10.0.1.2", "-p", "5202", "-1")
	go func() { _ = server2.Run() }()

	time.Sleep(2 * time.Second)

	// Apply traffic control with two classes (smaller values for CI compatibility)
	tcController := api.NetworkInterface(device)
	tcController.WithHardLimitBandwidth("20mbit")

	// High priority class - more bandwidth
	tcController.CreateTrafficClass("high-priority").
		WithGuaranteedBandwidth("12mbit").
		WithSoftLimitBandwidth("15mbit").
		WithPriority(1).
		ForPort(5201)

	// Low priority class - less bandwidth
	tcController.CreateTrafficClass("low-priority").
		WithGuaranteedBandwidth("4mbit").
		WithSoftLimitBandwidth("8mbit").
		WithPriority(6).
		ForPort(5202)

	err := tcController.Apply()
	require.NoError(t, err, "Failed to apply traffic control")

	// Run concurrent iperf3 tests
	var wg sync.WaitGroup
	var highBandwidth, lowBandwidth float64
	var highErr, lowErr error

	wg.Add(2)

	// High priority traffic
	go func() {
		defer wg.Done()
		cmd := exec.Command("iperf3", "-c", "10.0.1.2", "-p", "5201", "-t", "10", "-f", "m")
		output, err := cmd.CombinedOutput()
		if err != nil {
			highErr = err
			return
		}
		highBandwidth = parseIperf3Bandwidth(t, string(output))
	}()

	// Low priority traffic
	go func() {
		defer wg.Done()
		// Start slightly later to ensure both are running concurrently
		time.Sleep(1 * time.Second)
		cmd := exec.Command("iperf3", "-c", "10.0.1.2", "-p", "5202", "-t", "8", "-f", "m")
		output, err := cmd.CombinedOutput()
		if err != nil {
			lowErr = err
			return
		}
		lowBandwidth = parseIperf3Bandwidth(t, string(output))
	}()

	wg.Wait()

	require.NoError(t, highErr, "High priority iperf3 failed")
	require.NoError(t, lowErr, "Low priority iperf3 failed")

	t.Logf("High priority bandwidth: %.2f Mbps", highBandwidth)
	t.Logf("Low priority bandwidth: %.2f Mbps", lowBandwidth)

	// High priority should get significantly more bandwidth
	require.Greater(t, highBandwidth, lowBandwidth*1.5,
		"High priority should get at least 1.5x more bandwidth than low priority")

	// High priority should be close to its guaranteed bandwidth (12mbit = 12 Mbps)
	require.Greater(t, highBandwidth, 8.0, "High priority bandwidth too low")
	require.Less(t, highBandwidth, 20.0, "High priority bandwidth too high")

	// Low priority should be limited (4mbit = 4 Mbps) 
	require.Greater(t, lowBandwidth, 2.0, "Low priority bandwidth too low")
	require.Less(t, lowBandwidth, 10.0, "Low priority bandwidth too high")
}

// TestDynamicBandwidthChange tests changing bandwidth limits during active traffic
func TestDynamicBandwidthChange(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping dynamic test in short mode")
	}

	if os.Getenv("CI") != "true" && os.Geteuid() != 0 {
		t.Skip("Test requires root privileges")
	}

	if _, err := exec.LookPath("iperf3"); err != nil {
		t.Skip("iperf3 not installed, skipping test")
	}

	device := findTestInterface(t)
	if device == "" {
		t.Skip("No suitable network interface found for testing")
	}

	// Start iperf3 server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := exec.CommandContext(ctx, "iperf3", "-s", "-p", "5201")
	err := server.Start()
	require.NoError(t, err, "Failed to start iperf3 server")
	defer func() { _ = server.Process.Kill() }()

	time.Sleep(1 * time.Second)

	// Initial traffic control - 50mbit
	tcController := api.NetworkInterface(device)
	tcController.WithHardLimitBandwidth("100mbit")
	tcController.CreateTrafficClass("dynamic").
		WithGuaranteedBandwidth("50mbit").
		WithPriority(4)

	err = tcController.Apply()
	require.NoError(t, err, "Failed to apply initial traffic control")

	// Start long-running iperf3 client in background
	clientCtx, clientCancel := context.WithCancel(context.Background())
	defer clientCancel()

	client := exec.CommandContext(clientCtx, "iperf3", "-c", "localhost", "-p", "5201", "-t", "20", "-i", "1")
	clientPipe, err := client.StdoutPipe()
	require.NoError(t, err)

	err = client.Start()
	require.NoError(t, err, "Failed to start iperf3 client")

	// Let it run for 5 seconds with initial settings
	time.Sleep(5 * time.Second)

	// Change bandwidth to 20mbit
	t.Log("Changing bandwidth limit to 20mbit")
	tcController = api.NetworkInterface(device)
	tcController.WithHardLimitBandwidth("100mbit")
	tcController.CreateTrafficClass("dynamic").
		WithGuaranteedBandwidth("20mbit").
		WithPriority(4)

	err = tcController.Apply()
	require.NoError(t, err, "Failed to apply new traffic control")

	// Let it run for another 5 seconds
	time.Sleep(5 * time.Second)

	// Change bandwidth to 80mbit
	t.Log("Changing bandwidth limit to 80mbit")
	tcController = api.NetworkInterface(device)
	tcController.WithHardLimitBandwidth("100mbit")
	tcController.CreateTrafficClass("dynamic").
		WithGuaranteedBandwidth("80mbit").
		WithPriority(4)

	err = tcController.Apply()
	require.NoError(t, err, "Failed to apply new traffic control")

	// Let it finish
	time.Sleep(5 * time.Second)
	clientCancel()

	// Read output
	output, _ := io.ReadAll(clientPipe)
	_ = client.Wait()

	t.Logf("Dynamic bandwidth test output:\n%s", string(output))
	// Visual inspection of output should show bandwidth changes
}

func parseIperf3Bandwidth(t *testing.T, output string) float64 {
	// Parse iperf3 output to extract bandwidth
	// Look for lines like:
	// [  4]   0.00-5.00   sec  6.25 MBytes  10.5 Mbits/sec                  sender
	// [  4]   0.00-5.04   sec  6.12 MBytes  10.2 Mbits/sec                  receiver

	// We want the receiver bandwidth as it's more accurate for our test
	re := regexp.MustCompile(`\[\s*\d+\]\s+[\d.-]+-[\d.-]+\s+sec\s+[\d.]+\s+\w+\s+([\d.]+)\s+Mbits/sec\s+receiver`)
	matches := re.FindStringSubmatch(output)

	if len(matches) < 2 {
		// Try sender if receiver not found
		re = regexp.MustCompile(`\[\s*\d+\]\s+[\d.-]+-[\d.-]+\s+sec\s+[\d.]+\s+\w+\s+([\d.]+)\s+Mbits/sec\s+sender`)
		matches = re.FindStringSubmatch(output)
	}

	if len(matches) < 2 {
		t.Logf("Could not parse bandwidth from iperf3 output:\n%s", output)
		t.Fatal("Failed to parse iperf3 bandwidth")
	}

	bandwidth, err := strconv.ParseFloat(matches[1], 64)
	require.NoError(t, err, "Failed to parse bandwidth value: %s", matches[1])

	return bandwidth
}
