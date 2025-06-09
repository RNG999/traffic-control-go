package application

import (
	"context"
	"testing"
	"time"

	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	"github.com/rng999/traffic-control-go/internal/infrastructure/timeseries"
	"github.com/rng999/traffic-control-go/pkg/tc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistoricalDataService(t *testing.T) {
	// Set up test dependencies
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	service := NewHistoricalDataService(timeSeriesStore)
	ctx := context.Background()
	deviceName, _ := tc.NewDeviceName("eth0")

	t.Run("store and retrieve raw data", func(t *testing.T) {
		// Create test statistics
		now := time.Now().Truncate(time.Second)
		stats := &DeviceStatistics{
			DeviceName: deviceName.String(),
			Timestamp:  now,
			QdiscStats: []QdiscStatistics{
				{
					Handle: "1:",
					Type:   "htb",
					Stats: netlink.QdiscStats{
						BytesSent:    1000000,
						PacketsSent:  1000,
						BytesDropped: 5,
						Overlimits:   10,
						Requeues:     2,
					},
				},
			},
			ClassStats: []ClassStatistics{
				{
					Handle: "1:10",
					Parent: "1:",
					Type:   "htb",
					Stats: netlink.ClassStats{
						BytesSent:    500000,
						PacketsSent:  500,
						BytesDropped: 2,
						Overlimits:   5,
						RateBPS:      50000,
						BacklogBytes: 250,
					},
				},
			},
			FilterStats: []FilterStatistics{
				{
					Handle:   "800:100",
					Parent:   "1:",
					Priority: 100,
					FlowID:   "1:10",
					Stats: FilterStats{
						Matches: 250,
						Bytes:   250000,
						Packets: 250,
						Rate:    25000,
					},
				},
			},
			LinkStats: LinkStatistics{
				RxBytes:   2000000,
				TxBytes:   1500000,
				RxPackets: 2000,
				TxPackets: 1500,
				RxRate:    200000,
				TxRate:    150000,
			},
		}

		// Store the data
		err := service.StoreRawData(ctx, deviceName, stats)
		require.NoError(t, err)

		// Retrieve the data
		start := now.Add(-1 * time.Hour)
		end := now.Add(1 * time.Hour)
		historicalData, err := service.GetHistoricalData(ctx, deviceName, start, end, "")
		require.NoError(t, err)
		require.Len(t, historicalData, 1)

		// Verify the data
		retrieved := historicalData[0]
		assert.Equal(t, deviceName.String(), retrieved.DeviceName)
		assert.Equal(t, now.Unix(), retrieved.Timestamp.Unix())
		assert.Len(t, retrieved.QdiscStats, 1)
		assert.Equal(t, "1:", retrieved.QdiscStats[0].Handle)
		assert.Equal(t, uint64(1000000), retrieved.QdiscStats[0].TotalBytes)
	})

	t.Run("data aggregation by hour", func(t *testing.T) {
		// Clear previous data
		timeSeriesStore.Close()
		timeSeriesStore = timeseries.NewMemoryTimeSeriesStore()
		service = NewHistoricalDataService(timeSeriesStore)

		now := time.Now().Truncate(time.Hour)
		
		// Create multiple data points within the same hour
		for i := 0; i < 5; i++ {
			timestamp := now.Add(time.Duration(i*10) * time.Minute)
			stats := &DeviceStatistics{
				DeviceName: deviceName.String(),
				Timestamp:  timestamp,
				QdiscStats: []QdiscStatistics{
					{
						Handle: "1:",
						Type:   "htb",
						Stats: netlink.QdiscStats{
							BytesSent:    uint64(1000000 * (i + 1)),
							PacketsSent:  uint64(1000 * (i + 1)),
							BytesDropped: uint64(5 * (i + 1)),
							Overlimits:   uint64(10 * (i + 1)),
							Requeues:     uint64(2 * (i + 1)),
						},
					},
				},
				ClassStats:  []ClassStatistics{},
				FilterStats: []FilterStatistics{},
				LinkStats:   LinkStatistics{},
			}

			err := service.StoreRawData(ctx, deviceName, stats)
			require.NoError(t, err)
		}

		// Perform hourly aggregation
		startTime := now
		endTime := now.Add(1 * time.Hour)
		err := service.AggregateData(ctx, deviceName, timeseries.IntervalHour, startTime, endTime)
		require.NoError(t, err)

		// Retrieve aggregated data
		aggregatedData, err := service.GetHistoricalData(ctx, deviceName, startTime, endTime, timeseries.IntervalHour)
		require.NoError(t, err)
		require.Len(t, aggregatedData, 1)

		// Verify aggregation
		aggregated := aggregatedData[0]
		assert.Equal(t, timeseries.IntervalHour, aggregated.Interval)
		assert.Len(t, aggregated.QdiscStats, 1)
		
		qdiscAgg := aggregated.QdiscStats[0]
		assert.Equal(t, "1:", qdiscAgg.Handle)
		assert.Equal(t, uint64(15000000), qdiscAgg.TotalBytes)   // Sum: 1M+2M+3M+4M+5M = 15M
		assert.Equal(t, uint64(15000), qdiscAgg.TotalPackets)    // Sum: 1K+2K+3K+4K+5K = 15K
		assert.Equal(t, uint64(75), qdiscAgg.TotalDrops)         // Sum: 5+10+15+20+25 = 75
		assert.Equal(t, uint64(100000), qdiscAgg.MinRate)        // Default rate from conversion
		assert.Equal(t, uint64(100000), qdiscAgg.MaxRate)        // Default rate from conversion
	})

	t.Run("scheduled aggregation", func(t *testing.T) {
		// Clear previous data
		timeSeriesStore.Close()
		timeSeriesStore = timeseries.NewMemoryTimeSeriesStore()
		service = NewHistoricalDataService(timeSeriesStore)

		// Create test data spanning multiple hours
		baseTime := time.Now().Add(-6 * time.Hour).Truncate(time.Hour)
		
		for hour := 0; hour < 6; hour++ {
			for minute := 0; minute < 60; minute += 10 {
				timestamp := baseTime.Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)
				stats := &DeviceStatistics{
					DeviceName: deviceName.String(),
					Timestamp:  timestamp,
					QdiscStats: []QdiscStatistics{
						{
							Handle: "1:",
							Type:   "htb",
							Stats: netlink.QdiscStats{
								BytesSent:   uint64(1000000 + hour*100000 + minute*1000),
								PacketsSent: uint64(1000 + hour*100 + minute),
								Overlimits:  uint64(hour*10 + minute/10),
							},
						},
					},
					ClassStats:  []ClassStatistics{},
					FilterStats: []FilterStatistics{},
					LinkStats:   LinkStatistics{},
				}

				err := service.StoreRawData(ctx, deviceName, stats)
				require.NoError(t, err)
			}
		}

		// Perform scheduled aggregation
		err := service.PerformScheduledAggregation(ctx, deviceName)
		require.NoError(t, err)

		// Verify hourly aggregations were created
		start := baseTime
		end := baseTime.Add(6 * time.Hour)
		hourlyData, err := service.GetHistoricalData(ctx, deviceName, start, end, timeseries.IntervalHour)
		require.NoError(t, err)
		assert.True(t, len(hourlyData) > 0, "Should have hourly aggregated data")
	})

	t.Run("data cleanup", func(t *testing.T) {
		// Create old data
		oldTime := time.Now().Add(-48 * time.Hour)
		oldStats := &DeviceStatistics{
			DeviceName:  deviceName.String(),
			Timestamp:   oldTime,
			QdiscStats:  []QdiscStatistics{{Handle: "1:", Type: "htb", Stats: netlink.QdiscStats{BytesSent: 1000}}},
			ClassStats:  []ClassStatistics{},
			FilterStats: []FilterStatistics{},
			LinkStats:   LinkStatistics{},
		}

		err := service.StoreRawData(ctx, deviceName, oldStats)
		require.NoError(t, err)

		// Verify old data exists
		start := oldTime.Add(-1 * time.Hour)
		end := oldTime.Add(1 * time.Hour)
		oldData, err := service.GetHistoricalData(ctx, deviceName, start, end, "")
		require.NoError(t, err)
		assert.Len(t, oldData, 1)

		// Clean up old data (24 hour retention)
		retentionPolicy := timeseries.RetentionPolicy{
			RawDataRetention: 24 * time.Hour,
		}
		err = service.CleanupOldData(ctx, deviceName, retentionPolicy)
		require.NoError(t, err)

		// Verify old data is cleaned up
		oldData, err = service.GetHistoricalData(ctx, deviceName, start, end, "")
		require.NoError(t, err)
		assert.Len(t, oldData, 0)
	})

	t.Run("data summary", func(t *testing.T) {
		// Clear previous data
		timeSeriesStore.Close()
		timeSeriesStore = timeseries.NewMemoryTimeSeriesStore()
		service = NewHistoricalDataService(timeSeriesStore)

		// Create test data
		now := time.Now()
		for i := 0; i < 10; i++ {
			timestamp := now.Add(-time.Duration(i) * time.Hour)
			stats := &DeviceStatistics{
				DeviceName:  deviceName.String(),
				Timestamp:   timestamp,
				QdiscStats:  []QdiscStatistics{{Handle: "1:", Type: "htb", Stats: netlink.QdiscStats{BytesSent: 1000}}},
				ClassStats:  []ClassStatistics{},
				FilterStats: []FilterStatistics{},
				LinkStats:   LinkStatistics{},
			}

			err := service.StoreRawData(ctx, deviceName, stats)
			require.NoError(t, err)
		}

		// Get data summary
		summary, err := service.GetDataSummary(ctx, deviceName)
		require.NoError(t, err)
		
		assert.True(t, summary.HasData)
		assert.Equal(t, deviceName.String(), summary.DeviceName)
		assert.Equal(t, 10, summary.TotalDataPoints)
		assert.True(t, summary.TimeSpan > 0)
		assert.True(t, summary.DataDensity > 0)
	})

	t.Run("empty device summary", func(t *testing.T) {
		emptyDevice, _ := tc.NewDeviceName("empty0")
		
		summary, err := service.GetDataSummary(ctx, emptyDevice)
		require.NoError(t, err)
		
		assert.False(t, summary.HasData)
		assert.Equal(t, emptyDevice.String(), summary.DeviceName)
		assert.Equal(t, 0, summary.TotalDataPoints)
	})

	t.Run("interval start calculation", func(t *testing.T) {
		testTime := time.Date(2023, 10, 18, 14, 35, 42, 0, time.UTC) // Wednesday

		// Test minute interval
		minuteStart := service.getIntervalStart(testTime, timeseries.IntervalMinute)
		expected := time.Date(2023, 10, 18, 14, 35, 0, 0, time.UTC)
		assert.Equal(t, expected, minuteStart)

		// Test hour interval
		hourStart := service.getIntervalStart(testTime, timeseries.IntervalHour)
		expected = time.Date(2023, 10, 18, 14, 0, 0, 0, time.UTC)
		assert.Equal(t, expected, hourStart)

		// Test day interval
		dayStart := service.getIntervalStart(testTime, timeseries.IntervalDay)
		expected = time.Date(2023, 10, 18, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, expected, dayStart)

		// Test week interval (should start on Monday)
		weekStart := service.getIntervalStart(testTime, timeseries.IntervalWeek)
		expected = time.Date(2023, 10, 16, 0, 0, 0, 0, time.UTC) // Monday
		assert.Equal(t, expected, weekStart)

		// Test month interval
		monthStart := service.getIntervalStart(testTime, timeseries.IntervalMonth)
		expected = time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, expected, monthStart)
	})
}

