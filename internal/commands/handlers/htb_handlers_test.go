package handlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/internal/commands/models"
	"github.com/rng999/traffic-control-go/internal/domain/aggregates"
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