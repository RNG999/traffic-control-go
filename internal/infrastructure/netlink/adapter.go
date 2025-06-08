//go:build linux
// +build linux

package netlink

import (
	"context"
	"fmt"
	"syscall"

	"github.com/vishvananda/netlink"

	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/pkg/logging"
	"github.com/rng999/traffic-control-go/pkg/tc"
	"github.com/rng999/traffic-control-go/pkg/types"
)

// RealNetlinkAdapter is the real implementation using netlink library
type RealNetlinkAdapter struct {
	logger logging.Logger
}

// NewRealNetlinkAdapter creates a new real netlink adapter
func NewRealNetlinkAdapter() *RealNetlinkAdapter {
	logger := logging.WithComponent(logging.ComponentNetlink)
	logger.Info("Initializing real netlink adapter")

	return &RealNetlinkAdapter{
		logger: logger,
	}
}

// AddQdisc adds a qdisc using netlink
func (a *RealNetlinkAdapter) AddQdisc(ctx context.Context, qdiscEntity *entities.Qdisc) error {
	a.logger.Info("Adding qdisc",
		logging.String("device", qdiscEntity.Device().String()),
		logging.String("qdisc_type", qdiscEntity.Type().String()),
		logging.String("handle", qdiscEntity.Handle().String()),
		logging.String("operation", logging.OperationCreateQdisc),
	)

	// Get the network link
	link, err := netlink.LinkByName(qdiscEntity.Device().String())
	if err != nil {
		return fmt.Errorf("failed to find device %s: %w", qdiscEntity.Device(), err)
	}

	// Create HTB qdisc
	qdisc := &netlink.Htb{
		QdiscAttrs: netlink.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    netlink.MakeHandle(qdiscEntity.Handle().Major(), qdiscEntity.Handle().Minor()),
			Parent:    netlink.HANDLE_ROOT,
		},
		Version:      3,
		Rate2Quantum: 10,
		Defcls:       0, // Will be set by the HTB configuration
	}

	// Handle parent if not root
	if qdiscEntity.Parent() != nil {
		attrs := qdisc.Attrs()
		attrs.Parent = netlink.MakeHandle(qdiscEntity.Parent().Major(), qdiscEntity.Parent().Minor())
	}

	// Add the qdisc
	if err := netlink.QdiscAdd(qdisc); err != nil {
		return fmt.Errorf("failed to add qdisc: %w", err)
	}

	a.logger.Info("Qdisc added successfully",
		logging.String("handle", qdiscEntity.Handle().String()),
		logging.String("type", qdiscEntity.Type().String()),
	)

	return nil
}

// DeleteQdisc deletes a qdisc using netlink
func (a *RealNetlinkAdapter) DeleteQdisc(device tc.DeviceName, handle tc.Handle) types.Result[Unit] {
	// Get the network link
	link, err := netlink.LinkByName(device.String())
	if err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to find device %s: %w", device, err))
	}

	// Create a generic qdisc with the handle to delete
	qdisc := &netlink.GenericQdisc{
		QdiscType: "htb", // Type doesn't matter for deletion
		QdiscAttrs: netlink.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    netlink.MakeHandle(handle.Major(), handle.Minor()),
			Parent:    netlink.HANDLE_ROOT,
		},
	}

	if err := netlink.QdiscDel(qdisc); err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to delete qdisc: %w", err))
	}

	return types.Success(Unit{})
}

// GetQdiscs returns all qdiscs for a device
func (a *RealNetlinkAdapter) GetQdiscs(device tc.DeviceName) types.Result[[]QdiscInfo] {
	// Get the network link
	link, err := netlink.LinkByName(device.String())
	if err != nil {
		return types.Failure[[]QdiscInfo](fmt.Errorf("failed to find device %s: %w", device, err))
	}

	// Get all qdiscs for the link
	qdiscs, err := netlink.QdiscList(link)
	if err != nil {
		return types.Failure[[]QdiscInfo](fmt.Errorf("failed to list qdiscs: %w", err))
	}

	// Convert to our domain types
	var result []QdiscInfo
	for _, qdisc := range qdiscs {
		info := QdiscInfo{
			Handle: tc.HandleFromUint32(qdisc.Attrs().Handle),
			Statistics: QdiscStats{
				BytesSent:    0, // Actual stats would be retrieved differently
				PacketsSent:  0, // Statistics struct differs across netlink versions
				BytesDropped: 0,
				Overlimits:   0,
				Requeues:     0,
			},
		}

		// Set parent if not root
		if qdisc.Attrs().Parent != netlink.HANDLE_ROOT {
			parent := tc.HandleFromUint32(qdisc.Attrs().Parent)
			info.Parent = &parent
		}

		// Determine type
		switch qdisc.Type() {
		case "htb":
			info.Type = entities.QdiscTypeHTB
		case "tbf":
			info.Type = entities.QdiscTypeTBF
		case "prio":
			info.Type = entities.QdiscTypePRIO
		case "fq_codel":
			info.Type = entities.QdiscTypeFQCODEL
		case "sfq":
			info.Type = entities.QdiscTypeSFQ
		case "cake":
			info.Type = entities.QdiscTypeCAKE
		}

		result = append(result, info)
	}

	return types.Success(result)
}

