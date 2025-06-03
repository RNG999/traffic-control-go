//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"

	"github.com/rng999/traffic-control-go/api"
)

func main() {
	fmt.Println("Priority Options Demo")
	fmt.Println("====================\n")

	// Demonstrate different priority options
	controller := api.NetworkInterface("eth0")
	controller.WithHardLimitBandwidth("1Gbps")

	// Using numeric priorities
	controller.CreateTrafficClass("Critical Traffic").
		WithGuaranteedBandwidth("100Mbps").
		WithPriority(0).     // Highest priority
		ForPort(5060, 5061)  // SIP/VoIP
		
	controller.CreateTrafficClass("Interactive Traffic").
		WithGuaranteedBandwidth("100Mbps").
		WithPriority(1). // High priority
		ForPort(22)      // SSH
		
	controller.CreateTrafficClass("Normal Traffic").
		WithGuaranteedBandwidth("400Mbps").
		WithPriority(4).  // Must set explicit priority
		ForPort(80, 443)  // HTTP/HTTPS
		
	controller.CreateTrafficClass("Background Traffic").
		WithGuaranteedBandwidth("100Mbps").
		WithPriority(6). // Low priority
		ForPort(873)     // rsync
		
	controller.CreateTrafficClass("Database Traffic").
		WithGuaranteedBandwidth("200Mbps").
		WithPriority(3). // Medium-high priority
		ForPort(3306)    // MySQL
		
	controller.CreateTrafficClass("Bulk Transfer").
		WithGuaranteedBandwidth("100Mbps").
		WithPriority(7). // Lowest priority
		ForPort(6881, 6882, 6883) // BitTorrent

	err := controller.Apply()

	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Println("✓ Priority-based traffic control configured successfully")
	}

	fmt.Println("\nPriority Values in Linux HTB:")
	fmt.Println("- Priority range: 0-7")
	fmt.Println("- Lower number = Higher priority")
	fmt.Println("- 0 = Highest priority (critical)")
	fmt.Println("- 4 = Normal priority (default)")
	fmt.Println("- 7 = Lowest priority (background)")
}