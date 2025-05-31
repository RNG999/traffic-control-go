package events

import (
	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
)

// FilterCreatedEvent is emitted when a filter is created
type FilterCreatedEvent struct {
	BaseEvent
	DeviceName valueobjects.DeviceName
	Parent     valueobjects.Handle
	Priority   uint16
	Handle     valueobjects.Handle
	FlowID     valueobjects.Handle
	Protocol   entities.Protocol
	Matches    []MatchData
}

// MatchData represents serializable match data
type MatchData struct {
	Type  entities.MatchType
	Value string
}

// NewFilterCreatedEvent creates a new FilterCreatedEvent
func NewFilterCreatedEvent(aggregateID string, version int, device valueobjects.DeviceName, parent valueobjects.Handle, priority uint16, handle valueobjects.Handle, flowID valueobjects.Handle) *FilterCreatedEvent {
	return &FilterCreatedEvent{
		BaseEvent:  NewBaseEvent(aggregateID, "FilterCreated", version),
		DeviceName: device,
		Parent:     parent,
		Priority:   priority,
		Handle:     handle,
		FlowID:     flowID,
		Protocol:   entities.ProtocolIP,
		Matches:    make([]MatchData, 0),
	}
}

// AddMatch adds a match to the event
func (e *FilterCreatedEvent) AddMatch(matchType entities.MatchType, value string) {
	e.Matches = append(e.Matches, MatchData{
		Type:  matchType,
		Value: value,
	})
}

// FilterDeletedEvent is emitted when a filter is deleted
type FilterDeletedEvent struct {
	BaseEvent
	DeviceName valueobjects.DeviceName
	Parent     valueobjects.Handle
	Priority   uint16
	Handle     valueobjects.Handle
}

// NewFilterDeletedEvent creates a new FilterDeletedEvent
func NewFilterDeletedEvent(aggregateID string, version int, device valueobjects.DeviceName, parent valueobjects.Handle, priority uint16, handle valueobjects.Handle) *FilterDeletedEvent {
	return &FilterDeletedEvent{
		BaseEvent:  NewBaseEvent(aggregateID, "FilterDeleted", version),
		DeviceName: device,
		Parent:     parent,
		Priority:   priority,
		Handle:     handle,
	}
}

// FilterModifiedEvent is emitted when a filter is modified
type FilterModifiedEvent struct {
	BaseEvent
	DeviceName valueobjects.DeviceName
	Parent     valueobjects.Handle
	Priority   uint16
	Handle     valueobjects.Handle
	NewFlowID  *valueobjects.Handle
	NewMatches []MatchData
}

// NewFilterModifiedEvent creates a new FilterModifiedEvent
func NewFilterModifiedEvent(aggregateID string, version int, device valueobjects.DeviceName, parent valueobjects.Handle, priority uint16, handle valueobjects.Handle) *FilterModifiedEvent {
	return &FilterModifiedEvent{
		BaseEvent:  NewBaseEvent(aggregateID, "FilterModified", version),
		DeviceName: device,
		Parent:     parent,
		Priority:   priority,
		Handle:     handle,
		NewMatches: make([]MatchData, 0),
	}
}

// SetNewFlowID sets a new flow ID
func (e *FilterModifiedEvent) SetNewFlowID(flowID valueobjects.Handle) {
	e.NewFlowID = &flowID
}

// AddNewMatch adds a new match to the modification
func (e *FilterModifiedEvent) AddNewMatch(matchType entities.MatchType, value string) {
	e.NewMatches = append(e.NewMatches, MatchData{
		Type:  matchType,
		Value: value,
	})
}