package unit

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rng999/traffic-control-go/api"
)

func TestPrioritySettings(t *testing.T) {
	t.Run("RequiredPriority", func(t *testing.T) {
		controller := api.NetworkInterface("eth0")
		controller.WithHardLimitBandwidth("1Gbps")

		// Test that missing priority causes an error
		controller.CreateTrafficClass("no-priority").
			WithGuaranteedBandwidth("100Mbps")
			// No priority set - should fail

		err := controller.Apply()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not have a priority set")
	})

	t.Run("NumericPriorities", func(t *testing.T) {
		controller := api.NetworkInterface("eth0")
		controller.WithHardLimitBandwidth("1Gbps")

		// Test numeric priorities 0-7
		for i := 0; i <= 7; i++ {
			controller.CreateTrafficClass(fmt.Sprintf("priority_%d", i)).
				WithGuaranteedBandwidth("100Mbps").
				WithPriority(i)
		}

		err := controller.Apply()
		assert.NoError(t, err)
	})

	t.Run("PriorityBounds", func(t *testing.T) {
		controller := api.NetworkInterface("eth0")
		controller.WithHardLimitBandwidth("1Gbps")

		// Test that priorities are clamped to 0-7
		controller.CreateTrafficClass("negative").
			WithGuaranteedBandwidth("100Mbps").
			WithPriority(-5) // Should become 0

		controller.CreateTrafficClass("too_high").
			WithGuaranteedBandwidth("100Mbps").
			WithPriority(10) // Should become 7

		err := controller.Apply()
		assert.NoError(t, err)
	})
}

func TestConfigPriorities(t *testing.T) {
	t.Run("RequiredPriorityInConfig", func(t *testing.T) {
		config := &api.TrafficControlConfig{
			Version:   "1.0",
			Device:    "eth0",
			Bandwidth: "1Gbps",
			Classes: []api.TrafficClassConfig{
				{Name: "no-priority", Guaranteed: "100Mbps"},
			},
		}

		controller := api.NetworkInterface("eth0")
		err := controller.ApplyConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not have a priority set")
	})

	t.Run("NumericPrioritiesInConfig", func(t *testing.T) {
		config := &api.TrafficControlConfig{
			Version:   "1.0",
			Device:    "eth0",
			Bandwidth: "1Gbps",
			Classes: []api.TrafficClassConfig{
				{Name: "p0", Guaranteed: "100Mbps", Priority: &[]int{0}[0]},
				{Name: "p3", Guaranteed: "100Mbps", Priority: &[]int{3}[0]},
				{Name: "p7", Guaranteed: "100Mbps", Priority: &[]int{7}[0]},
			},
		}

		controller := api.NetworkInterface("eth0")
		err := controller.ApplyConfig(config)
		assert.NoError(t, err)
	})
}
