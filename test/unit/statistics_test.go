package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/internal/application"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	"github.com/rng999/traffic-control-go/internal/projections"
	"github.com/rng999/traffic-control-go/pkg/logging"
)

func TestStatisticsService_GetDeviceStatistics(t *testing.T) {
	// Setup
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("test")
	readModelStore := projections.NewMemoryReadModelStore()

	service := application.NewTrafficControlService(eventStore, netlinkAdapter, logger)
	ctx := context.Background()

	// Test device
	deviceName := "eth0"
	device, err := valueobjects.NewDevice(deviceName)
	require.NoError(t, err)

	// Setup mock data
	mockNetlinkAdapter := netlinkAdapter.(*netlink.MockAdapter)
	
	// Mock qdisc data
	mockQdiscInfo := []netlink.QdiscInfo{
		{
			Handle: valueobjects.NewHandle(1, 0),
			Type:   valueobjects.NewQdiscType("htb"),
			Statistics: netlink.QdiscStats{
				BytesSent:    1000000,
				PacketsSent:  10000,
				BytesDropped: 100,
				Overlimits:   10,
				Requeues:     5,
			},
		},
	}
	mockNetlinkAdapter.SetQdiscs(device, mockQdiscInfo)

	// Mock class data
	mockClassInfo := []netlink.ClassInfo{
		{
			Handle: valueobjects.NewHandle(1, 10),
			Parent: valueobjects.NewHandle(1, 0),
			Name:   "web-traffic",
			Statistics: netlink.ClassStats{
				BytesSent:      500000,
				PacketsSent:    5000,
				BytesDropped:   50,
				Overlimits:     5,
				BacklogBytes:   1024,
				BacklogPackets: 10,
				RateBPS:        1000000,
			},
		},
	}
	mockNetlinkAdapter.SetClasses(device, mockClassInfo)

	// Create some configuration to populate read model
	err = service.CreateHTBQdisc(ctx, deviceName, "1:", "1:999")
	require.NoError(t, err)

	err = service.CreateHTBClass(ctx, deviceName, "1:", "1:10", "1mbit", "10mbit")
	require.NoError(t, err)

	// Test GetDeviceStatistics
	stats, err := service.GetDeviceStatistics(ctx, deviceName)
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, deviceName, stats.DeviceName)
	assert.NotEmpty(t, stats.Timestamp)

	// Verify qdisc statistics
	assert.Len(t, stats.QdiscStats, 1)
	qdiscStat := stats.QdiscStats[0]
	assert.Equal(t, "1:0", qdiscStat.Handle)
	assert.Equal(t, "htb", qdiscStat.Type)
	assert.Equal(t, uint64(1000000), qdiscStat.BytesSent)
	assert.Equal(t, uint64(10000), qdiscStat.PacketsSent)

	// Verify class statistics
	assert.Len(t, stats.ClassStats, 1)
	classStat := stats.ClassStats[0]
	assert.Equal(t, "1:10", classStat.Handle)
	assert.Equal(t, "1:0", classStat.Parent)
	assert.Equal(t, "web-traffic", classStat.Name)
	assert.Equal(t, uint64(500000), classStat.BytesSent)
	assert.Equal(t, uint64(5000), classStat.PacketsSent)
}

func TestStatisticsService_GetRealtimeStatistics(t *testing.T) {
	// Setup
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("test")

	service := application.NewTrafficControlService(eventStore, netlinkAdapter, logger)
	ctx := context.Background()

	// Test device
	deviceName := "eth0"
	device, err := valueobjects.NewDevice(deviceName)
	require.NoError(t, err)

	// Setup mock data
	mockNetlinkAdapter := netlinkAdapter.(*netlink.MockAdapter)
	
	// Mock qdisc data
	mockQdiscInfo := []netlink.QdiscInfo{
		{
			Handle: valueobjects.NewHandle(1, 0),
			Type:   valueobjects.NewQdiscType("htb"),
			Statistics: netlink.QdiscStats{
				BytesSent:    2000000,
				PacketsSent:  20000,
				BytesDropped: 200,
				Overlimits:   20,
				Requeues:     10,
			},
		},
	}
	mockNetlinkAdapter.SetQdiscs(device, mockQdiscInfo)

	// Test GetRealtimeStatistics (no read model needed)
	stats, err := service.GetRealtimeStatistics(ctx, deviceName)
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, deviceName, stats.DeviceName)

	// Verify qdisc statistics
	assert.Len(t, stats.QdiscStats, 1)
	qdiscStat := stats.QdiscStats[0]
	assert.Equal(t, "1:0", qdiscStat.Handle)
	assert.Equal(t, "htb", qdiscStat.Type)
	assert.Equal(t, uint64(2000000), qdiscStat.BytesSent)
	assert.Equal(t, uint64(20000), qdiscStat.PacketsSent)
}

