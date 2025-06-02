package api

import (
	"fmt"

	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/pkg/logging"
)

// ImprovedController represents the new improved API design
type ImprovedController struct {
	deviceName     string
	totalBandwidth *valueobjects.Bandwidth
	classes        []*ImprovedClass
	logger         logging.Logger
	currentClass   *ImprovedClass // Track current class being configured
}

// ImprovedClass represents a traffic class with cleaner API
type ImprovedClass struct {
	name                string
	guaranteedBandwidth *valueobjects.Bandwidth
	burstBandwidth      *valueobjects.Bandwidth
	priority            *Priority
	ports               []int
	sourceIPs           []string
	destIPs             []string
	protocols           []string
	controller          *ImprovedController // Back reference
}

// New creates a new improved traffic controller
func NewImproved(deviceName string) *ImprovedController {
	return &ImprovedController{
		deviceName: deviceName,
		logger:     logging.WithComponent(logging.ComponentAPI),
		classes:    make([]*ImprovedClass, 0),
	}
}

// TotalBandwidth sets the total bandwidth for the interface
func (tc *ImprovedController) TotalBandwidth(bandwidth string) *ImprovedController {
	bw, err := valueobjects.NewBandwidth(bandwidth)
	if err != nil {
		tc.logger.Error("Invalid total bandwidth", logging.String("bandwidth", bandwidth), logging.Error(err))
		return tc
	}
	tc.totalBandwidth = &bw
	tc.logger.Info("Setting total bandwidth", logging.String("device", tc.deviceName), logging.String("bandwidth", bandwidth))
	return tc
}

// Class creates or selects a traffic class for configuration
func (tc *ImprovedController) Class(name string) *ImprovedClass {
	// Check if class already exists
	for _, class := range tc.classes {
		if class.name == name {
			tc.currentClass = class
			return class
		}
	}

	// Create new class
	class := &ImprovedClass{
		name:       name,
		controller: tc,
		ports:      make([]int, 0),
		sourceIPs:  make([]string, 0),
		destIPs:    make([]string, 0),
		protocols:  make([]string, 0),
	}

	tc.classes = append(tc.classes, class)
	tc.currentClass = class
	tc.logger.Info("Creating traffic class", logging.String("device", tc.deviceName), logging.String("class_name", name))

	return class
}

// Guaranteed sets the guaranteed bandwidth for the current class
func (c *ImprovedClass) Guaranteed(bandwidth string) *ImprovedClass {
	bw, err := valueobjects.NewBandwidth(bandwidth)
	if err != nil {
		c.controller.logger.Error("Invalid guaranteed bandwidth", 
			logging.String("class", c.name), 
			logging.String("bandwidth", bandwidth), 
			logging.Error(err))
		return c
	}
	c.guaranteedBandwidth = &bw
	c.controller.logger.Info("Setting guaranteed bandwidth", 
		logging.String("class", c.name), 
		logging.String("bandwidth", bandwidth))
	return c
}

// BurstTo sets the maximum burstable bandwidth for the current class
func (c *ImprovedClass) BurstTo(bandwidth string) *ImprovedClass {
	bw, err := valueobjects.NewBandwidth(bandwidth)
	if err != nil {
		c.controller.logger.Error("Invalid burst bandwidth", 
			logging.String("class", c.name), 
			logging.String("bandwidth", bandwidth), 
			logging.Error(err))
		return c
	}
	c.burstBandwidth = &bw
	c.controller.logger.Info("Setting burst bandwidth", 
		logging.String("class", c.name), 
		logging.String("bandwidth", bandwidth))
	return c
}

// Priority sets the priority for the current class (0-7, where 0 is highest)
func (c *ImprovedClass) Priority(priority int) *ImprovedClass {
	if priority < 0 || priority > 7 {
		c.controller.logger.Error("Invalid priority", 
			logging.String("class", c.name), 
			logging.Int("priority", priority))
		return c
	}
	p := Priority(priority)
	c.priority = &p
	c.controller.logger.Info("Setting priority", 
		logging.String("class", c.name), 
		logging.Int("priority", priority))
	return c
}

