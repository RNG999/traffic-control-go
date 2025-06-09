package handlers

import (
	"context"
	"testing"

	"github.com/rng999/traffic-control-go/internal/commands/models"
	"github.com/rng999/traffic-control-go/internal/domain/aggregates"
	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/pkg/tc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateFilterHandler(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*aggregates.TrafficControlAggregate)
		command   *models.CreateFilterCommand
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *aggregates.TrafficControlAggregate)
	}{
		{
			name: "successful filter creation with IP match",
			setup: func(agg *aggregates.TrafficControlAggregate) {
				// Add HTB qdisc
				qHandle, _ := tc.ParseHandle("1:")
				defaultHandle, _ := tc.ParseHandle("1:10")
				require.NoError(t, agg.AddHTBQdisc(qHandle, defaultHandle))

				// Add parent class
				parentHandle, _ := tc.ParseHandle("1:")
				classHandle, _ := tc.ParseHandle("1:10")
				rate, _ := tc.ParseBandwidth("10Mbps")
				ceil, _ := tc.ParseBandwidth("20Mbps")
				require.NoError(t, agg.AddHTBClass(parentHandle, classHandle, "default", rate, ceil))

				// Add target class
				targetHandle, _ := tc.ParseHandle("1:20")
				require.NoError(t, agg.AddHTBClass(parentHandle, targetHandle, "priority", rate, ceil))
			},
			command: &models.CreateFilterCommand{
				DeviceName: "eth0",
				Parent:     "1:",
				Priority:   100,
				Protocol:   "ip",
				FlowID:     "1:20",
				Match: map[string]string{
					"src_ip": "192.168.1.0/24",
					"dst_ip": "10.0.0.0/8",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, agg *aggregates.TrafficControlAggregate) {
				filters := agg.GetFilters()
				require.Len(t, filters, 1)

				filter := filters[0]
				assert.Equal(t, uint16(100), filter.Priority())
				assert.Equal(t, "1:20", filter.FlowID().String())
				assert.Len(t, filter.Matches(), 2)
			},
		},
		{
			name: "successful filter creation with port match",
			setup: func(agg *aggregates.TrafficControlAggregate) {
				// Add HTB qdisc and classes
				qHandle, _ := tc.ParseHandle("1:")
				defaultHandle, _ := tc.ParseHandle("1:10")
				require.NoError(t, agg.AddHTBQdisc(qHandle, defaultHandle))

				parentHandle, _ := tc.ParseHandle("1:")
				classHandle, _ := tc.ParseHandle("1:10")
				rate, _ := tc.ParseBandwidth("10Mbps")
				ceil, _ := tc.ParseBandwidth("20Mbps")
				require.NoError(t, agg.AddHTBClass(parentHandle, classHandle, "default", rate, ceil))
			},
			command: &models.CreateFilterCommand{
				DeviceName: "eth0",
				Parent:     "1:",
				Priority:   200,
				Protocol:   "ip",
				FlowID:     "1:10",
				Match: map[string]string{
					"src_port": "80",
					"dst_port": "443",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, agg *aggregates.TrafficControlAggregate) {
				filters := agg.GetFilters()
				require.Len(t, filters, 1)

				filter := filters[0]
				assert.Equal(t, uint16(200), filter.Priority())
				assert.Len(t, filter.Matches(), 2)
			},
		},
		{
			name: "fail when parent does not exist",
			setup: func(agg *aggregates.TrafficControlAggregate) {
				// No setup - parent doesn't exist
			},
			command: &models.CreateFilterCommand{
				DeviceName: "eth0",
				Parent:     "1:",
				Priority:   100,
				Protocol:   "ip",
				FlowID:     "1:20",
				Match: map[string]string{
					"src_ip": "192.168.1.0/24",
				},
			},
			wantErr: true,
			errMsg:  "parent 1: does not exist",
		},
		{
			name: "fail when target class does not exist",
			setup: func(agg *aggregates.TrafficControlAggregate) {
				// Add HTB qdisc but no target class
				qHandle, _ := tc.ParseHandle("1:")
				defaultHandle, _ := tc.ParseHandle("1:10")
				require.NoError(t, agg.AddHTBQdisc(qHandle, defaultHandle))
			},
			command: &models.CreateFilterCommand{
				DeviceName: "eth0",
				Parent:     "1:",
				Priority:   100,
				Protocol:   "ip",
				FlowID:     "1:20",
				Match: map[string]string{
					"src_ip": "192.168.1.0/24",
				},
			},
			wantErr: true,
			errMsg:  "target class 1:20 does not exist",
		},
		{
			name: "parent can be a class",
			setup: func(agg *aggregates.TrafficControlAggregate) {
				// Add HTB qdisc
				qHandle, _ := tc.ParseHandle("1:")
				defaultHandle, _ := tc.ParseHandle("1:10")
				require.NoError(t, agg.AddHTBQdisc(qHandle, defaultHandle))

				// Add parent class
				parentHandle, _ := tc.ParseHandle("1:")
				classHandle, _ := tc.ParseHandle("1:10")
				rate, _ := tc.ParseBandwidth("10Mbps")
				ceil, _ := tc.ParseBandwidth("20Mbps")
				require.NoError(t, agg.AddHTBClass(parentHandle, classHandle, "parent", rate, ceil))

				// Add target class
				targetHandle, _ := tc.ParseHandle("1:20")
				require.NoError(t, agg.AddHTBClass(parentHandle, targetHandle, "target", rate, ceil))
			},
			command: &models.CreateFilterCommand{
				DeviceName: "eth0",
				Parent:     "1:10", // Parent is a class
				Priority:   100,
				Protocol:   "ip",
				FlowID:     "1:20",
				Match: map[string]string{
					"src_ip": "192.168.1.0/24",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, agg *aggregates.TrafficControlAggregate) {
				filters := agg.GetFilters()
				require.Len(t, filters, 1)
				assert.Equal(t, "1:10", filters[0].Parent().String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create event store and handler
			store := eventstore.NewMemoryEventStoreWithContext()
			handler := NewCreateFilterHandler(store)

			// Create device
			device, err := tc.NewDeviceName(tt.command.DeviceName)
			require.NoError(t, err)

			// Setup aggregate
			agg := aggregates.NewTrafficControlAggregate(device)
			if tt.setup != nil {
				tt.setup(agg)
			}

			// Save initial state
			err = store.SaveAggregate(context.Background(), agg)
			require.NoError(t, err)

			// Execute command
			err = handler.HandleTyped(context.Background(), tt.command)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)

				// Load aggregate and validate
				loadedAgg := aggregates.NewTrafficControlAggregate(device)
				err = store.Load(context.Background(), loadedAgg.GetID(), loadedAgg)
				require.NoError(t, err)

				if tt.validate != nil {
					tt.validate(t, loadedAgg)
				}
			}
		})
	}
}

func TestDeleteFilterHandler(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*aggregates.TrafficControlAggregate)
		command   *models.DeleteFilterCommand
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *aggregates.TrafficControlAggregate)
	}{
		{
			name: "successful filter deletion",
			setup: func(agg *aggregates.TrafficControlAggregate) {
				// Add HTB qdisc
				qHandle, _ := tc.ParseHandle("1:")
				defaultHandle, _ := tc.ParseHandle("1:10")
				require.NoError(t, agg.AddHTBQdisc(qHandle, defaultHandle))

				// Add classes
				parentHandle, _ := tc.ParseHandle("1:")
				classHandle, _ := tc.ParseHandle("1:10")
				rate, _ := tc.ParseBandwidth("10Mbps")
				ceil, _ := tc.ParseBandwidth("20Mbps")
				require.NoError(t, agg.AddHTBClass(parentHandle, classHandle, "default", rate, ceil))

				// Add filter
				filterHandle := tc.NewHandle(0x800, 100)
				flowID, _ := tc.ParseHandle("1:10")
				require.NoError(t, agg.AddFilter(parentHandle, 100, filterHandle, flowID, nil))
			},
			command: &models.DeleteFilterCommand{
				DeviceName: mustParseDeviceName("eth0"),
				Parent:     mustParseHandle("1:"),
				Priority:   100,
				Handle:     tc.NewHandle(0x800, 100),
			},
			wantErr: false,
			validate: func(t *testing.T, agg *aggregates.TrafficControlAggregate) {
				filters := agg.GetFilters()
				assert.Len(t, filters, 0)
			},
		},
		{
			name: "fail when filter does not exist",
			setup: func(agg *aggregates.TrafficControlAggregate) {
				// Add HTB qdisc but no filter
				qHandle, _ := tc.ParseHandle("1:")
				defaultHandle, _ := tc.ParseHandle("1:10")
				require.NoError(t, agg.AddHTBQdisc(qHandle, defaultHandle))
			},
			command: &models.DeleteFilterCommand{
				DeviceName: mustParseDeviceName("eth0"),
				Parent:     mustParseHandle("1:"),
				Priority:   100,
				Handle:     tc.NewHandle(0x800, 100),
			},
			wantErr: true,
			errMsg:  "filter with parent 1:, priority 100, handle 800:64 not found",
		},
		{
			name: "only delete specific filter",
			setup: func(agg *aggregates.TrafficControlAggregate) {
				// Add HTB qdisc
				qHandle, _ := tc.ParseHandle("1:")
				defaultHandle, _ := tc.ParseHandle("1:10")
				require.NoError(t, agg.AddHTBQdisc(qHandle, defaultHandle))

				// Add classes
				parentHandle, _ := tc.ParseHandle("1:")
				classHandle, _ := tc.ParseHandle("1:10")
				rate, _ := tc.ParseBandwidth("10Mbps")
				ceil, _ := tc.ParseBandwidth("20Mbps")
				require.NoError(t, agg.AddHTBClass(parentHandle, classHandle, "default", rate, ceil))

				// Add multiple filters
				flowID, _ := tc.ParseHandle("1:10")
				require.NoError(t, agg.AddFilter(parentHandle, 100, tc.NewHandle(0x800, 100), flowID, nil))
				require.NoError(t, agg.AddFilter(parentHandle, 200, tc.NewHandle(0x800, 200), flowID, nil))
			},
			command: &models.DeleteFilterCommand{
				DeviceName: mustParseDeviceName("eth0"),
				Parent:     mustParseHandle("1:"),
				Priority:   100,
				Handle:     tc.NewHandle(0x800, 100),
			},
			wantErr: false,
			validate: func(t *testing.T, agg *aggregates.TrafficControlAggregate) {
				filters := agg.GetFilters()
				assert.Len(t, filters, 1)
				assert.Equal(t, uint16(200), filters[0].Priority())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create event store and handler
			store := eventstore.NewMemoryEventStoreWithContext()
			handler := NewDeleteFilterHandler(store)

			// Setup aggregate
			agg := aggregates.NewTrafficControlAggregate(tt.command.DeviceName)
			if tt.setup != nil {
				tt.setup(agg)
			}

			// Save initial state
			err := store.SaveAggregate(context.Background(), agg)
			require.NoError(t, err)

			// Execute command
			err = handler.HandleTyped(context.Background(), tt.command)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)

				// Load aggregate and validate
				loadedAgg := aggregates.NewTrafficControlAggregate(tt.command.DeviceName)
				err = store.Load(context.Background(), loadedAgg.GetID(), loadedAgg)
				require.NoError(t, err)

				if tt.validate != nil {
					tt.validate(t, loadedAgg)
				}
			}
		})
	}
}

