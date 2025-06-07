package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/internal/application"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// TestNetworkInterface tests the main entry point function
func TestNetworkInterface(t *testing.T) {
	t.Run("creates_traffic_controller_with_valid_device", func(t *testing.T) {
		controller := NetworkInterface("eth0")

		assert.NotNil(t, controller)
		assert.Equal(t, "eth0", controller.deviceName)
		assert.Empty(t, controller.classes)
		assert.NotNil(t, controller.logger)
		assert.NotNil(t, controller.service)
	})

	t.Run("creates_controller_with_different_device_names", func(t *testing.T) {
		testCases := []string{
			"eth0",
			"wlan0",
			"lo",
			"enp0s3",
			"docker0",
		}

		for _, deviceName := range testCases {
			t.Run(deviceName, func(t *testing.T) {
				controller := NetworkInterface(deviceName)
				assert.Equal(t, deviceName, controller.deviceName)
			})
		}
	})
}

// TestTrafficController_WithHardLimitBandwidth tests bandwidth configuration
func TestTrafficController_WithHardLimitBandwidth(t *testing.T) {
	t.Run("sets_total_bandwidth_from_string", func(t *testing.T) {
		controller := NetworkInterface("eth0")

		result := controller.WithHardLimitBandwidth("100mbps")

		// Should return self for chaining
		assert.Equal(t, controller, result)
		assert.Equal(t, tc.MustParseBandwidth("100mbps"), controller.totalBandwidth)
	})

	t.Run("handles_various_bandwidth_formats", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected tc.Bandwidth
		}{
			{"1gbps", tc.MustParseBandwidth("1gbps")},
			{"500mbps", tc.MustParseBandwidth("500mbps")},
			{"10kbps", tc.MustParseBandwidth("10kbps")},
			{"2048bps", tc.MustParseBandwidth("2048bps")},
		}

		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				controller := NetworkInterface("eth0")
				controller.WithHardLimitBandwidth(tc.input)
				assert.Equal(t, tc.expected, controller.totalBandwidth)
			})
		}
	})

	t.Run("panics_with_invalid_bandwidth_format", func(t *testing.T) {
		controller := NetworkInterface("eth0")

		assert.Panics(t, func() {
			controller.WithHardLimitBandwidth("invalid")
		})
	})
}

// TestTrafficController_CreateTrafficClass tests traffic class creation
func TestTrafficController_CreateTrafficClass(t *testing.T) {
	t.Run("creates_traffic_class_builder", func(t *testing.T) {
		controller := NetworkInterface("eth0")

		builder := controller.CreateTrafficClass("web-traffic")

		assert.NotNil(t, builder)
		assert.Equal(t, controller, builder.controller)
		assert.Equal(t, "web-traffic", builder.class.name)
		assert.Nil(t, builder.class.priority) // Priority should be nil initially
		assert.False(t, builder.finalized)
		assert.Len(t, controller.pendingBuilders, 1)
	})

	t.Run("creates_multiple_classes", func(t *testing.T) {
		controller := NetworkInterface("eth0")

		builder1 := controller.CreateTrafficClass("web-traffic")
		builder2 := controller.CreateTrafficClass("database-traffic")

		assert.NotEqual(t, builder1, builder2)
		assert.Equal(t, "web-traffic", builder1.class.name)
		assert.Equal(t, "database-traffic", builder2.class.name)
		assert.Len(t, controller.pendingBuilders, 2)
	})
}

