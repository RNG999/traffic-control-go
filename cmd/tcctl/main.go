package main

import (
	"fmt"
	"log"

	tc "github.com/rng999/traffic-control-go/api"
	"github.com/rng999/traffic-control-go/pkg/logging"
)

func main() {
	// Initialize logging - try to load from environment, fallback to development config
	if err := logging.InitializeFromEnv(); err != nil {
		if err := logging.InitializeDevelopment(); err != nil {
			log.Fatalf("Failed to initialize logging: %v", err)
		}
	}

	// Ensure logs are flushed on exit
	defer logging.Sync()

	logger := logging.WithComponent(logging.ComponentAPI)
	logger.Info("Starting Traffic Control CLI Tool")

	fmt.Println("Traffic Control Command Line Tool (tcctl)")
	fmt.Println("=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=")

	// Example 1: Simple bandwidth limiting
	logger.Info("Running example 1: Simple bandwidth limiting")
	example1()

	// Example 2: Home network fair sharing
	logger.Info("Running example 2: Home network fair sharing")
	example2()

	// Example 3: Web server optimization
	logger.Info("Running example 3: Web server optimization")
	example3()

	logger.Info("Traffic Control CLI Tool completed")
}

func example1() {
	fmt.Println("\nExample 1: Simple Bandwidth Limiting")
	fmt.Println("-" + "-" + "-" + "-" + "-" + "-" + "-" + "-" + "-" + "-")

	controller := tc.New("eth0").
		SetTotalBandwidth("1Gbps")

	err := controller.
		CreateTrafficClass("database-traffic").
		WithGuaranteedBandwidth("100Mbps").
		WithMaxBandwidth("200Mbps").
		ForDestination("192.168.1.10").
		Apply()

	if err != nil {
		logger := logging.WithComponent(logging.ComponentAPI).WithDevice("eth0")
		logger.Error("Example 1 failed", logging.Error(err))
		log.Printf("Error: %v", err)
	}
}

func example2() {
	fmt.Println("\nExample 2: Home Network Fair Sharing")
	fmt.Println("-" + "-" + "-" + "-" + "-" + "-" + "-" + "-" + "-" + "-")

	controller := tc.New("eth0").
		SetTotalBandwidth("100Mbps")

	// Living room TV - can burst for streaming
	controller.CreateTrafficClass("living-room-tv").
		WithGuaranteedBandwidth("20Mbps").
		WithBurstableTo("40Mbps").
		ForDestination("192.168.1.50").
		And()

	// Kids' devices - limited bandwidth
	controller.CreateTrafficClass("kids-tablets").
		WithGuaranteedBandwidth("10Mbps").
		WithMaxBandwidth("20Mbps").
		WithPriority(6). // Low priority
		ForDestination("192.168.1.100").
		ForDestination("192.168.1.101").
		And()

	// Work laptop - high priority with good bandwidth
	err := controller.CreateTrafficClass("work-laptop").
		WithGuaranteedBandwidth("30Mbps").
		WithBurstableTo("80Mbps").
		WithPriority(1). // High priority
		ForDestination("192.168.1.10").
		Apply()

	if err != nil {
		logger := logging.WithComponent(logging.ComponentAPI).WithDevice("eth0")
		logger.Error("Example 1 failed", logging.Error(err))
		log.Printf("Error: %v", err)
	}
}

func example3() {
	fmt.Println("\nExample 3: Web Server Optimization")
	fmt.Println("-" + "-" + "-" + "-" + "-" + "-" + "-" + "-" + "-" + "-")

	controller := tc.New("eth0").
		SetTotalBandwidth("10Gbps")

	// Web traffic gets priority
	controller.CreateTrafficClass("web-traffic").
		WithGuaranteedBandwidth("4Gbps").
		WithBurstableTo("8Gbps").
		WithPriority(1). // High priority
		ForPort(80, 443).
		And()

	// Database replication
	controller.CreateTrafficClass("db-replication").
		WithGuaranteedBandwidth("1Gbps").
		WithMaxBandwidth("2Gbps").
		ForPort(3306).
		ForDestination("10.0.1.0/24").
		And()

	// Backup traffic - low priority
	err := controller.CreateTrafficClass("backup-traffic").
		WithGuaranteedBandwidth("500Mbps").
		WithMaxBandwidth("1Gbps").
		WithPriority(6). // Low priority
		ForPort(873).    // rsync
		Apply()

	if err != nil {
		logger := logging.WithComponent(logging.ComponentAPI).WithDevice("eth0")
		logger.Error("Example 1 failed", logging.Error(err))
		log.Printf("Error: %v", err)
	}
}

// Example showing validation errors
func exampleWithErrors() {
	fmt.Println("\nExample: Configuration Validation")
	fmt.Println("-" + "-" + "-" + "-" + "-" + "-" + "-" + "-" + "-" + "-")

	controller := tc.New("eth0").
		SetTotalBandwidth("50Mbps")

	// This will fail validation - guaranteed bandwidth exceeds total
	err := controller.
		CreateTrafficClass("video-streaming").
		WithGuaranteedBandwidth("100Mbps"). // Error: exceeds total!
		WithMaxBandwidth("200Mbps").
		ForPort(8080).
		Apply()

	if err != nil {
		fmt.Printf("Validation failed:\n%v\n", err)
	}
}
