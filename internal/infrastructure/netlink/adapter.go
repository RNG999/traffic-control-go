//go:build linux
// +build linux

package netlink

import (
	"context"
	"fmt"
	"syscall"

	"github.com/vishvananda/netlink"

	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/pkg/logging"
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

// createHTBQdisc creates an HTB qdisc
func (a *RealNetlinkAdapter) createHTBQdisc(link netlink.Link, config QdiscConfig) netlink.Qdisc {
	htb := netlink.NewHtb(netlink.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    netlink.MakeHandle(config.Handle.Major(), config.Handle.Minor()),
		Parent:    netlink.HANDLE_ROOT,
	})

	// Set HTB-specific parameters
	if defaultClass, ok := config.Parameters["defaultClass"].(valueobjects.Handle); ok {
		htb.Defcls = uint32(defaultClass.Minor())
	}
	if r2q, ok := config.Parameters["r2q"].(uint32); ok {
		htb.Rate2Quantum = r2q
	}

	return htb
}

// createTBFQdisc creates a TBF qdisc
func (a *RealNetlinkAdapter) createTBFQdisc(link netlink.Link, config QdiscConfig) netlink.Qdisc {
	tbf := &netlink.Tbf{
		QdiscAttrs: netlink.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    netlink.MakeHandle(config.Handle.Major(), config.Handle.Minor()),
			Parent:    netlink.HANDLE_ROOT,
		},
	}

	// Set TBF parameters from config
	if rate, ok := config.Parameters["rate"].(valueobjects.Bandwidth); ok {
		tbf.Rate = uint64(rate.BitsPerSecond()) / 8 // Convert to bytes per second
	}
	if buffer, ok := config.Parameters["buffer"].(uint32); ok {
		tbf.Buffer = buffer
	}
	if limit, ok := config.Parameters["limit"].(uint32); ok {
		tbf.Limit = limit
	}

	return tbf
}

// createPRIOQdisc creates a PRIO qdisc
func (a *RealNetlinkAdapter) createPRIOQdisc(link netlink.Link, config QdiscConfig) netlink.Qdisc {
	prio := netlink.NewPrio(netlink.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    netlink.MakeHandle(config.Handle.Major(), config.Handle.Minor()),
		Parent:    netlink.HANDLE_ROOT,
	})

	if bands, ok := config.Parameters["bands"].(uint8); ok {
		prio.Bands = bands
	}

	return prio
}

