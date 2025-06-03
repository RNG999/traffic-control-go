package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/api"
)

func TestConfigValidation(t *testing.T) {
	t.Run("ValidConfiguration", func(t *testing.T) {
		config := &api.TrafficControlConfig{
			Version:   "1.0",
			Device:    "eth0",
			Bandwidth: "1Gbps",
			Classes: []api.TrafficClassConfig{
				{
					Name:       "test",
					Guaranteed: "100Mbps",
					Priority:   &[]int{4}[0],
				},
			},
		}

		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("MissingDevice", func(t *testing.T) {
		config := &api.TrafficControlConfig{
			Version:   "1.0",
			Bandwidth: "1Gbps",
			Classes: []api.TrafficClassConfig{
				{Name: "test", Guaranteed: "100Mbps", Priority: &[]int{4}[0]},
			},
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "device is required")
	})

	t.Run("MissingBandwidth", func(t *testing.T) {
		config := &api.TrafficControlConfig{
			Version: "1.0",
			Device:  "eth0",
			Classes: []api.TrafficClassConfig{
				{Name: "test", Guaranteed: "100Mbps", Priority: &[]int{4}[0]},
			},
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bandwidth is required")
	})

	t.Run("NoClasses", func(t *testing.T) {
		config := &api.TrafficControlConfig{
			Version:   "1.0",
			Device:    "eth0",
			Bandwidth: "1Gbps",
			Classes:   []api.TrafficClassConfig{},
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one class")
	})

	t.Run("DuplicateClassName", func(t *testing.T) {
		config := &api.TrafficControlConfig{
			Version:   "1.0",
			Device:    "eth0",
			Bandwidth: "1Gbps",
			Classes: []api.TrafficClassConfig{
				{Name: "test", Guaranteed: "100Mbps", Priority: &[]int{4}[0]},
				{Name: "test", Guaranteed: "200Mbps", Priority: &[]int{4}[0]},
			},
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate class name")
	})

	t.Run("InvalidRuleTarget", func(t *testing.T) {
		config := &api.TrafficControlConfig{
			Version:   "1.0",
			Device:    "eth0",
			Bandwidth: "1Gbps",
			Classes: []api.TrafficClassConfig{
				{Name: "test", Guaranteed: "100Mbps", Priority: &[]int{4}[0]},
			},
			Rules: []api.TrafficRuleConfig{
				{
					Name:   "rule1",
					Target: "nonexistent",
					Match:  api.MatchConfig{},
				},
			},
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target class")
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestConfigHierarchy(t *testing.T) {
	config := &api.TrafficControlConfig{
		Version:   "1.0",
		Device:    "eth0",
		Bandwidth: "1Gbps",
		Classes: []api.TrafficClassConfig{
			{
				Name:       "parent",
				Guaranteed: "500Mbps",
				Priority:   &[]int{4}[0],
				Children: []api.TrafficClassConfig{
					{
						Name:       "child1",
						Guaranteed: "200Mbps",
						Priority:   &[]int{4}[0],
					},
					{
						Name:       "child2",
						Guaranteed: "300Mbps",
						Priority:   &[]int{5}[0],
						Children: []api.TrafficClassConfig{
							{
								Name:       "grandchild",
								Guaranteed: "100Mbps",
								Priority:   &[]int{6}[0],
							},
						},
					},
				},
			},
		},
		Rules: []api.TrafficRuleConfig{
			{
				Name:   "rule_to_child",
				Target: "parent.child1",
				Match: api.MatchConfig{
					DestPort: []int{80},
				},
			},
			{
				Name:   "rule_to_grandchild",
				Target: "parent.child2.grandchild",
				Match: api.MatchConfig{
					DestPort: []int{443},
				},
			},
		},
	}

	err := config.Validate()
	assert.NoError(t, err)

	// Test that hierarchical names are properly recognized
	classNames := make(map[string]bool)
	var collectClassNames func(class *api.TrafficClassConfig, parentPath string)
	collectClassNames = func(class *api.TrafficClassConfig, parentPath string) {
		fullName := class.Name
		if parentPath != "" {
			fullName = parentPath + "." + class.Name
		}
		classNames[fullName] = true

		for i := range class.Children {
			collectClassNames(&class.Children[i], fullName)
		}
	}

	for i := range config.Classes {
		collectClassNames(&config.Classes[i], "")
	}

	assert.True(t, classNames["parent"])
	assert.True(t, classNames["parent.child1"])
	assert.True(t, classNames["parent.child2"])
	assert.True(t, classNames["parent.child2.grandchild"])
}

func TestApplyConfig(t *testing.T) {
	t.Run("BasicConfiguration", func(t *testing.T) {
		config := &api.TrafficControlConfig{
			Version:   "1.0",
			Device:    "eth0",
			Bandwidth: "100Mbps",
			Classes: []api.TrafficClassConfig{
				{
					Name:       "high",
					Guaranteed: "40Mbps",
					Maximum:    "60Mbps",
					Priority:   &[]int{1}[0],
				},
				{
					Name:       "medium",
					Guaranteed: "30Mbps",
					Priority:   &[]int{4}[0],
				},
				{
					Name:       "low",
					Guaranteed: "30Mbps",
					Priority:   &[]int{6}[0],
				},
			},
		}

		controller := api.NetworkInterface("eth0")
		err := controller.ApplyConfig(config)
		assert.NoError(t, err)
	})

	t.Run("WithDefaults", func(t *testing.T) {
		config := &api.TrafficControlConfig{
			Version:   "1.0",
			Device:    "eth0",
			Bandwidth: "1Gbps",
			Defaults: &api.DefaultConfig{
				BurstRatio: 2.0,
			},
			Classes: []api.TrafficClassConfig{
				{
					Name:       "test1",
					Guaranteed: "100Mbps",
					Priority:   &[]int{4}[0],
					// Should get maximum = 200Mbps from burst ratio
				},
				{
					Name:       "test2",
					Guaranteed: "200Mbps",
					Maximum:    "300Mbps",    // Explicit maximum overrides burst ratio
					Priority:   &[]int{6}[0], // Explicit priority overrides default
				},
			},
		}

		controller := api.NetworkInterface("eth0")
		err := controller.ApplyConfig(config)
		assert.NoError(t, err)
	})

	t.Run("WithRules", func(t *testing.T) {
		config := &api.TrafficControlConfig{
			Version:   "1.0",
			Device:    "eth0",
			Bandwidth: "1Gbps",
			Classes: []api.TrafficClassConfig{
				{
					Name:       "web",
					Guaranteed: "500Mbps",
					Priority:   &[]int{3}[0],
				},
				{
					Name:       "ssh",
					Guaranteed: "100Mbps",
					Priority:   &[]int{1}[0],
				},
			},
			Rules: []api.TrafficRuleConfig{
				{
					Name: "web_traffic",
					Match: api.MatchConfig{
						DestPort: []int{80, 443},
						Protocol: "tcp",
					},
					Target:   "web",
					Priority: 1,
				},
				{
					Name: "ssh_traffic",
					Match: api.MatchConfig{
						DestPort:      []int{22},
						DestinationIP: "10.0.0.0/8",
					},
					Target: "ssh",
				},
			},
		}

		controller := api.NetworkInterface("eth0")
		err := controller.ApplyConfig(config)
		assert.NoError(t, err)
	})
}

func TestConfigurationExamples(t *testing.T) {
	// Test that example configurations are valid
	t.Run("SimpleHomeRouter", func(t *testing.T) {
		config := &api.TrafficControlConfig{
			Version:   "1.0",
			Device:    "eth0",
			Bandwidth: "100Mbps",
			Classes: []api.TrafficClassConfig{
				{Name: "streaming", Guaranteed: "40Mbps", Priority: &[]int{1}[0]},
				{Name: "gaming", Guaranteed: "30Mbps", Priority: &[]int{1}[0]},
				{Name: "general", Guaranteed: "20Mbps", Priority: &[]int{4}[0]},
				{Name: "iot", Guaranteed: "10Mbps", Priority: &[]int{6}[0]},
			},
		}

		require.NoError(t, config.Validate())
	})

	t.Run("EnterpriseMultiTier", func(t *testing.T) {
		config := &api.TrafficControlConfig{
			Version:   "1.0",
			Device:    "eth0",
			Bandwidth: "10Gbps",
			Classes: []api.TrafficClassConfig{
				{
					Name:       "critical",
					Guaranteed: "4Gbps",
					Maximum:    "8Gbps",
					Priority:   &[]int{1}[0],
					Children: []api.TrafficClassConfig{
						{Name: "voip", Guaranteed: "1Gbps", Priority: &[]int{0}[0]},
						{Name: "database", Guaranteed: "3Gbps", Priority: &[]int{2}[0]},
					},
				},
				{
					Name:       "standard",
					Guaranteed: "4Gbps",
					Maximum:    "6Gbps",
					Priority:   &[]int{4}[0],
				},
				{
					Name:       "bulk",
					Guaranteed: "2Gbps",
					Priority:   &[]int{6}[0],
				},
			},
		}

		require.NoError(t, config.Validate())
	})
}