// TestTrafficClassBuilder tests the fluent builder interface
func TestTrafficClassBuilder(t *testing.T) {
	controller := NetworkInterface("eth0")

	t.Run("fluent_interface_chaining", func(t *testing.T) {
		builder := controller.CreateTrafficClass("web-traffic")

		result := builder.
			WithGuaranteedBandwidth("10mbps").
			WithSoftLimitBandwidth("50mbps").
			WithPriority(1).
			ForDestination("192.168.1.100").
			ForPort(80, 443)

		// Should return self for chaining
		assert.Equal(t, builder, result)
	})

	t.Run("sets_guaranteed_bandwidth", func(t *testing.T) {
		builder := controller.CreateTrafficClass("web-traffic")

		builder.WithGuaranteedBandwidth("20mbps")

		expected := tc.MustParseBandwidth("20mbps")
		assert.Equal(t, expected, builder.class.guaranteedBandwidth)
	})

	t.Run("sets_soft_limit_bandwidth", func(t *testing.T) {
		builder := controller.CreateTrafficClass("web-traffic")

		builder.WithSoftLimitBandwidth("100mbps")

		expected := tc.MustParseBandwidth("100mbps")
		assert.Equal(t, expected, builder.class.maxBandwidth)
	})

	t.Run("sets_priority_within_valid_range", func(t *testing.T) {
		testCases := []struct {
			input    int
			expected uint8
		}{
			{0, 0},
			{3, 3},
			{7, 7},
			{-1, 0}, // Should clamp to 0
			{10, 7}, // Should clamp to 7
		}

		for _, tc := range testCases {
			t.Run(string(rune(tc.input+'0')), func(t *testing.T) {
				builder := controller.CreateTrafficClass("test-class")

				builder.WithPriority(tc.input)

				require.NotNil(t, builder.class.priority)
				assert.Equal(t, tc.expected, *builder.class.priority)
			})
		}
	})

	t.Run("adds_destination_ip_filters", func(t *testing.T) {
		builder := controller.CreateTrafficClass("web-traffic")

		builder.ForDestination("192.168.1.100")

		require.Len(t, builder.class.filters, 1)
		assert.Equal(t, DestinationIPFilter, builder.class.filters[0].filterType)
		assert.Equal(t, "192.168.1.100", builder.class.filters[0].value)
	})

	t.Run("adds_multiple_destination_ips", func(t *testing.T) {
		builder := controller.CreateTrafficClass("web-traffic")

		builder.ForDestinationIPs("192.168.1.100", "192.168.1.101")

		require.Len(t, builder.class.filters, 2)
		assert.Equal(t, DestinationIPFilter, builder.class.filters[0].filterType)
		assert.Equal(t, "192.168.1.100", builder.class.filters[0].value)
		assert.Equal(t, DestinationIPFilter, builder.class.filters[1].filterType)
		assert.Equal(t, "192.168.1.101", builder.class.filters[1].value)
	})

	t.Run("adds_source_ip_filters", func(t *testing.T) {
		builder := controller.CreateTrafficClass("web-traffic")

		builder.ForSource("10.0.0.1")

		require.Len(t, builder.class.filters, 1)
		assert.Equal(t, SourceIPFilter, builder.class.filters[0].filterType)
		assert.Equal(t, "10.0.0.1", builder.class.filters[0].value)
	})

	t.Run("adds_multiple_source_ips", func(t *testing.T) {
		builder := controller.CreateTrafficClass("web-traffic")

		builder.ForSourceIPs("10.0.0.1", "10.0.0.2")

		require.Len(t, builder.class.filters, 2)
		assert.Equal(t, SourceIPFilter, builder.class.filters[0].filterType)
		assert.Equal(t, "10.0.0.1", builder.class.filters[0].value)
		assert.Equal(t, SourceIPFilter, builder.class.filters[1].filterType)
		assert.Equal(t, "10.0.0.2", builder.class.filters[1].value)
	})

	t.Run("adds_port_filters", func(t *testing.T) {
		builder := controller.CreateTrafficClass("web-traffic")

		builder.ForPort(80, 443)

		require.Len(t, builder.class.filters, 2)
		assert.Equal(t, DestinationPortFilter, builder.class.filters[0].filterType)
		assert.Equal(t, 80, builder.class.filters[0].value)
		assert.Equal(t, DestinationPortFilter, builder.class.filters[1].filterType)
		assert.Equal(t, 443, builder.class.filters[1].value)
	})

	t.Run("adds_protocol_filters", func(t *testing.T) {
		builder := controller.CreateTrafficClass("web-traffic")

		builder.ForProtocols("tcp", "udp")

		require.Len(t, builder.class.filters, 2)
		assert.Equal(t, ProtocolFilter, builder.class.filters[0].filterType)
		assert.Equal(t, "tcp", builder.class.filters[0].value)
		assert.Equal(t, ProtocolFilter, builder.class.filters[1].filterType)
		assert.Equal(t, "udp", builder.class.filters[1].value)
	})
}

