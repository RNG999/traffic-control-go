package handlers

import (
	"context"
	"fmt"

	"github.com/rng999/traffic-control-go/internal/domain/aggregates"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/internal/queries/models"
)

// GetQdiscByDeviceHandler handles queries for qdiscs
type GetQdiscByDeviceHandler struct {
	eventStore eventstore.EventStore
}

// NewGetQdiscByDeviceHandler creates a new handler
func NewGetQdiscByDeviceHandler(eventStore eventstore.EventStore) *GetQdiscByDeviceHandler {
	return &GetQdiscByDeviceHandler{
		eventStore: eventStore,
	}
}

// Handle processes the query
func (h *GetQdiscByDeviceHandler) Handle(ctx context.Context, query interface{}) (interface{}, error) {
	// Type assert the query
	qdiscQuery, ok := query.(*models.GetQdiscByDeviceQuery)
	if !ok {
		return nil, fmt.Errorf("invalid query type: expected *GetQdiscByDeviceQuery, got %T", query)
	}

	// Load aggregate
	aggregateID := fmt.Sprintf("tc:%s", qdiscQuery.DeviceName())
	events, err := h.eventStore.GetEvents(aggregateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load events: %w", err)
	}

	// Reconstruct aggregate from events
	aggregate := aggregates.FromEvents(qdiscQuery.DeviceName(), events)

	// Convert to view models
	var views []models.QdiscView
	for _, qdisc := range aggregate.GetQdiscs() {
		view := models.NewQdiscView(qdiscQuery.DeviceName(), qdisc)
		views = append(views, view)
	}

	return views, nil
}

// GetClassesByDeviceHandler handles queries for classes
type GetClassesByDeviceHandler struct {
	eventStore eventstore.EventStore
}

// NewGetClassesByDeviceHandler creates a new handler
func NewGetClassesByDeviceHandler(eventStore eventstore.EventStore) *GetClassesByDeviceHandler {
	return &GetClassesByDeviceHandler{
		eventStore: eventStore,
	}
}

// Handle processes the query
func (h *GetClassesByDeviceHandler) Handle(ctx context.Context, query interface{}) (interface{}, error) {
	// Type assert the query
	classQuery, ok := query.(*models.GetClassesByDeviceQuery)
	if !ok {
		return nil, fmt.Errorf("invalid query type: expected *GetClassesByDeviceQuery, got %T", query)
	}

	// Load aggregate
	aggregateID := fmt.Sprintf("tc:%s", classQuery.DeviceName())
	events, err := h.eventStore.GetEvents(aggregateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load events: %w", err)
	}

	// Reconstruct aggregate from events
	aggregate := aggregates.FromEvents(classQuery.DeviceName(), events)

	// Convert to view models
	var views []models.ClassView
	for _, class := range aggregate.GetClasses() {
		view := models.NewClassView(classQuery.DeviceName(), class)
		views = append(views, view)
	}

	return views, nil
}

// GetFiltersByDeviceHandler handles queries for filters
type GetFiltersByDeviceHandler struct {
	eventStore eventstore.EventStore
}

// NewGetFiltersByDeviceHandler creates a new handler
func NewGetFiltersByDeviceHandler(eventStore eventstore.EventStore) *GetFiltersByDeviceHandler {
	return &GetFiltersByDeviceHandler{
		eventStore: eventStore,
	}
}

// Handle processes the query
func (h *GetFiltersByDeviceHandler) Handle(ctx context.Context, query interface{}) (interface{}, error) {
	// Type assert the query
	filterQuery, ok := query.(*models.GetFiltersByDeviceQuery)
	if !ok {
		return nil, fmt.Errorf("invalid query type: expected *GetFiltersByDeviceQuery, got %T", query)
	}

	// Load aggregate
	aggregateID := fmt.Sprintf("tc:%s", filterQuery.DeviceName())
	events, err := h.eventStore.GetEvents(aggregateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load events: %w", err)
	}

	// Reconstruct aggregate from events
	aggregate := aggregates.FromEvents(filterQuery.DeviceName(), events)

	// Convert to view models
	var views []models.FilterView
	for _, filter := range aggregate.GetFilters() {
		view := models.NewFilterView(filterQuery.DeviceName(), filter)
		views = append(views, view)
	}

	return views, nil
}

// GetTrafficControlConfigHandler handles queries for complete configuration
type GetTrafficControlConfigHandler struct {
	eventStore eventstore.EventStore
}

// NewGetTrafficControlConfigHandler creates a new handler
func NewGetTrafficControlConfigHandler(eventStore eventstore.EventStore) *GetTrafficControlConfigHandler {
	return &GetTrafficControlConfigHandler{
		eventStore: eventStore,
	}
}

// Handle processes the query
func (h *GetTrafficControlConfigHandler) Handle(ctx context.Context, query interface{}) (interface{}, error) {
	// Type assert the query
	configQuery, ok := query.(*models.GetTrafficControlConfigQuery)
	if !ok {
		return nil, fmt.Errorf("invalid query type: expected *GetTrafficControlConfigQuery, got %T", query)
	}

	// Load aggregate
	aggregateID := fmt.Sprintf("tc:%s", configQuery.DeviceName())
	events, err := h.eventStore.GetEvents(aggregateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load events: %w", err)
	}

	// Reconstruct aggregate from events
	aggregate := aggregates.FromEvents(configQuery.DeviceName(), events)

	// Build complete view
	view := models.TrafficControlConfigView{
		DeviceName: configQuery.DeviceName().String(),
		Qdiscs:     make([]models.QdiscView, 0),
		Classes:    make([]models.ClassView, 0),
		Filters:    make([]models.FilterView, 0),
	}

	// Convert qdiscs
	for _, qdisc := range aggregate.GetQdiscs() {
		qdiscView := models.NewQdiscView(configQuery.DeviceName(), qdisc)
		view.Qdiscs = append(view.Qdiscs, qdiscView)
	}

	// Convert classes
	for _, class := range aggregate.GetClasses() {
		classView := models.NewClassView(configQuery.DeviceName(), class)
		view.Classes = append(view.Classes, classView)
	}

	// Convert filters
	for _, filter := range aggregate.GetFilters() {
		filterView := models.NewFilterView(configQuery.DeviceName(), filter)
		view.Filters = append(view.Filters, filterView)
	}

	return view, nil
}
