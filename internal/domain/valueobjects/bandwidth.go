package valueobjects

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// BandwidthUnit represents the unit of bandwidth measurement
type BandwidthUnit int

const (
	BitsPerSecond BandwidthUnit = iota
	KilobitsPerSecond
	MegabitsPerSecond
	GigabitsPerSecond
)

// Bandwidth represents network bandwidth in a human-readable way
type Bandwidth struct {
	value uint64 // Always stored in bits per second
}

// Predefined bandwidth units for easy creation
func Bps(value uint64) Bandwidth   { return Bandwidth{value: value} }
func Kbps(value float64) Bandwidth { return Bandwidth{value: uint64(value * 1000)} }
func Mbps(value float64) Bandwidth { return Bandwidth{value: uint64(value * 1000 * 1000)} }
func Gbps(value float64) Bandwidth { return Bandwidth{value: uint64(value * 1000 * 1000 * 1000)} }

// MustParseBandwidth parses a bandwidth string like "100Mbps" or "1.5Gbps"
func MustParseBandwidth(s string) Bandwidth {
	b, err := ParseBandwidth(s)
	if err != nil {
		panic(fmt.Sprintf("invalid bandwidth format: %s", s))
	}
	return b
}

// ParseBandwidth parses a bandwidth string with error handling
func ParseBandwidth(s string) (Bandwidth, error) {
	// Regular expression to match number + unit
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*(bps|kbps|mbps|gbps|Bps|Kbps|Mbps|Gbps)$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(s))

	if len(matches) != 3 {
		return Bandwidth{}, fmt.Errorf("invalid bandwidth format: %s (expected format: '100Mbps')", s)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return Bandwidth{}, fmt.Errorf("invalid numeric value: %s", matches[1])
	}

	unit := strings.ToLower(matches[2])

	switch unit {
	case "bps":
		return Bps(uint64(value)), nil
	case "kbps":
		return Kbps(value), nil
	case "mbps":
		return Mbps(value), nil
	case "gbps":
		return Gbps(value), nil
	default:
		return Bandwidth{}, fmt.Errorf("unknown bandwidth unit: %s", unit)
	}
}

// BitsPerSecond returns the bandwidth in bits per second
func (b Bandwidth) BitsPerSecond() uint64 {
	return b.value
}

// KilobitsPerSecond returns the bandwidth in kilobits per second
func (b Bandwidth) KilobitsPerSecond() float64 {
	return float64(b.value) / 1000
}

// MegabitsPerSecond returns the bandwidth in megabits per second
func (b Bandwidth) MegabitsPerSecond() float64 {
	return float64(b.value) / (1000 * 1000)
}

// GigabitsPerSecond returns the bandwidth in gigabits per second
func (b Bandwidth) GigabitsPerSecond() float64 {
	return float64(b.value) / (1000 * 1000 * 1000)
}

// HumanReadable returns a human-friendly string representation
func (b Bandwidth) HumanReadable() string {
	switch {
	case b.value >= 1000*1000*1000:
		return fmt.Sprintf("%.1fGbps", b.GigabitsPerSecond())
	case b.value >= 1000*1000:
		return fmt.Sprintf("%.1fMbps", b.MegabitsPerSecond())
	case b.value >= 1000:
		return fmt.Sprintf("%.1fKbps", b.KilobitsPerSecond())
	default:
		return fmt.Sprintf("%dbps", b.value)
	}
}

// String implements the Stringer interface
func (b Bandwidth) String() string {
	return b.HumanReadable()
}

// Equals checks if two bandwidths are equal
func (b Bandwidth) Equals(other Bandwidth) bool {
	return b.value == other.value
}

// GreaterThan checks if this bandwidth is greater than another
func (b Bandwidth) GreaterThan(other Bandwidth) bool {
	return b.value > other.value
}

// LessThan checks if this bandwidth is less than another
func (b Bandwidth) LessThan(other Bandwidth) bool {
	return b.value < other.value
}

// Add returns a new Bandwidth that is the sum of two bandwidths
func (b Bandwidth) Add(other Bandwidth) Bandwidth {
	return Bandwidth{value: b.value + other.value}
}

// Subtract returns a new Bandwidth that is the difference of two bandwidths
func (b Bandwidth) Subtract(other Bandwidth) Bandwidth {
	if b.value < other.value {
		return Bandwidth{value: 0}
	}
	return Bandwidth{value: b.value - other.value}
}

// MultiplyBy returns a new Bandwidth multiplied by a factor
func (b Bandwidth) MultiplyBy(factor float64) Bandwidth {
	return Bandwidth{value: uint64(float64(b.value) * factor)}
}

// Percentage returns a percentage of the bandwidth
func (b Bandwidth) Percentage(percent float64) Bandwidth {
	return b.MultiplyBy(percent / 100.0)
}
