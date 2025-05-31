package valueobjects

import (
	"fmt"
	"strconv"
	"strings"
)

// Handle represents a TC handle (major:minor format)
type Handle struct {
	major uint16
	minor uint16
}

// NewHandle creates a new handle with major and minor numbers
func NewHandle(major, minor uint16) Handle {
	return Handle{major: major, minor: minor}
}

// ParseHandle parses a handle from string format "major:minor" or "major:"
func ParseHandle(s string) (Handle, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return Handle{}, fmt.Errorf("invalid handle format: %s (expected 'major:minor')", s)
	}

	// Parse major (required)
	major, err := strconv.ParseUint(parts[0], 16, 16)
	if err != nil {
		return Handle{}, fmt.Errorf("invalid major number: %s", parts[0])
	}

	// Parse minor (optional, can be empty)
	var minor uint64
	if parts[1] != "" {
		minor, err = strconv.ParseUint(parts[1], 16, 16)
		if err != nil {
			return Handle{}, fmt.Errorf("invalid minor number: %s", parts[1])
		}
	}

	return Handle{
		major: uint16(major),
		minor: uint16(minor),
	}, nil
}

// MustParseHandle parses a handle or panics
func MustParseHandle(s string) Handle {
	h, err := ParseHandle(s)
	if err != nil {
		panic(err)
	}
	return h
}

// Major returns the major number
func (h Handle) Major() uint16 {
	return h.major
}

// Minor returns the minor number
func (h Handle) Minor() uint16 {
	return h.minor
}

// String returns the handle in "major:minor" format
func (h Handle) String() string {
	if h.minor == 0 {
		return fmt.Sprintf("%x:", h.major)
	}
	return fmt.Sprintf("%x:%x", h.major, h.minor)
}

// IsRoot checks if this is a root handle (minor == 0)
func (h Handle) IsRoot() bool {
	return h.minor == 0
}

// Equals checks if two handles are equal
func (h Handle) Equals(other Handle) bool {
	return h.major == other.major && h.minor == other.minor
}

// ToUint32 converts the handle to a 32-bit representation used by netlink
func (h Handle) ToUint32() uint32 {
	return (uint32(h.major) << 16) | uint32(h.minor)
}

// HandleFromUint32 creates a Handle from a 32-bit netlink representation
func HandleFromUint32(u uint32) Handle {
	return Handle{
		major: uint16(u >> 16),
		minor: uint16(u & 0xFFFF),
	}
}
