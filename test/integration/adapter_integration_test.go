//go:build integration
// +build integration

package integration_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/internal/application"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	"github.com/rng999/traffic-control-go/pkg/logging"
)

// TestNetlinkAdapterIntegration tests the mock netlink adapter integration with the type-safe command bus
func TestNetlinkAdapterIntegration(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	mockAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("adapter-test")
	service := application.NewTrafficControlService(eventStore, mockAdapter, logger)
	ctx := context.Background()

	t.Run("Mock Adapter Qdisc Operations", func(t *testing.T) {
		deviceName := "mock-eth0"

		// Test HTB qdisc creation
		err := service.CreateHTBQdisc(ctx, deviceName, "1:0", "1:999")
		require.NoError(t, err, "HTB qdisc creation should succeed")

		// Test TBF qdisc creation
		err = service.CreateTBFQdisc(ctx, deviceName, "2:0", "100Mbps", 1600, 3000, 1600)
		require.NoError(t, err, "TBF qdisc creation should succeed")

		// Test PRIO qdisc creation
		priomap := []uint8{1, 2, 2, 2, 1, 2, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1}
		err = service.CreatePRIOQdisc(ctx, deviceName, "3:0", 3, priomap)
		require.NoError(t, err, "PRIO qdisc creation should succeed")

		// Test FQ_CODEL qdisc creation
		err = service.CreateFQCODELQdisc(ctx, deviceName, "4:0", 10240, 1024, 5000, 100000, 1514, true)
		require.NoError(t, err, "FQ_CODEL qdisc creation should succeed")

		t.Log("All qdisc types created successfully through mock adapter")
	})

	t.Run("Mock Adapter Class Operations", func(t *testing.T) {
		deviceName := "class-eth0"

		// Create HTB qdisc first
		err := service.CreateHTBQdisc(ctx, deviceName, "1:0", "1:999")
		require.NoError(t, err, "HTB qdisc creation should succeed")

		// Create parent class
		err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:1", "1Gbps", "1Gbps")
		require.NoError(t, err, "Parent class creation should succeed")

		// Create multiple child classes
		testClasses := []struct {
			classID string
			rate    string
			ceil    string
		}{
			{"1:10", "300Mbps", "500Mbps"},
			{"1:20", "200Mbps", "400Mbps"},
			{"1:30", "100Mbps", "300Mbps"},
		}

		for _, tc := range testClasses {
			err = service.CreateHTBClass(ctx, deviceName, "1:1", tc.classID, tc.rate, tc.ceil)
			require.NoError(t, err, "Class %s creation should succeed", tc.classID)
		}

		t.Log("HTB class hierarchy created successfully through mock adapter")
	})

	t.Run("Mock Adapter Filter Operations", func(t *testing.T) {
		deviceName := "filter-eth0"

		// Setup HTB configuration
		err := service.CreateHTBQdisc(ctx, deviceName, "1:0", "1:999")
		require.NoError(t, err, "HTB qdisc creation should succeed")

		err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:10", "100Mbps", "200Mbps")
		require.NoError(t, err, "HTB class creation should succeed")

		// Create various filter types
		testFilters := []struct {
			priority uint16
			flowID   string
			match    map[string]string
		}{
			{1, "1:10", map[string]string{"dst_port": "22", "protocol": "tcp"}},
			{2, "1:10", map[string]string{"dst_port": "80", "protocol": "tcp"}},
			{3, "1:10", map[string]string{"dst_ip": "192.168.1.100"}},
		}

		for _, tf := range testFilters {
			err = service.CreateFilter(ctx, deviceName, "1:0", tf.priority, "ip", tf.flowID, tf.match)
			require.NoError(t, err, "Filter with priority %d should be created", tf.priority)
		}

		t.Log("Multiple filters created successfully through mock adapter")
	})
}

// TestAdapterErrorHandling tests error handling scenarios with the mock adapter
func TestAdapterErrorHandling(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	mockAdapter := netlink.NewMockAdapter()
	logger := logging.NewSilentLogger() // Use silent logger for error tests
	service := application.NewTrafficControlService(eventStore, mockAdapter, logger)
	ctx := context.Background()

	t.Run("Invalid Device Names", func(t *testing.T) {
		// Test empty device name
		err := service.CreateHTBQdisc(ctx, "", "1:0", "1:999")
		assert.Error(t, err, "Empty device name should cause error")
		assert.Contains(t, err.Error(), "device name", "Error should mention device name")

		// Test device name too long (>15 chars)
		longDeviceName := "very-long-device-name-that-exceeds-limit"
		err = service.CreateHTBQdisc(ctx, longDeviceName, "1:0", "1:999")
		assert.Error(t, err, "Long device name should cause error")
		assert.Contains(t, err.Error(), "too long", "Error should mention length limit")
	})

	t.Run("Invalid Handle Formats", func(t *testing.T) {
		deviceName := "error-eth0"

		// Test invalid handle format
		err := service.CreateHTBQdisc(ctx, deviceName, "invalid-handle", "1:999")
		assert.Error(t, err, "Invalid handle format should cause error")
		assert.Contains(t, err.Error(), "handle", "Error should mention handle")

		// Test malformed handle
		err = service.CreateHTBQdisc(ctx, deviceName, "1:2:3", "1:999")
		assert.Error(t, err, "Malformed handle should cause error")
	})

	t.Run("Invalid Bandwidth Specifications", func(t *testing.T) {
		deviceName := "bandwidth-eth0"

		// Create qdisc first
		err := service.CreateHTBQdisc(ctx, deviceName, "1:0", "1:999")
		require.NoError(t, err, "QDisc creation should succeed")

		// Test invalid bandwidth format
		err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:10", "invalid-bandwidth", "200Mbps")
		assert.Error(t, err, "Invalid bandwidth should cause error")
		assert.Contains(t, err.Error(), "bandwidth", "Error should mention bandwidth")

		// Test negative bandwidth
		err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:20", "-100Mbps", "200Mbps")
		assert.Error(t, err, "Negative bandwidth should cause error")
	})

	t.Run("Dependency Violations", func(t *testing.T) {
		deviceName := "dependency-eth0"

		// Try to create class without parent qdisc
		err := service.CreateHTBClass(ctx, deviceName, "1:0", "1:10", "100Mbps", "200Mbps")
		assert.Error(t, err, "Class creation without parent qdisc should fail")

		// Try to create filter without qdisc
		matchRules := map[string]string{"dst_port": "80"}
		err = service.CreateFilter(ctx, deviceName, "1:0", 1, "ip", "1:10", matchRules)
		assert.Error(t, err, "Filter creation without qdisc should fail")
	})
}

