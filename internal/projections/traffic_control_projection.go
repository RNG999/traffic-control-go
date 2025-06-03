package projections

import (
	"context"
	"fmt"

	"github.com/rng999/traffic-control-go/internal/domain/events"
	"github.com/rng999/traffic-control-go/pkg/logging"
)

// TrafficControlReadModel represents the current state of traffic control configuration
type TrafficControlReadModel struct {
	DeviceName string            `json:"device_name"`
	Qdiscs     []QdiscReadModel  `json:"qdiscs"`
	Classes    []ClassReadModel  `json:"classes"`
	Filters    []FilterReadModel `json:"filters"`
	LastUpdate int64             `json:"last_update"`
	Version    int               `json:"version"`
}

// QdiscReadModel represents a qdisc in the read model
type QdiscReadModel struct {
	Handle       string                 `json:"handle"`
	Type         string                 `json:"type"`
	Parent       string                 `json:"parent,omitempty"`
	DefaultClass string                 `json:"default_class,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
}

// ClassReadModel represents a class in the read model
type ClassReadModel struct {
	Handle     string                 `json:"handle"`
	Parent     string                 `json:"parent"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Rate       string                 `json:"rate"`
	Ceil       string                 `json:"ceil"`
	Priority   int                    `json:"priority,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// FilterReadModel represents a filter in the read model
type FilterReadModel struct {
	ID       string            `json:"id"`
	Parent   string            `json:"parent"`
	Priority uint16            `json:"priority"`
	Handle   string            `json:"handle"`
	Protocol string            `json:"protocol"`
	FlowID   string            `json:"flow_id"`
	Matches  map[string]string `json:"matches"`
}

// TrafficControlProjection builds read models from traffic control events
type TrafficControlProjection struct {
	store  ReadModelStore
	logger logging.Logger
}

// NewTrafficControlProjection creates a new traffic control projection
func NewTrafficControlProjection(store ReadModelStore) *TrafficControlProjection {
	return &TrafficControlProjection{
		store:  store,
		logger: logging.WithComponent("projection.traffic-control"),
	}
}

// GetName returns the projection name
func (p *TrafficControlProjection) GetName() string {
	return "traffic-control"
}

// Handle processes an event to update the read model
func (p *TrafficControlProjection) Handle(ctx context.Context, event events.DomainEvent) error {
	switch e := event.(type) {
	case *events.HTBQdiscCreatedEvent:
		return p.handleQdiscCreated(ctx, e)
	case *events.HTBClassCreatedEvent:
		return p.handleClassCreated(ctx, e)
	case *events.FilterCreatedEvent:
		return p.handleFilterCreated(ctx, e)
	default:
		// Unknown event type, ignore
		return nil
	}
}

// Reset clears the projection state
func (p *TrafficControlProjection) Reset(ctx context.Context) error {
	return p.store.Clear(ctx, "traffic-control")
}

func (p *TrafficControlProjection) handleQdiscCreated(ctx context.Context, event *events.HTBQdiscCreatedEvent) error {
	// Get or create read model
	var model TrafficControlReadModel
	modelID := fmt.Sprintf("tc:%s", event.DeviceName)

	if err := p.store.Get(ctx, "traffic-control", modelID, &model); err != nil {
		// Create new model if not exists
		model = TrafficControlReadModel{
			DeviceName: event.DeviceName.String(),
			Qdiscs:     make([]QdiscReadModel, 0),
			Classes:    make([]ClassReadModel, 0),
			Filters:    make([]FilterReadModel, 0),
		}
	}

	// Add qdisc to model
	qdisc := QdiscReadModel{
		Handle:       event.Handle.String(),
		Type:         "htb",
		DefaultClass: event.DefaultClass.String(),
		Parameters:   make(map[string]interface{}),
	}

	// Check if qdisc already exists and update it
	found := false
	for i, q := range model.Qdiscs {
		if q.Handle == qdisc.Handle {
			model.Qdiscs[i] = qdisc
			found = true
			break
		}
	}
	if !found {
		model.Qdiscs = append(model.Qdiscs, qdisc)
	}

	// Update metadata
	model.LastUpdate = event.Timestamp().Unix()
	model.Version = event.EventVersion()

	// Save updated model
	return p.store.Save(ctx, "traffic-control", modelID, &model)
}

func (p *TrafficControlProjection) handleClassCreated(ctx context.Context, event *events.HTBClassCreatedEvent) error {
	// Get or create read model
	var model TrafficControlReadModel
	modelID := fmt.Sprintf("tc:%s", event.DeviceName)

	if err := p.store.Get(ctx, "traffic-control", modelID, &model); err != nil {
		// Create new model if not exists
		model = TrafficControlReadModel{
			DeviceName: event.DeviceName.String(),
			Qdiscs:     make([]QdiscReadModel, 0),
			Classes:    make([]ClassReadModel, 0),
			Filters:    make([]FilterReadModel, 0),
		}
	}

	// Add class to model
	class := ClassReadModel{
		Handle:     event.Handle.String(),
		Parent:     event.Parent.String(),
		Type:       "htb",
		Name:       event.Name,
		Rate:       event.Rate.String(),
		Ceil:       event.Ceil.String(),
		Parameters: make(map[string]interface{}),
	}

	// Check if class already exists and update it
	found := false
	for i, c := range model.Classes {
		if c.Handle == class.Handle {
			model.Classes[i] = class
			found = true
			break
		}
	}
	if !found {
		model.Classes = append(model.Classes, class)
	}

	// Update metadata
	model.LastUpdate = event.Timestamp().Unix()
	model.Version = event.EventVersion()

	// Save updated model
	return p.store.Save(ctx, "traffic-control", modelID, &model)
}

func (p *TrafficControlProjection) handleFilterCreated(ctx context.Context, event *events.FilterCreatedEvent) error {
	// Get or create read model
	var model TrafficControlReadModel
	modelID := fmt.Sprintf("tc:%s", event.DeviceName)

	if err := p.store.Get(ctx, "traffic-control", modelID, &model); err != nil {
		// Create new model if not exists
		model = TrafficControlReadModel{
			DeviceName: event.DeviceName.String(),
			Qdiscs:     make([]QdiscReadModel, 0),
			Classes:    make([]ClassReadModel, 0),
			Filters:    make([]FilterReadModel, 0),
		}
	}

	// Add filter to model
	filter := FilterReadModel{
		ID:       fmt.Sprintf("%s:%d:%s", event.Parent.String(), event.Priority, event.Handle.String()),
		Parent:   event.Parent.String(),
		Priority: event.Priority,
		Handle:   event.Handle.String(),
		Protocol: "ip",
		FlowID:   event.FlowID.String(),
		Matches:  convertMatchData(event.Matches),
	}

	// Check if filter already exists and update it
	found := false
	for i, f := range model.Filters {
		if f.ID == filter.ID {
			model.Filters[i] = filter
			found = true
			break
		}
	}
	if !found {
		model.Filters = append(model.Filters, filter)
	}

	// Update metadata
	model.LastUpdate = event.Timestamp().Unix()
	model.Version = event.EventVersion()

	// Save updated model
	return p.store.Save(ctx, "traffic-control", modelID, &model)
}
