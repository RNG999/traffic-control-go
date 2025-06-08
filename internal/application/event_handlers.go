package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/events"
	"github.com/rng999/traffic-control-go/pkg/logging"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// handleQdiscCreated handles QdiscCreated events and applies them to netlink
func (s *TrafficControlService) handleQdiscCreated(ctx context.Context, event interface{}) error {
	// Type assert to the event types we expect
	var device tc.DeviceName
	var handle tc.Handle
	var qdiscType entities.QdiscType
	var defaultClass string

	switch e := event.(type) {
	case *events.QdiscCreatedEvent:
		device = e.DeviceName
		handle = e.Handle
		qdiscType = e.QdiscType
	case *events.HTBQdiscCreatedEvent:
		device = e.DeviceName
		handle = e.Handle
		qdiscType = entities.QdiscTypeHTB
		defaultClass = e.DefaultClass.String()
	default:
		// Not a qdisc event we handle
		return nil
	}

	s.logger.Info("Applying qdisc to netlink",
		logging.String("device", device.String()),
		logging.String("handle", handle.String()),
		logging.Int("type", int(qdiscType)),
	)

	// Apply to netlink based on type
	switch qdiscType {
	case entities.QdiscTypeHTB:
		// Parse default class handle
		defaultHandle, err := tc.ParseHandle(defaultClass)
		if err != nil {
			return fmt.Errorf("invalid default class handle: %w", err)
		}
		qdisc := entities.NewHTBQdisc(device, handle, defaultHandle)
		return s.netlinkAdapter.AddQdisc(ctx, qdisc.Qdisc)
	case entities.QdiscTypeTBF:
		// TBF needs rate from event - skip for now
		s.logger.Warn("TBF qdisc netlink application not implemented")
		return nil
	default:
		s.logger.Warn("Unsupported qdisc type for netlink",
			logging.Int("type", int(qdiscType)),
		)
		return nil
	}
}

// handleClassCreated handles ClassCreated events and applies them to netlink
func (s *TrafficControlService) handleClassCreated(ctx context.Context, event interface{}) error {
	switch e := event.(type) {
	case *events.ClassCreatedEvent:
		// Basic class creation - not HTB specific
		s.logger.Info("Basic class created event - netlink application not implemented",
			logging.String("device", e.DeviceName.String()),
			logging.String("handle", e.Handle.String()),
		)
		return nil

	case *events.HTBClassCreatedEvent:
		s.logger.Info("Applying HTB class to netlink",
			logging.String("device", e.DeviceName.String()),
			logging.String("parent", e.Parent.String()),
			logging.String("handle", e.Handle.String()),
			logging.String("rate", e.Rate.String()),
			logging.String("ceil", e.Ceil.String()),
		)

		// Create HTB class with proper parameters
		// Note: HTB events don't have priority in the event, using default priority
		priority := entities.Priority(4) // Normal priority
		class := entities.NewHTBClass(e.DeviceName, e.Handle, e.Parent, e.Name, priority)

		// Set rate and ceil bandwidth
		class.SetRate(e.Rate)

		// If ceil is 0 or not set, default to rate (HTB requirement)
		if e.Ceil.BitsPerSecond() == 0 {
			class.SetCeil(e.Rate)
		} else {
			class.SetCeil(e.Ceil)
		}

		// Set burst values if provided, otherwise calculate them
		if e.Burst > 0 {
			class.SetBurst(e.Burst)
		} else {
			class.SetBurst(class.CalculateBurst())
		}

		if e.Cburst > 0 {
			class.SetCburst(e.Cburst)
		} else {
			class.SetCburst(class.CalculateCburst())
		}

		return s.netlinkAdapter.AddClass(ctx, class)

	default:
		// Not a class event we handle
		return nil
	}
}

// handleFilterCreated handles FilterCreated events and applies them to netlink
func (s *TrafficControlService) handleFilterCreated(ctx context.Context, event interface{}) error {
	e, ok := event.(*events.FilterCreatedEvent)
	if !ok {
		// Not a filter event
		return nil
	}

	s.logger.Info("Applying filter to netlink",
		logging.String("device", e.DeviceName.String()),
		logging.String("parent", e.Parent.String()),
		logging.String("handle", e.Handle.String()),
		logging.String("flow_id", e.FlowID.String()),
		logging.Int("priority", int(e.Priority)),
	)

	// Create filter entity directly from event data (no string parsing needed)
	filter := entities.NewFilter(e.DeviceName, e.Parent, e.Priority, e.Handle)

	// Set flow ID
	filter.SetFlowID(e.FlowID)

	// Set protocol
	filter.SetProtocol(e.Protocol)

	// Add matches from event data
	for _, matchData := range e.Matches {
		s.logger.Debug("Filter match",
			logging.Int("type", int(matchData.Type)),
			logging.String("value", matchData.Value),
		)

		// Convert match data back to proper match objects
		match, err := convertMatchData(matchData)
		if err != nil {
			s.logger.Error("Failed to convert match data",
				logging.Error(err),
				logging.Int("type", int(matchData.Type)),
				logging.String("value", matchData.Value),
			)
			continue
		}

		s.logger.Debug("Successfully converted and added match",
			logging.Int("type", int(matchData.Type)),
			logging.String("value", matchData.Value),
			logging.String("match_string", match.String()),
		)

		filter.AddMatch(match)
	}

	s.logger.Info("Adding filter via netlink adapter")
	return s.netlinkAdapter.AddFilter(ctx, filter)
}

