package main

import (
	"fmt"
	"log"
	
	"github.com/rng999/traffic-control-go/api"
)

func main() {
	// Demonstrate the human-readable API
	fmt.Println("Traffic Control Library Demo")
	fmt.Println("============================")
	
	// Example 1: Basic traffic shaping
	fmt.Println("\n1. Basic Traffic Shaping Example:")
	basicExample()
	
	// Example 2: Priority-based traffic control
	fmt.Println("\n2. Priority-based Traffic Control:")
	priorityExample()
	
	// Example 3: Server-specific traffic control
	fmt.Println("\n3. Server-specific Traffic Control:")
	serverExample()
	
	// Example 4: Configuration file based
	fmt.Println("\n4. Configuration File Example:")
	configExample()
}

func basicExample() {
	controller := api.New("eth0")
	
	err := controller.
		SetTotalBandwidth("100Mbps").
		CreateTrafficClass("Web Services").
			WithGuaranteedBandwidth("30Mbps").
			WithBurstableTo("60Mbps").
			WithPriority(4). // Normal priority
			ForPort(80, 443).
			And().
		CreateTrafficClass("SSH Management").
			WithGuaranteedBandwidth("5Mbps").
			WithBurstableTo("10Mbps").
			WithPriority(1). // High priority
			ForPort(22).
			Apply()
	
	if err != nil {
		log.Printf("Configuration error: %v", err)
	} else {
		fmt.Println("✓ Basic traffic shaping configured successfully")
	}
}

func priorityExample() {
	controller := api.New("eth0")
	
	// Configure priority groups separately
	controller.SetTotalBandwidth("1Gbps")
	
	// This is a conceptual example - the actual implementation would need
	// proper integration between priority groups and the controller
	err := controller.
		CreateTrafficClass("High Priority").
			WithGuaranteedBandwidth("300Mbps").
			WithPriority(1). // High priority
			ForPort(22). // SSH
			And().
		CreateTrafficClass("Medium Priority").
			WithGuaranteedBandwidth("500Mbps").
			WithPriority(4). // Medium priority
			ForPort(80, 443). // HTTP/HTTPS
			And().
		CreateTrafficClass("Low Priority").
			WithGuaranteedBandwidth("200Mbps").
			WithPriority(6). // Low priority
			ForApplication("torrent").
			Apply()
	
	if err != nil {
		log.Printf("Configuration error: %v", err)
	} else {
		fmt.Println("✓ Priority-based traffic control configured")
	}
}

func serverExample() {
	controller := api.New("eth0")
	
	err := controller.
		SetTotalBandwidth("500Mbps").
		CreateTrafficClass("Database Server").
			WithGuaranteedBandwidth("100Mbps").
			WithBurstableTo("200Mbps").
			WithPriority(1). // High priority
			ForDestination("192.168.1.10").
			And().
		CreateTrafficClass("Web Servers").
			WithGuaranteedBandwidth("150Mbps").
			WithBurstableTo("300Mbps").
			WithPriority(3). // Normal-high priority
			ForDestination("192.168.1.20").
			ForDestination("192.168.1.21").
			And().
		CreateTrafficClass("Backup Traffic").
			WithGuaranteedBandwidth("50Mbps").
			WithPriority(6). // Low priority
			ForSource("192.168.1.100").
			Apply()
	
	if err != nil {
		log.Printf("Configuration error: %v", err)
	} else {
		fmt.Println("✓ Server-specific traffic control configured")
	}
}

// Example of error handling
func errorExample() {
	controller := api.New("eth0")
	
	// This will fail due to over-allocation
	err := controller.
		SetTotalBandwidth("100Mbps").
		CreateTrafficClass("Service1").
			WithGuaranteedBandwidth("60Mbps").
			And().
		CreateTrafficClass("Service2").
			WithGuaranteedBandwidth("50Mbps").
			Apply()
	
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
		fmt.Println("The library correctly validates bandwidth allocation!")
	}
}

// Configuration file example
func configExample() {
	// Example 1: Load from YAML
	fmt.Println("Loading configuration from YAML file...")
	err := api.LoadAndApplyYAML("examples/config-example.yaml", "")
	if err != nil {
		log.Printf("YAML configuration error: %v", err)
	} else {
		fmt.Println("✓ YAML configuration applied successfully")
	}
	
	// Example 2: Load config and modify before applying
	config, err := api.LoadConfigFromYAML("examples/config-example.yaml")
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		return
	}
	
	// Modify configuration programmatically
	config.Device = "eth1"  // Change device
	config.Classes = append(config.Classes, api.TrafficClassConfig{
		Name:       "emergency",
		Guaranteed: "50Mbps",
		Maximum:    "100Mbps",
		Priority:   &[]int{0}[0], // Highest priority
	})
	
	// Apply modified configuration
	controller := api.New(config.Device)
	err = controller.ApplyConfig(config)
	if err != nil {
		log.Printf("Configuration error: %v", err)
	} else {
		fmt.Println("✓ Modified configuration applied successfully")
	}
	
	// Example 3: Build configuration programmatically
	customConfig := &api.TrafficControlConfig{
		Version:   "1.0",
		Device:    "eth0",
		Bandwidth: "100Mbps",
		Classes: []api.TrafficClassConfig{
			{
				Name:       "realtime",
				Guaranteed: "40Mbps",
				Maximum:    "60Mbps",
				Priority:   &[]int{1}[0], // High priority
			},
			{
				Name:       "interactive",
				Guaranteed: "30Mbps",
				Maximum:    "50Mbps",
				Priority:   &[]int{4}[0], // Default priority
			},
			{
				Name:       "bulk",
				Guaranteed: "30Mbps",
				Priority:   &[]int{6}[0], // Low priority
			},
		},
		Rules: []api.TrafficRuleConfig{
			{
				Name: "ssh_traffic",
				Match: api.MatchConfig{
					DestPort: []int{22},
				},
				Target:   "realtime",
				Priority: 1,
			},
			{
				Name: "web_traffic",
				Match: api.MatchConfig{
					DestPort: []int{80, 443},
				},
				Target: "interactive",
			},
		},
	}
	
	controller = api.New(customConfig.Device)
	err = controller.ApplyConfig(customConfig)
	if err != nil {
		log.Printf("Custom configuration error: %v", err)
	} else {
		fmt.Println("✓ Custom configuration applied successfully")
	}
}