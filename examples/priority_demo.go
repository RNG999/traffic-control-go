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

	// Using numeric priorities
	err := controller.
		WithHardLimitBandwidth("1Gbps").
		CreateTrafficClass("Critical Traffic").
		WithGuaranteedBandwidth("100Mbps").
		WithPriority(0).     // Highest priority
		ForPort(5060, 5061). // SIP/VoIP
		Done().
		CreateTrafficClass("Interactive Traffic").
		WithGuaranteedBandwidth("100Mbps").
		WithPriority(1). // High priority
		ForPort(22).     // SSH
		Done().
		CreateTrafficClass("Normal Traffic").
		WithGuaranteedBandwidth("400Mbps").
		WithPriority(4).  // Must set explicit priority
		ForPort(80, 443). // HTTP/HTTPS
		Done().
		CreateTrafficClass("Background Traffic").
		WithGuaranteedBandwidth("100Mbps").
		WithPriority(6). // Low priority
		ForPort(873).    // rsync
		Done().
		CreateTrafficClass("Database Traffic").
		WithGuaranteedBandwidth("200Mbps").
		WithPriority(3). // Medium-high priority
		ForPort(3306).   // MySQL
		Done().
		CreateTrafficClass("Bulk Transfer").
		WithGuaranteedBandwidth("100Mbps").
		WithPriority(7). // Lowest priority
		ForApplication("torrent").
		Apply()

	if err != nil {
		log.Printf("Error: %v\n", err)
	}

	fmt.Println("\nPriority Values in Linux HTB:")
	fmt.Println("- Priority range: 0-7")
	fmt.Println("- Lower number = Higher priority")
	fmt.Println("- Default priority = 4")
	fmt.Println("\nRecommended priority assignments:")
	fmt.Println("- 0: Critical real-time traffic (VoIP, video conferencing)")
	fmt.Println("- 1-2: Interactive traffic (SSH, remote desktop)")
	fmt.Println("- 3-4: Normal traffic (web, email)")
	fmt.Println("- 5-6: Background traffic (updates, backups)")
	fmt.Println("- 7: Bulk/lowest priority traffic")

	// Configuration file example with numeric priorities
	configExample := `
# YAML configuration with numeric priorities
version: "1.0"
device: eth0
bandwidth: 1Gbps
classes:
  - name: voip
    guaranteed: 100Mbps
    priority: 0    # Highest priority
  - name: interactive
    guaranteed: 300Mbps
    priority: 2    # High priority
  - name: bulk
    guaranteed: 200Mbps
    priority: 6    # Low priority
`
	fmt.Println("\nExample YAML configuration:")
	fmt.Println(configExample)
}
