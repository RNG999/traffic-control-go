package timeseries

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/rng999/traffic-control-go/pkg/tc"
)

// MemoryTimeSeriesStore implements TimeSeriesStore using in-memory storage
// This implementation is primarily for testing and development
type MemoryTimeSeriesStore struct {
	mu               sync.RWMutex
	data             map[string][]*TimeSeriesData     // deviceName -> data points
	aggregatedData   map[string]map[AggregationInterval][]*AggregatedData // deviceName -> interval -> data points
}

// NewMemoryTimeSeriesStore creates a new in-memory time-series store
func NewMemoryTimeSeriesStore() *MemoryTimeSeriesStore {
	return &MemoryTimeSeriesStore{
		data:           make(map[string][]*TimeSeriesData),
		aggregatedData: make(map[string]map[AggregationInterval][]*AggregatedData),
	}
}

// Store saves statistics data with timestamp
func (m *MemoryTimeSeriesStore) Store(ctx context.Context, deviceName tc.DeviceName, stats *TimeSeriesData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	deviceKey := deviceName.String()
	
	// Create a copy of the data to avoid mutation issues
	statsCopy := &TimeSeriesData{
		Timestamp:   stats.Timestamp,
		DeviceName:  stats.DeviceName,
		QdiscStats:  make([]QdiscDataPoint, len(stats.QdiscStats)),
		ClassStats:  make([]ClassDataPoint, len(stats.ClassStats)),
		FilterStats: make([]FilterDataPoint, len(stats.FilterStats)),
		LinkStats:   stats.LinkStats,
	}

	copy(statsCopy.QdiscStats, stats.QdiscStats)
	copy(statsCopy.ClassStats, stats.ClassStats)
	copy(statsCopy.FilterStats, stats.FilterStats)

	// Add to device data
	if m.data[deviceKey] == nil {
		m.data[deviceKey] = make([]*TimeSeriesData, 0)
	}
	
	m.data[deviceKey] = append(m.data[deviceKey], statsCopy)
	
	// Keep data sorted by timestamp
	sort.Slice(m.data[deviceKey], func(i, j int) bool {
		return m.data[deviceKey][i].Timestamp.Before(m.data[deviceKey][j].Timestamp)
	})

	return nil
}

// Query retrieves statistics data for a time range
func (m *MemoryTimeSeriesStore) Query(ctx context.Context, deviceName tc.DeviceName, start, end time.Time) ([]*TimeSeriesData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	deviceKey := deviceName.String()
	deviceData := m.data[deviceKey]
	if deviceData == nil {
		return []*TimeSeriesData{}, nil
	}

	var results []*TimeSeriesData
	for _, data := range deviceData {
		if (data.Timestamp.Equal(start) || data.Timestamp.After(start)) &&
		   (data.Timestamp.Equal(end) || data.Timestamp.Before(end)) {
			// Create a copy to avoid mutation
			dataCopy := &TimeSeriesData{
				Timestamp:   data.Timestamp,
				DeviceName:  data.DeviceName,
				QdiscStats:  make([]QdiscDataPoint, len(data.QdiscStats)),
				ClassStats:  make([]ClassDataPoint, len(data.ClassStats)),
				FilterStats: make([]FilterDataPoint, len(data.FilterStats)),
				LinkStats:   data.LinkStats,
			}
			copy(dataCopy.QdiscStats, data.QdiscStats)
			copy(dataCopy.ClassStats, data.ClassStats)
			copy(dataCopy.FilterStats, data.FilterStats)
			
			results = append(results, dataCopy)
		}
	}

	return results, nil
}

// QueryAggregated retrieves aggregated statistics for a time range with specified interval
func (m *MemoryTimeSeriesStore) QueryAggregated(ctx context.Context, deviceName tc.DeviceName, start, end time.Time, interval AggregationInterval) ([]*AggregatedData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	deviceKey := deviceName.String()
	
	if m.aggregatedData[deviceKey] == nil {
		return []*AggregatedData{}, nil
	}
	
	intervalData := m.aggregatedData[deviceKey][interval]
	if intervalData == nil {
		return []*AggregatedData{}, nil
	}

	var results []*AggregatedData
	for _, data := range intervalData {
		if (data.Timestamp.Equal(start) || data.Timestamp.After(start)) &&
		   (data.Timestamp.Equal(end) || data.Timestamp.Before(end)) {
			// Create a copy to avoid mutation
			dataCopy := &AggregatedData{
				Timestamp:   data.Timestamp,
				Interval:    data.Interval,
				DeviceName:  data.DeviceName,
				QdiscStats:  make([]AggregatedQdiscStats, len(data.QdiscStats)),
				ClassStats:  make([]AggregatedClassStats, len(data.ClassStats)),
				FilterStats: make([]AggregatedFilterStats, len(data.FilterStats)),
				LinkStats:   data.LinkStats,
			}
			copy(dataCopy.QdiscStats, data.QdiscStats)
			copy(dataCopy.ClassStats, data.ClassStats)
			copy(dataCopy.FilterStats, data.FilterStats)
			
			results = append(results, dataCopy)
		}
	}

	return results, nil
}