// TestTrafficController_Validation tests configuration validation
func TestTrafficController_Validation(t *testing.T) {
	t.Run("fails_when_total_bandwidth_not_set", func(t *testing.T) {
		controller := NetworkInterface("eth0")
		controller.CreateTrafficClass("web-traffic").
			WithGuaranteedBandwidth("10mbps").
			WithPriority(1)

		err := controller.Apply()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "total bandwidth not set")
	})

	t.Run("fails_when_class_priority_not_set", func(t *testing.T) {
		controller := NetworkInterface("eth0")
		controller.WithHardLimitBandwidth("100mbps")
		controller.CreateTrafficClass("web-traffic").
			WithGuaranteedBandwidth("10mbps")
		// Note: not setting priority

		err := controller.Apply()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not have a priority set")
		assert.Contains(t, err.Error(), "web-traffic")
	})

	t.Run("fails_when_max_bandwidth_exceeds_total", func(t *testing.T) {
		controller := NetworkInterface("eth0")
		controller.WithHardLimitBandwidth("100mbps")
		controller.CreateTrafficClass("web-traffic").
			WithGuaranteedBandwidth("10mbps").
			WithSoftLimitBandwidth("200mbps"). // Exceeds total
			WithPriority(1)

		err := controller.Apply()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max bandwidth")
		assert.Contains(t, err.Error(), "higher than total bandwidth")
	})

	t.Run("fails_when_guaranteed_exceeds_max", func(t *testing.T) {
		controller := NetworkInterface("eth0")
		controller.WithHardLimitBandwidth("100mbps")
		controller.CreateTrafficClass("web-traffic").
			WithGuaranteedBandwidth("50mbps").
			WithSoftLimitBandwidth("30mbps"). // Less than guaranteed
			WithPriority(1)

		err := controller.Apply()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "guaranteed bandwidth")
		assert.Contains(t, err.Error(), "higher than max bandwidth")
	})

	t.Run("fails_when_total_guaranteed_exceeds_total", func(t *testing.T) {
		controller := NetworkInterface("eth0")
		controller.WithHardLimitBandwidth("100mbps")
		controller.CreateTrafficClass("web-traffic").
			WithGuaranteedBandwidth("60mbps").
			WithPriority(1)
		controller.CreateTrafficClass("database-traffic").
			WithGuaranteedBandwidth("60mbps"). // Total: 120mbps > 100mbps
			WithPriority(2)

		err := controller.Apply()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "total guaranteed bandwidth")
		assert.Contains(t, err.Error(), "exceeds interface bandwidth")
	})

	t.Run("passes_with_valid_configuration", func(t *testing.T) {
		controller := NetworkInterface("eth0")
		controller.WithHardLimitBandwidth("100mbps")
		controller.CreateTrafficClass("web-traffic").
			WithGuaranteedBandwidth("30mbps").
			WithSoftLimitBandwidth("50mbps").
			WithPriority(1)
		controller.CreateTrafficClass("database-traffic").
			WithGuaranteedBandwidth("20mbps").
			WithSoftLimitBandwidth("40mbps").
			WithPriority(2)

		// Mock the service to avoid actual netlink calls
		mockEventStore := eventstore.NewMemoryEventStoreWithContext()
		mockNetlinkAdapter := netlink.NewMockAdapter()
		controller.service = application.NewTrafficControlService(mockEventStore, mockNetlinkAdapter, controller.logger)

		err := controller.Apply()

		assert.NoError(t, err)
	})
}

