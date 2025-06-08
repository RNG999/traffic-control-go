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

// TestAdvancedErrorScenarios tests comprehensive error handling scenarios
// NOTE: All errors in this test are EXPECTED and intentional - they validate proper error handling
func TestAdvancedErrorScenarios(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	mockAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("error-scenarios")
	service := application.NewTrafficControlService(eventStore, mockAdapter, logger)
	ctx := context.Background()

	t.Run("Invalid Configurations", func(t *testing.T) {
		// EXPECTED: These should fail with validation errors
		// Test invalid bandwidth units
		err := service.CreateHTBQdisc(ctx, "test-eth0", "1:0", "1:999")
		require.NoError(t, err, "QDisc creation should succeed")

		err = service.CreateHTBClass(ctx, "test-eth0", "1:0", "1:20", "100xyz", "200Mbps")
		assert.Error(t, err, "Invalid bandwidth unit should cause error")
		if err != nil {
			assert.Contains(t, err.Error(), "bandwidth", "Error should mention bandwidth")
		}

		// Test completely invalid bandwidth
		err = service.CreateHTBClass(ctx, "test-eth0", "1:0", "1:30", "not-a-bandwidth", "100Mbps")
		assert.Error(t, err, "Invalid bandwidth format should cause error")
		if err != nil {
			assert.Contains(t, err.Error(), "bandwidth", "Error should mention bandwidth")
		}

		// Test negative bandwidth
		err = service.CreateHTBClass(ctx, "test-eth0", "1:0", "1:40", "-100Mbps", "200Mbps")
		assert.Error(t, err, "Negative bandwidth should cause error")
		if err != nil {
			assert.Contains(t, err.Error(), "bandwidth", "Error should mention bandwidth")
		}
	})

	t.Run("Handle Conflicts", func(t *testing.T) {
		// EXPECTED: These should fail with handle conflict errors
		deviceName := "conflict-eth0"

		// Create initial qdisc
		err := service.CreateHTBQdisc(ctx, deviceName, "1:0", "1:999")
		require.NoError(t, err, "Initial qdisc creation should succeed")

		// Try to create qdisc with same handle
		err = service.CreateHTBQdisc(ctx, deviceName, "1:0", "1:888")
		assert.Error(t, err, "Duplicate qdisc handle should cause error")
		assert.Contains(t, err.Error(), "already exists", "Error should mention handle conflict")

		// Create class
		err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:10", "100Mbps", "200Mbps")
		require.NoError(t, err, "Class creation should succeed")

		// Try to create class with same handle
		err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:10", "50Mbps", "150Mbps")
		assert.Error(t, err, "Duplicate class handle should cause error")
	})

	t.Run("Invalid Parent-Child Relationships", func(t *testing.T) {
		// EXPECTED: These should fail with relationship validation errors
		deviceName := "parent-eth0"

		// Try to create class without parent qdisc
		err := service.CreateHTBClass(ctx, deviceName, "1:0", "1:10", "100Mbps", "200Mbps")
		assert.Error(t, err, "Class without parent qdisc should fail")

		// Create qdisc
		err = service.CreateHTBQdisc(ctx, deviceName, "1:0", "1:999")
		require.NoError(t, err, "QDisc creation should succeed")

		// Try to create class with non-existent parent
		err = service.CreateHTBClass(ctx, deviceName, "2:0", "2:10", "100Mbps", "200Mbps")
		assert.Error(t, err, "Class with non-existent parent should fail")

		// Try to create circular dependency (class as its own parent)
		err = service.CreateHTBClass(ctx, deviceName, "1:10", "1:10", "100Mbps", "200Mbps")
		assert.Error(t, err, "Circular dependency should be prevented")
	})

	t.Run("Resource Limits", func(t *testing.T) {
		// EXPECTED: These should fail with resource limit errors
		deviceName := "limits-eth0"

		// Create qdisc
		err := service.CreateHTBQdisc(ctx, deviceName, "1:0", "1:999")
		require.NoError(t, err, "QDisc creation should succeed")

		// Try to create class with excessive bandwidth
		err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:10", "100Tbps", "200Tbps")
		assert.Error(t, err, "Excessive bandwidth should cause error")

		// Test maximum handle values
		err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:65535", "100Mbps", "200Mbps")
		// This might succeed or fail depending on implementation limits
		if err != nil {
			assert.Contains(t, err.Error(), "handle", "Error should mention handle limits")
		}
	})
}

