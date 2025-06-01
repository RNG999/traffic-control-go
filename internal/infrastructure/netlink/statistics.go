//go:build linux
// +build linux

package netlink

import (
	"fmt"

	nl "github.com/vishvananda/netlink"

	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/pkg/types"
)

// GetDetailedQdiscStats returns detailed statistics for a qdisc
func (a *RealNetlinkAdapter) GetDetailedQdiscStats(device valueobjects.DeviceName, handle valueobjects.Handle) types.Result[DetailedQdiscStats] {
	// Get the network link
	link, err := nl.LinkByName(device.String())
	if err != nil {
		return types.Failure[DetailedQdiscStats](fmt.Errorf("failed to find device %s: %w", device, err))
	}

	// Get all qdiscs for the link
	qdiscs, err := nl.QdiscList(link)
	if err != nil {
		return types.Failure[DetailedQdiscStats](fmt.Errorf("failed to list qdiscs: %w", err))
	}

	// Find the specific qdisc
	for _, qdisc := range qdiscs {
		if qdisc.Attrs().Handle == nl.MakeHandle(handle.Major(), handle.Minor()) {
			stats := DetailedQdiscStats{}

			// Get basic statistics if available
			if qdisc.Attrs().Statistics != nil {
				qs := qdisc.Attrs().Statistics
				if qs.Basic != nil {
					stats.BasicStats = QdiscStats{
						BytesSent:   qs.Basic.Bytes,
						PacketsSent: uint64(qs.Basic.Packets),
						// Note: Drops, Overlimits, and Requeues are not available in GnetStatsBasic
						BytesDropped: 0,
						Overlimits:   0,
						Requeues:     0,
					}
				}
				if qs.Queue != nil {
					stats.Backlog = qs.Queue.Backlog
					stats.QueueLength = qs.Queue.Qlen
				}
			}

			// Get HTB-specific stats if applicable
			if htb, ok := qdisc.(*nl.Htb); ok {
				stats.HTBStats = &HTBQdiscStats{
					DirectPackets: htb.DirectPkts,
					Version:       htb.Version,
				}
			}

			return types.Success(stats)
		}
	}

	return types.Failure[DetailedQdiscStats](fmt.Errorf("qdisc %s not found on device %s", handle, device))
}

// GetDetailedClassStats returns detailed statistics for a class
func (a *RealNetlinkAdapter) GetDetailedClassStats(device valueobjects.DeviceName, handle valueobjects.Handle) types.Result[DetailedClassStats] {
	// Get the network link
	link, err := nl.LinkByName(device.String())
	if err != nil {
		return types.Failure[DetailedClassStats](fmt.Errorf("failed to find device %s: %w", device, err))
	}

	// Get all qdiscs first
	qdiscs, err := nl.QdiscList(link)
	if err != nil {
		return types.Failure[DetailedClassStats](fmt.Errorf("failed to list qdiscs: %w", err))
	}

	// Search for the class in each qdisc
	for _, qdisc := range qdiscs {
		classes, err := nl.ClassList(link, qdisc.Attrs().Handle)
		if err != nil {
			continue
		}

		for _, class := range classes {
			if class.Attrs().Handle == nl.MakeHandle(handle.Major(), handle.Minor()) {
				stats := DetailedClassStats{}

				// Get basic statistics
				if class.Attrs().Statistics != nil {
					cs := class.Attrs().Statistics
					if cs.Basic != nil {
						stats.BasicStats = ClassStats{
							BytesSent:    cs.Basic.Bytes,
							PacketsSent:  uint64(cs.Basic.Packets),
							BytesDropped: 0, // Not available in GnetStatsBasic
							Overlimits:   0, // Not available in GnetStatsBasic
						}
					}
					if cs.RateEst != nil {
						stats.BasicStats.RateBPS = uint64(cs.RateEst.Bps)
					}
					if cs.Queue != nil {
						stats.BasicStats.BacklogBytes = uint64(cs.Queue.Backlog)
						stats.BasicStats.BacklogPackets = uint64(cs.Queue.Qlen)
					}
				}

				// Get HTB-specific stats if applicable
				if htbClass, ok := class.(*nl.HtbClass); ok {
					// NOTE: HTB-specific statistics are not available in the current version
					// of vishvananda/netlink. The HtbClass struct doesn't have a Stats field.
					// To get detailed HTB statistics, you would need to use tc command or
					// update the netlink library to a version that supports these statistics.
					stats.HTBStats = &HTBClassStats{
						Lends:   0,
						Borrows: 0,
						Giants:  0,
						Tokens:  0,
						CTokens: 0,
						Rate:    htbClass.Rate,
						Ceil:    htbClass.Ceil,
						Level:   htbClass.Level,
					}
				}

				return types.Success(stats)
			}
		}
	}

	return types.Failure[DetailedClassStats](fmt.Errorf("class %s not found on device %s", handle, device))
}
