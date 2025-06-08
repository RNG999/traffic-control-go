package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/pkg/tc"
)

func TestPortMatch_Creation(t *testing.T) {
	t.Run("Destination Port Match", func(t *testing.T) {
		match := NewPortDestinationMatch(5201)
		
		assert.Equal(t, MatchTypePortDestination, match.Type())
		assert.Equal(t, uint16(5201), match.Port())
		assert.Equal(t, "ip dport 5201 0xffff", match.String())
	})

	t.Run("Source Port Match", func(t *testing.T) {
		match := NewPortSourceMatch(8080)
		
		assert.Equal(t, MatchTypePortSource, match.Type())
		assert.Equal(t, uint16(8080), match.Port())
		assert.Equal(t, "ip sport 8080 0xffff", match.String())
	})
}

func TestIPMatch_Creation(t *testing.T) {
	t.Run("Valid Source IP CIDR", func(t *testing.T) {
		match, err := NewIPSourceMatch("192.168.1.0/24")
		require.NoError(t, err)
		
		assert.Equal(t, MatchTypeIPSource, match.Type())
		assert.Contains(t, match.String(), "192.168.1.0/24")
	})

	t.Run("Valid Destination IP Single", func(t *testing.T) {
		match, err := NewIPDestinationMatch("10.0.0.1")
		require.NoError(t, err)
		
		assert.Equal(t, MatchTypeIPDestination, match.Type())
		assert.Contains(t, match.String(), "10.0.0.1/32")
	})

	t.Run("Invalid IP Format", func(t *testing.T) {
		_, err := NewIPSourceMatch("invalid-ip")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid IP or CIDR")
	})
}

func TestFilter_MatchManagement(t *testing.T) {
	t.Run("Add Multiple Matches", func(t *testing.T) {
		device, _ := tc.NewDeviceName("eth0")
		parent, _ := tc.ParseHandle("1:0")
		handle, _ := tc.ParseHandle("800:1")
		
		filter := NewFilter(device, parent, 100, handle)
		flowID, _ := tc.ParseHandle("1:10")
		filter.SetFlowID(flowID)
		
		// Add port match
		portMatch := NewPortDestinationMatch(443)
		filter.AddMatch(portMatch)
		
		// Add IP match
		ipMatch, err := NewIPDestinationMatch("192.168.1.100")
		require.NoError(t, err)
		filter.AddMatch(ipMatch)
		
		// Verify both matches
		matches := filter.Matches()
		assert.Len(t, matches, 2)
		
		// Check types
		var hasPortMatch, hasIPMatch bool
		for _, match := range matches {
			switch match.Type() {
			case MatchTypePortDestination:
				hasPortMatch = true
				portMatch := match.(*PortMatch)
				assert.Equal(t, uint16(443), portMatch.Port())
			case MatchTypeIPDestination:
				hasIPMatch = true
			}
		}
		assert.True(t, hasPortMatch)
		assert.True(t, hasIPMatch)
	})

	t.Run("Filter Properties", func(t *testing.T) {
		device, _ := tc.NewDeviceName("test-dev")
		parent, _ := tc.ParseHandle("1:0")
		handle, _ := tc.ParseHandle("800:5")
		
		filter := NewFilter(device, parent, 200, handle)
		
		// Test ID properties
		assert.Equal(t, device, filter.ID().Device())
		assert.Equal(t, parent, filter.ID().Parent())
		assert.Equal(t, uint16(200), filter.ID().Priority())
		assert.Equal(t, handle, filter.ID().Handle())
		
		// Test flow ID
		flowID, _ := tc.ParseHandle("1:20")
		filter.SetFlowID(flowID)
		assert.Equal(t, flowID, filter.FlowID())
		
		// Test protocol
		filter.SetProtocol(ProtocolIP)
		assert.Equal(t, ProtocolIP, filter.Protocol())
	})
}

func TestFilterID_String(t *testing.T) {
	device, _ := tc.NewDeviceName("eth0")
	parent, _ := tc.ParseHandle("1:0")
	handle, _ := tc.ParseHandle("800:100")
	
	id := NewFilterID(device, parent, 150, handle)
	
	expected := "eth0:1::prio150:800:100"
	assert.Equal(t, expected, id.String())
}

func TestMatchTypes_EdgeCases(t *testing.T) {
	t.Run("Zero Port Number", func(t *testing.T) {
		match := NewPortDestinationMatch(0)
		assert.Equal(t, uint16(0), match.Port())
		assert.Equal(t, "ip dport 0 0xffff", match.String())
	})

	t.Run("Maximum Port Number", func(t *testing.T) {
		match := NewPortSourceMatch(65535)
		assert.Equal(t, uint16(65535), match.Port())
		assert.Contains(t, match.String(), "65535")
	})

	t.Run("IPv6 Address", func(t *testing.T) {
		match, err := NewIPSourceMatch("2001:db8::1")
		require.NoError(t, err)
		assert.Contains(t, match.String(), "2001:db8::1")
	})

	t.Run("Empty CIDR", func(t *testing.T) {
		_, err := NewIPDestinationMatch("")
		assert.Error(t, err)
	})
}

func TestProtocolMatch_Creation(t *testing.T) {
	t.Run("TCP Protocol", func(t *testing.T) {
		match := NewProtocolMatch(TransportProtocolTCP)
		assert.Equal(t, MatchTypeProtocol, match.Type())
		assert.Contains(t, match.String(), "6") // TCP is protocol 6
	})

	t.Run("UDP Protocol", func(t *testing.T) {
		match := NewProtocolMatch(TransportProtocolUDP)
		assert.Equal(t, MatchTypeProtocol, match.Type())
		assert.Contains(t, match.String(), "17") // UDP is protocol 17
	})
}