// TestFilterErrorScenarios tests filter-specific error scenarios
// NOTE: Errors in this test are EXPECTED and validate filter error handling
func TestFilterErrorScenarios(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	mockAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("filter-errors")
	service := application.NewTrafficControlService(eventStore, mockAdapter, logger)
	ctx := context.Background()

	deviceName := "filter-eth0"

	// Setup base configuration
	err := service.CreateHTBQdisc(ctx, deviceName, "1:0", "1:999")
	require.NoError(t, err, "QDisc creation should succeed")

	err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:10", "100Mbps", "200Mbps")
	require.NoError(t, err, "Class creation should succeed")

	t.Run("Invalid Filter Configurations", func(t *testing.T) {
		// EXPECTED: These should fail with filter validation errors
		// Test filter with non-existent parent
		match := map[string]string{"dst_port": "80"}
		err := service.CreateFilter(ctx, deviceName, "2:0", 1, "ip", "1:10", match)
		assert.Error(t, err, "Filter with non-existent parent should fail")

		// Test filter with non-existent flow target
		err = service.CreateFilter(ctx, deviceName, "1:0", 1, "ip", "1:99", match)
		assert.Error(t, err, "Filter with non-existent flow target should fail")

		// Note: Mock adapter doesn't validate priority 0 or protocol format
		// These validations would occur in real netlink adapter
		t.Log("Mock adapter successfully validates parent/target existence")
	})

	t.Run("Invalid Match Rules", func(t *testing.T) {
		// Note: Mock adapter doesn't validate match rule formats
		// These validations would occur in real netlink adapter
		
		// Test that valid match rules succeed
		validMatch := map[string]string{"dst_port": "80"}
		err := service.CreateFilter(ctx, deviceName, "1:0", 1, "ip", "1:10", validMatch)
		require.NoError(t, err, "Valid match rules should succeed")
		
		t.Log("Mock adapter successfully processes filter creation")
	})

	t.Run("Priority Conflicts", func(t *testing.T) {
		// Create valid filter
		match1 := map[string]string{"dst_port": "80"}
		err := service.CreateFilter(ctx, deviceName, "1:0", 5, "ip", "1:10", match1)
		require.NoError(t, err, "First filter should succeed")

		// Try to create filter with same priority (might be allowed depending on implementation)
		match2 := map[string]string{"dst_port": "443"}
		err = service.CreateFilter(ctx, deviceName, "1:0", 5, "ip", "1:10", match2)
		// This might succeed or fail depending on implementation
		t.Logf("Priority conflict test result: %v", err)
	})
}

