//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"

	"github.com/rng999/traffic-control-go/api"
)

func main() {
	fmt.Println("Improved Traffic Control API Demo")
	fmt.Println("=================================")

	// Example 1: Clean, readable API without redundant And() calls
	fmt.Println("\n1. Improved API Example:")
	improvedExample()

	// Example 2: Complex multi-class configuration
	fmt.Println("\n2. Complex Multi-Class Configuration:")
	complexExample()

	// Example 3: Server infrastructure example
	fmt.Println("\n3. Server Infrastructure Example:")
	serverInfrastructureExample()
}

func improvedExample() {
	// New improved API - much cleaner!
	tc := api.NewImproved("eth0").
		TotalBandwidth("1Gbps")

	// Configure web services class
	tc.Class("Web Services").
		Guaranteed("300Mbps").
		BurstTo("500Mbps").
		Priority(4).
		Ports(80, 443, 8080, 8443)

	// Configure database class
	tc.Class("Database").
		Guaranteed("200Mbps").
		BurstTo("400Mbps").
		Priority(2).
		Ports(3306, 5432, 1521).
		DestIPs("192.168.1.10", "192.168.1.11")

	// Configure management class
	tc.Class("Management").
		Guaranteed("50Mbps").
		BurstTo("100Mbps").
		Priority(1).
		Ports(22, 3389, 5900).
		Protocols("ssh", "rdp", "vnc")

	fmt.Printf("Configuration:\n%s\n", tc.String())

	// Apply configuration
	if err := tc.Apply(); err != nil {
		log.Printf("Error applying configuration: %v", err)
	} else {
		fmt.Println("✓ Configuration applied successfully!")
	}
}

func complexExample() {
	tc := api.NewImproved("enp0s3").
		TotalBandwidth("10Gbps")

	// High priority - Critical services
	tc.Class("Critical Services").
		Guaranteed("2Gbps").
		BurstTo("4Gbps").
		Priority(0).
		Ports(53, 123, 389, 636). // DNS, NTP, LDAP
		Protocols("dns", "ntp", "ldap")

	// Normal priority - Application services
	tc.Class("Applications").
		Guaranteed("4Gbps").
		BurstTo("6Gbps").
		Priority(3).
		Ports(80, 443, 8080, 8443, 9090, 9091)

	// Low priority - File transfers
	tc.Class("File Transfers").
		Guaranteed("1Gbps").
		BurstTo("3Gbps").
		Priority(6).
		Ports(21, 22, 873, 445). // FTP, SCP, rsync, SMB
		Protocols("ftp", "scp", "rsync", "smb")

	// Background - Monitoring and backups
	tc.Class("Background").
		Guaranteed("500Mbps").
		BurstTo("1Gbps").
		Priority(7).
		SourceIPs("192.168.100.0/24"). // Monitoring subnet
		Protocols("snmp", "syslog")

	fmt.Printf("Complex Configuration:\n%s\n", tc.String())

	if err := tc.Apply(); err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Println("✓ Complex configuration applied!")
	}
}

func serverInfrastructureExample() {
	// Server infrastructure with role-based traffic shaping
	tc := api.NewImproved("bond0").
		TotalBandwidth("40Gbps")

	// Web tier
	tc.Class("Web Tier").
		Guaranteed("15Gbps").
		BurstTo("25Gbps").
		Priority(2).
		Ports(80, 443).
		SourceIPs("10.0.1.0/24") // Web server subnet

	// Application tier
	tc.Class("App Tier").
		Guaranteed("10Gbps").
		BurstTo("20Gbps").
		Priority(1).
		Ports(8080, 8443, 9000, 9443).
		SourceIPs("10.0.2.0/24", "10.0.3.0/24") // App server subnets

	// Database tier
	tc.Class("DB Tier").
		Guaranteed("8Gbps").
		BurstTo("15Gbps").
		Priority(0). // Highest priority for database
		Ports(3306, 5432, 27017, 6379). // MySQL, PostgreSQL, MongoDB, Redis
		SourceIPs("10.0.4.0/24")

	// Storage and backup
	tc.Class("Storage").
		Guaranteed("5Gbps").
		BurstTo("10Gbps").
		Priority(5).
		Ports(2049, 111, 445, 139). // NFS, SMB
		Protocols("nfs", "smb", "iscsi")

	// Management and monitoring
	tc.Class("Management").
		Guaranteed("2Gbps").
		BurstTo("5Gbps").
		Priority(3).
		Ports(22, 161, 162, 514, 5601). // SSH, SNMP, Syslog, Kibana
		SourceIPs("10.0.100.0/24") // Management subnet

	fmt.Printf("Server Infrastructure Configuration:\n%s\n", tc.String())

	if err := tc.Apply(); err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Println("✓ Server infrastructure configured!")
	}
}

// Comparison function showing old vs new API
func comparisonExample() {
	fmt.Println("API Comparison:")
	fmt.Println("===============")

	fmt.Println("\nOLD API (with redundant And() calls):")
	fmt.Println(`controller := api.New("eth0").
    SetTotalBandwidth("100Mbps").
    CreateTrafficClass("Web Services").
        WithGuaranteedBandwidth("30Mbps").
        WithBurstableTo("60Mbps").
        WithPriority(4).
        ForPort(80, 443).
    And().  // ← Redundant!
    CreateTrafficClass("SSH Management").
        WithGuaranteedBandwidth("5Mbps").
        WithBurstableTo("10Mbps").
        WithPriority(1).
        ForPort(22).
    And().  // ← Redundant!
    Apply()`)

	fmt.Println("\nNEW API (clean and natural):")
	fmt.Println(`tc := api.NewImproved("eth0").
    TotalBandwidth("100Mbps")

tc.Class("Web Services").
    Guaranteed("30Mbps").
    BurstTo("60Mbps").
    Priority(4).
    Ports(80, 443)

tc.Class("SSH Management").
    Guaranteed("5Mbps").
    BurstTo("10Mbps").
    Priority(1).
    Ports(22)

tc.Apply()`)

	fmt.Println("\n✓ Much cleaner and more readable!")
}