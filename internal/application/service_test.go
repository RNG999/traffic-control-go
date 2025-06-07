package application

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	qmodels "github.com/rng999/traffic-control-go/internal/queries/models"
	"github.com/rng999/traffic-control-go/pkg/logging"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

func TestNewTrafficControlService(t *testing.T) {
	t.Run("creates_service_with_all_dependencies", func(t *testing.T) {
		eventStore := eventstore.NewMemoryEventStoreWithContext()
		netlinkAdapter := netlink.NewMockAdapter()
		logger := logging.WithComponent("application")

		service := NewTrafficControlService(eventStore, netlinkAdapter, logger)

		assert.NotNil(t, service)
		assert.Equal(t, eventStore, service.eventStore)
		assert.Equal(t, netlinkAdapter, service.netlinkAdapter)
		assert.Equal(t, logger, service.logger)
		assert.NotNil(t, service.commandBus)
		assert.NotNil(t, service.queryBus)
		assert.NotNil(t, service.eventBus)
		assert.NotNil(t, service.projectionManager)
		assert.NotNil(t, service.readModelStore)
		assert.NotNil(t, service.statisticsService)
	})

	t.Run("registers_handlers_and_projections", func(t *testing.T) {
		eventStore := eventstore.NewMemoryEventStoreWithContext()
		netlinkAdapter := netlink.NewMockAdapter()
		logger := logging.WithComponent("application")

		service := NewTrafficControlService(eventStore, netlinkAdapter, logger)

		// Verify handlers are registered by checking the buses have handlers
		assert.NotEmpty(t, service.commandBus.handlers)
		assert.NotEmpty(t, service.eventBus.handlers)
	})
}

func TestTrafficControlService_CreateHTBQdisc(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("application")
	service := NewTrafficControlService(eventStore, netlinkAdapter, logger)

	t.Run("creates_htb_qdisc_successfully", func(t *testing.T) {
		ctx := context.Background()

		err := service.CreateHTBQdisc(ctx, "eth0", "1:0", "1:999")

		assert.NoError(t, err)
	})

	t.Run("fails_with_invalid_device", func(t *testing.T) {
		ctx := context.Background()

		err := service.CreateHTBQdisc(ctx, "", "1:0", "1:999")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create HTB qdisc")
	})
}

func TestTrafficControlService_CreateTBFQdisc(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("application")
	service := NewTrafficControlService(eventStore, netlinkAdapter, logger)

	t.Run("creates_tbf_qdisc_successfully", func(t *testing.T) {
		ctx := context.Background()

		err := service.CreateTBFQdisc(ctx, "eth0", "1:0", "10mbps", 32768, 10000, 0)

		assert.NoError(t, err)
	})

	t.Run("fails_with_invalid_parameters", func(t *testing.T) {
		ctx := context.Background()

		err := service.CreateTBFQdisc(ctx, "", "1:0", "invalid", 0, 0, 0)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create TBF qdisc")
	})
}

func TestTrafficControlService_CreatePRIOQdisc(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("application")
	service := NewTrafficControlService(eventStore, netlinkAdapter, logger)

	t.Run("creates_prio_qdisc_successfully", func(t *testing.T) {
		ctx := context.Background()
		priomap := []uint8{0, 1, 2, 0, 1, 2, 0, 1, 0, 1, 2, 0, 1, 2, 0, 1}

		err := service.CreatePRIOQdisc(ctx, "eth0", "1:0", 3, priomap)

		assert.NoError(t, err)
	})

	t.Run("fails_with_invalid_device", func(t *testing.T) {
		ctx := context.Background()
		priomap := []uint8{0, 1, 2, 0, 1, 2, 0, 1, 0, 1, 2, 0, 1, 2, 0, 1}

		err := service.CreatePRIOQdisc(ctx, "", "1:0", 3, priomap)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create PRIO qdisc")
	})
}

