//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"

	"github.com/rng999/traffic-control-go/api"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
)

func main() {
	fmt.Println("=== Traffic Control Go - Additional Qdisc Types Demo ===")
	fmt.Println()

	// This demo shows usage of the new qdisc types: TBF, PRIO, and FQ_CODEL
	// along with the existing HTB qdisc type

	deviceName := "demo0"

	fmt.Printf("1. Creating TrafficControl for device: %s\n", deviceName)
	tc := api.NetworkInterface(deviceName)
	fmt.Println("✓ TrafficControl instance created")
	fmt.Println()

	// Demo 1: Basic Traffic Control Setup
	fmt.Println("=== Demo 1: Basic Traffic Control Setup ===")
	fmt.Println("Setting up basic traffic control with HTB qdisc")

	err := tc.
		WithHardLimitBandwidth("100Mbps").
		CreateTrafficClass("Default").
		WithGuaranteedBandwidth("100Mbps").
		WithPriority(4).
		Done().
		Apply()

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
	err = tc2.
		WithHardLimitBandwidth("1Gbps").
		CreateTrafficClass("High Priority").
		WithGuaranteedBandwidth("300Mbps").
		WithPriority(1).
		ForPort(22).
		Done().
		CreateTrafficClass("Normal Priority").
		WithGuaranteedBandwidth("400Mbps").
		WithPriority(4).
		ForPort(80, 443).
		Done().
		CreateTrafficClass("Low Priority").
		WithGuaranteedBandwidth("300Mbps").
		WithPriority(6).
		ForApplication("torrent").
		Done().
		Apply()

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
	err = tc3.
		WithHardLimitBandwidth("500Mbps").
		CreateTrafficClass("Database").
		WithGuaranteedBandwidth("150Mbps").
		WithSoftLimitBandwidth("250Mbps").
		WithPriority(2).
		ForPort(3306, 5432).
		Done().
		CreateTrafficClass("Web Services").
		WithGuaranteedBandwidth("200Mbps").
		WithSoftLimitBandwidth("350Mbps").
		WithPriority(3).
		ForPort(80, 443, 8080).
		Done().
		Apply()

	if err != nil {
		log.Printf("Error applying bandwidth shaping: %v", err)
	} else {
		fmt.Println("✓ Bandwidth shaping configured successfully")
		fmt.Println("  - Database: 150Mbps guaranteed, 250Mbps ceiling")
		fmt.Println("  - Web Services: 200Mbps guaranteed, 350Mbps ceiling")
		fmt.Println("  - Borrowing: Enabled between classes")
	}
	fmt.Println()

	// Demo 4: IP-based Traffic Classification
	fmt.Println("=== Demo 4: IP-based Traffic Classification ===")
	fmt.Println("Setting up traffic classes based on source/destination IPs")

	tc4 := api.NetworkInterface(deviceName)
	err = tc4.
		WithHardLimitBandwidth("1Gbps").
		CreateTrafficClass("Internal Traffic").
		WithGuaranteedBandwidth("400Mbps").
		WithSoftLimitBandwidth("600Mbps").
		WithPriority(2).
		ForSourceIPs("192.168.0.0/16", "10.0.0.0/8").
		Done().
		CreateTrafficClass("External Traffic").
		WithGuaranteedBandwidth("300Mbps").
		WithSoftLimitBandwidth("500Mbps").
		WithPriority(4).
		Done().
		Apply()

	if err != nil {
		log.Printf("Error applying IP-based classification: %v", err)
	} else {
		fmt.Println("✓ IP-based traffic classification configured successfully")
		fmt.Println("  - Internal Traffic: RFC1918 addresses (higher priority)")
		fmt.Println("  - External Traffic: Public addresses (lower priority)")
		fmt.Println("  - Bandwidth borrowing: Enabled")
	}
	fmt.Println()

	// Demo 5: Real-world scenarios
	fmt.Println("=== Demo 5: Real-world Use Cases ===")
	fmt.Println()

	fmt.Println("Use Case 1: Home Internet Gateway")
	fmt.Println("- Set total bandwidth to match your ISP plan")
	fmt.Println("- Prioritize real-time traffic (VoIP, gaming)")
	fmt.Println()

	fmt.Println("Use Case 2: Enterprise Network")
	fmt.Println("- Use hierarchical traffic classes for departments")
	fmt.Println("- Guarantee minimum bandwidth for critical applications")
	fmt.Println()

	fmt.Println("Use Case 3: ISP Customer Management")
	fmt.Println("- Set per-customer bandwidth limits")
	fmt.Println("- Provide service-level differentiation")
	fmt.Println()

	fmt.Println("Use Case 4: Data Center")
	fmt.Println("- Isolate container/VM traffic with separate classes")
	fmt.Println("- Prioritize network control traffic")
	fmt.Println()

	// Show configuration comparison
	fmt.Println("=== Traffic Control Features ===")
	fmt.Println()

	fmt.Println("| Feature           | Description                      | Example Usage              |")
	fmt.Println("|-------------------|----------------------------------|----------------------------|")
	fmt.Println("| Guaranteed BW     | Minimum bandwidth allocation     | WithGuaranteedBandwidth()  |")
	fmt.Println("| Maximum BW        | Bandwidth ceiling (burst)        | WithSoftLimitBandwidth()         |")
	fmt.Println("| Priority          | Traffic prioritization (0-7)     | WithPriority()             |")
	fmt.Println("| Port Filtering    | Match by TCP/UDP ports           | ForPort()                  |")
	fmt.Println("| IP Filtering      | Match by source/dest IPs         | ForSourceIPs/ForDestIPs    |")
	fmt.Println("| Protocol Match    | Match by protocol name           | ForProtocols()             |")
	fmt.Println()

	fmt.Println("=== Demo completed successfully! ===")
	fmt.Println()
	fmt.Println("Note: This demo shows the API usage. In a real environment:")
	fmt.Println("1. Ensure you have appropriate permissions (CAP_NET_ADMIN)")
	fmt.Println("2. Replace 'demo0' with an actual network interface")
	fmt.Println("3. Adjust parameters based on your network requirements")
	fmt.Println("4. Test configurations in a safe environment first")
}

