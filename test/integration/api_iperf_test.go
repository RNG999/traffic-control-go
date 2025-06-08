//go:build integration
// +build integration

package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/api"
)

// IperfResult represents iperf3 JSON output
type IperfResult struct {
	End struct {
		SumSent struct {
			BitsPerSecond float64 `json:"bits_per_second"`
		} `json:"sum_sent"`
		SumReceived struct {
			BitsPerSecond float64 `json:"bits_per_second"`
		} `json:"sum_received"`
	} `json:"end"`
}


// runIperf3Client runs iperf3 client and returns bandwidth in bits per second
func runIperf3Client(t *testing.T, serverIP string, duration int) float64 {
	t.Helper()
	
	// Run iperf3 client with JSON output
	cmd := exec.Command("iperf3", "-c", serverIP, "-t", strconv.Itoa(duration), "-J")
	output, err := cmd.Output()
	require.NoError(t, err, "iperf3 client failed")
	
	// Parse JSON output
	var result IperfResult
	err = json.Unmarshal(output, &result)
	require.NoError(t, err, "Failed to parse iperf3 JSON output")
	
	return result.End.SumSent.BitsPerSecond
}

// startIperf3Server starts iperf3 server in background
func startIperf3Server(t *testing.T, ctx context.Context, bindIP string) {
	t.Helper()
	
	// Start iperf3 server
	cmd := exec.CommandContext(ctx, "iperf3", "-s", "-B", bindIP, "-1")
	go func() {
		cmd.Run()
	}()
	
	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)
}

