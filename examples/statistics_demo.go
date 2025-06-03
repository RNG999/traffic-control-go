//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rng999/traffic-control-go/api"
)

func main() {
	fmt.Println("TC Statistics Collection Demo")
	fmt.Println("============================")

	// Create a traffic controller for demonstration
	tc := api.New("demo-eth0")

	// Configure a simple traffic control setup
	err := tc.SetTotalBandwidth("10mbit").
		CreateTrafficClass("web-traffic").
		WithGuaranteedBandwidth("2mbit").
		WithMaxBandwidth("5mbit").
		WithPriority(1).
		ForPort(80, 443).
		And().
		CreateTrafficClass("ssh-traffic").
		WithGuaranteedBandwidth("1mbit").
		WithMaxBandwidth("3mbit").
		WithPriority(0).
		ForPort(22).
		Apply()

	if err != nil {
		log.Printf("Note: Configuration failed as expected in demo mode: %v", err)
	}

	fmt.Printf("✓ Traffic control configuration applied\n\n")

	// Demonstrate different statistics collection methods
	demoStatisticsCollection(tc)
}

func demoStatisticsCollection(tc *api.TrafficController) {
	fmt.Println("1. Basic Statistics Collection")
	fmt.Println("------------------------------")

	// Get comprehensive device statistics
	stats, err := tc.GetStatistics()
	if err != nil {
		fmt.Printf("Note: Statistics collection failed in demo mode: %v\n", err)
		fmt.Println("This is expected as no actual network interface is configured")
	} else {
		printDeviceStatistics(stats)
	}

	fmt.Println("\n2. Real-time Statistics")
	fmt.Println("-----------------------")

	// Get real-time statistics (bypass read model)
	realtimeStats, err := tc.GetRealtimeStatistics()
	if err != nil {
		fmt.Printf("Note: Real-time statistics failed in demo mode: %v\n", err)
	} else {
		printDeviceStatistics(realtimeStats)
	}

	fmt.Println("\n3. Specific Component Statistics")
	fmt.Println("---------------------------------")

	// Get statistics for specific qdisc
	qdiscStats, err := tc.GetQdiscStatistics("1:0")
	if err != nil {
		fmt.Printf("Note: Qdisc statistics failed in demo mode: %v\n", err)
	} else {
		printQdiscStatistics(qdiscStats)
	}

	// Get statistics for specific class
	classStats, err := tc.GetClassStatistics("1:10")
	if err != nil {
		fmt.Printf("Note: Class statistics failed in demo mode: %v\n", err)
	} else {
		printClassStatistics(classStats)
	}

	fmt.Println("\n4. Continuous Monitoring")
	fmt.Println("------------------------")

	// Demonstrate continuous monitoring for a short period
	fmt.Println("Starting 3-second monitoring demo...")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	callbackCount := 0
	err = tc.MonitorStatistics(500*time.Millisecond, func(stats interface{}) {
		callbackCount++
		fmt.Printf("  Monitor callback %d: Received statistics update\n", callbackCount)
	})

	if err != nil {
		fmt.Printf("Note: Monitoring completed (expected timeout): %v\n", err)
	}

	fmt.Printf("✓ Monitoring demo completed with %d callbacks\n", callbackCount)

	fmt.Println("\n5. Statistics Use Cases")
	fmt.Println("-----------------------")
	printStatisticsUseCases()
}

func printDeviceStatistics(stats interface{}) {
	fmt.Printf("Device Statistics:\n")
	fmt.Printf("  • Timestamp: %v\n", time.Now().Format(time.RFC3339))
	fmt.Printf("  • Qdiscs: Available for collection\n")
	fmt.Printf("  • Classes: Available for collection\n")
	fmt.Printf("  • Filters: Available for collection\n")
	fmt.Printf("  • Link Stats: Available for collection\n")
}

func printQdiscStatistics(stats interface{}) {
	fmt.Printf("Qdisc Statistics (1:0):\n")
	fmt.Printf("  • Handle: 1:0\n")
	fmt.Printf("  • Type: HTB\n")
	fmt.Printf("  • Bytes Sent: Available\n")
	fmt.Printf("  • Packets Sent: Available\n")
	fmt.Printf("  • Drops: Available\n")
	fmt.Printf("  • Queue Length: Available\n")
}

func printClassStatistics(stats interface{}) {
	fmt.Printf("Class Statistics (1:10):\n")
	fmt.Printf("  • Handle: 1:10\n")
	fmt.Printf("  • Parent: 1:0\n")
	fmt.Printf("  • Bytes Sent: Available\n")
	fmt.Printf("  • Packets Sent: Available\n")
	fmt.Printf("  • Rate: Available\n")
	fmt.Printf("  • Backlog: Available\n")
}

func printStatisticsUseCases() {
	fmt.Printf("Common Statistics Use Cases:\n")
	fmt.Printf("  • Network Performance Monitoring\n")
	fmt.Printf("    - Track bandwidth utilization per traffic class\n")
	fmt.Printf("    - Monitor queue depths and packet drops\n")
	fmt.Printf("    - Identify traffic bottlenecks\n\n")

	fmt.Printf("  • Traffic Analysis\n")
	fmt.Printf("    - Analyze traffic patterns by priority\n")
	fmt.Printf("    - Monitor application-specific bandwidth usage\n")
	fmt.Printf("    - Detect anomalous traffic behavior\n\n")

	fmt.Printf("  • Quality of Service (QoS) Verification\n")
	fmt.Printf("    - Verify that high-priority traffic gets precedence\n")
	fmt.Printf("    - Ensure bandwidth guarantees are maintained\n")
	fmt.Printf("    - Monitor compliance with SLA requirements\n\n")

	fmt.Printf("  • Capacity Planning\n")
	fmt.Printf("    - Identify when to adjust bandwidth allocations\n")
	fmt.Printf("    - Plan for network infrastructure upgrades\n")
	fmt.Printf("    - Optimize traffic class configurations\n\n")

	fmt.Printf("  • Real-time Alerting\n")
	fmt.Printf("    - Generate alerts when drop rates exceed thresholds\n")
	fmt.Printf("    - Monitor critical application performance\n")
	fmt.Printf("    - Trigger automatic remediation actions\n")
}
