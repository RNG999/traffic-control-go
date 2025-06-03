package application

import (
	"context"
	"fmt"

	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/events"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/pkg/logging"
)

// handleQdiscCreated handles QdiscCreated events and applies them to netlink
func (s *TrafficControlService) handleQdiscCreated(ctx context.Context, event interface{}) error {
	// Type assert to the event types we expect
	var device valueobjects.DeviceName
	var handle valueobjects.Handle
	var qdiscType entities.QdiscType
	var defaultClass string

	switch e := event.(type) {
	case *events.QdiscCreatedEvent:
		device = e.DeviceName
		handle = e.Handle
		qdiscType = e.QdiscType
	case *events.HTBQdiscCreatedEvent:
		device = e.DeviceName
		handle = e.Handle
		qdiscType = entities.QdiscTypeHTB
		defaultClass = e.DefaultClass.String()
	default:
		// Not a qdisc event we handle
		return nil
	}

	s.logger.Info("Applying qdisc to netlink",
		logging.String("device", device.String()),
		logging.String("handle", handle.String()),
		logging.Int("type", int(qdiscType)),
	)

	// Apply to netlink based on type
	switch qdiscType {
	case entities.QdiscTypeHTB:
		// Parse default class handle
		defaultHandle, err := valueobjects.ParseHandle(defaultClass)
		if err != nil {
			return fmt.Errorf("invalid default class handle: %w", err)
		}
		qdisc := entities.NewHTBQdisc(device, handle, defaultHandle)
		return s.netlinkAdapter.AddQdisc(ctx, qdisc.Qdisc)
	case entities.QdiscTypeTBF:
		// TBF needs rate from event - skip for now
		s.logger.Warn("TBF qdisc netlink application not implemented")
		return nil
	default:
		s.logger.Warn("Unsupported qdisc type for netlink",
			logging.Int("type", int(qdiscType)),
		)
		return nil
	}
}

// handleClassCreated handles ClassCreated events and applies them to netlink
func (s *TrafficControlService) handleClassCreated(ctx context.Context, event interface{}) error {
	switch e := event.(type) {
	case *events.ClassCreatedEvent:
		// Basic class creation - not HTB specific
		s.logger.Info("Basic class created event - netlink application not implemented",
			logging.String("device", e.DeviceName.String()),
			logging.String("handle", e.Handle.String()),
		)
		return nil

	case *events.HTBClassCreatedEvent:
		s.logger.Info("Applying HTB class to netlink",
			logging.String("device", e.DeviceName.String()),
			logging.String("parent", e.Parent.String()),
			logging.String("handle", e.Handle.String()),
			logging.String("rate", e.Rate.String()),
			logging.String("ceil", e.Ceil.String()),
		)

		// Create HTB class with proper parameters
		// Note: HTB events don't have priority in the event, using default priority
		priority := entities.Priority(4) // Normal priority
		class := entities.NewHTBClass(e.DeviceName, e.Handle, e.Parent, e.Name, priority)

		// Set rate and ceil bandwidth
		class.SetRate(e.Rate)

		// If ceil is 0 or not set, default to rate (HTB requirement)
		if e.Ceil.BitsPerSecond() == 0 {
			class.SetCeil(e.Rate)
		} else {
			class.SetCeil(e.Ceil)
		}

		// Set burst values if provided, otherwise calculate them
		if e.Burst > 0 {
			class.SetBurst(e.Burst)
		} else {
			class.SetBurst(class.CalculateBurst())
		}

		if e.Cburst > 0 {
			class.SetCburst(e.Cburst)
		} else {
			class.SetCburst(class.CalculateCburst())
		}

		return s.netlinkAdapter.AddClass(ctx, class)

	default:
		// Not a class event we handle
		return nil
	}
}

// handleFilterCreated handles FilterCreated events and applies them to netlink
func (s *TrafficControlService) handleFilterCreated(ctx context.Context, event interface{}) error {
	e, ok := event.(*events.FilterCreatedEvent)
	if !ok {
		// Not a filter event
		return nil
	}

	s.logger.Info("Applying filter to netlink",
		logging.String("device", e.DeviceName.String()),
		logging.String("parent", e.Parent.String()),
		logging.String("handle", e.Handle.String()),
		logging.String("flow_id", e.FlowID.String()),
		logging.Int("priority", int(e.Priority)),
	)

	// Create filter entity directly from event data (no string parsing needed)
	filter := entities.NewFilter(e.DeviceName, e.Parent, e.Priority, e.Handle)

	// Set flow ID
	filter.SetFlowID(e.FlowID)

	// Set protocol
	filter.SetProtocol(e.Protocol)

	// Add matches from event data
	for _, matchData := range e.Matches {
		// TODO: Convert match data to proper match objects
		// For now, just log them
		s.logger.Debug("Filter match",
			logging.Int("type", int(matchData.Type)),
			logging.String("value", matchData.Value),
		)
	}

	s.logger.Info("Adding filter via netlink adapter")
	return s.netlinkAdapter.AddFilter(ctx, filter)
}
