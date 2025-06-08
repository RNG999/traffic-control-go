package aggregates

import (
	"fmt"

	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/events"
	"github.com/rng999/traffic-control-go/pkg/tc"
	"github.com/rng999/traffic-control-go/pkg/types"
)

// TrafficControlAggregate is the root aggregate for traffic control configuration
type TrafficControlAggregate struct {
	// Aggregate identity
	id         string
	deviceName tc.DeviceName

	// Current state
	qdiscs  map[tc.Handle]*entities.Qdisc
	classes map[tc.Handle]*entities.Class
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
func NewTrafficControlAggregate(deviceName tc.DeviceName) *TrafficControlAggregate {
	return &TrafficControlAggregate{
		id:         fmt.Sprintf("tc:%s", deviceName),
		deviceName: deviceName,
		qdiscs:     make(map[tc.Handle]*entities.Qdisc),
		classes:    make(map[tc.Handle]*entities.Class),
		filters:    make([]*entities.Filter, 0),
		version:    0,
		changes:    make([]events.DomainEvent, 0),
	}
}

// FromEvents reconstructs an aggregate from events
func FromEvents(deviceName tc.DeviceName, eventList []events.DomainEvent) *TrafficControlAggregate {
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
func (ag *TrafficControlAggregate) ID() string {
	return ag.id
}

// Version returns the current version
func (ag *TrafficControlAggregate) Version() int {
	return ag.version
}

// DeviceName returns the device name
func (ag *TrafficControlAggregate) DeviceName() tc.DeviceName {
	return ag.deviceName
}

// AddHTBQdisc adds an HTB qdisc (DEPRECATED: use WithHTBQdisc)
func (ag *TrafficControlAggregate) AddHTBQdisc(handle tc.Handle, defaultClass tc.Handle) error {
	// Business rule: Check if qdisc already exists
	if _, exists := ag.qdiscs[handle]; exists {
		return fmt.Errorf("qdisc with handle %s already exists", handle)
	}

	// Business rule: Root qdisc must have minor = 0
	if !handle.IsRoot() {
		return fmt.Errorf("root qdisc handle must have minor = 0, got %s", handle)
	}

	// Create and apply event
	event := events.NewHTBQdiscCreatedEvent(
		ag.id,
		ag.version+1,
		ag.deviceName,
		handle,
		defaultClass,
	)

	ag.ApplyEvent(event)
	ag.changes = append(ag.changes, event)
	ag.version++

	return nil
}

// WithHTBQdisc returns a new aggregate with an HTB qdisc added (immutable)
func (ag *TrafficControlAggregate) WithHTBQdisc(handle tc.Handle, defaultClass tc.Handle) types.Result[*TrafficControlAggregate] {
	// Business rule: Check if qdisc already exists
	if _, exists := ag.qdiscs[handle]; exists {
		return types.Failure[*TrafficControlAggregate](fmt.Errorf("qdisc with handle %s already exists", handle))
	}

	// Business rule: Root qdisc must have minor = 0
	if !handle.IsRoot() {
		return types.Failure[*TrafficControlAggregate](fmt.Errorf("root qdisc handle must have minor = 0, got %s", handle))
	}

	// Create event
	event := events.NewHTBQdiscCreatedEvent(
		ag.id,
		ag.version+1,
		ag.deviceName,
		handle,
		defaultClass,
	)

	// Create new aggregate with the event applied
	return types.Success(ag.withEvent(event))
}

// withEvent creates a new aggregate with an event applied (immutable helper)
func (ag *TrafficControlAggregate) withEvent(event events.DomainEvent) *TrafficControlAggregate {
	// Create a deep copy of the aggregate
	newAggregate := &TrafficControlAggregate{
		id:         ag.id,
		deviceName: ag.deviceName,
		qdiscs:     make(map[tc.Handle]*entities.Qdisc),
		classes:    make(map[tc.Handle]*entities.Class),
		filters:    make([]*entities.Filter, len(ag.filters)),
		version:    ag.version + 1,
		changes:    make([]events.DomainEvent, len(ag.changes)+1),
	}

	// Copy qdiscs
	for k, v := range ag.qdiscs {
		newAggregate.qdiscs[k] = v
	}

	// Copy classes
	for k, v := range ag.classes {
		newAggregate.classes[k] = v
	}

	// Copy filters
	copy(newAggregate.filters, ag.filters)

	// Copy existing changes and add new event
	copy(newAggregate.changes, ag.changes)
	newAggregate.changes[len(ag.changes)] = event

	// Apply the new event
	newAggregate.ApplyEvent(event)

	return newAggregate
}

// Chain enables functional composition of aggregate operations
func (ag *TrafficControlAggregate) Chain(operation func(*TrafficControlAggregate) types.Result[*TrafficControlAggregate]) types.Result[*TrafficControlAggregate] {
	return operation(ag)
}

// WithOperations applies multiple operations in sequence (functional composition)
func (ag *TrafficControlAggregate) WithOperations(operations ...func(*TrafficControlAggregate) types.Result[*TrafficControlAggregate]) types.Result[*TrafficControlAggregate] {
	result := types.Success(ag)

	for _, operation := range operations {
		result = result.FlatMap(operation)
		if result.IsFailure() {
			return result
		}
	}

	return result
}

// AddTBFQdisc adds a TBF qdisc
func (ag *TrafficControlAggregate) AddTBFQdisc(handle tc.Handle, rate tc.Bandwidth, buffer, limit, burst uint32) error {
	// Business rule: Check if qdisc already exists
	if _, exists := ag.qdiscs[handle]; exists {
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
		ag.id,
		ag.version+1,
		ag.deviceName,
		handle,
		rate,
		buffer,
		limit,
		burst,
	)

	ag.ApplyEvent(event)
	ag.changes = append(ag.changes, event)
	ag.version++

	return nil
}

// AddPRIOQdisc adds a PRIO qdisc
func (ag *TrafficControlAggregate) AddPRIOQdisc(handle tc.Handle, bands uint8, priomap []uint8) error {
	// Business rule: Check if qdisc already exists
	if _, exists := ag.qdiscs[handle]; exists {
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
		ag.id,
		ag.version+1,
		ag.deviceName,
		handle,
		bands,
		priomap,
	)

	ag.ApplyEvent(event)
	ag.changes = append(ag.changes, event)
	ag.version++

	return nil
}

// AddFQCODELQdisc adds a FQ_CODEL qdisc
func (ag *TrafficControlAggregate) AddFQCODELQdisc(handle tc.Handle, limit, flows, target, interval, quantum uint32, ecn bool) error {
	// Business rule: Check if qdisc already exists
	if _, exists := ag.qdiscs[handle]; exists {
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
		ag.id,
		ag.version+1,
		ag.deviceName,
		handle,
		limit,
		flows,
		target,
		interval,
		quantum,
		ecn,
	)

	ag.ApplyEvent(event)
	ag.changes = append(ag.changes, event)
	ag.version++

	return nil
}

// AddHTBClass adds an HTB class
func (ag *TrafficControlAggregate) AddHTBClass(parent tc.Handle, classHandle tc.Handle, name string, rate tc.Bandwidth, ceil tc.Bandwidth) error {
	// Business rule: Parent qdisc must exist
	parentQdisc, parentExists := ag.qdiscs[parent]
	if !parentExists {
		// Check if parent is a class
		if _, classExists := ag.classes[parent]; !classExists {
			return fmt.Errorf("parent %s does not exist", parent)
		}
	}

	// Business rule: Class handle must not already exist
	if _, exists := ag.classes[classHandle]; exists {
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
		ag.id,
		ag.version+1,
		ag.deviceName,
		classHandle,
		parent,
		name,
		rate,
		ceil,
	)

	ag.ApplyEvent(event)
	ag.changes = append(ag.changes, event)
	ag.version++

	return nil
}

// AddHTBClassWithAdvancedParameters adds an HTB class with enhanced parameters
func (ag *TrafficControlAggregate) AddHTBClassWithAdvancedParameters(
	parent tc.Handle, 
	classHandle tc.Handle, 
	name string, 
	rate tc.Bandwidth, 
	ceil tc.Bandwidth,
	priority entities.Priority,
	quantum uint32,
	overhead uint32,
	mpu uint32,
	mtu uint32,
	htbPrio uint32,
	useDefaults bool,
) error {
	// Business rule: Parent qdisc must exist
	parentQdisc, parentExists := ag.qdiscs[parent]
	if !parentExists {
		// Check if parent is a class
		if _, classExists := ag.classes[parent]; !classExists {
			return fmt.Errorf("parent %s does not exist", parent)
		}
	}

	// Business rule: Class handle must not already exist
	if _, exists := ag.classes[classHandle]; exists {
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

	// Create and apply event with enhanced parameters
	event := events.NewHTBClassCreatedEventWithAdvancedParameters(
		ag.id,
		ag.version+1,
		ag.deviceName,
		classHandle,
		parent,
		name,
		rate,
		ceil,
		priority,
		quantum,
		overhead,
		mpu,
		mtu,
		htbPrio,
		useDefaults,
	)

	ag.ApplyEvent(event)
	ag.changes = append(ag.changes, event)
	ag.version++

	return nil
}

// WithHTBClass returns a new aggregate with an HTB class added (immutable)
func (ag *TrafficControlAggregate) WithHTBClass(parent tc.Handle, classHandle tc.Handle, name string, rate tc.Bandwidth, ceil tc.Bandwidth) types.Result[*TrafficControlAggregate] {
	// Business rule: Parent qdisc must exist
	parentQdisc, parentExists := ag.qdiscs[parent]
	if !parentExists {
		// Check if parent is a class
		if _, classExists := ag.classes[parent]; !classExists {
			return types.Failure[*TrafficControlAggregate](fmt.Errorf("parent %s does not exist", parent))
		}
	}

	// Business rule: Class handle must not already exist
	if _, exists := ag.classes[classHandle]; exists {
		return types.Failure[*TrafficControlAggregate](fmt.Errorf("class with handle %s already exists", classHandle))
	}

	// Business rule: HTB specific - parent must be HTB
	if parentExists && parentQdisc.Type() != entities.QdiscTypeHTB {
		return types.Failure[*TrafficControlAggregate](fmt.Errorf("parent qdisc must be HTB type"))
	}

	// Business rule: Ceil must be >= Rate
	if ceil.BitsPerSecond() > 0 && ceil.LessThan(rate) {
		return types.Failure[*TrafficControlAggregate](fmt.Errorf("ceil (%s) cannot be less than rate (%s)", ceil, rate))
	}

	// Create event
	event := events.NewHTBClassCreatedEvent(
		ag.id,
		ag.version+1,
		ag.deviceName,
		classHandle,
		parent,
		name,
		rate,
		ceil,
	)

	// Create new aggregate with the event applied
	return types.Success(ag.withEvent(event))
}

// AddFilter adds a filter
func (ag *TrafficControlAggregate) AddFilter(parent tc.Handle, priority uint16, handle tc.Handle, flowID tc.Handle, matches []entities.Match) error {
	// Business rule: Parent must exist (either qdisc or class)
	_, qdiscExists := ag.qdiscs[parent]
	_, classExists := ag.classes[parent]
	if !qdiscExists && !classExists {
		return fmt.Errorf("parent %s does not exist", parent)
	}

	// Business rule: Target class (flowID) must exist
	if _, exists := ag.classes[flowID]; !exists {
		return fmt.Errorf("target class %s does not exist", flowID)
	}

	// Create event
	event := events.NewFilterCreatedEvent(
		ag.id,
		ag.version+1,
		ag.deviceName,
		parent,
		priority,
		handle,
		flowID,
	)

	// Add matches to event
	for _, match := range matches {
		event.AddMatch(match.Type(), match.String())
	}

	ag.ApplyEvent(event)
	ag.changes = append(ag.changes, event)
	ag.version++

	return nil
}

// GetUncommittedEvents returns uncommitted events
func (ag *TrafficControlAggregate) GetUncommittedEvents() []events.DomainEvent {
	return ag.changes
}

// MarkEventsAsCommitted clears uncommitted events
func (ag *TrafficControlAggregate) MarkEventsAsCommitted() {
	ag.changes = make([]events.DomainEvent, 0)
}

// LoadFromHistory rebuilds aggregate state from events
func (ag *TrafficControlAggregate) LoadFromHistory(history []events.DomainEvent) {
	for _, event := range history {
		ag.ApplyEvent(event)
		ag.version++
	}
	// Clear changes as these are already committed
	ag.changes = make([]events.DomainEvent, 0)
}

// GetVersion returns the current version
func (ag *TrafficControlAggregate) GetVersion() int {
	return ag.version
}

// ApplyEvent applies a domain event to update aggregate state
func (ag *TrafficControlAggregate) ApplyEvent(event events.DomainEvent) {
	switch e := event.(type) {
	case *events.HTBQdiscCreatedEvent:
		qdisc := entities.NewHTBQdisc(e.DeviceName, e.Handle, e.DefaultClass)
		ag.qdiscs[e.Handle] = qdisc.Qdisc

	case *events.TBFQdiscCreatedEvent:
		qdisc := entities.NewTBFQdisc(e.DeviceName, e.Handle, e.Rate)
		qdisc.SetBuffer(e.Buffer)
		qdisc.SetLimit(e.Limit)
		qdisc.SetBurst(e.Burst)
		ag.qdiscs[e.Handle] = qdisc.Qdisc

	case *events.PRIOQdiscCreatedEvent:
		qdisc := entities.NewPRIOQdisc(e.DeviceName, e.Handle, e.Bands)
		qdisc.SetPriomap(e.Priomap)
		ag.qdiscs[e.Handle] = qdisc.Qdisc

	case *events.FQCODELQdiscCreatedEvent:
		qdisc := entities.NewFQCODELQdisc(e.DeviceName, e.Handle)
		qdisc.SetLimit(e.Limit)
		qdisc.SetFlows(e.Flows)
		qdisc.SetTarget(e.Target)
		qdisc.SetInterval(e.Interval)
		qdisc.SetQuantum(e.Quantum)
		qdisc.SetECN(e.ECN)
		ag.qdiscs[e.Handle] = qdisc.Qdisc

	case *events.HTBClassCreatedEvent:
		// Use a default priority of 4 for event reconstruction
		class := entities.NewHTBClass(e.DeviceName, e.Handle, e.Parent, e.Name, entities.Priority(4))
		class.SetRate(e.Rate)
		class.SetCeil(e.Ceil)
		ag.classes[e.Handle] = class.Class

	case *events.FilterCreatedEvent:
		filter := entities.NewFilter(e.DeviceName, e.Parent, e.Priority, e.Handle)
		filter.SetFlowID(e.FlowID)
		filter.SetProtocol(e.Protocol)

		// Reconstruct matches from event data
		for _, matchData := range e.Matches {
			switch matchData.Type {
			case entities.MatchTypeIPDestination:
				// Parse IP from string representation
				// Format: "ip dst 192.168.1.100/32"
				var cidr string
				if _, err := fmt.Sscanf(matchData.Value, "ip dst %s", &cidr); err == nil {
					if ipMatch, err := entities.NewIPDestinationMatch(cidr); err == nil {
						filter.AddMatch(ipMatch)
					}
				}
			case entities.MatchTypeIPSource:
				// Parse IP from string representation
				// Format: "ip src 10.0.0.0/24"
				var cidr string
				if _, err := fmt.Sscanf(matchData.Value, "ip src %s", &cidr); err == nil {
					if ipMatch, err := entities.NewIPSourceMatch(cidr); err == nil {
						filter.AddMatch(ipMatch)
					}
				}
			case entities.MatchTypePortDestination:
				// Parse port from string representation
				// Format: "ip dport 5201 0xffff"
				var port uint16
				if _, err := fmt.Sscanf(matchData.Value, "ip dport %d 0xffff", &port); err == nil {
					match := entities.NewPortDestinationMatch(port)
					filter.AddMatch(match)
				}
			case entities.MatchTypePortSource:
				// Parse port from string representation
				// Format: "ip sport 8080 0xffff"
				var port uint16
				if _, err := fmt.Sscanf(matchData.Value, "ip sport %d 0xffff", &port); err == nil {
					match := entities.NewPortSourceMatch(port)
					filter.AddMatch(match)
				}
			}
		}

		ag.filters = append(ag.filters, filter)

	case *events.QdiscDeletedEvent:
		delete(ag.qdiscs, e.Handle)

	case *events.ClassDeletedEvent:
		delete(ag.classes, e.Handle)
	}
}

// GetUncommittedChanges returns events that haven't been persisted
func (ag *TrafficControlAggregate) GetUncommittedChanges() []events.DomainEvent {
	return ag.changes
}

// MarkChangesAsCommitted clears the uncommitted changes
func (ag *TrafficControlAggregate) MarkChangesAsCommitted() {
	ag.changes = make([]events.DomainEvent, 0)
}

// GetQdiscs returns all qdiscs (for queries)
func (ag *TrafficControlAggregate) GetQdiscs() map[tc.Handle]*entities.Qdisc {
	// Return a copy to maintain immutability
	result := make(map[tc.Handle]*entities.Qdisc)
	for k, v := range ag.qdiscs {
		result[k] = v
	}
	return result
}

// GetClasses returns all classes (for queries)
func (ag *TrafficControlAggregate) GetClasses() map[tc.Handle]*entities.Class {
	// Return a copy to maintain immutability
	result := make(map[tc.Handle]*entities.Class)
	for k, v := range ag.classes {
		result[k] = v
	}
	return result
}

// GetFilters returns all filters (for queries)
func (ag *TrafficControlAggregate) GetFilters() []*entities.Filter {
	// Return a copy to maintain immutability
	result := make([]*entities.Filter, len(ag.filters))
	copy(result, ag.filters)
	return result
}
