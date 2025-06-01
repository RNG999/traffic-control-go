package api

import (
	"context"
	"fmt"
	"time"

	"github.com/rng999/traffic-control-go/internal/application"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	qmodels "github.com/rng999/traffic-control-go/internal/queries/models"
	"github.com/rng999/traffic-control-go/pkg/logging"
)

// TrafficController is the main entry point for traffic control configuration
type TrafficController struct {
	deviceName     string
	totalBandwidth valueobjects.Bandwidth
	classes        []*TrafficClass
	logger         logging.Logger
	service        *application.TrafficControlService
}

// TrafficClass represents a traffic classification with its rules
type TrafficClass struct {
	name                string
	guaranteedBandwidth valueobjects.Bandwidth
	maxBandwidth        valueobjects.Bandwidth
	priority            *Priority // Priority is now required and must be explicitly set
	filters             []Filter
}

// Priority represents the priority level of traffic (0-7, where 0 is highest priority)
type Priority int

// Filter represents a packet matching rule
type Filter struct {
	filterType FilterType
	value      interface{}
}

// FilterType represents the type of filter
type FilterType int

const (
	SourceIPFilter FilterType = iota
	DestinationIPFilter
	SourcePortFilter
	DestinationPortFilter
	ProtocolFilter
	ApplicationFilter
)

// New creates a new traffic controller for a network interface
func New(deviceName string) *TrafficController {
	logger := logging.WithComponent(logging.ComponentAPI).WithDevice(deviceName)
	logger.Info("Creating new traffic controller",
		logging.String("device", deviceName),
	)

	// Initialize the application service with default dependencies
	// In production, these would be injected
	eventStore := eventstore.NewMemoryEventStoreWithContext()
	netlinkAdapter := netlink.NewAdapter()
	service := application.NewTrafficControlService(eventStore, netlinkAdapter, logger)

	return &TrafficController{
		deviceName: deviceName,
		classes:    make([]*TrafficClass, 0),
		logger:     logger,
		service:    service,
	}
}

// SetTotalBandwidth sets the total available bandwidth for the interface
func (tc *TrafficController) SetTotalBandwidth(bandwidth string) *TrafficController {
	tc.logger.Info("Setting total bandwidth",
		logging.String("bandwidth", bandwidth),
		logging.String("operation", logging.OperationConfigLoad),
	)

	tc.totalBandwidth = valueobjects.MustParseBandwidth(bandwidth)
	return tc
}

// CreateTrafficClass creates a new traffic class with a human-readable name
func (tc *TrafficController) CreateTrafficClass(name string) *TrafficClassBuilder {
	tc.logger.Info("Creating traffic class",
		logging.String("class_name", name),
		logging.String("operation", logging.OperationCreateClass),
	)

	class := &TrafficClass{
		name: name,
		// priority is nil by default - must be set explicitly
	}

	return &TrafficClassBuilder{
		controller: tc,
		class:      class,
	}
}

// TrafficClassBuilder provides a fluent interface for building traffic classes
type TrafficClassBuilder struct {
	controller *TrafficController
	class      *TrafficClass
}

// WithGuaranteedBandwidth sets the minimum guaranteed bandwidth
func (b *TrafficClassBuilder) WithGuaranteedBandwidth(bandwidth string) *TrafficClassBuilder {
	b.class.guaranteedBandwidth = valueobjects.MustParseBandwidth(bandwidth)
	return b
}

// WithMaxBandwidth sets the maximum allowed bandwidth
func (b *TrafficClassBuilder) WithMaxBandwidth(bandwidth string) *TrafficClassBuilder {
	b.class.maxBandwidth = valueobjects.MustParseBandwidth(bandwidth)
	return b
}

// WithBurstableTo is an alias for WithMaxBandwidth for better readability
func (b *TrafficClassBuilder) WithBurstableTo(bandwidth string) *TrafficClassBuilder {
	return b.WithMaxBandwidth(bandwidth)
}

// WithPriority sets the traffic class to a specific priority level (0-7)
func (b *TrafficClassBuilder) WithPriority(priority int) *TrafficClassBuilder {
	// HTB supports priority values 0-7, where lower numbers = higher priority
	if priority < 0 {
		priority = 0
	} else if priority > 7 {
		priority = 7
	}
	p := Priority(priority)
	b.class.priority = &p
	return b
}

// ForDestination adds a destination IP filter
func (b *TrafficClassBuilder) ForDestination(ip string) *TrafficClassBuilder {
	b.class.filters = append(b.class.filters, Filter{
		filterType: DestinationIPFilter,
		value:      ip,
	})
	return b
}

