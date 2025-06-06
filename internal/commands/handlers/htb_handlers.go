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
	"github.com/rng999/traffic-control-go/pkg/types"
)

// Helper types for functional composition
type deviceCommandPair struct {
	device  tc.DeviceName
	command *models.CreateHTBQdiscCommand
}

type aggregateCommandPair struct {
	aggregate *aggregates.TrafficControlAggregate
	command   *models.CreateHTBQdiscCommand
}

type htbQdiscParams struct {
	aggregate    *aggregates.TrafficControlAggregate
	handle       tc.Handle
	defaultClass tc.Handle
}

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

// Handle processes the CreateHTBQdiscCommand (DEPRECATED: use HandleFunctional)
func (h *CreateHTBQdiscHandler) Handle(ctx context.Context, command interface{}) error {
	result := h.HandleFunctional(ctx, command)
	return result.Error()
}

// HandleFunctional processes the CreateHTBQdiscCommand using functional composition
func (h *CreateHTBQdiscHandler) HandleFunctional(ctx context.Context, command interface{}) types.Result[*aggregates.TrafficControlAggregate] {
	// Step 1: Validate command
	cmdResult := h.validateCommand(command)
	if cmdResult.IsFailure() {
		return types.Failure[*aggregates.TrafficControlAggregate](cmdResult.Error())
	}

	// Step 2: Create device
	deviceResult := h.createDeviceFromCommand(cmdResult.Value())
	if deviceResult.IsFailure() {
		return types.Failure[*aggregates.TrafficControlAggregate](deviceResult.Error())
	}

	// Step 3: Load aggregate
	aggregateResult := h.loadAggregateFromCommand(ctx)(deviceResult.Value())
	if aggregateResult.IsFailure() {
		return types.Failure[*aggregates.TrafficControlAggregate](aggregateResult.Error())
	}

	// Step 4: Parse handles
	paramsResult := h.parseHandlesFromCommand(aggregateResult.Value())
	if paramsResult.IsFailure() {
		return types.Failure[*aggregates.TrafficControlAggregate](paramsResult.Error())
	}

	// Step 5: Execute business logic
	businessResult := h.executeHTBQdiscBusinessLogic(paramsResult.Value())
	if businessResult.IsFailure() {
		return businessResult
	}

	// Step 6: Save aggregate
	return h.saveAggregate(ctx)(businessResult.Value())
}

// validateCommand validates the command type
func (h *CreateHTBQdiscHandler) validateCommand(command interface{}) types.Result[*models.CreateHTBQdiscCommand] {
	cmd, ok := command.(*models.CreateHTBQdiscCommand)
	if !ok {
		return types.Failure[*models.CreateHTBQdiscCommand](fmt.Errorf("invalid command type"))
	}
	return types.Success(cmd)
}

// createDeviceFromCommand creates device value object from command
func (h *CreateHTBQdiscHandler) createDeviceFromCommand(cmd *models.CreateHTBQdiscCommand) types.Result[deviceCommandPair] {
	device, err := tc.NewDevice(cmd.DeviceName)
	if err != nil {
		return types.Failure[deviceCommandPair](fmt.Errorf("invalid device name: %w", err))
	}
	return types.Success(deviceCommandPair{device: device, command: cmd})
}

// loadAggregateFromCommand loads aggregate from event store
func (h *CreateHTBQdiscHandler) loadAggregateFromCommand(ctx context.Context) func(pair deviceCommandPair) types.Result[aggregateCommandPair] {
	return func(pair deviceCommandPair) types.Result[aggregateCommandPair] {
		aggregate := aggregates.NewTrafficControlAggregate(pair.device)
		if err := h.eventStore.Load(ctx, aggregate.GetID(), aggregate); err != nil {
			return types.Failure[aggregateCommandPair](fmt.Errorf("failed to load aggregate: %w", err))
		}
		return types.Success(aggregateCommandPair{aggregate: aggregate, command: pair.command})
	}
}

