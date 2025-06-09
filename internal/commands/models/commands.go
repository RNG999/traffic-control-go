package models

import (
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// Command is the base interface for all commands
type Command interface {
	DeviceName() tc.DeviceName
}

// CreateHTBQdiscCommand creates an HTB qdisc
type CreateHTBQdiscCommand struct {
	DeviceName   string
	Handle       string
	DefaultClass string
}

// CreateTBFQdiscCommand creates a TBF qdisc
type CreateTBFQdiscCommand struct {
	DeviceName string
	Handle     string
	Rate       string // bandwidth string like "100Mbps"
	Buffer     uint32
	Limit      uint32
	Burst      uint32
}

// CreatePRIOQdiscCommand creates a PRIO qdisc
type CreatePRIOQdiscCommand struct {
	DeviceName string
	Handle     string
	Bands      uint8
	Priomap    []uint8
}

// CreateFQCODELQdiscCommand creates a FQ_CODEL qdisc
type CreateFQCODELQdiscCommand struct {
	DeviceName string
	Handle     string
	Limit      uint32
	Flows      uint32
	Target     uint32 // microseconds
	Interval   uint32 // microseconds
	Quantum    uint32
	ECN        bool
}

// CreateHTBClassCommand creates an HTB class
type CreateHTBClassCommand struct {
	DeviceName string
	Parent     string
	ClassID    string
	Name       string // Human-readable name for the class
	Rate       string
	Ceil       string
	Priority   int    // HTB priority (0-7, where 0 is highest)
	// WP2 parameters
	Burst      uint32 // Burst size in bytes (0 = auto-calculate)
	Cburst     uint32 // Ceil burst size in bytes (0 = auto-calculate)
	// Enhanced HTB parameters from main
	Quantum     uint32 // Quantum for borrowing (bytes)
	Overhead    uint32 // Packet overhead (bytes)
	MPU         uint32 // Minimum packet unit (bytes)
	MTU         uint32 // Maximum transmission unit (bytes)
	HTBPrio     uint32 // Internal HTB priority (0-7)
	UseDefaults bool   // Apply default parameters automatically
}

// CreateFilterCommand creates a filter
type CreateFilterCommand struct {
	DeviceName string
	Parent     string
	Priority   uint16
	Protocol   string
	FlowID     string
	Match      map[string]string
}

// CreateAdvancedFilterCommand creates an advanced filter with enhanced capabilities
type CreateAdvancedFilterCommand struct {
	DeviceName string
	Parent     string
	Priority   uint16
	Handle     string
	Protocol   string
	FlowID     string
	// Enhanced filtering options
	IPSourceRange     *IPRange   // Source IP range
	IPDestRange       *IPRange   // Destination IP range
	PortSourceRange   *PortRange // Source port range
	PortDestRange     *PortRange // Destination port range
	TransportProtocol string     // TCP, UDP, ICMP
	TOSValue          uint8      // Type of Service
	DSCPValue         uint8      // DSCP marking
	QoSPriority       uint8      // QoS priority (0-7)
	RateLimit         string     // Rate limiting bandwidth
	BurstLimit        uint32     // Burst limit in bytes
	Action            string     // classify, drop, ratelimit, mark
}

// IPRange represents an IP address range
type IPRange struct {
	StartIP string
	EndIP   string
	CIDR    string // Alternative to range
}

// PortRange represents a port range
type PortRange struct {
	StartPort uint16
	EndPort   uint16
}

// DeleteQdiscCommand deletes a qdisc
type DeleteQdiscCommand struct {
	deviceName tc.DeviceName
	handle     tc.Handle
}

// NewDeleteQdiscCommand creates a new DeleteQdiscCommand
func NewDeleteQdiscCommand(deviceName tc.DeviceName, handle tc.Handle) *DeleteQdiscCommand {
	return &DeleteQdiscCommand{
		deviceName: deviceName,
		handle:     handle,
	}
}

// DeviceName returns the device name
func (c *DeleteQdiscCommand) DeviceName() tc.DeviceName {
	return c.deviceName
}

// Handle returns the qdisc handle
func (c *DeleteQdiscCommand) Handle() tc.Handle {
	return c.handle
}

// DeleteFilterCommand deletes a filter
type DeleteFilterCommand struct {
	DeviceName tc.DeviceName
	Parent     tc.Handle
	Priority   uint16
	Handle     tc.Handle
}

// NewDeleteFilterCommand creates a new DeleteFilterCommand
func NewDeleteFilterCommand(deviceName tc.DeviceName, parent tc.Handle, priority uint16, handle tc.Handle) *DeleteFilterCommand {
	return &DeleteFilterCommand{
		DeviceName: deviceName,
		Parent:     parent,
		Priority:   priority,
		Handle:     handle,
	}
}
