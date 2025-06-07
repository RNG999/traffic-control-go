package tc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/pkg/tc"
)

func TestBandwidthCreation(t *testing.T) {
	tests := []struct {
		name     string
		create   func() tc.Bandwidth
		expected uint64 // bits per second
	}{
		{
			name:     "Bps creation",
			create:   func() tc.Bandwidth { return tc.Bps(1000) },
			expected: 1000,
		},
		{
			name:     "Kbps creation",
			create:   func() tc.Bandwidth { return tc.Kbps(100) },
			expected: 100_000,
		},
		{
			name:     "Mbps creation",
			create:   func() tc.Bandwidth { return tc.Mbps(10) },
			expected: 10_000_000,
		},
		{
			name:     "Gbps creation",
			create:   func() tc.Bandwidth { return tc.Gbps(1) },
			expected: 1_000_000_000,
		},
		{
			name:     "Fractional Mbps",
			create:   func() tc.Bandwidth { return tc.Mbps(1.5) },
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
			b, err := tc.ParseBandwidth(tt.input)

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
		b := tc.MustParseBandwidth("100Mbps")
		assert.Equal(t, uint64(100_000_000), b.BitsPerSecond())
	})

	t.Run("Invalid input panics", func(t *testing.T) {
		assert.Panics(t, func() {
			tc.MustParseBandwidth("invalid")
		})
	})
}

func TestBandwidthConversions(t *testing.T) {
	b := tc.Gbps(1.5)

	assert.Equal(t, uint64(1_500_000_000), b.BitsPerSecond())
	assert.Equal(t, 1_500_000.0, b.KilobitsPerSecond())
	assert.Equal(t, 1_500.0, b.MegabitsPerSecond())
	assert.Equal(t, 1.5, b.GigabitsPerSecond())
}

func TestBandwidthHumanReadable(t *testing.T) {
	tests := []struct {
		name      string
		bandwidth tc.Bandwidth
		expected  string
	}{
		{
			name:      "Display as bps",
			bandwidth: tc.Bps(500),
			expected:  "500bps",
		},
		{
			name:      "Display as Kbps",
			bandwidth: tc.Kbps(100),
			expected:  "100.0Kbps",
		},
		{
			name:      "Display as Mbps",
			bandwidth: tc.Mbps(50),
			expected:  "50.0Mbps",
		},
		{
			name:      "Display as Gbps",
			bandwidth: tc.Gbps(2.5),
			expected:  "2.5Gbps",
		},
		{
			name:      "Auto-format large bps as Kbps",
			bandwidth: tc.Bps(5000),
			expected:  "5.0Kbps",
		},
		{
			name:      "Auto-format large Kbps as Mbps",
			bandwidth: tc.Kbps(5000),
			expected:  "5.0Mbps",
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
	b1 := tc.Mbps(100)
	b2 := tc.Mbps(50)
	b3 := tc.Mbps(100)

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
	b1 := tc.Mbps(100)
	b2 := tc.Mbps(50)

	t.Run("Addition", func(t *testing.T) {
		result := b1.Add(b2)
		assert.Equal(t, tc.Mbps(150).BitsPerSecond(), result.BitsPerSecond())
	})

	t.Run("Subtraction", func(t *testing.T) {
		result := b1.Subtract(b2)
		assert.Equal(t, tc.Mbps(50).BitsPerSecond(), result.BitsPerSecond())
	})

	t.Run("Subtraction with underflow", func(t *testing.T) {
		result := b2.Subtract(b1)
		assert.Equal(t, uint64(0), result.BitsPerSecond())
	})

	t.Run("Multiply", func(t *testing.T) {
		result := b1.MultiplyBy(1.5)
		assert.Equal(t, tc.Mbps(150).BitsPerSecond(), result.BitsPerSecond())
	})

	t.Run("Percentage", func(t *testing.T) {
		result := b1.Percentage(25)
		assert.Equal(t, tc.Mbps(25).BitsPerSecond(), result.BitsPerSecond())
	})
}

func TestBandwidthImmutability(t *testing.T) {
	original := tc.Mbps(100)

	// All operations should return new instances
	_ = original.Add(tc.Mbps(50))
	_ = original.Subtract(tc.Mbps(50))
	_ = original.MultiplyBy(2)
	_ = original.Percentage(50)

	// Original should remain unchanged
	assert.Equal(t, tc.Mbps(100).BitsPerSecond(), original.BitsPerSecond())
}

// =============================================================================
// BENCHMARK TESTS
// =============================================================================

func BenchmarkBandwidthParsing(b *testing.B) {
	testCases := []string{
		"100bps", "1kbps", "100kbps", "1mbps", "100mbps", "1gbps",
		"1.5Mbps", "10.5Gbps", "500Kbps", "2048bps",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, testCase := range testCases {
			_, _ = tc.ParseBandwidth(testCase)
		}
	}
}

func BenchmarkBandwidthParsingSimple(b *testing.B) {
	input := "100mbps"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tc.ParseBandwidth(input)
	}
}

func BenchmarkBandwidthParsingComplex(b *testing.B) {
	input := "  1.5 Gbps  " // With spaces and decimal
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tc.ParseBandwidth(input)
	}
}

func BenchmarkMustParseBandwidth(b *testing.B) {
	input := "100mbps"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tc.MustParseBandwidth(input)
	}
}

func BenchmarkBandwidthCreation(b *testing.B) {
	b.Run("Bps", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = tc.Bps(1000)
		}
	})

	b.Run("Kbps", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = tc.Kbps(100)
		}
	})

	b.Run("Mbps", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = tc.Mbps(100)
		}
	})

	b.Run("Gbps", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = tc.Gbps(1)
		}
	})
}

func BenchmarkBandwidthArithmetic(b *testing.B) {
	b1 := tc.Mbps(100)
	b2 := tc.Mbps(50)

	b.Run("Add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = b1.Add(b2)
		}
	})

	b.Run("Subtract", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = b1.Subtract(b2)
		}
	})

	b.Run("MultiplyBy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = b1.MultiplyBy(1.5)
		}
	})

	b.Run("Percentage", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = b1.Percentage(25)
		}
	})
}

func BenchmarkBandwidthComparisons(b *testing.B) {
	b1 := tc.Mbps(100)
	b2 := tc.Mbps(50)
	b3 := tc.Mbps(100)

	b.Run("Equals", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = b1.Equals(b3)
		}
	})

	b.Run("GreaterThan", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = b1.GreaterThan(b2)
		}
	})

	b.Run("LessThan", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = b2.LessThan(b1)
		}
	})
}

func BenchmarkBandwidthFormatting(b *testing.B) {
	bandwidth := tc.Mbps(150.75)

	b.Run("String", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bandwidth.String()
		}
	})

	b.Run("HumanReadable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bandwidth.HumanReadable()
		}
	})

	b.Run("BitsPerSecond", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bandwidth.BitsPerSecond()
		}
	})

	b.Run("MegabitsPerSecond", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bandwidth.MegabitsPerSecond()
		}
	})
}

func BenchmarkBandwidthParsingVsCreation(b *testing.B) {
	b.Run("ParseBandwidth", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = tc.ParseBandwidth("100mbps")
		}
	})

	b.Run("DirectCreation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = tc.Mbps(100)
		}
	})
}
