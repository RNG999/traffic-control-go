package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/events"
)

func TestConvertMatchData(t *testing.T) {
	t.Run("Port Destination Match", func(t *testing.T) {
		matchData := events.MatchData{
			Type:  entities.MatchTypePortDestination,
			Value: "ip dport 5201 0xffff",
		}

		match, err := convertMatchData(matchData)
		require.NoError(t, err)

		portMatch, ok := match.(*entities.PortMatch)
		require.True(t, ok, "Should be a PortMatch")
		assert.Equal(t, entities.MatchTypePortDestination, portMatch.Type())
		assert.Equal(t, uint16(5201), portMatch.Port())
	})

	t.Run("Port Source Match", func(t *testing.T) {
		matchData := events.MatchData{
			Type:  entities.MatchTypePortSource,
			Value: "ip sport 8080 0xffff",
		}

		match, err := convertMatchData(matchData)
		require.NoError(t, err)

		portMatch, ok := match.(*entities.PortMatch)
		require.True(t, ok, "Should be a PortMatch")
		assert.Equal(t, entities.MatchTypePortSource, portMatch.Type())
		assert.Equal(t, uint16(8080), portMatch.Port())
	})

	t.Run("IP Destination Match", func(t *testing.T) {
		matchData := events.MatchData{
			Type:  entities.MatchTypeIPDestination,
			Value: "ip dst 192.168.1.100/32",
		}

		match, err := convertMatchData(matchData)
		require.NoError(t, err)

		ipMatch, ok := match.(*entities.IPMatch)
		require.True(t, ok, "Should be an IPMatch")
		assert.Equal(t, entities.MatchTypeIPDestination, ipMatch.Type())
		assert.Contains(t, ipMatch.String(), "192.168.1.100/32")
	})

	t.Run("IP Source Match", func(t *testing.T) {
		matchData := events.MatchData{
			Type:  entities.MatchTypeIPSource,
			Value: "ip src 10.0.0.0/24",
		}

		match, err := convertMatchData(matchData)
		require.NoError(t, err)

		ipMatch, ok := match.(*entities.IPMatch)
		require.True(t, ok, "Should be an IPMatch")
		assert.Equal(t, entities.MatchTypeIPSource, ipMatch.Type())
		assert.Contains(t, ipMatch.String(), "10.0.0.0/24")
	})

	t.Run("Protocol Match", func(t *testing.T) {
		matchData := events.MatchData{
			Type:  entities.MatchTypeProtocol,
			Value: "ip protocol 6 0xff",
		}

		match, err := convertMatchData(matchData)
		require.NoError(t, err)

		protocolMatch, ok := match.(*entities.ProtocolMatch)
		require.True(t, ok, "Should be a ProtocolMatch")
		assert.Equal(t, entities.MatchTypeProtocol, protocolMatch.Type())
		assert.Equal(t, entities.TransportProtocolTCP, protocolMatch.Protocol())
	})

	t.Run("Invalid Port Format", func(t *testing.T) {
		matchData := events.MatchData{
			Type:  entities.MatchTypePortDestination,
			Value: "invalid port format",
		}

		_, err := convertMatchData(matchData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid destination port match value")
	})

	t.Run("Invalid IP Format", func(t *testing.T) {
		matchData := events.MatchData{
			Type:  entities.MatchTypeIPDestination,
			Value: "invalid ip format",
		}

		_, err := convertMatchData(matchData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid destination IP match value")
	})

	t.Run("Unsupported Match Type", func(t *testing.T) {
		matchData := events.MatchData{
			Type:  999, // Invalid type
			Value: "test",
		}

		_, err := convertMatchData(matchData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported match type")
	})
}

func TestParseIPFromString(t *testing.T) {
	t.Run("Valid Destination IP", func(t *testing.T) {
		cidr, err := parseIPFromString("ip dst 192.168.1.100/32", "ip dst")
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.100/32", cidr)
	})

	t.Run("Valid Source IP", func(t *testing.T) {
		cidr, err := parseIPFromString("ip src 10.0.0.0/24", "ip src")
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.0/24", cidr)
	})

	t.Run("IPv6 Address", func(t *testing.T) {
		cidr, err := parseIPFromString("ip dst 2001:db8::1/128", "ip dst")
		require.NoError(t, err)
		assert.Equal(t, "2001:db8::1/128", cidr)
	})

	t.Run("Invalid Format", func(t *testing.T) {
		_, err := parseIPFromString("invalid format", "ip dst")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid IP match format")
	})

	t.Run("Wrong Prefix", func(t *testing.T) {
		_, err := parseIPFromString("ip src 10.0.0.0/24", "ip dst")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid IP match format")
	})
}

func TestParsePortFromString(t *testing.T) {
	t.Run("Valid Destination Port", func(t *testing.T) {
		port, err := parsePortFromString("ip dport 5201 0xffff")
		require.NoError(t, err)
		assert.Equal(t, uint16(5201), port)
	})

	t.Run("Valid Source Port", func(t *testing.T) {
		port, err := parsePortFromString("ip sport 8080 0xffff")
		require.NoError(t, err)
		assert.Equal(t, uint16(8080), port)
	})

	t.Run("Port 0", func(t *testing.T) {
		port, err := parsePortFromString("ip dport 0 0xffff")
		require.NoError(t, err)
		assert.Equal(t, uint16(0), port)
	})

	t.Run("Port 65535", func(t *testing.T) {
		port, err := parsePortFromString("ip sport 65535 0xffff")
		require.NoError(t, err)
		assert.Equal(t, uint16(65535), port)
	})

	t.Run("Invalid Format", func(t *testing.T) {
		_, err := parsePortFromString("invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid port match format")
	})

	t.Run("Invalid Port Number", func(t *testing.T) {
		_, err := parsePortFromString("ip dport invalid 0xffff")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid port number")
	})
}