// AddClass adds a class using netlink
func (a *RealNetlinkAdapter) AddClass(ctx context.Context, classEntity interface{}) error {
	switch class := classEntity.(type) {
	case *entities.Class:
		a.logger.Info("Adding class",
			logging.String("device", class.ID().Device().String()),
			logging.String("operation", logging.OperationCreateClass),
		)

		// For basic classes, we'd need to implement specific logic
		// This is a simplified implementation
		return fmt.Errorf("basic class creation not implemented")

	case *entities.HTBClass:
		a.logger.Info("Adding HTB class",
			logging.String("device", class.ID().Device().String()),
			logging.String("operation", logging.OperationCreateClass),
		)

		// Get the network link
		link, err := netlink.LinkByName(class.ID().Device().String())
		if err != nil {
			return fmt.Errorf("failed to find device %s: %w", class.ID().Device(), err)
		}

		// Create netlink HTB class
		nlClass := netlink.NewHtbClass(netlink.ClassAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    netlink.MakeHandle(class.Handle().Major(), class.Handle().Minor()),
			Parent:    netlink.MakeHandle(class.Parent().Major(), class.Parent().Minor()),
		}, netlink.HtbClassAttrs{})

		// Set HTB class parameters
		nlClass.Rate = uint64(class.Rate().BitsPerSecond()) / 8 // Convert to bytes per second
		nlClass.Ceil = uint64(class.Ceil().BitsPerSecond()) / 8
		
		// Set burst parameters - use enhanced calculation if available
		if class.Burst() > 0 {
			nlClass.Buffer = class.Burst()
		} else {
			nlClass.Buffer = class.CalculateEnhancedBurst()
		}
		
		if class.Cburst() > 0 {
			nlClass.Cbuffer = class.Cburst()  
		} else {
			nlClass.Cbuffer = class.CalculateEnhancedCburst()
		}

		// Set enhanced HTB parameters if available
		if class.Quantum() > 0 {
			nlClass.Quantum = class.Quantum()
		} else {
			nlClass.Quantum = class.CalculateQuantum()
		}

		// Note: Advanced parameters (Overhead, MPU, MTU) are not supported by the current netlink library version
		// These are tracked in the domain model but not applied via netlink for now
		
		// Set HTB priority if specified and supported
		if class.HTBPrio() > 0 {
			nlClass.Prio = class.HTBPrio()
		}

		a.logger.Debug("HTB class parameters",
			logging.String("rate", fmt.Sprintf("%d", nlClass.Rate)),
			logging.String("ceil", fmt.Sprintf("%d", nlClass.Ceil)),
			logging.String("buffer", fmt.Sprintf("%d", nlClass.Buffer)),
			logging.String("cbuffer", fmt.Sprintf("%d", nlClass.Cbuffer)),
			logging.String("quantum", fmt.Sprintf("%d", nlClass.Quantum)),
			logging.String("prio", fmt.Sprintf("%d", nlClass.Prio)),
		)
		
		// Log advanced parameters for debugging (domain model only)
		if class.Overhead() > 0 || class.MPU() > 0 || class.MTU() > 0 {
			a.logger.Debug("Advanced HTB parameters (domain model only)",
				logging.String("overhead", fmt.Sprintf("%d", class.Overhead())),
				logging.String("mpu", fmt.Sprintf("%d", class.MPU())),
				logging.String("mtu", fmt.Sprintf("%d", class.MTU())),
			)
		}

		if err := netlink.ClassAdd(nlClass); err != nil {
			return fmt.Errorf("failed to add HTB class: %w", err)
		}

		a.logger.Info("HTB class added successfully",
			logging.String("handle", class.Handle().String()),
			logging.String("parent", class.Parent().String()),
		)

		return nil

	default:
		return fmt.Errorf("unsupported class type: %T", classEntity)
	}
}

// DeleteClass deletes a class using netlink
func (a *RealNetlinkAdapter) DeleteClass(device tc.DeviceName, handle tc.Handle) types.Result[Unit] {
	// Get the network link
	link, err := netlink.LinkByName(device.String())
	if err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to find device %s: %w", device, err))
	}

	// Create a generic class with the handle to delete
	class := &netlink.GenericClass{
		ClassType: "htb", // Type doesn't matter for deletion
		ClassAttrs: netlink.ClassAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    netlink.MakeHandle(handle.Major(), handle.Minor()),
		},
	}

	if err := netlink.ClassDel(class); err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to delete class: %w", err))
	}

	return types.Success(Unit{})
}

