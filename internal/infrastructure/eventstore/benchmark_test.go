package eventstore_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rng999/traffic-control-go/internal/domain/events"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// =============================================================================
// BENCHMARK TESTS FOR EVENT STORE PERFORMANCE
// =============================================================================

func BenchmarkEventStoreSave(b *testing.B) {
	ctx := context.Background()

	b.Run("MemoryEventStore", func(b *testing.B) {
		store := eventstore.NewMemoryEventStore()
		event := createTestEvent()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			store.AppendEvent(ctx, "aggregate-1", int64(i), event)
		}
	})

	b.Run("SQLiteEventStore", func(b *testing.B) {
		// Create temporary database
		tmpFile, err := os.CreateTemp("", "benchmark-*.db")
		if err != nil {
			b.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		store, err := eventstore.NewSQLiteEventStore(tmpFile.Name())
		if err != nil {
			b.Fatal(err)
		}
		defer store.Close()

		event := createTestEvent()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			store.AppendEvent(ctx, "aggregate-1", int64(i), event)
		}
	})
}

func BenchmarkEventStoreRetrieve(b *testing.B) {
	ctx := context.Background()
	numEvents := 1000

	b.Run("MemoryEventStore", func(b *testing.B) {
		store := eventstore.NewMemoryEventStore()
		
		// Populate with test events
		event := createTestEvent()
		for i := 0; i < numEvents; i++ {
			store.AppendEvent(ctx, "aggregate-1", int64(i), event)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.GetEvents(ctx, "aggregate-1")
		}
	})

	b.Run("SQLiteEventStore", func(b *testing.B) {
		// Create temporary database
		tmpFile, err := os.CreateTemp("", "benchmark-*.db")
		if err != nil {
			b.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		store, err := eventstore.NewSQLiteEventStore(tmpFile.Name())
		if err != nil {
			b.Fatal(err)
		}
		defer store.Close()

		// Populate with test events
		event := createTestEvent()
		for i := 0; i < numEvents; i++ {
			store.AppendEvent(ctx, "aggregate-1", int64(i), event)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.GetEvents(ctx, "aggregate-1")
		}
	})
}

func BenchmarkEventStoreRetrieveFromVersion(b *testing.B) {
	ctx := context.Background()
	numEvents := 1000

	b.Run("MemoryEventStore", func(b *testing.B) {
		store := eventstore.NewMemoryEventStore()
		
		// Populate with test events
		event := createTestEvent()
		for i := 0; i < numEvents; i++ {
			store.AppendEvent(ctx, "aggregate-1", int64(i), event)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.GetEventsFromVersion(ctx, "aggregate-1", 500)
		}
	})

	b.Run("SQLiteEventStore", func(b *testing.B) {
		// Create temporary database
		tmpFile, err := os.CreateTemp("", "benchmark-*.db")
		if err != nil {
			b.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		store, err := eventstore.NewSQLiteEventStore(tmpFile.Name())
		if err != nil {
			b.Fatal(err)
		}
		defer store.Close()

		// Populate with test events
		event := createTestEvent()
		for i := 0; i < numEvents; i++ {
			store.AppendEvent(ctx, "aggregate-1", int64(i), event)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.GetEventsFromVersion(ctx, "aggregate-1", 500)
		}
	})
}

func BenchmarkEventStoreConcurrentReads(b *testing.B) {
	ctx := context.Background()
	numEvents := 100

	b.Run("MemoryEventStore", func(b *testing.B) {
		store := eventstore.NewMemoryEventStore()
		
		// Populate with test events
		event := createTestEvent()
		for i := 0; i < numEvents; i++ {
			store.AppendEvent(ctx, "aggregate-1", int64(i), event)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = store.GetEvents(ctx, "aggregate-1")
			}
		})
	})

	b.Run("SQLiteEventStore", func(b *testing.B) {
		// Create temporary database
		tmpFile, err := os.CreateTemp("", "benchmark-*.db")
		if err != nil {
			b.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		store, err := eventstore.NewSQLiteEventStore(tmpFile.Name())
		if err != nil {
			b.Fatal(err)
		}
		defer store.Close()

		// Populate with test events
		event := createTestEvent()
		for i := 0; i < numEvents; i++ {
			store.AppendEvent(ctx, "aggregate-1", int64(i), event)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = store.GetEvents(ctx, "aggregate-1")
			}
		})
	})
}