// TestHTBQdiscBuilder tests HTB qdisc builder
func TestHTBQdiscBuilder(t *testing.T) {
	controller := NetworkInterface("eth0")

	t.Run("creates_htb_qdisc_builder", func(t *testing.T) {
		builder := controller.CreateHTBQdisc("1:0", "1:1")

		assert.NotNil(t, builder)
		assert.Equal(t, controller, builder.controller)
		assert.Equal(t, "1:0", builder.handle)
		assert.Equal(t, "1:1", builder.defaultClass)
		assert.Empty(t, builder.classes)
	})

	t.Run("adds_classes_to_qdisc", func(t *testing.T) {
		builder := controller.CreateHTBQdisc("1:0", "1:1")

		result := builder.AddClass("1:0", "1:10", "web-traffic", "10mbps", "50mbps")

		assert.Equal(t, builder, result) // Should return self for chaining
		require.Len(t, builder.classes, 1)

		class := builder.classes[0]
		assert.Equal(t, "1:0", class.parent)
		assert.Equal(t, "1:10", class.handle)
		assert.Equal(t, "web-traffic", class.name)
		assert.Equal(t, "10mbps", class.rate)
		assert.Equal(t, "50mbps", class.ceil)
	})
}

// TestTBFQdiscBuilder tests TBF qdisc builder
func TestTBFQdiscBuilder(t *testing.T) {
	controller := NetworkInterface("eth0")

	t.Run("creates_tbf_qdisc_builder_with_defaults", func(t *testing.T) {
		builder := controller.CreateTBFQdisc("1:0", "10mbps")

		assert.NotNil(t, builder)
		assert.Equal(t, "1:0", builder.handle)
		assert.Equal(t, "10mbps", builder.rate)
		assert.Equal(t, uint32(32768), builder.buffer) // Default buffer
		assert.Equal(t, uint32(10000), builder.limit)  // Default limit
		assert.Equal(t, uint32(0), builder.burst)      // Default burst
	})

	t.Run("allows_customization_with_fluent_interface", func(t *testing.T) {
		builder := controller.CreateTBFQdisc("1:0", "10mbps")

		result := builder.
			WithBuffer(65536).
			WithLimit(20000).
			WithBurst(1000)

		assert.Equal(t, builder, result) // Should return self for chaining
		assert.Equal(t, uint32(65536), builder.buffer)
		assert.Equal(t, uint32(20000), builder.limit)
		assert.Equal(t, uint32(1000), builder.burst)
	})
}

// TestPRIOQdiscBuilder tests PRIO qdisc builder
func TestPRIOQdiscBuilder(t *testing.T) {
	controller := NetworkInterface("eth0")

	t.Run("creates_prio_qdisc_builder_with_defaults", func(t *testing.T) {
		builder := controller.CreatePRIOQdisc("1:0", 3)

		assert.NotNil(t, builder)
		assert.Equal(t, "1:0", builder.handle)
		assert.Equal(t, uint8(3), builder.bands)
		assert.Len(t, builder.priomap, 16) // Default priomap length
	})

	t.Run("allows_custom_priomap", func(t *testing.T) {
		builder := controller.CreatePRIOQdisc("1:0", 3)
		customPriomap := []uint8{0, 1, 2, 0, 1, 2, 0, 1, 0, 1, 2, 0, 1, 2, 0, 1}

		result := builder.WithPriomap(customPriomap)

		assert.Equal(t, builder, result)
		assert.Equal(t, customPriomap, builder.priomap)
	})

	t.Run("ignores_invalid_priomap_length", func(t *testing.T) {
		builder := controller.CreatePRIOQdisc("1:0", 3)
		originalPriomap := builder.priomap
		invalidPriomap := []uint8{0, 1, 2} // Too short

		builder.WithPriomap(invalidPriomap)

		assert.Equal(t, originalPriomap, builder.priomap) // Should remain unchanged
	})
}