// GetClasses returns all classes for a device
func (a *RealNetlinkAdapter) GetClasses(device tc.DeviceName) types.Result[[]ClassInfo] {
	// Get the network link
	link, err := netlink.LinkByName(device.String())
	if err != nil {
		return types.Failure[[]ClassInfo](fmt.Errorf("failed to find device %s: %w", device, err))
	}

	// Get all classes for the link
	// Note: netlink library doesn't have a direct ClassList function
	// We need to list qdiscs and then get classes for each
	qdiscs, err := netlink.QdiscList(link)
	if err != nil {
		return types.Failure[[]ClassInfo](fmt.Errorf("failed to list qdiscs: %w", err))
	}

	var result []ClassInfo
	for _, qdisc := range qdiscs {
		classes, err := netlink.ClassList(link, qdisc.Attrs().Handle)
		if err != nil {
			continue // Skip if we can't get classes for this qdisc
		}

		for _, class := range classes {
			info := ClassInfo{
				Handle: tc.HandleFromUint32(class.Attrs().Handle),
				Parent: tc.HandleFromUint32(class.Attrs().Parent),
				Statistics: ClassStats{
					BytesSent:      0, // Actual stats would be retrieved differently
					PacketsSent:    0, // Statistics struct differs across netlink versions
					BytesDropped:   0,
					Overlimits:     0,
					RateBPS:        0,
					BacklogBytes:   0,
					BacklogPackets: 0,
				},
			}

			// Determine type based on class type
			switch class.Type() {
			case "htb":
				info.Type = entities.QdiscTypeHTB
			}

			result = append(result, info)
		}
	}

	return types.Success(result)
}

// AddFilter adds a filter using netlink
func (a *RealNetlinkAdapter) AddFilter(ctx context.Context, filterEntity *entities.Filter) error {
	a.logger.Info("Adding filter",
		logging.String("device", filterEntity.ID().Device().String()),
		logging.String("operation", logging.OperationCreateFilter),
	)

	// Get the network link
	link, err := netlink.LinkByName(filterEntity.ID().Device().String())
	if err != nil {
		return fmt.Errorf("failed to find device %s: %w", filterEntity.ID().Device(), err)
	}

	// Create u32 filter with match conditions
	filter := &netlink.U32{
		FilterAttrs: netlink.FilterAttrs{
			LinkIndex: link.Attrs().Index,
			Parent:    netlink.MakeHandle(filterEntity.ID().Parent().Major(), filterEntity.ID().Parent().Minor()),
			Priority:  filterEntity.ID().Priority(),
			Protocol:  syscall.ETH_P_IP,
		},
		ClassId: netlink.MakeHandle(filterEntity.FlowID().Major(), filterEntity.FlowID().Minor()),
	}

	// Configure match conditions based on filter matches
	if err := a.configureU32Matches(filter, filterEntity.Matches()); err != nil {
		return fmt.Errorf("failed to configure filter matches: %w", err)
	}

	a.logger.Debug("Filter configuration",
		logging.String("parent", filterEntity.ID().Parent().String()),
		logging.String("handle", filterEntity.ID().Handle().String()),
		logging.String("flow_id", filterEntity.FlowID().String()),
		logging.Int("priority", int(filterEntity.ID().Priority())),
	)

	if err := netlink.FilterAdd(filter); err != nil {
		return fmt.Errorf("failed to add filter: %w", err)
	}

	a.logger.Info("Filter added successfully",
		logging.String("handle", filterEntity.ID().Handle().String()),
		logging.String("flow_id", filterEntity.FlowID().String()),
	)

	return nil
}

// DeleteFilter deletes a filter using netlink
func (a *RealNetlinkAdapter) DeleteFilter(device tc.DeviceName, parent tc.Handle, priority uint16, handle tc.Handle) types.Result[Unit] {
	// Get the network link
	link, err := netlink.LinkByName(device.String())
	if err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to find device %s: %w", device, err))
	}

	// Create filter to delete
	filter := &netlink.U32{
		FilterAttrs: netlink.FilterAttrs{
			LinkIndex: link.Attrs().Index,
			Parent:    netlink.MakeHandle(parent.Major(), parent.Minor()),
			Priority:  priority,
			Handle:    netlink.MakeHandle(handle.Major(), handle.Minor()),
		},
	}

	if err := netlink.FilterDel(filter); err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to delete filter: %w", err))
	}

	return types.Success(Unit{})
}

