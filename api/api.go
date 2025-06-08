package api

import (
	"context"
	"fmt"
	"time"

	"github.com/rng999/traffic-control-go/internal/application"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	qmodels "github.com/rng999/traffic-control-go/internal/queries/models"
	"github.com/rng999/traffic-control-go/pkg/logging"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// TrafficController is the main entry point for traffic control configuration
type TrafficController struct {
	deviceName      string
	totalBandwidth  tc.Bandwidth
	classes         []*TrafficClass
	pendingBuilders []*TrafficClassBuilder
	logger          logging.Logger
	service         *application.TrafficControlService
}

// TrafficClass represents a traffic classification with its rules
type TrafficClass struct {
	name                string
	guaranteedBandwidth tc.Bandwidth
	maxBandwidth        tc.Bandwidth
	priority            *uint8 // Priority is now required and must be explicitly set (0-7, where 0 is highest)
	filters             []Filter
}

// Priority型は削除: uint8を直接使用

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
)

// NetworkInterface creates a new traffic controller for a network interface
func NetworkInterface(deviceName string) *TrafficController {
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

// WithHardLimitBandwidth sets the absolute physical bandwidth limit for the network interface
func (controller *TrafficController) WithHardLimitBandwidth(bandwidth string) *TrafficController {
	controller.logger.Info("Setting hard limit bandwidth",
		logging.String("bandwidth", bandwidth),
		logging.String("operation", logging.OperationConfigLoad),
	)

	controller.totalBandwidth = tc.MustParseBandwidth(bandwidth)
	return controller
}

// CreateTrafficClass creates a new traffic class with a human-readable name
func (controller *TrafficController) CreateTrafficClass(name string) *TrafficClassBuilder {
	controller.logger.Info("Creating traffic class",
		logging.String("class_name", name),
		logging.String("operation", logging.OperationCreateClass),
	)

	class := &TrafficClass{
		name: name,
		// priority is nil by default - must be set explicitly
	}

	builder := &TrafficClassBuilder{
		controller: controller,
		class:      class,
	}

	// Add to pending builders list for automatic registration on Apply()
	controller.pendingBuilders = append(controller.pendingBuilders, builder)

	return builder
}

// TrafficClassBuilder provides a fluent interface for building traffic classes
type TrafficClassBuilder struct {
	controller *TrafficController
	class      *TrafficClass
	finalized  bool
}

// WithGuaranteedBandwidth sets the minimum guaranteed bandwidth
func (b *TrafficClassBuilder) WithGuaranteedBandwidth(bandwidth string) *TrafficClassBuilder {
	b.class.guaranteedBandwidth = tc.MustParseBandwidth(bandwidth)
	return b
}

// WithSoftLimitBandwidth sets the policy-based bandwidth limit (borrowing allowed)
func (b *TrafficClassBuilder) WithSoftLimitBandwidth(bandwidth string) *TrafficClassBuilder {
	b.class.maxBandwidth = tc.MustParseBandwidth(bandwidth)
	return b
}

// WithPriority sets the traffic class to a specific priority level (0-7)
func (b *TrafficClassBuilder) WithPriority(priority int) *TrafficClassBuilder {
	// HTB supports priority values 0-7, where lower numbers = higher priority
	if priority < 0 {
		priority = 0
	} else if priority > 7 {
		priority = 7
	}
	// #nosec G115 -- priority is explicitly clamped to 0-7 range above
	p := uint8(priority)
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

// ForDestinationIPs adds multiple destination IP filters
func (b *TrafficClassBuilder) ForDestinationIPs(ips ...string) *TrafficClassBuilder {
	for _, ip := range ips {
		b.ForDestination(ip)
	}
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

// ForSourceIPs adds multiple source IP filters
func (b *TrafficClassBuilder) ForSourceIPs(ips ...string) *TrafficClassBuilder {
	for _, ip := range ips {
		b.ForSource(ip)
	}
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

// ForProtocols adds protocol filters
func (b *TrafficClassBuilder) ForProtocols(protocols ...string) *TrafficClassBuilder {
	for _, protocol := range protocols {
		b.class.filters = append(b.class.filters, Filter{
			filterType: ProtocolFilter,
			value:      protocol,
		})
	}
	return b
}

// Apply completes the builder and adds the class to the controller
func (b *TrafficClassBuilder) Apply() error {
	return b.controller.Apply()
}

// CreateHTBQdisc creates an HTB (Hierarchical Token Bucket) qdisc with fluent interface
func (controller *TrafficController) CreateHTBQdisc(handle, defaultClass string) *HTBQdiscBuilder {
	return &HTBQdiscBuilder{
		controller:   controller,
		handle:       handle,
		defaultClass: defaultClass,
	}
}

// CreateTBFQdisc creates a TBF (Token Bucket Filter) qdisc with fluent interface
func (controller *TrafficController) CreateTBFQdisc(handle, rate string) *TBFQdiscBuilder {
	return &TBFQdiscBuilder{
		controller: controller,
		handle:     handle,
		rate:       rate,
		buffer:     32768, // default buffer
		limit:      10000, // default limit
		burst:      0,     // will be calculated if not set
	}
}

// CreatePRIOQdisc creates a PRIO (Priority Scheduler) qdisc with fluent interface
func (controller *TrafficController) CreatePRIOQdisc(handle string, bands uint8) *PRIOQdiscBuilder {
	return &PRIOQdiscBuilder{
		controller: controller,
		handle:     handle,
		bands:      bands,
		priomap:    []uint8{1, 2, 2, 2, 1, 2, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1}, // default priomap
	}
}

// CreateFQCODELQdisc creates a FQ_CODEL (Fair Queuing Controlled Delay) qdisc with fluent interface
func (controller *TrafficController) CreateFQCODELQdisc(handle string) *FQCODELQdiscBuilder {
	return &FQCODELQdiscBuilder{
		controller: controller,
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

// AddClass adds an HTB class to the qdisc
func (b *HTBQdiscBuilder) AddClass(parent, handle, name, rate, ceil string) *HTBQdiscBuilder {
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

// finalizePendingClasses automatically registers all pending class builders
func (controller *TrafficController) finalizePendingClasses() {
	for _, builder := range controller.pendingBuilders {
		if !builder.finalized {
			controller.classes = append(controller.classes, builder.class)
			builder.finalized = true
		}
	}
	controller.pendingBuilders = nil // Clear pending builders
}

// Apply applies the configuration
func (controller *TrafficController) Apply() error {
	// Finalize any pending class builders
	controller.finalizePendingClasses()

	controller.logger.Info("Starting traffic control configuration application",
		logging.String("operation", logging.OperationApplyConfig),
		logging.Int("class_count", len(controller.classes)),
	)

	// Validation
	if err := controller.validate(); err != nil {
		controller.logger.Error("Configuration validation failed",
			logging.Error(err),
			logging.String("operation", logging.OperationValidation),
		)
		return err
	}

	controller.logger.Info("Configuration validation successful")

	// Apply configuration through the application service
	ctx := context.Background()

	// Create HTB qdisc
	handle := "1:0"
	defaultClass := "1:999" // Default class for unclassified traffic
	if err := controller.service.CreateHTBQdisc(ctx, controller.deviceName, handle, defaultClass); err != nil {
		controller.logger.Error("Failed to create HTB qdisc",
			logging.Error(err),
			logging.String("device", controller.deviceName),
		)
		return fmt.Errorf("failed to create HTB qdisc: %w", err)
	}

	// Create classes
	for i, class := range controller.classes {
		classID := fmt.Sprintf("1:%d", int(*class.priority)+10) // Use priority to determine handle (1:10-1:17)
		parent := "1:0"                                         // Parent is the root qdisc

		controller.logger.Debug("Creating HTB class",
			logging.String("class_name", class.name),
			logging.String("class_id", classID),
			logging.String("guaranteed_bandwidth", class.guaranteedBandwidth.String()),
			logging.String("max_bandwidth", class.maxBandwidth.String()),
		)

		if err := controller.service.CreateHTBClass(ctx, controller.deviceName, parent, classID,
			class.guaranteedBandwidth.String(), class.maxBandwidth.String()); err != nil {
			controller.logger.Error("Failed to create HTB class",
				logging.Error(err),
				logging.String("class_name", class.name),
			)
			return fmt.Errorf("failed to create HTB class %s: %w", class.name, err)
		}

		// Create filters for the class
		if len(class.filters) == 0 {
			// Create a catch-all filter if no specific filters are defined
			priority := uint16(100) // Default priority for catch-all
			protocol := "ip"
			flowID := classID
			match := make(map[string]string) // Empty match = catch all

			if err := controller.service.CreateFilter(ctx, controller.deviceName, parent, priority,
				protocol, flowID, match); err != nil {
				controller.logger.Error("Failed to create catch-all filter",
					logging.Error(err),
					logging.String("class_name", class.name),
				)
				return fmt.Errorf("failed to create catch-all filter for class %s: %w", class.name, err)
			}
		} else {
			// Create explicit filters
			for j, filter := range class.filters {
				// Use different priority ranges for each class to avoid conflicts
				// Check for potential overflow before conversion
				baseValue := 100 + i*10
				if baseValue > 65525 || j > 9 { // Prevent overflow
					return fmt.Errorf("too many filters or classes: would overflow uint16")
				}
				// #nosec G115 -- overflow check performed above
				basePriority := uint16(baseValue) // Class 0: 100-109, Class 1: 110-119, etc.
				// #nosec G115 -- overflow check performed above
				priority := basePriority + uint16(j)
				protocol := "ip"
				flowID := classID

				match := controller.buildFilterMatch(filter)
				if len(match) == 0 {
					continue // Skip unsupported filters
				}

				if err := controller.service.CreateFilter(ctx, controller.deviceName, parent, priority,
					protocol, flowID, match); err != nil {
					controller.logger.Error("Failed to create filter",
						logging.Error(err),
						logging.String("class_name", class.name),
						logging.String("filter_type", fmt.Sprintf("%v", filter.filterType)),
					)
					return fmt.Errorf("failed to create filter for class %s: %w", class.name, err)
				}
			}
		}
	}

	// Create default class for unclassified traffic
	if err := controller.service.CreateHTBClass(ctx, controller.deviceName, "1:0", "1:999",
		"1mbit", controller.totalBandwidth.String()); err != nil {
		controller.logger.Error("Failed to create default HTB class",
			logging.Error(err),
		)
		return fmt.Errorf("failed to create default HTB class: %w", err)
	}

	controller.logger.Info("Traffic control configuration applied successfully",
		logging.String("device", controller.deviceName),
		logging.String("total_bandwidth", controller.totalBandwidth.String()),
		logging.Int("classes_applied", len(controller.classes)),
	)

	return nil
}

// buildFilterMatch converts a Filter to a match map for the CQRS command
func (controller *TrafficController) buildFilterMatch(filter Filter) map[string]string {
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
	}

	return match
}

// validate checks if the configuration is valid
func (controller *TrafficController) validate() error {
	controller.logger.Debug("Starting configuration validation",
		logging.String("operation", logging.OperationValidation),
		logging.Int("class_count", len(controller.classes)),
	)

	if controller.totalBandwidth.BitsPerSecond() == 0 {
		controller.logger.Warn("Total bandwidth not set",
			logging.String("validation_error", "missing_total_bandwidth"),
		)
		return fmt.Errorf("total bandwidth not set. Use WithHardLimitBandwidth() to specify the interface bandwidth")
	}

	// Check if all classes have priority set
	for _, class := range controller.classes {
		if class.priority == nil {
			controller.logger.Warn("Traffic class missing priority",
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
	var totalGuaranteed tc.Bandwidth
	for _, class := range controller.classes {
		totalGuaranteed = totalGuaranteed.Add(class.guaranteedBandwidth)

		controller.logger.Debug("Validating traffic class",
			logging.String("class_name", class.name),
			logging.String("guaranteed_bandwidth", class.guaranteedBandwidth.String()),
			logging.String("max_bandwidth", class.maxBandwidth.String()),
			logging.Int("priority", int(*class.priority)),
		)

		// Check if max bandwidth exceeds total
		if class.maxBandwidth.GreaterThan(controller.totalBandwidth) {
			controller.logger.Warn("Class max bandwidth exceeds total bandwidth",
				logging.String("class_name", class.name),
				logging.String("max_bandwidth", class.maxBandwidth.String()),
				logging.String("total_bandwidth", controller.totalBandwidth.String()),
				logging.String("validation_error", "max_exceeds_total"),
			)
			return fmt.Errorf(
				"class '%s' has max bandwidth (%s) higher than total bandwidth (%s)\n"+
					"Suggestion: Either reduce the max bandwidth or increase the total bandwidth",
				class.name,
				class.maxBandwidth,
				controller.totalBandwidth,
			)
		}

		// Check if guaranteed > max
		if class.guaranteedBandwidth.GreaterThan(class.maxBandwidth) && class.maxBandwidth.BitsPerSecond() > 0 {
			controller.logger.Warn("Class guaranteed bandwidth exceeds max bandwidth",
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

	if totalGuaranteed.GreaterThan(controller.totalBandwidth) {
		controller.logger.Warn("Total guaranteed bandwidth exceeds interface bandwidth",
			logging.String("total_guaranteed", totalGuaranteed.String()),
			logging.String("total_bandwidth", controller.totalBandwidth.String()),
			logging.String("validation_error", "total_guaranteed_exceeds_total"),
		)
		return fmt.Errorf(
			"total guaranteed bandwidth (%s) exceeds interface bandwidth (%s)\n"+
				"Suggestion: Reduce guaranteed bandwidths or increase total bandwidth",
			totalGuaranteed,
			controller.totalBandwidth,
		)
	}

	controller.logger.Debug("Configuration validation completed successfully",
		logging.String("total_guaranteed", totalGuaranteed.String()),
		logging.String("total_bandwidth", controller.totalBandwidth.String()),
	)

	return nil
}

// GetStatistics retrieves current traffic control statistics
func (controller *TrafficController) GetStatistics() (*qmodels.DeviceStatisticsView, error) {
	ctx := context.Background()
	return controller.service.GetDeviceStatistics(ctx, controller.deviceName)
}

// GetRealtimeStatistics retrieves real-time statistics
func (controller *TrafficController) GetRealtimeStatistics() (*qmodels.DeviceStatisticsView, error) {
	ctx := context.Background()
	return controller.service.GetRealtimeStatistics(ctx, controller.deviceName)
}

// MonitorStatistics starts continuous monitoring of statistics
func (controller *TrafficController) MonitorStatistics(interval time.Duration, callback func(*qmodels.DeviceStatisticsView)) error {
	ctx := context.Background()
	return controller.service.MonitorStatistics(ctx, controller.deviceName, interval, callback)
}

// GetQdiscStatistics retrieves statistics for a specific qdisc
func (controller *TrafficController) GetQdiscStatistics(handle string) (*qmodels.QdiscStatisticsView, error) {
	ctx := context.Background()
	return controller.service.GetQdiscStatistics(ctx, controller.deviceName, handle)
}

// GetClassStatistics retrieves statistics for a specific class
func (controller *TrafficController) GetClassStatistics(handle string) (*qmodels.ClassStatisticsView, error) {
	ctx := context.Background()
	return controller.service.GetClassStatistics(ctx, controller.deviceName, handle)
}
