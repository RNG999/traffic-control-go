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

// GetQdiscQuery queries for a specific qdisc
type GetQdiscQuery struct {
	DeviceName string
	Handle     string
}

// GetClassQuery queries for a specific class
type GetClassQuery struct {
	DeviceName string
	ClassID    string
}

// GetFilterQuery queries for a specific filter
type GetFilterQuery struct {
	DeviceName string
	Parent     string
	Priority   uint16
	Handle     string
}

// GetConfigurationQuery queries for the complete configuration
type GetConfigurationQuery struct {
	DeviceName string
}

// GetDeviceStatisticsQuery queries for device statistics
type GetDeviceStatisticsQuery struct {
	deviceName valueobjects.DeviceName
}

// NewGetDeviceStatisticsQuery creates a new query
func NewGetDeviceStatisticsQuery(deviceName valueobjects.DeviceName) *GetDeviceStatisticsQuery {
	return &GetDeviceStatisticsQuery{
		deviceName: deviceName,
	}
}

// DeviceName returns the device name
func (q *GetDeviceStatisticsQuery) DeviceName() valueobjects.DeviceName {
	return q.deviceName
}

// GetQdiscStatisticsQuery queries for qdisc statistics
type GetQdiscStatisticsQuery struct {
	deviceName valueobjects.DeviceName
	handle     valueobjects.Handle
}

// NewGetQdiscStatisticsQuery creates a new query
func NewGetQdiscStatisticsQuery(deviceName valueobjects.DeviceName, handle valueobjects.Handle) *GetQdiscStatisticsQuery {
	return &GetQdiscStatisticsQuery{
		deviceName: deviceName,
		handle:     handle,
	}
}

// DeviceName returns the device name
func (q *GetQdiscStatisticsQuery) DeviceName() valueobjects.DeviceName {
	return q.deviceName
}

// Handle returns the qdisc handle
func (q *GetQdiscStatisticsQuery) Handle() valueobjects.Handle {
	return q.handle
}

// GetClassStatisticsQuery queries for class statistics
type GetClassStatisticsQuery struct {
	deviceName valueobjects.DeviceName
	handle     valueobjects.Handle
}

// NewGetClassStatisticsQuery creates a new query
func NewGetClassStatisticsQuery(deviceName valueobjects.DeviceName, handle valueobjects.Handle) *GetClassStatisticsQuery {
	return &GetClassStatisticsQuery{
		deviceName: deviceName,
		handle:     handle,
	}
}

// DeviceName returns the device name
func (q *GetClassStatisticsQuery) DeviceName() valueobjects.DeviceName {
	return q.deviceName
}

// Handle returns the class handle
func (q *GetClassStatisticsQuery) Handle() valueobjects.Handle {
	return q.handle
}

// GetRealtimeStatisticsQuery queries for realtime statistics
type GetRealtimeStatisticsQuery struct {
	deviceName valueobjects.DeviceName
}

// NewGetRealtimeStatisticsQuery creates a new query
func NewGetRealtimeStatisticsQuery(deviceName valueobjects.DeviceName) *GetRealtimeStatisticsQuery {
	return &GetRealtimeStatisticsQuery{
		deviceName: deviceName,
	}
}

// DeviceName returns the device name
func (q *GetRealtimeStatisticsQuery) DeviceName() valueobjects.DeviceName {
	return q.deviceName
}
