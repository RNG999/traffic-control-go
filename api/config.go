package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"gopkg.in/yaml.v3"
	
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
)

// TrafficControlConfig represents a structured configuration for traffic control
type TrafficControlConfig struct {
	Version   string                    `yaml:"version" json:"version"`
	Device    string                    `yaml:"device" json:"device"`
	Bandwidth string                    `yaml:"bandwidth" json:"bandwidth"`
	Defaults  *DefaultConfig            `yaml:"defaults,omitempty" json:"defaults,omitempty"`
	Classes   []TrafficClassConfig      `yaml:"classes" json:"classes"`
	Rules     []TrafficRuleConfig       `yaml:"rules,omitempty" json:"rules,omitempty"`
}

// DefaultConfig represents default settings
type DefaultConfig struct {
	BurstRatio float64 `yaml:"burst_ratio,omitempty" json:"burst_ratio,omitempty"` // Default: 1.5
	// Priority is no longer supported in defaults - must be set explicitly per class
}

// TrafficClassConfig represents a traffic class configuration
type TrafficClassConfig struct {
	Name       string                 `yaml:"name" json:"name"`
	Parent     string                 `yaml:"parent,omitempty" json:"parent,omitempty"`
	Guaranteed string                 `yaml:"guaranteed" json:"guaranteed"`
	Maximum    string                 `yaml:"maximum,omitempty" json:"maximum,omitempty"`
	Priority   *int                   `yaml:"priority,omitempty" json:"priority,omitempty"`
	Children   []TrafficClassConfig   `yaml:"children,omitempty" json:"children,omitempty"`
}

// TrafficRuleConfig represents a traffic rule configuration
type TrafficRuleConfig struct {
	Name     string           `yaml:"name" json:"name"`
	Match    MatchConfig      `yaml:"match" json:"match"`
	Target   string           `yaml:"target" json:"target"`
	Priority int              `yaml:"priority,omitempty" json:"priority,omitempty"`
}

// MatchConfig represents match conditions
type MatchConfig struct {
	SourceIP      string   `yaml:"source_ip,omitempty" json:"source_ip,omitempty"`
	DestinationIP string   `yaml:"destination_ip,omitempty" json:"destination_ip,omitempty"`
	SourcePort    []int    `yaml:"source_port,omitempty" json:"source_port,omitempty"`
	DestPort      []int    `yaml:"dest_port,omitempty" json:"dest_port,omitempty"`
	Protocol      string   `yaml:"protocol,omitempty" json:"protocol,omitempty"`
	Application   []string `yaml:"application,omitempty" json:"application,omitempty"`
}

// LoadConfigFromYAML loads configuration from a YAML file
func LoadConfigFromYAML(filename string) (*TrafficControlConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config TrafficControlConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return &config, nil
}

// LoadConfigFromJSON loads configuration from a JSON file
func LoadConfigFromJSON(filename string) (*TrafficControlConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config TrafficControlConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return &config, nil
}

// Validate validates the configuration
func (c *TrafficControlConfig) Validate() error {
	if c.Device == "" {
		return fmt.Errorf("device is required")
	}
	
	if c.Bandwidth == "" {
		return fmt.Errorf("bandwidth is required")
	}
	
	if len(c.Classes) == 0 {
		return fmt.Errorf("at least one class is required")
	}
	
	// Validate class names are unique
	classNames := make(map[string]bool)
	for _, class := range c.Classes {
		if err := validateClassConfig(&class, classNames, ""); err != nil {
			return err
		}
	}
	
	// Validate rules reference existing classes
	for _, rule := range c.Rules {
		if rule.Target == "" {
			return fmt.Errorf("rule %s: target is required", rule.Name)
		}
		if !classNames[rule.Target] {
			return fmt.Errorf("rule %s: target class '%s' not found", rule.Name, rule.Target)
		}
	}
	
	return nil
}

// validateClassConfig recursively validates class configuration
func validateClassConfig(class *TrafficClassConfig, classNames map[string]bool, parentPath string) error {
	if class.Name == "" {
		return fmt.Errorf("class name is required")
	}
	
	fullName := class.Name
	if parentPath != "" {
		fullName = parentPath + "." + class.Name
	}
	
	if classNames[fullName] {
		return fmt.Errorf("duplicate class name: %s", fullName)
	}
	classNames[fullName] = true
	
	if class.Guaranteed == "" {
		return fmt.Errorf("class %s: guaranteed bandwidth is required", fullName)
	}
	
	if class.Priority == nil {
		return fmt.Errorf("class %s: priority is required. Set a value between 0-7 (0=highest, 7=lowest)", fullName)
	}
	
	// Validate children recursively
	for i := range class.Children {
		if err := validateClassConfig(&class.Children[i], classNames, fullName); err != nil {
			return err
		}
	}
	
	return nil
}

