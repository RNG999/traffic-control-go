package entities

import (
	"fmt"

	"github.com/rng999/traffic-control-go/pkg/tc"
)

// ClassID represents a unique identifier for a traffic class
type ClassID struct {
	device tc.DeviceName
	handle tc.Handle
}

// NewClassID creates a new ClassID
func NewClassID(device tc.DeviceName, handle tc.Handle) ClassID {
	return ClassID{device: device, handle: handle}
}

// String returns the string representation of ClassID
func (id ClassID) String() string {
	return fmt.Sprintf("%s:%s", id.device, id.handle)
}

// Device returns the device name
func (id ClassID) Device() tc.DeviceName {
	return id.device
}

// Class represents a traffic class entity
type Class struct {
	id       ClassID
	parent   tc.Handle
	name     string    // Human-readable name
	priority *Priority // Priority must be explicitly set
}

// Priority represents the priority level of a class (0-7, where 0 is highest priority)
type Priority int

// NewClass creates a new Class entity
func NewClass(device tc.DeviceName, handle tc.Handle, parent tc.Handle, name string, priority Priority) *Class {
	return &Class{
		id:       NewClassID(device, handle),
		parent:   parent,
		name:     name,
		priority: &priority,
	}
}

// ID returns the class ID
func (c *Class) ID() ClassID {
	return c.id
}

// Handle returns the class handle
func (c *Class) Handle() tc.Handle {
	return c.id.handle
}

// Parent returns the parent handle
func (c *Class) Parent() tc.Handle {
	return c.parent
}

// Name returns the human-readable name
func (c *Class) Name() string {
	return c.name
}

// Priority returns the priority
func (c *Class) Priority() *Priority {
	return c.priority
}

// SetPriority sets the priority
func (c *Class) SetPriority(p Priority) {
	c.priority = &p
}

// HTBClass represents an HTB-specific traffic class
type HTBClass struct {
	*Class
	rate   tc.Bandwidth
	ceil   tc.Bandwidth
	burst  uint32
	cburst uint32
}

// NewHTBClass creates a new HTB class
func NewHTBClass(device tc.DeviceName, handle tc.Handle, parent tc.Handle, name string, priority Priority) *HTBClass {
	class := NewClass(device, handle, parent, name, priority)
	return &HTBClass{
		Class: class,
	}
}

// SetRate sets the guaranteed rate
func (h *HTBClass) SetRate(rate tc.Bandwidth) {
	h.rate = rate
}

// Rate returns the guaranteed rate
func (h *HTBClass) Rate() tc.Bandwidth {
	return h.rate
}

// SetCeil sets the maximum rate
func (h *HTBClass) SetCeil(ceil tc.Bandwidth) {
	h.ceil = ceil
}

// Ceil returns the maximum rate
func (h *HTBClass) Ceil() tc.Bandwidth {
	return h.ceil
}

// SetBurst sets the burst size
func (h *HTBClass) SetBurst(burst uint32) {
	h.burst = burst
}

// Burst returns the burst size
func (h *HTBClass) Burst() uint32 {
	return h.burst
}

// SetCburst sets the ceil burst size
func (h *HTBClass) SetCburst(cburst uint32) {
	h.cburst = cburst
}

// Cburst returns the ceil burst size
func (h *HTBClass) Cburst() uint32 {
	return h.cburst
}

// CalculateBurst calculates appropriate burst size based on rate
func (h *HTBClass) CalculateBurst() uint32 {
	// Basic calculation: rate_bps / 8 * 0.01s (10ms timer)
	// Multiply by 10 for safety margin
	bytesPerSecond := h.rate.BitsPerSecond() / 8
	// Avoid floating point: 0.01 * 10 = 0.1 = 1/10
	burstValue := bytesPerSecond / 10
	if burstValue > 0xFFFFFFFF {
		return 0xFFFFFFFF // Cap at maximum uint32 value
	}
	return uint32(burstValue)
}

// CalculateCburst calculates appropriate cburst size based on ceil
func (h *HTBClass) CalculateCburst() uint32 {
	// Use ceil if set, otherwise use rate
	bandwidth := h.ceil
	if bandwidth.BitsPerSecond() == 0 {
		bandwidth = h.rate
	}

	bytesPerSecond := bandwidth.BitsPerSecond() / 8
	// Avoid floating point: 0.01 * 10 = 0.1 = 1/10
	cburstValue := bytesPerSecond / 10
	if cburstValue > 0xFFFFFFFF {
		return 0xFFFFFFFF // Cap at maximum uint32 value
	}
	return uint32(cburstValue)
}
