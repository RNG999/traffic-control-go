package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rng999/traffic-control-go/api"
	"github.com/rng999/traffic-control-go/pkg/logging"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// ProductionTrafficManager demonstrates a production-ready traffic control setup
// with monitoring, error handling, and graceful shutdown
type ProductionTrafficManager struct {
	controller *api.TrafficController
	config     *TrafficConfiguration
	logger     logging.Logger
	stats      *StatisticsCollector
	mu         sync.RWMutex
}

// TrafficConfiguration represents the traffic control configuration
type TrafficConfiguration struct {
	Device    string         `json:"device"`
	Bandwidth string         `json:"bandwidth"`
	Classes   []ClassConfig  `json:"classes"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// ClassConfig represents a traffic class configuration
type ClassConfig struct {
	Name       string   `json:"name"`
	Guaranteed string   `json:"guaranteed"`
	SoftLimit  string   `json:"soft_limit"`
	Priority   uint8    `json:"priority"`
	Ports      []uint16 `json:"ports,omitempty"`
	IPs        []string `json:"ips,omitempty"`
}

// StatisticsCollector collects and monitors traffic statistics
type StatisticsCollector struct {
	controller *api.TrafficController
	logger     logging.Logger
	interval   time.Duration
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func main() {
	// Initialize logging
	logging.Initialize(&logging.Config{
		Level:       "info",
		Format:      "json",
		Development: false,
		Outputs:     []string{"stdout"},
	})

	logger := logging.WithComponent("traffic-manager")
	logger.Info("Starting production traffic manager")

	// Load configuration
	config, err := loadConfiguration("traffic-config.json")
	if err != nil {
		logger.Fatal("Failed to load configuration", logging.Error(err))
	}

	// Create traffic manager
	manager, err := NewProductionTrafficManager(config)
	if err != nil {
		logger.Fatal("Failed to create traffic manager", logging.Error(err))
	}

	// Apply configuration
	if err := manager.Apply(); err != nil {
		logger.Fatal("Failed to apply traffic control", logging.Error(err))
	}

	// Start monitoring
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	manager.StartMonitoring(ctx, 10*time.Second)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutting down traffic manager")

	// Cleanup
	manager.Shutdown()
}

// NewProductionTrafficManager creates a new production traffic manager
func NewProductionTrafficManager(config *TrafficConfiguration) (*ProductionTrafficManager, error) {
	logger := logging.WithComponent("traffic-manager").
		WithDevice(config.Device)

	// Validate configuration
	if err := validateConfiguration(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create controller
	controller := api.NetworkInterface(config.Device)
	controller.WithHardLimitBandwidth(config.Bandwidth)

	return &ProductionTrafficManager{
		controller: controller,
		config:     config,
		logger:     logger,
	}, nil
}

// Apply applies the traffic control configuration
func (m *ProductionTrafficManager) Apply() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Applying traffic control configuration",
		logging.String("bandwidth", m.config.Bandwidth),
		logging.Int("class_count", len(m.config.Classes)))

	// Apply each traffic class
	for _, class := range m.config.Classes {
		tc := m.controller.CreateTrafficClass(class.Name).
			WithGuaranteedBandwidth(class.Guaranteed).
			WithSoftLimitBandwidth(class.SoftLimit).
			WithPriority(class.Priority)

		// Add port filters
		if len(class.Ports) > 0 {
			tc.ForPort(class.Ports...)
		}

		// Add IP filters
		for _, ip := range class.IPs {
			tc.ForDestination(ip)
		}

		m.logger.Debug("Configured traffic class",
			logging.String("class", class.Name),
			logging.String("guaranteed", class.Guaranteed),
			logging.String("soft_limit", class.SoftLimit),
			logging.Uint8("priority", class.Priority))
	}

	// Apply configuration with retry logic
	var lastErr error
	for i := 0; i < 3; i++ {
		if err := m.controller.Apply(); err != nil {
			lastErr = err
			m.logger.Warn("Failed to apply configuration, retrying",
				logging.Error(err),
				logging.Int("attempt", i+1))
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		m.logger.Info("Traffic control configuration applied successfully")
		return nil
	}

	return fmt.Errorf("failed after 3 attempts: %w", lastErr)
}

// StartMonitoring starts the statistics monitoring
func (m *ProductionTrafficManager) StartMonitoring(ctx context.Context, interval time.Duration) {
	statsCtx, cancel := context.WithCancel(ctx)
	m.stats = &StatisticsCollector{
		controller: m.controller,
		logger:     m.logger.WithComponent("statistics"),
		interval:   interval,
		ctx:        statsCtx,
		cancel:     cancel,
	}

	m.stats.Start()
}

// Start begins collecting statistics
func (sc *StatisticsCollector) Start() {
	sc.wg.Add(1)
	go func() {
		defer sc.wg.Done()
		ticker := time.NewTicker(sc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				sc.collect()
			case <-sc.ctx.Done():
				return
			}
		}
	}()
}

// collect gathers and logs statistics
func (sc *StatisticsCollector) collect() {
	// This is a simplified version - in production you'd get real stats
	sc.logger.Info("Collecting traffic statistics")

	// Example of what you might collect:
	// - Bandwidth usage per class
	// - Packet drop rates
	// - Queue lengths
	// - Latency measurements

	// Alert on issues
	// if dropRate > threshold {
	//     sc.logger.Warn("High packet drop rate detected",
	//         logging.Float64("drop_rate", dropRate))
	// }
}

// Shutdown gracefully shuts down the traffic manager
func (m *ProductionTrafficManager) Shutdown() {
	m.logger.Info("Shutting down traffic manager")

	// Stop monitoring
	if m.stats != nil {
		m.stats.cancel()
		m.stats.wg.Wait()
	}

	// In production, you might want to:
	// 1. Save current statistics
	// 2. Clear traffic control rules
	// 3. Notify monitoring systems
}

// UpdateConfiguration updates the traffic control configuration
func (m *ProductionTrafficManager) UpdateConfiguration(newConfig *TrafficConfiguration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate new configuration
	if err := validateConfiguration(newConfig); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Store old config for rollback
	oldConfig := m.config

	// Apply new configuration
	m.config = newConfig
	m.controller = api.NetworkInterface(newConfig.Device)
	m.controller.WithHardLimitBandwidth(newConfig.Bandwidth)

	if err := m.Apply(); err != nil {
		// Rollback on failure
		m.logger.Error("Failed to apply new configuration, rolling back",
			logging.Error(err))
		m.config = oldConfig
		m.controller = api.NetworkInterface(oldConfig.Device)
		m.controller.WithHardLimitBandwidth(oldConfig.Bandwidth)
		
		// Try to restore old configuration
		if rollbackErr := m.Apply(); rollbackErr != nil {
			m.logger.Error("Rollback failed",
				logging.Error(rollbackErr))
		}
		
		return err
	}

	m.logger.Info("Configuration updated successfully")
	return nil
}

// validateConfiguration validates the traffic control configuration
func validateConfiguration(config *TrafficConfiguration) error {
	if config.Device == "" {
		return fmt.Errorf("device name is required")
	}

	// Parse and validate total bandwidth
	totalBW, err := tc.ParseBandwidth(config.Bandwidth)
	if err != nil {
		return fmt.Errorf("invalid total bandwidth: %w", err)
	}

	// Track guaranteed bandwidth sum
	var guaranteedSum uint64
	priorities := make(map[uint8]bool)

	for _, class := range config.Classes {
		// Check unique priorities
		if priorities[class.Priority] {
			return fmt.Errorf("duplicate priority %d", class.Priority)
		}
		priorities[class.Priority] = true

		// Validate priority range
		if class.Priority > 7 {
			return fmt.Errorf("priority must be 0-7, got %d", class.Priority)
		}

		// Parse and validate bandwidths
		guaranteed, err := tc.ParseBandwidth(class.Guaranteed)
		if err != nil {
			return fmt.Errorf("class %s: invalid guaranteed bandwidth: %w",
				class.Name, err)
		}

		softLimit, err := tc.ParseBandwidth(class.SoftLimit)
		if err != nil {
			return fmt.Errorf("class %s: invalid soft limit: %w",
				class.Name, err)
		}

		// Validate bandwidth relationships
		if softLimit.Bps() < guaranteed.Bps() {
			return fmt.Errorf("class %s: soft limit must be >= guaranteed",
				class.Name)
		}

		guaranteedSum += guaranteed.Bps()
	}

	// Check for oversubscription
	if guaranteedSum > totalBW.Bps() {
		return fmt.Errorf("total guaranteed bandwidth (%d) exceeds total (%d)",
			guaranteedSum, totalBW.Bps())
	}

	return nil
}

// loadConfiguration loads configuration from a JSON file
func loadConfiguration(filename string) (*TrafficConfiguration, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config TrafficConfiguration
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	config.UpdatedAt = time.Now()
	return &config, nil
}

// Example configuration file (traffic-config.json):
/*
{
  "device": "eth0",
  "bandwidth": "1gbps",
  "classes": [
    {
      "name": "critical",
      "guaranteed": "400mbps",
      "soft_limit": "600mbps",
      "priority": 0,
      "ports": [22, 3306]
    },
    {
      "name": "web",
      "guaranteed": "300mbps",
      "soft_limit": "500mbps",
      "priority": 2,
      "ports": [80, 443]
    },
    {
      "name": "background",
      "guaranteed": "100mbps",
      "soft_limit": "200mbps",
      "priority": 7
    }
  ]
}
*/