// ApplyConfig applies a structured configuration using the chain API
func (tc *TrafficController) ApplyConfig(config *TrafficControlConfig) error {
	// Set device and bandwidth
	tc.deviceName = config.Device
	tc.totalBandwidth = valueobjects.MustParseBandwidth(config.Bandwidth)
	
	// Apply defaults
	defaults := config.Defaults
	if defaults == nil {
		defaults = &DefaultConfig{
			BurstRatio: 1.5,
		}
	}
	
	// Create classes
	if err := tc.createClassesFromConfig(config.Classes, defaults, ""); err != nil {
		return err
	}
	
	// Apply rules
	for _, rule := range config.Rules {
		if err := tc.createRuleFromConfig(&rule); err != nil {
			return err
		}
	}
	
	// Apply the configuration
	return tc.Apply()
}

// createClassesFromConfig recursively creates classes from configuration
func (tc *TrafficController) createClassesFromConfig(classes []TrafficClassConfig, defaults *DefaultConfig, parentName string) error {
	for _, classConfig := range classes {
		// Create full class name
		fullName := classConfig.Name
		if parentName != "" {
			fullName = parentName + "." + classConfig.Name
		}
		
		// Create class using chain API
		builder := tc.CreateTrafficClass(fullName).
			WithGuaranteedBandwidth(classConfig.Guaranteed)
		
		// Apply maximum bandwidth
		if classConfig.Maximum != "" {
			builder = builder.WithMaxBandwidth(classConfig.Maximum)
		} else if defaults.BurstRatio > 1.0 {
			// Calculate burst based on guaranteed and ratio
			guaranteed := valueobjects.MustParseBandwidth(classConfig.Guaranteed)
			burst := fmt.Sprintf("%dMbps", int(float64(guaranteed.MegabitsPerSecond())*defaults.BurstRatio))
			builder = builder.WithMaxBandwidth(burst)
		}
		
		// Apply priority - required field
		if classConfig.Priority != nil {
			builder = builder.WithPriority(*classConfig.Priority)
		}
		// Note: validation will catch missing priority later
		
		// Store the builder
		tc.classes = append(tc.classes, builder.class)
		
		// Create children
		if len(classConfig.Children) > 0 {
			if err := tc.createClassesFromConfig(classConfig.Children, defaults, fullName); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// createRuleFromConfig creates a rule from configuration
func (tc *TrafficController) createRuleFromConfig(rule *TrafficRuleConfig) error {
	// Find target class
	var targetClass *TrafficClass
	for _, class := range tc.classes {
		if class.name == rule.Target {
			targetClass = class
			break
		}
	}
	
	if targetClass == nil {
		return fmt.Errorf("target class not found: %s", rule.Target)
	}
	
	// Apply filters based on match configuration
	match := &rule.Match
	
	if match.SourceIP != "" {
		targetClass.filters = append(targetClass.filters, Filter{
			filterType: SourceIPFilter,
			value:      match.SourceIP,
		})
	}
	
	if match.DestinationIP != "" {
		targetClass.filters = append(targetClass.filters, Filter{
			filterType: DestinationIPFilter,
			value:      match.DestinationIP,
		})
	}
	
	for _, port := range match.DestPort {
		targetClass.filters = append(targetClass.filters, Filter{
			filterType: DestinationPortFilter,
			value:      port,
		})
	}
	
	for _, port := range match.SourcePort {
		targetClass.filters = append(targetClass.filters, Filter{
			filterType: SourcePortFilter,
			value:      port,
		})
	}
	
	if match.Protocol != "" {
		targetClass.filters = append(targetClass.filters, Filter{
			filterType: ProtocolFilter,
			value:      match.Protocol,
		})
	}
	
	for _, app := range match.Application {
		targetClass.filters = append(targetClass.filters, Filter{
			filterType: ApplicationFilter,
			value:      app,
		})
	}
	
	return nil
}

// LoadAndApplyYAML is a convenience method to load and apply YAML configuration
func LoadAndApplyYAML(filename string, device string) error {
	config, err := LoadConfigFromYAML(filename)
	if err != nil {
		return err
	}
	
	if device != "" {
		config.Device = device
	}
	
	tc := New(config.Device)
	return tc.ApplyConfig(config)
}

// LoadAndApplyJSON is a convenience method to load and apply JSON configuration
func LoadAndApplyJSON(filename string, device string) error {
	config, err := LoadConfigFromJSON(filename)
	if err != nil {
		return err
	}
	
	if device != "" {
		config.Device = device
	}
	
	tc := New(config.Device)
	return tc.ApplyConfig(config)
}