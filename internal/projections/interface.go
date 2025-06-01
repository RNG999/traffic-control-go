package projections

import (
	"context"

	"github.com/rng999/traffic-control-go/internal/domain/events"
)

// Projection represents a projection that builds read models from events
type Projection interface {
	// Handle processes an event to update the read model
	Handle(ctx context.Context, event events.DomainEvent) error
	
	// GetName returns the projection name
	GetName() string
	
	// Reset clears the projection state
	Reset(ctx context.Context) error
}

// ProjectionManager manages multiple projections
type ProjectionManager interface {
	// Register registers a projection
	Register(projection Projection)
	
	// ProcessEvent sends an event to all registered projections
	ProcessEvent(ctx context.Context, event events.DomainEvent) error
	
	// RebuildProjections rebuilds all projections from the event store
	RebuildProjections(ctx context.Context) error
}

// ReadModelStore is the interface for storing read models
type ReadModelStore interface {
	// Save saves a read model
	Save(ctx context.Context, collection string, id string, data interface{}) error
	
	// Get retrieves a read model
	Get(ctx context.Context, collection string, id string, result interface{}) error
	
	// Query queries read models
	Query(ctx context.Context, collection string, filter interface{}) ([]interface{}, error)
	
	// Delete deletes a read model
	Delete(ctx context.Context, collection string, id string) error
	
	// Clear clears all data in a collection
	Clear(ctx context.Context, collection string) error
}