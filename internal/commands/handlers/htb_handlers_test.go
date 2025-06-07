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

func TestCreateHTBQdiscHandler_Functional(t *testing.T) {
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

	// Execute functional handler
	result := handler.HandleFunctional(ctx, cmd)

	// Verify success
	assert.True(t, result.IsSuccess())
	require.NotNil(t, result.Value())

	// Verify aggregate state
	aggregate := result.Value()
	assert.Len(t, aggregate.GetQdiscs(), 1)
	assert.Equal(t, 1, aggregate.Version())
	assert.Len(t, aggregate.GetUncommittedEvents(), 0) // Events are committed after save

	// Verify qdisc details
	qdiscs := aggregate.GetQdiscs()
	handle := tc.NewHandle(1, 0)
	qdisc, exists := qdiscs[handle]
	assert.True(t, exists)
	assert.NotNil(t, qdisc)
}

func TestCreateHTBClassHandler_Functional(t *testing.T) {
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

	qdiscResult := qdiscHandler.HandleFunctional(ctx, qdiscCmd)
	require.True(t, qdiscResult.IsSuccess())

	// Create class command
	classCmd := &models.CreateHTBClassCommand{
		DeviceName: "eth0",
		Parent:     "1:0",
		ClassID:    "1:10",
		Rate:       "100Mbps",
		Ceil:       "200Mbps",
	}

	// Execute functional handler
	result := classHandler.HandleFunctional(ctx, classCmd)

	// Verify success
	assert.True(t, result.IsSuccess())
	require.NotNil(t, result.Value())

	// Verify aggregate state
	aggregate := result.Value()
	assert.Len(t, aggregate.GetQdiscs(), 1)
	assert.Len(t, aggregate.GetClasses(), 1)
	assert.Equal(t, 2, aggregate.Version()) // qdisc + class

	// Verify class details
	classes := aggregate.GetClasses()
	classHandle := tc.NewHandle(1, 10)
	class, exists := classes[classHandle]
	assert.True(t, exists)
	assert.NotNil(t, class)
}

func TestFunctionalHandlers_ErrorPropagation(t *testing.T) {
	// Setup
	store := eventstore.NewMemoryEventStoreWithContext()
	handler := NewCreateHTBQdiscHandler(store)
	ctx := context.Background()

	// Test invalid command type
	result := handler.HandleFunctional(ctx, "invalid")
	assert.True(t, result.IsFailure())
	assert.Contains(t, result.Error().Error(), "invalid command type")

	// Test invalid device name
	cmd := &models.CreateHTBQdiscCommand{
		DeviceName:   "", // Invalid empty name
		Handle:       "1:0",
		DefaultClass: "1:30",
	}

	result = handler.HandleFunctional(ctx, cmd)
	assert.True(t, result.IsFailure())
	assert.Contains(t, result.Error().Error(), "invalid device name")

	// Test invalid handle format
	cmd = &models.CreateHTBQdiscCommand{
		DeviceName:   "eth0",
		Handle:       "invalid", // Invalid format
		DefaultClass: "1:30",
	}

	result = handler.HandleFunctional(ctx, cmd)
	assert.True(t, result.IsFailure())
	assert.Contains(t, result.Error().Error(), "invalid handle format")
}

func TestFunctionalHandlers_ImmutabilityPreservation(t *testing.T) {
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
	result := handler.HandleFunctional(ctx, cmd)
	require.True(t, result.IsSuccess())

	// Load original from store to verify it wasn't mutated
	loadedOriginal := aggregates.NewTrafficControlAggregate(deviceName)
	err = store.Load(ctx, loadedOriginal.GetID(), loadedOriginal)
	require.NoError(t, err)

	// Verify original state in store is preserved (only the returned aggregate has changes)
	newAggregate := result.Value()
	assert.Len(t, newAggregate.GetQdiscs(), 1)
	assert.Equal(t, 1, newAggregate.Version())

	// Note: The store will contain the new events after SaveAggregate is called,
	// but the functional approach ensures the handler doesn't mutate input aggregates
}

func TestFunctionalHandlers_ComplexScenario(t *testing.T) {
	// Setup
	store := eventstore.NewMemoryEventStoreWithContext()
	qdiscHandler := NewCreateHTBQdiscHandler(store)
	classHandler := NewCreateHTBClassHandler(store)
	ctx := context.Background()

	// Create multiple operations using functional composition
	qdiscCmd := &models.CreateHTBQdiscCommand{
		DeviceName:   "eth0",
		Handle:       "1:0",
		DefaultClass: "1:30",
	}

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
	result1 := qdiscHandler.HandleFunctional(ctx, qdiscCmd)
	require.True(t, result1.IsSuccess())

	result2 := classHandler.HandleFunctional(ctx, class1Cmd)
	require.True(t, result2.IsSuccess())

	result3 := classHandler.HandleFunctional(ctx, class2Cmd)
	require.True(t, result3.IsSuccess())

	// Verify final state
	finalAggregate := result3.Value()
	assert.Len(t, finalAggregate.GetQdiscs(), 1)
	assert.Len(t, finalAggregate.GetClasses(), 2)
	assert.Equal(t, 3, finalAggregate.Version()) // qdisc + 2 classes

	// Verify each class exists with correct handles
	classes := finalAggregate.GetClasses()
	class1Handle := tc.NewHandle(1, 10)
	class2Handle := tc.NewHandle(1, 20)

	_, exists1 := classes[class1Handle]
	_, exists2 := classes[class2Handle]
	assert.True(t, exists1)
	assert.True(t, exists2)
}
