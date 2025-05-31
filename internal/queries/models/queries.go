package models

import (
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
)

// Query is the base interface for all queries
type Query interface {
	DeviceName() valueobjects.DeviceName
}

// GetQdiscByDeviceQuery retrieves all qdiscs for a device
type GetQdiscByDeviceQuery struct {
	deviceName valueobjects.DeviceName
}

// NewGetQdiscByDeviceQuery creates a new query
func NewGetQdiscByDeviceQuery(deviceName valueobjects.DeviceName) *GetQdiscByDeviceQuery {
	return &GetQdiscByDeviceQuery{
		deviceName: deviceName,
	}
}

// DeviceName returns the device name
func (q *GetQdiscByDeviceQuery) DeviceName() valueobjects.DeviceName {
	return q.deviceName
}

// GetClassesByDeviceQuery retrieves all classes for a device
type GetClassesByDeviceQuery struct {
	deviceName valueobjects.DeviceName
}

// NewGetClassesByDeviceQuery creates a new query
func NewGetClassesByDeviceQuery(deviceName valueobjects.DeviceName) *GetClassesByDeviceQuery {
	return &GetClassesByDeviceQuery{
		deviceName: deviceName,
	}
}

// DeviceName returns the device name
func (q *GetClassesByDeviceQuery) DeviceName() valueobjects.DeviceName {
	return q.deviceName
}

// GetFiltersByDeviceQuery retrieves all filters for a device
type GetFiltersByDeviceQuery struct {
	deviceName valueobjects.DeviceName
}

// NewGetFiltersByDeviceQuery creates a new query
func NewGetFiltersByDeviceQuery(deviceName valueobjects.DeviceName) *GetFiltersByDeviceQuery {
	return &GetFiltersByDeviceQuery{
		deviceName: deviceName,
	}
}

// DeviceName returns the device name
func (q *GetFiltersByDeviceQuery) DeviceName() valueobjects.DeviceName {
	return q.deviceName
}

// GetTrafficControlConfigQuery retrieves the complete TC configuration
type GetTrafficControlConfigQuery struct {
	deviceName valueobjects.DeviceName
}

// NewGetTrafficControlConfigQuery creates a new query
func NewGetTrafficControlConfigQuery(deviceName valueobjects.DeviceName) *GetTrafficControlConfigQuery {
	return &GetTrafficControlConfigQuery{
		deviceName: deviceName,
	}
}

// DeviceName returns the device name
func (q *GetTrafficControlConfigQuery) DeviceName() valueobjects.DeviceName {
	return q.deviceName
}