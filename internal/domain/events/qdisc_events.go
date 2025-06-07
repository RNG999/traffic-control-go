package events

import (
	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// QdiscCreatedEvent is emitted when a qdisc is created
type QdiscCreatedEvent struct {
	BaseEvent
	DeviceName tc.DeviceName
	Handle     tc.Handle
	QdiscType  entities.QdiscType
	Parent     *tc.Handle
}

// NewQdiscCreatedEvent creates a new QdiscCreatedEvent
func NewQdiscCreatedEvent(aggregateID string, version int, device tc.DeviceName, handle tc.Handle, qdiscType entities.QdiscType, parent *tc.Handle) *QdiscCreatedEvent {
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
	DeviceName   tc.DeviceName
	Handle       tc.Handle
	DefaultClass tc.Handle
	R2Q          uint32
}

// NewHTBQdiscCreatedEvent creates a new HTBQdiscCreatedEvent
func NewHTBQdiscCreatedEvent(aggregateID string, version int, device tc.DeviceName, handle tc.Handle, defaultClass tc.Handle) *HTBQdiscCreatedEvent {
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
	DeviceName tc.DeviceName
	Handle     tc.Handle
}

// NewQdiscDeletedEvent creates a new QdiscDeletedEvent
func NewQdiscDeletedEvent(aggregateID string, version int, device tc.DeviceName, handle tc.Handle) *QdiscDeletedEvent {
	return &QdiscDeletedEvent{
		BaseEvent:  NewBaseEvent(aggregateID, "QdiscDeleted", version),
		DeviceName: device,
		Handle:     handle,
	}
}

// QdiscModifiedEvent is emitted when a qdisc is modified
type QdiscModifiedEvent struct {
	BaseEvent
	DeviceName tc.DeviceName
	Handle     tc.Handle
	Parameters map[string]interface{}
}

// NewQdiscModifiedEvent creates a new QdiscModifiedEvent
func NewQdiscModifiedEvent(aggregateID string, version int, device tc.DeviceName, handle tc.Handle, parameters map[string]interface{}) *QdiscModifiedEvent {
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

// TBFQdiscCreatedEvent is emitted when a TBF qdisc is created
type TBFQdiscCreatedEvent struct {
	BaseEvent
	DeviceName tc.DeviceName
	Handle     tc.Handle
	Rate       tc.Bandwidth
	Buffer     uint32
	Limit      uint32
	Burst      uint32
}

// NewTBFQdiscCreatedEvent creates a new TBFQdiscCreatedEvent
func NewTBFQdiscCreatedEvent(aggregateID string, version int, device tc.DeviceName, handle tc.Handle, rate tc.Bandwidth, buffer, limit, burst uint32) *TBFQdiscCreatedEvent {
	return &TBFQdiscCreatedEvent{
		BaseEvent:  NewBaseEvent(aggregateID, "TBFQdiscCreated", version),
		DeviceName: device,
		Handle:     handle,
		Rate:       rate,
		Buffer:     buffer,
		Limit:      limit,
		Burst:      burst,
	}
}

// PRIOQdiscCreatedEvent is emitted when a PRIO qdisc is created
type PRIOQdiscCreatedEvent struct {
	BaseEvent
	DeviceName tc.DeviceName
	Handle     tc.Handle
	Bands      uint8
	Priomap    []uint8
}

// NewPRIOQdiscCreatedEvent creates a new PRIOQdiscCreatedEvent
func NewPRIOQdiscCreatedEvent(aggregateID string, version int, device tc.DeviceName, handle tc.Handle, bands uint8, priomap []uint8) *PRIOQdiscCreatedEvent {
	// Create a copy of priomap to ensure immutability
	priomapCopy := make([]uint8, len(priomap))
	copy(priomapCopy, priomap)

	return &PRIOQdiscCreatedEvent{
		BaseEvent:  NewBaseEvent(aggregateID, "PRIOQdiscCreated", version),
		DeviceName: device,
		Handle:     handle,
		Bands:      bands,
		Priomap:    priomapCopy,
	}
}

// FQCODELQdiscCreatedEvent is emitted when a FQ_CODEL qdisc is created
type FQCODELQdiscCreatedEvent struct {
	BaseEvent
	DeviceName tc.DeviceName
	Handle     tc.Handle
	Limit      uint32
	Flows      uint32
	Target     uint32
	Interval   uint32
	Quantum    uint32
	ECN        bool
}

// NewFQCODELQdiscCreatedEvent creates a new FQCODELQdiscCreatedEvent
func NewFQCODELQdiscCreatedEvent(aggregateID string, version int, device tc.DeviceName, handle tc.Handle, limit, flows, target, interval, quantum uint32, ecn bool) *FQCODELQdiscCreatedEvent {
	return &FQCODELQdiscCreatedEvent{
		BaseEvent:  NewBaseEvent(aggregateID, "FQCODELQdiscCreated", version),
		DeviceName: device,
		Handle:     handle,
		Limit:      limit,
		Flows:      flows,
		Target:     target,
		Interval:   interval,
		Quantum:    quantum,
		ECN:        ecn,
	}
}
