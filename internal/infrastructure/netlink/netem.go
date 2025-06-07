//go:build linux
// +build linux

package netlink

import (
	"fmt"
	"time"

	nl "github.com/vishvananda/netlink"

	"github.com/rng999/traffic-control-go/pkg/tc"
	"github.com/rng999/traffic-control-go/pkg/types"
)

// NetemConfig represents NETEM qdisc configuration
type NetemConfig struct {
	// Basic parameters
	Delay       *time.Duration
	DelayJitter *time.Duration
	Loss        *float32 // Percentage 0-100
	Duplicate   *float32 // Percentage 0-100
	Corrupt     *float32 // Percentage 0-100
	Reorder     *float32 // Percentage 0-100

	// Advanced parameters
	Gap          *uint32 // Packet gap for reordering
	Limit        *uint32 // Queue limit
	Distribution string  // "normal", "pareto", "paretonormal"
}

// AddNetemQdisc adds a NETEM qdisc for network emulation
func (a *RealNetlinkAdapter) AddNetemQdisc(device tc.DeviceName, handle tc.Handle, config NetemConfig) types.Result[Unit] {
	// Get the network link
	link, err := nl.LinkByName(device.String())
	if err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to find device %s: %w", device, err))
	}

	// Create NETEM qdisc
	netem := nl.NewNetem(nl.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    nl.MakeHandle(handle.Major(), handle.Minor()),
		Parent:    nl.HANDLE_ROOT,
	}, nl.NetemQdiscAttrs{})

	// Set delay parameters
	if config.Delay != nil {
		delayMicros := config.Delay.Nanoseconds() / 1000 // Convert to microseconds
		if delayMicros > 0x7FFFFFFF {                    // Use signed max to avoid potential issues
			return types.Failure[Unit](fmt.Errorf("delay %v exceeds maximum value", config.Delay))
		}
		netem.Latency = uint32(delayMicros) // #nosec G115 - range checked above

		if config.DelayJitter != nil {
			jitterMicros := config.DelayJitter.Nanoseconds() / 1000
			if jitterMicros > 0x7FFFFFFF {
				return types.Failure[Unit](fmt.Errorf("jitter %v exceeds maximum value", config.DelayJitter))
			}
			netem.Jitter = uint32(jitterMicros) // #nosec G115 - range checked above
		}
	}

	// Set loss parameters (convert percentage to fixed point: 0-100% -> 0-UINT32_MAX)
	if config.Loss != nil {
		// Convert percentage (0-100) to kernel's representation
		// Kernel uses: 0 = 0%, UINT32_MAX = 100%
		netem.Loss = uint32((*config.Loss / 100.0) * float32(^uint32(0)))
	}

	// Set duplicate parameters
	if config.Duplicate != nil {
		// Convert percentage to kernel's representation
		netem.Duplicate = uint32((*config.Duplicate / 100.0) * float32(^uint32(0)))
	}

	// Set corrupt parameters
	if config.Corrupt != nil {
		// Convert percentage to kernel's representation
		netem.CorruptProb = uint32((*config.Corrupt / 100.0) * float32(^uint32(0)))
	}

	// Set reorder parameters
	if config.Reorder != nil {
		// Convert percentage to kernel's representation
		netem.ReorderProb = uint32((*config.Reorder / 100.0) * float32(^uint32(0)))
		if config.Gap != nil {
			netem.Gap = *config.Gap
		}
	}

	// Set limit if specified
	if config.Limit != nil {
		netem.Limit = *config.Limit
	}

	// Add the qdisc
	if err := nl.QdiscAdd(netem); err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to add NETEM qdisc: %w", err))
	}

	return types.Success(Unit{})
}

// Example implementation in the adapter switch statement:
// This function shows how NETEM would be integrated, kept for reference
func _addNetemToSwitch(config QdiscConfig) (nl.Qdisc, error) { //nolint:unused
	// This would be added to the main adapter's switch statement
	netem := nl.NewNetem(nl.QdiscAttrs{}, nl.NetemQdiscAttrs{})

	// Parse NETEM-specific parameters from config.Parameters
	if delay, ok := config.Parameters["delay"].(time.Duration); ok {
		delayMicros := delay.Nanoseconds() / 1000
		if delayMicros > 0x7FFFFFFF { // Use signed max to avoid potential issues
			return nil, fmt.Errorf("delay %v exceeds maximum value", delay)
		}
		netem.Latency = uint32(delayMicros) // #nosec G115 - range checked above
	}
	if jitter, ok := config.Parameters["jitter"].(time.Duration); ok {
		jitterMicros := jitter.Nanoseconds() / 1000
		if jitterMicros > 0x7FFFFFFF {
			return nil, fmt.Errorf("jitter %v exceeds maximum value", jitter)
		}
		netem.Jitter = uint32(jitterMicros) // #nosec G115 - range checked above
	}
	if loss, ok := config.Parameters["loss"].(float32); ok {
		// Convert percentage to kernel's representation
		netem.Loss = uint32((loss / 100.0) * float32(^uint32(0)))
	}
	if duplicate, ok := config.Parameters["duplicate"].(float32); ok {
		// Convert percentage to kernel's representation
		netem.Duplicate = uint32((duplicate / 100.0) * float32(^uint32(0)))
	}
	if corrupt, ok := config.Parameters["corrupt"].(float32); ok {
		// Convert percentage to kernel's representation
		netem.CorruptProb = uint32((corrupt / 100.0) * float32(^uint32(0)))
	}
	if reorder, ok := config.Parameters["reorder"].(float32); ok {
		// Convert percentage to kernel's representation
		netem.ReorderProb = uint32((reorder / 100.0) * float32(^uint32(0)))
	}
	if limit, ok := config.Parameters["limit"].(uint32); ok {
		netem.Limit = limit
	}

	return netem, nil
}
