package netlink

import (
	"fmt"
	
	nl "github.com/vishvananda/netlink"
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

// AddQdisc adds a qdisc using netlink
func (a *RealNetlinkAdapter) AddQdisc(device valueobjects.DeviceName, config QdiscConfig) types.Result[Unit] {
	logger := a.logger.WithDevice(device.String()).WithOperation(logging.OperationCreateQdisc)
	logger.Info("Adding qdisc", 
		logging.String("qdisc_type", string(config.Type)),
		logging.String("handle", config.Handle.String()),
	)
	
	// Get the network link
	link, err := nl.LinkByName(device.String())
	if err != nil {
		logger.Error("Failed to find network device", logging.Error(err))
		return types.Failure[Unit](fmt.Errorf("failed to find device %s: %w", device, err))
	}
	
	// Create qdisc based on type
	var qdisc nl.Qdisc
	
	switch config.Type {
	case entities.QdiscTypeHTB:
		htb := nl.NewHtb(nl.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    nl.MakeHandle(config.Handle.Major(), config.Handle.Minor()),
			Parent:    nl.HANDLE_ROOT,
		})
		
		// Set HTB-specific parameters
		if defaultClass, ok := config.Parameters["defaultClass"].(valueobjects.Handle); ok {
			htb.Defcls = uint32(defaultClass.Minor())
		}
		if r2q, ok := config.Parameters["r2q"].(uint32); ok {
			htb.Rate2Quantum = r2q
		}
		
		qdisc = htb
		
	case entities.QdiscTypeTBF:
		// TBF implementation
		tbf := &nl.Tbf{
			QdiscAttrs: nl.QdiscAttrs{
				LinkIndex: link.Attrs().Index,
				Handle:    nl.MakeHandle(config.Handle.Major(), config.Handle.Minor()),
				Parent:    nl.HANDLE_ROOT,
			},
		}
		
		// Set TBF parameters from config
		if rate, ok := config.Parameters["rate"].(valueobjects.Bandwidth); ok {
			tbf.Rate = uint64(rate.BitsPerSecond() / 8) // Convert to bytes per second
		}
		if buffer, ok := config.Parameters["buffer"].(uint32); ok {
			tbf.Buffer = buffer
		}
		if limit, ok := config.Parameters["limit"].(uint32); ok {
			tbf.Limit = limit
		}
		
		qdisc = tbf
		
	case entities.QdiscTypePRIO:
		prio := nl.NewPrio(nl.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    nl.MakeHandle(config.Handle.Major(), config.Handle.Minor()),
			Parent:    nl.HANDLE_ROOT,
		})
		
		if bands, ok := config.Parameters["bands"].(uint8); ok {
			prio.Bands = bands
		}
		
		qdisc = prio
		
	case entities.QdiscTypeFQ_CODEL:
		fqCodel := nl.NewFqCodel(nl.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    nl.MakeHandle(config.Handle.Major(), config.Handle.Minor()),
			Parent:    nl.HANDLE_ROOT,
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
		
		qdisc = fqCodel
		
	default:
		return types.Failure[Unit](fmt.Errorf("unsupported qdisc type: %s", config.Type))
	}
	
	// Handle parent if not root
	if config.Parent != nil {
		attrs := qdisc.Attrs()
		attrs.Parent = nl.MakeHandle(config.Parent.Major(), config.Parent.Minor())
	}
	
	// Add the qdisc
	if err := nl.QdiscAdd(qdisc); err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to add qdisc: %w", err))
	}
	
	return types.Success(Unit{})
}

// DeleteQdisc deletes a qdisc using netlink
func (a *RealNetlinkAdapter) DeleteQdisc(device valueobjects.DeviceName, handle valueobjects.Handle) types.Result[Unit] {
	// Get the network link
	link, err := nl.LinkByName(device.String())
	if err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to find device %s: %w", device, err))
	}
	
	// Create a generic qdisc with the handle to delete
	qdisc := &nl.GenericQdisc{
		QdiscType: "htb", // Type doesn't matter for deletion
		QdiscAttrs: nl.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    nl.MakeHandle(handle.Major(), handle.Minor()),
			Parent:    nl.HANDLE_ROOT,
		},
	}
	
	if err := nl.QdiscDel(qdisc); err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to delete qdisc: %w", err))
	}
	
	return types.Success(Unit{})
}