// parseHandlesFromCommand parses handles from string format
func (h *CreateHTBQdiscHandler) parseHandlesFromCommand(pair aggregateCommandPair) types.Result[htbQdiscParams] {
	cmd := pair.command

	// Parse main handle
	var handleMajor, handleMinor uint16
	if n, err := fmt.Sscanf(cmd.Handle, "%d:%d", &handleMajor, &handleMinor); err != nil || n != 2 {
		return types.Failure[htbQdiscParams](fmt.Errorf("invalid handle format: %s", cmd.Handle))
	}
	handle := tc.NewHandle(handleMajor, handleMinor)

	// Parse default class handle
	var defaultMajor, defaultMinor uint16
	if n, err := fmt.Sscanf(cmd.DefaultClass, "%d:%d", &defaultMajor, &defaultMinor); err != nil || n != 2 {
		return types.Failure[htbQdiscParams](fmt.Errorf("invalid default class format: %s", cmd.DefaultClass))
	}
	defaultClass := tc.NewHandle(defaultMajor, defaultMinor)

	return types.Success(htbQdiscParams{
		aggregate:    pair.aggregate,
		handle:       handle,
		defaultClass: defaultClass,
	})
}

// executeHTBQdiscBusinessLogic executes the core business logic
func (h *CreateHTBQdiscHandler) executeHTBQdiscBusinessLogic(params htbQdiscParams) types.Result[*aggregates.TrafficControlAggregate] {
	return params.aggregate.WithHTBQdisc(params.handle, params.defaultClass)
}

// saveAggregate saves the aggregate to event store
func (h *CreateHTBQdiscHandler) saveAggregate(ctx context.Context) func(*aggregates.TrafficControlAggregate) types.Result[*aggregates.TrafficControlAggregate] {
	return func(aggregate *aggregates.TrafficControlAggregate) types.Result[*aggregates.TrafficControlAggregate] {
		if err := h.eventStore.SaveAggregate(ctx, aggregate); err != nil {
			return types.Failure[*aggregates.TrafficControlAggregate](fmt.Errorf("failed to save aggregate: %w", err))
		}
		return types.Success(aggregate)
	}
}

// Helper types for HTB class creation
type htbClassCommand struct {
	device  tc.DeviceName
	command *models.CreateHTBClassCommand
}

type htbClassAggregate struct {
	aggregate *aggregates.TrafficControlAggregate
	command   *models.CreateHTBClassCommand
}

