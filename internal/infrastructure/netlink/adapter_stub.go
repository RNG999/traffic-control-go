//go:build !linux
// +build !linux

package netlink

import (
	"context"
	"fmt"

	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/pkg/logging"
	"github.com/rng999/traffic-control-go/pkg/tc"
	"github.com/rng999/traffic-control-go/pkg/types"
)

// RealNetlinkAdapter is a stub implementation for non-Linux platforms
type RealNetlinkAdapter struct {
	logger logging.Logger
}

// NewRealNetlinkAdapter creates a new stub netlink adapter for non-Linux platforms
func NewRealNetlinkAdapter() *RealNetlinkAdapter {
	logger := logging.WithComponent(logging.ComponentNetlink)
	logger.Warn("Traffic control operations are not supported on this platform")

	return &RealNetlinkAdapter{
		logger: logger,
	}
}

// AddQdisc is not supported on non-Linux platforms
func (a *RealNetlinkAdapter) AddQdisc(ctx context.Context, qdisc *entities.Qdisc) error {
	return fmt.Errorf("traffic control operations are not supported on this platform")
}

// DeleteQdisc is not supported on non-Linux platforms
func (a *RealNetlinkAdapter) DeleteQdisc(device tc.DeviceName, handle tc.Handle) types.Result[Unit] {
	return types.Failure[Unit](fmt.Errorf("traffic control operations are not supported on this platform"))
}

// GetQdiscs is not supported on non-Linux platforms
func (a *RealNetlinkAdapter) GetQdiscs(device tc.DeviceName) types.Result[[]QdiscInfo] {
	return types.Failure[[]QdiscInfo](fmt.Errorf("traffic control operations are not supported on this platform"))
}

// AddClass is not supported on non-Linux platforms
func (a *RealNetlinkAdapter) AddClass(ctx context.Context, class interface{}) error {
	return fmt.Errorf("traffic control operations are not supported on this platform")
}

// DeleteClass is not supported on non-Linux platforms
func (a *RealNetlinkAdapter) DeleteClass(device tc.DeviceName, handle tc.Handle) types.Result[Unit] {
	return types.Failure[Unit](fmt.Errorf("traffic control operations are not supported on this platform"))
}

// GetClasses is not supported on non-Linux platforms
func (a *RealNetlinkAdapter) GetClasses(device tc.DeviceName) types.Result[[]ClassInfo] {
	return types.Failure[[]ClassInfo](fmt.Errorf("traffic control operations are not supported on this platform"))
}

// AddFilter is not supported on non-Linux platforms
func (a *RealNetlinkAdapter) AddFilter(ctx context.Context, filter *entities.Filter) error {
	return fmt.Errorf("traffic control operations are not supported on this platform")
}

// DeleteFilter is not supported on non-Linux platforms
func (a *RealNetlinkAdapter) DeleteFilter(device tc.DeviceName, parent tc.Handle, priority uint16, handle tc.Handle) types.Result[Unit] {
	return types.Failure[Unit](fmt.Errorf("traffic control operations are not supported on this platform"))
}

// GetFilters is not supported on non-Linux platforms
func (a *RealNetlinkAdapter) GetFilters(device tc.DeviceName) types.Result[[]FilterInfo] {
	return types.Failure[[]FilterInfo](fmt.Errorf("traffic control operations are not supported on this platform"))
}

// GetDetailedQdiscStats is not supported on non-Linux platforms
func (a *RealNetlinkAdapter) GetDetailedQdiscStats(device tc.DeviceName, handle tc.Handle) types.Result[DetailedQdiscStats] {
	return types.Failure[DetailedQdiscStats](fmt.Errorf("traffic control operations are not supported on this platform"))
}

// GetDetailedClassStats is not supported on non-Linux platforms
func (a *RealNetlinkAdapter) GetDetailedClassStats(device tc.DeviceName, handle tc.Handle) types.Result[DetailedClassStats] {
	return types.Failure[DetailedClassStats](fmt.Errorf("traffic control operations are not supported on this platform"))
}
