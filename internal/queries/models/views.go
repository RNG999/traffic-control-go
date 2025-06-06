package models

import (
	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// QdiscView is a read model for qdiscs
type QdiscView struct {
	DeviceName   string
	Handle       string
	Type         string
	Parent       string
	DefaultClass string // For HTB
	Parameters   map[string]interface{}
}

// ClassView is a read model for classes
type ClassView struct {
	DeviceName string                 `json:"device_name"`
	Handle     string                 `json:"handle"`
	Parent     string                 `json:"parent"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Rate       string                 `json:"rate"`
	Ceil       string                 `json:"ceil"`
	Priority   int                    `json:"priority,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	// Legacy fields for compatibility
	GuaranteedBandwidth string `json:"guaranteed_bandwidth,omitempty"`
	MaxBandwidth        string `json:"max_bandwidth,omitempty"`
	CurrentBandwidth    string `json:"current_bandwidth,omitempty"`
	DroppedPackets      uint64 `json:"dropped_packets,omitempty"`
}

// FilterView is a read model for filters
type FilterView struct {
	DeviceName string            `json:"device_name"`
	Parent     string            `json:"parent"`
	Priority   uint16            `json:"priority"`
	Handle     string            `json:"handle"`
	Protocol   string            `json:"protocol"`
	FlowID     string            `json:"flow_id"`
	Matches    map[string]string `json:"matches"`
}

// MatchView is a read model for filter matches
type MatchView struct {
	Type  string
	Value string
}

// TrafficControlConfigView is a complete view of TC configuration
type TrafficControlConfigView struct {
	DeviceName string
	Qdiscs     []QdiscView
	Classes    []ClassView
	Filters    []FilterView
}

// NewQdiscView creates a QdiscView from domain entity
func NewQdiscView(device tc.DeviceName, qdisc interface{}) QdiscView {
	// First check if it's a basic Qdisc
	basicQdisc, isBasicQdisc := qdisc.(*entities.Qdisc)
	if !isBasicQdisc {
		// If not, check if it's HTBQdisc which embeds Qdisc
		if htb, ok := qdisc.(*entities.HTBQdisc); ok {
			basicQdisc = htb.Qdisc
		} else {
			// Return empty view if type assertion fails
			return QdiscView{}
		}
	}

	view := QdiscView{
		DeviceName: device.String(),
		Handle:     basicQdisc.Handle().String(),
		Type:       basicQdisc.Type().String(),
		Parameters: make(map[string]interface{}),
	}

	if basicQdisc.Parent() != nil {
		view.Parent = basicQdisc.Parent().String()
	}

	// Add HTB-specific parameters
	if htb, ok := qdisc.(*entities.HTBQdisc); ok {
		view.DefaultClass = htb.DefaultClass().String()
		view.Parameters["r2q"] = htb.R2Q()
	}

	return view
}

// NewClassView creates a ClassView from domain entity
func NewClassView(device tc.DeviceName, class interface{}) ClassView {
	// First check if it's a basic Class
	basicClass, isBasicClass := class.(*entities.Class)
	if !isBasicClass {
		// If not, check if it's HTBClass which embeds Class
		if htb, ok := class.(*entities.HTBClass); ok {
			basicClass = htb.Class
		} else {
			// Return empty view if type assertion fails
			return ClassView{}
		}
	}

	view := ClassView{
		DeviceName: device.String(),
		Handle:     basicClass.Handle().String(),
		Parent:     basicClass.Parent().String(),
		Name:       basicClass.Name(),
	}

	// Convert priority to int
	if basicClass.Priority() != nil {
		view.Priority = int(*basicClass.Priority())
	}

	// Add HTB-specific parameters
	if htb, ok := class.(*entities.HTBClass); ok {
		view.GuaranteedBandwidth = htb.Rate().HumanReadable()
		view.MaxBandwidth = htb.Ceil().HumanReadable()
	}

	return view
}

// NewFilterView creates a FilterView from domain entity
func NewFilterView(device tc.DeviceName, filter *entities.Filter) FilterView {
	view := FilterView{
		DeviceName: device.String(),
		Parent:     filter.ID().String(),
		Priority:   filter.ID().Priority(),
		Handle:     filter.ID().Handle().String(),
		FlowID:     filter.FlowID().String(),
		Matches:    make(map[string]string),
	}

	// Convert protocol
	switch filter.Protocol() {
	case entities.ProtocolAll:
		view.Protocol = "all"
	case entities.ProtocolIP:
		view.Protocol = "ip"
	case entities.ProtocolIPv6:
		view.Protocol = "ipv6"
	}

	// Convert matches
	for _, match := range filter.Matches() {
		view.Matches[getMatchTypeName(match.Type())] = match.String()
	}

	return view
}

func getMatchTypeName(matchType entities.MatchType) string {
	switch matchType {
	case entities.MatchTypeIPSource:
		return "Source IP"
	case entities.MatchTypeIPDestination:
		return "Destination IP"
	case entities.MatchTypePortSource:
		return "Source Port"
	case entities.MatchTypePortDestination:
		return "Destination Port"
	case entities.MatchTypeProtocol:
		return "Protocol"
	case entities.MatchTypeMark:
		return "Firewall Mark"
	default:
		return "Unknown"
	}
}

// PrettyPrint returns a human-readable representation of the config
func (v *TrafficControlConfigView) PrettyPrint() string {
	// Implementation would format the config nicely
	// This is a placeholder
	return "Traffic Control Configuration for " + v.DeviceName
}

// ConfigurationView represents the complete configuration with version
type ConfigurationView struct {
	DeviceName string       `json:"device_name"`
	Qdiscs     []QdiscView  `json:"qdiscs"`
	Classes    []ClassView  `json:"classes"`
	Filters    []FilterView `json:"filters"`
	Version    int          `json:"version"`
}

// DeviceStatisticsView represents statistics for a device
type DeviceStatisticsView struct {
	DeviceName  string                 `json:"device_name"`
	Timestamp   string                 `json:"timestamp"`
	QdiscStats  []QdiscStatisticsView  `json:"qdisc_stats"`
	ClassStats  []ClassStatisticsView  `json:"class_stats"`
	FilterStats []FilterStatisticsView `json:"filter_stats"`
	LinkStats   LinkStatisticsView     `json:"link_stats"`
}

// QdiscStatisticsView represents qdisc statistics with metadata
type QdiscStatisticsView struct {
	Handle        string                 `json:"handle"`
	Type          string                 `json:"type"`
	BytesSent     uint64                 `json:"bytes_sent"`
	PacketsSent   uint64                 `json:"packets_sent"`
	BytesDropped  uint64                 `json:"bytes_dropped"`
	Overlimits    uint64                 `json:"overlimits"`
	Requeues      uint64                 `json:"requeues"`
	Backlog       uint32                 `json:"backlog"`
	QueueLength   uint32                 `json:"queue_length"`
	DetailedStats map[string]interface{} `json:"detailed_stats,omitempty"`
}

// ClassStatisticsView represents class statistics with metadata
type ClassStatisticsView struct {
	Handle         string                 `json:"handle"`
	Parent         string                 `json:"parent"`
	Name           string                 `json:"name"`
	BytesSent      uint64                 `json:"bytes_sent"`
	PacketsSent    uint64                 `json:"packets_sent"`
	BytesDropped   uint64                 `json:"bytes_dropped"`
	Overlimits     uint64                 `json:"overlimits"`
	BacklogBytes   uint64                 `json:"backlog_bytes"`
	BacklogPackets uint64                 `json:"backlog_packets"`
	RateBPS        uint64                 `json:"rate_bps"`
	DetailedStats  map[string]interface{} `json:"detailed_stats,omitempty"`
}

// FilterStatisticsView represents filter statistics with metadata
type FilterStatisticsView struct {
	Parent     string `json:"parent"`
	Priority   uint16 `json:"priority"`
	Protocol   string `json:"protocol"`
	Handle     string `json:"handle"`
	MatchCount int    `json:"match_count"`
	FlowID     string `json:"flow_id"`
}

// LinkStatisticsView represents network interface statistics
type LinkStatisticsView struct {
	RxBytes   uint64 `json:"rx_bytes"`
	TxBytes   uint64 `json:"tx_bytes"`
	RxPackets uint64 `json:"rx_packets"`
	TxPackets uint64 `json:"tx_packets"`
	RxErrors  uint64 `json:"rx_errors"`
	TxErrors  uint64 `json:"tx_errors"`
	RxDropped uint64 `json:"rx_dropped"`
	TxDropped uint64 `json:"tx_dropped"`
}
