package valueobjects_test

import (
	"strings"
	"testing"
	
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			d, err := valueobjects.NewDeviceName(tt.input)
			
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
		d := valueobjects.MustNewDeviceName("eth0")
		assert.Equal(t, "eth0", d.String())
	})
	
	t.Run("Invalid input panics", func(t *testing.T) {
		assert.Panics(t, func() {
			valueobjects.MustNewDeviceName("")
		})
	})
}

func TestDeviceNameEquals(t *testing.T) {
	d1 := valueobjects.MustNewDeviceName("eth0")
	d2 := valueobjects.MustNewDeviceName("eth0")
	d3 := valueobjects.MustNewDeviceName("eth1")
	
	assert.True(t, d1.Equals(d2))
	assert.False(t, d1.Equals(d3))
}

func TestDeviceNameString(t *testing.T) {
	d := valueobjects.MustNewDeviceName("wlan0")
	assert.Equal(t, "wlan0", d.String())
}

func TestDeviceNameImmutability(t *testing.T) {
	// DeviceName should be immutable
	name := "eth0"
	d := valueobjects.MustNewDeviceName(name)
	
	// Even if we modify the original string variable,
	// the DeviceName should remain unchanged
	originalStr := d.String()
	name = "modified"
	
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
			_, err := valueobjects.NewDeviceName(name)
			assert.NoError(t, err, "Expected %s to be a valid device name", name)
		})
	}
}

func TestDeviceNameBoundaryLengths(t *testing.T) {
	// Test boundary conditions for name length
	tests := []struct {
		name    string
		length  int
		valid   bool
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
			
			_, err := valueobjects.NewDeviceName(name)
			
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "too long")
			}
		})
	}
}