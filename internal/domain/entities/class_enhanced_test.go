package entities

import (
	"testing"

	"github.com/rng999/traffic-control-go/pkg/tc"
	"github.com/stretchr/testify/assert"
)

func TestHTBClass_EnhancedParameters(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *HTBClass
		testFunc func(t *testing.T, class *HTBClass)
	}{
		{
			name: "SetAndGetQuantum",
			setup: func() *HTBClass {
				device, _ := tc.NewDeviceName("eth0")
				handle := tc.NewHandle(1, 10)
				parent := tc.NewHandle(1, 0)
				return NewHTBClass(device, handle, parent, "test-class", Priority(1))
			},
			testFunc: func(t *testing.T, class *HTBClass) {
				quantum := uint32(1500)
				class.SetQuantum(quantum)
				assert.Equal(t, quantum, class.Quantum())
			},
		},
		{
			name: "SetAndGetOverhead",
			setup: func() *HTBClass {
				device, _ := tc.NewDeviceName("eth0")
				handle := tc.NewHandle(1, 10)
				parent := tc.NewHandle(1, 0)
				return NewHTBClass(device, handle, parent, "test-class", Priority(1))
			},
			testFunc: func(t *testing.T, class *HTBClass) {
				overhead := uint32(8)
				class.SetOverhead(overhead)
				assert.Equal(t, overhead, class.Overhead())
			},
		},
		{
			name: "SetAndGetMPU",
			setup: func() *HTBClass {
				device, _ := tc.NewDeviceName("eth0")
				handle := tc.NewHandle(1, 10)
				parent := tc.NewHandle(1, 0)
				return NewHTBClass(device, handle, parent, "test-class", Priority(1))
			},
			testFunc: func(t *testing.T, class *HTBClass) {
				mpu := uint32(64)
				class.SetMPU(mpu)
				assert.Equal(t, mpu, class.MPU())
			},
		},
		{
			name: "SetAndGetMTU",
			setup: func() *HTBClass {
				device, _ := tc.NewDeviceName("eth0")
				handle := tc.NewHandle(1, 10)
				parent := tc.NewHandle(1, 0)
				return NewHTBClass(device, handle, parent, "test-class", Priority(1))
			},
			testFunc: func(t *testing.T, class *HTBClass) {
				mtu := uint32(1500)
				class.SetMTU(mtu)
				assert.Equal(t, mtu, class.MTU())
			},
		},
		{
			name: "SetAndGetHTBPrio",
			setup: func() *HTBClass {
				device, _ := tc.NewDeviceName("eth0")
				handle := tc.NewHandle(1, 10)
				parent := tc.NewHandle(1, 0)
				return NewHTBClass(device, handle, parent, "test-class", Priority(1))
			},
			testFunc: func(t *testing.T, class *HTBClass) {
				htbPrio := uint32(3)
				class.SetHTBPrio(htbPrio)
				assert.Equal(t, htbPrio, class.HTBPrio())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			class := tt.setup()
			tt.testFunc(t, class)
		})
	}
}

func TestHTBClass_CalculateQuantum(t *testing.T) {
	tests := []struct {
		name           string
		rate           string
		expectedMin    uint32
		expectedMax    uint32
		description    string
	}{
		{
			name:           "ZeroRate",
			rate:           "1bps", // Use minimal valid rate instead of 0
			expectedMin:    1000,
			expectedMax:    1000,
			description:    "Very low rate should return minimum quantum",
		},
		{
			name:           "LowRate_1Mbps",
			rate:           "1Mbps",
			expectedMin:    1000,
			expectedMax:    1000,
			description:    "Low rate should return minimum quantum",
		},
		{
			name:           "MediumRate_10Mbps",
			rate:           "10Mbps",
			expectedMin:    1000,
			expectedMax:    2000,
			description:    "Medium rate should calculate proportional quantum",
		},
		{
			name:           "HighRate_100Mbps",
			rate:           "100Mbps",
			expectedMin:    10000,
			expectedMax:    15000,
			description:    "High rate should calculate proportional quantum",
		},
		{
			name:           "VeryHighRate_1Gbps",
			rate:           "1Gbps",
			expectedMin:    60000,
			expectedMax:    60000,
			description:    "Very high rate should be capped at maximum quantum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device, _ := tc.NewDeviceName("eth0")
			handle := tc.NewHandle(1, 10)
			parent := tc.NewHandle(1, 0)
			class := NewHTBClass(device, handle, parent, "test-class", Priority(1))

			bandwidth, err := tc.NewBandwidth(tt.rate)
			assert.NoError(t, err, "Failed to create bandwidth")
			
			class.SetRate(bandwidth)
			quantum := class.CalculateQuantum()

			assert.GreaterOrEqual(t, quantum, tt.expectedMin, 
				"Quantum should be >= expected minimum for %s", tt.description)
			assert.LessOrEqual(t, quantum, tt.expectedMax, 
				"Quantum should be <= expected maximum for %s", tt.description)
		})
	}
}