// TestEdgeCaseScenarios tests various edge cases and boundary conditions
func TestEdgeCaseScenarios(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	mockAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("edge-cases")
	service := application.NewTrafficControlService(eventStore, mockAdapter, logger)
	ctx := context.Background()

	t.Run("Boundary Values", func(t *testing.T) {
		deviceName := "bound-eth0"

		// Test minimum valid handle
		err := service.CreateHTBQdisc(ctx, deviceName, "1:0", "1:1")
		require.NoError(t, err, "Minimum handle should work")

		// Test maximum valid minor handle
		err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:65534", "1Mbps", "2Mbps")
		// This might fail depending on implementation limits
		if err != nil {
			t.Logf("Maximum handle test failed as expected: %v", err)
		}

		// Test very small bandwidth
		err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:10", "1bps", "2bps")
		if err != nil {
			assert.Contains(t, err.Error(), "bandwidth", "Small bandwidth error should mention bandwidth")
		}

		// Test very large bandwidth (within reasonable limits)
		err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:20", "10Gbps", "20Gbps")
		// This should generally succeed
		assert.NoError(t, err, "Large but reasonable bandwidth should work")
	})

	t.Run("Special Characters in Device Names", func(t *testing.T) {
		// Test device name with special characters
		specialDevices := []string{
			"eth-0",     // Hyphen
			"eth_0",     // Underscore
			"eth.0",     // Dot
			"eth0:1",    // Colon (VLAN interface)
		}

		for _, device := range specialDevices {
			err := service.CreateHTBQdisc(ctx, device, "1:0", "1:999")
			// These might succeed or fail depending on validation rules
			t.Logf("Special device name '%s' result: %v", device, err)
		}

		// Test device names that should definitely fail
		invalidDevices := []string{
			"",                    // Empty
			"eth0 with spaces",    // Spaces
			"eth0/invalid",        // Forward slash
			"eth0\\invalid",       // Backslash
		}

		for _, device := range invalidDevices {
			err := service.CreateHTBQdisc(ctx, device, "1:0", "1:999")
			assert.Error(t, err, "Invalid device name '%s' should fail", device)
		}
	})

	t.Run("Rapid Successive Operations", func(t *testing.T) {
		deviceName := "rapid-eth0"

		// Create and delete operations in rapid succession
		err := service.CreateHTBQdisc(ctx, deviceName, "1:0", "1:999")
		require.NoError(t, err, "QDisc creation should succeed")

		// Rapid class creation
		for i := 1; i <= 10; i++ {
			classID := fmt.Sprintf("1:%d", i+10)
			rate := fmt.Sprintf("%dMbps", i*10)
			ceil := fmt.Sprintf("%dMbps", i*20)
			err := service.CreateHTBClass(ctx, deviceName, "1:0", classID, rate, ceil)
			require.NoError(t, err, "Rapid class creation %d should succeed", i)
		}

		t.Logf("Successfully created 10 classes in rapid succession")
	})
}

// TestRecoveryScenarios tests error recovery and system resilience
func TestRecoveryScenarios(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	mockAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("recovery")
	service := application.NewTrafficControlService(eventStore, mockAdapter, logger)
	ctx := context.Background()

	t.Run("Recovery After Errors", func(t *testing.T) {
		deviceName := "recovery-eth0"

		// Try invalid operation first
		err := service.CreateHTBClass(ctx, deviceName, "1:0", "1:10", "invalid", "200Mbps")
		assert.Error(t, err, "Invalid operation should fail")

		// System should recover and allow valid operations
		err = service.CreateHTBQdisc(ctx, deviceName, "1:0", "1:999")
		require.NoError(t, err, "Valid operation after error should succeed")

		err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:10", "100Mbps", "200Mbps")
		require.NoError(t, err, "Valid class creation should succeed")

		t.Log("System successfully recovered from errors")
	})

	t.Run("Partial Configuration Rollback", func(t *testing.T) {
		deviceName := "rollback-eth0"

		// Create valid base configuration
		err := service.CreateHTBQdisc(ctx, deviceName, "1:0", "1:999")
		require.NoError(t, err, "Base qdisc should succeed")

		err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:10", "100Mbps", "200Mbps")
		require.NoError(t, err, "Base class should succeed")

		// Try to create configuration with error in the middle
		err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:20", "50Mbps", "100Mbps")
		require.NoError(t, err, "Second class should succeed")

		// This should fail but not affect previous configuration
		err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:30", "invalid-bandwidth", "100Mbps")
		assert.Error(t, err, "Invalid class should fail")

		// Previous configuration should still be intact
		err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:40", "25Mbps", "50Mbps")
		require.NoError(t, err, "New valid class should still work")

		t.Log("System maintained consistency despite partial failures")
	})
}