// Delete removes old data based on retention policy
func (m *MemoryTimeSeriesStore) Delete(ctx context.Context, deviceName tc.DeviceName, before time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	deviceKey := deviceName.String()

	// Delete raw data
	if deviceData := m.data[deviceKey]; deviceData != nil {
		var filtered []*TimeSeriesData
		for _, data := range deviceData {
			if data.Timestamp.After(before) || data.Timestamp.Equal(before) {
				filtered = append(filtered, data)
			}
		}
		m.data[deviceKey] = filtered
	}

	// Delete aggregated data
	if deviceAggData := m.aggregatedData[deviceKey]; deviceAggData != nil {
		for interval, intervalData := range deviceAggData {
			var filtered []*AggregatedData
			for _, data := range intervalData {
				if data.Timestamp.After(before) || data.Timestamp.Equal(before) {
					filtered = append(filtered, data)
				}
			}
			m.aggregatedData[deviceKey][interval] = filtered
		}
	}

	return nil
}

// StoreAggregated stores aggregated data
func (m *MemoryTimeSeriesStore) StoreAggregated(ctx context.Context, deviceName tc.DeviceName, data *AggregatedData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	deviceKey := deviceName.String()
	
	// Initialize device aggregated data if needed
	if m.aggregatedData[deviceKey] == nil {
		m.aggregatedData[deviceKey] = make(map[AggregationInterval][]*AggregatedData)
	}
	
	// Initialize interval data if needed
	if m.aggregatedData[deviceKey][data.Interval] == nil {
		m.aggregatedData[deviceKey][data.Interval] = make([]*AggregatedData, 0)
	}

	// Create a copy of the data
	dataCopy := &AggregatedData{
		Timestamp:   data.Timestamp,
		Interval:    data.Interval,
		DeviceName:  data.DeviceName,
		QdiscStats:  make([]AggregatedQdiscStats, len(data.QdiscStats)),
		ClassStats:  make([]AggregatedClassStats, len(data.ClassStats)),
		FilterStats: make([]AggregatedFilterStats, len(data.FilterStats)),
		LinkStats:   data.LinkStats,
	}
	copy(dataCopy.QdiscStats, data.QdiscStats)
	copy(dataCopy.ClassStats, data.ClassStats)
	copy(dataCopy.FilterStats, data.FilterStats)

	// Check if data already exists for this timestamp (replace if exists)
	intervalData := m.aggregatedData[deviceKey][data.Interval]
	replaced := false
	for i, existing := range intervalData {
		if existing.Timestamp.Equal(data.Timestamp) {
			intervalData[i] = dataCopy
			replaced = true
			break
		}
	}

	// If not replaced, append new data
	if !replaced {
		m.aggregatedData[deviceKey][data.Interval] = append(intervalData, dataCopy)
		
		// Keep data sorted by timestamp
		sort.Slice(m.aggregatedData[deviceKey][data.Interval], func(i, j int) bool {
			return m.aggregatedData[deviceKey][data.Interval][i].Timestamp.Before(
				m.aggregatedData[deviceKey][data.Interval][j].Timestamp)
		})
	}

	return nil
}

// Close closes the store and cleans up resources
func (m *MemoryTimeSeriesStore) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Clear all data
	m.data = make(map[string][]*TimeSeriesData)
	m.aggregatedData = make(map[string]map[AggregationInterval][]*AggregatedData)
	
	return nil
}

// GetStorageStats returns memory usage statistics
func (m *MemoryTimeSeriesStore) GetStorageStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})
	
	// Count raw data points per device
	rawDataCount := 0
	for device, data := range m.data {
		deviceCount := len(data)
		rawDataCount += deviceCount
		stats["raw_data_"+device] = deviceCount
	}
	stats["total_raw_data_points"] = rawDataCount

	// Count aggregated data points per device and interval
	aggregatedDataCount := 0
	for device, deviceData := range m.aggregatedData {
		deviceTotal := 0
		for interval, intervalData := range deviceData {
			intervalCount := len(intervalData)
			deviceTotal += intervalCount
			aggregatedDataCount += intervalCount
			stats["aggregated_data_"+device+"_"+string(interval)] = intervalCount
		}
		stats["aggregated_data_"+device+"_total"] = deviceTotal
	}
	stats["total_aggregated_data_points"] = aggregatedDataCount

	stats["total_devices"] = len(m.data)
	
	return stats
}

// GetDataRange returns the time range of stored data for a device
func (m *MemoryTimeSeriesStore) GetDataRange(deviceName tc.DeviceName) (start, end time.Time, count int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	deviceKey := deviceName.String()
	deviceData := m.data[deviceKey]
	
	if len(deviceData) == 0 {
		return time.Time{}, time.Time{}, 0
	}

	// Data is sorted by timestamp
	start = deviceData[0].Timestamp
	end = deviceData[len(deviceData)-1].Timestamp
	count = len(deviceData)

	return start, end, count
}