func BenchmarkEventStoreConcurrentWrites(b *testing.B) {
	ctx := context.Background()

	b.Run("MemoryEventStore", func(b *testing.B) {
		store := eventstore.NewMemoryEventStore()
		event := createTestEvent()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				aggregateID := fmt.Sprintf("aggregate-%d", i%10) // Distribute across 10 aggregates
				store.AppendEvent(ctx, aggregateID, int64(i), event)
				i++
			}
		})
	})

	b.Run("SQLiteEventStore", func(b *testing.B) {
		// Create temporary database
		tmpFile, err := os.CreateTemp("", "benchmark-*.db")
		if err != nil {
			b.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		store, err := eventstore.NewSQLiteEventStore(tmpFile.Name())
		if err != nil {
			b.Fatal(err)
		}
		defer store.Close()

		event := createTestEvent()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				aggregateID := fmt.Sprintf("aggregate-%d", i%10) // Distribute across 10 aggregates
				store.AppendEvent(ctx, aggregateID, int64(i), event)
				i++
			}
		})
	})
}

func BenchmarkEventSerialization(b *testing.B) {
	event := createTestEvent()

	b.Run("QdiscCreatedEvent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// This tests the JSON marshaling performance
			data, _ := event.Serialize()
			_ = data
		}
	})

	b.Run("EventDeserialization", func(b *testing.B) {
		// First serialize the event to get test data
		data, _ := event.Serialize()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Test deserialization performance
			var evt events.QdiscCreatedEvent
			_ = evt.Deserialize(data)
		}
	})
}

func BenchmarkEventStoreGetAllEvents(b *testing.B) {
	ctx := context.Background()
	numAggregates := 10
	numEventsPerAggregate := 100

	b.Run("MemoryEventStore", func(b *testing.B) {
		store := eventstore.NewMemoryEventStore()
		
		// Populate with test events across multiple aggregates
		event := createTestEvent()
		for aggIdx := 0; aggIdx < numAggregates; aggIdx++ {
			aggregateID := fmt.Sprintf("aggregate-%d", aggIdx)
			for eventIdx := 0; eventIdx < numEventsPerAggregate; eventIdx++ {
				store.AppendEvent(ctx, aggregateID, int64(eventIdx), event)
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.GetAllEvents(ctx)
		}
	})

	b.Run("SQLiteEventStore", func(b *testing.B) {
		// Create temporary database
		tmpFile, err := os.CreateTemp("", "benchmark-*.db")
		if err != nil {
			b.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		store, err := eventstore.NewSQLiteEventStore(tmpFile.Name())
		if err != nil {
			b.Fatal(err)
		}
		defer store.Close()

		// Populate with test events across multiple aggregates
		event := createTestEvent()
		for aggIdx := 0; aggIdx < numAggregates; aggIdx++ {
			aggregateID := fmt.Sprintf("aggregate-%d", aggIdx)
			for eventIdx := 0; eventIdx < numEventsPerAggregate; eventIdx++ {
				store.AppendEvent(ctx, aggregateID, int64(eventIdx), event)
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.GetAllEvents(ctx)
		}
	})
}

func BenchmarkEventStoreComparison(b *testing.B) {
	ctx := context.Background()
	event := createTestEvent()

	// Benchmark different operations to compare Memory vs SQLite performance
	operations := []struct {
		name string
		fn   func(store eventstore.EventStore)
	}{
		{
			name: "SingleWrite",
			fn: func(store eventstore.EventStore) {
				store.AppendEvent(ctx, "aggregate-1", 1, event)
			},
		},
		{
			name: "SingleRead",
			fn: func(store eventstore.EventStore) {
				store.GetEvents(ctx, "aggregate-1")
			},
		},
		{
			name: "BatchWrite",
			fn: func(store eventstore.EventStore) {
				for i := 0; i < 10; i++ {
					store.AppendEvent(ctx, "aggregate-1", int64(i), event)
				}
			},
		},
	}

	for _, op := range operations {
		b.Run(fmt.Sprintf("Memory_%s", op.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				store := eventstore.NewMemoryEventStore()
				op.fn(store)
			}
		})

		b.Run(fmt.Sprintf("SQLite_%s", op.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tmpFile, err := os.CreateTemp("", "benchmark-*.db")
				if err != nil {
					b.Fatal(err)
				}
				tmpFile.Close()

				store, err := eventstore.NewSQLiteEventStore(tmpFile.Name())
				if err != nil {
					b.Fatal(err)
				}

				op.fn(store)
				
				store.Close()
				os.Remove(tmpFile.Name())
			}
		})
	}
}

// Helper function to create a test event
func createTestEvent() events.QdiscCreatedEvent {
	device, _ := tc.NewDeviceName("eth0")
	handle := tc.NewHandle(1, 0)
	
	return events.QdiscCreatedEvent{
		BaseEvent: events.BaseEvent{
			EventID:   "test-event-id",
			Timestamp: time.Now(),
			Version:   1,
		},
		Device:     device,
		Handle:     handle,
		QdiscType:  "htb",
		Properties: map[string]interface{}{
			"default": "30",
			"rate":    "1000mbit",
		},
	}
}