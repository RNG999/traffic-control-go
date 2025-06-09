package aggregates

import (
	"testing"

	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/events"
	"github.com/rng999/traffic-control-go/pkg/tc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddFilter(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*TrafficControlAggregate)
		parent    string
		priority  uint16
		handle    string
		flowID    string
		matches   []entities.Match
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *TrafficControlAggregate)
	}{
		{
			name: "successful filter creation with IP matches",
			setup: func(agg *TrafficControlAggregate) {
				// Add HTB qdisc
				qHandle, _ := tc.ParseHandle("1:")
				defaultHandle, _ := tc.ParseHandle("1:10")
				agg.AddHTBQdisc(qHandle, defaultHandle)

				// Add target class
				parentHandle, _ := tc.ParseHandle("1:")
				classHandle, _ := tc.ParseHandle("1:10")
				rate, _ := tc.ParseBandwidth("10Mbps")
				ceil, _ := tc.ParseBandwidth("20Mbps")
				agg.AddHTBClass(parentHandle, classHandle, "default", rate, ceil)
			},
			parent:   "1:",
			priority: 100,
			handle:   "800:100",
			flowID:   "1:10",
			matches: []entities.Match{
				mustNewIPSourceMatch("192.168.1.0/24"),
				mustNewIPDestinationMatch("10.0.0.0/8"),
			},
			wantErr: false,
			validate: func(t *testing.T, agg *TrafficControlAggregate) {
				filters := agg.GetFilters()
				assert.Len(t, filters, 1)

				filter := filters[0]
				assert.Equal(t, uint16(100), filter.Priority())
				assert.Equal(t, "1:10", filter.FlowID().String())
				assert.Len(t, filter.Matches(), 2)

				// Check filter matches were stored correctly
				assert.Equal(t, entities.MatchTypeIPSource, filter.Matches()[0].Type())
				assert.Equal(t, entities.MatchTypeIPDestination, filter.Matches()[1].Type())
			},
		},
		{
			name: "successful filter with port matches",
			setup: func(agg *TrafficControlAggregate) {
				// Add HTB qdisc and class
				qHandle, _ := tc.ParseHandle("1:")
				defaultHandle, _ := tc.ParseHandle("1:10")
				agg.AddHTBQdisc(qHandle, defaultHandle)

				parentHandle, _ := tc.ParseHandle("1:")
				classHandle, _ := tc.ParseHandle("1:10")
				rate, _ := tc.ParseBandwidth("10Mbps")
				ceil, _ := tc.ParseBandwidth("20Mbps")
				agg.AddHTBClass(parentHandle, classHandle, "default", rate, ceil)
			},
			parent:   "1:",
			priority: 200,
			handle:   "800:200",
			flowID:   "1:10",
			matches: []entities.Match{
				entities.NewPortSourceMatch(80),
				entities.NewPortDestinationMatch(443),
			},
			wantErr: false,
			validate: func(t *testing.T, agg *TrafficControlAggregate) {
				filters := agg.GetFilters()
				assert.Len(t, filters, 1)

				filter := filters[0]
				assert.Len(t, filter.Matches(), 2)
				assert.Equal(t, entities.MatchTypePortSource, filter.Matches()[0].Type())
				assert.Equal(t, entities.MatchTypePortDestination, filter.Matches()[1].Type())
			},
		},
		{
			name: "successful filter with protocol match",
			setup: func(agg *TrafficControlAggregate) {
				// Add HTB qdisc and class
				qHandle, _ := tc.ParseHandle("1:")
				defaultHandle, _ := tc.ParseHandle("1:10")
				agg.AddHTBQdisc(qHandle, defaultHandle)

				parentHandle, _ := tc.ParseHandle("1:")
				classHandle, _ := tc.ParseHandle("1:10")
				rate, _ := tc.ParseBandwidth("10Mbps")
				ceil, _ := tc.ParseBandwidth("20Mbps")
				agg.AddHTBClass(parentHandle, classHandle, "default", rate, ceil)
			},
			parent:   "1:",
			priority: 300,
			handle:   "800:300",
			flowID:   "1:10",
			matches: []entities.Match{
				entities.NewProtocolMatch(entities.TransportProtocolTCP),
			},
			wantErr: false,
			validate: func(t *testing.T, agg *TrafficControlAggregate) {
				filters := agg.GetFilters()
				assert.Len(t, filters, 1)

				filter := filters[0]
				assert.Len(t, filter.Matches(), 1)
				assert.Equal(t, entities.MatchTypeProtocol, filter.Matches()[0].Type())
			},
		},
		{
			name:     "fail when parent does not exist",
			setup:    func(agg *TrafficControlAggregate) {},
			parent:   "1:",
			priority: 100,
			handle:   "800:100",
			flowID:   "1:10",
			matches:  []entities.Match{},
			wantErr:  true,
			errMsg:   "parent 1: does not exist",
		},
		{
			name: "fail when target class does not exist",
			setup: func(agg *TrafficControlAggregate) {
				// Add HTB qdisc but no target class
				qHandle, _ := tc.ParseHandle("1:")
				defaultHandle, _ := tc.ParseHandle("1:10")
				agg.AddHTBQdisc(qHandle, defaultHandle)
			},
			parent:   "1:",
			priority: 100,
			handle:   "800:100",
			flowID:   "1:20", // non-existent class
			matches:  []entities.Match{},
			wantErr:  true,
			errMsg:   "target class 1:20 does not exist",
		},
		{
			name: "parent can be a class",
			setup: func(agg *TrafficControlAggregate) {
				// Add HTB qdisc
				qHandle, _ := tc.ParseHandle("1:")
				defaultHandle, _ := tc.ParseHandle("1:10")
				agg.AddHTBQdisc(qHandle, defaultHandle)

				// Add parent class
				parentHandle, _ := tc.ParseHandle("1:")
				classHandle, _ := tc.ParseHandle("1:10")
				rate, _ := tc.ParseBandwidth("10Mbps")
				ceil, _ := tc.ParseBandwidth("20Mbps")
				agg.AddHTBClass(parentHandle, classHandle, "parent", rate, ceil)

				// Add target class
				targetHandle, _ := tc.ParseHandle("1:20")
				agg.AddHTBClass(parentHandle, targetHandle, "target", rate, ceil)
			},
			parent:   "1:10", // parent is a class
			priority: 100,
			handle:   "800:100",
			flowID:   "1:20",
			matches: []entities.Match{
				mustNewIPSourceMatch("192.168.1.0/24"),
			},
			wantErr: false,
			validate: func(t *testing.T, agg *TrafficControlAggregate) {
				filters := agg.GetFilters()
				assert.Len(t, filters, 1)
				assert.Equal(t, "1:10", filters[0].Parent().String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create aggregate
			device, _ := tc.NewDeviceName("eth0")
			agg := NewTrafficControlAggregate(device)

			// Setup
			if tt.setup != nil {
				tt.setup(agg)
			}

			// Clear uncommitted changes from setup
			agg.MarkEventsAsCommitted()

			// Parse handles
			parentHandle, _ := tc.ParseHandle(tt.parent)
			handle, _ := tc.ParseHandle(tt.handle)
			flowID, _ := tc.ParseHandle(tt.flowID)

			// Execute
			err := agg.AddFilter(parentHandle, tt.priority, handle, flowID, tt.matches)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)

				// Check event was created
				changes := agg.GetUncommittedEvents()
				assert.Len(t, changes, 1)

				event, ok := changes[0].(*events.FilterCreatedEvent)
				require.True(t, ok)
				assert.Equal(t, parentHandle, event.Parent)
				assert.Equal(t, tt.priority, event.Priority)
				assert.Equal(t, handle, event.Handle)
				assert.Equal(t, flowID, event.FlowID)
				assert.Len(t, event.Matches, len(tt.matches))

				// Validate aggregate state
				if tt.validate != nil {
					tt.validate(t, agg)
				}
			}
		})
	}
}

