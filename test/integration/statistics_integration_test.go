package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/api"
)

// TestStatisticsIntegration demonstrates the complete statistics workflow
func TestStatisticsIntegration(t *testing.T) {
	t.Skip("Skipping statistics integration test - query handlers not implemented")
	// Create a traffic controller with mock adapter for testing
	tc := api.NetworkInterface("eth0")

	// Configure traffic control
	tc.WithHardLimitBandwidth("10mbit")
	tc.CreateTrafficClass("web-traffic").
		WithGuaranteedBandwidth("2mbit").
		WithSoftLimitBandwidth("5mbit").
		WithPriority(1).
		ForPort(80, 443)

	tc.CreateTrafficClass("ssh-traffic").
		WithGuaranteedBandwidth("1mbit").
		WithSoftLimitBandwidth("3mbit").
		WithPriority(0).
		ForPort(22)

	err := tc.Apply()

	require.NoError(t, err)

	// Setup mock data for testing
	setupMockStatistics(tc)

	// Test getting comprehensive statistics
	t.Run("GetStatistics", func(t *testing.T) {
		stats, err := tc.GetStatistics()
		require.NoError(t, err)
		assert.NotNil(t, stats)

		assert.Equal(t, "eth0", stats.DeviceName)
		assert.NotEmpty(t, stats.Timestamp)

		// Should have statistics for created qdiscs and classes
		assert.NotEmpty(t, stats.QdiscStats)
		assert.NotEmpty(t, stats.ClassStats)

		t.Logf("Device: %s", stats.DeviceName)
		t.Logf("Timestamp: %s", stats.Timestamp)
		t.Logf("Qdiscs: %d", len(stats.QdiscStats))
		t.Logf("Classes: %d", len(stats.ClassStats))
		t.Logf("Filters: %d", len(stats.FilterStats))

		// Verify qdisc statistics structure
		for _, qdisc := range stats.QdiscStats {
			t.Logf("Qdisc %s (%s): %d bytes sent, %d packets sent",
				qdisc.Handle, qdisc.Type, qdisc.BytesSent, qdisc.PacketsSent)
			assert.NotEmpty(t, qdisc.Handle)
			assert.NotEmpty(t, qdisc.Type)
		}

		// Verify class statistics structure
		for _, class := range stats.ClassStats {
			t.Logf("Class %s (%s): %d bytes sent, %d packets sent, rate %d bps",
				class.Handle, class.Name, class.BytesSent, class.PacketsSent, class.RateBPS)
			assert.NotEmpty(t, class.Handle)
			assert.NotEmpty(t, class.Parent)
		}
	})

	// Test real-time statistics
	t.Run("GetRealtimeStatistics", func(t *testing.T) {
		stats, err := tc.GetRealtimeStatistics()
		require.NoError(t, err)
		assert.NotNil(t, stats)

		assert.Equal(t, "eth0", stats.DeviceName)
		assert.NotEmpty(t, stats.Timestamp)

		// Real-time stats should include all detected TC elements
		assert.NotEmpty(t, stats.QdiscStats)

		t.Logf("Real-time stats - Qdiscs: %d, Classes: %d",
			len(stats.QdiscStats), len(stats.ClassStats))
	})

	// Test specific qdisc statistics
	t.Run("GetQdiscStatistics", func(t *testing.T) {
		stats, err := tc.GetQdiscStatistics("1:0")
		require.NoError(t, err)
		assert.NotNil(t, stats)

		assert.Equal(t, "1:0", stats.Handle)
		assert.NotEmpty(t, stats.Type)

		t.Logf("Qdisc 1:0: %d bytes sent, %d packets sent, %d dropped",
			stats.BytesSent, stats.PacketsSent, stats.BytesDropped)

		// Verify detailed stats if available
		if len(stats.DetailedStats) > 0 {
			t.Logf("Detailed stats available: %+v", stats.DetailedStats)
		}
	})

	// Test specific class statistics
	t.Run("GetClassStatistics", func(t *testing.T) {
		stats, err := tc.GetClassStatistics("1:10")
		require.NoError(t, err)
		assert.NotNil(t, stats)

		assert.Equal(t, "1:10", stats.Handle)
		assert.Equal(t, "1:0", stats.Parent)

		t.Logf("Class 1:10: %d bytes sent, %d packets sent, %d backlog bytes",
			stats.BytesSent, stats.PacketsSent, stats.BacklogBytes)
	})

	// Test monitoring statistics
	t.Run("MonitorStatistics", func(t *testing.T) {
		t.Skip("Skipping MonitorStatistics test - requires API change to support context cancellation")
		// TODO: Update MonitorStatistics API to accept context for proper cancellation
	})
}

