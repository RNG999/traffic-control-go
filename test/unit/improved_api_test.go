package unit

import (
	"strings"
	"testing"

	"github.com/rng999/traffic-control-go/api"
	"github.com/stretchr/testify/assert"
)

func TestImprovedAPI_BasicUsage(t *testing.T) {
	tc := api.NewImproved("eth0").
		TotalBandwidth("1Gbps")

	// Configure a class
	tc.Class("Web Services").
		Guaranteed("300Mbps").
		BurstTo("500Mbps").
		Priority(4).
		Ports(80, 443)

	// Verify configuration
	config := tc.String()
	assert.Contains(t, config, "TrafficController[eth0]")
	assert.Contains(t, config, "Total Bandwidth: 1.0Gbps")
	assert.Contains(t, config, "Web Services")
	assert.Contains(t, config, "guaranteed=300.0Mbps")
	assert.Contains(t, config, "burst=500.0Mbps")
	assert.Contains(t, config, "priority=4")
	assert.Contains(t, config, "ports=[80 443]")
}

func TestImprovedAPI_MultipleClasses(t *testing.T) {
	tc := api.NewImproved("enp0s3").
		TotalBandwidth("10Gbps")

	// Configure multiple classes
	tc.Class("High Priority").
		Guaranteed("4Gbps").
		BurstTo("6Gbps").
		Priority(1).
		Ports(22, 53).
		SourceIPs("192.168.1.0/24")

	tc.Class("Normal Priority").
		Guaranteed("3Gbps").
		BurstTo("5Gbps").
		Priority(4).
		Ports(80, 443).
		DestIPs("10.0.1.100", "10.0.1.101")

	tc.Class("Low Priority").
		Guaranteed("1Gbps").
		BurstTo("2Gbps").
		Priority(7).
		Protocols("ftp", "rsync")

	config := tc.String()

	// Verify all classes are present
	assert.Contains(t, config, "High Priority")
	assert.Contains(t, config, "Normal Priority")
	assert.Contains(t, config, "Low Priority")
	assert.Contains(t, config, "Classes: 3")
}

func TestImprovedAPI_ClassReuse(t *testing.T) {
	tc := api.NewImproved("eth0").
		TotalBandwidth("1Gbps")

	// Create a class
	tc.Class("Web Services").
		Guaranteed("300Mbps").
		Priority(4).
		Ports(80)

	// Reconfigure the same class (should reuse existing)
	tc.Class("Web Services").
		Ports(443, 8080)

	config := tc.String()

	// Should only have one class
	assert.Contains(t, config, "Classes: 1")
	assert.Contains(t, config, "Web Services")

	// Should have all ports (80 from first call, 443, 8080 from second)
	assert.Contains(t, config, "ports=[80 443 8080]")
}

func TestImprovedAPI_Validation(t *testing.T) {
	tc := api.NewImproved("eth0")

	// Test missing total bandwidth
	tc.Class("Test").
		Guaranteed("100Mbps").
		Priority(4)

	err := tc.Apply()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "total bandwidth must be set")

	// Test missing classes
	tc2 := api.NewImproved("eth0").
		TotalBandwidth("1Gbps")

	err = tc2.Apply()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one traffic class must be defined")

	// Test missing guaranteed bandwidth
	tc3 := api.NewImproved("eth0").
		TotalBandwidth("1Gbps")

	tc3.Class("Test").Priority(4)

	err = tc3.Apply()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "guaranteed bandwidth must be set")

	// Test missing priority
	tc4 := api.NewImproved("eth0").
		TotalBandwidth("1Gbps")

	tc4.Class("Test").Guaranteed("100Mbps")

	err = tc4.Apply()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "priority must be set")
}

