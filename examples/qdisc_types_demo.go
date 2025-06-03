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
	tc := api.New(deviceName)
	fmt.Println("✓ TrafficControl instance created")
	fmt.Println()

	// Demo 1: Token Bucket Filter (TBF) for simple rate limiting
	fmt.Println("=== Demo 1: Token Bucket Filter (TBF) ===")
	fmt.Println("TBF provides simple rate limiting with token bucket algorithm")

	err := tc.
		TBFQdisc("1:0", "100Mbps").
		WithBuffer(32768).
		WithLimit(10000).
		WithBurst(1024).
		Apply()

	if err != nil {
		log.Printf("Error applying TBF qdisc: %v", err)
	} else {
		fmt.Println("✓ TBF qdisc configured successfully")
		fmt.Println("  - Handle: 1:0")
		fmt.Println("  - Rate: 100Mbps")
		fmt.Println("  - Buffer: 32768 bytes")
		fmt.Println("  - Limit: 10000 packets")
		fmt.Println("  - Burst: 1024 bytes")
	}
	fmt.Println()

	// Demo 2: Priority (PRIO) qdisc for simple priority classes
	fmt.Println("=== Demo 2: Priority (PRIO) Qdisc ===")
	fmt.Println("PRIO provides simple priority-based packet scheduling")

	err = tc.
		PRIOQdisc("2:0", 3).
		WithPriomap([]uint8{1, 2, 2, 2, 1, 2, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1}).
		Apply()

	if err != nil {
		log.Printf("Error applying PRIO qdisc: %v", err)
	} else {
		fmt.Println("✓ PRIO qdisc configured successfully")
		fmt.Println("  - Handle: 2:0")
		fmt.Println("  - Bands: 3 (high, normal, low priority)")
		fmt.Println("  - Priomap: Custom priority mapping")
	}
	fmt.Println()

	// Demo 3: Fair Queue CoDel (FQ_CODEL) for fair queuing with AQM
	fmt.Println("=== Demo 3: Fair Queue CoDel (FQ_CODEL) ===")
	fmt.Println("FQ_CODEL combines fair queuing with CoDel AQM for low latency")

	err = tc.
		FQCODELQdisc("3:0").
		WithLimit(10240).
		WithFlows(1024).
		WithTarget(5000).     // 5ms target delay
		WithInterval(100000). // 100ms interval
		WithQuantum(1518).    // MTU + headers
		WithECN(true).        // Enable ECN marking
		Apply()

	if err != nil {
		log.Printf("Error applying FQ_CODEL qdisc: %v", err)
	} else {
		fmt.Println("✓ FQ_CODEL qdisc configured successfully")
		fmt.Println("  - Handle: 3:0")
		fmt.Println("  - Limit: 10240 packets")
		fmt.Println("  - Flows: 1024 queues")
		fmt.Println("  - Target: 5ms")
		fmt.Println("  - Interval: 100ms")
		fmt.Println("  - Quantum: 1518 bytes")
		fmt.Println("  - ECN: enabled")
	}
	fmt.Println()

	// Demo 4: HTB with hierarchical classes (existing functionality)
	fmt.Println("=== Demo 4: HTB with Hierarchical Classes ===")
	fmt.Println("HTB provides hierarchical token bucket with borrowing")

	err = tc.
		HTBQdisc("4:0", "4:999").
		HTBClass("4:0", "4:1", "parent", "100Mbps", "100Mbps").
		HTBClass("4:1", "4:10", "high", "60Mbps", "80Mbps").
		HTBClass("4:1", "4:20", "medium", "30Mbps", "50Mbps").
		HTBClass("4:1", "4:30", "low", "10Mbps", "20Mbps").
		Apply()

	if err != nil {
		log.Printf("Error applying HTB hierarchy: %v", err)
	} else {
		fmt.Println("✓ HTB hierarchy configured successfully")
		fmt.Println("  - Root qdisc: 4:0 with default class 4:999")
		fmt.Println("  - Parent class: 4:1 (100Mbps)")
		fmt.Println("  - High priority: 4:10 (60Mbps, ceil 80Mbps)")
		fmt.Println("  - Medium priority: 4:20 (30Mbps, ceil 50Mbps)")
		fmt.Println("  - Low priority: 4:30 (10Mbps, ceil 20Mbps)")
	}
	fmt.Println()

	// Demo 5: Real-world scenarios
	fmt.Println("=== Demo 5: Real-world Use Cases ===")
	fmt.Println()

	fmt.Println("Use Case 1: Home Internet Gateway")
	fmt.Println("- Use TBF for simple bandwidth limiting on WAN interface")
	fmt.Println("- Use PRIO for separating real-time traffic (VoIP, gaming)")
	fmt.Println()

	fmt.Println("Use Case 2: Enterprise Network")
	fmt.Println("- Use HTB for complex hierarchical bandwidth allocation")
	fmt.Println("- Use FQ_CODEL for low-latency internal traffic")
	fmt.Println()

	fmt.Println("Use Case 3: ISP Customer Management")
	fmt.Println("- Use TBF for per-customer rate limiting")
	fmt.Println("- Use HTB for service-level differentiation")
	fmt.Println()

	fmt.Println("Use Case 4: Data Center")
	fmt.Println("- Use FQ_CODEL for container/VM traffic isolation")
	fmt.Println("- Use PRIO for network control traffic prioritization")
	fmt.Println()

	// Show configuration comparison
	fmt.Println("=== Qdisc Type Comparison ===")
	fmt.Println()

	fmt.Println("| Qdisc    | Use Case              | Complexity | Features                    |")
	fmt.Println("|----------|----------------------|------------|------------------------------|")
	fmt.Println("| TBF      | Simple rate limiting | Low        | Token bucket, burst          |")
	fmt.Println("| PRIO     | Priority classes     | Low        | Static priorities            |")
	fmt.Println("| HTB      | Hierarchical shaping | High       | Borrowing, guarantees        |")
	fmt.Println("| FQ_CODEL | Fair queuing + AQM   | Medium     | Low latency, fairness        |")
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

	// Simple rate limiting on WAN interface with TBF
	err := tc.
		TBFQdisc("1:0", wanBandwidth).
		WithBuffer(32768).
		Apply()

	if err != nil {
		return fmt.Errorf("failed to configure WAN rate limiting: %w", err)
	}

	// Priority qdisc for LAN traffic separation
	err = tc.
		PRIOQdisc("2:0", 3).
		Apply()

	if err != nil {
		return fmt.Errorf("failed to configure LAN prioritization: %w", err)
	}

	return nil
}

