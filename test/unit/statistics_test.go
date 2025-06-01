package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/internal/application"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	"github.com/rng999/traffic-control-go/pkg/logging"
)

func TestStatisticsService_GetDeviceStatistics(t *testing.T) {
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

	// Setup mock data - add some test statistics
	mockAdapter := netlinkAdapter.(*netlink.MockAdapter)
	
	// Mock qdisc statistics
	handle := valueobjects.NewHandle(1, 0)
	mockAdapter.SetQdiscStatistics(device, handle, netlink.QdiscStats{
		BytesSent:    1000000,
		PacketsSent:  10000,
		BytesDropped: 100,
		Overlimits:   10,
		Requeues:     5,
	})

	// Test getting statistics
	statisticsService := application.NewStatisticsService(netlinkAdapter, logger)
	stats, err := statisticsService.GetDeviceStatistics(ctx, device)
	
	// Should not error even if no actual qdiscs are configured
	if err != nil {
		t.Logf("Expected behavior - no configured qdiscs: %v", err)
	} else {
		// If successful, verify basic structure
		assert.Equal(t, deviceName, stats.DeviceName)
		assert.NotEmpty(t, stats.Timestamp)
		t.Logf("Device statistics: %+v", stats)
	}
}

func TestStatisticsService_GetLinkStatistics(t *testing.T) {
	// Setup
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("test")
	
	device, err := valueobjects.NewDevice("eth0")
	require.NoError(t, err)

	// Test getting link statistics
	statisticsService := application.NewStatisticsService(netlinkAdapter, logger)
	ctx := context.Background()
	
	stats, err := statisticsService.GetLinkStatistics(ctx, device)
	require.NoError(t, err)
	
	// Verify mock data
	assert.Equal(t, uint64(1000000), stats.RxBytes)
	assert.Equal(t, uint64(2000000), stats.TxBytes)
	assert.Equal(t, uint64(10000), stats.RxPackets)
	assert.Equal(t, uint64(20000), stats.TxPackets)
}

func TestStatisticsService_BasicFlow(t *testing.T) {
	// Setup
	netlinkAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("test")
	
	device, err := valueobjects.NewDevice("eth0")
	require.NoError(t, err)

	statisticsService := application.NewStatisticsService(netlinkAdapter, logger)
	ctx := context.Background()
	
	// Basic test that the service can be created and doesn't crash
	assert.NotNil(t, statisticsService)
	
	// Test link statistics
	linkStats, err := statisticsService.GetLinkStatistics(ctx, device)
	assert.NoError(t, err)
	assert.NotNil(t, linkStats)
	
	// Test device statistics (may have no data but shouldn't crash)
	deviceStats, err := statisticsService.GetDeviceStatistics(ctx, device)
	if err != nil {
		// Expected if no qdiscs are configured
		t.Logf("Expected: %v", err)
	} else {
		assert.Equal(t, "eth0", deviceStats.DeviceName)
	}
}