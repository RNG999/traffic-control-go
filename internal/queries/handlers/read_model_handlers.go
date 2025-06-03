package handlers

import (
	"context"
	"fmt"

	"github.com/rng999/traffic-control-go/internal/projections"
	"github.com/rng999/traffic-control-go/internal/queries/models"
)

// GetQdiscHandler handles qdisc queries using read models
type GetQdiscHandler struct {
	readModelStore projections.ReadModelStore
}

// NewGetQdiscHandler creates a new handler
func NewGetQdiscHandler(readModelStore projections.ReadModelStore) *GetQdiscHandler {
	return &GetQdiscHandler{
		readModelStore: readModelStore,
	}
}

// Handle processes the query
func (h *GetQdiscHandler) Handle(ctx context.Context, query interface{}) (interface{}, error) {
	q, ok := query.(*models.GetQdiscQuery)
	if !ok {
		return nil, fmt.Errorf("invalid query type")
	}

	// Get read model
	var readModel projections.TrafficControlReadModel
	modelID := fmt.Sprintf("tc:%s", q.DeviceName)

	if err := h.readModelStore.Get(ctx, "traffic-control", modelID, &readModel); err != nil {
		return nil, fmt.Errorf("failed to get read model: %w", err)
	}

	// Find specific qdisc
	for _, qdisc := range readModel.Qdiscs {
		if qdisc.Handle == q.Handle {
			return &models.QdiscView{
				DeviceName:   readModel.DeviceName,
				Handle:       qdisc.Handle,
				Type:         qdisc.Type,
				Parent:       qdisc.Parent,
				DefaultClass: qdisc.DefaultClass,
				Parameters:   qdisc.Parameters,
			}, nil
		}
	}

	return nil, fmt.Errorf("qdisc not found")
}

// GetClassHandler handles class queries using read models
type GetClassHandler struct {
	readModelStore projections.ReadModelStore
}

// NewGetClassHandler creates a new handler
func NewGetClassHandler(readModelStore projections.ReadModelStore) *GetClassHandler {
	return &GetClassHandler{
		readModelStore: readModelStore,
	}
}

// Handle processes the query
func (h *GetClassHandler) Handle(ctx context.Context, query interface{}) (interface{}, error) {
	q, ok := query.(*models.GetClassQuery)
	if !ok {
		return nil, fmt.Errorf("invalid query type")
	}

	// Get read model
	var readModel projections.TrafficControlReadModel
	modelID := fmt.Sprintf("tc:%s", q.DeviceName)

	if err := h.readModelStore.Get(ctx, "traffic-control", modelID, &readModel); err != nil {
		return nil, fmt.Errorf("failed to get read model: %w", err)
	}

	// Find specific class
	for _, class := range readModel.Classes {
		if class.Handle == q.ClassID {
			return &models.ClassView{
				DeviceName: readModel.DeviceName,
				Handle:     class.Handle,
				Parent:     class.Parent,
				Type:       class.Type,
				Name:       class.Name,
				Rate:       class.Rate,
				Ceil:       class.Ceil,
				Priority:   class.Priority,
				Parameters: class.Parameters,
			}, nil
		}
	}

	return nil, fmt.Errorf("class not found")
}

// GetFilterHandler handles filter queries using read models
type GetFilterHandler struct {
	readModelStore projections.ReadModelStore
}

// NewGetFilterHandler creates a new handler
func NewGetFilterHandler(readModelStore projections.ReadModelStore) *GetFilterHandler {
	return &GetFilterHandler{
		readModelStore: readModelStore,
	}
}

// Handle processes the query
func (h *GetFilterHandler) Handle(ctx context.Context, query interface{}) (interface{}, error) {
	q, ok := query.(*models.GetFilterQuery)
	if !ok {
		return nil, fmt.Errorf("invalid query type")
	}

	// Get read model
	var readModel projections.TrafficControlReadModel
	modelID := fmt.Sprintf("tc:%s", q.DeviceName)

	if err := h.readModelStore.Get(ctx, "traffic-control", modelID, &readModel); err != nil {
		return nil, fmt.Errorf("failed to get read model: %w", err)
	}

	// Find specific filter
	filterID := fmt.Sprintf("%s:%d:%s", q.Parent, q.Priority, q.Handle)
	for _, filter := range readModel.Filters {
		if filter.ID == filterID {
			return &models.FilterView{
				DeviceName: readModel.DeviceName,
				Parent:     filter.Parent,
				Priority:   filter.Priority,
				Handle:     filter.Handle,
				Protocol:   filter.Protocol,
				FlowID:     filter.FlowID,
				Matches:    filter.Matches,
			}, nil
		}
	}

	return nil, fmt.Errorf("filter not found")
}

// GetConfigurationHandler handles configuration queries using read models
type GetConfigurationHandler struct {
	readModelStore projections.ReadModelStore
}

// NewGetConfigurationHandler creates a new handler
func NewGetConfigurationHandler(readModelStore projections.ReadModelStore) *GetConfigurationHandler {
	return &GetConfigurationHandler{
		readModelStore: readModelStore,
	}
}

// Handle processes the query
func (h *GetConfigurationHandler) Handle(ctx context.Context, query interface{}) (interface{}, error) {
	q, ok := query.(*models.GetConfigurationQuery)
	if !ok {
		return nil, fmt.Errorf("invalid query type")
	}

	// Get read model
	var readModel projections.TrafficControlReadModel
	modelID := fmt.Sprintf("tc:%s", q.DeviceName)

	if err := h.readModelStore.Get(ctx, "traffic-control", modelID, &readModel); err != nil {
		return nil, fmt.Errorf("failed to get read model: %w", err)
	}

	// Convert to configuration view
	config := &models.ConfigurationView{
		DeviceName: readModel.DeviceName,
		Qdiscs:     make([]models.QdiscView, 0, len(readModel.Qdiscs)),
		Classes:    make([]models.ClassView, 0, len(readModel.Classes)),
		Filters:    make([]models.FilterView, 0, len(readModel.Filters)),
		Version:    readModel.Version,
	}

	// Convert qdiscs
	for _, qdisc := range readModel.Qdiscs {
		config.Qdiscs = append(config.Qdiscs, models.QdiscView{
			DeviceName:   readModel.DeviceName,
			Handle:       qdisc.Handle,
			Type:         qdisc.Type,
			Parent:       qdisc.Parent,
			DefaultClass: qdisc.DefaultClass,
			Parameters:   qdisc.Parameters,
		})
	}

	// Convert classes
	for _, class := range readModel.Classes {
		config.Classes = append(config.Classes, models.ClassView{
			DeviceName: readModel.DeviceName,
			Handle:     class.Handle,
			Parent:     class.Parent,
			Type:       class.Type,
			Name:       class.Name,
			Rate:       class.Rate,
			Ceil:       class.Ceil,
			Priority:   class.Priority,
			Parameters: class.Parameters,
		})
	}

	// Convert filters
	for _, filter := range readModel.Filters {
		config.Filters = append(config.Filters, models.FilterView{
			DeviceName: readModel.DeviceName,
			Parent:     filter.Parent,
			Priority:   filter.Priority,
			Handle:     filter.Handle,
			Protocol:   filter.Protocol,
			FlowID:     filter.FlowID,
			Matches:    filter.Matches,
		})
	}

	return config, nil
}
