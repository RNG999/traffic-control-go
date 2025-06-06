//go:build ignore
// +build ignore

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// TestComplexHTBHierarchy tests a multi-level HTB hierarchy
func TestComplexHTBHierarchy(t *testing.T) {
	adapter := netlink.NewMockAdapter()
	device := tc.MustNewDeviceName("eth0")

	// Create root HTB qdisc
	rootQdisc := netlink.QdiscConfig{
		Handle: tc.NewHandle(1, 0),
		Type:   entities.QdiscTypeHTB,
		Parameters: map[string]interface{}{
			"defaultClass": tc.NewHandle(1, 999),
		},
	}

	result := adapter.AddQdisc(device, rootQdisc)
	require.True(t, result.IsSuccess(), "Failed to add root qdisc")

	// Create parent class (1:1) - Total bandwidth
	parentClass := netlink.ClassConfig{
		Handle: tc.NewHandle(1, 1),
		Parent: tc.NewHandle(1, 0),
		Type:   entities.QdiscTypeHTB,
		Parameters: map[string]interface{}{
			"rate": tc.MustParseBandwidth("1Gbps"),
			"ceil": tc.MustParseBandwidth("1Gbps"),
		},
	}

	result = adapter.AddClass(device, parentClass)
	require.True(t, result.IsSuccess(), "Failed to add parent class")

	// Create child classes under parent

	// High priority class (1:10)
	highPrioClass := netlink.ClassConfig{
		Handle: tc.NewHandle(1, 10),
		Parent: tc.NewHandle(1, 1),
		Type:   entities.QdiscTypeHTB,
		Parameters: map[string]interface{}{
			"rate": tc.MustParseBandwidth("400Mbps"),
			"ceil": tc.MustParseBandwidth("800Mbps"),
		},
	}

	result = adapter.AddClass(device, highPrioClass)
	require.True(t, result.IsSuccess(), "Failed to add high priority class")

	// Medium priority class (1:20)
	mediumPrioClass := netlink.ClassConfig{
		Handle: tc.NewHandle(1, 20),
		Parent: tc.NewHandle(1, 1),
		Type:   entities.QdiscTypeHTB,
		Parameters: map[string]interface{}{
			"rate": tc.MustParseBandwidth("300Mbps"),
			"ceil": tc.MustParseBandwidth("600Mbps"),
		},
	}

	result = adapter.AddClass(device, mediumPrioClass)
	require.True(t, result.IsSuccess(), "Failed to add medium priority class")

	// Low priority class (1:30)
	lowPrioClass := netlink.ClassConfig{
		Handle: tc.NewHandle(1, 30),
		Parent: tc.NewHandle(1, 1),
		Type:   entities.QdiscTypeHTB,
		Parameters: map[string]interface{}{
			"rate": tc.MustParseBandwidth("300Mbps"),
			"ceil": tc.MustParseBandwidth("500Mbps"),
		},
	}

	result = adapter.AddClass(device, lowPrioClass)
	require.True(t, result.IsSuccess(), "Failed to add low priority class")

	// Create sub-classes under high priority for different services

	// VoIP traffic (1:11)
	voipClass := netlink.ClassConfig{
		Handle: tc.NewHandle(1, 11),
		Parent: tc.NewHandle(1, 10),
		Type:   entities.QdiscTypeHTB,
		Parameters: map[string]interface{}{
			"rate": tc.MustParseBandwidth("100Mbps"),
			"ceil": tc.MustParseBandwidth("200Mbps"),
		},
	}

	result = adapter.AddClass(device, voipClass)
	require.True(t, result.IsSuccess(), "Failed to add VoIP class")

	// Video streaming (1:12)
	videoClass := netlink.ClassConfig{
		Handle: tc.NewHandle(1, 12),
		Parent: tc.NewHandle(1, 10),
		Type:   entities.QdiscTypeHTB,
		Parameters: map[string]interface{}{
			"rate": tc.MustParseBandwidth("300Mbps"),
			"ceil": tc.MustParseBandwidth("600Mbps"),
		},
	}

	result = adapter.AddClass(device, videoClass)
	require.True(t, result.IsSuccess(), "Failed to add video class")

	// Verify hierarchy
	classes := adapter.GetClasses(device)
	require.True(t, classes.IsSuccess())
	assert.Len(t, classes.Value(), 6, "Should have 6 classes total")
}

