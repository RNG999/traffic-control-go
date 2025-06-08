package application

import (
	"context"
	"fmt"
	"time"

	chandlers "github.com/rng999/traffic-control-go/internal/commands/handlers"
	"github.com/rng999/traffic-control-go/internal/commands/models"
	"github.com/rng999/traffic-control-go/internal/domain/events"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	"github.com/rng999/traffic-control-go/internal/projections"
	qhandlers "github.com/rng999/traffic-control-go/internal/queries/handlers"
	qmodels "github.com/rng999/traffic-control-go/internal/queries/models"
	"github.com/rng999/traffic-control-go/pkg/logging"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// TrafficControlService is the main application service that coordinates
// between the API layer, CQRS handlers, and infrastructure
type TrafficControlService struct {
	eventStore        eventstore.EventStoreWithContext
	netlinkAdapter    netlink.Adapter
	commandBus        *CommandBus
	queryBus          *QueryBus
	eventBus          *EventBus
	projectionManager *projections.Manager
	readModelStore    projections.ReadModelStore
	statisticsService *StatisticsService
	logger            logging.Logger
}

// NewTrafficControlService creates a new traffic control service
func NewTrafficControlService(
	eventStore eventstore.EventStoreWithContext,
	netlinkAdapter netlink.Adapter,
	logger logging.Logger,
) *TrafficControlService {
	// Create read model store
	readModelStore := projections.NewMemoryReadModelStore()

	// Create projection manager
	var baseEventStore eventstore.EventStore
	if wrapper, ok := eventStore.(*eventstore.MemoryEventStoreWrapper); ok {
		baseEventStore = wrapper.MemoryEventStore
	} else {
		// Fallback - this shouldn't happen in normal usage
		baseEventStore = eventStore
	}
	projectionManager := projections.NewManager(baseEventStore)

	service := &TrafficControlService{
		eventStore:        eventStore,
		netlinkAdapter:    netlinkAdapter,
		projectionManager: projectionManager,
		readModelStore:    readModelStore,
		logger:            logger,
	}

	// Initialize statistics service
	service.statisticsService = NewStatisticsService(netlinkAdapter, readModelStore)

	// Initialize buses
	service.commandBus = NewCommandBus(service)
	service.queryBus = NewQueryBus(service)
	service.eventBus = NewEventBus(service)

	// Setup event publishing from event store to event bus
	if wrapper, ok := eventStore.(*eventstore.MemoryEventStoreWrapper); ok {
		wrapper.SetEventPublisher(service.publishEvent)
	}

	// Register handlers
	service.registerHandlers()

	// Register projections
	service.registerProjections()

	return service
}

// registerHandlers registers all command and query handlers
func (s *TrafficControlService) registerHandlers() {
	// Legacy command handlers removed - now using type-safe generic handlers only

	// Register type-safe command handlers
	RegisterHandlerFor[*models.CreateHTBQdiscCommand](s.commandBus, chandlers.NewCreateHTBQdiscHandler(s.eventStore))
	RegisterHandlerFor[*models.CreateHTBClassCommand](s.commandBus, chandlers.NewCreateHTBClassHandler(s.eventStore))
	RegisterHandlerFor[*models.CreateFilterCommand](s.commandBus, chandlers.NewCreateFilterHandler(s.eventStore))
	RegisterHandlerFor[*models.CreateTBFQdiscCommand](s.commandBus, chandlers.NewCreateTBFQdiscHandler(s.eventStore))
	RegisterHandlerFor[*models.CreatePRIOQdiscCommand](s.commandBus, chandlers.NewCreatePRIOQdiscHandler(s.eventStore))
	RegisterHandlerFor[*models.CreateFQCODELQdiscCommand](s.commandBus, chandlers.NewCreateFQCODELQdiscHandler(s.eventStore))

	// Register query handlers with event store access for aggregate reconstruction
	if baseEventStore, ok := s.eventStore.(eventstore.EventStore); ok {
		s.queryBus.Register("GetQdisc", qhandlers.NewGetQdiscByDeviceHandler(baseEventStore))
		s.queryBus.Register("GetClass", qhandlers.NewGetClassesByDeviceHandler(baseEventStore))
		s.queryBus.Register("GetFilter", qhandlers.NewGetFiltersByDeviceHandler(baseEventStore))
		s.queryBus.Register("GetConfiguration", qhandlers.NewGetTrafficControlConfigHandler(baseEventStore))
	}

	// Create statistics query service
	statisticsQueryService := qhandlers.NewStatisticsQueryService(s.netlinkAdapter, s.readModelStore)

	// Register statistics query handlers
	s.queryBus.Register("GetDeviceStatistics", qhandlers.NewGetDeviceStatisticsHandler(statisticsQueryService))
	s.queryBus.Register("GetQdiscStatistics", qhandlers.NewGetQdiscStatisticsHandler(s.netlinkAdapter))
	s.queryBus.Register("GetClassStatistics", qhandlers.NewGetClassStatisticsHandler(s.netlinkAdapter))
	s.queryBus.Register("GetRealtimeStatistics", qhandlers.NewGetRealtimeStatisticsHandler(statisticsQueryService))

	// Register event handlers for netlink integration
	s.eventBus.Subscribe("QdiscCreated", s.handleQdiscCreated)
	s.eventBus.Subscribe("HTBQdiscCreated", s.handleQdiscCreated)
	s.eventBus.Subscribe("ClassCreated", s.handleClassCreated)
	s.eventBus.Subscribe("HTBClassCreated", s.handleClassCreated)
	s.eventBus.Subscribe("FilterCreated", s.handleFilterCreated)

	// Register event handlers for projections
	s.eventBus.Subscribe("QdiscCreated", s.handleEventForProjections)
	s.eventBus.Subscribe("ClassCreated", s.handleEventForProjections)
	s.eventBus.Subscribe("FilterCreated", s.handleEventForProjections)
}

// registerProjections registers all projections
func (s *TrafficControlService) registerProjections() {
	// Register traffic control projection
	tcProjection := projections.NewTrafficControlProjection(s.readModelStore)
	s.projectionManager.Register(tcProjection)
}

// CreateHTBQdisc creates a new HTB qdisc
func (s *TrafficControlService) CreateHTBQdisc(ctx context.Context, device string, handle string, defaultClass string) error {
	cmd := &models.CreateHTBQdiscCommand{
		DeviceName:   device,
		Handle:       handle,
		DefaultClass: defaultClass,
	}

	if err := s.commandBus.ExecuteCommand(ctx, cmd); err != nil {
		return fmt.Errorf("failed to create HTB qdisc: %w", err)
	}

	return nil
}

// CreateTBFQdisc creates a new TBF qdisc
func (s *TrafficControlService) CreateTBFQdisc(ctx context.Context, device string, handle string, rate string, buffer, limit, burst uint32) error {
	cmd := &models.CreateTBFQdiscCommand{
		DeviceName: device,
		Handle:     handle,
		Rate:       rate,
		Buffer:     buffer,
		Limit:      limit,
		Burst:      burst,
	}

	if err := s.commandBus.ExecuteCommand(ctx, cmd); err != nil {
		return fmt.Errorf("failed to create TBF qdisc: %w", err)
	}

	return nil
}

// CreatePRIOQdisc creates a new PRIO qdisc
func (s *TrafficControlService) CreatePRIOQdisc(ctx context.Context, device string, handle string, bands uint8, priomap []uint8) error {
	cmd := &models.CreatePRIOQdiscCommand{
		DeviceName: device,
		Handle:     handle,
		Bands:      bands,
		Priomap:    priomap,
	}

	if err := s.commandBus.ExecuteCommand(ctx, cmd); err != nil {
		return fmt.Errorf("failed to create PRIO qdisc: %w", err)
	}

	return nil
}

// CreateFQCODELQdisc creates a new FQ_CODEL qdisc
func (s *TrafficControlService) CreateFQCODELQdisc(ctx context.Context, device string, handle string, limit, flows, target, interval, quantum uint32, ecn bool) error {
	cmd := &models.CreateFQCODELQdiscCommand{
		DeviceName: device,
		Handle:     handle,
		Limit:      limit,
		Flows:      flows,
		Target:     target,
		Interval:   interval,
		Quantum:    quantum,
		ECN:        ecn,
	}

	if err := s.commandBus.ExecuteCommand(ctx, cmd); err != nil {
		return fmt.Errorf("failed to create FQ_CODEL qdisc: %w", err)
	}

	return nil
}

// CreateHTBClass creates a new HTB class
func (s *TrafficControlService) CreateHTBClass(ctx context.Context, device string, parent string, classID string, rate string, ceil string) error {
	cmd := &models.CreateHTBClassCommand{
		DeviceName: device,
		Parent:     parent,
		ClassID:    classID,
		Rate:       rate,
		Ceil:       ceil,
	}

	if err := s.commandBus.ExecuteCommand(ctx, cmd); err != nil {
		return fmt.Errorf("failed to create HTB class: %w", err)
	}

	return nil
}

// CreateHTBClassWithAdvancedParameters creates a new HTB class with advanced parameters including priority
func (s *TrafficControlService) CreateHTBClassWithAdvancedParameters(ctx context.Context, device string, parent string, classID string, name string, rate string, ceil string, priority uint8) error {
	cmd := &models.CreateHTBClassCommand{
		DeviceName:  device,
		Parent:      parent,
		ClassID:     classID,
		Name:        name,
		Rate:        rate,
		Ceil:        ceil,
		Priority:    int(priority),
		UseDefaults: true, // Use sensible defaults for advanced parameters
	}

	if err := s.commandBus.ExecuteCommand(ctx, cmd); err != nil {
		return fmt.Errorf("failed to create HTB class with advanced parameters: %w", err)
	}

	return nil
}

// CreateFilter creates a new filter
func (s *TrafficControlService) CreateFilter(ctx context.Context, device string, parent string, priority uint16, protocol string, flowID string, match map[string]string) error {
	cmd := &models.CreateFilterCommand{
		DeviceName: device,
		Parent:     parent,
		Priority:   priority,
		Protocol:   protocol,
		FlowID:     flowID,
		Match:      match,
	}

	if err := s.commandBus.ExecuteCommand(ctx, cmd); err != nil {
		return fmt.Errorf("failed to create filter: %w", err)
	}

	return nil
}

// GetConfiguration retrieves the current traffic control configuration
func (s *TrafficControlService) GetConfiguration(ctx context.Context, device string) (*qmodels.ConfigurationView, error) {
	deviceName, err := tc.NewDevice(device)
	if err != nil {
		return nil, fmt.Errorf("invalid device name: %w", err)
	}

	query := qmodels.NewGetTrafficControlConfigQuery(deviceName)

	result, err := s.queryBus.Execute(ctx, "GetConfiguration", query)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	config, ok := result.(qmodels.TrafficControlConfigView)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	// Convert to expected return type
	view := &qmodels.ConfigurationView{
		DeviceName: config.DeviceName,
		Qdiscs:     config.Qdiscs,
		Classes:    config.Classes,
		Filters:    config.Filters,
	}

	return view, nil
}

// GetDeviceStatistics retrieves comprehensive statistics for a device
func (s *TrafficControlService) GetDeviceStatistics(ctx context.Context, device string) (*qmodels.DeviceStatisticsView, error) {
	deviceName, err := tc.NewDevice(device)
	if err != nil {
		return nil, fmt.Errorf("invalid device name: %w", err)
	}

	query := qmodels.NewGetDeviceStatisticsQuery(deviceName)

	result, err := s.queryBus.Execute(ctx, "GetDeviceStatistics", query)
	if err != nil {
		return nil, fmt.Errorf("failed to get device statistics: %w", err)
	}

	stats, ok := result.(qmodels.DeviceStatisticsView)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	return &stats, nil
}

// GetQdiscStatistics retrieves statistics for a specific qdisc
func (s *TrafficControlService) GetQdiscStatistics(ctx context.Context, device string, handle string) (*qmodels.QdiscStatisticsView, error) {
	deviceName, err := tc.NewDevice(device)
	if err != nil {
		return nil, fmt.Errorf("invalid device name: %w", err)
	}

	qdiscHandle, err := tc.ParseHandle(handle)
	if err != nil {
		return nil, fmt.Errorf("invalid handle: %w", err)
	}

	query := qmodels.NewGetQdiscStatisticsQuery(deviceName, qdiscHandle)

	result, err := s.queryBus.Execute(ctx, "GetQdiscStatistics", query)
	if err != nil {
		return nil, fmt.Errorf("failed to get qdisc statistics: %w", err)
	}

	stats, ok := result.(qmodels.QdiscStatisticsView)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	return &stats, nil
}

// GetClassStatistics retrieves statistics for a specific class
func (s *TrafficControlService) GetClassStatistics(ctx context.Context, device string, handle string) (*qmodels.ClassStatisticsView, error) {
	deviceName, err := tc.NewDevice(device)
	if err != nil {
		return nil, fmt.Errorf("invalid device name: %w", err)
	}

	classHandle, err := tc.ParseHandle(handle)
	if err != nil {
		return nil, fmt.Errorf("invalid handle: %w", err)
	}

	query := qmodels.NewGetClassStatisticsQuery(deviceName, classHandle)

	result, err := s.queryBus.Execute(ctx, "GetClassStatistics", query)
	if err != nil {
		return nil, fmt.Errorf("failed to get class statistics: %w", err)
	}

	stats, ok := result.(qmodels.ClassStatisticsView)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	return &stats, nil
}

// GetRealtimeStatistics retrieves real-time statistics without using read models
func (s *TrafficControlService) GetRealtimeStatistics(ctx context.Context, device string) (*qmodels.DeviceStatisticsView, error) {
	deviceName, err := tc.NewDevice(device)
	if err != nil {
		return nil, fmt.Errorf("invalid device name: %w", err)
	}

	query := qmodels.NewGetRealtimeStatisticsQuery(deviceName)

	result, err := s.queryBus.Execute(ctx, "GetRealtimeStatistics", query)
	if err != nil {
		return nil, fmt.Errorf("failed to get realtime statistics: %w", err)
	}

	stats, ok := result.(qmodels.DeviceStatisticsView)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	return &stats, nil
}

// MonitorStatistics starts continuous monitoring of statistics
func (s *TrafficControlService) MonitorStatistics(ctx context.Context, device string, interval time.Duration, callback func(*qmodels.DeviceStatisticsView)) error {
	return s.statisticsService.MonitorStatistics(ctx, device, interval, func(stats *DeviceStatistics) {
		// Convert to view and call callback
		view := convertApplicationStatsToView(stats)
		callback(&view)
	})
}

// 削除: tc.ParseHandle()を直接使用するため不要

// convertApplicationStatsToView converts application model to view model
func convertApplicationStatsToView(stats *DeviceStatistics) qmodels.DeviceStatisticsView {
	view := qmodels.DeviceStatisticsView{
		DeviceName:  stats.DeviceName,
		Timestamp:   stats.Timestamp.Format(time.RFC3339),
		QdiscStats:  make([]qmodels.QdiscStatisticsView, 0, len(stats.QdiscStats)),
		ClassStats:  make([]qmodels.ClassStatisticsView, 0, len(stats.ClassStats)),
		FilterStats: make([]qmodels.FilterStatisticsView, 0, len(stats.FilterStats)),
		LinkStats: qmodels.LinkStatisticsView{
			RxBytes:   stats.LinkStats.RxBytes,
			TxBytes:   stats.LinkStats.TxBytes,
			RxPackets: stats.LinkStats.RxPackets,
			TxPackets: stats.LinkStats.TxPackets,
			RxErrors:  stats.LinkStats.RxErrors,
			TxErrors:  stats.LinkStats.TxErrors,
			RxDropped: stats.LinkStats.RxDropped,
			TxDropped: stats.LinkStats.TxDropped,
		},
	}

	// Convert qdisc statistics
	for _, qdisc := range stats.QdiscStats {
		qdiscView := qmodels.QdiscStatisticsView{
			Handle:        qdisc.Handle,
			Type:          qdisc.Type,
			BytesSent:     qdisc.Stats.BytesSent,
			PacketsSent:   qdisc.Stats.PacketsSent,
			BytesDropped:  qdisc.Stats.BytesDropped,
			Overlimits:    qdisc.Stats.Overlimits,
			Requeues:      qdisc.Stats.Requeues,
			DetailedStats: make(map[string]interface{}),
		}

		if qdisc.DetailedStats != nil {
			qdiscView.Backlog = qdisc.DetailedStats.Backlog
			qdiscView.QueueLength = qdisc.DetailedStats.QueueLength
			qdiscView.DetailedStats["backlog_bytes"] = qdisc.DetailedStats.BacklogBytes
			qdiscView.DetailedStats["bytes_per_second"] = qdisc.DetailedStats.BytesPerSecond
			qdiscView.DetailedStats["packets_per_second"] = qdisc.DetailedStats.PacketsPerSecond
			if qdisc.DetailedStats.HTBStats != nil {
				qdiscView.DetailedStats["htb_direct_packets"] = qdisc.DetailedStats.HTBStats.DirectPackets
				qdiscView.DetailedStats["htb_version"] = qdisc.DetailedStats.HTBStats.Version
			}
		}

		view.QdiscStats = append(view.QdiscStats, qdiscView)
	}

	// Convert class statistics
	for _, class := range stats.ClassStats {
		classView := qmodels.ClassStatisticsView{
			Handle:         class.Handle,
			Parent:         class.Parent,
			Name:           class.Name,
			BytesSent:      class.Stats.BytesSent,
			PacketsSent:    class.Stats.PacketsSent,
			BytesDropped:   class.Stats.BytesDropped,
			Overlimits:     class.Stats.Overlimits,
			BacklogBytes:   class.Stats.BacklogBytes,
			BacklogPackets: class.Stats.BacklogPackets,
			RateBPS:        class.Stats.RateBPS,
			DetailedStats:  make(map[string]interface{}),
		}

		if class.DetailedStats != nil && class.DetailedStats.HTBStats != nil {
			classView.DetailedStats["htb_lends"] = class.DetailedStats.HTBStats.Lends
			classView.DetailedStats["htb_borrows"] = class.DetailedStats.HTBStats.Borrows
			classView.DetailedStats["htb_giants"] = class.DetailedStats.HTBStats.Giants
			classView.DetailedStats["htb_tokens"] = class.DetailedStats.HTBStats.Tokens
			classView.DetailedStats["htb_ctokens"] = class.DetailedStats.HTBStats.CTokens
			classView.DetailedStats["htb_rate"] = class.DetailedStats.HTBStats.Rate
			classView.DetailedStats["htb_ceil"] = class.DetailedStats.HTBStats.Ceil
			classView.DetailedStats["htb_level"] = class.DetailedStats.HTBStats.Level
		}

		view.ClassStats = append(view.ClassStats, classView)
	}

	// Convert filter statistics
	for _, filter := range stats.FilterStats {
		filterView := qmodels.FilterStatisticsView{
			Parent:     filter.Parent,
			Priority:   filter.Priority,
			Protocol:   filter.Protocol,
			MatchCount: filter.MatchCount,
		}
		view.FilterStats = append(view.FilterStats, filterView)
	}

	return view
}

// publishEvent publishes an event to the event bus
func (s *TrafficControlService) publishEvent(ctx context.Context, event interface{}) error {
	// Determine event type from the event itself
	eventType := ""
	switch event.(type) {
	case *events.QdiscCreatedEvent:
		eventType = "QdiscCreated"
	case *events.HTBQdiscCreatedEvent:
		eventType = "HTBQdiscCreated"
	case *events.ClassCreatedEvent:
		eventType = "ClassCreated"
	case *events.HTBClassCreatedEvent:
		eventType = "HTBClassCreated"
	case *events.FilterCreatedEvent:
		eventType = "FilterCreated"
	default:
		s.logger.Debug("Unknown event type, skipping publish", logging.String("type", fmt.Sprintf("%T", event)))
		return nil
	}

	return s.eventBus.Publish(ctx, eventType, event)
}

// handleEventForProjections forwards events to the projection manager
func (s *TrafficControlService) handleEventForProjections(ctx context.Context, event interface{}) error {
	// Get the latest event from the event store to get the full event with metadata
	_, err := s.eventStore.GetEventsWithContext(ctx, "", 0, 1)
	if err != nil {
		return err
	}

	// TODO: Fix event type processing for projections
	// if len(events) > 0 {
	//     if domainEvent, ok := events[0].(events.DomainEvent); ok {
	//         return s.projectionManager.ProcessEvent(ctx, domainEvent)
	//     }
	// }

	return nil
}
