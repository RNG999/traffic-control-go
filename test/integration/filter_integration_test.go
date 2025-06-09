package integration

import (
	"context"
	"testing"

	"github.com/rng999/traffic-control-go/internal/application"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	"github.com/rng999/traffic-control-go/pkg/logging"
	"github.com/stretchr/testify/require"
)

func TestFilterIntegrationBasic(t *testing.T) {
	// Create application service with mock adapter
	logger, _ := logging.NewLogger(logging.Config{Level: "error"})
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	adapter := netlink.NewMockAdapter()
	service := application.NewTrafficControlService(eventStore, adapter, logger)

	// Test basic filter creation workflow
	t.Run("create HTB qdisc, class, and filter", func(t *testing.T) {
		// Create HTB qdisc
		err := service.CreateHTBQdisc(context.Background(), "eth0", "1:", "1:10")
		require.NoError(t, err)

		// Create target class
		err = service.CreateHTBClass(context.Background(), "eth0", "1:", "1:20", "5Mbps", "10Mbps")
		require.NoError(t, err)

		// Create filter
		filterMatch := map[string]string{
			"dst_port": "80",
			"src_ip":   "192.168.1.0/24",
		}
		err = service.CreateFilter(context.Background(), "eth0", "1:", 100, "ip", "1:20", filterMatch)
		require.NoError(t, err)
	})

	t.Run("create multiple filters with different priorities", func(t *testing.T) {
		// Create additional classes
		err := service.CreateHTBClass(context.Background(), "eth0", "1:", "1:30", "2Mbps", "8Mbps")
		require.NoError(t, err)

		// Create filters with different priorities
		filters := []struct {
			priority uint16
			port     string
			classID  string
		}{
			{10, "22", "1:20"},   // SSH -> High priority
			{50, "443", "1:10"},  // HTTPS -> Default
			{100, "8080", "1:30"}, // Bulk -> Low priority
		}

		for _, filter := range filters {
			filterMatch := map[string]string{"dst_port": filter.port}
			err = service.CreateFilter(context.Background(), "eth0", "1:", filter.priority, "ip", filter.classID, filterMatch)
			require.NoError(t, err)
		}
	})

	t.Run("create complex filter with multiple match criteria", func(t *testing.T) {
		// Create complex filter with multiple match criteria
		filterMatch := map[string]string{
			"src_ip":   "10.0.0.0/8",
			"dst_ip":   "172.16.0.0/12",
			"dst_port": "25",
			"src_port": "1024",
		}
		err := service.CreateFilter(context.Background(), "eth0", "1:", 25, "ip", "1:20", filterMatch)
		require.NoError(t, err)
	})
}

func TestFilterConfigurationRetrieval(t *testing.T) {
	// Create application service
	logger, _ := logging.NewLogger(logging.Config{Level: "error"})
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	adapter := netlink.NewMockAdapter()
	service := application.NewTrafficControlService(eventStore, adapter, logger)

	// Setup basic configuration
	err := service.CreateHTBQdisc(context.Background(), "eth0", "1:", "1:10")
	require.NoError(t, err)

	err = service.CreateHTBClass(context.Background(), "eth0", "1:", "1:10", "5Mbps", "10Mbps")
	require.NoError(t, err)

	err = service.CreateHTBClass(context.Background(), "eth0", "1:", "1:20", "3Mbps", "8Mbps")
	require.NoError(t, err)

	// Create filter
	filterMatch := map[string]string{
		"dst_port": "80",
		"src_ip":   "192.168.1.0/24",
	}
	err = service.CreateFilter(context.Background(), "eth0", "1:", 100, "ip", "1:20", filterMatch)
	require.NoError(t, err)

	// Test that configuration can be retrieved
	config, err := service.GetConfiguration(context.Background(), "eth0")
	require.NoError(t, err)
	require.NotNil(t, config)

	// Basic verification that configuration contains expected elements
	require.NotEmpty(t, config.Qdiscs)
	require.NotEmpty(t, config.Classes)
	require.NotEmpty(t, config.Filters)
}

func TestFilterStatisticsRetrieval(t *testing.T) {
	// Create application service
	logger, _ := logging.NewLogger(logging.Config{Level: "error"})
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	adapter := netlink.NewMockAdapter()
	service := application.NewTrafficControlService(eventStore, adapter, logger)

	// Setup basic configuration
	err := service.CreateHTBQdisc(context.Background(), "eth0", "1:", "1:10")
	require.NoError(t, err)

	err = service.CreateHTBClass(context.Background(), "eth0", "1:", "1:10", "5Mbps", "10Mbps")
	require.NoError(t, err)

	// Create filter
	filterMatch := map[string]string{"dst_port": "80"}
	err = service.CreateFilter(context.Background(), "eth0", "1:", 100, "ip", "1:10", filterMatch)
	require.NoError(t, err)

	// Test that device statistics can be retrieved
	stats, err := service.GetDeviceStatistics(context.Background(), "eth0")
	require.NoError(t, err)
	require.NotNil(t, stats)

	// Test qdisc statistics
	qdiscStats, err := service.GetQdiscStatistics(context.Background(), "eth0", "1:")
	require.NoError(t, err)
	require.NotNil(t, qdiscStats)

	// Test class statistics
	classStats, err := service.GetClassStatistics(context.Background(), "eth0", "1:10")
	require.NoError(t, err)
	require.NotNil(t, classStats)
}