// ForSource adds a source IP filter
func (b *TrafficClassBuilder) ForSource(ip string) *TrafficClassBuilder {
	b.class.filters = append(b.class.filters, Filter{
		filterType: SourceIPFilter,
		value:      ip,
	})
	return b
}

// ForPort adds a destination port filter
func (b *TrafficClassBuilder) ForPort(ports ...int) *TrafficClassBuilder {
	for _, port := range ports {
		b.class.filters = append(b.class.filters, Filter{
			filterType: DestinationPortFilter,
			value:      port,
		})
	}
	return b
}

// ForApplication adds an application-based filter (predefined port sets)
func (b *TrafficClassBuilder) ForApplication(apps ...string) *TrafficClassBuilder {
	for _, app := range apps {
		b.class.filters = append(b.class.filters, Filter{
			filterType: ApplicationFilter,
			value:      app,
		})
	}
	return b
}

// And returns the traffic controller for method chaining
func (b *TrafficClassBuilder) And() *TrafficController {
	b.controller.classes = append(b.controller.classes, b.class)
	return b.controller
}

// Apply completes the builder and adds the class to the controller
func (b *TrafficClassBuilder) Apply() error {
	b.controller.classes = append(b.controller.classes, b.class)
	return b.controller.Apply()
}

// PriorityGroupBuilder builds priority-based traffic groups
type PriorityGroupBuilder struct {
	controller *TrafficController //nolint:unused
	priority   Priority           //nolint:unused
	filters    []Filter
}

// ForSSH adds SSH traffic to this priority group
func (p *PriorityGroupBuilder) ForSSH() *PriorityGroupBuilder {
	p.filters = append(p.filters, Filter{
		filterType: DestinationPortFilter,
		value:      22,
	})
	return p
}

// ForHTTP adds HTTP traffic to this priority group
func (p *PriorityGroupBuilder) ForHTTP() *PriorityGroupBuilder {
	p.filters = append(p.filters, Filter{
		filterType: DestinationPortFilter,
		value:      80,
	})
	return p
}

// ForHTTPS adds HTTPS traffic to this priority group
func (p *PriorityGroupBuilder) ForHTTPS() *PriorityGroupBuilder {
	p.filters = append(p.filters, Filter{
		filterType: DestinationPortFilter,
		value:      443,
	})
	return p
}

// HTBQdisc creates an HTB qdisc with fluent interface
func (tc *TrafficController) HTBQdisc(handle, defaultClass string) *HTBQdiscBuilder {
	return &HTBQdiscBuilder{
		controller:   tc,
		handle:       handle,
		defaultClass: defaultClass,
	}
}

// TBFQdisc creates a TBF qdisc with fluent interface  
func (tc *TrafficController) TBFQdisc(handle, rate string) *TBFQdiscBuilder {
	return &TBFQdiscBuilder{
		controller: tc,
		handle:     handle,
		rate:       rate,
		buffer:     32768, // default buffer
		limit:      10000, // default limit
		burst:      0,     // will be calculated if not set
	}
}

// PRIOQdisc creates a PRIO qdisc with fluent interface
func (tc *TrafficController) PRIOQdisc(handle string, bands uint8) *PRIOQdiscBuilder {
	return &PRIOQdiscBuilder{
		controller: tc,
		handle:     handle,
		bands:      bands,
		priomap:    []uint8{1, 2, 2, 2, 1, 2, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1}, // default priomap
	}
}

// FQCODELQdisc creates a FQ_CODEL qdisc with fluent interface
func (tc *TrafficController) FQCODELQdisc(handle string) *FQCODELQdiscBuilder {
	return &FQCODELQdiscBuilder{
		controller: tc,
		handle:     handle,
		limit:      10240,  // default limit
		flows:      1024,   // default flows
		target:     5000,   // 5ms target
		interval:   100000, // 100ms interval
		quantum:    1518,   // default quantum
		ecn:        false,  // ECN disabled by default
	}
}

// HTBQdiscBuilder provides fluent interface for HTB qdiscs
type HTBQdiscBuilder struct {
	controller   *TrafficController
	handle       string
	defaultClass string
	classes      []*HTBClassConfig
}

type HTBClassConfig struct {
	parent string
	handle string
	name   string
	rate   string
	ceil   string
}

func (b *HTBQdiscBuilder) HTBClass(parent, handle, name, rate, ceil string) *HTBQdiscBuilder {
	b.classes = append(b.classes, &HTBClassConfig{
		parent: parent,
		handle: handle,
		name:   name,
		rate:   rate,
		ceil:   ceil,
	})
	return b
}

