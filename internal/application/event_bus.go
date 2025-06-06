package application

import (
	"context"
	"fmt"
	"sync"

	"github.com/rng999/traffic-control-go/internal/domain/events"
	"github.com/rng999/traffic-control-go/pkg/logging"
	"github.com/rng999/traffic-control-go/pkg/types"
)

// EventHandler handles domain events (legacy interface)
type EventHandler func(ctx context.Context, event interface{}) error

// FunctionalEventHandler is a pure function that handles domain events immutably
type FunctionalEventHandler[TEvent events.DomainEvent, TResult any] func(ctx context.Context, event TEvent) types.Result[TResult]

// EventBus publishes domain events to subscribers with support for both legacy and functional handlers
type EventBus struct {
	// Legacy handlers for backward compatibility
	handlers map[string][]EventHandler
	// Functional handlers for type-safe immutable event processing
	functionalHandlers map[string][]func(ctx context.Context, event events.DomainEvent) types.Result[interface{}]
	mu                 sync.RWMutex
	service            *TrafficControlService
	logger             logging.Logger
}

// NewEventBus creates a new event bus with support for both legacy and functional handlers
func NewEventBus(service *TrafficControlService) *EventBus {
	return &EventBus{
		handlers:           make(map[string][]EventHandler),
		functionalHandlers: make(map[string][]func(ctx context.Context, event events.DomainEvent) types.Result[interface{}]),
		service:            service,
		logger:             service.logger,
	}
}

// Subscribe subscribes a legacy handler to an event type (backward compatibility)
func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
	eb.logger.Debug("Subscribed legacy handler to event", logging.String("type", eventType))
}

// SubscribeFunctional subscribes a type-safe functional handler to an event type
func SubscribeFunctional[TEvent events.DomainEvent, TResult any](
	eb *EventBus,
	eventTypeName string,
	handler FunctionalEventHandler[TEvent, TResult],
) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Wrap the typed handler to fit the interface
	wrappedHandler := func(ctx context.Context, event events.DomainEvent) types.Result[interface{}] {
		// Type-safe cast
		typedEvent, ok := event.(TEvent)
		if !ok {
			eb.logger.Debug("Event type mismatch, skipping functional handler",
				logging.String("expected", eventTypeName),
				logging.String("actual", event.EventType()),
			)
			return types.Success[interface{}](nil) // Skip this handler
		}

		// Call the typed handler
		result := handler(ctx, typedEvent)
		if result.IsFailure() {
			return types.Failure[interface{}](result.Error())
		}

		// Convert result to interface{}
		return types.Success[interface{}](result.Value())
	}

	eb.functionalHandlers[eventTypeName] = append(eb.functionalHandlers[eventTypeName], wrappedHandler)
	eb.logger.Debug("Subscribed functional handler to event", logging.String("type", eventTypeName))
}

// SubscribeMultiple allows subscribing to multiple event types with the same handler
func (eb *EventBus) SubscribeMultiple(eventTypeNames []string, handler EventHandler) {
	for _, eventTypeName := range eventTypeNames {
		eb.Subscribe(eventTypeName, handler)
	}
}