// convertMatchData converts event match data back to entities.Match objects
func convertMatchData(matchData events.MatchData) (entities.Match, error) {
	switch matchData.Type {
	case entities.MatchTypeIPSource:
		// Parse IP from string representation
		ip, err := parseIPFromString(matchData.Value, "ip src")
		if err != nil {
			return nil, fmt.Errorf("invalid source IP match value: %w", err)
		}
		return entities.NewIPSourceMatch(ip)
	case entities.MatchTypeIPDestination:
		// Parse IP from string representation
		ip, err := parseIPFromString(matchData.Value, "ip dst")
		if err != nil {
			return nil, fmt.Errorf("invalid destination IP match value: %w", err)
		}
		return entities.NewIPDestinationMatch(ip)
	case entities.MatchTypePortSource:
		// Parse port number from string representation
		port, err := parsePortFromString(matchData.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid source port match value: %w", err)
		}
		return entities.NewPortSourceMatch(port), nil
	case entities.MatchTypePortDestination:
		// Parse port number from string representation
		port, err := parsePortFromString(matchData.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid destination port match value: %w", err)
		}
		return entities.NewPortDestinationMatch(port), nil
	case entities.MatchTypeProtocol:
		// Parse protocol from string representation
		protocol, err := parseProtocolFromString(matchData.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid protocol match value: %w", err)
		}
		return entities.NewProtocolMatch(protocol), nil
	case entities.MatchTypeMark:
		// Parse mark value from string representation
		mark, err := parseMarkFromString(matchData.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid mark match value: %w", err)
		}
		return entities.NewMarkMatch(mark), nil
	default:
		return nil, fmt.Errorf("unsupported match type: %v", matchData.Type)
	}
}

// parsePortFromString parses port number from the string representation
// Expected format: "ip dport 80 0xffff" or "ip sport 80 0xffff"
func parsePortFromString(value string) (uint16, error) {
	// Split the string by spaces: ["ip", "dport/sport", "80", "0xffff"]
	parts := strings.Fields(value)
	if len(parts) < 3 {
		return 0, fmt.Errorf("invalid port match format: %s", value)
	}

	// Parse the port number (third part)
	port, err := strconv.ParseUint(parts[2], 10, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid port number: %s", parts[2])
	}

	return uint16(port), nil
}

// parseProtocolFromString parses protocol from string representation
// Expected format: "ip protocol 6 0xff"
func parseProtocolFromString(value string) (entities.TransportProtocol, error) {
	parts := strings.Fields(value)
	if len(parts) < 3 {
		return 0, fmt.Errorf("invalid protocol match format: %s", value)
	}

	// Parse the protocol number (third part)
	protocol, err := strconv.ParseUint(parts[2], 10, 8)
	if err != nil {
		return 0, fmt.Errorf("invalid protocol number: %s", parts[2])
	}

	return entities.TransportProtocol(protocol), nil
}

// parseIPFromString parses IP/CIDR from string representation
// Expected format: "ip dst 192.168.1.100/32" or "ip src 10.0.0.0/24"
func parseIPFromString(value, expectedPrefix string) (string, error) {
	// Use fmt.Sscanf to extract the IP/CIDR part after the prefix
	var cidr string
	pattern := expectedPrefix + " %s"
	if _, err := fmt.Sscanf(value, pattern, &cidr); err != nil {
		return "", fmt.Errorf("invalid IP match format: %s (expected '%s CIDR')", value, expectedPrefix)
	}

	return cidr, nil
}

// parseMarkFromString parses mark value from string representation
// Expected format: "mark 0x123 0xffffffff"
func parseMarkFromString(value string) (uint32, error) {
	parts := strings.Fields(value)
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid mark match format: %s", value)
	}

	// Parse the mark value (second part)
	mark, err := strconv.ParseUint(parts[1], 0, 32) // 0 base allows 0x prefix
	if err != nil {
		return 0, fmt.Errorf("invalid mark value: %s", parts[1])
	}

	return uint32(mark), nil
}
