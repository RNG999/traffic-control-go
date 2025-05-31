package netlink

import (
	"fmt"
	"time"
	
	nl "github.com/vishvananda/netlink"
	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
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
	Gap         *uint32  // Packet gap for reordering
	Limit       *uint32  // Queue limit
	Distribution string   // "normal", "pareto", "paretonormal"
}

// AddNetemQdisc adds a NETEM qdisc for network emulation
func (a *RealNetlinkAdapter) AddNetemQdisc(device valueobjects.DeviceName, handle valueobjects.Handle, config NetemConfig) types.Result[Unit] {
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
		netem.Latency = uint32(config.Delay.Nanoseconds() / 1000) // Convert to microseconds
		if config.DelayJitter != nil {
			netem.Jitter = uint32(config.DelayJitter.Nanoseconds() / 1000)
		}
	}
	
	// Set loss parameters
	if config.Loss != nil {
		netem.Loss = *config.Loss
	}
	
	// Set duplicate parameters
	if config.Duplicate != nil {
		netem.Duplicate = *config.Duplicate
	}
	
	// Set corrupt parameters
	if config.Corrupt != nil {
		netem.Corrupt = *config.Corrupt
	}
	
	// Set reorder parameters
	if config.Reorder != nil {
		netem.Reorder = *config.Reorder
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
func addNetemToSwitch(config QdiscConfig) (nl.Qdisc, error) {
	// This would be added to the main adapter's switch statement
	netem := nl.NewNetem(nl.QdiscAttrs{}, nl.NetemQdiscAttrs{})
	
	// Parse NETEM-specific parameters from config.Parameters
	if delay, ok := config.Parameters["delay"].(time.Duration); ok {
		netem.Latency = uint32(delay.Nanoseconds() / 1000)
	}
	if jitter, ok := config.Parameters["jitter"].(time.Duration); ok {
		netem.Jitter = uint32(jitter.Nanoseconds() / 1000)
	}
	if loss, ok := config.Parameters["loss"].(float32); ok {
		netem.Loss = loss
	}
	if duplicate, ok := config.Parameters["duplicate"].(float32); ok {
		netem.Duplicate = duplicate
	}
	if corrupt, ok := config.Parameters["corrupt"].(float32); ok {
		netem.Corrupt = corrupt
	}
	if reorder, ok := config.Parameters["reorder"].(float32); ok {
		netem.Reorder = reorder
	}
	if limit, ok := config.Parameters["limit"].(uint32); ok {
		netem.Limit = limit
	}
	
	return netem, nil
}