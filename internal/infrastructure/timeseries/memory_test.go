package timeseries

import (
	"context"
	"testing"
	"time"

	"github.com/rng999/traffic-control-go/pkg/tc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryTimeSeriesStore(t *testing.T) {
	store := NewMemoryTimeSeriesStore()
	defer store.Close()

	ctx := context.Background()
	deviceName, _ := tc.NewDeviceName("eth0")

	t.Run("store and query time-series data", func(t *testing.T) {
		// Create test data
		now := time.Now().Truncate(time.Second)
		testData := &TimeSeriesData{
			Timestamp:  now,
			DeviceName: deviceName.String(),
			QdiscStats: []QdiscDataPoint{
				{
					Handle:      "1:",
					Type:        "htb",
					Bytes:       1000000,
					Packets:     1000,
					Drops:       5,
					Overlimits:  10,
					Rate:        100000,
					PacketRate:  100,
				},
			},
			ClassStats: []ClassDataPoint{
				{
					Handle:     "1:10",
					Parent:     "1:",
					Type:       "htb",
					Bytes:      500000,
					Packets:    500,
					Drops:      2,
					Rate:       50000,
					PacketRate: 50,
				},
			},
			FilterStats: []FilterDataPoint{
				{
					Handle:     "800:100",
					Parent:     "1:",
					Priority:   100,
					FlowID:     "1:10",
					Matches:    250,
					Bytes:      250000,
					Packets:    250,
					Rate:       25000,
					PacketRate: 25,
				},
			},
			LinkStats: LinkDataPoint{
				RxBytes:   2000000,
				TxBytes:   1500000,
				RxPackets: 2000,
				TxPackets: 1500,
				RxRate:    200000,
				TxRate:    150000,
			},
		}

		// Store the data
		err := store.Store(ctx, deviceName, testData)
		require.NoError(t, err)

		// Query the data
		start := now.Add(-1 * time.Hour)
		end := now.Add(1 * time.Hour)
		results, err := store.Query(ctx, deviceName, start, end)
		require.NoError(t, err)
		require.Len(t, results, 1)

		// Verify the data
		result := results[0]
		assert.Equal(t, testData.DeviceName, result.DeviceName)
		assert.Equal(t, testData.Timestamp, result.Timestamp)
		assert.Len(t, result.QdiscStats, 1)
		assert.Equal(t, testData.QdiscStats[0].Handle, result.QdiscStats[0].Handle)
		assert.Equal(t, testData.QdiscStats[0].Bytes, result.QdiscStats[0].Bytes)
		assert.Len(t, result.ClassStats, 1)
		assert.Equal(t, testData.ClassStats[0].Handle, result.ClassStats[0].Handle)
		assert.Len(t, result.FilterStats, 1)
		assert.Equal(t, testData.FilterStats[0].Handle, result.FilterStats[0].Handle)
		assert.Equal(t, testData.LinkStats.RxBytes, result.LinkStats.RxBytes)
	})

	t.Run("store multiple data points and verify ordering", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		
		// Store data points in non-chronological order
		times := []time.Time{
			now.Add(2 * time.Minute),
			now,
			now.Add(1 * time.Minute),
		}

		for i, timestamp := range times {
			data := &TimeSeriesData{
				Timestamp:   timestamp,
				DeviceName:  deviceName.String(),
				QdiscStats:  []QdiscDataPoint{{Handle: "1:", Bytes: uint64(1000 * (i + 1))}},
				ClassStats:  []ClassDataPoint{},
				FilterStats: []FilterDataPoint{},
				LinkStats:   LinkDataPoint{},
			}
			err := store.Store(ctx, deviceName, data)
			require.NoError(t, err)
		}

		// Query all data
		start := now.Add(-1 * time.Hour)
		end := now.Add(1 * time.Hour)
		results, err := store.Query(ctx, deviceName, start, end)
		require.NoError(t, err)
		require.Len(t, results, 4) // 3 new + 1 from previous test

		// Verify chronological ordering (should be sorted by timestamp)
		for i := 1; i < len(results); i++ {
			assert.True(t, results[i-1].Timestamp.Before(results[i].Timestamp) || 
						 results[i-1].Timestamp.Equal(results[i].Timestamp),
				"Data should be sorted chronologically")
		}
	})

	t.Run("query empty time range", func(t *testing.T) {
		// Query a time range with no data
		start := time.Now().Add(-10 * time.Hour)
		end := time.Now().Add(-9 * time.Hour)
		results, err := store.Query(ctx, deviceName, start, end)
		require.NoError(t, err)
		assert.Len(t, results, 0)
	})

	t.Run("store and query aggregated data", func(t *testing.T) {
		// Create test aggregated data
		now := time.Now().Truncate(time.Hour) // Round to hour for aggregation
		testAggData := &AggregatedData{
			Timestamp:  now,
			Interval:   IntervalHour,
			DeviceName: deviceName.String(),
			QdiscStats: []AggregatedQdiscStats{
				{
					Handle:          "1:",
					Type:            "htb",
					TotalBytes:      10000000,
					TotalPackets:    10000,
					TotalDrops:      50,
					TotalOverlimits: 100,
					AvgRate:         100000,
					MaxRate:         200000,
					AvgPacketRate:   100,
					MaxPacketRate:   200,
				},
			},
			ClassStats: []AggregatedClassStats{
				{
					Handle:        "1:10",
					Parent:        "1:",
					Type:          "htb",
					TotalBytes:    5000000,
					TotalPackets:  5000,
					TotalDrops:    25,
					AvgRate:       50000,
					MaxRate:       100000,
					AvgPacketRate: 50,
				},
			},
			FilterStats: []AggregatedFilterStats{
				{
					Handle:        "800:100",
					Parent:        "1:",
					Priority:      100,
					FlowID:        "1:10",
					TotalMatches:  2500,
					TotalBytes:    2500000,
					TotalPackets:  2500,
					AvgRate:       25000,
					MaxRate:       50000,
					AvgPacketRate: 25,
					MaxPacketRate: 50,
				},
			},
			LinkStats: AggregatedLinkStats{
				TotalRxBytes:   20000000,
				TotalTxBytes:   15000000,
				TotalRxPackets: 20000,
				TotalTxPackets: 15000,
				AvgRxRate:      200000,
				MaxRxRate:      400000,
				AvgTxRate:      150000,
				MaxTxRate:      300000,
			},
		}

		// Store the aggregated data
		err := store.StoreAggregated(ctx, deviceName, testAggData)
		require.NoError(t, err)

		// Query the aggregated data
		start := now.Add(-1 * time.Hour)
		end := now.Add(1 * time.Hour)
		results, err := store.QueryAggregated(ctx, deviceName, start, end, IntervalHour)
		require.NoError(t, err)
		require.Len(t, results, 1)

		// Verify the data
		result := results[0]
		assert.Equal(t, testAggData.DeviceName, result.DeviceName)
		assert.Equal(t, testAggData.Interval, result.Interval)
		assert.Equal(t, testAggData.Timestamp, result.Timestamp)
		assert.Len(t, result.QdiscStats, 1)
		assert.Equal(t, testAggData.QdiscStats[0].Handle, result.QdiscStats[0].Handle)
		assert.Equal(t, testAggData.QdiscStats[0].TotalBytes, result.QdiscStats[0].TotalBytes)
	})

	t.Run("replace existing aggregated data", func(t *testing.T) {
		now := time.Now().Truncate(time.Hour)
		
		// Store initial aggregated data
		initialData := &AggregatedData{
			Timestamp:  now,
			Interval:   IntervalHour,
			DeviceName: deviceName.String(),
			QdiscStats: []AggregatedQdiscStats{
				{Handle: "1:", TotalBytes: 1000000},
			},
			ClassStats:  []AggregatedClassStats{},
			FilterStats: []AggregatedFilterStats{},
			LinkStats:   AggregatedLinkStats{TotalRxBytes: 2000000},
		}
		err := store.StoreAggregated(ctx, deviceName, initialData)
		require.NoError(t, err)

		// Store updated data for same timestamp (should replace)
		updatedData := &AggregatedData{
			Timestamp:  now,
			Interval:   IntervalHour,
			DeviceName: deviceName.String(),
			QdiscStats: []AggregatedQdiscStats{
				{Handle: "1:", TotalBytes: 2000000}, // Updated value
			},
			ClassStats:  []AggregatedClassStats{},
			FilterStats: []AggregatedFilterStats{},
			LinkStats:   AggregatedLinkStats{TotalRxBytes: 4000000}, // Updated value
		}
		err = store.StoreAggregated(ctx, deviceName, updatedData)
		require.NoError(t, err)

		// Query and verify updated data
		start := now.Add(-1 * time.Hour)
		end := now.Add(1 * time.Hour)
		results, err := store.QueryAggregated(ctx, deviceName, start, end, IntervalHour)
		require.NoError(t, err)
		require.Len(t, results, 1) // Should still be only 1 entry (replaced, not added)

		result := results[0]
		assert.Equal(t, uint64(2000000), result.QdiscStats[0].TotalBytes)
		assert.Equal(t, uint64(4000000), result.LinkStats.TotalRxBytes)
	})

	t.Run("delete old data", func(t *testing.T) {
		// Store some old data
		oldTime := time.Now().Add(-48 * time.Hour)
		oldData := &TimeSeriesData{
			Timestamp:   oldTime,
			DeviceName:  deviceName.String(),
			QdiscStats:  []QdiscDataPoint{{Handle: "1:", Type: "htb", Bytes: 1000}},
			ClassStats:  []ClassDataPoint{},
			FilterStats: []FilterDataPoint{},
			LinkStats:   LinkDataPoint{},
		}

		err := store.Store(ctx, deviceName, oldData)
		require.NoError(t, err)

		// Verify old data exists
		start := oldTime.Add(-1 * time.Hour)
		end := oldTime.Add(1 * time.Hour)
		results, err := store.Query(ctx, deviceName, start, end)
		require.NoError(t, err)
		assert.Len(t, results, 1)

		// Delete old data (older than 24 hours)
		cutoff := time.Now().Add(-24 * time.Hour)
		err = store.Delete(ctx, deviceName, cutoff)
		require.NoError(t, err)

		// Verify old data is deleted
		results, err = store.Query(ctx, deviceName, start, end)
		require.NoError(t, err)
		assert.Len(t, results, 0)
	})

	t.Run("storage stats", func(t *testing.T) {
		stats := store.GetStorageStats()
		assert.Contains(t, stats, "total_raw_data_points")
		assert.Contains(t, stats, "total_aggregated_data_points")
		assert.Contains(t, stats, "total_devices")
		assert.IsType(t, 0, stats["total_raw_data_points"])
	})

	t.Run("data range", func(t *testing.T) {
		// Clear store for clean test
		newStore := NewMemoryTimeSeriesStore()
		defer newStore.Close()
		
		now := time.Now().Truncate(time.Second)
		
		// Store data points at different times
		times := []time.Time{
			now,
			now.Add(1 * time.Hour),
			now.Add(2 * time.Hour),
		}

		for _, timestamp := range times {
			data := &TimeSeriesData{
				Timestamp:   timestamp,
				DeviceName:  deviceName.String(),
				QdiscStats:  []QdiscDataPoint{{Handle: "1:", Bytes: 1000}},
				ClassStats:  []ClassDataPoint{},
				FilterStats: []FilterDataPoint{},
				LinkStats:   LinkDataPoint{},
			}
			err := newStore.Store(ctx, deviceName, data)
			require.NoError(t, err)
		}

		// Get data range
		start, end, count := newStore.GetDataRange(deviceName)
		assert.Equal(t, now, start)
		assert.Equal(t, now.Add(2*time.Hour), end)
		assert.Equal(t, 3, count)

		// Test empty device
		nonexistentDevice, _ := tc.NewDeviceName("nonexistent")
		emptyStart, emptyEnd, emptyCount := newStore.GetDataRange(nonexistentDevice)
		assert.True(t, emptyStart.IsZero())
		assert.True(t, emptyEnd.IsZero())
		assert.Equal(t, 0, emptyCount)
	})
}

