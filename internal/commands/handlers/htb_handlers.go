package handlers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/rng999/traffic-control-go/internal/commands/models"
	"github.com/rng999/traffic-control-go/internal/domain/aggregates"
	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// CreateHTBQdiscHandler handles CreateHTBQdiscCommand with type safety
type CreateHTBQdiscHandler struct {
	eventStore eventstore.EventStoreWithContext
}

// NewCreateHTBQdiscHandler creates a new type-safe handler
func NewCreateHTBQdiscHandler(eventStore eventstore.EventStoreWithContext) *CreateHTBQdiscHandler {
	return &CreateHTBQdiscHandler{
		eventStore: eventStore,
	}
}

// HandleTyped processes the CreateHTBQdiscCommand with compile-time type safety
func (h *CreateHTBQdiscHandler) HandleTyped(ctx context.Context, command *models.CreateHTBQdiscCommand) error {
	// No type assertion needed - we receive the exact type we expect

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

	// Parse handle with better error handling
	handle, err := tc.ParseHandle(command.Handle)
	if err != nil {
		return fmt.Errorf("invalid handle format: %w", err)
	}

	// Parse default class handle
	defaultHandle, err := tc.ParseHandle(command.DefaultClass)
	if err != nil {
		return fmt.Errorf("invalid default class handle: %w", err)
	}

	// Execute business logic
	if err := aggregate.AddHTBQdisc(handle, defaultHandle); err != nil {
		return err
	}

	// Save aggregate
	if err := h.eventStore.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate: %w", err)
	}

	return nil
}

// CreateHTBClassHandler handles CreateHTBClassCommand with type safety
type CreateHTBClassHandler struct {
	eventStore eventstore.EventStoreWithContext
}

// NewCreateHTBClassHandler creates a new type-safe handler
func NewCreateHTBClassHandler(eventStore eventstore.EventStoreWithContext) *CreateHTBClassHandler {
	return &CreateHTBClassHandler{
		eventStore: eventStore,
	}
}

// HandleTyped processes the CreateHTBClassCommand with compile-time type safety
func (h *CreateHTBClassHandler) HandleTyped(ctx context.Context, command *models.CreateHTBClassCommand) error {
	// No type assertion needed - compile-time type safety guaranteed

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

	// Parse handles
	parentHandle, err := tc.ParseHandle(command.Parent)
	if err != nil {
		return fmt.Errorf("invalid parent handle: %w", err)
	}

	classHandle, err := tc.ParseHandle(command.ClassID)
	if err != nil {
		return fmt.Errorf("invalid class handle: %w", err)
	}

	// Parse bandwidths
	rate, err := tc.ParseBandwidth(command.Rate)
	if err != nil {
		return fmt.Errorf("invalid rate: %w", err)
	}

	ceil, err := tc.ParseBandwidth(command.Ceil)
	if err != nil {
		return fmt.Errorf("invalid ceil: %w", err)
	}

	// Execute business logic
	// Use ClassID as the name for now - in a real implementation you might want a separate name field
	className := command.ClassID
	if err := aggregate.AddHTBClass(parentHandle, classHandle, className, rate, ceil); err != nil {
		return err
	}

	// Save aggregate
	if err := h.eventStore.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate: %w", err)
	}

	return nil
}

// CreateFilterHandler handles CreateFilterCommand with type safety
type CreateFilterHandler struct {
	eventStore eventstore.EventStoreWithContext
}

// NewCreateFilterHandler creates a new type-safe handler
func NewCreateFilterHandler(eventStore eventstore.EventStoreWithContext) *CreateFilterHandler {
	return &CreateFilterHandler{
		eventStore: eventStore,
	}
}

// HandleTyped processes the CreateFilterCommand with compile-time type safety
func (h *CreateFilterHandler) HandleTyped(ctx context.Context, command *models.CreateFilterCommand) error {
	// Type safety guaranteed at compile time - no runtime assertions needed

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

	// Parse parent handle
	parentHandle, err := tc.ParseHandle(command.Parent)
	if err != nil {
		return fmt.Errorf("invalid parent handle: %w", err)
	}

	// Parse flow ID handle
	flowHandle, err := tc.ParseHandle(command.FlowID)
	if err != nil {
		return fmt.Errorf("invalid flow ID handle: %w", err)
	}

	// Create a handle for the filter (using priority as a simple approach)
	filterHandle := tc.NewHandle(0x800, uint16(command.Priority))

	// Convert map matches to entities.Match
	matches := make([]entities.Match, 0, len(command.Match))
	for key, value := range command.Match {
		switch key {
		case "src_ip":
			if match, err := entities.NewIPSourceMatch(value); err == nil {
				matches = append(matches, match)
			}
		case "dst_ip":
			if match, err := entities.NewIPDestinationMatch(value); err == nil {
				matches = append(matches, match)
			}
		case "src_port":
			if port, err := strconv.ParseUint(value, 10, 16); err == nil {
				match := entities.NewPortSourceMatch(uint16(port))
				matches = append(matches, match)
			}
		case "dst_port":
			if port, err := strconv.ParseUint(value, 10, 16); err == nil {
				match := entities.NewPortDestinationMatch(uint16(port))
				matches = append(matches, match)
			}
		}
	}

	// Execute business logic
	if err := aggregate.AddFilter(
		parentHandle,
		command.Priority,
		filterHandle,
		flowHandle,
		matches,
	); err != nil {
		return err
	}

	// Save aggregate
	if err := h.eventStore.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate: %w", err)
	}

	return nil
}
