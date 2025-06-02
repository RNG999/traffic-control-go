package handlers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/rng999/traffic-control-go/internal/commands/models"
	"github.com/rng999/traffic-control-go/internal/domain/aggregates"
	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
)

// CreateHTBQdiscHandler handles the creation of HTB qdiscs
type CreateHTBQdiscHandler struct {
	eventStore eventstore.EventStoreWithContext
}

// NewCreateHTBQdiscHandler creates a new handler
func NewCreateHTBQdiscHandler(eventStore eventstore.EventStoreWithContext) *CreateHTBQdiscHandler {
	return &CreateHTBQdiscHandler{
		eventStore: eventStore,
	}
}

// Handle processes the CreateHTBQdiscCommand
func (h *CreateHTBQdiscHandler) Handle(ctx context.Context, command interface{}) error {
	cmd, ok := command.(*models.CreateHTBQdiscCommand)
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

	// Parse handles
	// Parse handles from string format "major:minor" (decimal format like "1:0")
	var handleMajor, handleMinor uint16
	if n, err := fmt.Sscanf(cmd.Handle, "%d:%d", &handleMajor, &handleMinor); err != nil || n != 2 {
		return fmt.Errorf("invalid handle format: %s", cmd.Handle)
	}
	handle := valueobjects.NewHandle(handleMajor, handleMinor)

	var defaultMajor, defaultMinor uint16
	if n, err := fmt.Sscanf(cmd.DefaultClass, "%d:%d", &defaultMajor, &defaultMinor); err != nil || n != 2 {
		return fmt.Errorf("invalid default class format: %s", cmd.DefaultClass)
	}
	defaultClass := valueobjects.NewHandle(defaultMajor, defaultMinor)

	// Execute business logic
	if err := aggregate.AddHTBQdisc(handle, defaultClass); err != nil {
		return err
	}

	// Save aggregate
	if err := h.eventStore.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate: %w", err)
	}

	return nil
}

// CreateHTBClassHandler handles the creation of HTB classes
type CreateHTBClassHandler struct {
	eventStore eventstore.EventStoreWithContext
}

// NewCreateHTBClassHandler creates a new handler
func NewCreateHTBClassHandler(eventStore eventstore.EventStoreWithContext) *CreateHTBClassHandler {
	return &CreateHTBClassHandler{
		eventStore: eventStore,
	}
}

// Handle processes the CreateHTBClassCommand
func (h *CreateHTBClassHandler) Handle(ctx context.Context, command interface{}) error {
	cmd, ok := command.(*models.CreateHTBClassCommand)
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

	// Parse handles from string format "major:minor" (decimal format like "1:0")
	var parentMajor, parentMinor uint16
	if n, err := fmt.Sscanf(cmd.Parent, "%d:%d", &parentMajor, &parentMinor); err != nil || n != 2 {
		return fmt.Errorf("invalid parent handle format: %s (error: %v, matched: %d)", cmd.Parent, err, n)
	}
	parentHandle := valueobjects.NewHandle(parentMajor, parentMinor)

	var classMajor, classMinor uint16
	if n, err := fmt.Sscanf(cmd.ClassID, "%d:%d", &classMajor, &classMinor); err != nil || n != 2 {
		return fmt.Errorf("invalid class ID format: %s", cmd.ClassID)
	}
	classHandle := valueobjects.NewHandle(classMajor, classMinor)

	// Parse bandwidth values
	rate, err := valueobjects.NewBandwidth(cmd.Rate)
	if err != nil {
		return fmt.Errorf("invalid rate: %w", err)
	}
	ceil, err := valueobjects.NewBandwidth(cmd.Ceil)
	if err != nil {
		return fmt.Errorf("invalid ceil: %w", err)
	}

	// Execute business logic
	if err := aggregate.AddHTBClass(
		parentHandle,
		classHandle,
		"class", // Default name
		rate,
		ceil,
	); err != nil {
		return err
	}

	// Save aggregate
	if err := h.eventStore.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("failed to save aggregate: %w", err)
	}

	return nil
}

// CreateFilterHandler handles the creation of filters
type CreateFilterHandler struct {
	eventStore eventstore.EventStoreWithContext
}

// NewCreateFilterHandler creates a new handler
func NewCreateFilterHandler(eventStore eventstore.EventStoreWithContext) *CreateFilterHandler {
	return &CreateFilterHandler{
		eventStore: eventStore,
	}
}

// Handle processes the CreateFilterCommand
func (h *CreateFilterHandler) Handle(ctx context.Context, command interface{}) error {
	cmd, ok := command.(*models.CreateFilterCommand)
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

	// Parse handles from string format "major:minor" (decimal format like "1:0")
	var parentMajor, parentMinor uint16
	if n, err := fmt.Sscanf(cmd.Parent, "%d:%d", &parentMajor, &parentMinor); err != nil || n != 2 {
		return fmt.Errorf("invalid parent handle format: %s (error: %v, matched: %d)", cmd.Parent, err, n)
	}
	parentHandle := valueobjects.NewHandle(parentMajor, parentMinor)

	// Generate filter handle (use hex format for handles)
	handle := valueobjects.NewHandle(0x800, 0x800)

	var flowMajor, flowMinor uint16
	if n, err := fmt.Sscanf(cmd.FlowID, "%d:%d", &flowMajor, &flowMinor); err != nil || n != 2 {
		return fmt.Errorf("invalid flow ID format: %s (error: %v, matched: %d)", cmd.FlowID, err, n)
	}
	flowID := valueobjects.NewHandle(flowMajor, flowMinor)

	// Convert match map to Match entities
	matches := make([]entities.Match, 0)
	for matchType, value := range cmd.Match {
		switch matchType {
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
				matches = append(matches, entities.NewPortSourceMatch(uint16(port)))
			}
		case "dst_port":
			if port, err := strconv.ParseUint(value, 10, 16); err == nil {
				matches = append(matches, entities.NewPortDestinationMatch(uint16(port)))
			}
		case "protocol":
			var proto entities.TransportProtocol
			switch value {
			case "tcp":
				proto = entities.TransportProtocolTCP
			case "udp":
				proto = entities.TransportProtocolUDP
			case "icmp":
				proto = entities.TransportProtocolICMP
			default:
				if p, err := strconv.ParseUint(value, 10, 8); err == nil {
					proto = entities.TransportProtocol(p)
				}
			}
			matches = append(matches, entities.NewProtocolMatch(proto))
		}
	}

	// Execute business logic
	if err := aggregate.AddFilter(
		parentHandle,
		cmd.Priority,
		handle,
		flowID,
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
