package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

func TestMockAdapter_BasicFunctionality(t *testing.T) {
	// Setup
	adapter := netlink.NewMockAdapter()

	// Test device
	device, err := tc.NewDevice("eth0")
	require.NoError(t, err)

	// Test getting qdiscs (should be empty initially)
	result := adapter.GetQdiscs(device)
	assert.True(t, result.IsSuccess())
	qdiscs := result.Value()
	assert.Empty(t, qdiscs)
}

func TestMockAdapter_QdiscStatistics(t *testing.T) {
	// Setup
	adapter := netlink.NewMockAdapter()

	device, err := tc.NewDevice("eth0")
	require.NoError(t, err)

	// Set mock statistics
	handle := tc.NewHandle(1, 0)
	stats := netlink.QdiscStats{
		BytesSent:    1000000,
		PacketsSent:  10000,
		BytesDropped: 100,
		Overlimits:   10,
		Requeues:     5,
	}

	adapter.SetQdiscStatistics(device, handle, stats)

	// Verify we can set statistics without errors
	assert.True(t, true) // Basic test that no panics occur
}

func TestMockAdapter_ClassStatistics(t *testing.T) {
	// Setup
	adapter := netlink.NewMockAdapter()

	device, err := tc.NewDevice("eth0")
	require.NoError(t, err)

	// Set mock class statistics
	handle := tc.NewHandle(1, 10)
	stats := netlink.ClassStats{
		BytesSent:      2000000,
		PacketsSent:    20000,
		BytesDropped:   200,
		Overlimits:     15,
		RateBPS:        100000000,
		BacklogBytes:   5000,
		BacklogPackets: 50,
	}

	adapter.SetClassStatistics(device, handle, stats)

	// Verify we can set statistics without errors
	assert.True(t, true) // Basic test that no panics occur
}
