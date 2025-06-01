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

// TestTrafficControlWithVethPair tests TC with virtual ethernet pair
func TestTrafficControlWithVethPair(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping veth test in short mode")
	}

	if os.Geteuid() != 0 {
		t.Skip("Test requires root privileges")
	}

	if _, err := exec.LookPath("iperf3"); err != nil {
		t.Skip("iperf3 not installed, skipping test")
	}

	// Create virtual ethernet pair
	veth0 := "veth0-tc-test"
	veth1 := "veth1-tc-test"
	
	// Clean up any existing interfaces
	cleanupVeth(veth0, veth1)
	defer cleanupVeth(veth0, veth1)
	
	// Create veth pair
	err := createVethPair(veth0, veth1)
	require.NoError(t, err, "Failed to create veth pair")
	
	// Configure IP addresses
	err = configureVethIPs(veth0, veth1)
	require.NoError(t, err, "Failed to configure veth IPs")
	
	// Start iperf server in network namespace
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	serverCmd := exec.CommandContext(ctx, "ip", "netns", "exec", "tc-test-ns", 
		"iperf3", "-s", "-B", "192.168.100.2", "-p", "5201", "-1")
	go func() {
		_ = serverCmd.Run()
	}()
	
	// Wait for server to start
	time.Sleep(2 * time.Second)
	
	// Apply traffic control on veth0
	tcController := api.New(veth0)
	err = tcController.
		SetTotalBandwidth("100mbit").
		CreateTrafficClass("limited").
		WithGuaranteedBandwidth("10mbit").
		WithPriority(4). // Normal priority
		And().
		Apply()
	require.NoError(t, err, "Failed to apply traffic control")
	
	// Run iperf client
	clientCmd := exec.Command("iperf3", "-c", "192.168.100.2", "-p", "5201", 
		"-t", "5", "-f", "m", "-b", "100M") // Try to send at 100Mbps
	output, err := clientCmd.CombinedOutput()
	require.NoError(t, err, "Failed to run iperf client: %s", string(output))
	
	// Parse and verify bandwidth
	actualMbps := parseIperf3Bandwidth(t, string(output))
	t.Logf("Measured bandwidth: %.1f Mbps (expected ~10 Mbps)", actualMbps)
	
	// Verify bandwidth is limited (with tolerance)
	require.Less(t, actualMbps, 15.0, "Bandwidth should be limited to ~10 Mbps")
	require.Greater(t, actualMbps, 5.0, "Bandwidth too low, TC might not be working correctly")
}

// TestBurstTraffic tests token bucket burst behavior
func TestBurstTraffic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping burst test in short mode")
	}

	if os.Geteuid() != 0 {
		t.Skip("Test requires root privileges")
	}

	device := "lo"
	
	// Clean up
	cleanupTC(t, device)
	defer cleanupTC(t, device)
	
	// Apply TBF (Token Bucket Filter) with burst
	cmd := exec.Command("tc", "qdisc", "add", "dev", device, "root", "handle", "1:", 
		"tbf", "rate", "1mbit", "burst", "10kb", "latency", "50ms")
	err := cmd.Run()
	require.NoError(t, err, "Failed to apply TBF qdisc")
	
	// Send burst traffic using ping with large packet size
	// Initial burst should go through quickly, then rate limited
	cmd = exec.Command("ping", "-c", "20", "-s", "1400", "-i", "0.01", "127.0.0.1")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Ping failed: %s", string(output))
	
	// Parse ping statistics
	t.Logf("Ping output:\n%s", string(output))
	// First few packets should have low RTT (burst), later packets higher RTT (rate limited)
}

// Helper functions for veth tests

func createVethPair(veth0, veth1 string) error {
	// Create network namespace
	if err := exec.Command("ip", "netns", "add", "tc-test-ns").Run(); err != nil {
		// Ignore if already exists
	}
	
	// Create veth pair
	cmd := exec.Command("ip", "link", "add", veth0, "type", "veth", "peer", "name", veth1)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create veth pair: %w", err)
	}
	
	// Move veth1 to namespace
	cmd = exec.Command("ip", "link", "set", veth1, "netns", "tc-test-ns")
	return cmd.Run()
}

func configureVethIPs(veth0, veth1 string) error {
	// Configure veth0
	cmds := [][]string{
		{"ip", "addr", "add", "192.168.100.1/24", "dev", veth0},
		{"ip", "link", "set", veth0, "up"},
		// Configure veth1 in namespace
		{"ip", "netns", "exec", "tc-test-ns", "ip", "addr", "add", "192.168.100.2/24", "dev", veth1},
		{"ip", "netns", "exec", "tc-test-ns", "ip", "link", "set", veth1, "up"},
		{"ip", "netns", "exec", "tc-test-ns", "ip", "link", "set", "lo", "up"},
	}
	
	for _, args := range cmds {
		if err := exec.Command(args[0], args[1:]...).Run(); err != nil {
			return fmt.Errorf("failed to run %v: %w", args, err)
		}
	}
	
	return nil
}

func cleanupVeth(veth0, veth1 string) {
	// Delete veth pair (this also removes veth1 from namespace)
	_ = exec.Command("ip", "link", "del", veth0).Run()
	
	// Delete namespace
	_ = exec.Command("ip", "netns", "del", "tc-test-ns").Run()
}