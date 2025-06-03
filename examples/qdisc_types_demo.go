//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"

	"github.com/rng999/traffic-control-go/api"
)

func main() {
	fmt.Println("=== Traffic Control Go - Qdisc Types Demo ===")
	fmt.Println()

	deviceName := "demo0"

	fmt.Printf("1. Creating TrafficControl for device: %s\n", deviceName)
	tc := api.NetworkInterface(deviceName)
	fmt.Println("✓ TrafficControl instance created")
	fmt.Println()

	// Demo 1: Basic Traffic Control Setup
	fmt.Println("=== Demo 1: Basic Traffic Control Setup ===")
	fmt.Println("Setting up basic traffic control with HTB qdisc")

	tc.WithHardLimitBandwidth("100Mbps")
	tc.CreateTrafficClass("Default").
		WithGuaranteedBandwidth("100Mbps").
		WithPriority(4)

	err := tc.Apply()

	if err != nil {
		log.Printf("Error applying traffic control: %v", err)
	} else {
		fmt.Println("✓ Basic traffic control configured successfully")
		fmt.Println("  - Device: demo0")
		fmt.Println("  - Total Bandwidth: 100Mbps")
		fmt.Println("  - Default Class: Configured")
	}
	fmt.Println()

	// Demo 2: Priority-based Traffic Classes
	fmt.Println("=== Demo 2: Priority-based Traffic Classes ===")
	fmt.Println("Setting up multiple traffic classes with different priorities")

	tc2 := api.NetworkInterface(deviceName)
	tc2.WithHardLimitBandwidth("1Gbps")
	
	tc2.CreateTrafficClass("High Priority").
		WithGuaranteedBandwidth("300Mbps").
		WithPriority(1).
		ForPort(22)
		
	tc2.CreateTrafficClass("Normal Priority").
		WithGuaranteedBandwidth("400Mbps").
		WithPriority(4).
		ForPort(80, 443)
		
	tc2.CreateTrafficClass("Low Priority").
		WithGuaranteedBandwidth("300Mbps").
		WithPriority(6).
		ForPort(6881, 6882, 6883) // BitTorrent

	err = tc2.Apply()

	if err != nil {
		log.Printf("Error applying priority classes: %v", err)
	} else {
		fmt.Println("✓ Priority-based traffic classes configured successfully")
		fmt.Println("  - High Priority: SSH traffic (priority 1)")
		fmt.Println("  - Normal Priority: Web traffic (priority 4)")
		fmt.Println("  - Low Priority: P2P traffic (priority 6)")
	}
	fmt.Println()

	// Demo 3: Bandwidth Shaping with Ceiling
	fmt.Println("=== Demo 3: Bandwidth Shaping with Ceiling ===")
	fmt.Println("Setting up traffic classes with guaranteed and maximum bandwidth")

	tc3 := api.NetworkInterface(deviceName)
	tc3.WithHardLimitBandwidth("500Mbps")
	
	tc3.CreateTrafficClass("Database").
		WithGuaranteedBandwidth("150Mbps").
		WithSoftLimitBandwidth("250Mbps").
		WithPriority(2).
		ForPort(3306, 5432)
		
	tc3.CreateTrafficClass("Web Services").
		WithGuaranteedBandwidth("200Mbps").
		WithSoftLimitBandwidth("350Mbps").
		WithPriority(3).
		ForPort(80, 443, 8080)

	err = tc3.Apply()

	if err != nil {
		log.Printf("Error applying bandwidth shaping: %v", err)
	} else {
		fmt.Println("✓ Bandwidth shaping configured successfully")
		fmt.Println("  - Database: 150Mbps guaranteed, 250Mbps ceiling")
		fmt.Println("  - Web Services: 200Mbps guaranteed, 350Mbps ceiling")
		fmt.Println("  - Borrowing: Enabled between classes")
	}
	fmt.Println()

	// Demo 4: Real-world use cases
	fmt.Println("=== Demo 4: Real-world Use Cases ===")
	fmt.Println()

	fmt.Println("Use Case 1: Home Internet Gateway")
	fmt.Println("- Set total bandwidth to match your ISP plan")
	fmt.Println("- Prioritize real-time traffic (VoIP, gaming)")
	fmt.Println()

	fmt.Println("Use Case 2: Enterprise Network")
	fmt.Println("- Use hierarchical traffic classes for departments")
	fmt.Println("- Guarantee minimum bandwidth for critical applications")
	fmt.Println()

	fmt.Println("Use Case 3: Data Center")
	fmt.Println("- Isolate container/VM traffic with separate classes")
	fmt.Println("- Prioritize network control traffic")
	fmt.Println()

	fmt.Println("=== Demo completed successfully! ===")
	fmt.Println()
	fmt.Println("Note: This demo shows the API usage. In a real environment:")
	fmt.Println("1. Ensure you have appropriate permissions (CAP_NET_ADMIN)")
	fmt.Println("2. Replace 'demo0' with an actual network interface")
	fmt.Println("3. Adjust parameters based on your network requirements")
	fmt.Println("4. Test configurations in a safe environment first")
}