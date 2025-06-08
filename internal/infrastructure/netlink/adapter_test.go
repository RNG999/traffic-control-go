//go:build linux
// +build linux

package netlink

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"

	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/pkg/logging"
)

func TestConfigureU32Matches_PortFiltering(t *testing.T) {
	// Create adapter with test logger
	logger, err := logging.NewLogger(logging.DevelopmentConfig())
	require.NoError(t, err)
	adapter := &RealNetlinkAdapter{
		logger: logger,
	}

	t.Run("Destination Port Match", func(t *testing.T) {
		// Create U32 filter
		filter := &netlink.U32{
			FilterAttrs: netlink.FilterAttrs{
				LinkIndex: 1,
				Parent:    netlink.MakeHandle(1, 0),
				Priority:  100,
				Protocol:  0x0800, // ETH_P_IP
			},
			ClassId: netlink.MakeHandle(1, 10),
		}

		// Create port match
		portMatch := entities.NewPortDestinationMatch(5201)
		matches := []entities.Match{portMatch}

		// Configure matches
		err := adapter.configureU32Matches(filter, matches)
		require.NoError(t, err)

		// Verify U32 selector was configured
		require.NotNil(t, filter.Sel, "U32 selector should be configured")
		assert.Equal(t, uint8(1), filter.Sel.Nkeys, "Should have one key")

		// Verify key configuration
		key := filter.Sel.Keys[0]
		assert.Equal(t, uint32(0x0000ffff), key.Mask, "Should match 2 bytes (port)")
		assert.Equal(t, uint32(5201), key.Val, "Should match port 5201")
		assert.Equal(t, int32(22), key.Off, "Should be at offset 22 (dest port)")
	})

	t.Run("Source Port Match", func(t *testing.T) {
		filter := &netlink.U32{
			FilterAttrs: netlink.FilterAttrs{
				LinkIndex: 1,
				Parent:    netlink.MakeHandle(1, 0),
				Priority:  100,
				Protocol:  0x0800,
			},
			ClassId: netlink.MakeHandle(1, 10),
		}

		portMatch := entities.NewPortSourceMatch(8080)
		matches := []entities.Match{portMatch}

		err := adapter.configureU32Matches(filter, matches)
		require.NoError(t, err)

		require.NotNil(t, filter.Sel)
		assert.Equal(t, uint8(1), filter.Sel.Nkeys)

		key := filter.Sel.Keys[0]
		assert.Equal(t, uint32(0xffff0000), key.Mask, "Should match high 2 bytes")
		assert.Equal(t, uint32(8080<<16), key.Val, "Should match port 8080 shifted")
		assert.Equal(t, int32(20), key.Off, "Should be at offset 20 (source port)")
	})

	t.Run("No Matches - Match All", func(t *testing.T) {
		filter := &netlink.U32{
			FilterAttrs: netlink.FilterAttrs{
				LinkIndex: 1,
				Parent:    netlink.MakeHandle(1, 0),
				Priority:  100,
				Protocol:  0x0800,
			},
			ClassId: netlink.MakeHandle(1, 10),
		}

		// No matches - should remain match-all
		matches := []entities.Match{}

		err := adapter.configureU32Matches(filter, matches)
		require.NoError(t, err)

		// Should remain nil (match-all behavior)
		assert.Nil(t, filter.Sel, "Should remain match-all filter")
	})

	t.Run("Unsupported Match Type", func(t *testing.T) {
		filter := &netlink.U32{
			FilterAttrs: netlink.FilterAttrs{
				LinkIndex: 1,
				Parent:    netlink.MakeHandle(1, 0),
				Priority:  100,
				Protocol:  0x0800,
			},
			ClassId: netlink.MakeHandle(1, 10),
		}

		// Create IP match (not yet implemented)
		ipMatch, err := entities.NewIPDestinationMatch("192.168.1.100")
		require.NoError(t, err)
		matches := []entities.Match{ipMatch}

		// Should not fail, just skip unsupported matches
		err = adapter.configureU32Matches(filter, matches)
		require.NoError(t, err)

		// Should remain nil since IP matching is not implemented yet
		assert.Nil(t, filter.Sel, "Should skip unsupported IP match")
	})

	t.Run("Multiple Port Matches", func(t *testing.T) {
		filter := &netlink.U32{
			FilterAttrs: netlink.FilterAttrs{
				LinkIndex: 1,
				Parent:    netlink.MakeHandle(1, 0),
				Priority:  100,
				Protocol:  0x0800,
			},
			ClassId: netlink.MakeHandle(1, 10),
		}

		// Create both source and destination port matches
		// Note: In real U32 filters, only one can be configured per filter
		// The last one will overwrite the previous
		srcMatch := entities.NewPortSourceMatch(8080)
		dstMatch := entities.NewPortDestinationMatch(5201)
		matches := []entities.Match{srcMatch, dstMatch}

		err := adapter.configureU32Matches(filter, matches)
		require.NoError(t, err)

		require.NotNil(t, filter.Sel)
		// Last match (destination) should win
		key := filter.Sel.Keys[0]
		assert.Equal(t, uint32(5201), key.Val, "Destination port should be configured")
		assert.Equal(t, int32(22), key.Off, "Should be destination port offset")
	})
}