func TestHTBClass_CalculateEnhancedBurst(t *testing.T) {
	tests := []struct {
		name        string
		rate        string
		mtu         uint32
		overhead    uint32
		expectedMin uint32
		description string
	}{
		{
			name:        "ZeroRate", 
			rate:        "1bps", // Use minimal valid rate instead of 0
			expectedMin: 1600,
			description: "Very low rate should return default minimum burst",
		},
		{
			name:        "LowRate_1Mbps",
			rate:        "1Mbps",
			expectedMin: 1600,
			description: "Low rate should consider minimum burst",
		},
		{
			name:        "MediumRate_10Mbps",
			rate:        "10Mbps",
			expectedMin: 3000,
			description: "Medium rate should calculate proportional burst",
		},
		{
			name:        "HighRate_100Mbps",
			rate:        "100Mbps",
			expectedMin: 80000,
			description: "High rate should calculate large burst",
		},
		{
			name:        "WithMTU",
			rate:        "10Mbps",
			mtu:         9000, // Jumbo frames
			expectedMin: 18000, // At least 2 * MTU
			description: "Should consider MTU for minimum burst",
		},
		{
			name:        "WithOverhead",
			rate:        "10Mbps",
			overhead:    20,
			expectedMin: 3000,
			description: "Should add overhead to burst calculation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device, _ := tc.NewDeviceName("eth0")
			handle := tc.NewHandle(1, 10)
			parent := tc.NewHandle(1, 0)
			class := NewHTBClass(device, handle, parent, "test-class", Priority(1))

			bandwidth, err := tc.NewBandwidth(tt.rate)
			assert.NoError(t, err, "Failed to create bandwidth")
			
			class.SetRate(bandwidth)
			if tt.mtu > 0 {
				class.SetMTU(tt.mtu)
			}
			if tt.overhead > 0 {
				class.SetOverhead(tt.overhead)
			}

			burst := class.CalculateEnhancedBurst()

			assert.GreaterOrEqual(t, burst, tt.expectedMin, 
				"Enhanced burst should be >= expected minimum for %s", tt.description)
			assert.Greater(t, burst, uint32(0), 
				"Enhanced burst should be positive for %s", tt.description)
		})
	}
}

func TestHTBClass_ApplyDefaultParameters(t *testing.T) {
	device, _ := tc.NewDeviceName("eth0")
	handle := tc.NewHandle(1, 10)
	parent := tc.NewHandle(1, 0)
	class := NewHTBClass(device, handle, parent, "test-class", Priority(1))

	// Set a rate to enable calculations
	bandwidth, err := tc.NewBandwidth("10Mbps")
	assert.NoError(t, err)
	class.SetRate(bandwidth)

	// Apply defaults
	class.ApplyDefaultParameters()

	// Verify all parameters are set
	assert.Greater(t, class.Quantum(), uint32(0), "Quantum should be set")
	assert.Equal(t, uint32(1500), class.MTU(), "MTU should be set to default")
	assert.Equal(t, uint32(64), class.MPU(), "MPU should be set to default")
	assert.Equal(t, uint32(4), class.Overhead(), "Overhead should be set to default")
	assert.Greater(t, class.Burst(), uint32(0), "Burst should be calculated")
	assert.Greater(t, class.Cburst(), uint32(0), "Cburst should be calculated")
}

func TestHTBClass_CalculateEnhancedCburst(t *testing.T) {
	device, _ := tc.NewDeviceName("eth0")
	handle := tc.NewHandle(1, 10)
	parent := tc.NewHandle(1, 0)
	class := NewHTBClass(device, handle, parent, "test-class", Priority(1))

	// Set rate and ceil
	rate, err := tc.NewBandwidth("10Mbps")
	assert.NoError(t, err)
	ceil, err := tc.NewBandwidth("20Mbps")
	assert.NoError(t, err)

	class.SetRate(rate)
	class.SetCeil(ceil)

	cburst := class.CalculateEnhancedCburst()
	burst := class.CalculateEnhancedBurst()

	// Cburst should be larger than burst since ceil > rate
	assert.Greater(t, cburst, burst, "Cburst should be larger than burst when ceil > rate")
	assert.Greater(t, cburst, uint32(0), "Cburst should be positive")
}

func TestHTBClass_QuantumBounds(t *testing.T) {
	device, _ := tc.NewDeviceName("eth0")
	handle := tc.NewHandle(1, 10)
	parent := tc.NewHandle(1, 0)
	class := NewHTBClass(device, handle, parent, "test-class", Priority(1))

	tests := []struct {
		name string
		rate string
	}{
		{"VeryLowRate", "1bps"},
		{"LowRate", "1Kbps"},
		{"MediumRate", "1Mbps"},
		{"HighRate", "100Mbps"},
		{"VeryHighRate", "10Gbps"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bandwidth, err := tc.NewBandwidth(tt.rate)
			assert.NoError(t, err)
			
			class.SetRate(bandwidth)
			quantum := class.CalculateQuantum()

			// Quantum should always be within bounds
			assert.GreaterOrEqual(t, quantum, uint32(1000), "Quantum should be >= 1000")
			assert.LessOrEqual(t, quantum, uint32(60000), "Quantum should be <= 60000")
		})
	}
}