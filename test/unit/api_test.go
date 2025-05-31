package unit

import (
	"testing"
	
	"github.com/rng999/traffic-control-go/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrafficControllerValidation(t *testing.T) {
	t.Run("ValidConfiguration", func(t *testing.T) {
		controller := api.New("eth0")
		
		err := controller.
			SetTotalBandwidth("100Mbps").
			CreateTrafficClass("Test").
				WithGuaranteedBandwidth("50Mbps").
				WithMaxBandwidth("80Mbps").
				WithPriority(4).
				Apply()
		
		assert.NoError(t, err)
	})
	
	t.Run("MissingTotalBandwidth", func(t *testing.T) {
		controller := api.New("eth0")
		
		err := controller.
			CreateTrafficClass("Test").
				WithGuaranteedBandwidth("50Mbps").
				WithPriority(4).
				Apply()
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "total bandwidth not set")
	})
	
	t.Run("OverAllocatedGuaranteed", func(t *testing.T) {
		controller := api.New("eth0")
		
		err := controller.
			SetTotalBandwidth("100Mbps").
			CreateTrafficClass("Class1").
				WithGuaranteedBandwidth("60Mbps").
				WithPriority(4).
				And().
			CreateTrafficClass("Class2").
				WithGuaranteedBandwidth("50Mbps").
				WithPriority(4).
				Apply()
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "total guaranteed bandwidth")
	})
	
	t.Run("MaxExceedsTotal", func(t *testing.T) {
		controller := api.New("eth0")
		
		err := controller.
			SetTotalBandwidth("100Mbps").
			CreateTrafficClass("Test").
				WithGuaranteedBandwidth("50Mbps").
				WithMaxBandwidth("150Mbps").
				WithPriority(4).
				Apply()
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max bandwidth")
		assert.Contains(t, err.Error(), "higher than total bandwidth")
	})
	
	t.Run("GuaranteedExceedsMax", func(t *testing.T) {
		controller := api.New("eth0")
		
		err := controller.
			SetTotalBandwidth("100Mbps").
			CreateTrafficClass("Test").
				WithGuaranteedBandwidth("60Mbps").
				WithMaxBandwidth("50Mbps").
				WithPriority(4).
				Apply()
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "guaranteed bandwidth")
		assert.Contains(t, err.Error(), "higher than max bandwidth")
	})
}

func TestBuilderPatterns(t *testing.T) {
	t.Run("TrafficClassBuilder", func(t *testing.T) {
		controller := api.New("eth0")
		
		// Test fluent interface
		controller.
			SetTotalBandwidth("100Mbps").
			CreateTrafficClass("Web").
				WithGuaranteedBandwidth("30Mbps").
				WithBurstableTo("60Mbps").
				WithPriority(1). // High priority
				ForPort(80, 443).
				ForDestination("192.168.1.10").
				And().
			CreateTrafficClass("SSH").
				WithGuaranteedBandwidth("5Mbps").
				WithPriority(6). // Low priority
				ForPort(22, 2222)
		
		// Should not panic and should be chainable
		require.NotNil(t, controller)
	})
	
	t.Run("PriorityValues", func(t *testing.T) {
		controller := api.New("eth0")
		
		// Test priority setting
		controller.
			SetTotalBandwidth("1Gbps").
			CreateTrafficClass("Critical").
				WithGuaranteedBandwidth("100Mbps").
				WithPriority(0). // Highest priority
				And().
			CreateTrafficClass("Normal").
				WithGuaranteedBandwidth("100Mbps").
				WithPriority(4). // Must set explicit priority
				And().
			CreateTrafficClass("Background").
				WithGuaranteedBandwidth("100Mbps").
				WithPriority(7) // Lowest priority
		
		require.NotNil(t, controller)
	})
}

func TestFilterTypes(t *testing.T) {
	controller := api.New("eth0")
	
	// Test all filter types compile correctly
	err := controller.
		SetTotalBandwidth("100Mbps").
		CreateTrafficClass("AllFilters").
			WithGuaranteedBandwidth("10Mbps").
			WithPriority(4).
			ForDestination("192.168.1.1").
			ForSource("10.0.0.1").
			ForPort(80, 443, 8080).
			ForApplication("http", "ssh").
			Apply()
	
	assert.NoError(t, err)
}