func TestTrafficControlService_CreateFQCODELQdisc(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("application")
	service := NewTrafficControlService(eventStore, netlinkAdapter, logger)

	t.Run("creates_fq_codel_qdisc_successfully", func(t *testing.T) {
		ctx := context.Background()

		err := service.CreateFQCODELQdisc(ctx, "eth0", "1:0", 10240, 1024, 5000, 100000, 1518, false)

		assert.NoError(t, err)
	})

	t.Run("creates_fq_codel_qdisc_with_ecn", func(t *testing.T) {
		ctx := context.Background()

		err := service.CreateFQCODELQdisc(ctx, "eth1", "2:0", 10240, 1024, 5000, 100000, 1518, true)

		assert.NoError(t, err)
	})
}

func TestTrafficControlService_CreateHTBClass(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("application")
	service := NewTrafficControlService(eventStore, netlinkAdapter, logger)

	t.Run("creates_htb_class_successfully", func(t *testing.T) {
		ctx := context.Background()

		// First create the parent qdisc
		err := service.CreateHTBQdisc(ctx, "eth0", "1:0", "1:999")
		assert.NoError(t, err)

		// Then create the HTB class
		err = service.CreateHTBClass(ctx, "eth0", "1:0", "1:10", "10mbps", "50mbps")
		assert.NoError(t, err)
	})

	t.Run("fails_with_invalid_parameters", func(t *testing.T) {
		ctx := context.Background()

		err := service.CreateHTBClass(ctx, "", "1:0", "1:10", "invalid", "invalid")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create HTB class")
	})
}

func TestTrafficControlService_CreateFilter(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("application")
	service := NewTrafficControlService(eventStore, netlinkAdapter, logger)

	t.Run("creates_filter_successfully", func(t *testing.T) {
		ctx := context.Background()
		
		// First create the parent qdisc and class
		err := service.CreateHTBQdisc(ctx, "eth0", "1:0", "1:999")
		assert.NoError(t, err)
		
		err = service.CreateHTBClass(ctx, "eth0", "1:0", "1:10", "10mbps", "50mbps")
		assert.NoError(t, err)

		match := map[string]string{
			"dst_ip": "192.168.1.100",
		}

		err = service.CreateFilter(ctx, "eth0", "1:0", 100, "ip", "1:10", match)
		assert.NoError(t, err)
	})

	t.Run("creates_filter_with_multiple_matches", func(t *testing.T) {
		ctx := context.Background()
		
		// Create separate qdisc to avoid conflicts
		err := service.CreateHTBQdisc(ctx, "eth1", "2:0", "2:999")
		assert.NoError(t, err)
		
		err = service.CreateHTBClass(ctx, "eth1", "2:0", "2:10", "10mbps", "50mbps")
		assert.NoError(t, err)

		match := map[string]string{
			"dst_ip":   "192.168.1.100",
			"dst_port": "80",
			"protocol": "tcp",
		}

		err = service.CreateFilter(ctx, "eth1", "2:0", 100, "ip", "2:10", match)
		assert.NoError(t, err)
	})

	t.Run("fails_with_invalid_device", func(t *testing.T) {
		ctx := context.Background()
		match := map[string]string{
			"dst_ip": "192.168.1.100",
		}

		err := service.CreateFilter(ctx, "", "1:0", 100, "ip", "1:10", match)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create filter")
	})
}

func TestTrafficControlService_ParseHandle(t *testing.T) {
	testCases := []struct {
		name        string
		handleStr   string
		expectError bool
	}{
		{
			name:        "valid_handle_format",
			handleStr:   "1:0",
			expectError: false,
		},
		{
			name:        "valid_handle_with_hex",
			handleStr:   "a:b",
			expectError: false,
		},
		{
			name:        "invalid_handle_format",
			handleStr:   "invalid",
			expectError: true,
		},
		{
			name:        "empty_handle",
			handleStr:   "",
			expectError: true,
		},
		{
			name:        "partial_handle",
			handleStr:   "1:",
			expectError: false, // tc.ParseHandle handles this correctly
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			handle, err := tc.ParseHandle(testCase.handleStr)

			if testCase.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, handle.String(), "")
			}
		})
	}
}

