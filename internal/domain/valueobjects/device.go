package valueobjects

import (
	"fmt"
	"regexp"
)

// DeviceName represents a network interface name
type DeviceName struct {
	value string
}

// NewDevice creates a new DeviceName with validation (alias for consistency)
func NewDevice(name string) (DeviceName, error) {
	return NewDeviceName(name)
}

// NewDeviceName creates a new DeviceName with validation
func NewDeviceName(name string) (DeviceName, error) {
	if err := validateDeviceName(name); err != nil {
		return DeviceName{}, err
	}
	return DeviceName{value: name}, nil
}

// MustNewDeviceName creates a new DeviceName or panics
func MustNewDeviceName(name string) DeviceName {
	d, err := NewDeviceName(name)
	if err != nil {
		panic(err)
	}
	return d
}

// validateDeviceName checks if the device name is valid
func validateDeviceName(name string) error {
	if name == "" {
		return fmt.Errorf("device name cannot be empty")
	}

	if len(name) > 15 { // Linux IFNAMSIZ limit
		return fmt.Errorf("device name too long (max 15 characters): %s", name)
	}

	// Valid characters: alphanumeric, dash, dot, @
	validName := regexp.MustCompile(`^[a-zA-Z0-9\-\.@]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid device name format: %s", name)
	}

	return nil
}

// String returns the device name
func (d DeviceName) String() string {
	return d.value
}

// Equals checks if two device names are equal
func (d DeviceName) Equals(other DeviceName) bool {
	return d.value == other.value
}
