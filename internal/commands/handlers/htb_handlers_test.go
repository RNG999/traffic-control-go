package handlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/internal/commands/models"
	"github.com/rng999/traffic-control-go/internal/domain/aggregates"
	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

func TestCreateHTBQdiscHandler_TypeSafe(t *testing.T) {
	// Setup
	store := eventstore.NewMemoryEventStoreWithContext()
	handler := NewCreateHTBQdiscHandler(store)
	ctx := context.Background()

	// Create command
	cmd := &models.CreateHTBQdiscCommand{
		DeviceName:   "eth0",
		Handle:       "1:0",
		DefaultClass: "1:30",
	}

	// Execute handler
	err := handler.HandleTyped(ctx, cmd)
	
	// Verify success
	assert.NoError(t, err)

	// Verify aggregate was saved by loading it
	deviceName, _ := tc.NewDeviceName("eth0")
	aggregate := aggregates.NewTrafficControlAggregate(deviceName)
	err = store.Load(ctx, aggregate.GetID(), aggregate)
	require.NoError(t, err)

	// Verify aggregate state
	assert.Len(t, aggregate.GetQdiscs(), 1)
	assert.Equal(t, 1, aggregate.Version())

	// Verify qdisc details
	qdiscs := aggregate.GetQdiscs()
	handle := tc.NewHandle(1, 0)
	qdisc, exists := qdiscs[handle]
	assert.True(t, exists)
	assert.NotNil(t, qdisc)
}

func TestCreateHTBClassHandler_TypeSafe(t *testing.T) {
	// Setup
	store := eventstore.NewMemoryEventStoreWithContext()
	qdiscHandler := NewCreateHTBQdiscHandler(store)
	classHandler := NewCreateHTBClassHandler(store)
	ctx := context.Background()

	// First create a qdisc
	qdiscCmd := &models.CreateHTBQdiscCommand{
		DeviceName:   "eth0",
		Handle:       "1:0",
		DefaultClass: "1:30",
	}

	err := qdiscHandler.HandleTyped(ctx, qdiscCmd)
	require.NoError(t, err)

	// Create class command
	classCmd := &models.CreateHTBClassCommand{
		DeviceName: "eth0",
		Parent:     "1:0",
		ClassID:    "1:10",
		Rate:       "100Mbps",
		Ceil:       "200Mbps",
	}

	// Execute handler
	err = classHandler.HandleTyped(ctx, classCmd)
	
	// Verify success
	assert.NoError(t, err)

	// Verify aggregate state by loading it
	deviceName, _ := tc.NewDeviceName("eth0")
	aggregate := aggregates.NewTrafficControlAggregate(deviceName)
	err = store.Load(ctx, aggregate.GetID(), aggregate)
	require.NoError(t, err)

	// Verify aggregate state
	assert.Len(t, aggregate.GetQdiscs(), 1)
	assert.Len(t, aggregate.GetClasses(), 1)
	assert.Equal(t, 2, aggregate.Version()) // qdisc + class

	// Verify class was created (exact verification depends on domain logic)
	// The important part is that no error occurred during handling
	assert.True(t, aggregate.Version() >= 2)
}

