package handlers

import (
	"fmt"
	
	"github.com/rng999/traffic-control-go/internal/commands/models"
	"github.com/rng999/traffic-control-go/internal/domain/aggregates"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/pkg/types"
)

// CreateHTBQdiscHandler handles the creation of HTB qdiscs
type CreateHTBQdiscHandler struct {
	eventStore eventstore.EventStore
}

// NewCreateHTBQdiscHandler creates a new handler
func NewCreateHTBQdiscHandler(eventStore eventstore.EventStore) *CreateHTBQdiscHandler {
	return &CreateHTBQdiscHandler{
		eventStore: eventStore,
	}
}

// Handle processes the CreateHTBQdiscCommand
func (h *CreateHTBQdiscHandler) Handle(cmd *models.CreateHTBQdiscCommand) types.Result[struct{}] {
	// Load aggregate
	aggregateID := fmt.Sprintf("tc:%s", cmd.DeviceName())
	events, err := h.eventStore.GetEvents(aggregateID)
	if err != nil {
		return types.Failure[struct{}](fmt.Errorf("failed to load events: %w", err))
	}
	
	// Reconstruct aggregate from events
	aggregate := aggregates.FromEvents(cmd.DeviceName(), events)
	
	// Execute business logic
	if err := aggregate.AddHTBQdisc(cmd.Handle(), cmd.DefaultClass()); err != nil {
		return types.Failure[struct{}](err)
	}
	
	// Save uncommitted events
	uncommitted := aggregate.GetUncommittedChanges()
	if err := h.eventStore.Save(aggregateID, uncommitted, aggregate.Version()-len(uncommitted)); err != nil {
		return types.Failure[struct{}](fmt.Errorf("failed to save events: %w", err))
	}
	
	return types.Success(struct{}{})
}

// CreateHTBClassHandler handles the creation of HTB classes
type CreateHTBClassHandler struct {
	eventStore eventstore.EventStore
}

// NewCreateHTBClassHandler creates a new handler
func NewCreateHTBClassHandler(eventStore eventstore.EventStore) *CreateHTBClassHandler {
	return &CreateHTBClassHandler{
		eventStore: eventStore,
	}
}

// Handle processes the CreateHTBClassCommand
func (h *CreateHTBClassHandler) Handle(cmd *models.CreateHTBClassCommand) types.Result[struct{}] {
	// Load aggregate
	aggregateID := fmt.Sprintf("tc:%s", cmd.DeviceName())
	events, err := h.eventStore.GetEvents(aggregateID)
	if err != nil {
		return types.Failure[struct{}](fmt.Errorf("failed to load events: %w", err))
	}
	
	// Reconstruct aggregate from events
	aggregate := aggregates.FromEvents(cmd.DeviceName(), events)
	
	// Execute business logic
	if err := aggregate.AddHTBClass(
		cmd.Parent(),
		cmd.Handle(),
		cmd.Name(),
		cmd.Rate(),
		cmd.Ceil(),
	); err != nil {
		return types.Failure[struct{}](err)
	}
	
	// Save uncommitted events
	uncommitted := aggregate.GetUncommittedChanges()
	if err := h.eventStore.Save(aggregateID, uncommitted, aggregate.Version()-len(uncommitted)); err != nil {
		return types.Failure[struct{}](fmt.Errorf("failed to save events: %w", err))
	}
	
	return types.Success(struct{}{})
}

// CreateFilterHandler handles the creation of filters
type CreateFilterHandler struct {
	eventStore eventstore.EventStore
}

// NewCreateFilterHandler creates a new handler
func NewCreateFilterHandler(eventStore eventstore.EventStore) *CreateFilterHandler {
	return &CreateFilterHandler{
		eventStore: eventStore,
	}
}

// Handle processes the CreateFilterCommand
func (h *CreateFilterHandler) Handle(cmd *models.CreateFilterCommand) types.Result[struct{}] {
	// Load aggregate
	aggregateID := fmt.Sprintf("tc:%s", cmd.DeviceName())
	events, err := h.eventStore.GetEvents(aggregateID)
	if err != nil {
		return types.Failure[struct{}](fmt.Errorf("failed to load events: %w", err))
	}
	
	// Reconstruct aggregate from events
	aggregate := aggregates.FromEvents(cmd.DeviceName(), events)
	
	// Execute business logic
	if err := aggregate.AddFilter(
		cmd.Parent(),
		cmd.Priority(),
		cmd.Handle(),
		cmd.FlowID(),
		cmd.Matches(),
	); err != nil {
		return types.Failure[struct{}](err)
	}
	
	// Save uncommitted events
	uncommitted := aggregate.GetUncommittedChanges()
	if err := h.eventStore.Save(aggregateID, uncommitted, aggregate.Version()-len(uncommitted)); err != nil {
		return types.Failure[struct{}](fmt.Errorf("failed to save events: %w", err))
	}
	
	return types.Success(struct{}{})
}