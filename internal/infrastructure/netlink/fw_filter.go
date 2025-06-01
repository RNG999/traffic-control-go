//go:build linux
// +build linux

package netlink

import (
	"fmt"

	nl "github.com/vishvananda/netlink"

	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/pkg/types"
)

// FwFilterConfig represents a firewall mark filter configuration
type FwFilterConfig struct {
	Parent   valueobjects.Handle
	Priority uint16
	Mark     uint32              // Firewall mark value
	Mask     uint32              // Mark mask (optional)
	FlowID   valueobjects.Handle // Target class
}

// AddFwFilter adds a firewall mark filter
func (a *RealNetlinkAdapter) AddFwFilter(device valueobjects.DeviceName, config FwFilterConfig) types.Result[Unit] {
	// Get the network link
	link, err := nl.LinkByName(device.String())
	if err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to find device %s: %w", device, err))
	}

	// Create FW filter
	// NOTE: The current version of vishvananda/netlink doesn't support Mark field in FwFilter.
	// FwFilter is primarily for classifying based on firewall marks that are already set on packets.
	// For mark-based filtering, consider using U32 filter or updating the netlink library.
	filter := &nl.FwFilter{
		FilterAttrs: nl.FilterAttrs{
			LinkIndex: link.Attrs().Index,
			Parent:    nl.MakeHandle(config.Parent.Major(), config.Parent.Minor()),
			Priority:  config.Priority,
			Protocol:  0x0300, // ETH_P_ALL
		},
		ClassId: nl.MakeHandle(config.FlowID.Major(), config.FlowID.Minor()),
	}

	// Set mask if provided
	if config.Mask != 0 {
		filter.Mask = config.Mask
	} else {
		filter.Mask = 0xffffffff // Default to exact match
	}

	// TODO: Implement mark-based filtering using U32 filter or alternative approach
	_ = config.Mark // Mark field is not used in current implementation

	// Add the filter
	if err := nl.FilterAdd(filter); err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to add FW filter: %w", err))
	}

	return types.Success(Unit{})
}

// DeleteFwFilter deletes a firewall mark filter
func (a *RealNetlinkAdapter) DeleteFwFilter(device valueobjects.DeviceName, parent valueobjects.Handle, priority uint16) types.Result[Unit] {
	// Get the network link
	link, err := nl.LinkByName(device.String())
	if err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to find device %s: %w", device, err))
	}

	// Create filter to delete
	filter := &nl.FwFilter{
		FilterAttrs: nl.FilterAttrs{
			LinkIndex: link.Attrs().Index,
			Parent:    nl.MakeHandle(parent.Major(), parent.Minor()),
			Priority:  priority,
		},
	}

	if err := nl.FilterDel(filter); err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to delete FW filter: %w", err))
	}

	return types.Success(Unit{})
}
