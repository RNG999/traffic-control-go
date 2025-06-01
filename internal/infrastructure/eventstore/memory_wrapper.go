package eventstore

import (
	"context"
)

// MemoryEventStoreWrapper wraps the memory event store to provide context support
type MemoryEventStoreWrapper struct {
	*MemoryEventStore
}

// NewMemoryEventStoreWithContext creates a new memory event store with context support
func NewMemoryEventStoreWithContext() EventStoreWithContext {
	return &MemoryEventStoreWrapper{
		MemoryEventStore: NewMemoryEventStore(),
	}
}

// Load loads an aggregate from the event store
func (m *MemoryEventStoreWrapper) Load(ctx context.Context, aggregateID string, aggregate EventSourcedAggregate) error {
	events, err := m.MemoryEventStore.GetEvents(aggregateID)
	if err != nil {
		return err
	}

	if len(events) > 0 {
		aggregate.LoadFromHistory(events)
	}

	return nil
}

// SaveAggregate saves an aggregate to the event store
func (m *MemoryEventStoreWrapper) SaveAggregate(ctx context.Context, aggregate EventSourcedAggregate) error {
	uncommittedEvents := aggregate.GetUncommittedEvents()
	if len(uncommittedEvents) == 0 {
		return nil // No events to save
	}

	expectedVersion := aggregate.GetVersion() - len(uncommittedEvents)
	if err := m.MemoryEventStore.Save(aggregate.GetID(), uncommittedEvents, expectedVersion); err != nil {
		return err
	}

	aggregate.MarkEventsAsCommitted()
	return nil
}

// GetEventsWithContext gets events with context
func (m *MemoryEventStoreWrapper) GetEventsWithContext(ctx context.Context, aggregateID string, fromVersion int, maxEvents int) ([]interface{}, error) {
	events, err := m.MemoryEventStore.GetEventsFromVersion(aggregateID, fromVersion)
	if err != nil {
		return nil, err
	}

	// Convert to interface{} slice and apply maxEvents limit
	result := make([]interface{}, 0, len(events))
	for i, event := range events {
		if maxEvents > 0 && i >= maxEvents {
			break
		}
		result = append(result, event)
	}

	return result, nil
}

// Ensure MemoryEventStoreWrapper implements EventStoreWithContext
var _ EventStoreWithContext = (*MemoryEventStoreWrapper)(nil)