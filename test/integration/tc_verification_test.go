//go:build integration
// +build integration

package integration_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/api"
)

// IperfStats represents iperf3 JSON output
type IperfStats struct {
	End struct {
		SumSent struct {
			BitsPerSecond float64 `json:"bits_per_second"`
		} `json:"sum_sent"`
		SumReceived struct {
			BitsPerSecond float64 `json:"bits_per_second"`
		} `json:"sum_received"`
	} `json:"end"`
}

// setupTCVethPair creates veth pair for TC testing
func setupTCVethPair(t *testing.T, name string) (string, func()) {
	t.Helper()
	
	peer := name + "-peer"
	
	// Clean up any existing interfaces
	exec.Command("ip", "link", "delete", name).Run()
	
	// Create veth pair
	cmd := exec.Command("ip", "link", "add", name, "type", "veth", "peer", "name", peer)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("Failed to create veth pair: %v, output: %s", err, string(output))
		return "", func() {}
	}
	
	// Configure both interfaces
	exec.Command("ip", "link", "set", name, "up").Run()
	exec.Command("ip", "link", "set", peer, "up").Run()
	exec.Command("ip", "addr", "add", "192.168.100.1/24", "dev", name).Run()
	exec.Command("ip", "addr", "add", "192.168.100.2/24", "dev", peer).Run()
	
	// Add routing to ensure traffic goes through the interface
	exec.Command("ip", "route", "add", "192.168.100.2/32", "dev", name).Run()
	
	return peer, func() {
		exec.Command("ip", "link", "delete", name).Run()
	}
}

// runIperf3Test runs iperf3 and returns bandwidth in bps
func runIperf3Test(t *testing.T, serverIP string, duration int, port int) float64 {
	t.Helper()
	
	portStr := strconv.Itoa(port)
	
	// Run iperf3 client
	cmd := exec.Command("iperf3", "-c", serverIP, "-t", strconv.Itoa(duration), "-p", portStr, "-J")
	output, err := cmd.Output()
	require.NoError(t, err, "iperf3 client failed")
	
	var stats IperfStats
	err = json.Unmarshal(output, &stats)
	require.NoError(t, err, "Failed to parse iperf3 output")
	
	return stats.End.SumSent.BitsPerSecond
}

// startIperf3ServerOnPort starts iperf3 server on specific port
func startIperf3ServerOnPort(t *testing.T, ctx context.Context, ip string, port int) {
	t.Helper()
	
	portStr := strconv.Itoa(port)
	cmd := exec.CommandContext(ctx, "iperf3", "-s", "-B", ip, "-p", portStr, "-1")
	go func() {
		cmd.Run()
	}()
	time.Sleep(200 * time.Millisecond) // Wait for server to start
}

// verifyTCConfiguration checks that TC rules are applied
func verifyTCConfiguration(t *testing.T, device string) {
	t.Helper()
	
	// Check qdisc
	cmd := exec.Command("tc", "qdisc", "show", "dev", device)
	output, err := cmd.Output()
	require.NoError(t, err, "Failed to show qdisc")
	t.Logf("TC qdisc on %s: %s", device, string(output))
	
	// Check classes
	cmd = exec.Command("tc", "class", "show", "dev", device)
	output, err = cmd.Output()
	require.NoError(t, err, "Failed to show classes")
	t.Logf("TC classes on %s: %s", device, string(output))
	
	// Check filters
	cmd = exec.Command("tc", "filter", "show", "dev", device)
	output, err = cmd.Output()
	require.NoError(t, err, "Failed to show filters")
	t.Logf("TC filters on %s: %s", device, string(output))
}

