package timeseries

import (
	"context"
	"time"

	"github.com/rng999/traffic-control-go/pkg/tc"
)

// TimeSeriesStore defines the interface for time-series statistics storage
type TimeSeriesStore interface {
	// Store saves statistics data with timestamp
	Store(ctx context.Context, deviceName tc.DeviceName, stats *TimeSeriesData) error
	
	// Query retrieves statistics data for a time range
	Query(ctx context.Context, deviceName tc.DeviceName, start, end time.Time) ([]*TimeSeriesData, error)
	
	// QueryAggregated retrieves aggregated statistics for a time range with specified interval
	QueryAggregated(ctx context.Context, deviceName tc.DeviceName, start, end time.Time, interval AggregationInterval) ([]*AggregatedData, error)
	
	// StoreAggregated stores aggregated data in the appropriate table
	StoreAggregated(ctx context.Context, deviceName tc.DeviceName, data *AggregatedData) error
	
	// Delete removes old data based on retention policy
	Delete(ctx context.Context, deviceName tc.DeviceName, before time.Time) error
	
	// Close closes the store and cleans up resources
	Close() error
}

// TimeSeriesData represents a single point in time statistics
type TimeSeriesData struct {
	Timestamp   time.Time          `json:"timestamp"`
	DeviceName  string             `json:"device_name"`
	QdiscStats  []QdiscDataPoint   `json:"qdisc_stats"`
	ClassStats  []ClassDataPoint   `json:"class_stats"`
	FilterStats []FilterDataPoint  `json:"filter_stats"`
	LinkStats   LinkDataPoint      `json:"link_stats"`
}

// QdiscDataPoint represents qdisc statistics at a point in time
type QdiscDataPoint struct {
	Handle       string    `json:"handle"`
	Type         string    `json:"type"`
	Bytes        uint64    `json:"bytes"`
	Packets      uint64    `json:"packets"`
	Drops        uint64    `json:"drops"`
	Overlimits   uint64    `json:"overlimits"`
	Requeues     uint64    `json:"requeues"`
	Backlog      uint64    `json:"backlog"`
	QueueLength  uint32    `json:"queue_length"`
	Rate         uint64    `json:"rate_bps"`         // bytes per second
	PacketRate   uint64    `json:"packet_rate_pps"`  // packets per second
}

// ClassDataPoint represents class statistics at a point in time
type ClassDataPoint struct {
	Handle       string    `json:"handle"`
	Parent       string    `json:"parent"`
	Type         string    `json:"type"`
	Bytes        uint64    `json:"bytes"`
	Packets      uint64    `json:"packets"`
	Drops        uint64    `json:"drops"`
	Overlimits   uint64    `json:"overlimits"`
	Requeues     uint64    `json:"requeues"`
	Backlog      uint64    `json:"backlog"`
	Rate         uint64    `json:"rate_bps"`
	PacketRate   uint64    `json:"packet_rate_pps"`
	// HTB-specific fields
	Lends        uint64    `json:"lends,omitempty"`
	Borrows      uint64    `json:"borrows,omitempty"`
	Giants       uint64    `json:"giants,omitempty"`
	Tokens       int64     `json:"tokens,omitempty"`
	CTokens      int64     `json:"ctokens,omitempty"`
}

// FilterDataPoint represents filter statistics at a point in time
type FilterDataPoint struct {
	Handle       string    `json:"handle"`
	Parent       string    `json:"parent"`
	Priority     uint16    `json:"priority"`
	FlowID       string    `json:"flow_id"`
	Matches      uint64    `json:"matches"`
	Bytes        uint64    `json:"bytes"`
	Packets      uint64    `json:"packets"`
	Rate         uint64    `json:"rate_bps"`
	PacketRate   uint64    `json:"packet_rate_pps"`
}

// LinkDataPoint represents link-level statistics at a point in time
type LinkDataPoint struct {
	RxBytes      uint64    `json:"rx_bytes"`
	TxBytes      uint64    `json:"tx_bytes"`
	RxPackets    uint64    `json:"rx_packets"`
	TxPackets    uint64    `json:"tx_packets"`
	RxErrors     uint64    `json:"rx_errors"`
	TxErrors     uint64    `json:"tx_errors"`
	RxDropped    uint64    `json:"rx_dropped"`
	TxDropped    uint64    `json:"tx_dropped"`
	RxRate       uint64    `json:"rx_rate_bps"`
	TxRate       uint64    `json:"tx_rate_bps"`
}

// AggregationInterval defines time intervals for data aggregation
type AggregationInterval string

const (
	IntervalMinute AggregationInterval = "1m"
	IntervalHour   AggregationInterval = "1h"
	IntervalDay    AggregationInterval = "1d"
	IntervalWeek   AggregationInterval = "1w"
	IntervalMonth  AggregationInterval = "1M"
)

// AggregatedData represents aggregated statistics over a time interval
type AggregatedData struct {
	Timestamp    time.Time                `json:"timestamp"`
	Interval     AggregationInterval      `json:"interval"`
	DeviceName   string                   `json:"device_name"`
	QdiscStats   []AggregatedQdiscStats   `json:"qdisc_stats"`
	ClassStats   []AggregatedClassStats   `json:"class_stats"`
	FilterStats  []AggregatedFilterStats  `json:"filter_stats"`
	LinkStats    AggregatedLinkStats      `json:"link_stats"`
}

