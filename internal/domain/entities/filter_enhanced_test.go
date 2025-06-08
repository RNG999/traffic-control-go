package entities

import (
	"fmt"
	"testing"

	"github.com/rng999/traffic-control-go/pkg/tc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdvancedFilter_Creation(t *testing.T) {
	device, _ := tc.NewDeviceName("eth0")
	parent := tc.NewHandle(1, 0)
	handle := tc.NewHandle(1, 10)

	filter := NewAdvancedFilter(device, parent, 100, handle)

	assert.NotNil(t, filter)
	assert.Equal(t, ActionClassify, filter.Action())
	assert.Equal(t, uint8(0), filter.QoSPriority())
	assert.Nil(t, filter.RateLimit())
}

func TestAdvancedFilter_SetQoSPriority(t *testing.T) {
	device, _ := tc.NewDeviceName("eth0")
	parent := tc.NewHandle(1, 0)
	handle := tc.NewHandle(1, 10)

	filter := NewAdvancedFilter(device, parent, 100, handle)
	filter.SetQoSPriority(5)

	assert.Equal(t, uint8(5), filter.QoSPriority())
}

func TestAdvancedFilter_SetRateLimit(t *testing.T) {
	device, _ := tc.NewDeviceName("eth0")
	parent := tc.NewHandle(1, 0)
	handle := tc.NewHandle(1, 10)

	filter := NewAdvancedFilter(device, parent, 100, handle)

	bandwidth, err := tc.NewBandwidth("10Mbps")
	require.NoError(t, err)

	filter.SetRateLimit(bandwidth, 1500)

	assert.NotNil(t, filter.RateLimit())
	assert.Equal(t, uint32(1500), filter.BurstLimit())
}

func TestAdvancedFilter_SetAction(t *testing.T) {
	device, _ := tc.NewDeviceName("eth0")
	parent := tc.NewHandle(1, 0)
	handle := tc.NewHandle(1, 10)

	filter := NewAdvancedFilter(device, parent, 100, handle)

	tests := []FilterAction{
		ActionClassify,
		ActionDrop,
		ActionRateLimit,
		ActionMark,
	}

	for _, action := range tests {
		filter.SetAction(action)
		assert.Equal(t, action, filter.Action())
	}
}

func TestPortRangeMatch(t *testing.T) {
	tests := []struct {
		name      string
		startPort uint16
		endPort   uint16
		isSource  bool
	}{
		{"SourcePortRange", 8000, 8080, true},
		{"DestinationPortRange", 80, 443, false},
		{"SinglePortAsRange", 22, 22, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var match *PortRangeMatch
			if tt.isSource {
				match = NewPortSourceRangeMatch(tt.startPort, tt.endPort)
			} else {
				match = NewPortDestinationRangeMatch(tt.startPort, tt.endPort)
			}

			assert.Equal(t, MatchTypePortRange, match.Type())
			assert.Equal(t, tt.startPort, match.StartPort())
			assert.Equal(t, tt.endPort, match.EndPort())
			assert.Contains(t, match.String(), "port range")
		})
	}
}

func TestTOSMatch(t *testing.T) {
	tests := []struct {
		name string
		tos  uint8
	}{
		{"MinimumDelay", 0x10},
		{"MaximumThroughput", 0x08},
		{"MaximumReliability", 0x04},
		{"MinimumCost", 0x02},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := NewTOSMatch(tt.tos)

			assert.Equal(t, MatchTypeTOS, match.Type())
			assert.Equal(t, tt.tos, match.TOS())
			assert.Contains(t, match.String(), "ip tos")
		})
	}
}

func TestDSCPMatch(t *testing.T) {
	tests := []struct {
		name string
		dscp uint8
	}{
		{"Default", 0},
		{"ClassSelector1", 8},
		{"AssuredForwarding", 10},
		{"ExpeditedForwarding", 46},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := NewDSCPMatch(tt.dscp)

			assert.Equal(t, MatchTypeDSCP, match.Type())
			assert.Equal(t, tt.dscp, match.DSCP())
			assert.Contains(t, match.String(), "ip tos")
			assert.Contains(t, match.String(), "0xfc") // DSCP mask
		})
	}
}