// GetFilters returns all filters for a device
func (a *RealNetlinkAdapter) GetFilters(device tc.DeviceName) types.Result[[]FilterInfo] {
	// Get the network link
	link, err := netlink.LinkByName(device.String())
	if err != nil {
		return types.Failure[[]FilterInfo](fmt.Errorf("failed to find device %s: %w", device, err))
	}

	// Get all qdiscs first
	qdiscs, err := netlink.QdiscList(link)
	if err != nil {
		return types.Failure[[]FilterInfo](fmt.Errorf("failed to list qdiscs: %w", err))
	}

	var result []FilterInfo

	// Get filters for each qdisc
	for _, qdisc := range qdiscs {
		filters, err := netlink.FilterList(link, qdisc.Attrs().Handle)
		if err != nil {
			continue
		}

		for _, filter := range filters {
			info := FilterInfo{
				Parent:   tc.HandleFromUint32(filter.Attrs().Parent),
				Priority: filter.Attrs().Priority,
				Handle:   tc.HandleFromUint32(filter.Attrs().Handle),
				Protocol: convertProtocolBack(filter.Attrs().Protocol),
			}

			// Handle U32 filters
			if u32, ok := filter.(*netlink.U32); ok {
				info.FlowID = tc.HandleFromUint32(u32.ClassId)
				// Extract matches - this is simplified
				// Real implementation would need to parse U32 sel
			}

			result = append(result, info)
		}
	}

	return types.Success(result)
}

// Helper functions

func convertProtocolBack(p uint16) entities.Protocol {
	switch p {
	case 0x0000:
		return entities.ProtocolAll
	case 0x0800:
		return entities.ProtocolIP
	case 0x86DD:
		return entities.ProtocolIPv6
	default:
		return entities.ProtocolIP
	}
}

// configureU32Matches configures U32 filter match conditions
func (a *RealNetlinkAdapter) configureU32Matches(filter *netlink.U32, matches []entities.Match) error {
	if len(matches) == 0 {
		// No match conditions - create a match-all filter
		return nil
	}

	// For now, we'll implement port matching which is the most common case
	// U32 filters use selectors to match fields in the packet
	for _, match := range matches {
		switch match.Type() {
		case entities.MatchTypePortDestination:
			if portMatch, ok := match.(*entities.PortMatch); ok {
				// For destination port matching in TCP/UDP:
				// The port is at offset 2 in the TCP/UDP header
				// TCP/UDP header starts at IP header length (variable)
				// For simplicity, assume standard 20-byte IP header (offset 20)
				// Destination port is at bytes 22-23 (offset 22)

				port := portMatch.Port()

				// Create U32 selector for destination port
				// This matches the destination port field in TCP/UDP header
				sel := &netlink.TcU32Sel{
					Flags:    0,
					Offshift: 0,
					Nkeys:    1,
					Offmask:  0,
					Off:      0,
					Offoff:   0,
					Hoff:     0,
					Hmask:    0,
				}

				// Configure the key to match destination port
				// Key matches 2 bytes at offset 22 (destination port in TCP/UDP)
				key := netlink.TcU32Key{
					Mask:    0x0000ffff,   // Match 2 bytes (port) in lower 16 bits
					Val:     uint32(port), // Port value
					Off:     22,           // Offset 22 for destination port in TCP/UDP
					OffMask: 0,
				}

				sel.Keys = []netlink.TcU32Key{key}
				filter.Sel = sel

				a.logger.Debug("Configured destination port match",
					logging.Int("port", int(port)),
					logging.String("mask", fmt.Sprintf("0x%08x", key.Mask)),
					logging.String("val", fmt.Sprintf("0x%08x", key.Val)),
				)
			}
		case entities.MatchTypePortSource:
			if portMatch, ok := match.(*entities.PortMatch); ok {
				// Source port is at offset 20 in TCP/UDP header (after IP header)
				port := portMatch.Port()

				sel := &netlink.TcU32Sel{
					Flags:    0,
					Offshift: 0,
					Nkeys:    1,
					Offmask:  0,
					Off:      0,
					Offoff:   0,
					Hoff:     0,
					Hmask:    0,
				}

				key := netlink.TcU32Key{
					Mask:    0xffff0000,         // Match 2 bytes (port) at high bits
					Val:     uint32(port) << 16, // Port value shifted for high bits
					Off:     20,                 // Offset 20 for source port in TCP/UDP
					OffMask: 0,
				}

				sel.Keys = []netlink.TcU32Key{key}
				filter.Sel = sel

				a.logger.Debug("Configured source port match",
					logging.Int("port", int(port)),
				)
			}
		default:
			// For now, skip other match types (IP addresses, etc.)
			// They can be implemented later as needed
			a.logger.Debug("Skipping unsupported match type",
				logging.String("type", fmt.Sprintf("%v", match.Type())),
			)
		}
	}

	return nil
}
