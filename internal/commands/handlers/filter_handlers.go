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

// DeleteFilterHandler handles DeleteFilterCommand with type safety
type DeleteFilterHandler struct {
	eventStore eventstore.EventStoreWithContext
}

// NewDeleteFilterHandler creates a new type-safe handler
func NewDeleteFilterHandler(eventStore eventstore.EventStoreWithContext) *DeleteFilterHandler {
	return &DeleteFilterHandler{
		eventStore: eventStore,
	}
}

// HandleTyped processes the DeleteFilterCommand with compile-time type safety
func (h *DeleteFilterHandler) HandleTyped(ctx context.Context, command *models.DeleteFilterCommand) error {
	// Load aggregate
	aggregate := aggregates.NewTrafficControlAggregate(command.DeviceName)
	if err := h.eventStore.Load(ctx, aggregate.GetID(), aggregate); err != nil {
		return fmt.Errorf("failed to load aggregate: %w", err)
	}

	// Execute business logic
	if err := aggregate.DeleteFilter(command.Parent, command.Priority, command.Handle); err != nil {
		return err
	}

	// Save aggregate
	if err := h.eventStore.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate: %w", err)
	}

	return nil
}

// CreateAdvancedFilterHandler handles CreateAdvancedFilterCommand with type safety
type CreateAdvancedFilterHandler struct {
	eventStore eventstore.EventStoreWithContext
}

// NewCreateAdvancedFilterHandler creates a new type-safe handler
func NewCreateAdvancedFilterHandler(eventStore eventstore.EventStoreWithContext) *CreateAdvancedFilterHandler {
	return &CreateAdvancedFilterHandler{
		eventStore: eventStore,
	}
}

// HandleTyped processes the CreateAdvancedFilterCommand with compile-time type safety
func (h *CreateAdvancedFilterHandler) HandleTyped(ctx context.Context, command *models.CreateAdvancedFilterCommand) error {
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

	// Parse filter handle
	filterHandle, err := tc.ParseHandle(command.Handle)
	if err != nil {
		return fmt.Errorf("invalid filter handle: %w", err)
	}

	// Build matches based on advanced filter options
	matches := make([]entities.Match, 0)

	// IP Source Range
	if command.IPSourceRange != nil {
		if command.IPSourceRange.CIDR != "" {
			if match, err := entities.NewIPSourceMatch(command.IPSourceRange.CIDR); err == nil {
				matches = append(matches, match)
			}
		}
	}

	// IP Destination Range
	if command.IPDestRange != nil {
		if command.IPDestRange.CIDR != "" {
			if match, err := entities.NewIPDestinationMatch(command.IPDestRange.CIDR); err == nil {
				matches = append(matches, match)
			}
		}
	}

	// Port Source Range
	if command.PortSourceRange != nil {
		// For simplicity, use start port if range specified
		match := entities.NewPortSourceMatch(command.PortSourceRange.StartPort)
		matches = append(matches, match)
	}

	// Port Destination Range
	if command.PortDestRange != nil {
		// For simplicity, use start port if range specified
		match := entities.NewPortDestinationMatch(command.PortDestRange.StartPort)
		matches = append(matches, match)
	}

	// Transport Protocol
	if command.TransportProtocol != "" {
		protocolNum := getProtocolNumber(command.TransportProtocol)
		if protocolNum > 0 {
			match := entities.NewProtocolMatch(entities.TransportProtocol(protocolNum))
			matches = append(matches, match)
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

// getProtocolNumber converts protocol name to number
func getProtocolNumber(protocol string) uint8 {
	switch protocol {
	case "tcp", "TCP":
		return 6
	case "udp", "UDP":
		return 17
	case "icmp", "ICMP":
		return 1
	default:
		// Try to parse as number
		if num, err := strconv.ParseUint(protocol, 10, 8); err == nil {
			return uint8(num)
		}
		return 0
	}
}