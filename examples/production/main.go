package main

import (
	"fmt"
	"os"

	"github.com/rng999/traffic-control-go/api"
	"github.com/rng999/traffic-control-go/pkg/logging"
)

// ProductionExample demonstrates a simple traffic control setup
func main() {
	// Initialize logging
	if err := logging.Initialize(logging.Config{
		Level:       "info",
		Format:      "json",
		Development: false,
	}); err != nil {
		panic(fmt.Sprintf("Failed to initialize logging: %v", err))
	}

	logger := logging.WithComponent("traffic-manager")
	logger.Info("Starting production traffic manager")

	// Check if running as root
	if os.Geteuid() != 0 {
		logger.Error("Traffic control requires root privileges")
		os.Exit(1)
	}

	// Create traffic controller for ethernet interface
	device := "eth0"
	controller := api.NetworkInterface(device)
	
	// Set interface bandwidth limit
	controller.WithHardLimitBandwidth("1gbps")

	// Create high priority class for SSH
	controller.CreateTrafficClass("SSH Management").
		WithGuaranteedBandwidth("50mbps").
		WithSoftLimitBandwidth("100mbps").
		WithPriority(0).
		ForPort(22)

	// Create medium priority class for web traffic
	controller.CreateTrafficClass("Web Services").
		WithGuaranteedBandwidth("300mbps").
		WithSoftLimitBandwidth("600mbps").
		WithPriority(1).
		ForPort(80, 443)

	// Apply traffic control configuration
	if err := controller.Apply(); err != nil {
		logger.Error("Failed to apply traffic control", logging.Error(err))
		os.Exit(1)
	}

	logger.Info("Traffic control configuration applied successfully")

	// Get statistics
	stats, err := controller.GetStatistics()
	if err != nil {
		logger.Error("Failed to get statistics", logging.Error(err))
	} else {
		logger.Info("Current statistics", 
			logging.String("device", stats.DeviceName),
			logging.Int("total_classes", len(stats.ClassStats)))
	}
}