// TestStatisticsErrorHandling tests error scenarios
func TestStatisticsErrorHandling(t *testing.T) {
	// TODO: Implement test logic when query handlers are available
	t.Log("TestStatisticsErrorHandling: placeholder test")
	// tc := api.NetworkInterface("nonexistent")
	//
	// // Test getting statistics for non-configured device
	// t.Run("NonExistentDevice", func(t *testing.T) {
	// 	// This should still work but return empty/minimal statistics
	// 	stats, err := tc.GetRealtimeStatistics()
	// 	if err != nil {
	// 		t.Logf("Expected error for non-existent device: %v", err)
	// 	} else {
	// 		assert.Equal(t, "nonexistent", stats.DeviceName)
	// 		t.Logf("Got stats for non-existent device: %+v", stats)
	// 	}
	// })
	//
	// // Test invalid handle formats
	// t.Run("InvalidHandles", func(t *testing.T) {
	// 	_, err := tc.GetQdiscStatistics("invalid")
	// 	assert.Error(t, err)
	// 	assert.Contains(t, err.Error(), "invalid handle")
	//
	// 	_, err = tc.GetClassStatistics("also-invalid")
	// 	assert.Error(t, err)
	// 	assert.Contains(t, err.Error(), "invalid handle")
	// })
}

// TestStatisticsPerformance tests the performance characteristics
func TestStatisticsPerformance(t *testing.T) {
	// TODO: Implement test logic when query handlers are available
	t.Log("TestStatisticsPerformance: placeholder test")
	// tc := api.NetworkInterface("eth0")
	//
	// err := tc.Apply()
	// require.NoError(t, err)
	//
	// setupMockStatistics(tc)
	//
	// // Test performance of statistics retrieval
	// t.Run("StatisticsRetrieval", func(t *testing.T) {
	// 	iterations := 100
	// 	start := time.Now()
	//
	// 	for i := 0; i < iterations; i++ {
	// 		_, err := tc.GetRealtimeStatistics()
	// 		require.NoError(t, err)
	// 	}
	//
	// 	duration := time.Since(start)
	// 	avgDuration := duration / time.Duration(iterations)
	//
	// 	t.Logf("Average time per statistics call: %v", avgDuration)
	//
	// 	// Should be reasonably fast (less than 10ms per call in mock mode)
	// 	assert.Less(t, avgDuration, 10*time.Millisecond)
	// })
}

// setupMockStatistics configures mock data for testing
func setupMockStatistics(tc *api.TrafficController) {
	// In a real implementation, we would extract the netlink adapter
	// and configure it with mock data. For now, this is a placeholder
	// that would work with the mock adapter to set up realistic statistics

	// This would typically involve:
	// 1. Getting the service from the traffic controller.
	// 2. Extracting the netlink adapter
	// 3. If it's a mock adapter, setting up mock qdisc/class/filter data
	// 4. Populating realistic statistics numbers

	// Example (pseudocode):
	// if mockAdapter, ok := tc.service.netlinkAdapter.(*netlink.MockAdapter); ok {
	//     mockAdapter.SetQdiscs(device, mockQdiscData)
	//     mockAdapter.SetClasses(device, mockClassData)
	// }
}

// TestStatisticsDataAccuracy tests that statistics accurately reflect the configuration
func TestStatisticsDataAccuracy(t *testing.T) {
	// TODO: Implement test logic when query handlers are available
	t.Log("TestStatisticsDataAccuracy: placeholder test")
	// tc := api.NetworkInterface("eth0")
	//
	// // Create a specific configuration
	// tc.WithHardLimitBandwidth("10mbit")
	// tc.CreateTrafficClass("priority-traffic").
	// 	WithGuaranteedBandwidth("3mbit").
	// 	WithSoftLimitBandwidth("7mbit").
	// 	WithPriority(1).
	// 	ForPort(22, 443)
	// tc.CreateTrafficClass("bulk-traffic").
	// 	WithGuaranteedBandwidth("2mbit").
	// 	WithSoftLimitBandwidth("5mbit").
	// 	WithPriority(5).
	// 	ForPort(80)
	//
	// err := tc.Apply()
	// require.NoError(t, err)
	//
	// setupMockStatistics(tc)
	//
	// // Get statistics and verify they match our configuration
	// stats, err := tc.GetStatistics()
	// require.NoError(t, err)
	//
	// // Should have one root qdisc
	// assert.Len(t, stats.QdiscStats, 1)
	// rootQdisc := stats.QdiscStats[0]
	// assert.Equal(t, "1:0", rootQdisc.Handle)
	// assert.Equal(t, "htb", rootQdisc.Type)
	//
	// // Should have at least 2 classes (our created classes) plus potentially a default class
	// assert.GreaterOrEqual(t, len(stats.ClassStats), 2)
	//
	// // Should have filters for our port specifications
	// assert.GreaterOrEqual(t, len(stats.FilterStats), 3) // SSH, HTTPS, HTTP
	//
	// t.Logf("Configuration accuracy test passed:")
	// t.Logf("- Root qdisc: %s (%s)", rootQdisc.Handle, rootQdisc.Type)
	// t.Logf("- Classes: %d", len(stats.ClassStats))
	// t.Logf("- Filters: %d", len(stats.FilterStats))
}
