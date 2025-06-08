//go:build integration
// +build integration

package integration_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/internal/application"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	"github.com/rng999/traffic-control-go/pkg/logging"
)

// TestComplexHTBHierarchy tests a multi-level HTB hierarchy using the type-safe command bus
// NOTE: This test should PASS completely - demonstrates successful complex configurations
func TestComplexHTBHierarchy(t *testing.T) {
	// Setup service with type-safe command bus
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	mockAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("complex-test")
	service := application.NewTrafficControlService(eventStore, mockAdapter, logger)
	ctx := context.Background()

	deviceName := "complex-eth0"

	// Create root HTB qdisc
	err := service.CreateHTBQdisc(ctx, deviceName, "1:0", "1:999")
	require.NoError(t, err, "Failed to create root HTB qdisc")

	// Create parent class (1:1) - Total bandwidth
	err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:1", "1Gbps", "1Gbps")
	require.NoError(t, err, "Failed to create parent class")

	// Create child classes under parent

	// High priority class (1:10)
	err = service.CreateHTBClass(ctx, deviceName, "1:1", "1:10", "400Mbps", "800Mbps")
	require.NoError(t, err, "Failed to create high priority class")

	// Medium priority class (1:20)
	err = service.CreateHTBClass(ctx, deviceName, "1:1", "1:20", "300Mbps", "600Mbps")
	require.NoError(t, err, "Failed to create medium priority class")

	// Low priority class (1:30)
	err = service.CreateHTBClass(ctx, deviceName, "1:1", "1:30", "300Mbps", "500Mbps")
	require.NoError(t, err, "Failed to create low priority class")

	// Create sub-classes under high priority for different services

	// VoIP traffic (1:11)
	err = service.CreateHTBClass(ctx, deviceName, "1:10", "1:11", "200Mbps", "400Mbps")
	require.NoError(t, err, "Failed to create VoIP class")

	// Video streaming (1:12)
	err = service.CreateHTBClass(ctx, deviceName, "1:10", "1:12", "200Mbps", "400Mbps")
	require.NoError(t, err, "Failed to create video streaming class")

	t.Log("Complex multi-level HTB hierarchy created successfully through type-safe command bus")
}

// TestMultipleQdiscTypes tests using different qdisc types in combination
func TestMultipleQdiscTypes(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	mockAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("multi-qdisc-test")
	service := application.NewTrafficControlService(eventStore, mockAdapter, logger)
	ctx := context.Background()

	t.Run("HTB with TBF leaf qdiscs", func(t *testing.T) {
		deviceName := "htb-tbf-eth0"

		// Create root HTB qdisc
		err := service.CreateHTBQdisc(ctx, deviceName, "1:0", "1:999")
		require.NoError(t, err, "Failed to create HTB root qdisc")

		// Create HTB classes
		err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:10", "100Mbps", "200Mbps")
		require.NoError(t, err, "Failed to create HTB class")

		// Create TBF qdisc attached to the class (leaf qdisc)
		err = service.CreateTBFQdisc(ctx, deviceName, "10:0", "100Mbps", 1600, 3000, 1600)
		require.NoError(t, err, "Failed to create TBF leaf qdisc")

		t.Log("HTB with TBF leaf qdisc configuration successful")
	})

	t.Run("PRIO qdisc with HTB leaves", func(t *testing.T) {
		deviceName := "prio-htb-eth0"

		// Create PRIO qdisc
		priomap := []uint8{1, 2, 2, 2, 1, 2, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1}
		err := service.CreatePRIOQdisc(ctx, deviceName, "1:0", 3, priomap)
		require.NoError(t, err, "Failed to create PRIO qdisc")

		// Create HTB qdiscs on each band
		err = service.CreateHTBQdisc(ctx, deviceName, "2:0", "2:999")
		require.NoError(t, err, "Failed to create HTB qdisc for band 0")

		err = service.CreateHTBQdisc(ctx, deviceName, "3:0", "3:999")
		require.NoError(t, err, "Failed to create HTB qdisc for band 1")

		err = service.CreateHTBQdisc(ctx, deviceName, "4:0", "4:999")
		require.NoError(t, err, "Failed to create HTB qdisc for band 2")

		t.Log("PRIO qdisc with HTB leaf qdiscs configuration successful")
	})
}

// TestFilterConfigurationScenarios tests complex filter scenarios
func TestFilterConfigurationScenarios(t *testing.T) {
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	mockAdapter := netlink.NewMockAdapter()
	logger := logging.WithComponent("filter-test")
	service := application.NewTrafficControlService(eventStore, mockAdapter, logger)
	ctx := context.Background()

	deviceName := "filter-eth0"

	// Setup base HTB configuration
	err := service.CreateHTBQdisc(ctx, deviceName, "1:0", "1:999")
	require.NoError(t, err, "Failed to create HTB qdisc")

	err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:10", "100Mbps", "200Mbps")
	require.NoError(t, err, "Failed to create high priority class")

	err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:20", "50Mbps", "100Mbps")
	require.NoError(t, err, "Failed to create normal priority class")

	err = service.CreateHTBClass(ctx, deviceName, "1:0", "1:30", "25Mbps", "50Mbps")
	require.NoError(t, err, "Failed to create low priority class")

	t.Run("Port-based filtering", func(t *testing.T) {
		// High priority: SSH (port 22)
		sshFilter := map[string]string{
			"dst_port": "22",
			"protocol": "tcp",
		}
		err := service.CreateFilter(ctx, deviceName, "1:0", 1, "ip", "1:10", sshFilter)
		require.NoError(t, err, "Failed to create SSH filter")

		// Normal priority: HTTP/HTTPS (ports 80, 443)
		httpFilter := map[string]string{
			"dst_port": "80",
			"protocol": "tcp",
		}
		err = service.CreateFilter(ctx, deviceName, "1:0", 2, "ip", "1:20", httpFilter)
		require.NoError(t, err, "Failed to create HTTP filter")

		httpsFilter := map[string]string{
			"dst_port": "443",
			"protocol": "tcp",
		}
		err = service.CreateFilter(ctx, deviceName, "1:0", 3, "ip", "1:20", httpsFilter)
		require.NoError(t, err, "Failed to create HTTPS filter")

		t.Log("Port-based filtering configuration successful")
	})

	t.Run("IP-based filtering", func(t *testing.T) {
		// High priority traffic to specific server
		serverFilter := map[string]string{
			"dst_ip": "192.168.1.100",
		}
		err := service.CreateFilter(ctx, deviceName, "1:0", 4, "ip", "1:10", serverFilter)
		require.NoError(t, err, "Failed to create server IP filter")

		// Low priority traffic to backup server
		backupFilter := map[string]string{
			"dst_ip": "192.168.1.200",
		}
		err = service.CreateFilter(ctx, deviceName, "1:0", 5, "ip", "1:30", backupFilter)
		require.NoError(t, err, "Failed to create backup server filter")

		t.Log("IP-based filtering configuration successful")
	})

	t.Log("Complex filter configuration scenarios completed successfully")
}
