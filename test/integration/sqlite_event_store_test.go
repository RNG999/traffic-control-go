package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/internal/domain/aggregates"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
)

func TestSQLiteEventStore(t *testing.T) {
	t.Skip("Skipping SQLite event store test - implementation pending")
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_events.db")

	// Create event store
	store, err := eventstore.NewSQLiteEventStoreWithContext(dbPath)
	require.NoError(t, err)
	defer func() {
		if sqliteStore, ok := store.(*eventstore.SQLiteEventStoreWrapper); ok {
			if err := sqliteStore.Close(); err != nil {
				t.Logf("Failed to close SQLite store: %v", err)
			}
		}
	}()

	ctx := context.Background()

	t.Run("save and load aggregate", func(t *testing.T) {
		// Create aggregate
		device, err := valueobjects.NewDevice("eth0")
		require.NoError(t, err)
		
		aggregate := aggregates.NewTrafficControlAggregate(device)
		
		// Add HTB qdisc
		handle := valueobjects.NewHandle(1, 0)
		defaultClass := valueobjects.NewHandle(1, 999)
		err = aggregate.AddHTBQdisc(handle, defaultClass)
		require.NoError(t, err)

		// Save aggregate
		err = store.SaveAggregate(ctx, aggregate)
		require.NoError(t, err)

		// Load aggregate
		newAggregate := aggregates.NewTrafficControlAggregate(device)
		err = store.Load(ctx, aggregate.GetID(), newAggregate)
		require.NoError(t, err)

		// Verify
		assert.Equal(t, 1, newAggregate.GetVersion())
		assert.Equal(t, 1, len(newAggregate.GetQdiscs()))
	})

	t.Run("optimistic concurrency control", func(t *testing.T) {
		// Create aggregate
		device, err := valueobjects.NewDevice("eth1")
		require.NoError(t, err)
		
		aggregate1 := aggregates.NewTrafficControlAggregate(device)
		aggregate2 := aggregates.NewTrafficControlAggregate(device)

		// Add qdisc to first aggregate
		handle := valueobjects.NewHandle(1, 0)
		defaultClass := valueobjects.NewHandle(1, 999)
		err = aggregate1.AddHTBQdisc(handle, defaultClass)
		require.NoError(t, err)

		// Save first aggregate
		err = store.SaveAggregate(ctx, aggregate1)
		require.NoError(t, err)

		// Load into second aggregate
		err = store.Load(ctx, aggregate1.GetID(), aggregate2)
		require.NoError(t, err)

		// Modify both aggregates
		parentHandle := valueobjects.NewHandle(1, 0)
		classHandle1 := valueobjects.NewHandle(1, 10)
		classHandle2 := valueobjects.NewHandle(1, 20)
		rate, _ := valueobjects.NewBandwidth("10mbps")
		ceil, _ := valueobjects.NewBandwidth("100mbps")

		err = aggregate1.AddHTBClass(parentHandle, classHandle1, "class1", rate, ceil)
		require.NoError(t, err)

		err = aggregate2.AddHTBClass(parentHandle, classHandle2, "class2", rate, ceil)
		require.NoError(t, err)

		// Save first aggregate - should succeed
		err = store.SaveAggregate(ctx, aggregate1)
		require.NoError(t, err)

		// Save second aggregate - should fail due to version conflict
		err = store.SaveAggregate(ctx, aggregate2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "concurrency conflict")
	})

	t.Run("retrieve events with context", func(t *testing.T) {
		// Create and save aggregate with multiple events
		device, err := valueobjects.NewDevice("eth2")
		require.NoError(t, err)
		
		aggregate := aggregates.NewTrafficControlAggregate(device)
		
		// Add multiple items
		handle := valueobjects.NewHandle(1, 0)
		defaultClass := valueobjects.NewHandle(1, 999)
		err = aggregate.AddHTBQdisc(handle, defaultClass)
		require.NoError(t, err)

		parentHandle := valueobjects.NewHandle(1, 0)
		classHandle := valueobjects.NewHandle(1, 10)
		rate, _ := valueobjects.NewBandwidth("10mbps")
		ceil, _ := valueobjects.NewBandwidth("100mbps")
		err = aggregate.AddHTBClass(parentHandle, classHandle, "class1", rate, ceil)
		require.NoError(t, err)

		// Save aggregate
		err = store.SaveAggregate(ctx, aggregate)
		require.NoError(t, err)

		// Get events with context
		events, err := store.GetEventsWithContext(ctx, aggregate.GetID(), 0, 10)
		require.NoError(t, err)
		assert.Equal(t, 2, len(events))

		// Get only one event
		events, err = store.GetEventsWithContext(ctx, aggregate.GetID(), 1, 1)
		require.NoError(t, err)
		assert.Equal(t, 1, len(events))
	})
}

func TestSQLiteEventStorePersistence(t *testing.T) {
	t.Skip("Skipping SQLite event store persistence test - implementation pending")
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "persist_test.db")

	ctx := context.Background()

	// Create and populate event store
	func() {
		store, err := eventstore.NewSQLiteEventStoreWithContext(dbPath)
		require.NoError(t, err)
		defer func() {
			if sqliteStore, ok := store.(*eventstore.SQLiteEventStoreWrapper); ok {
				if err := sqliteStore.Close(); err != nil {
					// Cannot use t.Logf here as t is not in scope
					_ = err
				}
			}
		}()

		// Create aggregate
		device, err := valueobjects.NewDevice("eth0")
		require.NoError(t, err)
		
		aggregate := aggregates.NewTrafficControlAggregate(device)
		
		// Add HTB qdisc
		handle := valueobjects.NewHandle(1, 0)
		defaultClass := valueobjects.NewHandle(1, 999)
		err = aggregate.AddHTBQdisc(handle, defaultClass)
		require.NoError(t, err)

		// Save aggregate
		err = store.SaveAggregate(ctx, aggregate)
		require.NoError(t, err)
	}()

	// Verify data persists after closing
	store, err := eventstore.NewSQLiteEventStoreWithContext(dbPath)
	require.NoError(t, err)
	defer func() {
		if sqliteStore, ok := store.(*eventstore.SQLiteEventStoreWrapper); ok {
			if err := sqliteStore.Close(); err != nil {
				t.Logf("Failed to close SQLite store: %v", err)
			}
		}
	}()

	// Load aggregate
	device, err := valueobjects.NewDevice("eth0")
	require.NoError(t, err)
	
	aggregate := aggregates.NewTrafficControlAggregate(device)
	err = store.Load(ctx, aggregate.GetID(), aggregate)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, 1, aggregate.GetVersion())
	assert.Equal(t, 1, len(aggregate.GetQdiscs()))
}

func TestSQLiteEventStoreFileCheck(t *testing.T) {
	t.Skip("Skipping SQLite event store file check test - implementation pending")
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "file_test.db")

	// Create event store
	store, err := eventstore.NewSQLiteEventStoreWithContext(dbPath)
	require.NoError(t, err)
	defer func() {
		if sqliteStore, ok := store.(*eventstore.SQLiteEventStoreWrapper); ok {
			if err := sqliteStore.Close(); err != nil {
				t.Logf("Failed to close SQLite store: %v", err)
			}
		}
	}()

	// Check that database file was created
	_, err = os.Stat(dbPath)
	assert.NoError(t, err, "SQLite database file should exist")
}