// TestAPIWithIperf3BandwidthLimiting tests that API actually limits bandwidth using iperf3
func TestAPIWithIperf3BandwidthLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping iperf3 test in short mode")
	}
	
	if os.Geteuid() != 0 {
		t.Skip("Test requires root privileges for TC operations")
		return
	}
	
	
	if _, err := exec.LookPath("iperf3"); err != nil {
		t.Skip("iperf3 not installed, skipping bandwidth test")
	}
	
	
	t.Run("Basic Bandwidth Limiting", func(t *testing.T) {
		// Create veth pair with IPs
		_, cleanup := setupIperfVethPair(t, "bw-test")
		defer cleanup()
		
		// Test without traffic control first (baseline)
		ctx1, cancel1 := context.WithCancel(context.Background())
		startIperf3Server(t, ctx1, "10.0.1.2")
		
		baselineBandwidth := runIperf3Client(t, "10.0.1.2", 2)
		cancel1()
		
		t.Logf("Baseline bandwidth: %.2f Mbps", baselineBandwidth/1_000_000)
		
		// Apply traffic control with 10 Mbps limit
		controller := api.NetworkInterface("bw-test")
		controller.WithHardLimitBandwidth("50mbps") // Interface limit
		
		controller.CreateTrafficClass("Limited").
			WithGuaranteedBandwidth("10mbps").
			WithSoftLimitBandwidth("10mbps").  // Hard limit at 10 Mbps
			WithPriority(1).
			ForPort(5201) // iperf3 default port
		
		err := controller.Apply()
		require.NoError(t, err, "Failed to apply traffic control")
		
		// Debug: Show TC configuration
		if os.Getenv("CI") == "true" {
			t.Log("=== DEBUG: TC Configuration ===")
			if output, err := exec.Command("tc", "qdisc", "show", "dev", "bw-test").CombinedOutput(); err == nil {
				t.Logf("TC Qdisc: %s", string(output))
			}
			if output, err := exec.Command("tc", "class", "show", "dev", "bw-test").CombinedOutput(); err == nil {
				t.Logf("TC Class: %s", string(output))
			}
			if output, err := exec.Command("tc", "filter", "show", "dev", "bw-test").CombinedOutput(); err == nil {
				t.Logf("TC Filter: %s", string(output))
			}
		}
		
		// Test with traffic control
		ctx2, cancel2 := context.WithCancel(context.Background())
		defer cancel2()
		startIperf3Server(t, ctx2, "10.0.1.2")
		
		limitedBandwidth := runIperf3Client(t, "10.0.1.2", 3)
		
		t.Logf("Limited bandwidth: %.2f Mbps", limitedBandwidth/1_000_000)
		t.Logf("Bandwidth reduction: %.1f%%", (1-limitedBandwidth/baselineBandwidth)*100)
		
		// Verify bandwidth is actually limited (should be significantly less than baseline)
		// Allow some tolerance for measurement variance
		maxExpectedBandwidth := 15_000_000 // 15 Mbps (10 Mbps + 50% tolerance)
		assert.Less(t, limitedBandwidth, maxExpectedBandwidth, 
			"Bandwidth should be limited to around 10 Mbps")
		
		// Verify there was actual limiting (at least 20% reduction from baseline)
		assert.Less(t, limitedBandwidth, baselineBandwidth*0.8, 
			"Traffic control should reduce bandwidth significantly")
	})
	
	t.Run("Multiple Traffic Classes with Priorities", func(t *testing.T) {
		// Create veth pair
		_, cleanup := setupIperfVethPair(t, "prio-test")
		defer cleanup()
		
		// Apply traffic control with multiple classes
		controller := api.NetworkInterface("prio-test")
		controller.WithHardLimitBandwidth("100mbps")
		
		// High priority class
		controller.CreateTrafficClass("High Priority").
			WithGuaranteedBandwidth("40mbps").
			WithSoftLimitBandwidth("80mbps").
			WithPriority(0). // Highest priority
			ForPort(5001)    // iperf3 default port
		
		// Low priority class
		controller.CreateTrafficClass("Low Priority").
			WithGuaranteedBandwidth("20mbps").
			WithSoftLimitBandwidth("60mbps").
			WithPriority(7). // Lowest priority
			ForPort(5002)    // Alternative port
		
		err := controller.Apply()
		require.NoError(t, err, "Failed to apply multi-class traffic control")
		
		// Test high priority traffic
		ctx1, cancel1 := context.WithCancel(context.Background())
		startIperf3Server(t, ctx1, "10.0.1.2") // Default port 5001
		
		highPriorityBandwidth := runIperf3Client(t, "10.0.1.2", 3)
		cancel1()
		
		t.Logf("High priority bandwidth: %.2f Mbps", highPriorityBandwidth/1_000_000)
		
		// Verify high priority gets reasonable bandwidth
		minExpectedHighPriority := 30_000_000 // At least 30 Mbps
		assert.Greater(t, highPriorityBandwidth, minExpectedHighPriority,
			"High priority class should get substantial bandwidth")
		
		t.Log("Multi-class traffic control test completed successfully")
	})
	
	t.Run("README Example Bandwidth Verification", func(t *testing.T) {
		// Create veth pair
		_, cleanup := setupIperfVethPair(t, "readme")
		defer cleanup()
		
		// Apply exact README.md example
		controller := api.NetworkInterface("readme")
		controller.WithHardLimitBandwidth("100mbps")
		
		controller.CreateTrafficClass("Web Services").
			WithGuaranteedBandwidth("30mbps").
			WithSoftLimitBandwidth("60mbps").
			WithPriority(1).
			ForPort(80, 443)
		
		controller.CreateTrafficClass("SSH Management").
			WithGuaranteedBandwidth("5mbps").
			WithSoftLimitBandwidth("10mbps").
			WithPriority(0).
			ForPort(22)
			
		// Add test traffic class for iperf3 on port 5201
		controller.CreateTrafficClass("Test Traffic").
			WithGuaranteedBandwidth("20mbps").
			WithSoftLimitBandwidth("30mbps").
			WithPriority(2).
			ForPort(5201)
		
		err := controller.Apply()
		require.NoError(t, err, "Failed to apply README example configuration")
		
		// Test general traffic (should use default class)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		startIperf3Server(t, ctx, "10.0.1.2")
		
		generalBandwidth := runIperf3Client(t, "10.0.1.2", 3)
		
		t.Logf("General traffic bandwidth: %.2f Mbps", generalBandwidth/1_000_000)
		
		// Verify traffic control is working (bandwidth should be limited)
		maxExpectedGeneral := 80_000_000 // Should be less than 80 Mbps
		assert.Less(t, generalBandwidth, maxExpectedGeneral,
			"General traffic should be limited by traffic control")
		
		t.Log("README example bandwidth verification completed")
	})
	
	t.Run("Bandwidth Format Compatibility", func(t *testing.T) {
		// Test different bandwidth formats work in practice
		testFormats := []struct {
			format   string
			expected float64 // Expected approximate bandwidth in bps
		}{
			{"10mbps", 10_000_000},
			{"1gbps", 1_000_000_000},
			{"500kbps", 500_000},
		}
		
		for _, tf := range testFormats {
			t.Run(fmt.Sprintf("Format_%s", tf.format), func(t *testing.T) {
				// Create veth pair (keep device name short)
				deviceName := fmt.Sprintf("f%d", int(tf.expected/1000000))
				_, cleanup := setupIperfVethPair(t, deviceName)
				defer cleanup()
				
				// Apply traffic control with specific format
				controller := api.NetworkInterface(deviceName)
				controller.WithHardLimitBandwidth("1gbps") // High interface limit
				
				controller.CreateTrafficClass("Test Format").
					WithGuaranteedBandwidth(tf.format).
					WithSoftLimitBandwidth(tf.format). // Hard limit
					WithPriority(1).
					ForPort(5201) // iperf3 default port
				
				err := controller.Apply()
				require.NoError(t, err, "Failed to apply traffic control for format %s", tf.format)
				
				// Test bandwidth
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				startIperf3Server(t, ctx, "10.0.1.2")
				
				actualBandwidth := runIperf3Client(t, "10.0.1.2", 2)
				
				t.Logf("Format %s: Expected ~%.2f Mbps, Actual %.2f Mbps", 
					tf.format, tf.expected/1_000_000, actualBandwidth/1_000_000)
				
				// Verify bandwidth is in reasonable range (allow significant tolerance for small values)
				tolerance := 0.5 // 50% tolerance
				if tf.expected < 1_000_000 { // For very small bandwidths (< 1 Mbps)
					tolerance = 2.0 // 200% tolerance
				}
				
				assert.Less(t, actualBandwidth, tf.expected*(1+tolerance),
					"Bandwidth should not exceed expected limit by too much")
				
				// For reasonable bandwidths, check lower bound too
				if tf.expected > 5_000_000 { // > 5 Mbps
					assert.Greater(t, actualBandwidth, tf.expected*0.3,
						"Bandwidth should not be too far below expected")
				}
			})
		}
	})
}

