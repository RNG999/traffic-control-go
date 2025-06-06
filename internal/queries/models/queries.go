package models

import (
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// Query is the base interface for all queries
type Query interface {
	DeviceName() tc.DeviceName
}

// GetQdiscByDeviceQuery retrieves all qdiscs for a device
type GetQdiscByDeviceQuery struct {
	deviceName tc.DeviceName
}

// NewGetQdiscByDeviceQuery creates a new query
func NewGetQdiscByDeviceQuery(deviceName tc.DeviceName) *GetQdiscByDeviceQuery {
	return &GetQdiscByDeviceQuery{
		deviceName: deviceName,
	}
}

// DeviceName returns the device name
func (q *GetQdiscByDeviceQuery) DeviceName() tc.DeviceName {
	return q.deviceName
}

// GetClassesByDeviceQuery retrieves all classes for a device
type GetClassesByDeviceQuery struct {
	deviceName tc.DeviceName
}

// NewGetClassesByDeviceQuery creates a new query
func NewGetClassesByDeviceQuery(deviceName tc.DeviceName) *GetClassesByDeviceQuery {
	return &GetClassesByDeviceQuery{
		deviceName: deviceName,
	}
}

// DeviceName returns the device name
func (q *GetClassesByDeviceQuery) DeviceName() tc.DeviceName {
	return q.deviceName
}

// GetFiltersByDeviceQuery retrieves all filters for a device
type GetFiltersByDeviceQuery struct {
	deviceName tc.DeviceName
}

// NewGetFiltersByDeviceQuery creates a new query
func NewGetFiltersByDeviceQuery(deviceName tc.DeviceName) *GetFiltersByDeviceQuery {
	return &GetFiltersByDeviceQuery{
		deviceName: deviceName,
	}
}

// DeviceName returns the device name
func (q *GetFiltersByDeviceQuery) DeviceName() tc.DeviceName {
	return q.deviceName
}

// GetTrafficControlConfigQuery retrieves the complete TC configuration
type GetTrafficControlConfigQuery struct {
	deviceName tc.DeviceName
}

// NewGetTrafficControlConfigQuery creates a new query
func NewGetTrafficControlConfigQuery(deviceName tc.DeviceName) *GetTrafficControlConfigQuery {
	return &GetTrafficControlConfigQuery{
		deviceName: deviceName,
	}
}

// DeviceName returns the device name
func (q *GetTrafficControlConfigQuery) DeviceName() tc.DeviceName {
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
	deviceName tc.DeviceName
}

// NewGetDeviceStatisticsQuery creates a new query
func NewGetDeviceStatisticsQuery(deviceName tc.DeviceName) *GetDeviceStatisticsQuery {
	return &GetDeviceStatisticsQuery{
		deviceName: deviceName,
	}
}

// DeviceName returns the device name
func (q *GetDeviceStatisticsQuery) DeviceName() tc.DeviceName {
	return q.deviceName
}

// GetQdiscStatisticsQuery queries for qdisc statistics
type GetQdiscStatisticsQuery struct {
	deviceName tc.DeviceName
	handle     tc.Handle
}

// NewGetQdiscStatisticsQuery creates a new query
func NewGetQdiscStatisticsQuery(deviceName tc.DeviceName, handle tc.Handle) *GetQdiscStatisticsQuery {
	return &GetQdiscStatisticsQuery{
		deviceName: deviceName,
		handle:     handle,
	}
}

// DeviceName returns the device name
func (q *GetQdiscStatisticsQuery) DeviceName() tc.DeviceName {
	return q.deviceName
}

// Handle returns the qdisc handle
func (q *GetQdiscStatisticsQuery) Handle() tc.Handle {
	return q.handle
}

// GetClassStatisticsQuery queries for class statistics
type GetClassStatisticsQuery struct {
	deviceName tc.DeviceName
	handle     tc.Handle
}

// NewGetClassStatisticsQuery creates a new query
func NewGetClassStatisticsQuery(deviceName tc.DeviceName, handle tc.Handle) *GetClassStatisticsQuery {
	return &GetClassStatisticsQuery{
		deviceName: deviceName,
		handle:     handle,
	}
}

// DeviceName returns the device name
func (q *GetClassStatisticsQuery) DeviceName() tc.DeviceName {
	return q.deviceName
}

// Handle returns the class handle
func (q *GetClassStatisticsQuery) Handle() tc.Handle {
	return q.handle
}

// GetRealtimeStatisticsQuery queries for realtime statistics
type GetRealtimeStatisticsQuery struct {
	deviceName tc.DeviceName
}

// NewGetRealtimeStatisticsQuery creates a new query
func NewGetRealtimeStatisticsQuery(deviceName tc.DeviceName) *GetRealtimeStatisticsQuery {
	return &GetRealtimeStatisticsQuery{
		deviceName: deviceName,
	}
}

// DeviceName returns the device name
func (q *GetRealtimeStatisticsQuery) DeviceName() tc.DeviceName {
	return q.deviceName
}
