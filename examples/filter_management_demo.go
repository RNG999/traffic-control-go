//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"

	"github.com/rng999/traffic-control-go/api"
)

func main() {
	fmt.Println("Traffic Control Filter Management Demo")
	fmt.Println("====================================")

	// Example 1: Basic filter setup
	fmt.Println("\n1. Basic Filter Setup:")
	basicFilterExample()

	// Example 2: Complex multi-filter configuration
	fmt.Println("\n2. Complex Multi-Filter Configuration:")
	complexFilterExample()

	// Example 3: Protocol-based filtering
	fmt.Println("\n3. Protocol-Based Filtering:")
	protocolFilterExample()
}

func basicFilterExample() {
	// Create traffic controller for eth0
	tc := api.NetworkInterface("eth0")

	// Set total bandwidth limit
	tc.WithHardLimitBandwidth("100Mbps")

	// Create high priority class for web traffic
	tc.CreateTrafficClass("Web Traffic").
		WithGuaranteedBandwidth("30Mbps").
		WithSoftLimitBandwidth("60Mbps").
		WithPriority(1).
		ForPort(80, 443)

	// Create medium priority class for SSH
	tc.CreateTrafficClass("SSH Traffic").
		WithGuaranteedBandwidth("10Mbps").
		WithSoftLimitBandwidth("20Mbps").
		WithPriority(2).
		ForPort(22)

	// Create low priority class for bulk transfers
	tc.CreateTrafficClass("Bulk Transfer").
		WithGuaranteedBandwidth("20Mbps").
		WithSoftLimitBandwidth("40Mbps").
		WithPriority(5).
		ForPort(21, 8080)

	// Apply configuration
	if err := tc.Apply(); err != nil {
		log.Printf("Failed to apply configuration: %v", err)
		return
	}

	fmt.Printf("✅ Basic filter configuration applied successfully\n")
	fmt.Printf("   - Web traffic: ports 80,443 → Priority 1 (30-60 Mbps)\n")
	fmt.Printf("   - SSH traffic: port 22 → Priority 2 (10-20 Mbps)\n")
	fmt.Printf("   - Bulk transfer: ports 21,8080 → Priority 5 (20-40 Mbps)\n")
}

func complexFilterExample() {
	tc := api.NetworkInterface("eth0")

	tc.WithHardLimitBandwidth("1Gbps")

	// Critical services - database traffic from specific subnets
	tc.CreateTrafficClass("Database Services").
		WithGuaranteedBandwidth("300Mbps").
		WithSoftLimitBandwidth("500Mbps").
		WithPriority(0).
		ForPort(3306, 5432, 27017). // MySQL, PostgreSQL, MongoDB
		ForSourceIPs("10.0.10.0/24", "10.0.11.0/24")

	// Application servers - web application traffic
	tc.CreateTrafficClass("Application Tier").
		WithGuaranteedBandwidth("400Mbps").
		WithSoftLimitBandwidth("700Mbps").
		WithPriority(1).
		ForPort(8080, 8443, 9000, 9443).
		ForDestinationIPs("10.0.20.0/24")

	// Management traffic - monitoring and admin
	tc.CreateTrafficClass("Management").
		WithGuaranteedBandwidth("50Mbps").
		WithSoftLimitBandwidth("100Mbps").
		WithPriority(2).
		ForPort(22, 161, 162, 514). // SSH, SNMP, Syslog
		ForSourceIPs("10.0.100.0/24")

	// Background services - backups and maintenance
	tc.CreateTrafficClass("Background").
		WithGuaranteedBandwidth("100Mbps").
		WithSoftLimitBandwidth("200Mbps").
		WithPriority(6).
		ForPort(873, 445). // Rsync, SMB
		ForDestinationIPs("10.0.200.0/24")

	if err := tc.Apply(); err != nil {
		log.Printf("Failed to apply complex configuration: %v", err)
		return
	}

	fmt.Printf("✅ Complex filter configuration applied successfully\n")
	fmt.Printf("   - Database: ports 3306,5432,27017 from 10.0.10-11.0/24 → Priority 0\n")
	fmt.Printf("   - App tier: ports 8080,8443,9000,9443 to 10.0.20.0/24 → Priority 1\n")
	fmt.Printf("   - Management: ports 22,161,162,514 from 10.0.100.0/24 → Priority 2\n")
	fmt.Printf("   - Background: ports 873,445 to 10.0.200.0/24 → Priority 6\n")
}

func protocolFilterExample() {
	tc := api.NetworkInterface("eth0")

	tc.WithHardLimitBandwidth("500Mbps")

	// Real-time communication - highest priority
	tc.CreateTrafficClass("Real-time").
		WithGuaranteedBandwidth("150Mbps").
		WithSoftLimitBandwidth("250Mbps").
		WithPriority(0).
		ForProtocols("rtp", "rtcp"). // Real-time protocols
		ForPort(5060, 5061) // SIP

	// Interactive traffic - high priority
	tc.CreateTrafficClass("Interactive").
		WithGuaranteedBandwidth("100Mbps").
		WithSoftLimitBandwidth("200Mbps").
		WithPriority(1).
		ForProtocols("ssh", "telnet", "rdp").
		ForPort(22, 23, 3389)

	// Bulk data - lower priority
	tc.CreateTrafficClass("Bulk Data").
		WithGuaranteedBandwidth("150Mbps").
		WithSoftLimitBandwidth("300Mbps").
		WithPriority(4).
		ForProtocols("ftp", "http", "https").
		ForPort(21, 80, 443)

	// Best effort - lowest priority
	tc.CreateTrafficClass("Best Effort").
		WithGuaranteedBandwidth("50Mbps").
		WithSoftLimitBandwidth("150Mbps").
		WithPriority(7) // Catch-all for unclassified traffic

	if err := tc.Apply(); err != nil {
		log.Printf("Failed to apply protocol configuration: %v", err)
		return
	}

	fmt.Printf("✅ Protocol-based filter configuration applied successfully\n")
	fmt.Printf("   - Real-time: RTP/RTCP/SIP → Priority 0 (150-250 Mbps)\n")
	fmt.Printf("   - Interactive: SSH/Telnet/RDP → Priority 1 (100-200 Mbps)\n")
	fmt.Printf("   - Bulk Data: FTP/HTTP/HTTPS → Priority 4 (150-300 Mbps)\n")
	fmt.Printf("   - Best Effort: Everything else → Priority 7 (50-150 Mbps)\n")
}