package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
)

// TestMockNetlinkAdapter tests the mock implementation
func TestMockNetlinkAdapter(t *testing.T) {
	deviceName := valueobjects.MustNewDeviceName("eth0")

	t.Run("AddAndGetQdisc", func(t *testing.T) {
		adapter := netlink.NewMockAdapter()
		// Test adding a qdisc
		config := netlink.QdiscConfig{
			Handle: valueobjects.NewHandle(1, 0),
			Type:   entities.QdiscTypeHTB,
			Parameters: map[string]interface{}{
				"defaultClass": valueobjects.NewHandle(1, 1),
			},
		}

		result := adapter.AddQdisc(deviceName, config)
		require.True(t, result.IsSuccess())

		// Test getting qdiscs
		qdiscsResult := adapter.GetQdiscs(deviceName)
		require.True(t, qdiscsResult.IsSuccess())

		qdiscs := qdiscsResult.Value()
		assert.Len(t, qdiscs, 1)
		assert.Equal(t, config.Handle, qdiscs[0].Handle)
		assert.Equal(t, config.Type, qdiscs[0].Type)
	})

	t.Run("AddAndGetClass", func(t *testing.T) {
		adapter := netlink.NewMockAdapter()
		// First add a parent qdisc
		qdiscConfig := netlink.QdiscConfig{
			Handle: valueobjects.NewHandle(1, 0),
			Type:   entities.QdiscTypeHTB,
		}
		result := adapter.AddQdisc(deviceName, qdiscConfig)
		require.True(t, result.IsSuccess())

		// Add a class
		classConfig := netlink.ClassConfig{
			Handle: valueobjects.NewHandle(1, 1),
			Parent: valueobjects.NewHandle(1, 0),
			Type:   entities.QdiscTypeHTB,
			Parameters: map[string]interface{}{
				"rate": valueobjects.MustParseBandwidth("10Mbps"),
				"ceil": valueobjects.MustParseBandwidth("20Mbps"),
			},
		}

		result = adapter.AddClass(deviceName, classConfig)
		if result.IsFailure() {
			t.Fatalf("AddClass failed: %v", result.Error())
		}
		require.True(t, result.IsSuccess())

		// Test getting classes
		classesResult := adapter.GetClasses(deviceName)
		require.True(t, classesResult.IsSuccess())

		classes := classesResult.Value()
		assert.Len(t, classes, 1)
		assert.Equal(t, classConfig.Handle, classes[0].Handle)
		assert.Equal(t, classConfig.Parent, classes[0].Parent)
	})

	t.Run("AddAndGetFilter", func(t *testing.T) {
		adapter := netlink.NewMockAdapter()
		// Add filter
		filterConfig := netlink.FilterConfig{
			Parent:   valueobjects.NewHandle(1, 0),
			Priority: 1,
			Handle:   valueobjects.NewHandle(800, 1),
			Protocol: entities.ProtocolIP,
			FlowID:   valueobjects.NewHandle(1, 1),
			Matches: []netlink.FilterMatch{
				{
					Type:  entities.MatchTypeIPDestination,
					Value: "192.168.1.100",
				},
			},
		}

		result := adapter.AddFilter(deviceName, filterConfig)
		require.True(t, result.IsSuccess())

		// Test getting filters
		filtersResult := adapter.GetFilters(deviceName)
		require.True(t, filtersResult.IsSuccess())

		filters := filtersResult.Value()
		assert.Len(t, filters, 1)
		assert.Equal(t, filterConfig.Parent, filters[0].Parent)
		assert.Equal(t, filterConfig.Priority, filters[0].Priority)
	})

	t.Run("ErrorCases", func(t *testing.T) {
		adapter := netlink.NewMockAdapter()
		// Test duplicate qdisc
		config := netlink.QdiscConfig{
			Handle: valueobjects.NewHandle(1, 0),
			Type:   entities.QdiscTypeHTB,
		}

		result1 := adapter.AddQdisc(deviceName, config)
		require.True(t, result1.IsSuccess())

		result2 := adapter.AddQdisc(deviceName, config)
		require.True(t, result2.IsFailure())
		assert.Contains(t, result2.Error().Error(), "already exists")

		// Test class without parent
		classConfig := netlink.ClassConfig{
			Handle: valueobjects.NewHandle(2, 1),
			Parent: valueobjects.NewHandle(2, 0), // Non-existent parent
			Type:   entities.QdiscTypeHTB,
		}

		result3 := adapter.AddClass(deviceName, classConfig)
		if result3.IsSuccess() {
			t.Fatalf("AddClass should have failed but succeeded")
		}
		require.True(t, result3.IsFailure())
		assert.Contains(t, result3.Error().Error(), "parent")
	})
}
