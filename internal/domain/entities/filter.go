package entities

import (
	"fmt"
	"net"

	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
)

// FilterID represents a unique identifier for a filter
type FilterID struct {
	device   valueobjects.DeviceName
	parent   valueobjects.Handle
	priority uint16
	handle   valueobjects.Handle
}

// NewFilterID creates a new FilterID
func NewFilterID(device valueobjects.DeviceName, parent valueobjects.Handle, priority uint16, handle valueobjects.Handle) FilterID {
	return FilterID{
		device:   device,
		parent:   parent,
		priority: priority,
		handle:   handle,
	}
}

// String returns the string representation of FilterID
func (id FilterID) String() string {
	return fmt.Sprintf("%s:%s:prio%d:%s", id.device, id.parent, id.priority, id.handle)
}

// Priority returns the filter priority
func (id FilterID) Priority() uint16 {
	return id.priority
}

// Handle returns the filter handle
func (id FilterID) Handle() valueobjects.Handle {
	return id.handle
}

// Filter represents a packet classification filter
type Filter struct {
	id       FilterID
	flowID   valueobjects.Handle // Target class
	protocol Protocol
	matches  []Match
}

// Protocol represents network protocol
type Protocol int

const (
	ProtocolAll Protocol = iota
	ProtocolIP
	ProtocolIPv6
)

// NewFilter creates a new Filter entity
func NewFilter(device valueobjects.DeviceName, parent valueobjects.Handle, priority uint16, handle valueobjects.Handle) *Filter {
	return &Filter{
		id:       NewFilterID(device, parent, priority, handle),
		protocol: ProtocolIP,
		matches:  make([]Match, 0),
	}
}

// ID returns the filter ID
func (f *Filter) ID() FilterID {
	return f.id
}

// SetFlowID sets the target class handle
func (f *Filter) SetFlowID(flowID valueobjects.Handle) {
	f.flowID = flowID
}

// FlowID returns the target class handle
func (f *Filter) FlowID() valueobjects.Handle {
	return f.flowID
}

// SetProtocol sets the protocol
func (f *Filter) SetProtocol(p Protocol) {
	f.protocol = p
}

// Protocol returns the protocol
func (f *Filter) Protocol() Protocol {
	return f.protocol
}

// AddMatch adds a match condition
func (f *Filter) AddMatch(match Match) {
	f.matches = append(f.matches, match)
}

// Matches returns all match conditions
func (f *Filter) Matches() []Match {
	return f.matches
}

// Match represents a filter matching condition
type Match interface {
	Type() MatchType
	String() string
}

// MatchType represents the type of match
type MatchType int

const (
	MatchTypeIPSource MatchType = iota
	MatchTypeIPDestination
	MatchTypePortSource
	MatchTypePortDestination
	MatchTypeProtocol
	MatchTypeMark
)

// IPMatch represents an IP address match
type IPMatch struct {
	matchType MatchType
	network   *net.IPNet
}

// NewIPSourceMatch creates a source IP match
func NewIPSourceMatch(cidr string) (*IPMatch, error) {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		// Try parsing as single IP
		ip := net.ParseIP(cidr)
		if ip == nil {
			return nil, fmt.Errorf("invalid IP or CIDR: %s", cidr)
		}
		// Convert to CIDR
		if ip.To4() != nil {
			network = &net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)}
		} else {
			network = &net.IPNet{IP: ip, Mask: net.CIDRMask(128, 128)}
		}
	}

	return &IPMatch{
		matchType: MatchTypeIPSource,
		network:   network,
	}, nil
}

// NewIPDestinationMatch creates a destination IP match
func NewIPDestinationMatch(cidr string) (*IPMatch, error) {
	match, err := NewIPSourceMatch(cidr)
	if err != nil {
		return nil, err
	}
	match.matchType = MatchTypeIPDestination
	return match, nil
}

// Type returns the match type
func (m *IPMatch) Type() MatchType {
	return m.matchType
}

// String returns the string representation
func (m *IPMatch) String() string {
	prefix := "src"
	if m.matchType == MatchTypeIPDestination {
		prefix = "dst"
	}
	return fmt.Sprintf("ip %s %s", prefix, m.network)
}

// Network returns the IP network
func (m *IPMatch) Network() *net.IPNet {
	return m.network
}

// PortMatch represents a port match
type PortMatch struct {
	matchType MatchType
	port      uint16
	mask      uint16
}

// NewPortSourceMatch creates a source port match
func NewPortSourceMatch(port uint16) *PortMatch {
	return &PortMatch{
		matchType: MatchTypePortSource,
		port:      port,
		mask:      0xFFFF,
	}
}

// NewPortDestinationMatch creates a destination port match
func NewPortDestinationMatch(port uint16) *PortMatch {
	return &PortMatch{
		matchType: MatchTypePortDestination,
		port:      port,
		mask:      0xFFFF,
	}
}

// Type returns the match type
func (m *PortMatch) Type() MatchType {
	return m.matchType
}

// String returns the string representation
func (m *PortMatch) String() string {
	prefix := "sport"
	if m.matchType == MatchTypePortDestination {
		prefix = "dport"
	}
	return fmt.Sprintf("ip %s %d 0x%x", prefix, m.port, m.mask)
}

// Port returns the port number
func (m *PortMatch) Port() uint16 {
	return m.port
}

// ProtocolMatch represents a protocol match
type ProtocolMatch struct {
	protocol TransportProtocol
}

// TransportProtocol represents transport layer protocol
type TransportProtocol int

const (
	TransportProtocolTCP  TransportProtocol = 6
	TransportProtocolUDP  TransportProtocol = 17
	TransportProtocolICMP TransportProtocol = 1
)

// NewProtocolMatch creates a protocol match
func NewProtocolMatch(protocol TransportProtocol) *ProtocolMatch {
	return &ProtocolMatch{protocol: protocol}
}

// Type returns the match type
func (m *ProtocolMatch) Type() MatchType {
	return MatchTypeProtocol
}

// String returns the string representation
func (m *ProtocolMatch) String() string {
	return fmt.Sprintf("ip protocol %d 0xff", m.protocol)
}

// Protocol returns the transport protocol
func (m *ProtocolMatch) Protocol() TransportProtocol {
	return m.protocol
}

// MarkMatch represents a firewall mark match
type MarkMatch struct {
	mark uint32
	mask uint32
}

// NewMarkMatch creates a mark match
func NewMarkMatch(mark uint32) *MarkMatch {
	return &MarkMatch{
		mark: mark,
		mask: 0xFFFFFFFF,
	}
}

// Type returns the match type
func (m *MarkMatch) Type() MatchType {
	return MatchTypeMark
}

// String returns the string representation
func (m *MarkMatch) String() string {
	return fmt.Sprintf("mark 0x%x 0x%x", m.mark, m.mask)
}

// Mark returns the mark value
func (m *MarkMatch) Mark() uint32 {
	return m.mark
}