// GetQdiscs returns all qdiscs for a device
func (a *RealNetlinkAdapter) GetQdiscs(device valueobjects.DeviceName) types.Result[[]QdiscInfo] {
	// Get the network link
	link, err := nl.LinkByName(device.String())
	if err != nil {
		return types.Failure[[]QdiscInfo](fmt.Errorf("failed to find device %s: %w", device, err))
	}
	
	// Get all qdiscs for the link
	qdiscs, err := nl.QdiscList(link)
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
		if qdisc.Attrs().Parent != nl.HANDLE_ROOT {
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
			info.Type = entities.QdiscTypeFQ_CODEL
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
func (a *RealNetlinkAdapter) AddClass(device valueobjects.DeviceName, config ClassConfig) types.Result[Unit] {
	// Get the network link
	link, err := nl.LinkByName(device.String())
	if err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to find device %s: %w", device, err))
	}
	
	// Create class based on type
	switch config.Type {
	case entities.QdiscTypeHTB:
		htbClass := nl.NewHtbClass(nl.ClassAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    nl.MakeHandle(config.Handle.Major(), config.Handle.Minor()),
			Parent:    nl.MakeHandle(config.Parent.Major(), config.Parent.Minor()),
		}, nl.HtbClassAttrs{})
		
		// Set HTB class parameters
		if rate, ok := config.Parameters["rate"].(valueobjects.Bandwidth); ok {
			htbClass.Rate = uint64(rate.BitsPerSecond() / 8) // Convert to bytes per second
		}
		if ceil, ok := config.Parameters["ceil"].(valueobjects.Bandwidth); ok {
			htbClass.Ceil = uint64(ceil.BitsPerSecond() / 8)
		}
		if buffer, ok := config.Parameters["buffer"].(uint32); ok {
			htbClass.Buffer = buffer
		}
		if cbuffer, ok := config.Parameters["cbuffer"].(uint32); ok {
			htbClass.Cbuffer = cbuffer
		}
		
		if err := nl.ClassAdd(htbClass); err != nil {
			return types.Failure[Unit](fmt.Errorf("failed to add HTB class: %w", err))
		}
		
	default:
		return types.Failure[Unit](fmt.Errorf("unsupported class type: %s", config.Type))
	}
	
	return types.Success(Unit{})
}

// DeleteClass deletes a class using netlink
func (a *RealNetlinkAdapter) DeleteClass(device valueobjects.DeviceName, handle valueobjects.Handle) types.Result[Unit] {
	// Get the network link
	link, err := nl.LinkByName(device.String())
	if err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to find device %s: %w", device, err))
	}
	
	// Create a generic class with the handle to delete
	class := &nl.GenericClass{
		ClassType: "htb", // Type doesn't matter for deletion
		ClassAttrs: nl.ClassAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    nl.MakeHandle(handle.Major(), handle.Minor()),
		},
	}
	
	if err := nl.ClassDel(class); err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to delete class: %w", err))
	}
	
	return types.Success(Unit{})
}

// GetClasses returns all classes for a device
func (a *RealNetlinkAdapter) GetClasses(device valueobjects.DeviceName) types.Result[[]ClassInfo] {
	// Get the network link
	link, err := nl.LinkByName(device.String())
	if err != nil {
		return types.Failure[[]ClassInfo](fmt.Errorf("failed to find device %s: %w", device, err))
	}
	
	// Get all classes for the link
	// Note: netlink library doesn't have a direct ClassList function
	// We need to list qdiscs and then get classes for each
	qdiscs, err := nl.QdiscList(link)
	if err != nil {
		return types.Failure[[]ClassInfo](fmt.Errorf("failed to list qdiscs: %w", err))
	}
	
	var result []ClassInfo
	for _, qdisc := range qdiscs {
		classes, err := nl.ClassList(link, qdisc.Attrs().Handle)
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
func (a *RealNetlinkAdapter) AddFilter(device valueobjects.DeviceName, config FilterConfig) types.Result[Unit] {
	// Get the network link
	link, err := nl.LinkByName(device.String())
	if err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to find device %s: %w", device, err))
	}
	
	// Create u32 filter
	filter := &nl.U32{
		FilterAttrs: nl.FilterAttrs{
			LinkIndex: link.Attrs().Index,
			Parent:    nl.MakeHandle(config.Parent.Major(), config.Parent.Minor()),
			Priority:  config.Priority,
			Handle:    nl.MakeHandle(config.Handle.Major(), config.Handle.Minor()),
			Protocol:  convertProtocol(config.Protocol),
		},
		ClassId: nl.MakeHandle(config.FlowID.Major(), config.FlowID.Minor()),
	}
	
	// Add matches
	for _, match := range config.Matches {
		if err := addU32Match(filter, match); err != nil {
			return types.Failure[Unit](fmt.Errorf("failed to add match: %w", err))
		}
	}
	
	if err := nl.FilterAdd(filter); err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to add filter: %w", err))
	}
	
	return types.Success(Unit{})
}