func TestHistoricalDataService_AggregationCalculations(t *testing.T) {
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	service := NewHistoricalDataService(timeSeriesStore)

	t.Run("qdisc stats aggregation", func(t *testing.T) {
		// Create test data points
		dataPoints := []*timeseries.TimeSeriesData{
			{
				QdiscStats: []timeseries.QdiscDataPoint{
					{Handle: "1:", Type: "htb", Bytes: 1000, Packets: 10, Rate: 100000, Backlog: 500},
					{Handle: "2:", Type: "fq", Bytes: 2000, Packets: 20, Rate: 200000, Backlog: 1000},
				},
			},
			{
				QdiscStats: []timeseries.QdiscDataPoint{
					{Handle: "1:", Type: "htb", Bytes: 1500, Packets: 15, Rate: 150000, Backlog: 750},
					{Handle: "2:", Type: "fq", Bytes: 2500, Packets: 25, Rate: 250000, Backlog: 1250},
				},
			},
		}

		aggregated := service.aggregateQdiscStats(dataPoints)
		require.Len(t, aggregated, 2)

		// Find specific qdiscs
		var htbStats, fqStats *timeseries.AggregatedQdiscStats
		for i := range aggregated {
			if aggregated[i].Handle == "1:" {
				htbStats = &aggregated[i]
			} else if aggregated[i].Handle == "2:" {
				fqStats = &aggregated[i]
			}
		}

		require.NotNil(t, htbStats)
		require.NotNil(t, fqStats)

		// Verify HTB aggregation
		assert.Equal(t, "1:", htbStats.Handle)
		assert.Equal(t, "htb", htbStats.Type)
		assert.Equal(t, uint64(2500), htbStats.TotalBytes)  // 1000 + 1500
		assert.Equal(t, uint64(25), htbStats.TotalPackets)  // 10 + 15
		assert.Equal(t, uint64(100000), htbStats.MinRate)   // min(100000, 150000)
		assert.Equal(t, uint64(150000), htbStats.MaxRate)   // max(100000, 150000)
		assert.Equal(t, uint64(750), htbStats.MaxBacklog)   // max(500, 750)

		// Verify FQ aggregation
		assert.Equal(t, "2:", fqStats.Handle)
		assert.Equal(t, "fq", fqStats.Type)
		assert.Equal(t, uint64(4500), fqStats.TotalBytes)   // 2000 + 2500
		assert.Equal(t, uint64(45), fqStats.TotalPackets)   // 20 + 25
		assert.Equal(t, uint64(200000), fqStats.MinRate)    // min(200000, 250000)
		assert.Equal(t, uint64(250000), fqStats.MaxRate)    // max(200000, 250000)
	})

	t.Run("class stats aggregation", func(t *testing.T) {
		dataPoints := []*timeseries.TimeSeriesData{
			{
				ClassStats: []timeseries.ClassDataPoint{
					{Handle: "1:10", Parent: "1:", Type: "htb", Bytes: 1000, Packets: 10, Rate: 100000, Lends: 5, Borrows: 2},
				},
			},
			{
				ClassStats: []timeseries.ClassDataPoint{
					{Handle: "1:10", Parent: "1:", Type: "htb", Bytes: 1500, Packets: 15, Rate: 150000, Lends: 7, Borrows: 3},
				},
			},
		}

		aggregated := service.aggregateClassStats(dataPoints)
		require.Len(t, aggregated, 1)

		classStats := aggregated[0]
		assert.Equal(t, "1:10", classStats.Handle)
		assert.Equal(t, "1:", classStats.Parent)
		assert.Equal(t, "htb", classStats.Type)
		assert.Equal(t, uint64(2500), classStats.TotalBytes)    // 1000 + 1500
		assert.Equal(t, uint64(25), classStats.TotalPackets)    // 10 + 15
		assert.Equal(t, uint64(12), classStats.TotalLends)      // 5 + 7
		assert.Equal(t, uint64(5), classStats.TotalBorrows)     // 2 + 3
		assert.Equal(t, uint64(100000), classStats.MinRate)     // min(100000, 150000)
		assert.Equal(t, uint64(150000), classStats.MaxRate)     // max(100000, 150000)
	})

	t.Run("filter stats aggregation", func(t *testing.T) {
		dataPoints := []*timeseries.TimeSeriesData{
			{
				FilterStats: []timeseries.FilterDataPoint{
					{Handle: "800:100", Parent: "1:", Priority: 100, FlowID: "1:10", Matches: 50, Bytes: 5000, Packets: 50, Rate: 50000},
				},
			},
			{
				FilterStats: []timeseries.FilterDataPoint{
					{Handle: "800:100", Parent: "1:", Priority: 100, FlowID: "1:10", Matches: 75, Bytes: 7500, Packets: 75, Rate: 75000},
				},
			},
		}

		aggregated := service.aggregateFilterStats(dataPoints)
		require.Len(t, aggregated, 1)

		filterStats := aggregated[0]
		assert.Equal(t, "800:100", filterStats.Handle)
		assert.Equal(t, "1:", filterStats.Parent)
		assert.Equal(t, uint16(100), filterStats.Priority)
		assert.Equal(t, "1:10", filterStats.FlowID)
		assert.Equal(t, uint64(125), filterStats.TotalMatches)  // 50 + 75
		assert.Equal(t, uint64(12500), filterStats.TotalBytes)  // 5000 + 7500
		assert.Equal(t, uint64(125), filterStats.TotalPackets)  // 50 + 75
		assert.Equal(t, uint64(75000), filterStats.MaxRate)     // max(50000, 75000)
	})

	t.Run("link stats aggregation", func(t *testing.T) {
		dataPoints := []*timeseries.TimeSeriesData{
			{
				LinkStats: timeseries.LinkDataPoint{
					RxBytes: 10000, TxBytes: 8000, RxPackets: 100, TxPackets: 80,
					RxRate: 100000, TxRate: 80000,
				},
			},
			{
				LinkStats: timeseries.LinkDataPoint{
					RxBytes: 15000, TxBytes: 12000, RxPackets: 150, TxPackets: 120,
					RxRate: 150000, TxRate: 120000,
				},
			},
		}

		aggregated := service.aggregateLinkStats(dataPoints)
		
		// Should use the latest data point for totals
		assert.Equal(t, uint64(15000), aggregated.TotalRxBytes)
		assert.Equal(t, uint64(12000), aggregated.TotalTxBytes)
		assert.Equal(t, uint64(150), aggregated.TotalRxPackets)
		assert.Equal(t, uint64(120), aggregated.TotalTxPackets)
		
		// Should calculate averages and maximums for rates
		assert.Equal(t, uint64(125000), aggregated.AvgRxRate)   // (100000 + 150000) / 2
		assert.Equal(t, uint64(100000), aggregated.AvgTxRate)   // (80000 + 120000) / 2
		assert.Equal(t, uint64(150000), aggregated.MaxRxRate)   // max(100000, 150000)
		assert.Equal(t, uint64(120000), aggregated.MaxTxRate)   // max(80000, 120000)
	})
}

