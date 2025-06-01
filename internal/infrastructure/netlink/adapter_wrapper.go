package netlink

import (
	"context"
	"fmt"

	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/pkg/logging"
	"github.com/rng999/traffic-control-go/pkg/types"
)

// AdapterWrapper wraps the RealNetlinkAdapter to implement the new interface
type AdapterWrapper struct {
	adapter *RealNetlinkAdapter
	logger  logging.Logger
}

// NewAdapter creates a new netlink adapter
func NewAdapter() Adapter {
	return &AdapterWrapper{
		adapter: NewRealNetlinkAdapter(),
		logger:  logging.WithComponent(logging.ComponentNetlink),
	}
}

// AddQdisc adds a qdisc from domain entity
func (a *AdapterWrapper) AddQdisc(ctx context.Context, qdisc *entities.Qdisc) error {
	config := QdiscConfig{
		Handle:     qdisc.Handle(),
		Type:       qdisc.Type(),
		Parameters: make(map[string]interface{}),
	}

	// Add parent if present
	if qdisc.Parent() != nil {
		config.Parent = qdisc.Parent()
	}

	// Convert qdisc-specific parameters
	switch qdisc.Type() {
	case entities.QdiscTypeHTB:
		// Check if this is an HTB qdisc and extract parameters
		if defaultClass, ok := qdisc.GetParameter("defaultClass"); ok {
			config.Parameters["defaultClass"] = defaultClass
		}
		if r2q, ok := qdisc.GetParameter("r2q"); ok {
			config.Parameters["r2q"] = r2q
		}
	case entities.QdiscTypeTBF:
		if rate, ok := qdisc.GetParameter("rate"); ok {
			config.Parameters["rate"] = rate
		}
		if buffer, ok := qdisc.GetParameter("buffer"); ok {
			config.Parameters["buffer"] = buffer
		}
		if limit, ok := qdisc.GetParameter("limit"); ok {
			config.Parameters["limit"] = limit
		}
	}

	result := a.adapter.AddQdisc(qdisc.Device(), config)
	if result.IsFailure() {
		return result.Error()
	}

	return nil
}

// AddClass adds a class from domain entity
func (a *AdapterWrapper) AddClass(ctx context.Context, class *entities.Class) error {
	// Extract device from class ID string - format is "device:handle"
	idStr := class.ID().String()
	colonIndex := -1
	for i, ch := range idStr {
		if ch == ':' {
			colonIndex = i
			break
		}
	}
	
	if colonIndex < 0 {
		return fmt.Errorf("invalid class ID format: %s", idStr)
	}
	
	deviceStr := idStr[:colonIndex]
	device, err := valueobjects.NewDevice(deviceStr)
	if err != nil {
		return fmt.Errorf("invalid device in class ID: %w", err)
	}

	config := ClassConfig{
		Handle:     class.Handle(),
		Parent:     class.Parent(),
		Type:       entities.QdiscTypeHTB, // Assume HTB for now
		Parameters: make(map[string]interface{}),
	}

	// For now, we'll use default parameters since the aggregate
	// stores only the base Class entity, not the HTBClass
	// TODO: Modify the aggregate to store HTBClass entities or 
	//       add rate/ceil information to the base Class entity
	// config.Parameters["rate"] = "1mbit"    // Default rate
	// config.Parameters["ceil"] = "10mbit"   // Default ceil

	result := a.adapter.AddClass(device, config)
	if result.IsFailure() {
		return result.Error()
	}

	return nil
}