// AggregatedQdiscStats represents aggregated qdisc statistics
type AggregatedQdiscStats struct {
	Handle           string    `json:"handle"`
	Type             string    `json:"type"`
	// Cumulative counters (sum)
	TotalBytes       uint64    `json:"total_bytes"`
	TotalPackets     uint64    `json:"total_packets"`
	TotalDrops       uint64    `json:"total_drops"`
	TotalOverlimits  uint64    `json:"total_overlimits"`
	// Rate statistics (average, min, max)
	AvgRate          uint64    `json:"avg_rate_bps"`
	MinRate          uint64    `json:"min_rate_bps"`
	MaxRate          uint64    `json:"max_rate_bps"`
	AvgPacketRate    uint64    `json:"avg_packet_rate_pps"`
	MinPacketRate    uint64    `json:"min_packet_rate_pps"`
	MaxPacketRate    uint64    `json:"max_packet_rate_pps"`
	// Queue statistics
	AvgBacklog       uint64    `json:"avg_backlog"`
	MaxBacklog       uint64    `json:"max_backlog"`
	AvgQueueLength   uint32    `json:"avg_queue_length"`
	MaxQueueLength   uint32    `json:"max_queue_length"`
}

// AggregatedClassStats represents aggregated class statistics
type AggregatedClassStats struct {
	Handle           string    `json:"handle"`
	Parent           string    `json:"parent"`
	Type             string    `json:"type"`
	// Cumulative counters
	TotalBytes       uint64    `json:"total_bytes"`
	TotalPackets     uint64    `json:"total_packets"`
	TotalDrops       uint64    `json:"total_drops"`
	TotalOverlimits  uint64    `json:"total_overlimits"`
	// Rate statistics
	AvgRate          uint64    `json:"avg_rate_bps"`
	MinRate          uint64    `json:"min_rate_bps"`
	MaxRate          uint64    `json:"max_rate_bps"`
	AvgPacketRate    uint64    `json:"avg_packet_rate_pps"`
	// Queue statistics
	AvgBacklog       uint64    `json:"avg_backlog"`
	MaxBacklog       uint64    `json:"max_backlog"`
	// HTB-specific aggregations
	TotalLends       uint64    `json:"total_lends,omitempty"`
	TotalBorrows     uint64    `json:"total_borrows,omitempty"`
	TotalGiants      uint64    `json:"total_giants,omitempty"`
}

// AggregatedFilterStats represents aggregated filter statistics
type AggregatedFilterStats struct {
	Handle           string    `json:"handle"`
	Parent           string    `json:"parent"`
	Priority         uint16    `json:"priority"`
	FlowID           string    `json:"flow_id"`
	TotalMatches     uint64    `json:"total_matches"`
	TotalBytes       uint64    `json:"total_bytes"`
	TotalPackets     uint64    `json:"total_packets"`
	AvgRate          uint64    `json:"avg_rate_bps"`
	MaxRate          uint64    `json:"max_rate_bps"`
	AvgPacketRate    uint64    `json:"avg_packet_rate_pps"`
	MaxPacketRate    uint64    `json:"max_packet_rate_pps"`
}

// AggregatedLinkStats represents aggregated link statistics
type AggregatedLinkStats struct {
	TotalRxBytes     uint64    `json:"total_rx_bytes"`
	TotalTxBytes     uint64    `json:"total_tx_bytes"`
	TotalRxPackets   uint64    `json:"total_rx_packets"`
	TotalTxPackets   uint64    `json:"total_tx_packets"`
	TotalRxErrors    uint64    `json:"total_rx_errors"`
	TotalTxErrors    uint64    `json:"total_tx_errors"`
	TotalRxDropped   uint64    `json:"total_rx_dropped"`
	TotalTxDropped   uint64    `json:"total_tx_dropped"`
	AvgRxRate        uint64    `json:"avg_rx_rate_bps"`
	MaxRxRate        uint64    `json:"max_rx_rate_bps"`
	AvgTxRate        uint64    `json:"avg_tx_rate_bps"`
	MaxTxRate        uint64    `json:"max_tx_rate_bps"`
}

// RetentionPolicy defines how long to keep time-series data
type RetentionPolicy struct {
	RawDataRetention        time.Duration // How long to keep raw data points
	MinuteAggRetention      time.Duration // How long to keep minute aggregations
	HourAggRetention        time.Duration // How long to keep hour aggregations
	DayAggRetention         time.Duration // How long to keep day aggregations
	WeekAggRetention        time.Duration // How long to keep week aggregations
	MonthAggRetention       time.Duration // How long to keep month aggregations
}

// DefaultRetentionPolicy returns sensible default retention periods
func DefaultRetentionPolicy() RetentionPolicy {
	return RetentionPolicy{
		RawDataRetention:        24 * time.Hour,        // 1 day of raw data
		MinuteAggRetention:      7 * 24 * time.Hour,    // 1 week of minute data
		HourAggRetention:        30 * 24 * time.Hour,   // 1 month of hour data
		DayAggRetention:         365 * 24 * time.Hour,  // 1 year of day data
		WeekAggRetention:        2 * 365 * 24 * time.Hour, // 2 years of week data
		MonthAggRetention:       5 * 365 * 24 * time.Hour, // 5 years of month data
	}
}