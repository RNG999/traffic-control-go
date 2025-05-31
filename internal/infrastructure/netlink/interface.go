package netlink

import (
	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/pkg/types"
)

// NetlinkAdapter defines the interface for netlink operations
type NetlinkAdapter interface {
	// Qdisc operations
	AddQdisc(device valueobjects.DeviceName, qdisc QdiscConfig) types.Result[Unit]
	DeleteQdisc(device valueobjects.DeviceName, handle valueobjects.Handle) types.Result[Unit]
	GetQdiscs(device valueobjects.DeviceName) types.Result[[]QdiscInfo]
	
	// Class operations
	AddClass(device valueobjects.DeviceName, class ClassConfig) types.Result[Unit]
	DeleteClass(device valueobjects.DeviceName, handle valueobjects.Handle) types.Result[Unit]
	GetClasses(device valueobjects.DeviceName) types.Result[[]ClassInfo]
	
	// Filter operations
	AddFilter(device valueobjects.DeviceName, filter FilterConfig) types.Result[Unit]
	DeleteFilter(device valueobjects.DeviceName, parent valueobjects.Handle, priority uint16, handle valueobjects.Handle) types.Result[Unit]
	GetFilters(device valueobjects.DeviceName) types.Result[[]FilterInfo]
}

// Unit represents an empty value (like void)
type Unit struct{}

// QdiscConfig represents configuration for creating a qdisc
type QdiscConfig struct {
	Handle      valueobjects.Handle
	Parent      *valueobjects.Handle
	Type        entities.QdiscType
	Parameters  map[string]interface{}
}

// QdiscInfo represents information about an existing qdisc
type QdiscInfo struct {
	Handle      valueobjects.Handle
	Parent      *valueobjects.Handle
	Type        entities.QdiscType
	Statistics  QdiscStats
}

// QdiscStats represents qdisc statistics
type QdiscStats struct {
	BytesSent    uint64
	PacketsSent  uint64
	BytesDropped uint64
	Overlimits   uint64
	Requeues     uint64
}

// ClassConfig represents configuration for creating a class
type ClassConfig struct {
	Handle     valueobjects.Handle
	Parent     valueobjects.Handle
	Type       entities.QdiscType
	Parameters map[string]interface{}
}

// ClassInfo represents information about an existing class
type ClassInfo struct {
	Handle     valueobjects.Handle
	Parent     valueobjects.Handle
	Type       entities.QdiscType
	Statistics ClassStats
}

// ClassStats represents class statistics
type ClassStats struct {
	BytesSent      uint64
	PacketsSent    uint64
	BytesDropped   uint64
	Overlimits     uint64
	RateBPS        uint64 // Current rate in bits per second
	BacklogBytes   uint64
	BacklogPackets uint64
}

// FilterConfig represents configuration for creating a filter
type FilterConfig struct {
	Parent   valueobjects.Handle
	Priority uint16
	Handle   valueobjects.Handle
	Protocol entities.Protocol
	FlowID   valueobjects.Handle
	Matches  []FilterMatch
}

// FilterMatch represents a filter match configuration
type FilterMatch struct {
	Type  entities.MatchType
	Value interface{}
}

// FilterInfo represents information about an existing filter
type FilterInfo struct {
	Parent   valueobjects.Handle
	Priority uint16
	Handle   valueobjects.Handle
	Protocol entities.Protocol
	FlowID   valueobjects.Handle
	Matches  []FilterMatch
}