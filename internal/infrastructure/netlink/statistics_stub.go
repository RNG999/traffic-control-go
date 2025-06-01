//go:build !linux
// +build !linux

package netlink

import (
	"fmt"

	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/pkg/types"
)

var errNotSupported = fmt.Errorf("traffic control operations are not supported on this platform")

// GetQdiscStatistics returns an error on non-Linux platforms
func GetQdiscStatistics(device valueobjects.DeviceName, handle valueobjects.Handle) types.Result[QdiscStats] {
	return types.Failure[QdiscStats](errNotSupported)
}

// GetClassStatistics returns an error on non-Linux platforms
func GetClassStatistics(device valueobjects.DeviceName, handle valueobjects.Handle) types.Result[ClassStats] {
	return types.Failure[ClassStats](errNotSupported)
}