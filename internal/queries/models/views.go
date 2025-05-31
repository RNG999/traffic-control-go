package models

import (
	"fmt"
	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
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
	DeviceName          string
	Handle              string
	Parent              string
	Name                string
	Priority            string
	GuaranteedBandwidth string
	MaxBandwidth        string
	CurrentBandwidth    string // For statistics
	DroppedPackets      uint64 // For statistics
}

// FilterView is a read model for filters
type FilterView struct {
	DeviceName string
	Parent     string
	Priority   uint16
	Handle     string
	FlowID     string
	Protocol   string
	Matches    []MatchView
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
func NewQdiscView(device valueobjects.DeviceName, qdisc interface{}) QdiscView {
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
func NewClassView(device valueobjects.DeviceName, class interface{}) ClassView {
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

	// Convert priority to string
	view.Priority = fmt.Sprintf("%d", basicClass.Priority())

	// Add HTB-specific parameters
	if htb, ok := class.(*entities.HTBClass); ok {
		view.GuaranteedBandwidth = htb.Rate().HumanReadable()
		view.MaxBandwidth = htb.Ceil().HumanReadable()
	}

	return view
}

// NewFilterView creates a FilterView from domain entity
func NewFilterView(device valueobjects.DeviceName, filter *entities.Filter) FilterView {
	view := FilterView{
		DeviceName: device.String(),
		Parent:     filter.ID().String(),
		Priority:   filter.ID().Priority(),
		Handle:     filter.ID().Handle().String(),
		FlowID:     filter.FlowID().String(),
		Matches:    make([]MatchView, 0),
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
		view.Matches = append(view.Matches, MatchView{
			Type:  getMatchTypeName(match.Type()),
			Value: match.String(),
		})
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