// AddFilter adds a filter from domain entity
func (a *AdapterWrapper) AddFilter(ctx context.Context, filter *entities.Filter) error {
	// Get filter components from ID
	id := filter.ID()
	
	// Extract device from filter ID string - format is "device:parent:prioN:handle"
	idStr := id.String()
	colonIndex := 0
	for i, ch := range idStr {
		if ch == ':' {
			colonIndex = i
			break
		}
	}
	deviceStr := idStr[:colonIndex]
	device, err := valueobjects.NewDevice(deviceStr)
	if err != nil {
		return fmt.Errorf("invalid device in filter ID: %w", err)
	}

	// Extract parent handle from filter ID string
	firstColon := colonIndex
	secondColon := firstColon + 1
	for i := secondColon; i < len(idStr); i++ {
		if idStr[i] == ':' {
			secondColon = i
			break
		}
	}
	parentStr := idStr[firstColon+1:secondColon]
	parentMajor := uint16(0)
	parentMinor := uint16(0)
	if n, err := fmt.Sscanf(parentStr, "%x:%x", &parentMajor, &parentMinor); err != nil || n != 2 {
		return fmt.Errorf("invalid parent handle format: %w", err)
	}
	parentHandle := valueobjects.NewHandle(parentMajor, parentMinor)

	config := FilterConfig{
		Parent:   parentHandle,
		Priority: id.Priority(),
		Handle:   id.Handle(),
		Protocol: filter.Protocol(),
		FlowID:   filter.FlowID(),
		Matches:  []FilterMatch{},
	}

	// Convert match configurations - filters with simple map match
	// This is a temporary solution - should be refactored to use Filter.Matches()
	for matchType, value := range map[string]interface{}{} {
		var mt entities.MatchType
		switch matchType {
		case "src_ip":
			mt = entities.MatchTypeIPSource
		case "dst_ip":
			mt = entities.MatchTypeIPDestination
		case "src_port":
			mt = entities.MatchTypePortSource
			// Convert string port to uint16
			if portStr, ok := value.(string); ok {
				var port uint16
				if n, err := fmt.Sscanf(portStr, "%d", &port); err != nil || n != 1 {
					port = 0 // Default to 0 on error
				}
				value = port
			}
		case "dst_port":
			mt = entities.MatchTypePortDestination
			// Convert string port to uint16
			if portStr, ok := value.(string); ok {
				var port uint16
					if n, err := fmt.Sscanf(portStr, "%d", &port); err != nil || n != 1 {
						port = 0 // Default to 0 on error
					}
				value = port
			}
		case "protocol":
			mt = entities.MatchTypeProtocol
			// Convert protocol string to uint8
			if protoStr, ok := value.(string); ok {
				var proto uint8
				switch protoStr {
				case "tcp":
					proto = 6
				case "udp":
					proto = 17
				case "icmp":
					proto = 1
				default:
					_, _ = fmt.Sscanf(protoStr, "%d", &proto)
				}
				value = proto
			}
		default:
			continue // Skip unknown match types
		}

		config.Matches = append(config.Matches, FilterMatch{
			Type:  mt,
			Value: value,
		})
	}

	result := a.adapter.AddFilter(device, config)
	if result.IsFailure() {
		return result.Error()
	}

	return nil
}

// DeleteQdisc deletes a qdisc
func (a *AdapterWrapper) DeleteQdisc(device valueobjects.DeviceName, handle valueobjects.Handle) types.Result[Unit] {
	return a.adapter.DeleteQdisc(device, handle)
}

// GetQdiscs returns all qdiscs for a device
func (a *AdapterWrapper) GetQdiscs(device valueobjects.DeviceName) types.Result[[]QdiscInfo] {
	return a.adapter.GetQdiscs(device)
}

// DeleteClass deletes a class
func (a *AdapterWrapper) DeleteClass(device valueobjects.DeviceName, handle valueobjects.Handle) types.Result[Unit] {
	return a.adapter.DeleteClass(device, handle)
}

// GetClasses returns all classes for a device
func (a *AdapterWrapper) GetClasses(device valueobjects.DeviceName) types.Result[[]ClassInfo] {
	return a.adapter.GetClasses(device)
}

// DeleteFilter deletes a filter
func (a *AdapterWrapper) DeleteFilter(device valueobjects.DeviceName, parent valueobjects.Handle, priority uint16, handle valueobjects.Handle) types.Result[Unit] {
	return a.adapter.DeleteFilter(device, parent, priority, handle)
}

// GetFilters returns all filters for a device
func (a *AdapterWrapper) GetFilters(device valueobjects.DeviceName) types.Result[[]FilterInfo] {
	return a.adapter.GetFilters(device)
}