func TestDeleteFilter(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*TrafficControlAggregate)
		parent    string
		priority  uint16
		handle    string
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *TrafficControlAggregate)
	}{
		{
			name: "successful filter deletion",
			setup: func(agg *TrafficControlAggregate) {
				// Add HTB qdisc
				qHandle, _ := tc.ParseHandle("1:")
				defaultHandle, _ := tc.ParseHandle("1:10")
				agg.AddHTBQdisc(qHandle, defaultHandle)

				// Add class
				parentHandle, _ := tc.ParseHandle("1:")
				classHandle, _ := tc.ParseHandle("1:10")
				rate, _ := tc.ParseBandwidth("10Mbps")
				ceil, _ := tc.ParseBandwidth("20Mbps")
				agg.AddHTBClass(parentHandle, classHandle, "default", rate, ceil)

				// Add filter
				filterHandle, _ := tc.ParseHandle("800:100")
				flowID, _ := tc.ParseHandle("1:10")
				agg.AddFilter(parentHandle, 100, filterHandle, flowID, nil)
			},
			parent:   "1:",
			priority: 100,
			handle:   "800:100",
			wantErr:  false,
			validate: func(t *testing.T, agg *TrafficControlAggregate) {
				filters := agg.GetFilters()
				assert.Len(t, filters, 0)
			},
		},
		{
			name: "fail when filter does not exist",
			setup: func(agg *TrafficControlAggregate) {
				// Add HTB qdisc but no filter
				qHandle, _ := tc.ParseHandle("1:")
				defaultHandle, _ := tc.ParseHandle("1:10")
				agg.AddHTBQdisc(qHandle, defaultHandle)
			},
			parent:   "1:",
			priority: 100,
			handle:   "800:100",
			wantErr:  true,
			errMsg:   "filter with parent 1:, priority 100, handle 800:100 not found",
		},
		{
			name: "delete only specific filter",
			setup: func(agg *TrafficControlAggregate) {
				// Add HTB qdisc
				qHandle, _ := tc.ParseHandle("1:")
				defaultHandle, _ := tc.ParseHandle("1:10")
				agg.AddHTBQdisc(qHandle, defaultHandle)

				// Add class
				parentHandle, _ := tc.ParseHandle("1:")
				classHandle, _ := tc.ParseHandle("1:10")
				rate, _ := tc.ParseBandwidth("10Mbps")
				ceil, _ := tc.ParseBandwidth("20Mbps")
				agg.AddHTBClass(parentHandle, classHandle, "default", rate, ceil)

				// Add multiple filters
				flowID, _ := tc.ParseHandle("1:10")
				filterHandle1, _ := tc.ParseHandle("800:100")
				filterHandle2, _ := tc.ParseHandle("800:200")
				agg.AddFilter(parentHandle, 100, filterHandle1, flowID, nil)
				agg.AddFilter(parentHandle, 200, filterHandle2, flowID, nil)
			},
			parent:   "1:",
			priority: 100,
			handle:   "800:100",
			wantErr:  false,
			validate: func(t *testing.T, agg *TrafficControlAggregate) {
				filters := agg.GetFilters()
				assert.Len(t, filters, 1)
				assert.Equal(t, uint16(200), filters[0].Priority())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create aggregate
			device, _ := tc.NewDeviceName("eth0")
			agg := NewTrafficControlAggregate(device)

			// Setup
			if tt.setup != nil {
				tt.setup(agg)
			}

			// Clear uncommitted changes from setup
			agg.MarkEventsAsCommitted()

			// Parse handles
			parentHandle, _ := tc.ParseHandle(tt.parent)
			handle, _ := tc.ParseHandle(tt.handle)

			// Execute
			err := agg.DeleteFilter(parentHandle, tt.priority, handle)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)

				// Check event was created
				changes := agg.GetUncommittedEvents()
				assert.Len(t, changes, 1)

				event, ok := changes[0].(*events.FilterDeletedEvent)
				require.True(t, ok)
				assert.Equal(t, parentHandle, event.Parent)
				assert.Equal(t, tt.priority, event.Priority)
				assert.Equal(t, handle, event.Handle)

				// Validate aggregate state
				if tt.validate != nil {
					tt.validate(t, agg)
				}
			}
		})
	}
}

