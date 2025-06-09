//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"

	"github.com/rng999/traffic-control-go/api"
)

func main() {
	fmt.Println("HTB Advanced Configuration Demo")
	fmt.Println("==============================")

	// Example 1: Enterprise network setup
	fmt.Println("\n1. Enterprise Network Setup:")
	enterpriseExample()

	// Example 2: ISP customer management
	fmt.Println("\n2. ISP Customer Management:")
	ispExample()

	// Example 3: Server farm traffic management
	fmt.Println("\n3. Server Farm Traffic Management:")
	serverFarmExample()
}

func enterpriseExample() {
	tc := api.NetworkInterface("eth0")

	// Total enterprise bandwidth: 10 Gbps
	tc.WithHardLimitBandwidth("10Gbps")

	// Executive/Management - highest priority
	tc.CreateTrafficClass("Executive").
		WithGuaranteedBandwidth("2Gbps").
		WithSoftLimitBandwidth("4Gbps").
		WithPriority(0).
		ForSourceIPs("192.168.10.0/24"). // Executive network
		ForPort(80, 443, 22, 25, 993, 995) // Web, SSH, Email

	// Development team - high priority for productivity
	tc.CreateTrafficClass("Development").
		WithGuaranteedBandwidth("3Gbps").
		WithSoftLimitBandwidth("6Gbps").
		WithPriority(1).
		ForSourceIPs("192.168.20.0/24", "192.168.21.0/24"). // Dev networks
		ForPort(22, 80, 443, 8080, 9000, 3000) // SSH, Web, Dev servers

	// General staff - normal priority
	tc.CreateTrafficClass("General Staff").
		WithGuaranteedBandwidth("2Gbps").
		WithSoftLimitBandwidth("4Gbps").
		WithPriority(3).
		ForSourceIPs("192.168.30.0/24", "192.168.31.0/24"). // Staff networks
		ForPort(80, 443, 25, 110, 143, 993, 995) // Web, Email

	// Guest network - limited bandwidth
	tc.CreateTrafficClass("Guest Network").
		WithGuaranteedBandwidth("500Mbps").
		WithSoftLimitBandwidth("1Gbps").
		WithPriority(6).
		ForSourceIPs("192.168.100.0/24"). // Guest network
		ForPort(80, 443) // Web only

	// Background services - system maintenance
	tc.CreateTrafficClass("System Services").
		WithGuaranteedBandwidth("1Gbps").
		WithSoftLimitBandwidth("2Gbps").
		WithPriority(5).
		ForPort(53, 123, 161, 162, 514, 389, 636) // DNS, NTP, SNMP, Syslog, LDAP

	if err := tc.Apply(); err != nil {
		log.Printf("Failed to apply enterprise configuration: %v", err)
		return
	}

	fmt.Printf("✅ Enterprise configuration applied (10 Gbps total)\n")
	fmt.Printf("   - Executive: 2-4 Gbps (Priority 0)\n")
	fmt.Printf("   - Development: 3-6 Gbps (Priority 1)\n")
	fmt.Printf("   - General Staff: 2-4 Gbps (Priority 3)\n")
	fmt.Printf("   - Guest Network: 0.5-1 Gbps (Priority 6)\n")
	fmt.Printf("   - System Services: 1-2 Gbps (Priority 5)\n")
}