// Publish publishes an event to all subscribers (legacy interface)
func (eb *EventBus) Publish(ctx context.Context, eventType string, event interface{}) error {
	eb.mu.RLock()
	legacyHandlers := eb.handlers[eventType]
	functionalHandlers := eb.functionalHandlers[eventType]
	eb.mu.RUnlock()

	totalHandlers := len(legacyHandlers) + len(functionalHandlers)
	if totalHandlers == 0 {
		eb.logger.Debug("No handlers for event type", logging.String("type", eventType))
		return nil
	}

	eb.logger.Debug("Publishing event",
		logging.String("type", eventType),
		logging.Int("legacy_handlers", len(legacyHandlers)),
		logging.Int("functional_handlers", len(functionalHandlers)))

	var errs []error

	// Execute legacy handlers
	for _, handler := range legacyHandlers {
		if err := handler(ctx, event); err != nil {
			eb.logger.Error("Legacy event handler failed",
				logging.String("type", eventType),
				logging.Error(err))
			errs = append(errs, err)
		}
	}

	// Execute functional handlers if event is a domain event
	if domainEvent, ok := event.(events.DomainEvent); ok {
		for _, handler := range functionalHandlers {
			result := handler(ctx, domainEvent)
			if result.IsFailure() {
				eb.logger.Error("Functional event handler failed",
					logging.String("type", eventType),
					logging.Error(result.Error()))
				errs = append(errs, result.Error())
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("event handling failed with %d errors: %v", len(errs), errs)
	}

	eb.logger.Debug("Event published successfully", logging.String("type", eventType))
	return nil
}

// PublishFunctional publishes a domain event to functional handlers and returns typed results
func (eb *EventBus) PublishFunctional(ctx context.Context, event events.DomainEvent) types.Result[[]interface{}] {
	eventTypeName := event.EventType()

	eb.mu.RLock()
	handlers := eb.functionalHandlers[eventTypeName]
	eb.mu.RUnlock()

	if len(handlers) == 0 {
		eb.logger.Debug("No functional handlers for event type", logging.String("type", eventTypeName))
		return types.Success([]interface{}{})
	}

	eb.logger.Debug("Publishing event functionally",
		logging.String("type", eventTypeName),
		logging.Int("handlers", len(handlers)),
	)

	// Use functional composition to process all handlers
	results := types.Map(handlers, func(handler func(ctx context.Context, event events.DomainEvent) types.Result[interface{}]) types.Result[interface{}] {
		return handler(ctx, event)
	})

	// Collect successful results and errors separately
	var successResults []interface{}
	var errors []error

	for _, result := range results {
		if result.IsSuccess() && result.Value() != nil {
			successResults = append(successResults, result.Value())
		} else if result.IsFailure() {
			errors = append(errors, result.Error())
			eb.logger.Error("Functional event handler failed",
				logging.String("type", eventTypeName),
				logging.Error(result.Error()),
			)
		}
	}

	// If any handlers failed, return failure with collected errors
	if len(errors) > 0 {
		eb.logger.Error("Some functional event handlers failed",
			logging.String("type", eventTypeName),
			logging.Int("failed", len(errors)),
			logging.Int("succeeded", len(successResults)),
		)
		// Return the first error for simplicity
		return types.Failure[[]interface{}](errors[0])
	}

	eb.logger.Debug("Event published functionally with success",
		logging.String("type", eventTypeName),
		logging.Int("results", len(successResults)),
	)
	return types.Success(successResults)
}

// PublishAndCollect publishes an event and collects all successful results of a specific type
func PublishAndCollect[TResult any](
	eb *EventBus,
	ctx context.Context,
	event events.DomainEvent,
) types.Result[[]TResult] {
	publishResult := eb.PublishFunctional(ctx, event)
	if publishResult.IsFailure() {
		return types.Failure[[]TResult](publishResult.Error())
	}

	// Filter and convert results to the desired type
	results := publishResult.Value()
	typedResults := types.Filter(results, func(result interface{}) bool {
		_, ok := result.(TResult)
		return ok
	})

	// Convert to typed slice
	converted := types.Map(typedResults, func(result interface{}) TResult {
		return result.(TResult)
	})

	return types.Success(converted)
}

// SubscribeToMultiple allows subscribing to multiple event types with the same handler
func SubscribeToMultiple[TEvent events.DomainEvent, TResult any](
	eb *EventBus,
	eventTypeNames []string,
	handler FunctionalEventHandler[TEvent, TResult],
) {
	for _, eventTypeName := range eventTypeNames {
		SubscribeFunctional(eb, eventTypeName, handler)
	}
}

// Chain allows functional composition of event processing
func Chain[TEvent events.DomainEvent, TIntermediate, TResult any](
	firstHandler FunctionalEventHandler[TEvent, TIntermediate],
	secondHandler func(ctx context.Context, intermediate TIntermediate) types.Result[TResult],
) FunctionalEventHandler[TEvent, TResult] {
	return func(ctx context.Context, event TEvent) types.Result[TResult] {
		firstResult := firstHandler(ctx, event)
		if firstResult.IsFailure() {
			return types.Failure[TResult](firstResult.Error())
		}
		return secondHandler(ctx, firstResult.Value())
	}
}

// Compose creates a handler that applies multiple handlers in sequence
func Compose[TEvent events.DomainEvent, TResult any](
	handlers ...FunctionalEventHandler[TEvent, TResult],
) FunctionalEventHandler[TEvent, []TResult] {
	return func(ctx context.Context, event TEvent) types.Result[[]TResult] {
		results := make([]TResult, 0, len(handlers))

		for _, handler := range handlers {
			result := handler(ctx, event)
			if result.IsFailure() {
				return types.Failure[[]TResult](result.Error())
			}
			results = append(results, result.Value())
		}

		return types.Success(results)
	}
}

// Filter creates a conditional handler that only processes events matching a predicate
func Filter[TEvent events.DomainEvent, TResult any](
	predicate func(TEvent) bool,
	handler FunctionalEventHandler[TEvent, TResult],
) FunctionalEventHandler[TEvent, types.Option[TResult]] {
	return func(ctx context.Context, event TEvent) types.Result[types.Option[TResult]] {
		if !predicate(event) {
			return types.Success(types.None[TResult]())
		}

		result := handler(ctx, event)
		if result.IsFailure() {
			return types.Failure[types.Option[TResult]](result.Error())
		}

		return types.Success(types.Some(result.Value()))
	}
}
