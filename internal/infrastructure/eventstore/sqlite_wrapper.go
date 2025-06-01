package eventstore

import (
	"context"
)

// SQLiteEventStoreWrapper wraps the SQLite event store to provide context support
type SQLiteEventStoreWrapper struct {
	*SQLiteEventStore
}

// NewSQLiteEventStoreWithContext creates a new SQLite event store with context support
func NewSQLiteEventStoreWithContext(dbPath string) (EventStoreWithContext, error) {
	sqliteStore, err := NewSQLiteEventStore(dbPath)
	if err != nil {
		return nil, err
	}

	return &SQLiteEventStoreWrapper{
		SQLiteEventStore: sqliteStore,
	}, nil
}

// Load loads an aggregate from the event store
func (s *SQLiteEventStoreWrapper) Load(ctx context.Context, aggregateID string, aggregate EventSourcedAggregate) error {
	events, err := s.GetEvents(aggregateID)
	if err != nil {
		return err
	}

	if len(events) > 0 {
		aggregate.LoadFromHistory(events)
	}

	return nil
}

// SaveAggregate saves an aggregate to the event store
func (s *SQLiteEventStoreWrapper) SaveAggregate(ctx context.Context, aggregate EventSourcedAggregate) error {
	uncommittedEvents := aggregate.GetUncommittedEvents()
	if len(uncommittedEvents) == 0 {
		return nil // No events to save
	}

	expectedVersion := aggregate.GetVersion() - len(uncommittedEvents)
	if err := s.SQLiteEventStore.Save(aggregate.GetID(), uncommittedEvents, expectedVersion); err != nil {
		return err
	}

	aggregate.MarkEventsAsCommitted()
	return nil
}

// GetEventsWithContext gets events with context
func (s *SQLiteEventStoreWrapper) GetEventsWithContext(ctx context.Context, aggregateID string, fromVersion int, maxEvents int) ([]interface{}, error) {
	events, err := s.SQLiteEventStore.GetEventsFromVersion(aggregateID, fromVersion)
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

// Ensure SQLiteEventStoreWrapper implements EventStoreWithContext
var _ EventStoreWithContext = (*SQLiteEventStoreWrapper)(nil)