package eventstore

import (
	"context"

	"github.com/rng999/traffic-control-go/internal/domain/events"
)

// EventStore defines the interface for event storage
type EventStore interface {
	// Save saves events for an aggregate with optimistic concurrency control
	Save(aggregateID string, events []events.DomainEvent, expectedVersion int) error

	// GetEvents retrieves all events for an aggregate
	GetEvents(aggregateID string) ([]events.DomainEvent, error)

	// GetEventsFromVersion retrieves events starting from a specific version
	GetEventsFromVersion(aggregateID string, fromVersion int) ([]events.DomainEvent, error)

	// GetAllEvents returns all events in the store (for projections)
	GetAllEvents() ([]events.DomainEvent, error)
}

// EventStoreWithContext extends EventStore with context support
type EventStoreWithContext interface {
	EventStore
	// Load loads an aggregate from the event store
	Load(ctx context.Context, aggregateID string, aggregate EventSourcedAggregate) error
	// SaveAggregate saves an aggregate to the event store (renamed to avoid conflict)
	SaveAggregate(ctx context.Context, aggregate EventSourcedAggregate) error
	// GetEventsWithContext gets events with context (renamed to avoid conflict)
	GetEventsWithContext(ctx context.Context, aggregateID string, fromVersion int, maxEvents int) ([]interface{}, error)
}

// EventSourcedAggregate is implemented by aggregates that use event sourcing
type EventSourcedAggregate interface {
	GetID() string
	GetUncommittedEvents() []events.DomainEvent
	MarkEventsAsCommitted()
	LoadFromHistory(events []events.DomainEvent)
	GetVersion() int
}