func (b *HTBQdiscBuilder) Apply() error {
	ctx := context.Background()
	
	// Create HTB qdisc
	if err := b.controller.service.CreateHTBQdisc(ctx, b.controller.deviceName, b.handle, b.defaultClass); err != nil {
		return fmt.Errorf("failed to create HTB qdisc: %w", err)
	}
	
	// Create classes
	for _, class := range b.classes {
		if err := b.controller.service.CreateHTBClass(ctx, b.controller.deviceName, class.parent, class.handle, class.rate, class.ceil); err != nil {
			return fmt.Errorf("failed to create HTB class %s: %w", class.name, err)
		}
	}
	
	return nil
}

// TBFQdiscBuilder provides fluent interface for TBF qdiscs
type TBFQdiscBuilder struct {
	controller *TrafficController
	handle     string
	rate       string
	buffer     uint32
	limit      uint32
	burst      uint32
}

func (b *TBFQdiscBuilder) WithBuffer(buffer uint32) *TBFQdiscBuilder {
	b.buffer = buffer
	return b
}

func (b *TBFQdiscBuilder) WithLimit(limit uint32) *TBFQdiscBuilder {
	b.limit = limit
	return b
}

func (b *TBFQdiscBuilder) WithBurst(burst uint32) *TBFQdiscBuilder {
	b.burst = burst
	return b
}

func (b *TBFQdiscBuilder) Apply() error {
	ctx := context.Background()
	return b.controller.service.CreateTBFQdisc(ctx, b.controller.deviceName, b.handle, b.rate, b.buffer, b.limit, b.burst)
}

// PRIOQdiscBuilder provides fluent interface for PRIO qdiscs
type PRIOQdiscBuilder struct {
	controller *TrafficController
	handle     string
	bands      uint8
	priomap    []uint8
}

func (b *PRIOQdiscBuilder) WithPriomap(priomap []uint8) *PRIOQdiscBuilder {
	if len(priomap) == 16 {
		b.priomap = priomap
	}
	return b
}

func (b *PRIOQdiscBuilder) Apply() error {
	ctx := context.Background()
	return b.controller.service.CreatePRIOQdisc(ctx, b.controller.deviceName, b.handle, b.bands, b.priomap)
}

// FQCODELQdiscBuilder provides fluent interface for FQ_CODEL qdiscs
type FQCODELQdiscBuilder struct {
	controller *TrafficController
	handle     string
	limit      uint32
	flows      uint32
	target     uint32
	interval   uint32
	quantum    uint32
	ecn        bool
}

func (b *FQCODELQdiscBuilder) WithLimit(limit uint32) *FQCODELQdiscBuilder {
	b.limit = limit
	return b
}

func (b *FQCODELQdiscBuilder) WithFlows(flows uint32) *FQCODELQdiscBuilder {
	b.flows = flows
	return b
}

func (b *FQCODELQdiscBuilder) WithTarget(target uint32) *FQCODELQdiscBuilder {
	b.target = target
	return b
}

func (b *FQCODELQdiscBuilder) WithInterval(interval uint32) *FQCODELQdiscBuilder {
	b.interval = interval
	return b
}

func (b *FQCODELQdiscBuilder) WithQuantum(quantum uint32) *FQCODELQdiscBuilder {
	b.quantum = quantum
	return b
}

func (b *FQCODELQdiscBuilder) WithECN(ecn bool) *FQCODELQdiscBuilder {
	b.ecn = ecn
	return b
}

func (b *FQCODELQdiscBuilder) Apply() error {
	ctx := context.Background()
	return b.controller.service.CreateFQCODELQdisc(ctx, b.controller.deviceName, b.handle, b.limit, b.flows, b.target, b.interval, b.quantum, b.ecn)
}

