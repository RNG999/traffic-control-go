package projections

import (
	"context"
	"fmt"
	"sync"

	"github.com/rng999/traffic-control-go/internal/domain/events"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/pkg/logging"
)

// Manager manages multiple projections
type Manager struct {
	projections []Projection
	eventStore  eventstore.EventStore
	logger      logging.Logger
	mu          sync.RWMutex
}

// NewManager creates a new projection manager
func NewManager(eventStore eventstore.EventStore) *Manager {
	return &Manager{
		projections: make([]Projection, 0),
		eventStore:  eventStore,
		logger:      logging.WithComponent("projection.manager"),
	}
}

// Register registers a projection
func (m *Manager) Register(projection Projection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.projections = append(m.projections, projection)
	m.logger.Info("Registered projection",
		logging.String("name", projection.GetName()))
}

// ProcessEvent sends an event to all registered projections
func (m *Manager) ProcessEvent(ctx context.Context, event events.DomainEvent) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errs []error
	for _, projection := range m.projections {
		if err := projection.Handle(ctx, event); err != nil {
			m.logger.Error("Projection failed to handle event",
				logging.String("projection", projection.GetName()),
				logging.String("event_type", fmt.Sprintf("%T", event)),
				logging.Error(err))
			errs = append(errs, fmt.Errorf("projection %s: %w", projection.GetName(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("projection errors: %v", errs)
	}

	return nil
}

// RebuildProjections rebuilds all projections from the event store
func (m *Manager) RebuildProjections(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.logger.Info("Starting projection rebuild")

	// Reset all projections
	for _, projection := range m.projections {
		if err := projection.Reset(ctx); err != nil {
			return fmt.Errorf("failed to reset projection %s: %w", projection.GetName(), err)
		}
		m.logger.Debug("Reset projection",
			logging.String("name", projection.GetName()))
	}

	// Get all events from event store
	allEvents, err := m.eventStore.GetAllEvents()
	if err != nil {
		return fmt.Errorf("failed to get all events: %w", err)
	}

	m.logger.Info("Processing events for rebuild",
		logging.Int("event_count", len(allEvents)))

	// Process each event through all projections
	for i, event := range allEvents {
		if err := m.ProcessEvent(ctx, event); err != nil {
			return fmt.Errorf("failed to process event %d: %w", i, err)
		}
	}

	m.logger.Info("Projection rebuild completed successfully",
		logging.Int("events_processed", len(allEvents)),
		logging.Int("projections_rebuilt", len(m.projections)))

	return nil
}

// GetProjections returns the list of registered projections
func (m *Manager) GetProjections() []Projection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Projection, len(m.projections))
	copy(result, m.projections)
	return result
}
