package models

import (
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
)

// Command is the base interface for all commands
type Command interface {
	DeviceName() valueobjects.DeviceName
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
	Rate       string
	Ceil       string
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


// DeleteQdiscCommand deletes a qdisc
type DeleteQdiscCommand struct {
	deviceName valueobjects.DeviceName
	handle     valueobjects.Handle
}

// NewDeleteQdiscCommand creates a new DeleteQdiscCommand
func NewDeleteQdiscCommand(deviceName valueobjects.DeviceName, handle valueobjects.Handle) *DeleteQdiscCommand {
	return &DeleteQdiscCommand{
		deviceName: deviceName,
		handle:     handle,
	}
}

// DeviceName returns the device name
func (c *DeleteQdiscCommand) DeviceName() valueobjects.DeviceName {
	return c.deviceName
}

// Handle returns the qdisc handle
func (c *DeleteQdiscCommand) Handle() valueobjects.Handle {
	return c.handle
}