// Apply applies the configuration
func (tc *TrafficController) Apply() error {
	tc.logger.Info("Starting traffic control configuration application",
		logging.String("operation", logging.OperationApplyConfig),
		logging.Int("class_count", len(tc.classes)),
	)

	// Validation
	if err := tc.validate(); err != nil {
		tc.logger.Error("Configuration validation failed",
			logging.Error(err),
			logging.String("operation", logging.OperationValidation),
		)
		return err
	}

	tc.logger.Info("Configuration validation successful")

	// Apply configuration through the application service
	ctx := context.Background()

	// Create HTB qdisc
	handle := "1:"
	defaultClass := "1:999" // Default class for unclassified traffic
	if err := tc.service.CreateHTBQdisc(ctx, tc.deviceName, handle, defaultClass); err != nil {
		tc.logger.Error("Failed to create HTB qdisc",
			logging.Error(err),
			logging.String("device", tc.deviceName),
		)
		return fmt.Errorf("failed to create HTB qdisc: %w", err)
	}

	// Create classes
	for i, class := range tc.classes {
		classID := fmt.Sprintf("1:%d", i+10) // Start class IDs at 1:10
		parent := "1:" // Parent is the root qdisc

		tc.logger.Debug("Creating HTB class",
			logging.String("class_name", class.name),
			logging.String("class_id", classID),
			logging.String("guaranteed_bandwidth", class.guaranteedBandwidth.String()),
			logging.String("max_bandwidth", class.maxBandwidth.String()),
		)

		if err := tc.service.CreateHTBClass(ctx, tc.deviceName, parent, classID,
			class.guaranteedBandwidth.String(), class.maxBandwidth.String()); err != nil {
			tc.logger.Error("Failed to create HTB class",
				logging.Error(err),
				logging.String("class_name", class.name),
			)
			return fmt.Errorf("failed to create HTB class %s: %w", class.name, err)
		}

		// Create filters for the class
		for j, filter := range class.filters {
			priority := uint16(100 + j) // Start filter priorities at 100
			protocol := "ip"
			flowID := classID

			match := tc.buildFilterMatch(filter)
			if len(match) == 0 {
				continue // Skip unsupported filters
			}

			if err := tc.service.CreateFilter(ctx, tc.deviceName, parent, priority,
				protocol, flowID, match); err != nil {
				tc.logger.Error("Failed to create filter",
					logging.Error(err),
					logging.String("class_name", class.name),
					logging.String("filter_type", fmt.Sprintf("%v", filter.filterType)),
				)
				return fmt.Errorf("failed to create filter for class %s: %w", class.name, err)
			}
		}
	}

	// Create default class for unclassified traffic
	if err := tc.service.CreateHTBClass(ctx, tc.deviceName, "1:", "1:999",
		"1mbit", tc.totalBandwidth.String()); err != nil {
		tc.logger.Error("Failed to create default HTB class",
			logging.Error(err),
		)
		return fmt.Errorf("failed to create default HTB class: %w", err)
	}

	tc.logger.Info("Traffic control configuration applied successfully",
		logging.String("device", tc.deviceName),
		logging.String("total_bandwidth", tc.totalBandwidth.String()),
		logging.Int("classes_applied", len(tc.classes)),
	)

	return nil
}

// buildFilterMatch converts a Filter to a match map for the CQRS command
func (tc *TrafficController) buildFilterMatch(filter Filter) map[string]string {
	match := make(map[string]string)

	switch filter.filterType {
	case SourceIPFilter:
		if ip, ok := filter.value.(string); ok {
			match["src_ip"] = ip
		}
	case DestinationIPFilter:
		if ip, ok := filter.value.(string); ok {
			match["dst_ip"] = ip
		}
	case SourcePortFilter:
		if port, ok := filter.value.(int); ok {
			match["src_port"] = fmt.Sprintf("%d", port)
		}
	case DestinationPortFilter:
		if port, ok := filter.value.(int); ok {
			match["dst_port"] = fmt.Sprintf("%d", port)
		}
	case ProtocolFilter:
		if proto, ok := filter.value.(string); ok {
			match["protocol"] = proto
		}
	case ApplicationFilter:
		// Convert known applications to port ranges
		if app, ok := filter.value.(string); ok {
			switch app {
			case "ssh":
				match["dst_port"] = "22"
			case "http":
				match["dst_port"] = "80"
			case "https":
				match["dst_port"] = "443"
			case "dns":
				match["dst_port"] = "53"
			}
		}
	}

	return match
}

