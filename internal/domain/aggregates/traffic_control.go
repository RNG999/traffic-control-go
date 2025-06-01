package aggregates

import (
	"fmt"

	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/events"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
)

// TrafficControlAggregate is the root aggregate for traffic control configuration
type TrafficControlAggregate struct {
	// Aggregate identity
	id         string
	deviceName valueobjects.DeviceName

	// Current state
	qdiscs  map[valueobjects.Handle]*entities.Qdisc
	classes map[valueobjects.Handle]*entities.Class
	filters []*entities.Filter

	// Event sourcing
	version int
	changes []events.DomainEvent
}

// GetID returns the aggregate ID
func (a *TrafficControlAggregate) GetID() string {
	return a.id
}

// NewTrafficControlAggregate creates a new aggregate
func NewTrafficControlAggregate(deviceName valueobjects.DeviceName) *TrafficControlAggregate {
	return &TrafficControlAggregate{
		id:         fmt.Sprintf("tc:%s", deviceName),
		deviceName: deviceName,
		qdiscs:     make(map[valueobjects.Handle]*entities.Qdisc),
		classes:    make(map[valueobjects.Handle]*entities.Class),
		filters:    make([]*entities.Filter, 0),
		version:    0,
		changes:    make([]events.DomainEvent, 0),
	}
}

// FromEvents reconstructs an aggregate from events
func FromEvents(deviceName valueobjects.DeviceName, eventList []events.DomainEvent) *TrafficControlAggregate {
	aggregate := NewTrafficControlAggregate(deviceName)

	for _, event := range eventList {
		aggregate.ApplyEvent(event)
		aggregate.version++
	}

	// Clear changes as these are already committed events
	aggregate.changes = make([]events.DomainEvent, 0)

	return aggregate
}

// ID returns the aggregate ID
func (tc *TrafficControlAggregate) ID() string {
	return tc.id
}

// Version returns the current version
func (tc *TrafficControlAggregate) Version() int {
	return tc.version
}

// DeviceName returns the device name
func (tc *TrafficControlAggregate) DeviceName() valueobjects.DeviceName {
	return tc.deviceName
}

// AddHTBQdisc adds an HTB qdisc
func (tc *TrafficControlAggregate) AddHTBQdisc(handle valueobjects.Handle, defaultClass valueobjects.Handle) error {
	// Business rule: Check if qdisc already exists
	if _, exists := tc.qdiscs[handle]; exists {
		return fmt.Errorf("qdisc with handle %s already exists", handle)
	}

	// Business rule: Root qdisc must have minor = 0
	if !handle.IsRoot() {
		return fmt.Errorf("root qdisc handle must have minor = 0, got %s", handle)
	}

	// Create and apply event
	event := events.NewHTBQdiscCreatedEvent(
		tc.id,
		tc.version+1,
		tc.deviceName,
		handle,
		defaultClass,
	)

	tc.ApplyEvent(event)
	tc.changes = append(tc.changes, event)
	tc.version++

	return nil
}

// AddTBFQdisc adds a TBF qdisc
func (tc *TrafficControlAggregate) AddTBFQdisc(handle valueobjects.Handle, rate valueobjects.Bandwidth, buffer, limit, burst uint32) error {
	// Business rule: Check if qdisc already exists
	if _, exists := tc.qdiscs[handle]; exists {
		return fmt.Errorf("qdisc with handle %s already exists", handle)
	}

	// Business rule: Root qdisc must have minor = 0
	if !handle.IsRoot() {
		return fmt.Errorf("root qdisc handle must have minor = 0, got %s", handle)
	}

	// Business rule: Rate must be positive
	if rate.BitsPerSecond() <= 0 {
		return fmt.Errorf("rate must be positive, got %s", rate)
	}

	// Create and apply event
	event := events.NewTBFQdiscCreatedEvent(
		tc.id,
		tc.version+1,
		tc.deviceName,
		handle,
		rate,
		buffer,
		limit,
		burst,
	)

	tc.ApplyEvent(event)
	tc.changes = append(tc.changes, event)
	tc.version++

	return nil
}

// AddPRIOQdisc adds a PRIO qdisc
func (tc *TrafficControlAggregate) AddPRIOQdisc(handle valueobjects.Handle, bands uint8, priomap []uint8) error {
	// Business rule: Check if qdisc already exists
	if _, exists := tc.qdiscs[handle]; exists {
		return fmt.Errorf("qdisc with handle %s already exists", handle)
	}

	// Business rule: Root qdisc must have minor = 0
	if !handle.IsRoot() {
		return fmt.Errorf("root qdisc handle must have minor = 0, got %s", handle)
	}

	// Business rule: Bands must be between 2 and 16
	if bands < 2 || bands > 16 {
		return fmt.Errorf("bands must be between 2 and 16, got %d", bands)
	}

	// Business rule: Priomap must have 16 elements
	if len(priomap) != 16 {
		return fmt.Errorf("priomap must have 16 elements, got %d", len(priomap))
	}

	// Business rule: All priomap values must be < bands
	for i, p := range priomap {
		if p >= bands {
			return fmt.Errorf("priomap[%d] = %d must be < bands (%d)", i, p, bands)
		}
	}

	// Create and apply event
	event := events.NewPRIOQdiscCreatedEvent(
		tc.id,
		tc.version+1,
		tc.deviceName,
		handle,
		bands,
		priomap,
	)

	tc.ApplyEvent(event)
	tc.changes = append(tc.changes, event)
	tc.version++

	return nil
}

