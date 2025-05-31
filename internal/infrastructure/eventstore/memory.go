package eventstore

import (
	"fmt"
	"sync"

	"github.com/rng999/traffic-control-go/internal/domain/events"
)

// MemoryEventStore is an in-memory implementation of event store
type MemoryEventStore struct {
	mu     sync.RWMutex
	events map[string][]events.DomainEvent // aggregateID -> events
}

// NewMemoryEventStore creates a new in-memory event store
func NewMemoryEventStore() *MemoryEventStore {
	return &MemoryEventStore{
		events: make(map[string][]events.DomainEvent),
	}
}

// Save saves events for an aggregate
func (m *MemoryEventStore) Save(aggregateID string, domainEvents []events.DomainEvent, expectedVersion int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get current events for the aggregate
	currentEvents, exists := m.events[aggregateID]
	currentVersion := 0
	if exists {
		currentVersion = len(currentEvents)
	}

	// Check for optimistic concurrency
	if currentVersion != expectedVersion {
		return fmt.Errorf("concurrency conflict: expected version %d but was %d", expectedVersion, currentVersion)
	}

	// Append new events
	if !exists {
		m.events[aggregateID] = make([]events.DomainEvent, 0, len(domainEvents))
	}
	m.events[aggregateID] = append(m.events[aggregateID], domainEvents...)

	return nil
}

// GetEvents retrieves all events for an aggregate
func (m *MemoryEventStore) GetEvents(aggregateID string) ([]events.DomainEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	domainEvents, exists := m.events[aggregateID]
	if !exists {
		return []events.DomainEvent{}, nil
	}

	// Return a copy to prevent external modification
	result := make([]events.DomainEvent, len(domainEvents))
	copy(result, domainEvents)

	return result, nil
}

// GetEventsFromVersion retrieves events starting from a specific version
func (m *MemoryEventStore) GetEventsFromVersion(aggregateID string, fromVersion int) ([]events.DomainEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	allEvents, exists := m.events[aggregateID]
	if !exists {
		return []events.DomainEvent{}, nil
	}

	if fromVersion >= len(allEvents) {
		return []events.DomainEvent{}, nil
	}

	// Return events from the specified version
	result := make([]events.DomainEvent, len(allEvents)-fromVersion)
	copy(result, allEvents[fromVersion:])

	return result, nil
}

// GetAllEvents returns all events in the store (for projections)
func (m *MemoryEventStore) GetAllEvents() ([]events.DomainEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var allEvents []events.DomainEvent
	for _, aggregateEvents := range m.events {
		allEvents = append(allEvents, aggregateEvents...)
	}

	return allEvents, nil
}

// Clear removes all events (for testing)
func (m *MemoryEventStore) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.events = make(map[string][]events.DomainEvent)
}