func TestHistoricalDataService_ConversionFunctions(t *testing.T) {
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	service := NewHistoricalDataService(timeSeriesStore)
	deviceName, _ := tc.NewDeviceName("eth0")

	t.Run("convert application stats to time-series data", func(t *testing.T) {
		stats := &DeviceStatistics{
			DeviceName: deviceName.String(),
			Timestamp:  time.Now(),
			QdiscStats: []QdiscStatistics{
				{
					Handle: "1:",
					Type:   "htb",
					Stats: netlink.QdiscStats{
						BytesSent:    1000000,
						PacketsSent:  1000,
						BytesDropped: 5,
						Overlimits:   10,
						Requeues:     2,
					},
				},
			},
			ClassStats: []ClassStatistics{
				{
					Handle: "1:10",
					Parent: "1:",
					Type:   "htb",
					Stats: netlink.ClassStats{
						BytesSent:      500000,
						PacketsSent:    500,
						BytesDropped:   2,
						Overlimits:     5,
						RateBPS:        50000,
						BacklogBytes:   250,
						BacklogPackets: 5,
					},
					DetailedStats: &netlink.DetailedClassStats{
						HTBStats: &netlink.HTBClassStats{
							Lends:   10,
							Borrows: 5,
							Giants:  1,
							Tokens:  1000,
							CTokens: 500,
						},
					},
				},
			},
			FilterStats: []FilterStatistics{
				{
					Handle:   "800:100",
					Parent:   "1:",
					Priority: 100,
					FlowID:   "1:10",
					Stats: FilterStats{
						Matches: 250,
						Bytes:   250000,
						Packets: 250,
						Rate:    25000,
						PacketRate: 25,
					},
				},
			},
			LinkStats: LinkStatistics{
				RxBytes:   2000000,
				TxBytes:   1500000,
				RxPackets: 2000,
				TxPackets: 1500,
				RxErrors:  10,
				TxErrors:  5,
				RxDropped: 2,
				TxDropped: 1,
				RxRate:    200000,
				TxRate:    150000,
			},
		}

		tsData := service.convertToTimeSeriesData(stats)

		// Verify qdisc conversion
		require.Len(t, tsData.QdiscStats, 1)
		qdisc := tsData.QdiscStats[0]
		assert.Equal(t, "1:", qdisc.Handle)
		assert.Equal(t, "htb", qdisc.Type)
		assert.Equal(t, uint64(1000000), qdisc.Bytes)
		assert.Equal(t, uint64(1000), qdisc.Packets)
		assert.Equal(t, uint64(5), qdisc.Drops)
		assert.Equal(t, uint64(100000), qdisc.Rate)

		// Verify class conversion
		require.Len(t, tsData.ClassStats, 1)
		class := tsData.ClassStats[0]
		assert.Equal(t, "1:10", class.Handle)
		assert.Equal(t, "1:", class.Parent)
		assert.Equal(t, "htb", class.Type)
		assert.Equal(t, uint64(500000), class.Bytes)
		assert.Equal(t, uint64(10), class.Lends)
		assert.Equal(t, uint64(5), class.Borrows)
		assert.Equal(t, int64(1000), class.Tokens)

		// Verify filter conversion
		require.Len(t, tsData.FilterStats, 1)
		filter := tsData.FilterStats[0]
		assert.Equal(t, "800:100", filter.Handle)
		assert.Equal(t, "1:", filter.Parent)
		assert.Equal(t, uint16(100), filter.Priority)
		assert.Equal(t, "1:10", filter.FlowID)
		assert.Equal(t, uint64(250), filter.Matches)

		// Verify link conversion
		link := tsData.LinkStats
		assert.Equal(t, uint64(2000000), link.RxBytes)
		assert.Equal(t, uint64(1500000), link.TxBytes)
		assert.Equal(t, uint64(200000), link.RxRate)
		assert.Equal(t, uint64(150000), link.TxRate)
	})

	t.Run("convert raw to aggregated data", func(t *testing.T) {
		rawData := &timeseries.TimeSeriesData{
			Timestamp:  time.Now(),
			DeviceName: deviceName.String(),
			QdiscStats: []timeseries.QdiscDataPoint{
				{Handle: "1:", Type: "htb", Bytes: 1000, Packets: 10, Rate: 100000},
			},
			ClassStats: []timeseries.ClassDataPoint{
				{Handle: "1:10", Parent: "1:", Type: "htb", Bytes: 500, Packets: 5, Rate: 50000},
			},
			FilterStats: []timeseries.FilterDataPoint{
				{Handle: "800:100", Parent: "1:", Priority: 100, FlowID: "1:10", Matches: 5, Bytes: 500, Rate: 50000},
			},
			LinkStats: timeseries.LinkDataPoint{
				RxBytes: 2000, TxBytes: 1500, RxRate: 200000, TxRate: 150000,
			},
		}

		aggregated := service.convertRawToAggregated(rawData)

		assert.Equal(t, rawData.Timestamp, aggregated.Timestamp)
		assert.Equal(t, rawData.DeviceName, aggregated.DeviceName)
		assert.Equal(t, "", string(aggregated.Interval))

		// Verify qdisc conversion
		require.Len(t, aggregated.QdiscStats, 1)
		qdiscAgg := aggregated.QdiscStats[0]
		assert.Equal(t, "1:", qdiscAgg.Handle)
		assert.Equal(t, uint64(1000), qdiscAgg.TotalBytes)
		assert.Equal(t, uint64(100000), qdiscAgg.AvgRate)
		assert.Equal(t, uint64(100000), qdiscAgg.MinRate)
		assert.Equal(t, uint64(100000), qdiscAgg.MaxRate)

		// Verify class conversion
		require.Len(t, aggregated.ClassStats, 1)
		classAgg := aggregated.ClassStats[0]
		assert.Equal(t, "1:10", classAgg.Handle)
		assert.Equal(t, uint64(500), classAgg.TotalBytes)
		assert.Equal(t, uint64(50000), classAgg.AvgRate)

		// Verify filter conversion
		require.Len(t, aggregated.FilterStats, 1)
		filterAgg := aggregated.FilterStats[0]
		assert.Equal(t, "800:100", filterAgg.Handle)
		assert.Equal(t, uint64(5), filterAgg.TotalMatches)
		assert.Equal(t, uint64(500), filterAgg.TotalBytes)

		// Verify link conversion
		linkAgg := aggregated.LinkStats
		assert.Equal(t, uint64(2000), linkAgg.TotalRxBytes)
		assert.Equal(t, uint64(1500), linkAgg.TotalTxBytes)
		assert.Equal(t, uint64(200000), linkAgg.AvgRxRate)
		assert.Equal(t, uint64(150000), linkAgg.AvgTxRate)
	})
}