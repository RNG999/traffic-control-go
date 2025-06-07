package netlink

import (
	"context"

	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/pkg/logging"
	"github.com/rng999/traffic-control-go/pkg/tc"
	"github.com/rng999/traffic-control-go/pkg/types"
)

// AdapterWrapper wraps the real netlink adapter and provides domain entity interfaces
type AdapterWrapper struct {
	adapter Adapter
	logger  logging.Logger
}

// NewAdapter creates a new wrapped adapter
func NewAdapter() Adapter {
	return &AdapterWrapper{
		adapter: NewRealNetlinkAdapter(),
		logger:  logging.WithComponent(logging.ComponentNetlink),
	}
}

// AddQdisc adds a qdisc from domain entity
func (a *AdapterWrapper) AddQdisc(ctx context.Context, qdisc *entities.Qdisc) error {
	// Delegate directly to the adapter
	return a.adapter.AddQdisc(ctx, qdisc)
}

// AddClass adds a class from domain entity
func (a *AdapterWrapper) AddClass(ctx context.Context, class interface{}) error {
	// Delegate directly to the adapter
	return a.adapter.AddClass(ctx, class)
}

// AddFilter adds a filter from domain entity
func (a *AdapterWrapper) AddFilter(ctx context.Context, filter *entities.Filter) error {
	// Delegate directly to the adapter
	return a.adapter.AddFilter(ctx, filter)
}

// DeleteQdisc deletes a qdisc
func (a *AdapterWrapper) DeleteQdisc(device tc.DeviceName, handle tc.Handle) types.Result[Unit] {
	return a.adapter.DeleteQdisc(device, handle)
}

// GetQdiscs returns all qdiscs for a device
func (a *AdapterWrapper) GetQdiscs(device tc.DeviceName) types.Result[[]QdiscInfo] {
	return a.adapter.GetQdiscs(device)
}

// DeleteClass deletes a class
func (a *AdapterWrapper) DeleteClass(device tc.DeviceName, handle tc.Handle) types.Result[Unit] {
	return a.adapter.DeleteClass(device, handle)
}

// GetClasses returns all classes for a device
func (a *AdapterWrapper) GetClasses(device tc.DeviceName) types.Result[[]ClassInfo] {
	return a.adapter.GetClasses(device)
}

// DeleteFilter deletes a filter
func (a *AdapterWrapper) DeleteFilter(device tc.DeviceName, parent tc.Handle, priority uint16, handle tc.Handle) types.Result[Unit] {
	return a.adapter.DeleteFilter(device, parent, priority, handle)
}

// GetFilters returns all filters for a device
func (a *AdapterWrapper) GetFilters(device tc.DeviceName) types.Result[[]FilterInfo] {
	return a.adapter.GetFilters(device)
}