func TestStatisticsService_GetQdiscStatistics(t *testing.T) {
	// Setup
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("test")

	service := application.NewTrafficControlService(eventStore, netlinkAdapter, logger)
	ctx := context.Background()

	// Test device
	deviceName := "eth0"
	device, err := valueobjects.NewDevice(deviceName)
	require.NoError(t, err)

	// Setup mock data
	mockNetlinkAdapter := netlinkAdapter.(*netlink.MockAdapter)
	
	// Mock qdisc data
	mockQdiscInfo := []netlink.QdiscInfo{
		{
			Handle: valueobjects.NewHandle(1, 0),
			Type:   valueobjects.NewQdiscType("htb"),
			Statistics: netlink.QdiscStats{
				BytesSent:    3000000,
				PacketsSent:  30000,
				BytesDropped: 300,
				Overlimits:   30,
				Requeues:     15,
			},
		},
	}
	mockNetlinkAdapter.SetQdiscs(device, mockQdiscInfo)

	// Test GetQdiscStatistics
	stats, err := service.GetQdiscStatistics(ctx, deviceName, "1:0")
	require.NoError(t, err)
	assert.NotNil(t, stats)
	
	assert.Equal(t, "1:0", stats.Handle)
	assert.Equal(t, "htb", stats.Type)
	assert.Equal(t, uint64(3000000), stats.BytesSent)
	assert.Equal(t, uint64(30000), stats.PacketsSent)
	assert.Equal(t, uint64(300), stats.BytesDropped)
	assert.Equal(t, uint64(30), stats.Overlimits)
	assert.Equal(t, uint64(15), stats.Requeues)
}

func TestStatisticsService_GetClassStatistics(t *testing.T) {
	// Setup
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("test")

	service := application.NewTrafficControlService(eventStore, netlinkAdapter, logger)
	ctx := context.Background()

	// Test device
	deviceName := "eth0"
	device, err := valueobjects.NewDevice(deviceName)
	require.NoError(t, err)

	// Setup mock data
	mockNetlinkAdapter := netlinkAdapter.(*netlink.MockAdapter)
	
	// Mock class data
	mockClassInfo := []netlink.ClassInfo{
		{
			Handle: valueobjects.NewHandle(1, 10),
			Parent: valueobjects.NewHandle(1, 0),
			Name:   "high-priority",
			Statistics: netlink.ClassStats{
				BytesSent:      1500000,
				PacketsSent:    15000,
				BytesDropped:   150,
				Overlimits:     15,
				BacklogBytes:   2048,
				BacklogPackets: 20,
				RateBPS:        2000000,
			},
		},
	}
	mockNetlinkAdapter.SetClasses(device, mockClassInfo)

	// Test GetClassStatistics
	stats, err := service.GetClassStatistics(ctx, deviceName, "1:10")
	require.NoError(t, err)
	assert.NotNil(t, stats)
	
	assert.Equal(t, "1:10", stats.Handle)
	assert.Equal(t, "1:0", stats.Parent)
	assert.Equal(t, "high-priority", stats.Name)
	assert.Equal(t, uint64(1500000), stats.BytesSent)
	assert.Equal(t, uint64(15000), stats.PacketsSent)
	assert.Equal(t, uint64(150), stats.BytesDropped)
	assert.Equal(t, uint64(15), stats.Overlimits)
	assert.Equal(t, uint64(2048), stats.BacklogBytes)
	assert.Equal(t, uint64(20), stats.BacklogPackets)
	assert.Equal(t, uint64(2000000), stats.RateBPS)
}

func TestStatisticsService_MonitorStatistics(t *testing.T) {
	// Setup
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("test")

	service := application.NewTrafficControlService(eventStore, netlinkAdapter, logger)
	
	// Test device
	deviceName := "eth0"
	device, err := valueobjects.NewDevice(deviceName)
	require.NoError(t, err)

	// Setup mock data
	mockNetlinkAdapter := netlinkAdapter.(*netlink.MockAdapter)
	
	// Mock qdisc data
	mockQdiscInfo := []netlink.QdiscInfo{
		{
			Handle: valueobjects.NewHandle(1, 0),
			Type:   valueobjects.NewQdiscType("htb"),
			Statistics: netlink.QdiscStats{
				BytesSent:    1000,
				PacketsSent:  10,
				BytesDropped: 1,
				Overlimits:   0,
				Requeues:     0,
			},
		},
	}
	mockNetlinkAdapter.SetQdiscs(device, mockQdiscInfo)

	// Test monitoring for a short duration
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	callbackCount := 0
	err = service.MonitorStatistics(ctx, deviceName, 25*time.Millisecond, func(stats *application.DeviceStatisticsView) {
		callbackCount++
		assert.Equal(t, deviceName, stats.DeviceName)
		assert.NotEmpty(t, stats.Timestamp)
	})

	// Should timeout (which is expected behavior)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
	
	// Should have received some callbacks
	assert.Greater(t, callbackCount, 0)
}

func TestStatisticsService_ErrorHandling(t *testing.T) {
	// Setup
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("test")

	service := application.NewTrafficControlService(eventStore, netlinkAdapter, logger)
	ctx := context.Background()

	// Test invalid device name
	_, err := service.GetDeviceStatistics(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid device name")

	// Test invalid handle format
	_, err = service.GetQdiscStatistics(ctx, "eth0", "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid handle")

	// Test non-existent qdisc
	deviceName := "eth0"
	device, err := valueobjects.NewDevice(deviceName)
	require.NoError(t, err)

	mockNetlinkAdapter := netlinkAdapter.(*netlink.MockAdapter)
	mockNetlinkAdapter.SetQdiscs(device, []netlink.QdiscInfo{})

	_, err = service.GetQdiscStatistics(ctx, deviceName, "1:0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAPI_StatisticsMethods(t *testing.T) {
	// This test would go in api_test.go, but included here for completeness
	// Setup would create a TrafficController with mock dependencies
	// and test the statistics methods exposed through the API
}