// TestAdapterConcurrency tests concurrent operations with the mock adapter
func TestAdapterConcurrency(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	mockAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("concurrency-test")
	service := application.NewTrafficControlService(eventStore, mockAdapter, logger)
	ctx := context.Background()

	t.Run("Concurrent Qdisc Creation", func(t *testing.T) {
		const numGoroutines = 50
		errors := make(chan error, numGoroutines)

		// Launch concurrent operations
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				deviceName := fmt.Sprintf("conc-eth%d", id)
				handleNum := id + 1
				handle := fmt.Sprintf("%d:0", handleNum)
				defaultClass := fmt.Sprintf("%d:999", handleNum)
				err := service.CreateHTBQdisc(ctx, deviceName, handle, defaultClass)
				errors <- err
			}(i)
		}

		// Collect results
		var errorCount int
		for i := 0; i < numGoroutines; i++ {
			if err := <-errors; err != nil {
				errorCount++
				t.Logf("Concurrent operation %d failed: %v", i, err)
			}
		}

		// All operations should succeed with mock adapter
		assert.Equal(t, 0, errorCount, "All concurrent operations should succeed")
		t.Logf("Successfully executed %d concurrent qdisc operations", numGoroutines)
	})

	t.Run("Mixed Concurrent Operations", func(t *testing.T) {
		const opsPerType = 10
		totalOps := opsPerType * 3 // qdisc, class, filter
		errors := make(chan error, totalOps)

		// Setup base configuration for classes and filters
		baseDevice := "mixed-base-eth0"
		err := service.CreateHTBQdisc(ctx, baseDevice, "1:0", "1:999")
		require.NoError(t, err, "Base qdisc creation should succeed")

		err = service.CreateHTBClass(ctx, baseDevice, "1:0", "1:10", "100Mbps", "200Mbps")
		require.NoError(t, err, "Base class creation should succeed")

		// Launch mixed concurrent operations
		// Qdisc operations
		for i := 0; i < opsPerType; i++ {
			go func(id int) {
				deviceName := fmt.Sprintf("mixed-q%d", id)
				handleNum := id + 10
				handle := fmt.Sprintf("%d:0", handleNum)
				defaultClass := fmt.Sprintf("%d:999", handleNum)
				err := service.CreateHTBQdisc(ctx, deviceName, handle, defaultClass)
				errors <- err
			}(i)
		}

		// Class operations
		for i := 0; i < opsPerType; i++ {
			go func(id int) {
				classID := fmt.Sprintf("1:%d", id+20)
				rate := fmt.Sprintf("%dMbps", (id+1)*10)
				ceil := fmt.Sprintf("%dMbps", (id+1)*20)
				err := service.CreateHTBClass(ctx, baseDevice, "1:0", classID, rate, ceil)
				errors <- err
			}(i)
		}

		// Filter operations
		for i := 0; i < opsPerType; i++ {
			go func(id int) {
				priority := uint16(id + 1)
				match := map[string]string{"dst_port": fmt.Sprintf("%d", 8000+id)}
				err := service.CreateFilter(ctx, baseDevice, "1:0", priority, "ip", "1:10", match)
				errors <- err
			}(i)
		}

		// Collect results
		var errorCount int
		for i := 0; i < totalOps; i++ {
			if err := <-errors; err != nil {
				errorCount++
				t.Logf("Mixed concurrent operation failed: %v", err)
			}
		}

		// Allow for some errors due to concurrency (version conflicts are expected)
		successRate := float64(totalOps-errorCount) / float64(totalOps)
		assert.GreaterOrEqual(t, successRate, 0.6, "At least 60%% of mixed concurrent operations should succeed")
		t.Logf("Mixed concurrent operations success rate: %.2f%% (%d/%d)", successRate*100, totalOps-errorCount, totalOps)
		t.Log("Note: Concurrency conflicts are expected behavior in event sourcing systems")
	})
}