func TestU32FilterConstruction(t *testing.T) {
	t.Run("Basic Filter Properties", func(t *testing.T) {
		filter := &netlink.U32{
			FilterAttrs: netlink.FilterAttrs{
				LinkIndex: 2,
				Parent:    netlink.MakeHandle(1, 0),
				Priority:  150,
				Protocol:  0x0800,
			},
			ClassId: netlink.MakeHandle(1, 20),
		}

		assert.Equal(t, 2, filter.LinkIndex)
		assert.Equal(t, uint32(0x10000), filter.Parent) // 1:0
		assert.Equal(t, uint16(150), filter.Priority)
		assert.Equal(t, uint16(0x0800), filter.Protocol)
		assert.Equal(t, uint32(0x10014), filter.ClassId) // 1:20
	})
}

func TestPortMatchValues(t *testing.T) {
	testCases := []struct {
		port     uint16
		expected struct {
			dstMask uint32
			dstVal  uint32
			srcMask uint32
			srcVal  uint32
		}
	}{
		{
			port: 80,
			expected: struct {
				dstMask uint32
				dstVal  uint32
				srcMask uint32
				srcVal  uint32
			}{
				dstMask: 0x0000ffff,
				dstVal:  80,
				srcMask: 0xffff0000,
				srcVal:  80 << 16,
			},
		},
		{
			port: 443,
			expected: struct {
				dstMask uint32
				dstVal  uint32
				srcMask uint32
				srcVal  uint32
			}{
				dstMask: 0x0000ffff,
				dstVal:  443,
				srcMask: 0xffff0000,
				srcVal:  443 << 16,
			},
		},
		{
			port: 65535, // Maximum port
			expected: struct {
				dstMask uint32
				dstVal  uint32
				srcMask uint32
				srcVal  uint32
			}{
				dstMask: 0x0000ffff,
				dstVal:  65535,
				srcMask: 0xffff0000,
				srcVal:  65535 << 16,
			},
		},
	}

	logger, err := logging.NewLogger(logging.DevelopmentConfig())
	require.NoError(t, err)
	adapter := &RealNetlinkAdapter{
		logger: logger,
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Port_%d", tc.port), func(t *testing.T) {
			// Test destination port
			filter := &netlink.U32{
				FilterAttrs: netlink.FilterAttrs{
					LinkIndex: 1,
					Parent:    netlink.MakeHandle(1, 0),
					Priority:  100,
					Protocol:  0x0800,
				},
				ClassId: netlink.MakeHandle(1, 10),
			}

			dstMatch := entities.NewPortDestinationMatch(tc.port)
			err := adapter.configureU32Matches(filter, []entities.Match{dstMatch})
			require.NoError(t, err)

			require.NotNil(t, filter.Sel)
			key := filter.Sel.Keys[0]
			assert.Equal(t, tc.expected.dstMask, key.Mask, "Destination mask mismatch")
			assert.Equal(t, tc.expected.dstVal, key.Val, "Destination value mismatch")

			// Test source port
			filter.Sel = nil // Reset
			srcMatch := entities.NewPortSourceMatch(tc.port)
			err = adapter.configureU32Matches(filter, []entities.Match{srcMatch})
			require.NoError(t, err)

			require.NotNil(t, filter.Sel)
			key = filter.Sel.Keys[0]
			assert.Equal(t, tc.expected.srcMask, key.Mask, "Source mask mismatch")
			assert.Equal(t, tc.expected.srcVal, key.Val, "Source value mismatch")
		})
	}
}