func TestFilterEventReplay(t *testing.T) {
	// Create initial aggregate
	device, _ := tc.NewDeviceName("eth0")
	agg1 := NewTrafficControlAggregate(device)

	// Add HTB qdisc
	qHandle, _ := tc.ParseHandle("1:")
	defaultHandle, _ := tc.ParseHandle("1:10")
	agg1.AddHTBQdisc(qHandle, defaultHandle)

	// Add classes
	parentHandle, _ := tc.ParseHandle("1:")
	classHandle1, _ := tc.ParseHandle("1:10")
	classHandle2, _ := tc.ParseHandle("1:20")
	rate, _ := tc.ParseBandwidth("10Mbps")
	ceil, _ := tc.ParseBandwidth("20Mbps")
	agg1.AddHTBClass(parentHandle, classHandle1, "default", rate, ceil)
	agg1.AddHTBClass(parentHandle, classHandle2, "priority", rate, ceil)

	// Add filters
	filterHandle1, _ := tc.ParseHandle("800:100")
	filterHandle2, _ := tc.ParseHandle("800:200")
	flowID1, _ := tc.ParseHandle("1:10")
	flowID2, _ := tc.ParseHandle("1:20")

	matches1 := []entities.Match{
		mustNewIPSourceMatch("192.168.1.0/24"),
		entities.NewPortDestinationMatch(80),
	}
	matches2 := []entities.Match{
		mustNewIPDestinationMatch("10.0.0.0/8"),
		entities.NewProtocolMatch(entities.TransportProtocolTCP),
	}

	agg1.AddFilter(parentHandle, 100, filterHandle1, flowID1, matches1)
	agg1.AddFilter(parentHandle, 200, filterHandle2, flowID2, matches2)

	// Delete one filter
	agg1.DeleteFilter(parentHandle, 100, filterHandle1)

	// Get all events
	events := agg1.GetUncommittedEvents()

	// Create new aggregate and replay events
	agg2 := NewTrafficControlAggregate(device)
	agg2.LoadFromHistory(events)

	// Verify state matches
	filters := agg2.GetFilters()
	assert.Len(t, filters, 1)
	assert.Equal(t, uint16(200), filters[0].Priority())
	assert.Equal(t, flowID2, filters[0].FlowID())
	assert.Len(t, filters[0].Matches(), 2)

	// Verify version
	assert.Equal(t, agg1.GetVersion(), agg2.GetVersion())
}

// Helper functions
func mustNewIPSourceMatch(cidr string) *entities.IPMatch {
	match, _ := entities.NewIPSourceMatch(cidr)
	return match
}

func mustNewIPDestinationMatch(cidr string) *entities.IPMatch {
	match, _ := entities.NewIPDestinationMatch(cidr)
	return match
}