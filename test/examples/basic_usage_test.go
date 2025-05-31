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

	controller := api.New("eth0")

	// Test the fluent interface without applying
	err := controller.
		SetTotalBandwidth("100Mbps").
		CreateTrafficClass("Critical Services").
		WithGuaranteedBandwidth("20Mbps").
		WithBurstableTo("40Mbps").
		WithPriority(1). // High priority
		ForPort(22, 443).
		And().
		CreateTrafficClass("Web Traffic").
		WithGuaranteedBandwidth("30Mbps").
		WithBurstableTo("60Mbps").
		WithPriority(4). // Normal priority
		ForPort(80, 443).
		And().
		CreateTrafficClass("Background").
		WithGuaranteedBandwidth("10Mbps").
		WithPriority(6). // Low priority
		ForDestination("192.168.1.100").
		Apply()

	// Should succeed with valid configuration
	assert.NoError(t, err)
}

// TestInvalidConfiguration demonstrates validation
func TestInvalidConfiguration(t *testing.T) {
	controller := api.New("eth0")

	// Test over-allocation of guaranteed bandwidth
	err := controller.
		SetTotalBandwidth("100Mbps").
		CreateTrafficClass("Service1").
		WithGuaranteedBandwidth("60Mbps").
		WithPriority(4).
		And().
		CreateTrafficClass("Service2").
		WithGuaranteedBandwidth("50Mbps").
		WithPriority(4).
		Apply()

	// Should fail due to over-allocation
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "total guaranteed bandwidth")
}

// TestInvalidBandwidthLimits tests bandwidth validation
func TestInvalidBandwidthLimits(t *testing.T) {
	controller := api.New("eth0")

	// Test guaranteed > max bandwidth
	err := controller.
		SetTotalBandwidth("100Mbps").
		CreateTrafficClass("Invalid").
		WithGuaranteedBandwidth("50Mbps").
		WithMaxBandwidth("30Mbps").
		WithPriority(4).
		Apply()

	// Should fail due to guaranteed > max
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "guaranteed bandwidth")
}

// TestMissingTotalBandwidth tests missing bandwidth configuration
func TestMissingTotalBandwidth(t *testing.T) {
	controller := api.New("eth0")

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
