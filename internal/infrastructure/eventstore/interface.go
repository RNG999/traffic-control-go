package eventstore

import (
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