// validate checks if the configuration is valid
func (tc *TrafficController) validate() error {
	tc.logger.Debug("Starting configuration validation",
		logging.String("operation", logging.OperationValidation),
		logging.Int("class_count", len(tc.classes)),
	)

	if tc.totalBandwidth.BitsPerSecond() == 0 {
		tc.logger.Warn("Total bandwidth not set",
			logging.String("validation_error", "missing_total_bandwidth"),
		)
		return fmt.Errorf("total bandwidth not set. Use SetTotalBandwidth() to specify the interface bandwidth")
	}

	// Check if all classes have priority set
	for _, class := range tc.classes {
		if class.priority == nil {
			tc.logger.Warn("Traffic class missing priority",
				logging.String("class_name", class.name),
				logging.String("validation_error", "missing_priority"),
			)
			return fmt.Errorf(
				"class '%s' does not have a priority set\n"+
					"Priority is required for all traffic classes. Use WithPriority(0-7) to set it",
				class.name,
			)
		}
	}

	// Check if guaranteed bandwidth sum doesn't exceed total
	var totalGuaranteed valueobjects.Bandwidth
	for _, class := range tc.classes {
		totalGuaranteed = totalGuaranteed.Add(class.guaranteedBandwidth)

		tc.logger.Debug("Validating traffic class",
			logging.String("class_name", class.name),
			logging.String("guaranteed_bandwidth", class.guaranteedBandwidth.String()),
			logging.String("max_bandwidth", class.maxBandwidth.String()),
			logging.Int("priority", int(*class.priority)),
		)

		// Check if max bandwidth exceeds total
		if class.maxBandwidth.GreaterThan(tc.totalBandwidth) {
			tc.logger.Warn("Class max bandwidth exceeds total bandwidth",
				logging.String("class_name", class.name),
				logging.String("max_bandwidth", class.maxBandwidth.String()),
				logging.String("total_bandwidth", tc.totalBandwidth.String()),
				logging.String("validation_error", "max_exceeds_total"),
			)
			return fmt.Errorf(
				"class '%s' has max bandwidth (%s) higher than total bandwidth (%s)\n"+
					"Suggestion: Either reduce the max bandwidth or increase the total bandwidth",
				class.name,
				class.maxBandwidth,
				tc.totalBandwidth,
			)
		}

		// Check if guaranteed > max
		if class.guaranteedBandwidth.GreaterThan(class.maxBandwidth) && class.maxBandwidth.BitsPerSecond() > 0 {
			tc.logger.Warn("Class guaranteed bandwidth exceeds max bandwidth",
				logging.String("class_name", class.name),
				logging.String("guaranteed_bandwidth", class.guaranteedBandwidth.String()),
				logging.String("max_bandwidth", class.maxBandwidth.String()),
				logging.String("validation_error", "guaranteed_exceeds_max"),
			)
			return fmt.Errorf(
				"class '%s' has guaranteed bandwidth (%s) higher than max bandwidth (%s)\n"+
					"Suggestion: Set max bandwidth higher than or equal to guaranteed bandwidth",
				class.name,
				class.guaranteedBandwidth,
				class.maxBandwidth,
			)
		}
	}

	if totalGuaranteed.GreaterThan(tc.totalBandwidth) {
		tc.logger.Warn("Total guaranteed bandwidth exceeds interface bandwidth",
			logging.String("total_guaranteed", totalGuaranteed.String()),
			logging.String("total_bandwidth", tc.totalBandwidth.String()),
			logging.String("validation_error", "total_guaranteed_exceeds_total"),
		)
		return fmt.Errorf(
			"total guaranteed bandwidth (%s) exceeds interface bandwidth (%s)\n"+
				"Suggestion: Reduce guaranteed bandwidths or increase total bandwidth",
			totalGuaranteed,
			tc.totalBandwidth,
		)
	}

	tc.logger.Debug("Configuration validation completed successfully",
		logging.String("total_guaranteed", totalGuaranteed.String()),
		logging.String("total_bandwidth", tc.totalBandwidth.String()),
	)

	return nil
}

// GetStatistics retrieves current traffic control statistics
func (tc *TrafficController) GetStatistics() (*qmodels.DeviceStatisticsView, error) {
	ctx := context.Background()
	return tc.service.GetDeviceStatistics(ctx, tc.deviceName)
}

// GetRealtimeStatistics retrieves real-time statistics
func (tc *TrafficController) GetRealtimeStatistics() (*qmodels.DeviceStatisticsView, error) {
	ctx := context.Background()
	return tc.service.GetRealtimeStatistics(ctx, tc.deviceName)
}

// MonitorStatistics starts continuous monitoring of statistics
func (tc *TrafficController) MonitorStatistics(interval time.Duration, callback func(*qmodels.DeviceStatisticsView)) error {
	ctx := context.Background()
	return tc.service.MonitorStatistics(ctx, tc.deviceName, interval, callback)
}

// GetQdiscStatistics retrieves statistics for a specific qdisc
func (tc *TrafficController) GetQdiscStatistics(handle string) (*qmodels.QdiscStatisticsView, error) {
	ctx := context.Background()
	return tc.service.GetQdiscStatistics(ctx, tc.deviceName, handle)
}

// GetClassStatistics retrieves statistics for a specific class
func (tc *TrafficController) GetClassStatistics(handle string) (*qmodels.ClassStatisticsView, error) {
	ctx := context.Background()
	return tc.service.GetClassStatistics(ctx, tc.deviceName, handle)
}