// TestTrafficControlFunctionality tests that TC actually limits bandwidth
func TestTrafficControlFunctionality(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TC functionality test in short mode")
	}
	
	if os.Geteuid() != 0 {
		t.Skip("Test requires root privileges")
	}
	
	if _, err := exec.LookPath("iperf3"); err != nil {
		t.Skip("iperf3 not installed")
	}
	
	t.Run("Port-Based Traffic Shaping", func(t *testing.T) {
		// Create test interface
		_, cleanup := setupTCVethPair(t, "tcport")
		defer cleanup()
		
		// Test baseline bandwidth without TC
		ctx1, cancel1 := context.WithCancel(context.Background())
		startIperf3ServerOnPort(t, ctx1, "192.168.100.2", 5201)
		baselineBW := runIperf3Test(t, "192.168.100.2", 3, 5201)
		cancel1()
		
		t.Logf("Baseline bandwidth: %.2f Mbps", baselineBW/1_000_000)
		
		// Apply traffic control for specific port
		controller := api.NetworkInterface("tcport")
		controller.WithHardLimitBandwidth("100mbps")
		
		// Create a class that specifically targets iperf3 traffic
		controller.CreateTrafficClass("IperfLimited").
			WithGuaranteedBandwidth("5mbps").
			WithSoftLimitBandwidth("5mbps"). // Hard limit at 5 Mbps
			WithPriority(1).
			ForPort(5201) // iperf3 default port
		
		err := controller.Apply()
		require.NoError(t, err, "Failed to apply TC configuration")
		
		// Verify TC configuration is applied
		verifyTCConfiguration(t, "tcport")
		
		// Test bandwidth with TC applied
		ctx2, cancel2 := context.WithCancel(context.Background())
		defer cancel2()
		startIperf3ServerOnPort(t, ctx2, "192.168.100.2", 5201)
		limitedBW := runIperf3Test(t, "192.168.100.2", 3, 5201)
		
		t.Logf("Limited bandwidth: %.2f Mbps", limitedBW/1_000_000)
		t.Logf("Reduction: %.1f%%", (1-limitedBW/baselineBW)*100)
		
		// Verify significant bandwidth reduction
		maxExpected := 10_000_000 // 10 Mbps max (5 Mbps + tolerance)
		assert.Less(t, limitedBW, maxExpected, "Bandwidth should be limited to ~5 Mbps")
		
		// Verify actual reduction occurred
		reductionRatio := limitedBW / baselineBW
		assert.Less(t, reductionRatio, 0.5, "Should see at least 50% bandwidth reduction")
	})
	
	t.Run("Interface-Level Bandwidth Limiting", func(t *testing.T) {
		// Create test interface
		_, cleanup := setupTCVethPair(t, "tc-if-test")
		defer cleanup()
		
		// Apply very restrictive interface-level limit
		controller := api.NetworkInterface("tc-if-test")
		controller.WithHardLimitBandwidth("10mbps") // Very low interface limit
		
		// Create a permissive class to test interface limit
		controller.CreateTrafficClass("TestClass").
			WithGuaranteedBandwidth("8mbps").
			WithSoftLimitBandwidth("50mbps"). // Higher than interface limit
			WithPriority(1)
		
		err := controller.Apply()
		require.NoError(t, err, "Failed to apply interface-level TC")
		
		verifyTCConfiguration(t, "tc-if-test")
		
		// Test bandwidth - should be limited by interface limit
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		startIperf3ServerOnPort(t, ctx, "192.168.100.2", 5201)
		
		bandwidth := runIperf3Test(t, "192.168.100.2", 3, 5201)
		
		t.Logf("Interface-limited bandwidth: %.2f Mbps", bandwidth/1_000_000)
		
		// Should be limited by 10 Mbps interface limit
		maxExpected := 15_000_000 // 15 Mbps (10 + tolerance)
		assert.Less(t, bandwidth, maxExpected, "Should be limited by interface bandwidth")
	})
	
	t.Run("Multiple Classes with Different Limits", func(t *testing.T) {
		// Create test interface
		_, cleanup := setupTCVethPair(t, "tc-multi")
		defer cleanup()
		
		// Apply multi-class configuration
		controller := api.NetworkInterface("tc-multi")
		controller.WithHardLimitBandwidth("50mbps")
		
		// High priority, high bandwidth class
		controller.CreateTrafficClass("HighPrio").
			WithGuaranteedBandwidth("20mbps").
			WithSoftLimitBandwidth("30mbps").
			WithPriority(0).
			ForPort(5201)
		
		// Low priority, low bandwidth class
		controller.CreateTrafficClass("LowPrio").
			WithGuaranteedBandwidth("5mbps").
			WithSoftLimitBandwidth("10mbps").
			WithPriority(7).
			ForPort(5202)
		
		err := controller.Apply()
		require.NoError(t, err, "Failed to apply multi-class TC")
		
		verifyTCConfiguration(t, "tc-multi")
		
		// Test high priority traffic
		ctx1, cancel1 := context.WithCancel(context.Background())
		startIperf3ServerOnPort(t, ctx1, "192.168.100.2", 5201)
		highPrioBW := runIperf3Test(t, "192.168.100.2", 3, 5201)
		cancel1()
		
		// Test low priority traffic
		ctx2, cancel2 := context.WithCancel(context.Background())
		defer cancel2()
		startIperf3ServerOnPort(t, ctx2, "192.168.100.2", 5202)
		lowPrioBW := runIperf3Test(t, "192.168.100.2", 3, 5202)
		
		t.Logf("High priority bandwidth: %.2f Mbps", highPrioBW/1_000_000)
		t.Logf("Low priority bandwidth: %.2f Mbps", lowPrioBW/1_000_000)
		
		// High priority should get more bandwidth than low priority
		assert.Greater(t, highPrioBW, lowPrioBW, "High priority should get more bandwidth")
		
		// Both should be reasonably limited
		assert.Less(t, highPrioBW, 40_000_000, "High priority should be limited to ~30 Mbps")
		assert.Less(t, lowPrioBW, 15_000_000, "Low priority should be limited to ~10 Mbps")
	})
	
	t.Run("Default Class Traffic", func(t *testing.T) {
		// Test that unclassified traffic goes to default class
		_, cleanup := setupTCVethPair(t, "tc-default")
		defer cleanup()
		
		controller := api.NetworkInterface("tc-default")
		controller.WithHardLimitBandwidth("20mbps")
		
		// Create a specific class for port 5201 only
		controller.CreateTrafficClass("Specific").
			WithGuaranteedBandwidth("10mbps").
			WithSoftLimitBandwidth("15mbps").
			WithPriority(1).
			ForPort(5201)
		
		err := controller.Apply()
		require.NoError(t, err, "Failed to apply default class TC")
		
		verifyTCConfiguration(t, "tc-default")
		
		// Test classified traffic (port 5201)
		ctx1, cancel1 := context.WithCancel(context.Background())
		startIperf3ServerOnPort(t, ctx1, "192.168.100.2", 5201)
		classifiedBW := runIperf3Test(t, "192.168.100.2", 3, 5201)
		cancel1()
		
		// Test unclassified traffic (different port)
		ctx2, cancel2 := context.WithCancel(context.Background())
		defer cancel2()
		startIperf3ServerOnPort(t, ctx2, "192.168.100.2", 5555)
		unclassifiedBW := runIperf3Test(t, "192.168.100.2", 3, 5555)
		
		t.Logf("Classified traffic (port 5201): %.2f Mbps", classifiedBW/1_000_000)
		t.Logf("Unclassified traffic (port 5555): %.2f Mbps", unclassifiedBW/1_000_000)
		
		// Classified traffic should be limited by class settings
		assert.Less(t, classifiedBW, 20_000_000, "Classified traffic should be limited")
		
		// Both should be under interface limit
		assert.Less(t, classifiedBW, 25_000_000, "Should be under interface limit")
		assert.Less(t, unclassifiedBW, 25_000_000, "Should be under interface limit")
		
		t.Log("Default class test completed - traffic properly classified")
	})
}