type htbClassParams struct {
	aggregate    *aggregates.TrafficControlAggregate
	parentHandle tc.Handle
	classHandle  tc.Handle
	rate         tc.Bandwidth
	ceil         tc.Bandwidth
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

// Handle processes the CreateHTBClassCommand (DEPRECATED: use HandleFunctional)
func (h *CreateHTBClassHandler) Handle(ctx context.Context, command interface{}) error {
	result := h.HandleFunctional(ctx, command)
	return result.Error()
}

// HandleFunctional processes the CreateHTBClassCommand using functional composition
func (h *CreateHTBClassHandler) HandleFunctional(ctx context.Context, command interface{}) types.Result[*aggregates.TrafficControlAggregate] {
	// Step 1: Validate command
	cmdResult := h.validateHTBClassCommand(command)
	if cmdResult.IsFailure() {
		return types.Failure[*aggregates.TrafficControlAggregate](cmdResult.Error())
	}

	// Step 2: Create device
	deviceResult := h.createDeviceFromHTBClassCommand(cmdResult.Value())
	if deviceResult.IsFailure() {
		return types.Failure[*aggregates.TrafficControlAggregate](deviceResult.Error())
	}

	// Step 3: Load aggregate
	aggregateResult := h.loadAggregateFromHTBClassCommand(ctx)(deviceResult.Value())
	if aggregateResult.IsFailure() {
		return types.Failure[*aggregates.TrafficControlAggregate](aggregateResult.Error())
	}

	// Step 4: Parse parameters
	paramsResult := h.parseHTBClassParameters(aggregateResult.Value())
	if paramsResult.IsFailure() {
		return types.Failure[*aggregates.TrafficControlAggregate](paramsResult.Error())
	}

	// Step 5: Execute business logic
	businessResult := h.executeHTBClassBusinessLogic(paramsResult.Value())
	if businessResult.IsFailure() {
		return businessResult
	}

	// Step 6: Save aggregate
	return h.saveAggregate(ctx)(businessResult.Value())
}

// validateHTBClassCommand validates the command type
func (h *CreateHTBClassHandler) validateHTBClassCommand(command interface{}) types.Result[*models.CreateHTBClassCommand] {
	cmd, ok := command.(*models.CreateHTBClassCommand)
	if !ok {
		return types.Failure[*models.CreateHTBClassCommand](fmt.Errorf("invalid command type"))
	}
	return types.Success(cmd)
}

// createDeviceFromHTBClassCommand creates device value object from command
func (h *CreateHTBClassHandler) createDeviceFromHTBClassCommand(cmd *models.CreateHTBClassCommand) types.Result[htbClassCommand] {
	device, err := tc.NewDevice(cmd.DeviceName)
	if err != nil {
		return types.Failure[htbClassCommand](fmt.Errorf("invalid device name: %w", err))
	}
	return types.Success(htbClassCommand{device: device, command: cmd})
}

// loadAggregateFromHTBClassCommand loads aggregate from event store
func (h *CreateHTBClassHandler) loadAggregateFromHTBClassCommand(ctx context.Context) func(pair htbClassCommand) types.Result[htbClassAggregate] {
	return func(pair htbClassCommand) types.Result[htbClassAggregate] {
		aggregate := aggregates.NewTrafficControlAggregate(pair.device)
		if err := h.eventStore.Load(ctx, aggregate.GetID(), aggregate); err != nil {
			return types.Failure[htbClassAggregate](fmt.Errorf("failed to load aggregate: %w", err))
		}
		return types.Success(htbClassAggregate{aggregate: aggregate, command: pair.command})
	}
}

// parseHTBClassParameters parses all parameters from command
func (h *CreateHTBClassHandler) parseHTBClassParameters(pair htbClassAggregate) types.Result[htbClassParams] {
	cmd := pair.command

	// Parse parent handle
	var parentMajor, parentMinor uint16
	if n, err := fmt.Sscanf(cmd.Parent, "%d:%d", &parentMajor, &parentMinor); err != nil || n != 2 {
		return types.Failure[htbClassParams](fmt.Errorf("invalid parent handle format: %s", cmd.Parent))
	}
	parentHandle := tc.NewHandle(parentMajor, parentMinor)

	// Parse class handle
	var classMajor, classMinor uint16
	if n, err := fmt.Sscanf(cmd.ClassID, "%d:%d", &classMajor, &classMinor); err != nil || n != 2 {
		return types.Failure[htbClassParams](fmt.Errorf("invalid class ID format: %s", cmd.ClassID))
	}
	classHandle := tc.NewHandle(classMajor, classMinor)

	// Parse bandwidth values
	rate, err := tc.NewBandwidth(cmd.Rate)
	if err != nil {
		return types.Failure[htbClassParams](fmt.Errorf("invalid rate: %w", err))
	}
	ceil, err := tc.NewBandwidth(cmd.Ceil)
	if err != nil {
		return types.Failure[htbClassParams](fmt.Errorf("invalid ceil: %w", err))
	}

	return types.Success(htbClassParams{
		aggregate:    pair.aggregate,
		parentHandle: parentHandle,
		classHandle:  classHandle,
		rate:         rate,
		ceil:         ceil,
	})
}

// executeHTBClassBusinessLogic executes the core business logic
func (h *CreateHTBClassHandler) executeHTBClassBusinessLogic(params htbClassParams) types.Result[*aggregates.TrafficControlAggregate] {
	return params.aggregate.WithHTBClass(
		params.parentHandle,
		params.classHandle,
		"class", // Default name
		params.rate,
		params.ceil,
	)
}

// saveAggregate saves the aggregate to event store (shared method)
func (h *CreateHTBClassHandler) saveAggregate(ctx context.Context) func(*aggregates.TrafficControlAggregate) types.Result[*aggregates.TrafficControlAggregate] {
	return func(aggregate *aggregates.TrafficControlAggregate) types.Result[*aggregates.TrafficControlAggregate] {
		if err := h.eventStore.SaveAggregate(ctx, aggregate); err != nil {
			return types.Failure[*aggregates.TrafficControlAggregate](fmt.Errorf("failed to save aggregate: %w", err))
		}
		return types.Success(aggregate)
	}
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
	device, err := tc.NewDevice(cmd.DeviceName)
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
	parentHandle := tc.NewHandle(parentMajor, parentMinor)

	// Generate filter handle (use hex format for handles)
	handle := tc.NewHandle(0x800, 0x800)

	var flowMajor, flowMinor uint16
	if n, err := fmt.Sscanf(cmd.FlowID, "%d:%d", &flowMajor, &flowMinor); err != nil || n != 2 {
		return fmt.Errorf("invalid flow ID format: %s (error: %v, matched: %d)", cmd.FlowID, err, n)
	}
	flowID := tc.NewHandle(flowMajor, flowMinor)

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
