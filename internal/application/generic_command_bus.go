package application

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/rng999/traffic-control-go/pkg/logging"
)

// GenericCommandHandler defines a type-safe command handler interface using generics
type GenericCommandHandler[T any] interface {
	HandleTyped(ctx context.Context, command T) error
}

// CommandHandlerWrapper wraps a generic handler to conform to the legacy interface
type CommandHandlerWrapper[T any] struct {
	handler GenericCommandHandler[T]
	logger  logging.Logger
}

// Handle implements the legacy CommandHandler interface using type-safe wrapper
func (w *CommandHandlerWrapper[T]) Handle(ctx context.Context, command interface{}) error {
	// Type assertion with better error handling
	typedCommand, ok := command.(T)
	if !ok {
		commandType := reflect.TypeOf(command)
		expectedType := reflect.TypeOf((*T)(nil)).Elem()
		w.logger.Error("Command type mismatch",
			logging.String("expected", expectedType.String()),
			logging.String("received", commandType.String()))
		return fmt.Errorf("expected command type %s, got %s", expectedType, commandType)
	}

	return w.handler.HandleTyped(ctx, typedCommand)
}

// NewCommandHandlerWrapper creates a wrapper for a generic handler
func NewCommandHandlerWrapper[T any](handler GenericCommandHandler[T], logger logging.Logger) *CommandHandlerWrapper[T] {
	return &CommandHandlerWrapper[T]{
		handler: handler,
		logger:  logger,
	}
}

// GenericCommandBus provides type-safe command execution capabilities
type GenericCommandBus struct {
	handlers map[reflect.Type]CommandHandler
	mu       sync.RWMutex
	service  *TrafficControlService
	logger   logging.Logger
}

// NewGenericCommandBus creates a new generic command bus
func NewGenericCommandBus(service *TrafficControlService) *GenericCommandBus {
	return &GenericCommandBus{
		handlers: make(map[reflect.Type]CommandHandler),
		service:  service,
		logger:   service.logger,
	}
}

// RegisterHandlerFor registers a generic command handler for a specific command type
func RegisterHandlerFor[T any](gcb *GenericCommandBus, handler GenericCommandHandler[T]) {
	gcb.mu.Lock()
	defer gcb.mu.Unlock()

	var commandType T
	reflectType := reflect.TypeOf(commandType)

	// If T is a pointer type, we want the element type for registration
	if reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}

	// Wrap the generic handler to work with the legacy interface
	wrapper := NewCommandHandlerWrapper(handler, gcb.logger)
	gcb.handlers[reflectType] = wrapper

	gcb.logger.Debug("Registered generic command handler",
		logging.String("type", reflectType.String()))
}

// ExecuteCommand executes a command with runtime type checking (simplified approach)
func (gcb *GenericCommandBus) ExecuteCommand(ctx context.Context, command interface{}) error {
	reflectType := reflect.TypeOf(command)

	// If command is a pointer, get the element type for lookup
	if reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}

	gcb.mu.RLock()
	handler, exists := gcb.handlers[reflectType]
	gcb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no handler registered for command type: %s", reflectType)
	}

	gcb.logger.Debug("Executing typed command",
		logging.String("type", reflectType.String()))

	// Execute the command through the wrapper
	if err := handler.Handle(ctx, command); err != nil {
		gcb.logger.Error("Typed command execution failed",
			logging.String("type", reflectType.String()),
			logging.Error(err))
		return err
	}

	// Publish events after successful command execution
	if err := gcb.publishCommandEvents(ctx, reflectType.String()); err != nil {
		gcb.logger.Error("Failed to publish command events",
			logging.String("type", reflectType.String()),
			logging.Error(err))
		return err
	}

	gcb.logger.Debug("Typed command executed successfully",
		logging.String("type", reflectType.String()))
	return nil
}

// publishCommandEvents publishes events after command execution (simplified)
func (gcb *GenericCommandBus) publishCommandEvents(ctx context.Context, commandType string) error {
	// Map command types to event types
	switch commandType {
	case "CreateHTBQdiscCommand":
		return gcb.service.eventBus.Publish(ctx, "QdiscCreated", nil)
	case "CreateHTBClassCommand":
		return gcb.service.eventBus.Publish(ctx, "ClassCreated", nil)
	case "CreateFilterCommand":
		return gcb.service.eventBus.Publish(ctx, "FilterCreated", nil)
	case "CreateTBFQdiscCommand":
		return gcb.service.eventBus.Publish(ctx, "QdiscCreated", nil)
	case "CreatePRIOQdiscCommand":
		return gcb.service.eventBus.Publish(ctx, "QdiscCreated", nil)
	case "CreateFQCODELQdiscCommand":
		return gcb.service.eventBus.Publish(ctx, "QdiscCreated", nil)
	}

	return nil
}

// ExecuteTypedCommand executes a command through the generic command bus
func (cb *CommandBus) ExecuteTypedCommand(ctx context.Context, command interface{}) error {
	// Delegate to the service's generic command bus
	if cb.service.genericCommandBus == nil {
		return fmt.Errorf("generic command bus not initialized")
	}

	return cb.service.genericCommandBus.ExecuteCommand(ctx, command)
}
