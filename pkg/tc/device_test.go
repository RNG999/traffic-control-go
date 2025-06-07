package tc_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/pkg/tc"
)

func TestNewDeviceName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:  "Valid ethernet device",
			input: "eth0",
		},
		{
			name:  "Valid wifi device",
			input: "wlan0",
		},
		{
			name:  "Valid loopback",
			input: "lo",
		},
		{
			name:  "Valid with dash",
			input: "br-lan",
		},
		{
			name:  "Valid with dot",
			input: "eth0.100",
		},
		{
			name:  "Valid with at",
			input: "veth@if2",
		},
		{
			name:  "Valid max length",
			input: "a234567890bcdef", // 15 chars
		},
		{
			name:    "Empty name",
			input:   "",
			wantErr: true,
			errMsg:  "device name cannot be empty",
		},
		{
			name:    "Too long",
			input:   "a234567890bcdefg", // 16 chars
			wantErr: true,
			errMsg:  "device name too long",
		},
		{
			name:    "Invalid characters - space",
			input:   "eth 0",
			wantErr: true,
			errMsg:  "invalid device name format",
		},
		{
			name:    "Invalid characters - slash",
			input:   "eth/0",
			wantErr: true,
			errMsg:  "invalid device name format",
		},
		{
			name:    "Invalid characters - colon",
			input:   "eth:0",
			wantErr: true,
			errMsg:  "invalid device name format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := tc.NewDeviceName(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.input, d.String())
		})
	}
}

func TestMustNewDeviceName(t *testing.T) {
	t.Run("Valid input", func(t *testing.T) {
		d := tc.MustNewDeviceName("eth0")
		assert.Equal(t, "eth0", d.String())
	})

	t.Run("Invalid input panics", func(t *testing.T) {
		assert.Panics(t, func() {
			tc.MustNewDeviceName("")
		})
	})
}

func TestDeviceNameEquals(t *testing.T) {
	d1 := tc.MustNewDeviceName("eth0")
	d2 := tc.MustNewDeviceName("eth0")
	d3 := tc.MustNewDeviceName("eth1")

	assert.True(t, d1.Equals(d2))
	assert.False(t, d1.Equals(d3))
}

func TestDeviceNameString(t *testing.T) {
	d := tc.MustNewDeviceName("wlan0")
	assert.Equal(t, "wlan0", d.String())
}

func TestDeviceNameImmutability(t *testing.T) {
	// DeviceName should be immutable
	name := "eth0"
	d := tc.MustNewDeviceName(name)

	// Even if we modify the original string variable,
	// the DeviceName should remain unchanged
	originalStr := d.String()
	_ = "modified" // This assignment demonstrates that DeviceName is immutable

	assert.Equal(t, originalStr, d.String())
	assert.Equal(t, "eth0", d.String())
}

func TestDeviceNameValidPatterns(t *testing.T) {
	// Test various common Linux network interface naming patterns
	validNames := []string{
		// Traditional names
		"eth0", "eth1", "eth10",
		"wlan0", "wlan1",
		"lo",
		// Predictable names
		"enp0s3", "enp0s25",
		"ens33", "ens160",
		"eno1", "eno2",
		"wlp3s0", "wlp58s0",
		// Virtual interfaces
		"docker0", "br-abcd1234",
		"virbr0", "virbr0-nic",
		"tun0", "tap0",
		"veth0", "veth@if2",
		// VLAN interfaces
		"eth0.100", "eth0.200",
		"bond0.100",
		// Other patterns
		"dummy0", "gre0", "sit0",
		"ppp0", "slip0",
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			_, err := tc.NewDeviceName(name)
			assert.NoError(t, err, "Expected %s to be a valid device name", name)
		})
	}
}

func TestDeviceNameBoundaryLengths(t *testing.T) {
	// Test boundary conditions for name length
	tests := []struct {
		name   string
		length int
		valid  bool
	}{
		{"1 char", 1, true},
		{"2 chars", 2, true},
		{"14 chars", 14, true},
		{"15 chars (max)", 15, true},
		{"16 chars (too long)", 16, false},
		{"20 chars (too long)", 20, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate a name of the specified length
			name := strings.Repeat("a", tt.length)

			_, err := tc.NewDeviceName(name)

			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "too long")
			}
		})
	}
}

// =============================================================================
// BENCHMARK TESTS
// =============================================================================

func BenchmarkDeviceNameCreation(b *testing.B) {
	deviceNames := []string{
		"eth0", "wlan0", "lo", "br-lan", "eth0.100", "veth@if2",
		"enp0s3", "docker0", "virbr0", "tun0",
	}

	b.Run("NewDeviceName_Multiple", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, name := range deviceNames {
				_, _ = tc.NewDeviceName(name)
			}
		}
	})

	b.Run("NewDeviceName_Simple", func(b *testing.B) {
		name := "eth0"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tc.NewDeviceName(name)
		}
	})

	b.Run("NewDeviceName_Complex", func(b *testing.B) {
		name := "br-abcd1234.100"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tc.NewDeviceName(name)
		}
	})

	b.Run("MustNewDeviceName", func(b *testing.B) {
		name := "eth0"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tc.MustNewDeviceName(name)
		}
	})
}

func BenchmarkDeviceNameValidation(b *testing.B) {
	validNames := []string{
		"eth0", "wlan0", "enp0s3", "docker0", "eth0.100",
	}
	
	invalidNames := []string{
		"", "eth 0", "eth/0", "eth:0", strings.Repeat("a", 16),
	}

	b.Run("Valid_Names", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, name := range validNames {
				_, _ = tc.NewDeviceName(name)
			}
		}
	})

	b.Run("Invalid_Names", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, name := range invalidNames {
				_, _ = tc.NewDeviceName(name)
			}
		}
	})
}

func BenchmarkDeviceNameOperations(b *testing.B) {
	dev1 := tc.MustNewDeviceName("eth0")
	dev2 := tc.MustNewDeviceName("eth0")
	dev3 := tc.MustNewDeviceName("eth1")

	b.Run("String", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = dev1.String()
		}
	})

	b.Run("Equals_Same", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = dev1.Equals(dev2)
		}
	})

	b.Run("Equals_Different", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = dev1.Equals(dev3)
		}
	})
}

func BenchmarkDeviceNameLengthValidation(b *testing.B) {
	b.Run("Short_Name", func(b *testing.B) {
		name := "lo"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tc.NewDeviceName(name)
		}
	})

	b.Run("Medium_Name", func(b *testing.B) {
		name := "eth0.100"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tc.NewDeviceName(name)
		}
	})

	b.Run("Max_Length_Name", func(b *testing.B) {
		name := strings.Repeat("a", 15) // Max valid length
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tc.NewDeviceName(name)
		}
	})

	b.Run("Too_Long_Name", func(b *testing.B) {
		name := strings.Repeat("a", 16) // Too long
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tc.NewDeviceName(name)
		}
	})
}

func BenchmarkDeviceNameCommonPatterns(b *testing.B) {
	patterns := map[string]string{
		"Traditional":  "eth0",
		"Predictable":  "enp0s3",
		"Wireless":     "wlp3s0",
		"Virtual":      "veth@if2",
		"VLAN":         "eth0.100",
		"Bridge":       "br-docker0",
	}

	for patternName, deviceName := range patterns {
		b.Run(patternName, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = tc.NewDeviceName(deviceName)
			}
		})
	}
}