// TestAPIPerformanceWithIperf3 tests API performance under load
func TestAPIPerformanceWithIperf3(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}
	
	if os.Geteuid() != 0 {
		t.Skip("Test requires root privileges")
		return
	}
	
	
	if _, err := exec.LookPath("iperf3"); err != nil {
		t.Skip("iperf3 not installed")
	}
	
	
	t.Run("API Response Time Under Load", func(t *testing.T) {
		// Create veth pair
		_, cleanup := setupIperfVethPair(t, "perf-test")
		defer cleanup()
		
		// Start background traffic
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		startIperf3Server(t, ctx, "10.0.1.2")
		
		// Start iperf3 client in background
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					runIperf3Client(t, "10.0.1.2", 1)
					time.Sleep(100 * time.Millisecond)
				}
			}
		}()
		
		// Measure API response time while traffic is flowing
		start := time.Now()
		
		controller := api.NetworkInterface("perf-test")
		controller.WithHardLimitBandwidth("100mbps")
		
		controller.CreateTrafficClass("Performance Test").
			WithGuaranteedBandwidth("50mbps").
			WithSoftLimitBandwidth("80mbps").
			WithPriority(1)
		
		err := controller.Apply()
		require.NoError(t, err, "API should work under traffic load")
		
		elapsed := time.Since(start)
		
		t.Logf("API response time under load: %v", elapsed)
		
		// API should respond within reasonable time even under load
		assert.Less(t, elapsed, 5*time.Second, "API should be responsive under load")
	})
}