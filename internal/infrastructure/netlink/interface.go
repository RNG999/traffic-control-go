package netlink

import (
	"context"

	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/pkg/tc"
	"github.com/rng999/traffic-control-go/pkg/types"
)

// Adapter defines the interface for netlink operations
type Adapter interface {
	// Qdisc operations
	AddQdisc(ctx context.Context, qdisc *entities.Qdisc) error
	DeleteQdisc(device tc.DeviceName, handle tc.Handle) types.Result[Unit]
	GetQdiscs(device tc.DeviceName) types.Result[[]QdiscInfo]

	// Class operations
	AddClass(ctx context.Context, class interface{}) error
	DeleteClass(device tc.DeviceName, handle tc.Handle) types.Result[Unit]
	GetClasses(device tc.DeviceName) types.Result[[]ClassInfo]

	// Filter operations
	AddFilter(ctx context.Context, filter *entities.Filter) error
	DeleteFilter(device tc.DeviceName, parent tc.Handle, priority uint16, handle tc.Handle) types.Result[Unit]
	GetFilters(device tc.DeviceName) types.Result[[]FilterInfo]

	// Statistics operations
	GetDetailedQdiscStats(device tc.DeviceName, handle tc.Handle) types.Result[DetailedQdiscStats]
	GetDetailedClassStats(device tc.DeviceName, handle tc.Handle) types.Result[DetailedClassStats]
	GetLinkStats(device tc.DeviceName) types.Result[LinkStats]
}

// Unit represents an empty value (like void)
type Unit struct{}

// QdiscConfig represents configuration for creating a qdisc
type QdiscConfig struct {
	Handle     tc.Handle
	Parent     *tc.Handle
	Type       entities.QdiscType
	Parameters map[string]interface{}
}

// QdiscInfo represents information about an existing qdisc
type QdiscInfo struct {
	Handle     tc.Handle
	Parent     *tc.Handle
	Type       entities.QdiscType
	Statistics QdiscStats
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
	Handle     tc.Handle
	Parent     tc.Handle
	Type       entities.QdiscType
	Parameters map[string]interface{}
}

// ClassInfo represents information about an existing class
type ClassInfo struct {
	Handle     tc.Handle
	Parent     tc.Handle
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
	Parent   tc.Handle
	Priority uint16
	Handle   tc.Handle
	Protocol entities.Protocol
	FlowID   tc.Handle
	Matches  []FilterMatch
}

// FilterMatch represents a filter match configuration
type FilterMatch struct {
	Type  entities.MatchType
	Value interface{}
}

// FilterInfo represents information about an existing filter
type FilterInfo struct {
	Parent   tc.Handle
	Priority uint16
	Handle   tc.Handle
	Protocol entities.Protocol
	FlowID   tc.Handle
	Matches  []FilterMatch
}

// LinkStats represents network interface statistics
type LinkStats struct {
	RxBytes   uint64
	TxBytes   uint64
	RxPackets uint64
	TxPackets uint64
	RxErrors  uint64
	TxErrors  uint64
	RxDropped uint64
	TxDropped uint64
}