func TestMemoryTimeSeriesStore_MultipleDevices(t *testing.T) {
	store := NewMemoryTimeSeriesStore()
	defer store.Close()

	ctx := context.Background()
	device1, _ := tc.NewDeviceName("eth0")
	device2, _ := tc.NewDeviceName("eth1")

	now := time.Now().Truncate(time.Second)

	// Store data for device1
	data1 := &TimeSeriesData{
		Timestamp:   now,
		DeviceName:  device1.String(),
		QdiscStats:  []QdiscDataPoint{{Handle: "1:", Type: "htb", Bytes: 1000}},
		ClassStats:  []ClassDataPoint{},
		FilterStats: []FilterDataPoint{},
		LinkStats:   LinkDataPoint{RxBytes: 2000},
	}
	err := store.Store(ctx, device1, data1)
	require.NoError(t, err)

	// Store data for device2
	data2 := &TimeSeriesData{
		Timestamp:   now.Add(1 * time.Minute),
		DeviceName:  device2.String(),
		QdiscStats:  []QdiscDataPoint{{Handle: "1:", Type: "htb", Bytes: 2000}},
		ClassStats:  []ClassDataPoint{},
		FilterStats: []FilterDataPoint{},
		LinkStats:   LinkDataPoint{RxBytes: 4000},
	}
	err = store.Store(ctx, device2, data2)
	require.NoError(t, err)

	// Query device1 data
	start := now.Add(-1 * time.Hour)
	end := now.Add(1 * time.Hour)
	results1, err := store.Query(ctx, device1, start, end)
	require.NoError(t, err)
	require.Len(t, results1, 1)
	assert.Equal(t, device1.String(), results1[0].DeviceName)
	assert.Equal(t, uint64(1000), results1[0].QdiscStats[0].Bytes)

	// Query device2 data
	results2, err := store.Query(ctx, device2, start, end)
	require.NoError(t, err)
	require.Len(t, results2, 1)
	assert.Equal(t, device2.String(), results2[0].DeviceName)
	assert.Equal(t, uint64(2000), results2[0].QdiscStats[0].Bytes)

	// Verify devices are isolated
	assert.NotEqual(t, results1[0].DeviceName, results2[0].DeviceName)
}