func TestFlowIDMatch(t *testing.T) {
	keys := []string{"src", "dst", "proto", "sport", "dport"}
	mask := uint32(0xFFFF)

	match := NewFlowIDMatch(keys, mask)

	assert.Equal(t, MatchTypeFlowID, match.Type())
	assert.Equal(t, keys, match.Keys())
	assert.Equal(t, mask, match.Mask())
	assert.Contains(t, match.String(), "flow keys")
}

func TestFilter_AddProtocolMatch(t *testing.T) {
	device, _ := tc.NewDeviceName("eth0")
	parent := tc.NewHandle(1, 0)
	handle := tc.NewHandle(1, 10)

	filter := NewFilter(device, parent, 100, handle)

	tests := []TransportProtocol{
		TransportProtocolTCP,
		TransportProtocolUDP,
		TransportProtocolICMP,
	}

	for _, protocol := range tests {
		filter.AddProtocolMatch(protocol)
	}

	matches := filter.Matches()
	assert.Len(t, matches, 3)

	for i, match := range matches {
		assert.Equal(t, MatchTypeProtocol, match.Type())
		protocolMatch := match.(*ProtocolMatch)
		assert.Equal(t, tests[i], protocolMatch.Protocol())
	}
}

func TestFilter_AddPortRangeMatch(t *testing.T) {
	device, _ := tc.NewDeviceName("eth0")
	parent := tc.NewHandle(1, 0)
	handle := tc.NewHandle(1, 10)

	filter := NewFilter(device, parent, 100, handle)

	// Add source port range
	filter.AddPortRangeMatch(8000, 8080, true)

	// Add destination port range
	filter.AddPortRangeMatch(80, 443, false)

	matches := filter.Matches()
	assert.Len(t, matches, 2)

	// Verify source port range
	sourceMatch := matches[0].(*PortRangeMatch)
	assert.Equal(t, MatchTypePortRange, sourceMatch.Type())
	assert.Equal(t, uint16(8000), sourceMatch.StartPort())
	assert.Equal(t, uint16(8080), sourceMatch.EndPort())

	// Verify destination port range
	destMatch := matches[1].(*PortRangeMatch)
	assert.Equal(t, MatchTypePortRange, destMatch.Type())
	assert.Equal(t, uint16(80), destMatch.StartPort())
	assert.Equal(t, uint16(443), destMatch.EndPort())
}

func TestFilter_AddTOSMatch(t *testing.T) {
	device, _ := tc.NewDeviceName("eth0")
	parent := tc.NewHandle(1, 0)
	handle := tc.NewHandle(1, 10)

	filter := NewFilter(device, parent, 100, handle)
	filter.AddTOSMatch(0x10) // Minimum delay

	matches := filter.Matches()
	assert.Len(t, matches, 1)

	tosMatch := matches[0].(*TOSMatch)
	assert.Equal(t, MatchTypeTOS, tosMatch.Type())
	assert.Equal(t, uint8(0x10), tosMatch.TOS())
}

func TestFilter_AddDSCPMatch(t *testing.T) {
	device, _ := tc.NewDeviceName("eth0")
	parent := tc.NewHandle(1, 0)
	handle := tc.NewHandle(1, 10)

	filter := NewFilter(device, parent, 100, handle)
	filter.AddDSCPMatch(46) // Expedited forwarding

	matches := filter.Matches()
	assert.Len(t, matches, 1)

	dscpMatch := matches[0].(*DSCPMatch)
	assert.Equal(t, MatchTypeDSCP, dscpMatch.Type())
	assert.Equal(t, uint8(46), dscpMatch.DSCP())
}