func TestTrafficControlService_GetDeviceStatistics(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("application")
	service := NewTrafficControlService(eventStore, netlinkAdapter, logger)

	t.Run("fails_with_invalid_device_name", func(t *testing.T) {
		ctx := context.Background()

		stats, err := service.GetDeviceStatistics(ctx, "")

		assert.Error(t, err)
		assert.Nil(t, stats)
		assert.Contains(t, err.Error(), "invalid device name")
	})

	t.Run("fails_when_query_bus_execution_fails", func(t *testing.T) {
		ctx := context.Background()

		stats, err := service.GetDeviceStatistics(ctx, "eth0")

		assert.Error(t, err)
		assert.Nil(t, stats)
		assert.Contains(t, err.Error(), "failed to get device statistics")
	})
}

func TestTrafficControlService_GetQdiscStatistics(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("application")
	service := NewTrafficControlService(eventStore, netlinkAdapter, logger)

	t.Run("fails_with_invalid_device_name", func(t *testing.T) {
		ctx := context.Background()

		stats, err := service.GetQdiscStatistics(ctx, "", "1:0")

		assert.Error(t, err)
		assert.Nil(t, stats)
		assert.Contains(t, err.Error(), "invalid device name")
	})

	t.Run("fails_with_invalid_handle", func(t *testing.T) {
		ctx := context.Background()

		stats, err := service.GetQdiscStatistics(ctx, "eth0", "invalid")

		assert.Error(t, err)
		assert.Nil(t, stats)
		assert.Contains(t, err.Error(), "invalid handle")
	})
}

func TestTrafficControlService_GetClassStatistics(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("application")
	service := NewTrafficControlService(eventStore, netlinkAdapter, logger)

	t.Run("fails_with_invalid_device_name", func(t *testing.T) {
		ctx := context.Background()

		stats, err := service.GetClassStatistics(ctx, "", "1:10")

		assert.Error(t, err)
		assert.Nil(t, stats)
		assert.Contains(t, err.Error(), "invalid device name")
	})

	t.Run("fails_with_invalid_handle", func(t *testing.T) {
		ctx := context.Background()

		stats, err := service.GetClassStatistics(ctx, "eth0", "invalid")

		assert.Error(t, err)
		assert.Nil(t, stats)
		assert.Contains(t, err.Error(), "invalid handle")
	})
}

func TestTrafficControlService_MonitorStatistics(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("application")
	service := NewTrafficControlService(eventStore, netlinkAdapter, logger)

	t.Run("fails_with_invalid_device", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		callbackCalled := false
		done := make(chan error, 1)

		go func() {
			err := service.MonitorStatistics(ctx, "", time.Second, func(stats *qmodels.DeviceStatisticsView) {
				callbackCalled = true
			})
			done <- err
		}()

		// Cancel context immediately to stop monitoring
		cancel()

		select {
		case err := <-done:
			assert.Error(t, err)
			assert.False(t, callbackCalled)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("MonitorStatistics should have returned quickly")
		}
	})
}