func TestCreateAdvancedFilterHandler(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*aggregates.TrafficControlAggregate)
		command   *models.CreateAdvancedFilterCommand
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *aggregates.TrafficControlAggregate)
	}{
		{
			name: "successful advanced filter with IP and port ranges",
			setup: func(agg *aggregates.TrafficControlAggregate) {
				// Add HTB qdisc
				qHandle, _ := tc.ParseHandle("1:")
				defaultHandle, _ := tc.ParseHandle("1:10")
				require.NoError(t, agg.AddHTBQdisc(qHandle, defaultHandle))

				// Add classes
				parentHandle, _ := tc.ParseHandle("1:")
				classHandle, _ := tc.ParseHandle("1:10")
				rate, _ := tc.ParseBandwidth("10Mbps")
				ceil, _ := tc.ParseBandwidth("20Mbps")
				require.NoError(t, agg.AddHTBClass(parentHandle, classHandle, "default", rate, ceil))
			},
			command: &models.CreateAdvancedFilterCommand{
				DeviceName: "eth0",
				Parent:     "1:",
				Priority:   100,
				Handle:     "800:100",
				Protocol:   "ip",
				FlowID:     "1:10",
				IPSourceRange: &models.IPRange{
					CIDR: "192.168.1.0/24",
				},
				IPDestRange: &models.IPRange{
					CIDR: "10.0.0.0/8",
				},
				PortSourceRange: &models.PortRange{
					StartPort: 1024,
					EndPort:   65535,
				},
				PortDestRange: &models.PortRange{
					StartPort: 80,
					EndPort:   443,
				},
				TransportProtocol: "tcp",
			},
			wantErr: false,
			validate: func(t *testing.T, agg *aggregates.TrafficControlAggregate) {
				filters := agg.GetFilters()
				require.Len(t, filters, 1)

				filter := filters[0]
				assert.Equal(t, uint16(100), filter.Priority())
				assert.Equal(t, "1:10", filter.FlowID().String())
				assert.GreaterOrEqual(t, len(filter.Matches()), 4)
			},
		},
		{
			name: "successful filter with protocol match",
			setup: func(agg *aggregates.TrafficControlAggregate) {
				// Add HTB qdisc
				qHandle, _ := tc.ParseHandle("1:")
				defaultHandle, _ := tc.ParseHandle("1:10")
				require.NoError(t, agg.AddHTBQdisc(qHandle, defaultHandle))

				// Add classes
				parentHandle, _ := tc.ParseHandle("1:")
				classHandle, _ := tc.ParseHandle("1:10")
				rate, _ := tc.ParseBandwidth("10Mbps")
				ceil, _ := tc.ParseBandwidth("20Mbps")
				require.NoError(t, agg.AddHTBClass(parentHandle, classHandle, "default", rate, ceil))
			},
			command: &models.CreateAdvancedFilterCommand{
				DeviceName:        "eth0",
				Parent:            "1:",
				Priority:          200,
				Handle:            "800:200",
				Protocol:          "ip",
				FlowID:            "1:10",
				TransportProtocol: "udp",
			},
			wantErr: false,
			validate: func(t *testing.T, agg *aggregates.TrafficControlAggregate) {
				filters := agg.GetFilters()
				require.Len(t, filters, 1)

				// Check that protocol match was added
				hasProtocolMatch := false
				for _, match := range filters[0].Matches() {
					if match.Type() == entities.MatchTypeProtocol {
						hasProtocolMatch = true
						break
					}
				}
				assert.True(t, hasProtocolMatch)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create event store and handler
			store := eventstore.NewMemoryEventStoreWithContext()
			handler := NewCreateAdvancedFilterHandler(store)

			// Create device
			device, err := tc.NewDeviceName(tt.command.DeviceName)
			require.NoError(t, err)

			// Setup aggregate
			agg := aggregates.NewTrafficControlAggregate(device)
			if tt.setup != nil {
				tt.setup(agg)
			}

			// Save initial state
			err = store.SaveAggregate(context.Background(), agg)
			require.NoError(t, err)

			// Execute command
			err = handler.HandleTyped(context.Background(), tt.command)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)

				// Load aggregate and validate
				loadedAgg := aggregates.NewTrafficControlAggregate(device)
				err = store.Load(context.Background(), loadedAgg.GetID(), loadedAgg)
				require.NoError(t, err)

				if tt.validate != nil {
					tt.validate(t, loadedAgg)
				}
			}
		})
	}
}

// Helper functions
func mustParseDeviceName(name string) tc.DeviceName {
	device, _ := tc.NewDeviceName(name)
	return device
}

func mustParseHandle(handle string) tc.Handle {
	h, _ := tc.ParseHandle(handle)
	return h
}