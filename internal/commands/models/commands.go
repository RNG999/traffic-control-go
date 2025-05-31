package models

import (
	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
)

// Command is the base interface for all commands
type Command interface {
	DeviceName() valueobjects.DeviceName
}

// CreateHTBQdiscCommand creates an HTB qdisc
type CreateHTBQdiscCommand struct {
	deviceName   valueobjects.DeviceName
	handle       valueobjects.Handle
	defaultClass valueobjects.Handle
}

// NewCreateHTBQdiscCommand creates a new CreateHTBQdiscCommand
func NewCreateHTBQdiscCommand(deviceName valueobjects.DeviceName, handle valueobjects.Handle, defaultClass valueobjects.Handle) *CreateHTBQdiscCommand {
	return &CreateHTBQdiscCommand{
		deviceName:   deviceName,
		handle:       handle,
		defaultClass: defaultClass,
	}
}

// DeviceName returns the device name
func (c *CreateHTBQdiscCommand) DeviceName() valueobjects.DeviceName {
	return c.deviceName
}

// Handle returns the qdisc handle
func (c *CreateHTBQdiscCommand) Handle() valueobjects.Handle {
	return c.handle
}

// DefaultClass returns the default class handle
func (c *CreateHTBQdiscCommand) DefaultClass() valueobjects.Handle {
	return c.defaultClass
}

// CreateHTBClassCommand creates an HTB class
type CreateHTBClassCommand struct {
	deviceName valueobjects.DeviceName
	parent     valueobjects.Handle
	handle     valueobjects.Handle
	name       string
	rate       valueobjects.Bandwidth
	ceil       valueobjects.Bandwidth
}

// NewCreateHTBClassCommand creates a new CreateHTBClassCommand
func NewCreateHTBClassCommand(
	deviceName valueobjects.DeviceName,
	parent valueobjects.Handle,
	handle valueobjects.Handle,
	name string,
	rate valueobjects.Bandwidth,
	ceil valueobjects.Bandwidth,
) *CreateHTBClassCommand {
	return &CreateHTBClassCommand{
		deviceName: deviceName,
		parent:     parent,
		handle:     handle,
		name:       name,
		rate:       rate,
		ceil:       ceil,
	}
}

// DeviceName returns the device name
func (c *CreateHTBClassCommand) DeviceName() valueobjects.DeviceName {
	return c.deviceName
}

// Parent returns the parent handle
func (c *CreateHTBClassCommand) Parent() valueobjects.Handle {
	return c.parent
}

// Handle returns the class handle
func (c *CreateHTBClassCommand) Handle() valueobjects.Handle {
	return c.handle
}

// Name returns the class name
func (c *CreateHTBClassCommand) Name() string {
	return c.name
}

// Rate returns the guaranteed rate
func (c *CreateHTBClassCommand) Rate() valueobjects.Bandwidth {
	return c.rate
}

// Ceil returns the maximum rate
func (c *CreateHTBClassCommand) Ceil() valueobjects.Bandwidth {
	return c.ceil
}

// CreateFilterCommand creates a filter
type CreateFilterCommand struct {
	deviceName valueobjects.DeviceName
	parent     valueobjects.Handle
	priority   uint16
	handle     valueobjects.Handle
	flowID     valueobjects.Handle
	matches    []entities.Match
}

// NewCreateFilterCommand creates a new CreateFilterCommand
func NewCreateFilterCommand(
	deviceName valueobjects.DeviceName,
	parent valueobjects.Handle,
	priority uint16,
	handle valueobjects.Handle,
	flowID valueobjects.Handle,
	matches []entities.Match,
) *CreateFilterCommand {
	return &CreateFilterCommand{
		deviceName: deviceName,
		parent:     parent,
		priority:   priority,
		handle:     handle,
		flowID:     flowID,
		matches:    matches,
	}
}

// DeviceName returns the device name
func (c *CreateFilterCommand) DeviceName() valueobjects.DeviceName {
	return c.deviceName
}

// Parent returns the parent handle
func (c *CreateFilterCommand) Parent() valueobjects.Handle {
	return c.parent
}

// Priority returns the filter priority
func (c *CreateFilterCommand) Priority() uint16 {
	return c.priority
}

// Handle returns the filter handle
func (c *CreateFilterCommand) Handle() valueobjects.Handle {
	return c.handle
}

// FlowID returns the target class handle
func (c *CreateFilterCommand) FlowID() valueobjects.Handle {
	return c.flowID
}

// Matches returns the filter matches
func (c *CreateFilterCommand) Matches() []entities.Match {
	return c.matches
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