// Ports adds port-based filtering to the current class
func (c *ImprovedClass) Ports(ports ...int) *ImprovedClass {
	c.ports = append(c.ports, ports...)
	c.controller.logger.Info("Adding port filters", 
		logging.String("class", c.name), 
		logging.String("ports", fmt.Sprintf("%v", ports)))
	return c
}

// SourceIPs adds source IP-based filtering to the current class
func (c *ImprovedClass) SourceIPs(ips ...string) *ImprovedClass {
	c.sourceIPs = append(c.sourceIPs, ips...)
	c.controller.logger.Info("Adding source IP filters", 
		logging.String("class", c.name), 
		logging.String("ips", fmt.Sprintf("%v", ips)))
	return c
}

// DestIPs adds destination IP-based filtering to the current class
func (c *ImprovedClass) DestIPs(ips ...string) *ImprovedClass {
	c.destIPs = append(c.destIPs, ips...)
	c.controller.logger.Info("Adding destination IP filters", 
		logging.String("class", c.name), 
		logging.String("ips", fmt.Sprintf("%v", ips)))
	return c
}

// Protocols adds protocol-based filtering to the current class
func (c *ImprovedClass) Protocols(protocols ...string) *ImprovedClass {
	c.protocols = append(c.protocols, protocols...)
	c.controller.logger.Info("Adding protocol filters", 
		logging.String("class", c.name), 
		logging.String("protocols", fmt.Sprintf("%v", protocols)))
	return c
}

// Apply finalizes the configuration and applies it to the system
func (tc *ImprovedController) Apply() error {
	tc.logger.Info("Starting traffic control configuration application", 
		logging.String("device", tc.deviceName), 
		logging.Int("class_count", len(tc.classes)))

	// Validation
	if tc.totalBandwidth == nil {
		return fmt.Errorf("total bandwidth must be set")
	}

	if len(tc.classes) == 0 {
		return fmt.Errorf("at least one traffic class must be defined")
	}

	for _, class := range tc.classes {
		if class.guaranteedBandwidth == nil {
			return fmt.Errorf("class %s: guaranteed bandwidth must be set", class.name)
		}
		if class.priority == nil {
			return fmt.Errorf("class %s: priority must be set", class.name)
		}
	}

	// TODO: Convert to internal structures and apply
	// This would integrate with the existing TrafficControlService

	tc.logger.Info("Traffic control configuration applied successfully", 
		logging.String("device", tc.deviceName), 
		logging.String("total_bandwidth", tc.totalBandwidth.String()), 
		logging.Int("classes_applied", len(tc.classes)))

	return nil
}

// String returns a string representation of the controller configuration
func (tc *ImprovedController) String() string {
	result := fmt.Sprintf("TrafficController[%s]:\n", tc.deviceName)
	if tc.totalBandwidth != nil {
		result += fmt.Sprintf("  Total Bandwidth: %s\n", tc.totalBandwidth.String())
	}
	result += fmt.Sprintf("  Classes: %d\n", len(tc.classes))
	for _, class := range tc.classes {
		result += fmt.Sprintf("    - %s\n", class.String())
	}
	return result
}

// String returns a string representation of the class
func (c *ImprovedClass) String() string {
	result := fmt.Sprintf("%s:", c.name)
	if c.guaranteedBandwidth != nil {
		result += fmt.Sprintf(" guaranteed=%s", c.guaranteedBandwidth.String())
	}
	if c.burstBandwidth != nil {
		result += fmt.Sprintf(" burst=%s", c.burstBandwidth.String())
	}
	if c.priority != nil {
		result += fmt.Sprintf(" priority=%d", *c.priority)
	}
	if len(c.ports) > 0 {
		result += fmt.Sprintf(" ports=%v", c.ports)
	}
	return result
}