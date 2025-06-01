package application

import (
	"context"
	"fmt"
	"sync"

	"github.com/rng999/traffic-control-go/pkg/logging"
)

// QueryHandler handles a specific query
type QueryHandler interface {
	Handle(ctx context.Context, query interface{}) (interface{}, error)
}

// QueryBus routes queries to their handlers
type QueryBus struct {
	handlers map[string]QueryHandler
	mu       sync.RWMutex
	service  *TrafficControlService
	logger   logging.Logger
}

// NewQueryBus creates a new query bus
func NewQueryBus(service *TrafficControlService) *QueryBus {
	return &QueryBus{
		handlers: make(map[string]QueryHandler),
		service:  service,
		logger:   service.logger,
	}
}

// Register registers a query handler
func (qb *QueryBus) Register(queryType string, handler QueryHandler) {
	qb.mu.Lock()
	defer qb.mu.Unlock()
	
	qb.handlers[queryType] = handler
	qb.logger.Debug("Registered query handler", logging.String("type", queryType))
}

// Execute executes a query
func (qb *QueryBus) Execute(ctx context.Context, queryType string, query interface{}) (interface{}, error) {
	qb.mu.RLock()
	handler, exists := qb.handlers[queryType]
	qb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no handler registered for query type: %s", queryType)
	}

	qb.logger.Debug("Executing query", logging.String("type", queryType))
	
	result, err := handler.Handle(ctx, query)
	if err != nil {
		qb.logger.Error("Query execution failed", 
			logging.String("type", queryType),
			logging.Error(err))
		return nil, err
	}

	qb.logger.Debug("Query executed successfully", logging.String("type", queryType))
	return result, nil
}