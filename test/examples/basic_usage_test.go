package examples

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rng999/traffic-control-go/api"
)

// TestBasicTrafficControlUsage demonstrates the human-readable API
func TestBasicTrafficControlUsage(t *testing.T) {
	// This test demonstrates the fluent API design
	// It's not actually applying TC rules but validates the API structure

	controller := api.NetworkInterface("eth0")

	// Test the fluent interface without applying
	controller.WithHardLimitBandwidth("100Mbps")

	controller.
		CreateTrafficClass("Critical Services").
		WithGuaranteedBandwidth("20Mbps").
		WithSoftLimitBandwidth("40Mbps").
		WithPriority(1). // High priority
		ForPort(22, 443).
		AddClass()

	controller.
		CreateTrafficClass("Web Traffic").
		WithGuaranteedBandwidth("30Mbps").
		WithSoftLimitBandwidth("60Mbps").
		WithPriority(4). // Normal priority
		ForPort(80, 443).
		AddClass()

	controller.
		CreateTrafficClass("Background").
		WithGuaranteedBandwidth("10Mbps").
		WithPriority(6). // Low priority
		ForDestination("192.168.1.100").
		AddClass()

	err := controller.Apply()

	// Should succeed with valid configuration
	assert.NoError(t, err)
}

// TestInvalidConfiguration demonstrates validation
func TestInvalidConfiguration(t *testing.T) {
	controller := api.NetworkInterface("eth0")

	// Test over-allocation of guaranteed bandwidth
	controller.WithHardLimitBandwidth("100Mbps")

	controller.
		CreateTrafficClass("Service1").
		WithGuaranteedBandwidth("60Mbps").
		WithPriority(4).
		AddClass()

	controller.
		CreateTrafficClass("Service2").
		WithGuaranteedBandwidth("50Mbps").
		WithPriority(4).
		AddClass()

	err := controller.Apply()

	// Should fail due to over-allocation
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "total guaranteed bandwidth")
}

// TestInvalidBandwidthLimits tests bandwidth validation
func TestInvalidBandwidthLimits(t *testing.T) {
	controller := api.NetworkInterface("eth0")

	// Test guaranteed > max bandwidth
	err := controller.
		WithHardLimitBandwidth("100Mbps").
		CreateTrafficClass("Invalid").
		WithGuaranteedBandwidth("50Mbps").
		WithSoftLimitBandwidth("30Mbps").
		WithPriority(4).
		Apply()

	// Should fail due to guaranteed > max
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "guaranteed bandwidth")
}

// TestMissingTotalBandwidth tests missing bandwidth configuration
func TestMissingTotalBandwidth(t *testing.T) {
	controller := api.NetworkInterface("eth0")

	// Test missing total bandwidth
	err := controller.
		CreateTrafficClass("Service").
		WithGuaranteedBandwidth("10Mbps").
		WithPriority(4).
		Apply()

	// Should fail due to missing total bandwidth
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "total bandwidth not set")
}
