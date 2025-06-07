package events

import (
	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// ClassCreatedEvent is emitted when a traffic class is created
type ClassCreatedEvent struct {
	BaseEvent
	DeviceName tc.DeviceName
	Handle     tc.Handle
	Parent     tc.Handle
	Name       string
	Priority   entities.Priority
}

// NewClassCreatedEvent creates a new ClassCreatedEvent
func NewClassCreatedEvent(aggregateID string, version int, device tc.DeviceName, handle tc.Handle, parent tc.Handle, name string, priority entities.Priority) *ClassCreatedEvent {
	return &ClassCreatedEvent{
		BaseEvent:  NewBaseEvent(aggregateID, "ClassCreated", version),
		DeviceName: device,
		Handle:     handle,
		Parent:     parent,
		Name:       name,
		Priority:   priority,
	}
}

// HTBClassCreatedEvent is emitted when an HTB class is created
type HTBClassCreatedEvent struct {
	BaseEvent
	DeviceName tc.DeviceName
	Handle     tc.Handle
	Parent     tc.Handle
	Name       string
	Rate       tc.Bandwidth
	Ceil       tc.Bandwidth
	Burst      uint32
	Cburst     uint32
}

// NewHTBClassCreatedEvent creates a new HTBClassCreatedEvent
func NewHTBClassCreatedEvent(aggregateID string, version int, device tc.DeviceName, handle tc.Handle, parent tc.Handle, name string, rate tc.Bandwidth, ceil tc.Bandwidth) *HTBClassCreatedEvent {
	return &HTBClassCreatedEvent{
		BaseEvent:  NewBaseEvent(aggregateID, "HTBClassCreated", version),
		DeviceName: device,
		Handle:     handle,
		Parent:     parent,
		Name:       name,
		Rate:       rate,
		Ceil:       ceil,
	}
}

// ClassDeletedEvent is emitted when a class is deleted
type ClassDeletedEvent struct {
	BaseEvent
	DeviceName tc.DeviceName
	Handle     tc.Handle
}

// NewClassDeletedEvent creates a new ClassDeletedEvent
func NewClassDeletedEvent(aggregateID string, version int, device tc.DeviceName, handle tc.Handle) *ClassDeletedEvent {
	return &ClassDeletedEvent{
		BaseEvent:  NewBaseEvent(aggregateID, "ClassDeleted", version),
		DeviceName: device,
		Handle:     handle,
	}
}

// ClassModifiedEvent is emitted when a class is modified
type ClassModifiedEvent struct {
	BaseEvent
	DeviceName tc.DeviceName
	Handle     tc.Handle
	Changes    map[string]interface{}
}

// NewClassModifiedEvent creates a new ClassModifiedEvent
func NewClassModifiedEvent(aggregateID string, version int, device tc.DeviceName, handle tc.Handle, changes map[string]interface{}) *ClassModifiedEvent {
	// Create a copy of changes to ensure immutability
	changesCopy := make(map[string]interface{})
	for k, v := range changes {
		changesCopy[k] = v
	}

	return &ClassModifiedEvent{
		BaseEvent:  NewBaseEvent(aggregateID, "ClassModified", version),
		DeviceName: device,
		Handle:     handle,
		Changes:    changesCopy,
	}
}

// ClassPriorityChangedEvent is emitted when a class priority is changed
type ClassPriorityChangedEvent struct {
	BaseEvent
	DeviceName  tc.DeviceName
	Handle      tc.Handle
	OldPriority entities.Priority
	NewPriority entities.Priority
}

// NewClassPriorityChangedEvent creates a new ClassPriorityChangedEvent
func NewClassPriorityChangedEvent(aggregateID string, version int, device tc.DeviceName, handle tc.Handle, oldPriority, newPriority entities.Priority) *ClassPriorityChangedEvent {
	return &ClassPriorityChangedEvent{
		BaseEvent:   NewBaseEvent(aggregateID, "ClassPriorityChanged", version),
		DeviceName:  device,
		Handle:      handle,
		OldPriority: oldPriority,
		NewPriority: newPriority,
	}
}
