package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/api"
)

func TestTrafficControllerValidation(t *testing.T) {
	t.Run("ValidConfiguration", func(t *testing.T) {
		controller := api.NetworkInterface("eth0")

		controller.WithHardLimitBandwidth("100Mbps")
		controller.CreateTrafficClass("Test").
			WithGuaranteedBandwidth("50Mbps").
			WithSoftLimitBandwidth("80Mbps").
			WithPriority(4)

		err := controller.Apply()

		assert.NoError(t, err)
	})

	t.Run("MissingTotalBandwidth", func(t *testing.T) {
		controller := api.NetworkInterface("eth0")

		controller.CreateTrafficClass("Test").
			WithGuaranteedBandwidth("50Mbps").
			WithPriority(4)

		err := controller.Apply()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "total bandwidth not set")
	})

	t.Run("OverAllocatedGuaranteed", func(t *testing.T) {
		controller := api.NetworkInterface("eth0")

		controller.WithHardLimitBandwidth("100Mbps")
		controller.CreateTrafficClass("Class1").
			WithGuaranteedBandwidth("60Mbps").
			WithPriority(4)
		controller.CreateTrafficClass("Class2").
			WithGuaranteedBandwidth("50Mbps").
			WithPriority(4)

		err := controller.Apply()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "total guaranteed bandwidth")
	})

	t.Run("MaxExceedsTotal", func(t *testing.T) {
		controller := api.NetworkInterface("eth0")

		controller.WithHardLimitBandwidth("100Mbps")
		controller.CreateTrafficClass("Test").
			WithGuaranteedBandwidth("50Mbps").
			WithSoftLimitBandwidth("150Mbps").
			WithPriority(4)

		err := controller.Apply()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max bandwidth")
		assert.Contains(t, err.Error(), "higher than total bandwidth")
	})

	t.Run("GuaranteedExceedsMax", func(t *testing.T) {
		controller := api.NetworkInterface("eth0")

		controller.WithHardLimitBandwidth("100Mbps")
		controller.CreateTrafficClass("Test").
			WithGuaranteedBandwidth("60Mbps").
			WithSoftLimitBandwidth("50Mbps").
			WithPriority(4)

		err := controller.Apply()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "guaranteed bandwidth")
		assert.Contains(t, err.Error(), "higher than max bandwidth")
	})
}

func TestBuilderPatterns(t *testing.T) {
	t.Run("TrafficClassBuilder", func(t *testing.T) {
		controller := api.NetworkInterface("eth0")

		// Test fluent interface
		controller.WithHardLimitBandwidth("100Mbps")
		controller.CreateTrafficClass("Web").
			WithGuaranteedBandwidth("30Mbps").
			WithSoftLimitBandwidth("60Mbps").
			WithPriority(1). // High priority
			ForPort(80, 443).
			ForDestination("192.168.1.10")
		controller.CreateTrafficClass("SSH").
			WithGuaranteedBandwidth("5Mbps").
			WithPriority(6). // Low priority
			ForPort(22, 2222)

		// Should not panic and should be chainable
		require.NotNil(t, controller)
	})

	t.Run("PriorityValues", func(t *testing.T) {
		controller := api.NetworkInterface("eth0")

		// Test priority setting
		controller.WithHardLimitBandwidth("1Gbps")
		controller.CreateTrafficClass("Critical").
			WithGuaranteedBandwidth("100Mbps").
			WithPriority(0) // Highest priority
		controller.CreateTrafficClass("Normal").
			WithGuaranteedBandwidth("100Mbps").
			WithPriority(4) // Must set explicit priority
		controller.CreateTrafficClass("Background").
			WithGuaranteedBandwidth("100Mbps").
			WithPriority(7) // Lowest priority

		require.NotNil(t, controller)
	})
}

func TestFilterTypes(t *testing.T) {
	controller := api.NetworkInterface("eth0")

	// Test all filter types compile correctly
	controller.WithHardLimitBandwidth("100Mbps")
	controller.CreateTrafficClass("AllFilters").
		WithGuaranteedBandwidth("10Mbps").
		WithPriority(4).
		ForDestination("192.168.1.1").
		ForSource("10.0.0.1").
		ForPort(80, 443, 8080)

	err := controller.Apply()

	assert.NoError(t, err)
}