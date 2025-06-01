package application

import (
	"context"
	"fmt"
	"sync"

	"github.com/rng999/traffic-control-go/pkg/logging"
)

// CommandHandler handles a specific command
type CommandHandler interface {
	Handle(ctx context.Context, command interface{}) error
}

// CommandBus routes commands to their handlers
type CommandBus struct {
	handlers map[string]CommandHandler
	mu       sync.RWMutex
	service  *TrafficControlService
	logger   logging.Logger
}

// NewCommandBus creates a new command bus
func NewCommandBus(service *TrafficControlService) *CommandBus {
	return &CommandBus{
		handlers: make(map[string]CommandHandler),
		service:  service,
		logger:   service.logger,
	}
}

// Register registers a command handler
func (cb *CommandBus) Register(commandType string, handler CommandHandler) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.handlers[commandType] = handler
	cb.logger.Debug("Registered command handler", logging.String("type", commandType))
}

// Execute executes a command
func (cb *CommandBus) Execute(ctx context.Context, commandType string, command interface{}) error {
	cb.mu.RLock()
	handler, exists := cb.handlers[commandType]
	cb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no handler registered for command type: %s", commandType)
	}

	cb.logger.Debug("Executing command", logging.String("type", commandType))
	
	// Execute the command
	if err := handler.Handle(ctx, command); err != nil {
		cb.logger.Error("Command execution failed", 
			logging.String("type", commandType),
			logging.Error(err))
		return err
	}

	// Publish events after successful command execution
	// This will trigger netlink updates
	if err := cb.publishCommandEvents(ctx, commandType); err != nil {
		cb.logger.Error("Failed to publish command events",
			logging.String("type", commandType),
			logging.Error(err))
		return err
	}

	cb.logger.Debug("Command executed successfully", logging.String("type", commandType))
	return nil
}

// publishCommandEvents publishes events after command execution
func (cb *CommandBus) publishCommandEvents(ctx context.Context, commandType string) error {
	// Get the latest events from event store and publish them
	// This is a simplified version - in production, you'd track which events
	// were created by the command
	
	switch commandType {
	case "CreateHTBQdisc":
		return cb.service.eventBus.Publish(ctx, "QdiscCreated", nil)
	case "CreateHTBClass":
		return cb.service.eventBus.Publish(ctx, "ClassCreated", nil)
	case "CreateFilter":
		return cb.service.eventBus.Publish(ctx, "FilterCreated", nil)
	}

	return nil
}