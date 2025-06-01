//go:build linux
// +build linux

package netlink

import (
	"fmt"

	nl "github.com/vishvananda/netlink"

	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/pkg/types"
)

// PoliceAction represents a policing action
type PoliceAction struct {
	Rate     valueobjects.Bandwidth  // Rate limit
	Burst    uint32                  // Burst size in bytes
	MTU      uint32                  // MTU
	Action   PoliceActionType        // Action when exceeded
	PeakRate *valueobjects.Bandwidth // Optional peak rate
}

// PoliceActionType represents the action to take when rate is exceeded
type PoliceActionType int

const (
	PoliceActionDrop PoliceActionType = iota
	PoliceActionPass
	PoliceActionReclassify
	PoliceActionContinue
	PoliceActionPipe
)

// AddPoliceFilter adds a filter with police action
func (a *RealNetlinkAdapter) AddPoliceFilter(device valueobjects.DeviceName, parent valueobjects.Handle, priority uint16, police PoliceAction) types.Result[Unit] {
	// Get the network link
	link, err := nl.LinkByName(device.String())
	if err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to find device %s: %w", device, err))
	}

	// Create basic filter with police action
	filter := &nl.U32{
		FilterAttrs: nl.FilterAttrs{
			LinkIndex: link.Attrs().Index,
			Parent:    nl.MakeHandle(parent.Major(), parent.Minor()),
			Priority:  priority,
			Protocol:  0x0800, // IPv4
		},
		Actions: []nl.Action{},
	}

	// Create police action
	rateBytes := police.Rate.BitsPerSecond() / 8
	if rateBytes > 0xFFFFFFFF {
		return types.Failure[Unit](fmt.Errorf("rate %d bytes/sec exceeds maximum uint32 value", rateBytes))
	}

	policeAction := &nl.PoliceAction{
		Rate:  uint32(rateBytes), // Convert to bytes per second
		Burst: police.Burst,
		Mtu:   police.MTU,
	}

	// Set action type
	switch police.Action {
	case PoliceActionDrop:
		policeAction.ExceedAction = nl.TC_POLICE_SHOT
	case PoliceActionPass:
		policeAction.ExceedAction = nl.TC_POLICE_OK
	case PoliceActionReclassify:
		policeAction.ExceedAction = nl.TC_POLICE_RECLASSIFY
	case PoliceActionContinue:
		policeAction.ExceedAction = nl.TC_POLICE_UNSPEC
	case PoliceActionPipe:
		policeAction.ExceedAction = nl.TC_POLICE_PIPE
	}

	// Set peak rate if specified
	if police.PeakRate != nil {
		peakRateBytes := police.PeakRate.BitsPerSecond() / 8
		if peakRateBytes > 0xFFFFFFFF {
			return types.Failure[Unit](fmt.Errorf("peak rate %d bytes/sec exceeds maximum uint32 value", peakRateBytes))
		}
		policeAction.PeakRate = uint32(peakRateBytes)
	}

	// Add action to filter
	filter.Actions = append(filter.Actions, policeAction)

	// Match all traffic (simplified - in practice you'd add specific matches)
	filter.Sel = &nl.TcU32Sel{
		Flags: 0,
		Nkeys: 1,
		Keys: []nl.TcU32Key{
			{
				Mask:    0,
				Val:     0,
				Off:     0,
				OffMask: 0,
			},
		},
	}

	// Add the filter
	if err := nl.FilterAdd(filter); err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to add police filter: %w", err))
	}

	return types.Success(Unit{})
}
