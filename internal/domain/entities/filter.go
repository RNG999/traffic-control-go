package entities

import (
	"fmt"
	"net"

	"github.com/rng999/traffic-control-go/pkg/tc"
)

// FilterID represents a unique identifier for a filter
type FilterID struct {
	device   tc.DeviceName
	parent   tc.Handle
	priority uint16
	handle   tc.Handle
}

// NewFilterID creates a new FilterID
func NewFilterID(device tc.DeviceName, parent tc.Handle, priority uint16, handle tc.Handle) FilterID {
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

// Device returns the device name
func (id FilterID) Device() tc.DeviceName {
	return id.device
}

// Parent returns the parent handle
func (id FilterID) Parent() tc.Handle {
	return id.parent
}

// Priority returns the filter priority
func (id FilterID) Priority() uint16 {
	return id.priority
}

// Handle returns the filter handle
func (id FilterID) Handle() tc.Handle {
	return id.handle
}

// Filter represents a packet classification filter
type Filter struct {
	id       FilterID
	flowID   tc.Handle // Target class
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
func NewFilter(device tc.DeviceName, parent tc.Handle, priority uint16, handle tc.Handle) *Filter {
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
func (f *Filter) SetFlowID(flowID tc.Handle) {
	f.flowID = flowID
}

// FlowID returns the target class handle
func (f *Filter) FlowID() tc.Handle {
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
	MatchTypePortRange
	MatchTypeTOS
	MatchTypeDSCP
	MatchTypeFlowID
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

// PortRangeMatch represents a port range match
type PortRangeMatch struct {
	matchType MatchType
	startPort uint16
	endPort   uint16
}

// NewPortSourceRangeMatch creates a source port range match
func NewPortSourceRangeMatch(startPort, endPort uint16) *PortRangeMatch {
	return &PortRangeMatch{
		matchType: MatchTypePortRange,
		startPort: startPort,
		endPort:   endPort,
	}
}

// NewPortDestinationRangeMatch creates a destination port range match
func NewPortDestinationRangeMatch(startPort, endPort uint16) *PortRangeMatch {
	return &PortRangeMatch{
		matchType: MatchTypePortRange,
		startPort: startPort,
		endPort:   endPort,
	}
}

// Type returns the match type
func (m *PortRangeMatch) Type() MatchType {
	return m.matchType
}

// String returns the string representation
func (m *PortRangeMatch) String() string {
	return fmt.Sprintf("port range %d-%d", m.startPort, m.endPort)
}

// StartPort returns the start port
func (m *PortRangeMatch) StartPort() uint16 {
	return m.startPort
}

// EndPort returns the end port
func (m *PortRangeMatch) EndPort() uint16 {
	return m.endPort
}

// TOSMatch represents a Type of Service match
type TOSMatch struct {
	tos  uint8
	mask uint8
}

// NewTOSMatch creates a TOS match
func NewTOSMatch(tos uint8) *TOSMatch {
	return &TOSMatch{
		tos:  tos,
		mask: 0xFF,
	}
}

// Type returns the match type
func (m *TOSMatch) Type() MatchType {
	return MatchTypeTOS
}

// String returns the string representation
func (m *TOSMatch) String() string {
	return fmt.Sprintf("ip tos 0x%x 0x%x", m.tos, m.mask)
}

// TOS returns the TOS value
func (m *TOSMatch) TOS() uint8 {
	return m.tos
}

// DSCPMatch represents a DSCP (Differentiated Services Code Point) match
type DSCPMatch struct {
	dscp uint8
}

// NewDSCPMatch creates a DSCP match
func NewDSCPMatch(dscp uint8) *DSCPMatch {
	return &DSCPMatch{dscp: dscp}
}

// Type returns the match type
func (m *DSCPMatch) Type() MatchType {
	return MatchTypeDSCP
}

// String returns the string representation
func (m *DSCPMatch) String() string {
	// DSCP is in the top 6 bits of TOS field
	tosValue := m.dscp << 2
	return fmt.Sprintf("ip tos 0x%x 0xfc", tosValue)
}

// DSCP returns the DSCP value
func (m *DSCPMatch) DSCP() uint8 {
	return m.dscp
}

// FlowIDMatch represents a flow-based match using hash tables
type FlowIDMatch struct {
	keys   []string
	mask   uint32
	hashID uint32
}

// NewFlowIDMatch creates a flow-based match
func NewFlowIDMatch(keys []string, mask uint32) *FlowIDMatch {
	return &FlowIDMatch{
		keys: keys,
		mask: mask,
	}
}

// Type returns the match type
func (m *FlowIDMatch) Type() MatchType {
	return MatchTypeFlowID
}

// String returns the string representation
func (m *FlowIDMatch) String() string {
	return fmt.Sprintf("flow keys %v mask 0x%x", m.keys, m.mask)
}

// Keys returns the flow keys
func (m *FlowIDMatch) Keys() []string {
	return m.keys
}

// Mask returns the flow mask
func (m *FlowIDMatch) Mask() uint32 {
	return m.mask
}

// AdvancedFilter represents an enhanced filter with complex matching capabilities
type AdvancedFilter struct {
	*Filter
	priority   uint8           // QoS priority (0-7)
	rateLimit  *tc.Bandwidth   // Rate limiting for matched traffic
	burstLimit uint32          // Burst limit in bytes
	action     FilterAction    // Action to take on match
}

// FilterAction represents the action to take when filter matches
type FilterAction int

const (
	ActionClassify FilterAction = iota // Classify to specific class
	ActionDrop                         // Drop the packet
	ActionRateLimit                    // Apply rate limiting
	ActionMark                         // Mark the packet
)

// NewAdvancedFilter creates a new advanced filter
func NewAdvancedFilter(device tc.DeviceName, parent tc.Handle, priority uint16, handle tc.Handle) *AdvancedFilter {
	return &AdvancedFilter{
		Filter: NewFilter(device, parent, priority, handle),
		action: ActionClassify,
	}
}

// SetQoSPriority sets the QoS priority
func (f *AdvancedFilter) SetQoSPriority(priority uint8) {
	f.priority = priority
}

// QoSPriority returns the QoS priority
func (f *AdvancedFilter) QoSPriority() uint8 {
	return f.priority
}

// SetRateLimit sets the rate limiting parameters
func (f *AdvancedFilter) SetRateLimit(rate tc.Bandwidth, burst uint32) {
	f.rateLimit = &rate
	f.burstLimit = burst
}

// RateLimit returns the rate limit
func (f *AdvancedFilter) RateLimit() *tc.Bandwidth {
	return f.rateLimit
}

// BurstLimit returns the burst limit
func (f *AdvancedFilter) BurstLimit() uint32 {
	return f.burstLimit
}

// SetAction sets the filter action
func (f *AdvancedFilter) SetAction(action FilterAction) {
	f.action = action
}

// Action returns the filter action
func (f *AdvancedFilter) Action() FilterAction {
	return f.action
}

// AddIPRangeMatch adds a match for IP address range
func (f *Filter) AddIPRangeMatch(startIP, endIP string, isSource bool) error {
	startIPNet := net.ParseIP(startIP)
	endIPNet := net.ParseIP(endIP)
	
	if startIPNet == nil || endIPNet == nil {
		return fmt.Errorf("invalid IP addresses: %s - %s", startIP, endIP)
	}

	// For IP ranges, we'll need to create multiple CIDR matches
	// This is a simplified implementation - in practice, you'd use a more sophisticated algorithm
	if isSource {
		match, err := NewIPSourceMatch(startIP + "/32")
		if err != nil {
			return err
		}
		f.AddMatch(match)
	} else {
		match, err := NewIPDestinationMatch(startIP + "/32")
		if err != nil {
			return err
		}
		f.AddMatch(match)
	}

	return nil
}

// AddProtocolMatch adds a protocol-based match
func (f *Filter) AddProtocolMatch(protocol TransportProtocol) {
	match := NewProtocolMatch(protocol)
	f.AddMatch(match)
}

// AddPortRangeMatch adds a port range match
func (f *Filter) AddPortRangeMatch(startPort, endPort uint16, isSource bool) {
	var match *PortRangeMatch
	if isSource {
		match = NewPortSourceRangeMatch(startPort, endPort)
	} else {
		match = NewPortDestinationRangeMatch(startPort, endPort)
	}
	f.AddMatch(match)
}

// AddTOSMatch adds a Type of Service match
func (f *Filter) AddTOSMatch(tos uint8) {
	match := NewTOSMatch(tos)
	f.AddMatch(match)
}

// AddDSCPMatch adds a DSCP match
func (f *Filter) AddDSCPMatch(dscp uint8) {
	match := NewDSCPMatch(dscp)
	f.AddMatch(match)
}

// ValidateMatches validates that all matches are compatible
func (f *Filter) ValidateMatches() error {
	protocolCount := 0
	ipCount := 0
	
	for _, match := range f.matches {
		switch match.Type() {
		case MatchTypeProtocol:
			protocolCount++
		case MatchTypeIPSource, MatchTypeIPDestination:
			ipCount++
		}
	}

	if protocolCount > 1 {
		return fmt.Errorf("multiple protocol matches not supported")
	}

	// Additional validation logic here...
	return nil
}