// Example helper functions for different scenarios

// ConfigureHomeGateway sets up traffic control for a home internet gateway
func ConfigureHomeGateway(tc *api.TrafficController, wanBandwidth string) error {
	fmt.Println("Configuring home gateway traffic control...")

	// Basic bandwidth limiting with prioritized traffic
	err := tc.
		WithHardLimitBandwidth(wanBandwidth).
		CreateTrafficClass("VoIP").
		WithGuaranteedBandwidth("1Mbps").
		WithPriority(0).
		ForPort(5060, 5061).
		Done().
		CreateTrafficClass("Gaming").
		WithGuaranteedBandwidth("5Mbps").
		WithPriority(1).
		ForPort(3000, 4000).
		Done().
		CreateTrafficClass("Web").
		WithGuaranteedBandwidth("10Mbps").
		WithPriority(4).
		ForPort(80, 443).
		Done().
		Apply()

	if err != nil {
		return fmt.Errorf("failed to configure home gateway traffic control: %w", err)
	}

	return nil
}

// ConfigureDataCenter sets up traffic control for data center
func ConfigureDataCenter(tc *api.TrafficController) error {
	fmt.Println("Configuring data center traffic control...")

	// High-performance traffic classes for data center
	err := tc.
		WithHardLimitBandwidth("10Gbps").
		CreateTrafficClass("Database").
		WithGuaranteedBandwidth("3Gbps").
		WithSoftLimitBandwidth("5Gbps").
		WithPriority(1).
		ForPort(3306, 5432, 27017).
		Done().
		CreateTrafficClass("Application").
		WithGuaranteedBandwidth("4Gbps").
		WithSoftLimitBandwidth("7Gbps").
		WithPriority(2).
		ForPort(8080, 8443, 9000).
		Done().
		CreateTrafficClass("Storage").
		WithGuaranteedBandwidth("2Gbps").
		WithSoftLimitBandwidth("4Gbps").
		WithPriority(3).
		ForPort(2049, 445).
		Done().
		Apply()

	return err
}

// ConfigureISPCustomer sets up per-customer traffic shaping
func ConfigureISPCustomer(tc *api.TrafficController, customerPlan string) error {
	fmt.Println("Configuring ISP customer traffic shaping...")

	var totalBW, guaranteedBW, maxBW string

	switch customerPlan {
	case "basic":
		totalBW, guaranteedBW, maxBW = "10Mbps", "8Mbps", "12Mbps"
	case "standard":
		totalBW, guaranteedBW, maxBW = "50Mbps", "40Mbps", "60Mbps"
	case "premium":
		totalBW, guaranteedBW, maxBW = "100Mbps", "80Mbps", "120Mbps"
	default:
		return fmt.Errorf("unknown customer plan: %s", customerPlan)
	}

	// Customer traffic shaping with burst capability
	err := tc.
		WithHardLimitBandwidth(totalBW).
		CreateTrafficClass("Customer").
		WithGuaranteedBandwidth(guaranteedBW).
		WithSoftLimitBandwidth(maxBW).
		WithPriority(4).
		Done().
		Apply()

	return err
}