func TestMemoryTimeSeriesStore_ConcurrentAccess(t *testing.T) {
	store := NewMemoryTimeSeriesStore()
	defer store.Close()

	ctx := context.Background()
	deviceName, _ := tc.NewDeviceName("eth0")
	now := time.Now().Truncate(time.Second)

	// Test concurrent writes and reads
	done := make(chan bool, 2)

	// Writer goroutine
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			data := &TimeSeriesData{
				Timestamp:   now.Add(time.Duration(i) * time.Second),
				DeviceName:  deviceName.String(),
				QdiscStats:  []QdiscDataPoint{{Handle: "1:", Bytes: uint64(i * 1000)}},
				ClassStats:  []ClassDataPoint{},
				FilterStats: []FilterDataPoint{},
				LinkStats:   LinkDataPoint{},
			}
			err := store.Store(ctx, deviceName, data)
			assert.NoError(t, err)
		}
	}()

	// Reader goroutine
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 10; i++ {
			start := now.Add(-1 * time.Hour)
			end := now.Add(1 * time.Hour)
			_, err := store.Query(ctx, deviceName, start, end)
			assert.NoError(t, err)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify final state
	start := now.Add(-1 * time.Hour)
	end := now.Add(2 * time.Hour)
	results, err := store.Query(ctx, deviceName, start, end)
	require.NoError(t, err)
	assert.Equal(t, 100, len(results))
}