// TestFQCODELQdiscBuilder tests FQ_CODEL qdisc builder
func TestFQCODELQdiscBuilder(t *testing.T) {
	controller := NetworkInterface("eth0")

	t.Run("creates_fq_codel_qdisc_builder_with_defaults", func(t *testing.T) {
		builder := controller.CreateFQCODELQdisc("1:0")

		assert.NotNil(t, builder)
		assert.Equal(t, "1:0", builder.handle)
		assert.Equal(t, uint32(10240), builder.limit)
		assert.Equal(t, uint32(1024), builder.flows)
		assert.Equal(t, uint32(5000), builder.target)
		assert.Equal(t, uint32(100000), builder.interval)
		assert.Equal(t, uint32(1518), builder.quantum)
		assert.False(t, builder.ecn)
	})

	t.Run("allows_customization_with_fluent_interface", func(t *testing.T) {
		builder := controller.CreateFQCODELQdisc("1:0")

		result := builder.
			WithLimit(20480).
			WithFlows(2048).
			WithTarget(10000).
			WithInterval(200000).
			WithQuantum(3036).
			WithECN(true)

		assert.Equal(t, builder, result)
		assert.Equal(t, uint32(20480), builder.limit)
		assert.Equal(t, uint32(2048), builder.flows)
		assert.Equal(t, uint32(10000), builder.target)
		assert.Equal(t, uint32(200000), builder.interval)
		assert.Equal(t, uint32(3036), builder.quantum)
		assert.True(t, builder.ecn)
	})
}

// TestBuildFilterMatch tests the internal filter matching logic
func TestBuildFilterMatch(t *testing.T) {
	controller := NetworkInterface("eth0")

	testCases := []struct {
		name           string
		filter         Filter
		expectedResult map[string]string
	}{
		{
			name:   "source_ip_filter",
			filter: Filter{filterType: SourceIPFilter, value: "192.168.1.100"},
			expectedResult: map[string]string{
				"src_ip": "192.168.1.100",
			},
		},
		{
			name:   "destination_ip_filter",
			filter: Filter{filterType: DestinationIPFilter, value: "10.0.0.1"},
			expectedResult: map[string]string{
				"dst_ip": "10.0.0.1",
			},
		},
		{
			name:   "source_port_filter",
			filter: Filter{filterType: SourcePortFilter, value: 8080},
			expectedResult: map[string]string{
				"src_port": "8080",
			},
		},
		{
			name:   "destination_port_filter",
			filter: Filter{filterType: DestinationPortFilter, value: 443},
			expectedResult: map[string]string{
				"dst_port": "443",
			},
		},
		{
			name:   "protocol_filter",
			filter: Filter{filterType: ProtocolFilter, value: "tcp"},
			expectedResult: map[string]string{
				"protocol": "tcp",
			},
		},
		{
			name:           "invalid_filter_type_returns_empty",
			filter:         Filter{filterType: FilterType(999), value: "invalid"},
			expectedResult: map[string]string{},
		},
		{
			name:           "invalid_value_type_returns_empty",
			filter:         Filter{filterType: SourceIPFilter, value: 123}, // Wrong type
			expectedResult: map[string]string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := controller.buildFilterMatch(tc.filter)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

// TestFinalizePendingClasses tests automatic class registration
func TestFinalizePendingClasses(t *testing.T) {
	t.Run("finalizes_pending_builders", func(t *testing.T) {
		controller := NetworkInterface("eth0")

		// Create some pending classes
		controller.CreateTrafficClass("web-traffic").WithPriority(1)
		controller.CreateTrafficClass("database-traffic").WithPriority(2)

		assert.Len(t, controller.pendingBuilders, 2)
		assert.Len(t, controller.classes, 0)

		controller.finalizePendingClasses()

		assert.Len(t, controller.pendingBuilders, 0)
		assert.Len(t, controller.classes, 2)
		assert.Equal(t, "web-traffic", controller.classes[0].name)
		assert.Equal(t, "database-traffic", controller.classes[1].name)
	})

	t.Run("marks_builders_as_finalized", func(t *testing.T) {
		controller := NetworkInterface("eth0")

		builder := controller.CreateTrafficClass("web-traffic").WithPriority(1)
		assert.False(t, builder.finalized)

		controller.finalizePendingClasses()

		assert.True(t, builder.finalized)
	})
}