func TestFilter_AddIPRangeMatch(t *testing.T) {
	device, _ := tc.NewDeviceName("eth0")
	parent := tc.NewHandle(1, 0)
	handle := tc.NewHandle(1, 10)

	filter := NewFilter(device, parent, 100, handle)

	tests := []struct {
		name     string
		startIP  string
		endIP    string
		isSource bool
		wantErr  bool
	}{
		{"ValidSourceRange", "192.168.1.1", "192.168.1.100", true, false},
		{"ValidDestRange", "10.0.0.1", "10.0.0.254", false, false},
		{"InvalidStartIP", "invalid", "192.168.1.100", true, true},
		{"InvalidEndIP", "192.168.1.1", "invalid", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := filter.AddIPRangeMatch(tt.startIP, tt.endIP, tt.isSource)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFilter_ValidateMatches(t *testing.T) {
	device, _ := tc.NewDeviceName("eth0")
	parent := tc.NewHandle(1, 0)
	handle := tc.NewHandle(1, 10)

	tests := []struct {
		name    string
		setup   func(f *Filter)
		wantErr bool
	}{
		{
			name: "ValidSingleProtocol",
			setup: func(f *Filter) {
				f.AddProtocolMatch(TransportProtocolTCP)
			},
			wantErr: false,
		},
		{
			name: "InvalidMultipleProtocols",
			setup: func(f *Filter) {
				f.AddProtocolMatch(TransportProtocolTCP)
				f.AddProtocolMatch(TransportProtocolUDP)
			},
			wantErr: true,
		},
		{
			name: "ValidMixedMatches",
			setup: func(f *Filter) {
				f.AddProtocolMatch(TransportProtocolTCP)
				f.AddPortRangeMatch(80, 443, false)
				f.AddTOSMatch(0x10)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewFilter(device, parent, 100, handle)
			tt.setup(filter)

			err := filter.ValidateMatches()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDSCPMatch_TOSCalculation(t *testing.T) {
	// Test that DSCP values are correctly converted to TOS field
	tests := []struct {
		dscp     uint8
		expected string
	}{
		{0, "0x0"},   // DSCP 0 -> TOS 0x00
		{10, "0x28"}, // DSCP 10 -> TOS 0x28 (10 << 2)
		{46, "0xb8"}, // DSCP 46 -> TOS 0xB8 (46 << 2)
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("DSCP_%d", tt.dscp), func(t *testing.T) {
			match := NewDSCPMatch(tt.dscp)
			tosStr := match.String()

			assert.Contains(t, tosStr, tt.expected)
			assert.Contains(t, tosStr, "0xfc") // DSCP mask
		})
	}
}

func TestAdvancedFilter_ComplexScenario(t *testing.T) {
	// Test a complex filter scenario with multiple match types
	device, _ := tc.NewDeviceName("eth0")
	parent := tc.NewHandle(1, 0)
	handle := tc.NewHandle(1, 10)

	filter := NewAdvancedFilter(device, parent, 100, handle)

	// Set up complex filtering scenario
	filter.SetQoSPriority(3)
	filter.SetAction(ActionRateLimit)

	bandwidth, err := tc.NewBandwidth("5Mbps")
	require.NoError(t, err)
	filter.SetRateLimit(bandwidth, 7500) // 1.5 * MTU burst

	// Add multiple match criteria
	filter.AddProtocolMatch(TransportProtocolTCP)
	filter.AddPortRangeMatch(8000, 8080, false) // Destination ports 8000-8080
	filter.AddDSCPMatch(26)                     // AF31 (Assured Forwarding)

	err = filter.AddIPRangeMatch("192.168.1.0", "192.168.1.255", true)
	require.NoError(t, err)

	// Validate the configuration
	assert.Equal(t, uint8(3), filter.QoSPriority())
	assert.Equal(t, ActionRateLimit, filter.Action())
	assert.NotNil(t, filter.RateLimit())
	assert.Equal(t, uint32(7500), filter.BurstLimit())

	matches := filter.Matches()
	assert.Len(t, matches, 4) // Protocol, Port Range, DSCP, IP Range

	// Validate that the filter configuration is valid
	err = filter.ValidateMatches()
	assert.NoError(t, err)
}
