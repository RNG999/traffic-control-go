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
		nlClass.Buffer = class.Burst()
		nlClass.Cbuffer = class.Cburst()

		a.logger.Debug("HTB class parameters",
			logging.String("rate", fmt.Sprintf("%d", nlClass.Rate)),
			logging.String("ceil", fmt.Sprintf("%d", nlClass.Ceil)),
			logging.String("buffer", fmt.Sprintf("%d", nlClass.Buffer)),
			logging.String("cbuffer", fmt.Sprintf("%d", nlClass.Cbuffer)),
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
