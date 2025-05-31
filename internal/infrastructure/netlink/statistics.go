package netlink

import (
	"fmt"
	
	nl "github.com/vishvananda/netlink"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/pkg/types"
)

// DetailedQdiscStats represents detailed qdisc statistics
type DetailedQdiscStats struct {
	BasicStats QdiscStats
	// Queue information
	QueueLength   uint32
	Backlog       uint32
	BacklogBytes  uint64
	// Rate information
	BytesPerSecond   uint64
	PacketsPerSecond uint64
	// HTB specific
	HTBStats *HTBQdiscStats
}

// HTBQdiscStats represents HTB-specific statistics
type HTBQdiscStats struct {
	DirectPackets uint32
	Version       uint32
}

// DetailedClassStats represents detailed class statistics
type DetailedClassStats struct {
	BasicStats ClassStats
	// HTB specific
	HTBStats *HTBClassStats
}

// HTBClassStats represents HTB class-specific statistics
type HTBClassStats struct {
	Lends    uint32
	Borrows  uint32
	Giants   uint32
	Tokens   uint32
	CTokens  uint32
	Rate     uint64
	Ceil     uint64
	Level    uint32
}

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
				if bs, ok := qdisc.Attrs().Statistics.(*nl.QdiscStatistics); ok {
					stats.BasicStats = QdiscStats{
						BytesSent:    bs.Bytes,
						PacketsSent:  bs.Packets,
						BytesDropped: bs.Drops,
						Overlimits:   bs.Overlimits,
						Requeues:     bs.Requeues,
					}
					stats.Backlog = bs.Backlog
					stats.QueueLength = bs.Qlen
				}
			}
			
			// Get HTB-specific stats if applicable
			if htb, ok := qdisc.(*nl.Htb); ok {
				stats.HTBStats = &HTBQdiscStats{
					DirectPackets: htb.DirectPkts,
					Version:      htb.Version,
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
					if cs, ok := class.Attrs().Statistics.(*nl.ClassStatistics); ok {
						stats.BasicStats = ClassStats{
							BytesSent:      cs.Basic.Bytes,
							PacketsSent:    cs.Basic.Packets,
							BytesDropped:   cs.Basic.Drops,
							Overlimits:     cs.Basic.Overlimits,
							RateBPS:        cs.RateEst.Bps,
							BacklogBytes:   cs.Queue.Backlog,
							BacklogPackets: cs.Queue.Qlen,
						}
					}
				}
				
				// Get HTB-specific stats if applicable
				if htbClass, ok := class.(*nl.HtbClass); ok {
					stats.HTBStats = &HTBClassStats{
						Lends:   htbClass.Stats.Lends,
						Borrows: htbClass.Stats.Borrows,
						Giants:  htbClass.Stats.Giants,
						Tokens:  htbClass.Stats.Tokens,
						CTokens: htbClass.Stats.Ctokens,
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