// DeleteFilter deletes a filter using netlink
func (a *RealNetlinkAdapter) DeleteFilter(device valueobjects.DeviceName, parent valueobjects.Handle, priority uint16, handle valueobjects.Handle) types.Result[Unit] {
	// Get the network link
	link, err := nl.LinkByName(device.String())
	if err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to find device %s: %w", device, err))
	}
	
	// Create filter to delete
	filter := &nl.U32{
		FilterAttrs: nl.FilterAttrs{
			LinkIndex: link.Attrs().Index,
			Parent:    nl.MakeHandle(parent.Major(), parent.Minor()),
			Priority:  priority,
			Handle:    nl.MakeHandle(handle.Major(), handle.Minor()),
		},
	}
	
	if err := nl.FilterDel(filter); err != nil {
		return types.Failure[Unit](fmt.Errorf("failed to delete filter: %w", err))
	}
	
	return types.Success(Unit{})
}

// GetFilters returns all filters for a device
func (a *RealNetlinkAdapter) GetFilters(device valueobjects.DeviceName) types.Result[[]FilterInfo] {
	// Get the network link
	link, err := nl.LinkByName(device.String())
	if err != nil {
		return types.Failure[[]FilterInfo](fmt.Errorf("failed to find device %s: %w", device, err))
	}
	
	// Get all qdiscs first
	qdiscs, err := nl.QdiscList(link)
	if err != nil {
		return types.Failure[[]FilterInfo](fmt.Errorf("failed to list qdiscs: %w", err))
	}
	
	var result []FilterInfo
	
	// Get filters for each qdisc
	for _, qdisc := range qdiscs {
		filters, err := nl.FilterList(link, qdisc.Attrs().Handle)
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
			if u32, ok := filter.(*nl.U32); ok {
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

func addU32Match(filter *nl.U32, match FilterMatch) error {
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
			selector := &nl.TcU32Sel{
				Flags: 0,
				Nkeys: 1,
				Keys: []nl.TcU32Key{
					{
						Mask: 0xffffffff, // Match all 32 bits
						Val:  ipAddr,     // IP address in network byte order
						Off:  16,         // Offset 16 bytes into IP header (destination IP)
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
			
			selector := &nl.TcU32Sel{
				Flags: 0,
				Nkeys: 1,
				Keys: []nl.TcU32Key{
					{
						Mask: 0xffffffff,
						Val:  ipAddr,
						Off:  12, // Source IP offset
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
			selector := &nl.TcU32Sel{
				Flags: 0,
				Nkeys: 1,
				Keys: []nl.TcU32Key{
					{
						Mask: 0x0000ffff, // Match 16 bits for port
						Val:  uint32(port) << 16, // Port in high 16 bits
						Off:  22, // TCP/UDP destination port offset (20 byte IP + 2 byte offset)
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
			selector := &nl.TcU32Sel{
				Flags: 0,
				Nkeys: 1,
				Keys: []nl.TcU32Key{
					{
						Mask: 0xffff0000, // Match high 16 bits for source port
						Val:  uint32(port) << 16,
						Off:  20, // TCP/UDP source port offset
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
			selector := &nl.TcU32Sel{
				Flags: 0,
				Nkeys: 1,
				Keys: []nl.TcU32Key{
					{
						Mask: 0x000000ff, // Match 8 bits for protocol
						Val:  uint32(protocol),
						Off:  9, // Protocol field offset in IP header
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