// AddFQCODELQdisc adds a FQ_CODEL qdisc
func (tc *TrafficControlAggregate) AddFQCODELQdisc(handle valueobjects.Handle, limit, flows, target, interval, quantum uint32, ecn bool) error {
	// Business rule: Check if qdisc already exists
	if _, exists := tc.qdiscs[handle]; exists {
		return fmt.Errorf("qdisc with handle %s already exists", handle)
	}

	// Business rule: Root qdisc must have minor = 0
	if !handle.IsRoot() {
		return fmt.Errorf("root qdisc handle must have minor = 0, got %s", handle)
	}

	// Business rule: Limit must be positive
	if limit == 0 {
		return fmt.Errorf("limit must be positive, got %d", limit)
	}

	// Business rule: Flows must be positive and power of 2
	if flows == 0 || (flows&(flows-1)) != 0 {
		return fmt.Errorf("flows must be positive and power of 2, got %d", flows)
	}

	// Business rule: Target must be positive
	if target == 0 {
		return fmt.Errorf("target must be positive, got %d microseconds", target)
	}

	// Business rule: Interval must be positive and >= target
	if interval == 0 || interval < target {
		return fmt.Errorf("interval must be positive and >= target (%d), got %d microseconds", target, interval)
	}

	// Create and apply event
	event := events.NewFQCODELQdiscCreatedEvent(
		tc.id,
		tc.version+1,
		tc.deviceName,
		handle,
		limit,
		flows,
		target,
		interval,
		quantum,
		ecn,
	)

	tc.ApplyEvent(event)
	tc.changes = append(tc.changes, event)
	tc.version++

	return nil
}

// AddHTBClass adds an HTB class
func (tc *TrafficControlAggregate) AddHTBClass(parent valueobjects.Handle, classHandle valueobjects.Handle, name string, rate valueobjects.Bandwidth, ceil valueobjects.Bandwidth) error {
	// Business rule: Parent qdisc must exist
	parentQdisc, parentExists := tc.qdiscs[parent]
	if !parentExists {
		// Check if parent is a class
		if _, classExists := tc.classes[parent]; !classExists {
			return fmt.Errorf("parent %s does not exist", parent)
		}
	}

	// Business rule: Class handle must not already exist
	if _, exists := tc.classes[classHandle]; exists {
		return fmt.Errorf("class with handle %s already exists", classHandle)
	}

	// Business rule: HTB specific - parent must be HTB
	if parentExists && parentQdisc.Type() != entities.QdiscTypeHTB {
		return fmt.Errorf("parent qdisc must be HTB type")
	}

	// Business rule: Ceil must be >= Rate
	if ceil.BitsPerSecond() > 0 && ceil.LessThan(rate) {
		return fmt.Errorf("ceil (%s) cannot be less than rate (%s)", ceil, rate)
	}

	// Create and apply event
	event := events.NewHTBClassCreatedEvent(
		tc.id,
		tc.version+1,
		tc.deviceName,
		classHandle,
		parent,
		name,
		rate,
		ceil,
	)

	tc.ApplyEvent(event)
	tc.changes = append(tc.changes, event)
	tc.version++

	return nil
}

// AddFilter adds a filter
func (tc *TrafficControlAggregate) AddFilter(parent valueobjects.Handle, priority uint16, handle valueobjects.Handle, flowID valueobjects.Handle, matches []entities.Match) error {
	// Business rule: Parent must exist (either qdisc or class)
	_, qdiscExists := tc.qdiscs[parent]
	_, classExists := tc.classes[parent]
	if !qdiscExists && !classExists {
		return fmt.Errorf("parent %s does not exist", parent)
	}

	// Business rule: Target class (flowID) must exist
	if _, exists := tc.classes[flowID]; !exists {
		return fmt.Errorf("target class %s does not exist", flowID)
	}

	// Create event
	event := events.NewFilterCreatedEvent(
		tc.id,
		tc.version+1,
		tc.deviceName,
		parent,
		priority,
		handle,
		flowID,
	)

	// Add matches to event
	for _, match := range matches {
		event.AddMatch(match.Type(), match.String())
	}

	tc.ApplyEvent(event)
	tc.changes = append(tc.changes, event)
	tc.version++

	return nil
}

// GetUncommittedEvents returns uncommitted events
func (tc *TrafficControlAggregate) GetUncommittedEvents() []events.DomainEvent {
	return tc.changes
}

