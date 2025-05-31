package api

import (
	"fmt"

	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/pkg/logging"
)

// TrafficController is the main entry point for traffic control configuration
type TrafficController struct {
	deviceName     string
	totalBandwidth valueobjects.Bandwidth
	classes        []*TrafficClass
	logger         logging.Logger
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

	return &TrafficController{
		deviceName: deviceName,
		classes:    make([]*TrafficClass, 0),
		logger:     logger,
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

	// TODO: Implement actual TC commands via netlink
	fmt.Printf("Applying traffic control configuration to %s\n", tc.deviceName)
	fmt.Printf("Total bandwidth: %s\n", tc.totalBandwidth)

	for _, class := range tc.classes {
		priority := "<not set>"
		if class.priority != nil {
			priority = fmt.Sprintf("%d", *class.priority)
		}

		tc.logger.Debug("Applying traffic class configuration",
			logging.String("class_name", class.name),
			logging.String("guaranteed_bandwidth", class.guaranteedBandwidth.String()),
			logging.String("max_bandwidth", class.maxBandwidth.String()),
			logging.String("priority", priority),
		)

		fmt.Printf("  Class '%s': guaranteed=%s, max=%s, priority=%s\n",
			class.name,
			class.guaranteedBandwidth,
			class.maxBandwidth,
			priority,
		)
	}

	tc.logger.Info("Traffic control configuration applied successfully",
		logging.String("device", tc.deviceName),
		logging.String("total_bandwidth", tc.totalBandwidth.String()),
		logging.Int("classes_applied", len(tc.classes)),
	)

	return nil
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