func ispExample() {
	tc := api.NetworkInterface("eth1")

	// ISP uplink: 1 Gbps
	tc.WithHardLimitBandwidth("1Gbps")

	// Premium customers - guaranteed high bandwidth
	tc.CreateTrafficClass("Premium Customers").
		WithGuaranteedBandwidth("400Mbps").
		WithSoftLimitBandwidth("600Mbps").
		WithPriority(0).
		ForSourceIPs("10.1.0.0/16"). // Premium customer range
		ForDestinationIPs("10.1.0.0/16")

	// Business customers - reliable service
	tc.CreateTrafficClass("Business Customers").
		WithGuaranteedBandwidth("300Mbps").
		WithSoftLimitBandwidth("500Mbps").
		WithPriority(1).
		ForSourceIPs("10.2.0.0/16"). // Business customer range
		ForDestinationIPs("10.2.0.0/16")

	// Residential customers - best effort
	tc.CreateTrafficClass("Residential").
		WithGuaranteedBandwidth("200Mbps").
		WithSoftLimitBandwidth("400Mbps").
		WithPriority(3).
		ForSourceIPs("10.3.0.0/16"). // Residential range
		ForDestinationIPs("10.3.0.0/16")

	// Bulk/P2P traffic - lowest priority
	tc.CreateTrafficClass("Bulk Traffic").
		WithGuaranteedBandwidth("50Mbps").
		WithSoftLimitBandwidth("200Mbps").
		WithPriority(7).
		ForPort(6881, 6882, 6883, 6884, 6885, 6886, 6887, 6888, 6889) // BitTorrent

	if err := tc.Apply(); err != nil {
		log.Printf("Failed to apply ISP configuration: %v", err)
		return
	}

	fmt.Printf("✅ ISP configuration applied (1 Gbps total)\n")
	fmt.Printf("   - Premium: 400-600 Mbps (Priority 0)\n")
	fmt.Printf("   - Business: 300-500 Mbps (Priority 1)\n")
	fmt.Printf("   - Residential: 200-400 Mbps (Priority 3)\n")
	fmt.Printf("   - Bulk/P2P: 50-200 Mbps (Priority 7)\n")
}

func serverFarmExample() {
	tc := api.NetworkInterface("eth2")

	// Server farm uplink: 40 Gbps
	tc.WithHardLimitBandwidth("40Gbps")

	// Database tier - mission critical
	tc.CreateTrafficClass("Database Tier").
		WithGuaranteedBandwidth("15Gbps").
		WithSoftLimitBandwidth("25Gbps").
		WithPriority(0).
		ForSourceIPs("10.0.10.0/24"). // DB server subnet
		ForDestinationIPs("10.0.10.0/24").
		ForPort(3306, 5432, 1521, 1433, 27017, 6379, 11211) // Various databases

	// Application tier - high priority
	tc.CreateTrafficClass("Application Tier").
		WithGuaranteedBandwidth("12Gbps").
		WithSoftLimitBandwidth("20Gbps").
		WithPriority(1).
		ForSourceIPs("10.0.20.0/24", "10.0.21.0/24"). // App server subnets
		ForDestinationIPs("10.0.20.0/24", "10.0.21.0/24").
		ForPort(8080, 8443, 9000, 9443, 8000, 8888) // App server ports

	// Web tier - public facing
	tc.CreateTrafficClass("Web Tier").
		WithGuaranteedBandwidth("8Gbps").
		WithSoftLimitBandwidth("15Gbps").
		WithPriority(2).
		ForSourceIPs("10.0.30.0/24"). // Web server subnet
		ForDestinationIPs("10.0.30.0/24").
		ForPort(80, 443, 8080, 8443) // HTTP/HTTPS

	// Cache tier - performance optimization
	tc.CreateTrafficClass("Cache Tier").
		WithGuaranteedBandwidth("3Gbps").
		WithSoftLimitBandwidth("8Gbps").
		WithPriority(2).
		ForSourceIPs("10.0.40.0/24"). // Cache server subnet
		ForDestinationIPs("10.0.40.0/24").
		ForPort(6379, 11211, 8091, 8092) // Redis, Memcached, Couchbase

	// Storage/Backup - background operations
	tc.CreateTrafficClass("Storage").
		WithGuaranteedBandwidth("2Gbps").
		WithSoftLimitBandwidth("5Gbps").
		WithPriority(5).
		ForSourceIPs("10.0.50.0/24"). // Storage subnet
		ForDestinationIPs("10.0.50.0/24").
		ForPort(2049, 111, 445, 139, 873) // NFS, SMB, Rsync

	if err := tc.Apply(); err != nil {
		log.Printf("Failed to apply server farm configuration: %v", err)
		return
	}

	fmt.Printf("✅ Server farm configuration applied (40 Gbps total)\n")
	fmt.Printf("   - Database: 15-25 Gbps (Priority 0)\n")
	fmt.Printf("   - Application: 12-20 Gbps (Priority 1)\n")
	fmt.Printf("   - Web: 8-15 Gbps (Priority 2)\n")
	fmt.Printf("   - Cache: 3-8 Gbps (Priority 2)\n")
	fmt.Printf("   - Storage: 2-5 Gbps (Priority 5)\n")
}