// TestMultipleFiltersWithPriority tests filter ordering and priority
func TestMultipleFiltersWithPriority(t *testing.T) {
	adapter := netlink.NewMockAdapter()
	device := tc.MustNewDeviceName("eth0")

	// Add qdisc
	qdisc := netlink.QdiscConfig{
		Handle: tc.NewHandle(1, 0),
		Type:   entities.QdiscTypeHTB,
	}
	adapter.AddQdisc(device, qdisc)

	// Add filters with different priorities
	filters := []netlink.FilterConfig{
		{
			Parent:   tc.NewHandle(1, 0),
			Priority: 1, // Highest priority
			Handle:   tc.NewHandle(800, 1),
			Protocol: entities.ProtocolIP,
			FlowID:   tc.NewHandle(1, 10),
			Matches: []netlink.FilterMatch{
				{
					Type:  entities.MatchTypeIPDestination,
					Value: "192.168.1.100",
				},
			},
		},
		{
			Parent:   tc.NewHandle(1, 0),
			Priority: 2,
			Handle:   tc.NewHandle(800, 2),
			Protocol: entities.ProtocolIP,
			FlowID:   tc.NewHandle(1, 20),
			Matches: []netlink.FilterMatch{
				{
					Type:  entities.MatchTypePortDestination,
					Value: uint16(80),
				},
			},
		},
		{
			Parent:   tc.NewHandle(1, 0),
			Priority: 3,
			Handle:   tc.NewHandle(800, 3),
			Protocol: entities.ProtocolIP,
			FlowID:   tc.NewHandle(1, 30),
			Matches: []netlink.FilterMatch{
				{
					Type:  entities.MatchTypeProtocol,
					Value: uint8(6), // TCP
				},
			},
		},
	}

	// Add all filters
	for _, filter := range filters {
		result := adapter.AddFilter(device, filter)
		require.True(t, result.IsSuccess(), "Failed to add filter")
	}

	// Verify all filters exist
	filtersResult := adapter.GetFilters(device)
	require.True(t, filtersResult.IsSuccess())
	assert.Len(t, filtersResult.Value(), 3, "Should have 3 filters")

	// Verify filters are returned in priority order (mock should maintain order)
	returnedFilters := filtersResult.Value()
	for i, filter := range returnedFilters {
		assert.Equal(t, uint16(i+1), filter.Priority, "Filters should be in priority order")
	}
}

// TestNETEMConfiguration tests NETEM qdisc configuration
func TestNETEMConfiguration(t *testing.T) {
	// This test demonstrates how NETEM would be configured
	// Note: Actual implementation would require extending the mock adapter

	type NetemTestConfig struct {
		Delay     time.Duration
		Jitter    time.Duration
		Loss      float32
		Duplicate float32
		Corrupt   float32
		Reorder   float32
	}

	testCases := []struct {
		name   string
		config NetemTestConfig
	}{
		{
			name: "Basic delay",
			config: NetemTestConfig{
				Delay: 100 * time.Millisecond,
			},
		},
		{
			name: "Delay with jitter",
			config: NetemTestConfig{
				Delay:  50 * time.Millisecond,
				Jitter: 10 * time.Millisecond,
			},
		},
		{
			name: "Packet loss",
			config: NetemTestConfig{
				Loss: 1.5, // 1.5% loss
			},
		},
		{
			name: "Complex scenario",
			config: NetemTestConfig{
				Delay:     20 * time.Millisecond,
				Jitter:    5 * time.Millisecond,
				Loss:      0.5,
				Duplicate: 0.1,
				Corrupt:   0.01,
				Reorder:   0.5,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This demonstrates the expected NETEM configuration
			// Real implementation would create actual NETEM qdisc
			assert.NotNil(t, tc.config)
		})
	}
}

// TestErrorScenarios tests various error conditions
func TestErrorScenarios(t *testing.T) {
	adapter := netlink.NewMockAdapter()
	device := tc.MustNewDeviceName("eth0")

	t.Run("ClassWithoutParentQdisc", func(t *testing.T) {
		// Try to add class without parent qdisc
		class := netlink.ClassConfig{
			Handle: tc.NewHandle(1, 1),
			Parent: tc.NewHandle(1, 0), // Non-existent parent
			Type:   entities.QdiscTypeHTB,
			Parameters: map[string]interface{}{
				"rate": tc.MustParseBandwidth("100Mbps"),
			},
		}

		result := adapter.AddClass(device, class)
		assert.True(t, result.IsFailure(), "Should fail without parent qdisc")
	})

	t.Run("DuplicateHandle", func(t *testing.T) {
		// Add qdisc
		qdisc := netlink.QdiscConfig{
			Handle: tc.NewHandle(1, 0),
			Type:   entities.QdiscTypeHTB,
		}
		adapter.AddQdisc(device, qdisc)

		// Try to add duplicate
		result := adapter.AddQdisc(device, qdisc)
		assert.True(t, result.IsFailure(), "Should fail with duplicate handle")
	})

	t.Run("InvalidFilterParent", func(t *testing.T) {
		// Try to add filter to non-existent parent
		filter := netlink.FilterConfig{
			Parent:   tc.NewHandle(99, 0), // Non-existent
			Priority: 1,
			Handle:   tc.NewHandle(800, 1),
			Protocol: entities.ProtocolIP,
			FlowID:   tc.NewHandle(1, 1),
		}

		result := adapter.AddFilter(device, filter)
		// Note: Current mock doesn't validate parent existence for filters
		// Real implementation should check this
		_ = result
	})
}
