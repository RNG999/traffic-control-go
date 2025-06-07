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

func TestCreateTBFQdiscHandler(t *testing.T) {
	// Setup
	store := eventstore.NewMemoryEventStoreWithContext()
	handler := NewCreateTBFQdiscHandler(store)
	ctx := context.Background()

	// Create command
	cmd := &models.CreateTBFQdiscCommand{
		DeviceName: "eth0",
		Handle:     "1:0",
		Rate:       "100Mbps",
		Buffer:     1000,
		Limit:      2000,
		Burst:      1500,
	}

	// Execute handler
	err := handler.Handle(ctx, cmd)

	// Verify success
	assert.NoError(t, err)

	// Load aggregate to verify state
	deviceName, _ := tc.NewDeviceName("eth0")
	aggregate := aggregates.NewTrafficControlAggregate(deviceName)
	err = store.Load(ctx, aggregate.GetID(), aggregate)
	require.NoError(t, err)

	// Verify qdisc was added
	qdiscs := aggregate.GetQdiscs()
	assert.Len(t, qdiscs, 1)

	handle := tc.NewHandle(1, 0)
	qdisc, exists := qdiscs[handle]
	assert.True(t, exists)
	assert.NotNil(t, qdisc)
}

func TestCreateTBFQdiscHandler_InvalidCommand(t *testing.T) {
	store := eventstore.NewMemoryEventStoreWithContext()
	handler := NewCreateTBFQdiscHandler(store)
	ctx := context.Background()

	// Test invalid command type
	err := handler.Handle(ctx, "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid command type")
}

func TestCreateTBFQdiscHandler_InvalidDevice(t *testing.T) {
	store := eventstore.NewMemoryEventStoreWithContext()
	handler := NewCreateTBFQdiscHandler(store)
	ctx := context.Background()

	cmd := &models.CreateTBFQdiscCommand{
		DeviceName: "", // Invalid empty name
		Handle:     "1:0",
		Rate:       "100Mbps",
		Buffer:     1000,
		Limit:      2000,
		Burst:      1500,
	}

	err := handler.Handle(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid device name")
}

func TestCreateTBFQdiscHandler_InvalidHandle(t *testing.T) {
	store := eventstore.NewMemoryEventStoreWithContext()
	handler := NewCreateTBFQdiscHandler(store)
	ctx := context.Background()

	cmd := &models.CreateTBFQdiscCommand{
		DeviceName: "eth0",
		Handle:     "invalid", // Invalid handle format
		Rate:       "100Mbps",
		Buffer:     1000,
		Limit:      2000,
		Burst:      1500,
	}

	err := handler.Handle(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid handle format")
}

func TestCreatePRIOQdiscHandler(t *testing.T) {
	// Setup
	store := eventstore.NewMemoryEventStoreWithContext()
	handler := NewCreatePRIOQdiscHandler(store)
	ctx := context.Background()

	// Create command
	cmd := &models.CreatePRIOQdiscCommand{
		DeviceName: "eth0",
		Handle:     "1:0",
		Bands:      3,
		Priomap:    []uint8{1, 2, 2, 2, 1, 2, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1},
	}

	// Execute handler
	err := handler.Handle(ctx, cmd)

	// Verify success
	assert.NoError(t, err)

	// Load aggregate to verify state
	deviceName, _ := tc.NewDeviceName("eth0")
	aggregate := aggregates.NewTrafficControlAggregate(deviceName)
	err = store.Load(ctx, aggregate.GetID(), aggregate)
	require.NoError(t, err)

	// Verify qdisc was added
	qdiscs := aggregate.GetQdiscs()
	assert.Len(t, qdiscs, 1)

	handle := tc.NewHandle(1, 0)
	qdisc, exists := qdiscs[handle]
	assert.True(t, exists)
	assert.NotNil(t, qdisc)
}

func TestCreatePRIOQdiscHandler_InvalidCommand(t *testing.T) {
	store := eventstore.NewMemoryEventStoreWithContext()
	handler := NewCreatePRIOQdiscHandler(store)
	ctx := context.Background()

	// Test invalid command type
	err := handler.Handle(ctx, "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid command type")
}

func TestCreateFQCODELQdiscHandler(t *testing.T) {
	// Setup
	store := eventstore.NewMemoryEventStoreWithContext()
	handler := NewCreateFQCODELQdiscHandler(store)
	ctx := context.Background()

	// Create command
	cmd := &models.CreateFQCODELQdiscCommand{
		DeviceName: "eth0",
		Handle:     "1:0",
		Limit:      10240,
		Flows:      1024,
		Target:     5000,   // 5ms in microseconds
		Interval:   100000, // 100ms in microseconds
		Quantum:    1514,
		ECN:        true,
	}

	// Execute handler
	err := handler.Handle(ctx, cmd)

	// Verify success
	assert.NoError(t, err)

	// Load aggregate to verify state
	deviceName, _ := tc.NewDeviceName("eth0")
	aggregate := aggregates.NewTrafficControlAggregate(deviceName)
	err = store.Load(ctx, aggregate.GetID(), aggregate)
	require.NoError(t, err)

	// Verify qdisc was added
	qdiscs := aggregate.GetQdiscs()
	assert.Len(t, qdiscs, 1)

	handle := tc.NewHandle(1, 0)
	qdisc, exists := qdiscs[handle]
	assert.True(t, exists)
	assert.NotNil(t, qdisc)
}

func TestCreateFQCODELQdiscHandler_InvalidCommand(t *testing.T) {
	store := eventstore.NewMemoryEventStoreWithContext()
	handler := NewCreateFQCODELQdiscHandler(store)
	ctx := context.Background()

	// Test invalid command type
	err := handler.Handle(ctx, "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid command type")
}

func TestCreateFQCODELQdiscHandler_InvalidDevice(t *testing.T) {
	store := eventstore.NewMemoryEventStoreWithContext()
	handler := NewCreateFQCODELQdiscHandler(store)
	ctx := context.Background()

	cmd := &models.CreateFQCODELQdiscCommand{
		DeviceName: "", // Invalid empty name
		Handle:     "1:0",
		Limit:      10240,
		Flows:      1024,
		Target:     5000,
		Interval:   100000,
		Quantum:    1514,
		ECN:        true,
	}

	err := handler.Handle(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid device name")
}

func TestCreateFQCODELQdiscHandler_InvalidHandle(t *testing.T) {
	store := eventstore.NewMemoryEventStoreWithContext()
	handler := NewCreateFQCODELQdiscHandler(store)
	ctx := context.Background()

	cmd := &models.CreateFQCODELQdiscCommand{
		DeviceName: "eth0",
		Handle:     "invalid", // Invalid handle format
		Limit:      10240,
		Flows:      1024,
		Target:     5000,
		Interval:   100000,
		Quantum:    1514,
		ECN:        true,
	}

	err := handler.Handle(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid handle format")
}
