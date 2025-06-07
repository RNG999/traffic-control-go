package eventstore_test

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/events"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// =============================================================================
// BENCHMARK TESTS FOR EVENT STORE PERFORMANCE
// =============================================================================

func BenchmarkEventStoreSave(b *testing.B) {
	b.Run("MemoryEventStore", func(b *testing.B) {
		store := eventstore.NewMemoryEventStore()
		event := createTestEvent()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			store.Save("aggregate-1", []events.DomainEvent{event}, i)
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
			store.Save("aggregate-1", []events.DomainEvent{event}, i)
		}
	})
}

func BenchmarkEventStoreRetrieve(b *testing.B) {
	numEvents := 100 // Reduced for benchmark efficiency

	b.Run("MemoryEventStore", func(b *testing.B) {
		store := eventstore.NewMemoryEventStore()
		
		// Populate with test events
		event := createTestEvent()
		for i := 0; i < numEvents; i++ {
			store.Save("aggregate-1", []events.DomainEvent{event}, i)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.GetEvents("aggregate-1")
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
			store.Save("aggregate-1", []events.DomainEvent{event}, i)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.GetEvents("aggregate-1")
		}
	})
}

func BenchmarkEventStoreRetrieveFromVersion(b *testing.B) {
	numEvents := 100

	b.Run("MemoryEventStore", func(b *testing.B) {
		store := eventstore.NewMemoryEventStore()
		
		// Populate with test events
		event := createTestEvent()
		for i := 0; i < numEvents; i++ {
			store.Save("aggregate-1", []events.DomainEvent{event}, i)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.GetEventsFromVersion("aggregate-1", 50)
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
			store.Save("aggregate-1", []events.DomainEvent{event}, i)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.GetEventsFromVersion("aggregate-1", 50)
		}
	})
}

func BenchmarkEventStoreConcurrentReads(b *testing.B) {
	numEvents := 50

	b.Run("MemoryEventStore", func(b *testing.B) {
		store := eventstore.NewMemoryEventStore()
		
		// Populate with test events
		event := createTestEvent()
		for i := 0; i < numEvents; i++ {
			store.Save("aggregate-1", []events.DomainEvent{event}, i)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = store.GetEvents("aggregate-1")
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
			store.Save("aggregate-1", []events.DomainEvent{event}, i)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = store.GetEvents("aggregate-1")
			}
		})
	})
}

func BenchmarkEventStoreConcurrentWrites(b *testing.B) {
	b.Run("MemoryEventStore", func(b *testing.B) {
		store := eventstore.NewMemoryEventStore()
		event := createTestEvent()
		var counter int64
		var mu sync.Mutex

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				mu.Lock()
				counter++
				aggregateID := fmt.Sprintf("aggregate-%d", counter%10) // Distribute across 10 aggregates
				version := int(counter)
				mu.Unlock()
				
				store.Save(aggregateID, []events.DomainEvent{event}, version)
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
		var counter int64
		var mu sync.Mutex

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				mu.Lock()
				counter++
				aggregateID := fmt.Sprintf("aggregate-%d", counter%10) // Distribute across 10 aggregates
				version := int(counter)
				mu.Unlock()
				
				store.Save(aggregateID, []events.DomainEvent{event}, version)
			}
		})
	})
}

func BenchmarkEventStoreGetAllEvents(b *testing.B) {
	numAggregates := 5
	numEventsPerAggregate := 10

	b.Run("MemoryEventStore", func(b *testing.B) {
		store := eventstore.NewMemoryEventStore()
		
		// Populate with test events across multiple aggregates
		event := createTestEvent()
		for aggIdx := 0; aggIdx < numAggregates; aggIdx++ {
			aggregateID := fmt.Sprintf("aggregate-%d", aggIdx)
			for eventIdx := 0; eventIdx < numEventsPerAggregate; eventIdx++ {
				store.Save(aggregateID, []events.DomainEvent{event}, eventIdx)
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.GetAllEvents()
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
				store.Save(aggregateID, []events.DomainEvent{event}, eventIdx)
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = store.GetAllEvents()
		}
	})
}

func BenchmarkEventStoreComparison(b *testing.B) {
	event := createTestEvent()

	// Benchmark different operations to compare Memory vs SQLite performance
	operations := []struct {
		name string
		fn   func(store eventstore.EventStore)
	}{
		{
			name: "SingleWrite",
			fn: func(store eventstore.EventStore) {
				store.Save("aggregate-1", []events.DomainEvent{event}, 1)
			},
		},
		{
			name: "SingleRead",
			fn: func(store eventstore.EventStore) {
				store.GetEvents("aggregate-1")
			},
		},
		{
			name: "BatchWrite",
			fn: func(store eventstore.EventStore) {
				for i := 0; i < 10; i++ {
					store.Save("aggregate-1", []events.DomainEvent{event}, i)
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
func createTestEvent() *events.QdiscCreatedEvent {
	device, _ := tc.NewDeviceName("eth0")
	handle := tc.NewHandle(1, 0)
	
	return events.NewQdiscCreatedEvent(
		"test-aggregate-1",
		1,
		device,
		handle,
		entities.QdiscTypeHTB,
		nil,
	)
}