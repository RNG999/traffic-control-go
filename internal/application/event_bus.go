package application

import (
	"context"
	"fmt"
	"sync"

	"github.com/rng999/traffic-control-go/pkg/logging"
)

// EventHandler handles domain events
type EventHandler func(ctx context.Context, event interface{}) error

// EventBus publishes domain events to subscribers
type EventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
	service  *TrafficControlService
	logger   logging.Logger
}

// NewEventBus creates a new event bus
func NewEventBus(service *TrafficControlService) *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
		service:  service,
		logger:   service.logger,
	}
}

// Subscribe subscribes a handler to an event type
func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
	eb.logger.Debug("Subscribed to event", logging.String("type", eventType))
}

// Publish publishes an event to all subscribers
func (eb *EventBus) Publish(ctx context.Context, eventType string, event interface{}) error {
	eb.mu.RLock()
	handlers := eb.handlers[eventType]
	eb.mu.RUnlock()

	if len(handlers) == 0 {
		eb.logger.Debug("No handlers for event type", logging.String("type", eventType))
		return nil
	}

	eb.logger.Debug("Publishing event", logging.String("type", eventType), logging.Int("handlers", len(handlers)))

	// Execute all handlers
	var errs []error
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			eb.logger.Error("Event handler failed",
				logging.String("type", eventType),
				logging.Error(err))
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("event handling failed with %d errors: %v", len(errs), errs)
	}

	eb.logger.Debug("Event published successfully", logging.String("type", eventType))
	return nil
}