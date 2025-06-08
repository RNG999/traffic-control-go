package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	"github.com/rng999/traffic-control-go/internal/queries/models"
	"github.com/rng999/traffic-control-go/pkg/logging"
)

// TestQueryHandlerRegistration tests that query handlers are properly registered
func TestQueryHandlerRegistration(t *testing.T) {
	// Setup test dependencies
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	mockAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("query-test")

	// Create service with query handlers
	service := NewTrafficControlService(eventStore, mockAdapter, logger)

	t.Run("Query Bus Has Registered Handlers", func(t *testing.T) {
		// Check that query bus was initialized and has handlers
		assert.NotNil(t, service.queryBus, "Query bus should be initialized")
		
		// Query bus should have handlers registered
		assert.Contains(t, service.queryBus.handlers, "GetQdisc", "GetQdisc handler should be registered")
		assert.Contains(t, service.queryBus.handlers, "GetClass", "GetClass handler should be registered") 
		assert.Contains(t, service.queryBus.handlers, "GetFilter", "GetFilter handler should be registered")
		assert.Contains(t, service.queryBus.handlers, "GetConfiguration", "GetConfiguration handler should be registered")
		assert.Contains(t, service.queryBus.handlers, "GetDeviceStatistics", "GetDeviceStatistics handler should be registered")
	})

	t.Run("Query Execution Without Errors", func(t *testing.T) {
		ctx := context.Background()
		deviceName := "test-eth0"

		// Create a device value object
		device, err := valueobjects.NewDevice(deviceName)
		require.NoError(t, err, "Device name should be valid")

		// Test GetConfiguration query - should not error even with no data
		query := models.NewGetTrafficControlConfigQuery(device)
		result, err := service.queryBus.Execute(ctx, "GetConfiguration", query)
		
		// Should execute without error (may return empty result)
		assert.NoError(t, err, "GetConfiguration query should execute without error")
		assert.NotNil(t, result, "GetConfiguration should return a result")
		
		// Result should be the correct type
		_, ok := result.(models.TrafficControlConfigView)
		assert.True(t, ok, "Result should be TrafficControlConfigView type")
	})
}

// TestQueryFunctionality tests end-to-end query functionality
func TestQueryFunctionality(t *testing.T) {
	// Setup test dependencies
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	mockAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("query-func-test")

	// Create service
	service := NewTrafficControlService(eventStore, mockAdapter, logger)
	ctx := context.Background()

	t.Run("GetConfiguration After HTB Creation", func(t *testing.T) {
		deviceName := "func-test-eth0"

		// Create an HTB qdisc first
		err := service.CreateHTBQdisc(ctx, deviceName, "1:0", "1:999")
		require.NoError(t, err, "HTB qdisc creation should succeed")

		// Now query the configuration
		config, err := service.GetConfiguration(ctx, deviceName)
		assert.NoError(t, err, "GetConfiguration should work after creating qdisc")
		assert.NotNil(t, config, "Configuration should not be nil")
		assert.Equal(t, deviceName, config.DeviceName, "Device name should match")
		
		// Should have at least one qdisc
		assert.GreaterOrEqual(t, len(config.Qdiscs), 1, "Should have at least one qdisc")
	})

	t.Run("Statistics Query Integration", func(t *testing.T) {
		deviceName := "stats-test-eth0"

		// Test device statistics query (should work even with no configuration)
		stats, err := service.GetDeviceStatistics(ctx, deviceName)
		
		// With mock adapter, this should succeed
		assert.NoError(t, err, "GetDeviceStatistics should work with mock adapter")
		assert.NotNil(t, stats, "Statistics should not be nil")
		assert.Equal(t, deviceName, stats.DeviceName, "Device name should match")
	})
}