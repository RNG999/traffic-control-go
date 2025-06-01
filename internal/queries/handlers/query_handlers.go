package handlers

import (
	"fmt"

	"github.com/rng999/traffic-control-go/internal/domain/aggregates"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/internal/queries/models"
	"github.com/rng999/traffic-control-go/pkg/types"
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
func (h *GetQdiscByDeviceHandler) Handle(query *models.GetQdiscByDeviceQuery) types.Result[[]models.QdiscView] {
	// Load aggregate
	aggregateID := fmt.Sprintf("tc:%s", query.DeviceName())
	events, err := h.eventStore.GetEvents(aggregateID)
	if err != nil {
		return types.Failure[[]models.QdiscView](fmt.Errorf("failed to load events: %w", err))
	}

	// Reconstruct aggregate from events
	aggregate := aggregates.FromEvents(query.DeviceName(), events)

	// Convert to view models
	var views []models.QdiscView
	for _, qdisc := range aggregate.GetQdiscs() {
		view := models.NewQdiscView(query.DeviceName(), qdisc)
		views = append(views, view)
	}

	return types.Success(views)
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
func (h *GetClassesByDeviceHandler) Handle(query *models.GetClassesByDeviceQuery) types.Result[[]models.ClassView] {
	// Load aggregate
	aggregateID := fmt.Sprintf("tc:%s", query.DeviceName())
	events, err := h.eventStore.GetEvents(aggregateID)
	if err != nil {
		return types.Failure[[]models.ClassView](fmt.Errorf("failed to load events: %w", err))
	}

	// Reconstruct aggregate from events
	aggregate := aggregates.FromEvents(query.DeviceName(), events)

	// Convert to view models
	var views []models.ClassView
	for _, class := range aggregate.GetClasses() {
		view := models.NewClassView(query.DeviceName(), class)
		views = append(views, view)
	}

	return types.Success(views)
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
func (h *GetFiltersByDeviceHandler) Handle(query *models.GetFiltersByDeviceQuery) types.Result[[]models.FilterView] {
	// Load aggregate
	aggregateID := fmt.Sprintf("tc:%s", query.DeviceName())
	events, err := h.eventStore.GetEvents(aggregateID)
	if err != nil {
		return types.Failure[[]models.FilterView](fmt.Errorf("failed to load events: %w", err))
	}

	// Reconstruct aggregate from events
	aggregate := aggregates.FromEvents(query.DeviceName(), events)

	// Convert to view models
	var views []models.FilterView
	for _, filter := range aggregate.GetFilters() {
		view := models.NewFilterView(query.DeviceName(), filter)
		views = append(views, view)
	}

	return types.Success(views)
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
func (h *GetTrafficControlConfigHandler) Handle(query *models.GetTrafficControlConfigQuery) types.Result[models.TrafficControlConfigView] {
	// Load aggregate
	aggregateID := fmt.Sprintf("tc:%s", query.DeviceName())
	events, err := h.eventStore.GetEvents(aggregateID)
	if err != nil {
		return types.Failure[models.TrafficControlConfigView](fmt.Errorf("failed to load events: %w", err))
	}

	// Reconstruct aggregate from events
	aggregate := aggregates.FromEvents(query.DeviceName(), events)

	// Build complete view
	view := models.TrafficControlConfigView{
		DeviceName: query.DeviceName().String(),
		Qdiscs:     make([]models.QdiscView, 0),
		Classes:    make([]models.ClassView, 0),
		Filters:    make([]models.FilterView, 0),
	}

	// Convert qdiscs
	for _, qdisc := range aggregate.GetQdiscs() {
		qdiscView := models.NewQdiscView(query.DeviceName(), qdisc)
		view.Qdiscs = append(view.Qdiscs, qdiscView)
	}

	// Convert classes
	for _, class := range aggregate.GetClasses() {
		classView := models.NewClassView(query.DeviceName(), class)
		view.Classes = append(view.Classes, classView)
	}

	// Convert filters
	for _, filter := range aggregate.GetFilters() {
		filterView := models.NewFilterView(query.DeviceName(), filter)
		view.Filters = append(view.Filters, filterView)
	}

	return types.Success[models.TrafficControlConfigView](view)
}