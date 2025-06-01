package events

import (
	"time"
)

// DomainEvent is the base interface for all domain events
type DomainEvent interface {
	AggregateID() string
	EventType() string
	Timestamp() time.Time
	EventVersion() int
}

// BaseEvent contains common fields for all events
type BaseEvent struct {
	aggregateID string
	eventType   string
	occurredAt  time.Time
	version     int
}

// NewBaseEvent creates a new base event
func NewBaseEvent(aggregateID string, eventType string, version int) BaseEvent {
	return BaseEvent{
		aggregateID: aggregateID,
		eventType:   eventType,
		occurredAt:  time.Now().UTC(),
		version:     version,
	}
}

// AggregateID returns the aggregate ID
func (e BaseEvent) AggregateID() string {
	return e.aggregateID
}

// EventType returns the event type
func (e BaseEvent) EventType() string {
	return e.eventType
}

// Timestamp returns when the event occurred
func (e BaseEvent) Timestamp() time.Time {
	return e.occurredAt
}

// EventVersion returns the event version
func (e BaseEvent) EventVersion() int {
	return e.version
}
