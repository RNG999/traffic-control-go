package events

import (
	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
)

// QdiscCreatedEvent is emitted when a qdisc is created
type QdiscCreatedEvent struct {
	BaseEvent
	DeviceName valueobjects.DeviceName
	Handle     valueobjects.Handle
	QdiscType  entities.QdiscType
	Parent     *valueobjects.Handle
}

// NewQdiscCreatedEvent creates a new QdiscCreatedEvent
func NewQdiscCreatedEvent(aggregateID string, version int, device valueobjects.DeviceName, handle valueobjects.Handle, qdiscType entities.QdiscType, parent *valueobjects.Handle) *QdiscCreatedEvent {
	return &QdiscCreatedEvent{
		BaseEvent:  NewBaseEvent(aggregateID, "QdiscCreated", version),
		DeviceName: device,
		Handle:     handle,
		QdiscType:  qdiscType,
		Parent:     parent,
	}
}

// HTBQdiscCreatedEvent is emitted when an HTB qdisc is created
type HTBQdiscCreatedEvent struct {
	BaseEvent
	DeviceName   valueobjects.DeviceName
	Handle       valueobjects.Handle
	DefaultClass valueobjects.Handle
	R2Q          uint32
}

// NewHTBQdiscCreatedEvent creates a new HTBQdiscCreatedEvent
func NewHTBQdiscCreatedEvent(aggregateID string, version int, device valueobjects.DeviceName, handle valueobjects.Handle, defaultClass valueobjects.Handle) *HTBQdiscCreatedEvent {
	return &HTBQdiscCreatedEvent{
		BaseEvent:    NewBaseEvent(aggregateID, "HTBQdiscCreated", version),
		DeviceName:   device,
		Handle:       handle,
		DefaultClass: defaultClass,
		R2Q:          10, // default value
	}
}

// QdiscDeletedEvent is emitted when a qdisc is deleted
type QdiscDeletedEvent struct {
	BaseEvent
	DeviceName valueobjects.DeviceName
	Handle     valueobjects.Handle
}

// NewQdiscDeletedEvent creates a new QdiscDeletedEvent
func NewQdiscDeletedEvent(aggregateID string, version int, device valueobjects.DeviceName, handle valueobjects.Handle) *QdiscDeletedEvent {
	return &QdiscDeletedEvent{
		BaseEvent:  NewBaseEvent(aggregateID, "QdiscDeleted", version),
		DeviceName: device,
		Handle:     handle,
	}
}

// QdiscModifiedEvent is emitted when a qdisc is modified
type QdiscModifiedEvent struct {
	BaseEvent
	DeviceName valueobjects.DeviceName
	Handle     valueobjects.Handle
	Parameters map[string]interface{}
}

// NewQdiscModifiedEvent creates a new QdiscModifiedEvent
func NewQdiscModifiedEvent(aggregateID string, version int, device valueobjects.DeviceName, handle valueobjects.Handle, parameters map[string]interface{}) *QdiscModifiedEvent {
	// Create a copy of parameters to ensure immutability
	paramsCopy := make(map[string]interface{})
	for k, v := range parameters {
		paramsCopy[k] = v
	}
	
	return &QdiscModifiedEvent{
		BaseEvent:  NewBaseEvent(aggregateID, "QdiscModified", version),
		DeviceName: device,
		Handle:     handle,
		Parameters: paramsCopy,
	}
}