func TestImprovedAPI_InvalidInputs(t *testing.T) {
	tc := api.NewImproved("eth0")

	// Test invalid bandwidth format
	tc.TotalBandwidth("invalid")
	// Should not crash, but bandwidth should be nil

	tc.Class("Test").
		Guaranteed("invalid").
		BurstTo("also-invalid").
		Priority(10) // Invalid priority (should be 0-7)

	// Configuration should handle invalid inputs gracefully
	config := tc.String()
	assert.Contains(t, config, "TrafficController[eth0]")
}

func TestImprovedAPI_ChainedConfiguration(t *testing.T) {
	// Test that we can chain configurations naturally
	tc := api.NewImproved("eth0").
		TotalBandwidth("1Gbps")

	class1 := tc.Class("Web").
		Guaranteed("400Mbps").
		BurstTo("600Mbps").
		Priority(3).
		Ports(80, 443)

	class2 := tc.Class("Database").
		Guaranteed("300Mbps").
		BurstTo("500Mbps").
		Priority(2).
		Ports(3306, 5432).
		DestIPs("192.168.1.10")

	// Both classes should be configured
	assert.NotNil(t, class1)
	assert.NotNil(t, class2)

	config := tc.String()
	assert.Contains(t, config, "Classes: 2")
	assert.Contains(t, config, "Web:")
	assert.Contains(t, config, "Database:")
}

func TestImprovedAPI_ComplexScenario(t *testing.T) {
	// Test a realistic complex scenario
	tc := api.NewImproved("bond0").
		TotalBandwidth("40Gbps")

	// Web tier
	tc.Class("Web Tier").
		Guaranteed("15Gbps").
		BurstTo("25Gbps").
		Priority(2).
		Ports(80, 443).
		SourceIPs("10.0.1.0/24")

	// Database tier
	tc.Class("DB Tier").
		Guaranteed("8Gbps").
		BurstTo("15Gbps").
		Priority(0).
		Ports(3306, 5432, 27017).
		SourceIPs("10.0.4.0/24")

	// Management
	tc.Class("Management").
		Guaranteed("2Gbps").
		BurstTo("5Gbps").
		Priority(3).
		Ports(22, 161).
		Protocols("ssh", "snmp")

	// Verify complete configuration
	err := tc.Apply()
	assert.NoError(t, err)

	config := tc.String()
	assert.Contains(t, config, "Classes: 3")
	assert.Contains(t, config, "Total Bandwidth: 40.0Gbps")

	// Verify each class has required settings
	lines := strings.Split(config, "\n")
	webFound := false
	dbFound := false
	mgmtFound := false

	for _, line := range lines {
		if strings.Contains(line, "Web Tier:") {
			assert.Contains(t, line, "guaranteed=15.0Gbps")
			assert.Contains(t, line, "priority=2")
			webFound = true
		}
		if strings.Contains(line, "DB Tier:") {
			assert.Contains(t, line, "guaranteed=8.0Gbps")
			assert.Contains(t, line, "priority=0")
			dbFound = true
		}
		if strings.Contains(line, "Management:") {
			assert.Contains(t, line, "guaranteed=2.0Gbps")
			assert.Contains(t, line, "priority=3")
			mgmtFound = true
		}
	}

	assert.True(t, webFound, "Web Tier configuration not found")
	assert.True(t, dbFound, "DB Tier configuration not found")
	assert.True(t, mgmtFound, "Management configuration not found")
}

// Benchmark the new API performance
func BenchmarkImprovedAPI_Configuration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tc := api.NewImproved("eth0").
			TotalBandwidth("1Gbps")

		tc.Class("Web").
			Guaranteed("400Mbps").
			BurstTo("600Mbps").
			Priority(3).
			Ports(80, 443)

		tc.Class("DB").
			Guaranteed("300Mbps").
			BurstTo("500Mbps").
			Priority(2).
			Ports(3306)

		_ = tc.String()
	}
}

func BenchmarkImprovedAPI_Apply(b *testing.B) {
	tc := api.NewImproved("eth0").
		TotalBandwidth("1Gbps")

	tc.Class("Test").
		Guaranteed("500Mbps").
		Priority(4)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tc.Apply()
	}
}
