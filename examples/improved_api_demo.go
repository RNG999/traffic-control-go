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
	tc := api.NetworkInterface("eth0").
		WithHardLimitBandwidth("1Gbps")

	// Configure web services class
	tc.CreateTrafficClass("Web Services").
		WithGuaranteedBandwidth("300Mbps").
		WithSoftLimitBandwidth("500Mbps").
		WithPriority(4).
		ForPort(80, 443, 8080, 8443).
		Done()

	// Configure database class
	tc.CreateTrafficClass("Database").
		WithGuaranteedBandwidth("200Mbps").
		WithSoftLimitBandwidth("400Mbps").
		WithPriority(2).
		ForPort(3306, 5432, 1521).
		ForDestinationIPs("192.168.1.10", "192.168.1.11").
		Done()

	// Configure management class
	tc.CreateTrafficClass("Management").
		WithGuaranteedBandwidth("50Mbps").
		WithSoftLimitBandwidth("100Mbps").
		WithPriority(1).
		ForPort(22, 3389, 5900).
		ForProtocols("ssh", "rdp", "vnc").
		Done()

	fmt.Printf("Configuration: Applied successfully\n")

	// Apply configuration
	if err := tc.Apply(); err != nil {
		log.Printf("Error applying configuration: %v", err)
	} else {
		fmt.Println("✓ Configuration applied successfully!")
	}
}

func complexExample() {
	tc := api.NetworkInterface("enp0s3").
		WithHardLimitBandwidth("10Gbps")

	// High priority - Critical services
	tc.CreateTrafficClass("Critical Services").
		WithGuaranteedBandwidth("2Gbps").
		WithSoftLimitBandwidth("4Gbps").
		WithPriority(0).
		ForPort(53, 123, 389, 636). // DNS, NTP, LDAP
		ForProtocols("dns", "ntp", "ldap").
		Done()

	// Normal priority - Application services
	tc.CreateTrafficClass("Applications").
		WithGuaranteedBandwidth("4Gbps").
		WithSoftLimitBandwidth("6Gbps").
		WithPriority(3).
		ForPort(80, 443, 8080, 8443, 9090, 9091).
		Done()

	// Low priority - File transfers
	tc.CreateTrafficClass("File Transfers").
		WithGuaranteedBandwidth("1Gbps").
		WithSoftLimitBandwidth("3Gbps").
		WithPriority(6).
		ForPort(21, 22, 873, 445). // FTP, SCP, rsync, SMB
		ForProtocols("ftp", "scp", "rsync", "smb").
		Done()

	// Background - Monitoring and backups
	tc.CreateTrafficClass("Background").
		WithGuaranteedBandwidth("500Mbps").
		WithSoftLimitBandwidth("1Gbps").
		WithPriority(7).
		ForSourceIPs("192.168.100.0/24"). // Monitoring subnet
		ForProtocols("snmp", "syslog").
		Done()

	fmt.Printf("Complex Configuration: Applied successfully\n")

	if err := tc.Apply(); err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Println("✓ Complex configuration applied!")
	}
}

func serverInfrastructureExample() {
	// Server infrastructure with role-based traffic shaping
	tc := api.NetworkInterface("bond0").
		WithHardLimitBandwidth("40Gbps")

	// Web tier
	tc.CreateTrafficClass("Web Tier").
		WithGuaranteedBandwidth("15Gbps").
		WithSoftLimitBandwidth("25Gbps").
		WithPriority(2).
		ForPort(80, 443).
		ForSourceIPs("10.0.1.0/24"). // Web server subnet
		Done()

	// Application tier
	tc.CreateTrafficClass("App Tier").
		WithGuaranteedBandwidth("10Gbps").
		WithSoftLimitBandwidth("20Gbps").
		WithPriority(1).
		ForPort(8080, 8443, 9000, 9443).
		ForSourceIPs("10.0.2.0/24", "10.0.3.0/24"). // App server subnets
		Done()

	// Database tier
	tc.CreateTrafficClass("DB Tier").
		WithGuaranteedBandwidth("8Gbps").
		WithSoftLimitBandwidth("15Gbps").
		WithPriority(0). // Highest priority for database
		ForPort(3306, 5432, 27017, 6379). // MySQL, PostgreSQL, MongoDB, Redis
		ForSourceIPs("10.0.4.0/24").
		Done()

	// Storage and backup
	tc.CreateTrafficClass("Storage").
		WithGuaranteedBandwidth("5Gbps").
		WithSoftLimitBandwidth("10Gbps").
		WithPriority(5).
		ForPort(2049, 111, 445, 139). // NFS, SMB
		ForProtocols("nfs", "smb", "iscsi").
		Done()

	// Management and monitoring
	tc.CreateTrafficClass("Management").
		WithGuaranteedBandwidth("2Gbps").
		WithSoftLimitBandwidth("5Gbps").
		WithPriority(3).
		ForPort(22, 161, 162, 514, 5601). // SSH, SNMP, Syslog, Kibana
		ForSourceIPs("10.0.100.0/24"). // Management subnet
		Done()

	fmt.Printf("Server Infrastructure Configuration: Applied successfully\n")

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
	fmt.Println(`controller := api.NetworkInterface("eth0").
    WithHardLimitBandwidth("100Mbps").
    CreateTrafficClass("Web Services").
        WithGuaranteedBandwidth("30Mbps").
        WithSoftLimitBandwidth("60Mbps").
        WithPriority(4).
        ForPort(80, 443).
    And().  // ← Redundant!
    CreateTrafficClass("SSH Management").
        WithGuaranteedBandwidth("5Mbps").
        WithSoftLimitBandwidth("10Mbps").
        WithPriority(1).
        ForPort(22).
    And().  // ← Redundant!
    Apply()`)

	fmt.Println("\nNEW API (clean and natural):")
	fmt.Println(`tc := api.NetworkInterface("eth0").
    WithHardLimitBandwidth("100Mbps")

tc.CreateTrafficClass("Web Services").
    WithGuaranteedBandwidth("30Mbps").
    WithSoftLimitBandwidth("60Mbps").
    WithPriority(4).
    ForPort(80, 443).
    Done()

tc.CreateTrafficClass("SSH Management").
    WithGuaranteedBandwidth("5Mbps").
    WithSoftLimitBandwidth("10Mbps").
    WithPriority(1).
    ForPort(22).
    Done()

tc.Apply()`)

	fmt.Println("\n✓ Much cleaner and more readable!")
}
