package handlers

import (
	"context"
	"fmt"

	"github.com/rng999/traffic-control-go/internal/commands/models"
	"github.com/rng999/traffic-control-go/internal/domain/aggregates"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
)

// CreateTBFQdiscHandler handles the creation of TBF qdiscs
type CreateTBFQdiscHandler struct {
	eventStore eventstore.EventStoreWithContext
}

// NewCreateTBFQdiscHandler creates a new handler
func NewCreateTBFQdiscHandler(eventStore eventstore.EventStoreWithContext) *CreateTBFQdiscHandler {
	return &CreateTBFQdiscHandler{
		eventStore: eventStore,
	}
}

// Handle processes the CreateTBFQdiscCommand
func (h *CreateTBFQdiscHandler) Handle(ctx context.Context, command interface{}) error {
	cmd, ok := command.(*models.CreateTBFQdiscCommand)
	if !ok {
		return fmt.Errorf("invalid command type")
	}

	// Create device value object
	device, err := valueobjects.NewDevice(cmd.DeviceName)
	if err != nil {
		return fmt.Errorf("invalid device name: %w", err)
	}

	// Load aggregate
	aggregate := aggregates.NewTrafficControlAggregate(device)
	if err := h.eventStore.Load(ctx, aggregate.GetID(), aggregate); err != nil {
		return fmt.Errorf("failed to load aggregate: %w", err)
	}

	// Parse handle
	var handleMajor, handleMinor uint16
	fmt.Sscanf(cmd.Handle, "%x:%x", &handleMajor, &handleMinor)
	handle := valueobjects.NewHandle(handleMajor, handleMinor)

	// Parse bandwidth
	rate, err := valueobjects.NewBandwidth(cmd.Rate)
	if err != nil {
		return fmt.Errorf("invalid rate: %w", err)
	}

	// Execute business logic
	if err := aggregate.AddTBFQdisc(handle, rate, cmd.Buffer, cmd.Limit, cmd.Burst); err != nil {
		return err
	}

	// Save aggregate
	if err := h.eventStore.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate: %w", err)
	}

	return nil
}

// CreatePRIOQdiscHandler handles the creation of PRIO qdiscs
type CreatePRIOQdiscHandler struct {
	eventStore eventstore.EventStoreWithContext
}

// NewCreatePRIOQdiscHandler creates a new handler
func NewCreatePRIOQdiscHandler(eventStore eventstore.EventStoreWithContext) *CreatePRIOQdiscHandler {
	return &CreatePRIOQdiscHandler{
		eventStore: eventStore,
	}
}

// Handle processes the CreatePRIOQdiscCommand
func (h *CreatePRIOQdiscHandler) Handle(ctx context.Context, command interface{}) error {
	cmd, ok := command.(*models.CreatePRIOQdiscCommand)
	if !ok {
		return fmt.Errorf("invalid command type")
	}

	// Create device value object
	device, err := valueobjects.NewDevice(cmd.DeviceName)
	if err != nil {
		return fmt.Errorf("invalid device name: %w", err)
	}

	// Load aggregate
	aggregate := aggregates.NewTrafficControlAggregate(device)
	if err := h.eventStore.Load(ctx, aggregate.GetID(), aggregate); err != nil {
		return fmt.Errorf("failed to load aggregate: %w", err)
	}

	// Parse handle
	var handleMajor, handleMinor uint16
	fmt.Sscanf(cmd.Handle, "%x:%x", &handleMajor, &handleMinor)
	handle := valueobjects.NewHandle(handleMajor, handleMinor)

	// Execute business logic
	if err := aggregate.AddPRIOQdisc(handle, cmd.Bands, cmd.Priomap); err != nil {
		return err
	}

	// Save aggregate
	if err := h.eventStore.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate: %w", err)
	}

	return nil
}

// CreateFQCODELQdiscHandler handles the creation of FQ_CODEL qdiscs
type CreateFQCODELQdiscHandler struct {
	eventStore eventstore.EventStoreWithContext
}

// NewCreateFQCODELQdiscHandler creates a new handler
func NewCreateFQCODELQdiscHandler(eventStore eventstore.EventStoreWithContext) *CreateFQCODELQdiscHandler {
	return &CreateFQCODELQdiscHandler{
		eventStore: eventStore,
	}
}

// Handle processes the CreateFQCODELQdiscCommand
func (h *CreateFQCODELQdiscHandler) Handle(ctx context.Context, command interface{}) error {
	cmd, ok := command.(*models.CreateFQCODELQdiscCommand)
	if !ok {
		return fmt.Errorf("invalid command type")
	}

	// Create device value object
	device, err := valueobjects.NewDevice(cmd.DeviceName)
	if err != nil {
		return fmt.Errorf("invalid device name: %w", err)
	}

	// Load aggregate
	aggregate := aggregates.NewTrafficControlAggregate(device)
	if err := h.eventStore.Load(ctx, aggregate.GetID(), aggregate); err != nil {
		return fmt.Errorf("failed to load aggregate: %w", err)
	}

	// Parse handle
	var handleMajor, handleMinor uint16
	fmt.Sscanf(cmd.Handle, "%x:%x", &handleMajor, &handleMinor)
	handle := valueobjects.NewHandle(handleMajor, handleMinor)

	// Execute business logic
	if err := aggregate.AddFQCODELQdisc(
		handle,
		cmd.Limit,
		cmd.Flows,
		cmd.Target,
		cmd.Interval,
		cmd.Quantum,
		cmd.ECN,
	); err != nil {
		return err
	}

	// Save aggregate
	if err := h.eventStore.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate: %w", err)
	}

	return nil
}