func TestHandlers_ErrorPropagation(t *testing.T) {
	// Setup
	store := eventstore.NewMemoryEventStoreWithContext()
	handler := NewCreateHTBQdiscHandler(store)
	ctx := context.Background()

	// Test invalid device name
	cmd := &models.CreateHTBQdiscCommand{
		DeviceName:   "", // Invalid empty name
		Handle:       "1:0",
		DefaultClass: "1:30",
	}

	err := handler.HandleTyped(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid device name")

	// Test invalid handle format
	cmd = &models.CreateHTBQdiscCommand{
		DeviceName:   "eth0",
		Handle:       "invalid", // Invalid format
		DefaultClass: "1:30",
	}

	err = handler.HandleTyped(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid handle format")
}

func TestHandlers_StatePersistence(t *testing.T) {
	// Setup
	store := eventstore.NewMemoryEventStoreWithContext()
	handler := NewCreateHTBQdiscHandler(store)
	ctx := context.Background()

	// Create initial aggregate manually
	deviceName, _ := tc.NewDeviceName("eth0")
	original := aggregates.NewTrafficControlAggregate(deviceName)

	// Save initial state
	err := store.SaveAggregate(ctx, original)
	require.NoError(t, err)

	// Create command
	cmd := &models.CreateHTBQdiscCommand{
		DeviceName:   "eth0",
		Handle:       "1:0",
		DefaultClass: "1:30",
	}

	// Execute handler
	err = handler.HandleTyped(ctx, cmd)
	require.NoError(t, err)

	// Load the updated aggregate
	updated := aggregates.NewTrafficControlAggregate(deviceName)
	err = store.Load(ctx, updated.GetID(), updated)
	require.NoError(t, err)

	// Verify the aggregate was updated correctly
	assert.Len(t, updated.GetQdiscs(), 1)
	assert.Equal(t, 1, updated.Version())
}

func TestMultipleOperations_SequentialExecution(t *testing.T) {
	// Setup
	store := eventstore.NewMemoryEventStoreWithContext()
	qdiscHandler := NewCreateHTBQdiscHandler(store)
	classHandler := NewCreateHTBClassHandler(store)
	ctx := context.Background()

	// Create qdisc
	qdiscCmd := &models.CreateHTBQdiscCommand{
		DeviceName:   "eth0",
		Handle:       "1:0",
		DefaultClass: "1:30",
	}

	// Create two classes
	class1Cmd := &models.CreateHTBClassCommand{
		DeviceName: "eth0",
		Parent:     "1:0",
		ClassID:    "1:10",
		Rate:       "100Mbps",
		Ceil:       "200Mbps",
	}

	class2Cmd := &models.CreateHTBClassCommand{
		DeviceName: "eth0",
		Parent:     "1:0",
		ClassID:    "1:20",
		Rate:       "50Mbps",
		Ceil:       "100Mbps",
	}

	// Execute operations in sequence
	err1 := qdiscHandler.HandleTyped(ctx, qdiscCmd)
	require.NoError(t, err1)

	err2 := classHandler.HandleTyped(ctx, class1Cmd)
	require.NoError(t, err2)

	err3 := classHandler.HandleTyped(ctx, class2Cmd)
	require.NoError(t, err3)

	// Verify final state
	deviceName, _ := tc.NewDeviceName("eth0")
	aggregate := aggregates.NewTrafficControlAggregate(deviceName)
	err := store.Load(ctx, aggregate.GetID(), aggregate)
	require.NoError(t, err)

	assert.Len(t, aggregate.GetQdiscs(), 1)
	assert.Len(t, aggregate.GetClasses(), 2)
	assert.Equal(t, 3, aggregate.Version()) // qdisc + 2 classes
}

func TestCreateFilterHandler_PortMatching(t *testing.T) {
	// Setup
	store := eventstore.NewMemoryEventStoreWithContext()
	handler := NewCreateFilterHandler(store)
	ctx := context.Background()

	// Create initial qdisc and class
	deviceName, _ := tc.NewDeviceName("eth0")
	aggregate := aggregates.NewTrafficControlAggregate(deviceName)
	
	// Add qdisc first
	qHandle, _ := tc.ParseHandle("1:0")
	defaultHandle, _ := tc.ParseHandle("1:30")
	err := aggregate.AddHTBQdisc(qHandle, defaultHandle)
	require.NoError(t, err)
	
	// Add class
	classHandle, _ := tc.ParseHandle("1:10")
	bandwidth := tc.MustParseBandwidth("100Mbps")
	err = aggregate.AddHTBClass(qHandle, classHandle, "testclass", bandwidth, bandwidth)
	require.NoError(t, err)
	
	// Save initial state
	err = store.SaveAggregate(ctx, aggregate)
	require.NoError(t, err)

	t.Run("Destination Port Filter", func(t *testing.T) {
		// Create filter command with destination port
		cmd := &models.CreateFilterCommand{
			DeviceName: "eth0",
			Parent:     "1:0",
			Priority:   100,
			Protocol:   "ip",
			FlowID:     "1:10",
			Match: map[string]string{
				"dst_port": "5201",
			},
		}

		err := handler.HandleTyped(ctx, cmd)
		if err != nil {
			t.Fatalf("Filter creation failed: %v", err)
		}

		// Verify filter was created with correct match
		aggregate := aggregates.NewTrafficControlAggregate(deviceName)
		err = store.Load(ctx, aggregate.GetID(), aggregate)
		require.NoError(t, err)

		filters := aggregate.GetFilters()
		require.Len(t, filters, 1)
		
		filter := filters[0]
		matches := filter.Matches()
		require.Len(t, matches, 1)
		
		// Verify it's a destination port match for port 5201
		portMatch, ok := matches[0].(*entities.PortMatch)
		require.True(t, ok, "Match should be a PortMatch")
		assert.Equal(t, entities.MatchTypePortDestination, portMatch.Type())
		assert.Equal(t, uint16(5201), portMatch.Port())
	})

	t.Run("Source Port Filter", func(t *testing.T) {
		// Create filter command with source port
		cmd := &models.CreateFilterCommand{
			DeviceName: "eth0",
			Parent:     "1:0",
			Priority:   200,
			Protocol:   "ip", 
			FlowID:     "1:10",
			Match: map[string]string{
				"src_port": "8080",
			},
		}

		err := handler.HandleTyped(ctx, cmd)
		require.NoError(t, err)

		// Verify filter was created
		aggregate := aggregates.NewTrafficControlAggregate(deviceName)
		err = store.Load(ctx, aggregate.GetID(), aggregate)
		require.NoError(t, err)

		filters := aggregate.GetFilters()
		require.Len(t, filters, 2) // Previous + this one
		
		// Find the source port filter
		var sourceFilter *entities.Filter
		for _, f := range filters {
			if f.ID().Priority() == 200 {
				sourceFilter = f
				break
			}
		}
		require.NotNil(t, sourceFilter, "Source port filter should exist")
		
		matches := sourceFilter.Matches()
		require.Len(t, matches, 1)
		
		portMatch, ok := matches[0].(*entities.PortMatch)
		require.True(t, ok)
		assert.Equal(t, entities.MatchTypePortSource, portMatch.Type())
		assert.Equal(t, uint16(8080), portMatch.Port())
	})

	t.Run("Multiple Match Conditions", func(t *testing.T) {
		// Create filter with both IP and port
		cmd := &models.CreateFilterCommand{
			DeviceName: "eth0",
			Parent:     "1:0",
			Priority:   300,
			Protocol:   "ip",
			FlowID:     "1:10",
			Match: map[string]string{
				"dst_port": "443",
				"dst_ip":   "192.168.1.100",
			},
		}

		err := handler.HandleTyped(ctx, cmd)
		require.NoError(t, err)

		// Verify both matches were created
		aggregate := aggregates.NewTrafficControlAggregate(deviceName)
		err = store.Load(ctx, aggregate.GetID(), aggregate)
		require.NoError(t, err)

		filters := aggregate.GetFilters()
		require.Len(t, filters, 3)
		
		// Find the multi-match filter
		var multiFilter *entities.Filter
		for _, f := range filters {
			if f.ID().Priority() == 300 {
				multiFilter = f
				break
			}
		}
		require.NotNil(t, multiFilter)
		
		matches := multiFilter.Matches()
		require.Len(t, matches, 2) // Port + IP match
		
		// Verify we have both types
		var hasPortMatch, hasIPMatch bool
		for _, match := range matches {
			switch match.Type() {
			case entities.MatchTypePortDestination:
				hasPortMatch = true
				portMatch := match.(*entities.PortMatch)
				assert.Equal(t, uint16(443), portMatch.Port())
			case entities.MatchTypeIPDestination:
				hasIPMatch = true
			}
		}
		assert.True(t, hasPortMatch, "Should have port match")
		assert.True(t, hasIPMatch, "Should have IP match")
	})

	t.Run("Invalid Port Number", func(t *testing.T) {
		// Create filter with invalid port
		cmd := &models.CreateFilterCommand{
			DeviceName: "eth0",
			Parent:     "1:0",
			Priority:   400,
			Protocol:   "ip",
			FlowID:     "1:10",
			Match: map[string]string{
				"dst_port": "99999", // > 65535
			},
		}

		err := handler.HandleTyped(ctx, cmd)
		require.NoError(t, err) // Should not fail, just skip invalid match

		// Verify filter was created but without port match
		aggregate := aggregates.NewTrafficControlAggregate(deviceName)
		err = store.Load(ctx, aggregate.GetID(), aggregate)
		require.NoError(t, err)

		filters := aggregate.GetFilters()
		require.Len(t, filters, 4)
		
		// Find the filter
		var invalidPortFilter *entities.Filter
		for _, f := range filters {
			if f.ID().Priority() == 400 {
				invalidPortFilter = f
				break
			}
		}
		require.NotNil(t, invalidPortFilter)
		
		// Should have no matches due to invalid port
		matches := invalidPortFilter.Matches()
		assert.Len(t, matches, 0)
	})
}