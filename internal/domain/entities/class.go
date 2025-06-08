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
	rate     tc.Bandwidth
	ceil     tc.Bandwidth
	burst    uint32
	cburst   uint32
	quantum  uint32 // Quantum for borrowing (bytes)
	overhead uint32 // Packet overhead calculation (bytes)
	mpu      uint32 // Minimum packet unit (bytes)
	mtu      uint32 // Maximum transmission unit (bytes)
	prio     uint32 // Internal HTB priority (0-7)
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

// SetQuantum sets the quantum for borrowing
func (h *HTBClass) SetQuantum(quantum uint32) {
	h.quantum = quantum
}

// Quantum returns the quantum for borrowing
func (h *HTBClass) Quantum() uint32 {
	return h.quantum
}

// SetOverhead sets the packet overhead
func (h *HTBClass) SetOverhead(overhead uint32) {
	h.overhead = overhead
}

// Overhead returns the packet overhead
func (h *HTBClass) Overhead() uint32 {
	return h.overhead
}

// SetMPU sets the minimum packet unit
func (h *HTBClass) SetMPU(mpu uint32) {
	h.mpu = mpu
}

// MPU returns the minimum packet unit
func (h *HTBClass) MPU() uint32 {
	return h.mpu
}

// SetMTU sets the maximum transmission unit
func (h *HTBClass) SetMTU(mtu uint32) {
	h.mtu = mtu
}

// MTU returns the maximum transmission unit
func (h *HTBClass) MTU() uint32 {
	return h.mtu
}

// SetHTBPrio sets the internal HTB priority
func (h *HTBClass) SetHTBPrio(prio uint32) {
	h.prio = prio
}

// HTBPrio returns the internal HTB priority
func (h *HTBClass) HTBPrio() uint32 {
	return h.prio
}

// CalculateQuantum calculates appropriate quantum based on rate
func (h *HTBClass) CalculateQuantum() uint32 {
	// Quantum calculation: rate_bps / 8 / HZ
	// Standard Linux HZ is typically 1000, so quantum = rate_bytes_per_second / 1000
	// Minimum quantum is typically 1000 bytes, maximum is 60000 bytes
	const (
		MinQuantum = 1000  // Minimum quantum (1KB)
		MaxQuantum = 60000 // Maximum quantum (60KB)
		HZ         = 1000  // Linux timer frequency
	)

	if h.rate.BitsPerSecond() == 0 {
		return MinQuantum
	}

	bytesPerSecond := h.rate.BitsPerSecond() / 8
	// Prevent integer overflow in conversion
	quantumCalc := bytesPerSecond / HZ
	var quantum uint32
	if quantumCalc > 0xFFFFFFFF {
		quantum = 0xFFFFFFFF
	} else {
		quantum = uint32(quantumCalc) // #nosec G115 - bounds checked above
	}

	// Ensure quantum is within reasonable bounds
	if quantum < MinQuantum {
		return MinQuantum
	}
	if quantum > MaxQuantum {
		return MaxQuantum
	}

	return quantum
}

// CalculateEnhancedBurst calculates burst with MTU and overhead considerations
func (h *HTBClass) CalculateEnhancedBurst() uint32 {
	// Enhanced burst calculation considering MTU, overhead, and timer resolution
	const TimerResolutionMS = 64 // Linux timer resolution in milliseconds

	if h.rate.BitsPerSecond() == 0 {
		return 1600 // Default minimum burst
	}

	// Calculate burst for timer resolution period
	bytesPerSecond := h.rate.BitsPerSecond() / 8
	// Prevent integer overflow in conversion
	burstCalc := bytesPerSecond * TimerResolutionMS / 1000
	var burstBytes uint32
	if burstCalc > 0xFFFFFFFF {
		burstBytes = 0xFFFFFFFF
	} else {
		burstBytes = uint32(burstCalc) // #nosec G115 - bounds checked above
	}

	// Add overhead consideration
	if h.overhead > 0 {
		// Assume average packet size for overhead calculation
		avgPacketSize := uint32(1500) // Standard Ethernet MTU
		if h.mtu > 0 {
			avgPacketSize = h.mtu
		}

		packetsPerBurst := burstBytes / avgPacketSize
		if packetsPerBurst == 0 {
			packetsPerBurst = 1
		}

		overheadTotal := packetsPerBurst * h.overhead
		burstBytes += overheadTotal
	}

	// Ensure minimum burst considering MTU
	minBurst := uint32(1600) // Default minimum
	if h.mtu > 0 {
		minBurst = h.mtu * 2 // At least 2 MTU-sized packets
	}

	if burstBytes < minBurst {
		burstBytes = minBurst
	}

	// burstBytes is already uint32, so no need to cap (would be caught at conversion)

	return burstBytes
}

// CalculateEnhancedCburst calculates cburst with advanced parameters
func (h *HTBClass) CalculateEnhancedCburst() uint32 {
	// Use ceil if set, otherwise use rate
	bandwidth := h.ceil
	if bandwidth.BitsPerSecond() == 0 {
		bandwidth = h.rate
	}

	// Temporarily store original rate to calculate cburst with ceil
	originalRate := h.rate
	h.rate = bandwidth

	cburst := h.CalculateEnhancedBurst()

	// Restore original rate
	h.rate = originalRate

	return cburst
}

// ApplyDefaultParameters applies sensible defaults for HTB parameters
func (h *HTBClass) ApplyDefaultParameters() {
	// Set quantum if not already set
	if h.quantum == 0 {
		h.quantum = h.CalculateQuantum()
	}

	// Set default MTU if not set
	if h.mtu == 0 {
		h.mtu = 1500 // Standard Ethernet MTU
	}

	// Set default MPU if not set
	if h.mpu == 0 {
		h.mpu = 64 // Minimum Ethernet frame payload
	}

	// Set default overhead if not set
	if h.overhead == 0 {
		h.overhead = 4 // Basic Ethernet overhead estimate
	}

	// Calculate burst and cburst using enhanced algorithms
	if h.burst == 0 {
		h.burst = h.CalculateEnhancedBurst()
	}

	if h.cburst == 0 {
		h.cburst = h.CalculateEnhancedCburst()
	}
}