func TestConvertApplicationStatsToView(t *testing.T) {
	t.Run("converts_application_stats_to_view", func(t *testing.T) {
		appStats := &DeviceStatistics{
			DeviceName: "eth0",
			QdiscStats: []QdiscStatistics{
				{
					Handle: "1:0",
					Type:   "htb",
					Stats: netlink.QdiscStats{
						BytesSent:   1000,
						PacketsSent: 10,
					},
				},
			},
			ClassStats: []ClassStatistics{
				{
					Handle: "1:10",
					Parent: "1:0",
					Name:   "web-traffic",
					Stats: netlink.ClassStats{
						BytesSent:   500,
						PacketsSent: 5,
					},
				},
			},
			FilterStats: []FilterStatistics{
				{
					Parent:     "1:0",
					Priority:   100,
					Protocol:   "ip",
					MatchCount: 5,
				},
			},
			LinkStats: LinkStatistics{
				RxBytes:   2000,
				TxBytes:   1500,
				RxPackets: 20,
				TxPackets: 15,
			},
		}

		view := convertApplicationStatsToView(appStats)

		assert.Equal(t, "eth0", view.DeviceName)
		assert.Len(t, view.QdiscStats, 1)
		assert.Equal(t, "1:0", view.QdiscStats[0].Handle)
		assert.Equal(t, "htb", view.QdiscStats[0].Type)
		assert.Equal(t, uint64(1000), view.QdiscStats[0].BytesSent)

		assert.Len(t, view.ClassStats, 1)
		assert.Equal(t, "1:10", view.ClassStats[0].Handle)
		assert.Equal(t, "1:0", view.ClassStats[0].Parent)
		assert.Equal(t, "web-traffic", view.ClassStats[0].Name)

		assert.Len(t, view.FilterStats, 1)
		assert.Equal(t, "1:0", view.FilterStats[0].Parent)
		assert.Equal(t, uint16(100), view.FilterStats[0].Priority)

		assert.Equal(t, uint64(2000), view.LinkStats.RxBytes)
		assert.Equal(t, uint64(1500), view.LinkStats.TxBytes)
	})

	t.Run("converts_stats_with_detailed_htb_stats", func(t *testing.T) {
		appStats := &DeviceStatistics{
			DeviceName: "eth0",
			QdiscStats: []QdiscStatistics{
				{
					Handle: "1:0",
					Type:   "htb",
					DetailedStats: &netlink.DetailedQdiscStats{
						Backlog:     100,
						QueueLength: 50,
						HTBStats: &netlink.HTBQdiscStats{
							DirectPackets: 25,
							Version:       3,
						},
					},
				},
			},
			ClassStats: []ClassStatistics{
				{
					Handle: "1:10",
					Parent: "1:0",
					DetailedStats: &netlink.DetailedClassStats{
						HTBStats: &netlink.HTBClassStats{
							Lends:   10,
							Borrows: 5,
							Tokens:  1000,
							Rate:    10485760, // 10 Mbps
							Ceil:    52428800, // 50 Mbps
							Level:   1,
						},
					},
				},
			},
		}

		view := convertApplicationStatsToView(appStats)

		require.Len(t, view.QdiscStats, 1)
		qdisc := view.QdiscStats[0]
		assert.Equal(t, uint32(100), qdisc.Backlog)
		assert.Equal(t, uint32(50), qdisc.QueueLength)
		assert.Equal(t, uint32(25), qdisc.DetailedStats["htb_direct_packets"])
		assert.Equal(t, uint32(3), qdisc.DetailedStats["htb_version"])

		require.Len(t, view.ClassStats, 1)
		class := view.ClassStats[0]
		assert.Equal(t, uint32(10), class.DetailedStats["htb_lends"])
		assert.Equal(t, uint32(5), class.DetailedStats["htb_borrows"])
		assert.Equal(t, uint32(1000), class.DetailedStats["htb_tokens"])
		assert.Equal(t, uint64(10485760), class.DetailedStats["htb_rate"])
		assert.Equal(t, uint64(52428800), class.DetailedStats["htb_ceil"])
		assert.Equal(t, uint32(1), class.DetailedStats["htb_level"])
	})
}

func TestTrafficControlService_PublishEvent(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("application")
	service := NewTrafficControlService(eventStore, netlinkAdapter, logger)

	t.Run("publishes_unknown_event_type", func(t *testing.T) {
		ctx := context.Background()
		unknownEvent := struct{ Name string }{Name: "test"}

		err := service.publishEvent(ctx, unknownEvent)

		assert.NoError(t, err) // Should not error, just skip unknown events
	})
}
