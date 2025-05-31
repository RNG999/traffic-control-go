package valueobjects_test

import (
	"testing"
	
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBandwidthCreation(t *testing.T) {
	tests := []struct {
		name     string
		create   func() valueobjects.Bandwidth
		expected uint64 // bits per second
	}{
		{
			name:     "Bps creation",
			create:   func() valueobjects.Bandwidth { return valueobjects.Bps(1000) },
			expected: 1000,
		},
		{
			name:     "Kbps creation",
			create:   func() valueobjects.Bandwidth { return valueobjects.Kbps(100) },
			expected: 100_000,
		},
		{
			name:     "Mbps creation",
			create:   func() valueobjects.Bandwidth { return valueobjects.Mbps(10) },
			expected: 10_000_000,
		},
		{
			name:     "Gbps creation",
			create:   func() valueobjects.Bandwidth { return valueobjects.Gbps(1) },
			expected: 1_000_000_000,
		},
		{
			name:     "Fractional Mbps",
			create:   func() valueobjects.Bandwidth { return valueobjects.Mbps(1.5) },
			expected: 1_500_000,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.create()
			assert.Equal(t, tt.expected, b.BitsPerSecond())
		})
	}
}

func TestParseBandwidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint64
		wantErr  bool
	}{
		{
			name:     "Parse 100bps",
			input:    "100bps",
			expected: 100,
		},
		{
			name:     "Parse 100kbps",
			input:    "100kbps",
			expected: 100_000,
		},
		{
			name:     "Parse 100mbps",
			input:    "100mbps",
			expected: 100_000_000,
		},
		{
			name:     "Parse 1gbps",
			input:    "1gbps",
			expected: 1_000_000_000,
		},
		{
			name:     "Parse with uppercase",
			input:    "100Mbps",
			expected: 100_000_000,
		},
		{
			name:     "Parse with spaces",
			input:    "  100 mbps  ",
			expected: 100_000_000,
		},
		{
			name:     "Parse decimal",
			input:    "1.5mbps",
			expected: 1_500_000,
		},
		{
			name:    "Invalid format",
			input:   "100",
			wantErr: true,
		},
		{
			name:    "Invalid unit",
			input:   "100tbps",
			wantErr: true,
		},
		{
			name:    "Invalid number",
			input:   "abcmbps",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := valueobjects.ParseBandwidth(tt.input)
			
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.expected, b.BitsPerSecond())
		})
	}
}

func TestMustParseBandwidth(t *testing.T) {
	t.Run("Valid input", func(t *testing.T) {
		b := valueobjects.MustParseBandwidth("100Mbps")
		assert.Equal(t, uint64(100_000_000), b.BitsPerSecond())
	})
	
	t.Run("Invalid input panics", func(t *testing.T) {
		assert.Panics(t, func() {
			valueobjects.MustParseBandwidth("invalid")
		})
	})
}

func TestBandwidthConversions(t *testing.T) {
	b := valueobjects.Gbps(1.5)
	
	assert.Equal(t, uint64(1_500_000_000), b.BitsPerSecond())
	assert.Equal(t, 1_500_000.0, b.KilobitsPerSecond())
	assert.Equal(t, 1_500.0, b.MegabitsPerSecond())
	assert.Equal(t, 1.5, b.GigabitsPerSecond())
}

func TestBandwidthHumanReadable(t *testing.T) {
	tests := []struct {
		name     string
		bandwidth valueobjects.Bandwidth
		expected string
	}{
		{
			name:     "Display as bps",
			bandwidth: valueobjects.Bps(500),
			expected: "500bps",
		},
		{
			name:     "Display as Kbps",
			bandwidth: valueobjects.Kbps(100),
			expected: "100.0Kbps",
		},
		{
			name:     "Display as Mbps",
			bandwidth: valueobjects.Mbps(50),
			expected: "50.0Mbps",
		},
		{
			name:     "Display as Gbps",
			bandwidth: valueobjects.Gbps(2.5),
			expected: "2.5Gbps",
		},
		{
			name:     "Auto-format large bps as Kbps",
			bandwidth: valueobjects.Bps(5000),
			expected: "5.0Kbps",
		},
		{
			name:     "Auto-format large Kbps as Mbps",
			bandwidth: valueobjects.Kbps(5000),
			expected: "5.0Mbps",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.bandwidth.HumanReadable())
			assert.Equal(t, tt.expected, tt.bandwidth.String())
		})
	}
}

func TestBandwidthComparisons(t *testing.T) {
	b1 := valueobjects.Mbps(100)
	b2 := valueobjects.Mbps(50)
	b3 := valueobjects.Mbps(100)
	
	assert.True(t, b1.Equals(b3))
	assert.False(t, b1.Equals(b2))
	
	assert.True(t, b1.GreaterThan(b2))
	assert.False(t, b2.GreaterThan(b1))
	assert.False(t, b1.GreaterThan(b3))
	
	assert.True(t, b2.LessThan(b1))
	assert.False(t, b1.LessThan(b2))
	assert.False(t, b1.LessThan(b3))
}

func TestBandwidthArithmetic(t *testing.T) {
	b1 := valueobjects.Mbps(100)
	b2 := valueobjects.Mbps(50)
	
	t.Run("Addition", func(t *testing.T) {
		result := b1.Add(b2)
		assert.Equal(t, valueobjects.Mbps(150).BitsPerSecond(), result.BitsPerSecond())
	})
	
	t.Run("Subtraction", func(t *testing.T) {
		result := b1.Subtract(b2)
		assert.Equal(t, valueobjects.Mbps(50).BitsPerSecond(), result.BitsPerSecond())
	})
	
	t.Run("Subtraction with underflow", func(t *testing.T) {
		result := b2.Subtract(b1)
		assert.Equal(t, uint64(0), result.BitsPerSecond())
	})
	
	t.Run("Multiply", func(t *testing.T) {
		result := b1.MultiplyBy(1.5)
		assert.Equal(t, valueobjects.Mbps(150).BitsPerSecond(), result.BitsPerSecond())
	})
	
	t.Run("Percentage", func(t *testing.T) {
		result := b1.Percentage(25)
		assert.Equal(t, valueobjects.Mbps(25).BitsPerSecond(), result.BitsPerSecond())
	})
}

func TestBandwidthImmutability(t *testing.T) {
	original := valueobjects.Mbps(100)
	
	// All operations should return new instances
	_ = original.Add(valueobjects.Mbps(50))
	_ = original.Subtract(valueobjects.Mbps(50))
	_ = original.MultiplyBy(2)
	_ = original.Percentage(50)
	
	// Original should remain unchanged
	assert.Equal(t, valueobjects.Mbps(100).BitsPerSecond(), original.BitsPerSecond())
}