// ConfigureDataCenter sets up low-latency traffic control for data center
func ConfigureDataCenter(tc *api.TrafficController) error {
	fmt.Println("Configuring data center traffic control...")

	// FQ_CODEL for fair queuing and low latency
	err := tc.
		FQCODELQdisc("1:0").
		WithLimit(20480).    // Higher limit for data center
		WithFlows(2048).     // More flows for many connections
		WithTarget(1000).    // 1ms target for low latency
		WithInterval(50000). // 50ms interval
		WithECN(true).       // Enable ECN for modern stacks
		Apply()

	return err
}

// ConfigureISPCustomer sets up per-customer traffic shaping
func ConfigureISPCustomer(tc *api.TrafficController, customerPlan string) error {
	fmt.Println("Configuring ISP customer traffic shaping...")

	var rate, ceil string

	switch customerPlan {
	case "basic":
		rate, ceil = "10Mbps", "15Mbps"
	case "standard":
		rate, ceil = "50Mbps", "75Mbps"
	case "premium":
		rate, ceil = "100Mbps", "150Mbps"
	default:
		return fmt.Errorf("unknown customer plan: %s", customerPlan)
	}

	// HTB for guaranteed rate with burst capability
	err := tc.
		HTBQdisc("1:0", "1:999").
		HTBClass("1:0", "1:1", "customer", rate, ceil).
		Apply()

	return err
}
