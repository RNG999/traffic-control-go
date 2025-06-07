package handlers

import (
	"context"
	"fmt"

	"github.com/rng999/traffic-control-go/internal/commands/models"
	"github.com/rng999/traffic-control-go/internal/domain/aggregates"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// GenericCreateTBFQdiscHandler handles CreateTBFQdiscCommand with type safety
type GenericCreateTBFQdiscHandler struct {
	eventStore eventstore.EventStoreWithContext
}

// NewGenericCreateTBFQdiscHandler creates a new type-safe TBF handler
func NewGenericCreateTBFQdiscHandler(eventStore eventstore.EventStoreWithContext) *GenericCreateTBFQdiscHandler {
	return &GenericCreateTBFQdiscHandler{
		eventStore: eventStore,
	}
}

// HandleTyped processes the CreateTBFQdiscCommand with compile-time type safety
func (h *GenericCreateTBFQdiscHandler) HandleTyped(ctx context.Context, command *models.CreateTBFQdiscCommand) error {
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

// GenericCreatePRIOQdiscHandler handles CreatePRIOQdiscCommand with type safety
type GenericCreatePRIOQdiscHandler struct {
	eventStore eventstore.EventStoreWithContext
}

// NewGenericCreatePRIOQdiscHandler creates a new type-safe PRIO handler
func NewGenericCreatePRIOQdiscHandler(eventStore eventstore.EventStoreWithContext) *GenericCreatePRIOQdiscHandler {
	return &GenericCreatePRIOQdiscHandler{
		eventStore: eventStore,
	}
}

// HandleTyped processes the CreatePRIOQdiscCommand with compile-time type safety
func (h *GenericCreatePRIOQdiscHandler) HandleTyped(ctx context.Context, command *models.CreatePRIOQdiscCommand) error {
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

// GenericCreateFQCODELQdiscHandler handles CreateFQCODELQdiscCommand with type safety
type GenericCreateFQCODELQdiscHandler struct {
	eventStore eventstore.EventStoreWithContext
}

// NewGenericCreateFQCODELQdiscHandler creates a new type-safe FQ_CODEL handler
func NewGenericCreateFQCODELQdiscHandler(eventStore eventstore.EventStoreWithContext) *GenericCreateFQCODELQdiscHandler {
	return &GenericCreateFQCODELQdiscHandler{
		eventStore: eventStore,
	}
}

// HandleTyped processes the CreateFQCODELQdiscCommand with compile-time type safety
func (h *GenericCreateFQCODELQdiscHandler) HandleTyped(ctx context.Context, command *models.CreateFQCODELQdiscCommand) error {
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