// MarkEventsAsCommitted clears uncommitted events
func (tc *TrafficControlAggregate) MarkEventsAsCommitted() {
	tc.changes = make([]events.DomainEvent, 0)
}

// LoadFromHistory rebuilds aggregate state from events
func (tc *TrafficControlAggregate) LoadFromHistory(history []events.DomainEvent) {
	for _, event := range history {
		tc.ApplyEvent(event)
		tc.version++
	}
	// Clear changes as these are already committed
	tc.changes = make([]events.DomainEvent, 0)
}

// GetVersion returns the current version
func (tc *TrafficControlAggregate) GetVersion() int {
	return tc.version
}

// ApplyEvent applies a domain event to update aggregate state
func (tc *TrafficControlAggregate) ApplyEvent(event events.DomainEvent) {
	switch e := event.(type) {
	case *events.HTBQdiscCreatedEvent:
		qdisc := entities.NewHTBQdisc(e.DeviceName, e.Handle, e.DefaultClass)
		tc.qdiscs[e.Handle] = qdisc.Qdisc

	case *events.TBFQdiscCreatedEvent:
		qdisc := entities.NewTBFQdisc(e.DeviceName, e.Handle, e.Rate)
		qdisc.SetBuffer(e.Buffer)
		qdisc.SetLimit(e.Limit)
		qdisc.SetBurst(e.Burst)
		tc.qdiscs[e.Handle] = qdisc.Qdisc

	case *events.PRIOQdiscCreatedEvent:
		qdisc := entities.NewPRIOQdisc(e.DeviceName, e.Handle, e.Bands)
		qdisc.SetPriomap(e.Priomap)
		tc.qdiscs[e.Handle] = qdisc.Qdisc

	case *events.FQCODELQdiscCreatedEvent:
		qdisc := entities.NewFQCODELQdisc(e.DeviceName, e.Handle)
		qdisc.SetLimit(e.Limit)
		qdisc.SetFlows(e.Flows)
		qdisc.SetTarget(e.Target)
		qdisc.SetInterval(e.Interval)
		qdisc.SetQuantum(e.Quantum)
		qdisc.SetECN(e.ECN)
		tc.qdiscs[e.Handle] = qdisc.Qdisc

	case *events.HTBClassCreatedEvent:
		// Use a default priority of 4 for event reconstruction
		class := entities.NewHTBClass(e.DeviceName, e.Handle, e.Parent, e.Name, entities.Priority(4))
		class.SetRate(e.Rate)
		class.SetCeil(e.Ceil)
		tc.classes[e.Handle] = class.Class

	case *events.FilterCreatedEvent:
		filter := entities.NewFilter(e.DeviceName, e.Parent, e.Priority, e.Handle)
		filter.SetFlowID(e.FlowID)
		filter.SetProtocol(e.Protocol)

		// Reconstruct matches from event data
		for _, matchData := range e.Matches {
			// This is simplified - in real implementation, we'd deserialize properly
			switch matchData.Type {
			case entities.MatchTypeIPDestination:
				if ipMatch, err := entities.NewIPDestinationMatch(matchData.Value); err == nil {
					filter.AddMatch(ipMatch)
				}
			case entities.MatchTypePortDestination:
				// Parse port from string - simplified
				// In real implementation, store structured data in events
			}
		}

		tc.filters = append(tc.filters, filter)

	case *events.QdiscDeletedEvent:
		delete(tc.qdiscs, e.Handle)

	case *events.ClassDeletedEvent:
		delete(tc.classes, e.Handle)
	}
}

// GetUncommittedChanges returns events that haven't been persisted
func (tc *TrafficControlAggregate) GetUncommittedChanges() []events.DomainEvent {
	return tc.changes
}

// MarkChangesAsCommitted clears the uncommitted changes
func (tc *TrafficControlAggregate) MarkChangesAsCommitted() {
	tc.changes = make([]events.DomainEvent, 0)
}

// GetQdiscs returns all qdiscs (for queries)
func (tc *TrafficControlAggregate) GetQdiscs() map[valueobjects.Handle]*entities.Qdisc {
	// Return a copy to maintain immutability
	result := make(map[valueobjects.Handle]*entities.Qdisc)
	for k, v := range tc.qdiscs {
		result[k] = v
	}
	return result
}

// GetClasses returns all classes (for queries)
func (tc *TrafficControlAggregate) GetClasses() map[valueobjects.Handle]*entities.Class {
	// Return a copy to maintain immutability
	result := make(map[valueobjects.Handle]*entities.Class)
	for k, v := range tc.classes {
		result[k] = v
	}
	return result
}

// GetFilters returns all filters (for queries)
func (tc *TrafficControlAggregate) GetFilters() []*entities.Filter {
	// Return a copy to maintain immutability
	result := make([]*entities.Filter, len(tc.filters))
	copy(result, tc.filters)
	return result
}
