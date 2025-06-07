package handlers

import (
	"context"
	"fmt"

	"github.com/rng999/traffic-control-go/internal/commands/models"
	"github.com/rng999/traffic-control-go/internal/domain/aggregates"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// CreateTBFQdiscHandler handles CreateTBFQdiscCommand with type safety
type CreateTBFQdiscHandler struct {
	eventStore eventstore.EventStoreWithContext
}

// NewCreateTBFQdiscHandler creates a new type-safe TBF handler
func NewCreateTBFQdiscHandler(eventStore eventstore.EventStoreWithContext) *CreateTBFQdiscHandler {
	return &CreateTBFQdiscHandler{
		eventStore: eventStore,
	}
}

// HandleTyped processes the CreateTBFQdiscCommand with compile-time type safety
func (h *CreateTBFQdiscHandler) HandleTyped(ctx context.Context, command *models.CreateTBFQdiscCommand) error {
	// Create device value object
	device, err := tc.NewDeviceName(command.DeviceName)
	if err != nil {
		return fmt.Errorf("invalid device name: %w", err)
	}

	// Load aggregate
	aggregate := aggregates.NewTrafficControlAggregate(device)
	if err := h.eventStore.Load(ctx, aggregate.GetID(), aggregate); err != nil {
		return fmt.Errorf("failed to load aggregate: %w", err)
	}

	// Parse handle
	handle, err := tc.ParseHandle(command.Handle)
	if err != nil {
		return fmt.Errorf("invalid handle format: %w", err)
	}

	// Parse rate
	rate, err := tc.ParseBandwidth(command.Rate)
	if err != nil {
		return fmt.Errorf("invalid rate: %w", err)
	}

	// Execute business logic
	if err := aggregate.AddTBFQdisc(handle, rate, command.Buffer, command.Limit, command.Burst); err != nil {
		return err
	}

	// Save aggregate
	if err := h.eventStore.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate: %w", err)
	}

	return nil
}

// CreatePRIOQdiscHandler handles CreatePRIOQdiscCommand with type safety
type CreatePRIOQdiscHandler struct {
	eventStore eventstore.EventStoreWithContext
}

// NewCreatePRIOQdiscHandler creates a new type-safe PRIO handler
func NewCreatePRIOQdiscHandler(eventStore eventstore.EventStoreWithContext) *CreatePRIOQdiscHandler {
	return &CreatePRIOQdiscHandler{
		eventStore: eventStore,
	}
}

// HandleTyped processes the CreatePRIOQdiscCommand with compile-time type safety
func (h *CreatePRIOQdiscHandler) HandleTyped(ctx context.Context, command *models.CreatePRIOQdiscCommand) error {
	// Create device value object
	device, err := tc.NewDeviceName(command.DeviceName)
	if err != nil {
		return fmt.Errorf("invalid device name: %w", err)
	}

	// Load aggregate
	aggregate := aggregates.NewTrafficControlAggregate(device)
	if err := h.eventStore.Load(ctx, aggregate.GetID(), aggregate); err != nil {
		return fmt.Errorf("failed to load aggregate: %w", err)
	}

	// Parse handle
	handle, err := tc.ParseHandle(command.Handle)
	if err != nil {
		return fmt.Errorf("invalid handle format: %w", err)
	}

	// Execute business logic
	if err := aggregate.AddPRIOQdisc(handle, command.Bands, command.Priomap); err != nil {
		return err
	}

	// Save aggregate
	if err := h.eventStore.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate: %w", err)
	}

	return nil
}

// CreateFQCODELQdiscHandler handles CreateFQCODELQdiscCommand with type safety
type CreateFQCODELQdiscHandler struct {
	eventStore eventstore.EventStoreWithContext
}

// NewCreateFQCODELQdiscHandler creates a new type-safe FQ_CODEL handler
func NewCreateFQCODELQdiscHandler(eventStore eventstore.EventStoreWithContext) *CreateFQCODELQdiscHandler {
	return &CreateFQCODELQdiscHandler{
		eventStore: eventStore,
	}
}

// HandleTyped processes the CreateFQCODELQdiscCommand with compile-time type safety
func (h *CreateFQCODELQdiscHandler) HandleTyped(ctx context.Context, command *models.CreateFQCODELQdiscCommand) error {
	// Create device value object
	device, err := tc.NewDeviceName(command.DeviceName)
	if err != nil {
		return fmt.Errorf("invalid device name: %w", err)
	}

	// Load aggregate
	aggregate := aggregates.NewTrafficControlAggregate(device)
	if err := h.eventStore.Load(ctx, aggregate.GetID(), aggregate); err != nil {
		return fmt.Errorf("failed to load aggregate: %w", err)
	}

	// Parse handle
	handle, err := tc.ParseHandle(command.Handle)
	if err != nil {
		return fmt.Errorf("invalid handle format: %w", err)
	}

	// Execute business logic
	if err := aggregate.AddFQCODELQdisc(handle, command.Limit, command.Flows, command.Target, command.Interval, command.Quantum, command.ECN); err != nil {
		return err
	}

	// Save aggregate
	if err := h.eventStore.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate: %w", err)
	}

	return nil
}