// createFQCODELQdisc creates a FQ_CODEL qdisc
func (a *RealNetlinkAdapter) createFQCODELQdisc(link netlink.Link, config QdiscConfig) netlink.Qdisc {
	fqCodel := netlink.NewFqCodel(netlink.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    netlink.MakeHandle(config.Handle.Major(), config.Handle.Minor()),
		Parent:    netlink.HANDLE_ROOT,
	})

	if limit, ok := config.Parameters["limit"].(uint32); ok {
		fqCodel.Limit = limit
	}
	if target, ok := config.Parameters["target"].(uint32); ok {
		fqCodel.Target = target
	}
	if interval, ok := config.Parameters["interval"].(uint32); ok {
		fqCodel.Interval = interval
	}

	return fqCodel
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

	// Create qdisc based on type
	var qdisc netlink.Qdisc

	// For now, assume all qdiscs are HTB qdiscs since that's what we're using
	// This is a simplification - in a full implementation we'd check the actual type
	qdisc = &netlink.Htb{
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
func (a *RealNetlinkAdapter) DeleteQdisc(device valueobjects.DeviceName, handle valueobjects.Handle) types.Result[Unit] {
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
func (a *RealNetlinkAdapter) GetQdiscs(device valueobjects.DeviceName) types.Result[[]QdiscInfo] {
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
			Handle: valueobjects.HandleFromUint32(qdisc.Attrs().Handle),
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
			parent := valueobjects.HandleFromUint32(qdisc.Attrs().Parent)
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
		nlClass.Buffer = class.Burst()
		nlClass.Cbuffer = class.Cburst()

		a.logger.Debug("HTB class parameters",
			logging.Int("rate", int(nlClass.Rate)),
			logging.Int("ceil", int(nlClass.Ceil)),
			logging.Int("buffer", int(nlClass.Buffer)),
			logging.Int("cbuffer", int(nlClass.Cbuffer)),
		)

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
func (a *RealNetlinkAdapter) DeleteClass(device valueobjects.DeviceName, handle valueobjects.Handle) types.Result[Unit] {
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
func (a *RealNetlinkAdapter) GetClasses(device valueobjects.DeviceName) types.Result[[]ClassInfo] {
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
				Handle: valueobjects.HandleFromUint32(class.Attrs().Handle),
				Parent: valueobjects.HandleFromUint32(class.Attrs().Parent),
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

	// Create simple u32 filter with basic match-all configuration
	filter := &netlink.U32{
		FilterAttrs: netlink.FilterAttrs{
			LinkIndex: link.Attrs().Index,
			Parent:    netlink.MakeHandle(filterEntity.ID().Parent().Major(), filterEntity.ID().Parent().Minor()),
			Priority:  filterEntity.ID().Priority(),
			Protocol:  syscall.ETH_P_IP,
		},
		ClassId: netlink.MakeHandle(filterEntity.FlowID().Major(), filterEntity.FlowID().Minor()),
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
func (a *RealNetlinkAdapter) DeleteFilter(device valueobjects.DeviceName, parent valueobjects.Handle, priority uint16, handle valueobjects.Handle) types.Result[Unit] {
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
func (a *RealNetlinkAdapter) GetFilters(device valueobjects.DeviceName) types.Result[[]FilterInfo] {
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
				Parent:   valueobjects.HandleFromUint32(filter.Attrs().Parent),
				Priority: filter.Attrs().Priority,
				Handle:   valueobjects.HandleFromUint32(filter.Attrs().Handle),
				Protocol: convertProtocolBack(filter.Attrs().Protocol),
			}

			// Handle U32 filters
			if u32, ok := filter.(*netlink.U32); ok {
				info.FlowID = valueobjects.HandleFromUint32(u32.ClassId)
				// Extract matches - this is simplified
				// Real implementation would need to parse U32 sel
			}

			result = append(result, info)
		}
	}

	return types.Success(result)
}

// Helper functions

func convertProtocol(p entities.Protocol) uint16 {
	switch p {
	case entities.ProtocolAll:
		return 0x0000
	case entities.ProtocolIP:
		return 0x0800
	case entities.ProtocolIPv6:
		return 0x86DD
	default:
		return 0x0800 // Default to IP
	}
}

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

func addU32Match(filter *netlink.U32, match FilterMatch) error {
	// U32 filter selectors are complex but follow specific patterns
	// Each selector matches specific bits at specific offsets in the packet

	switch match.Type {
	case entities.MatchTypeIPDestination:
		// Match destination IP (IPv4) at offset 16 in IP header
		if ipStr, ok := match.Value.(string); ok {
			ipAddr, err := parseIPAddress(ipStr)
			if err != nil {
				return fmt.Errorf("invalid IP address: %w", err)
			}

			// Create U32 selector for destination IP
			selector := &netlink.TcU32Sel{
				Flags: 0,
				Nkeys: 1,
				Keys: []netlink.TcU32Key{
					{
						Mask:    0xffffffff, // Match all 32 bits
						Val:     ipAddr,     // IP address in network byte order
						Off:     16,         // Offset 16 bytes into IP header (destination IP)
						OffMask: 0,
					},
				},
			}
			filter.Sel = selector
		} else {
			return fmt.Errorf("IP destination match requires string value")
		}

	case entities.MatchTypeIPSource:
		// Match source IP (IPv4) at offset 12 in IP header
		if ipStr, ok := match.Value.(string); ok {
			ipAddr, err := parseIPAddress(ipStr)
			if err != nil {
				return fmt.Errorf("invalid IP address: %w", err)
			}

			selector := &netlink.TcU32Sel{
				Flags: 0,
				Nkeys: 1,
				Keys: []netlink.TcU32Key{
					{
						Mask:    0xffffffff,
						Val:     ipAddr,
						Off:     12, // Source IP offset
						OffMask: 0,
					},
				},
			}
			filter.Sel = selector
		} else {
			return fmt.Errorf("IP source match requires string value")
		}

	case entities.MatchTypePortDestination:
		// Match destination port - requires protocol awareness
		// For TCP/UDP, port is at different offsets depending on IP header length
		if port, ok := match.Value.(uint16); ok {
			// This is a simplified implementation for standard 20-byte IP header
			// Real implementation would need to handle variable IP header length
			selector := &netlink.TcU32Sel{
				Flags: 0,
				Nkeys: 1,
				Keys: []netlink.TcU32Key{
					{
						Mask:    0x0000ffff,         // Match 16 bits for port
						Val:     uint32(port) << 16, // Port in high 16 bits
						Off:     22,                 // TCP/UDP destination port offset (20 byte IP + 2 byte offset)
						OffMask: 0,
					},
				},
			}
			filter.Sel = selector
		} else {
			return fmt.Errorf("port destination match requires uint16 value")
		}

	case entities.MatchTypePortSource:
		// Match source port
		if port, ok := match.Value.(uint16); ok {
			selector := &netlink.TcU32Sel{
				Flags: 0,
				Nkeys: 1,
				Keys: []netlink.TcU32Key{
					{
						Mask:    0xffff0000, // Match high 16 bits for source port
						Val:     uint32(port) << 16,
						Off:     20, // TCP/UDP source port offset
						OffMask: 0,
					},
				},
			}
			filter.Sel = selector
		} else {
			return fmt.Errorf("port source match requires uint16 value")
		}

	case entities.MatchTypeProtocol:
		// Match IP protocol field at offset 9 in IP header
		if protocol, ok := match.Value.(uint8); ok {
			selector := &netlink.TcU32Sel{
				Flags: 0,
				Nkeys: 1,
				Keys: []netlink.TcU32Key{
					{
						Mask:    0x000000ff, // Match 8 bits for protocol
						Val:     uint32(protocol),
						Off:     9, // Protocol field offset in IP header
						OffMask: 0,
					},
				},
			}
			filter.Sel = selector
		} else {
			return fmt.Errorf("protocol match requires uint8 value")
		}

	default:
		return fmt.Errorf("unsupported match type: %v", match.Type)
	}

	return nil
}

// parseIPAddress converts IP string to uint32 in network byte order
func parseIPAddress(ipStr string) (uint32, error) {
	parts := make([]uint32, 4)

	// Simple IP parsing - production code would use net.ParseIP
	ipParts := splitString(ipStr, ".")
	if len(ipParts) != 4 {
		return 0, fmt.Errorf("invalid IP format")
	}

	for i, part := range ipParts {
		var val uint64
		// Manual parsing to avoid external dependencies
		for _, char := range part {
			if char < '0' || char > '9' {
				return 0, fmt.Errorf("invalid IP octet")
			}
			val = val*10 + uint64(char-'0')
		}
		if val > 255 {
			return 0, fmt.Errorf("IP octet out of range")
		}
		parts[i] = uint32(val)
	}

	// Convert to network byte order (big-endian)
	return (parts[0] << 24) | (parts[1] << 16) | (parts[2] << 8) | parts[3], nil
}

// splitString splits string by delimiter (avoiding strings package)
func splitString(s, delimiter string) []string {
	if len(s) == 0 {
		return []string{}
	}

	var result []string
	start := 0

	for i := 0; i <= len(s)-len(delimiter); i++ {
		if s[i:i+len(delimiter)] == delimiter {
			result = append(result, s[start:i])
			start = i + len(delimiter)
			i += len(delimiter) - 1
		}
	}

	// Add the last part
	result = append(result, s[start:])
	return result
}
