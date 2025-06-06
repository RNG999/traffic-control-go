package tc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/pkg/tc"
)

func TestNewHandle(t *testing.T) {
	h := tc.NewHandle(1, 10)

	assert.Equal(t, uint16(1), h.Major())
	assert.Equal(t, uint16(10), h.Minor())
	assert.Equal(t, "1:a", h.String())
}

func TestParseHandle(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantMajor uint16
		wantMinor uint16
		wantErr   bool
	}{
		{
			name:      "Parse simple handle",
			input:     "1:10",
			wantMajor: 1,
			wantMinor: 16, // 0x10 = 16
		},
		{
			name:      "Parse hex handle",
			input:     "ff:20",
			wantMajor: 255,
			wantMinor: 32,
		},
		{
			name:      "Parse root handle",
			input:     "1:",
			wantMajor: 1,
			wantMinor: 0,
		},
		{
			name:      "Parse with leading zeros",
			input:     "01:0a",
			wantMajor: 1,
			wantMinor: 10,
		},
		{
			name:    "Invalid format - no colon",
			input:   "110",
			wantErr: true,
		},
		{
			name:    "Invalid format - too many colons",
			input:   "1:10:20",
			wantErr: true,
		},
		{
			name:    "Invalid major",
			input:   "xyz:10",
			wantErr: true,
		},
		{
			name:    "Invalid minor",
			input:   "1:xyz",
			wantErr: true,
		},
		{
			name:    "Major too large",
			input:   "10000:1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := tc.ParseHandle(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantMajor, h.Major())
			assert.Equal(t, tt.wantMinor, h.Minor())
		})
	}
}

func TestMustParseHandle(t *testing.T) {
	t.Run("Valid input", func(t *testing.T) {
		h := tc.MustParseHandle("1:10")
		assert.Equal(t, uint16(1), h.Major())
		assert.Equal(t, uint16(16), h.Minor())
	})

	t.Run("Invalid input panics", func(t *testing.T) {
		assert.Panics(t, func() {
			tc.MustParseHandle("invalid")
		})
	})
}

func TestHandleString(t *testing.T) {
	tests := []struct {
		name     string
		handle   tc.Handle
		expected string
	}{
		{
			name:     "Simple handle",
			handle:   tc.NewHandle(1, 10),
			expected: "1:a",
		},
		{
			name:     "Root handle",
			handle:   tc.NewHandle(1, 0),
			expected: "1:",
		},
		{
			name:     "Large numbers",
			handle:   tc.NewHandle(255, 65535),
			expected: "ff:ffff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.handle.String())
		})
	}
}

func TestHandleIsRoot(t *testing.T) {
	root := tc.NewHandle(1, 0)
	nonRoot := tc.NewHandle(1, 10)

	assert.True(t, root.IsRoot())
	assert.False(t, nonRoot.IsRoot())
}

func TestHandleEquals(t *testing.T) {
	h1 := tc.NewHandle(1, 10)
	h2 := tc.NewHandle(1, 10)
	h3 := tc.NewHandle(1, 20)
	h4 := tc.NewHandle(2, 10)

	assert.True(t, h1.Equals(h2))
	assert.False(t, h1.Equals(h3))
	assert.False(t, h1.Equals(h4))
}

func TestHandleUint32Conversion(t *testing.T) {
	tests := []struct {
		name     string
		handle   tc.Handle
		expected uint32
	}{
		{
			name:     "Simple conversion",
			handle:   tc.NewHandle(1, 10),
			expected: 0x0001000A,
		},
		{
			name:     "Root handle",
			handle:   tc.NewHandle(1, 0),
			expected: 0x00010000,
		},
		{
			name:     "Max values",
			handle:   tc.NewHandle(0xFFFF, 0xFFFF),
			expected: 0xFFFFFFFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.handle.ToUint32())
		})
	}
}

func TestHandleFromUint32(t *testing.T) {
	tests := []struct {
		name      string
		input     uint32
		wantMajor uint16
		wantMinor uint16
	}{
		{
			name:      "Simple conversion",
			input:     0x0001000A,
			wantMajor: 1,
			wantMinor: 10,
		},
		{
			name:      "Root handle",
			input:     0x00010000,
			wantMajor: 1,
			wantMinor: 0,
		},
		{
			name:      "Max values",
			input:     0xFFFFFFFF,
			wantMajor: 0xFFFF,
			wantMinor: 0xFFFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tc.HandleFromUint32(tt.input)
			assert.Equal(t, tt.wantMajor, h.Major())
			assert.Equal(t, tt.wantMinor, h.Minor())
		})
	}
}

func TestHandleRoundTripConversion(t *testing.T) {
	// Test that converting to uint32 and back preserves the handle
	original := tc.NewHandle(123, 456)
	uint32Val := original.ToUint32()
	restored := tc.HandleFromUint32(uint32Val)

	assert.True(t, original.Equals(restored))
}
