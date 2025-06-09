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

func TestDashboardDataService(t *testing.T) {
	// Set up test dependencies
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	historicalDataService := NewHistoricalDataService(timeSeriesStore)
	performanceMetricsService := NewPerformanceMetricsService(historicalDataService)
	service := NewDashboardDataService(historicalDataService, performanceMetricsService)
	
	ctx := context.Background()
	deviceName, _ := tc.NewDeviceName("eth0")

	// Create comprehensive test data
	now := time.Now().Truncate(time.Minute)
	
	// Store multiple data points for trend analysis
	for i := 0; i < 10; i++ {
		timestamp := now.Add(-time.Duration(i) * time.Minute)
		
		// Create realistic data with some variation
		baseRate := uint64(100000000) // 100 Mbps base
		variation := uint64(i * 5000000) // Add variation
		
		stats := &DeviceStatistics{
			DeviceName: deviceName.String(),
			Timestamp:  timestamp,
			QdiscStats: []QdiscStatistics{
				{
					Handle: "1:",
					Type:   "htb",
					Stats: netlink.QdiscStats{
						BytesSent:    uint64(1000000 * (10 - i)),
						PacketsSent:  uint64(1000 * (10 - i)),
						BytesDropped: uint64(5 * i), // Increasing drops
						Overlimits:   uint64(10 * i),
						Requeues:     uint64(2 * i),
					},
				},
			},
			ClassStats: []ClassStatistics{
				{
					Handle: "1:10",
					Parent: "1:",
					Type:   "htb",
					Stats: netlink.ClassStats{
						BytesSent:      uint64(500000 * (10 - i)),
						PacketsSent:    uint64(500 * (10 - i)),
						BytesDropped:   uint64(2 * i),
						Overlimits:     uint64(5 * i),
						RateBPS:        baseRate + variation,
						BacklogBytes:   uint64(100 + i*10),
						BacklogPackets: uint64(10 + i),
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
						Matches: uint64(250 * (10 - i)),
						Bytes:   uint64(250000 * (10 - i)),
						Packets: uint64(250 * (10 - i)),
						Rate:    baseRate/4 + variation/4,
					},
				},
			},
			LinkStats: LinkStatistics{
				RxBytes:   uint64(2000000 * (10 - i)),
				TxBytes:   uint64(1500000 * (10 - i)),
				RxPackets: uint64(2000 * (10 - i)),
				TxPackets: uint64(1500 * (10 - i)),
				RxErrors:  uint64(i),
				TxErrors:  uint64(i / 2),
				RxDropped: uint64(i),
				TxDropped: uint64(i / 2),
				RxRate:    baseRate + variation,
				TxRate:    (baseRate + variation) * 3 / 4,
			},
		}

		err := historicalDataService.StoreRawData(ctx, deviceName, stats)
		require.NoError(t, err)
	}

	t.Run("get real-time metrics", func(t *testing.T) {
		metrics, err := service.GetRealTimeMetrics(ctx, deviceName)
		require.NoError(t, err)

		// Verify basic structure
		assert.Equal(t, deviceName.String(), metrics.DeviceName)
		assert.NotZero(t, metrics.Timestamp)

		// Verify health indicator
		assert.NotEmpty(t, metrics.OverallHealth.Status)
		assert.GreaterOrEqual(t, metrics.OverallHealth.Score, float64(0))
		assert.LessOrEqual(t, metrics.OverallHealth.Score, float64(100))

		// Verify throughput metrics
		assert.Greater(t, metrics.ThroughputMetrics.CurrentRxRate, uint64(0))
		assert.Greater(t, metrics.ThroughputMetrics.CurrentTxRate, uint64(0))
		assert.GreaterOrEqual(t, metrics.ThroughputMetrics.UtilizationRx, float64(0))
		assert.LessOrEqual(t, metrics.ThroughputMetrics.UtilizationRx, float64(100))

		// Verify quality metrics
		assert.GreaterOrEqual(t, metrics.QualityMetrics.PacketLossRate, float64(0))
		assert.GreaterOrEqual(t, metrics.QualityMetrics.ErrorRate, float64(0))
		assert.GreaterOrEqual(t, metrics.QualityMetrics.QualityScore, float64(0))
		assert.LessOrEqual(t, metrics.QualityMetrics.QualityScore, float64(100))

		// Verify resource utilization
		assert.GreaterOrEqual(t, metrics.ResourceUtilization.BandwidthUtilization, float64(0))
		assert.NotNil(t, metrics.ResourceUtilization.ComponentUtilization)

		// Verify error counts
		assert.GreaterOrEqual(t, metrics.ErrorCounts.TotalErrors, uint64(0))
		assert.Contains(t, []string{"increasing", "decreasing", "stable"}, metrics.ErrorCounts.ErrorTrend)
	})

	t.Run("get trend data", func(t *testing.T) {
		timeWindow := 10 * time.Minute
		trends, err := service.GetTrendData(ctx, deviceName, timeWindow)
		require.NoError(t, err)

		// Verify basic structure
		assert.Equal(t, timeWindow, trends.TimeWindow)
		assert.GreaterOrEqual(t, trends.TrendConfidence, float64(0))
		assert.LessOrEqual(t, trends.TrendConfidence, float64(1))

		// Verify trend information
		assert.Contains(t, []string{"increasing", "decreasing", "stable"}, trends.ThroughputTrend.Direction)
		assert.Contains(t, []string{"increasing", "decreasing", "stable"}, trends.ErrorRateTrend.Direction)
		assert.Contains(t, []string{"increasing", "decreasing", "stable"}, trends.QualityTrend.Direction)

		// Verify predictions
		assert.NotNil(t, trends.PredictedValues.NextMinute)
		assert.NotNil(t, trends.PredictedValues.NextFiveMin)
		assert.NotNil(t, trends.PredictedValues.NextFifteenMin)
		assert.GreaterOrEqual(t, trends.PredictedValues.Confidence, float64(0))
	})

	t.Run("get dashboard update", func(t *testing.T) {
		deviceNames := []tc.DeviceName{deviceName}
		
		update, err := service.GetDashboardUpdate(ctx, deviceNames)
		require.NoError(t, err)

		// Verify basic structure
		assert.NotEmpty(t, update.UpdateID)
		assert.NotZero(t, update.Timestamp)
		assert.Len(t, update.DeviceUpdates, 1)

		// Verify device update
		deviceUpdate, exists := update.DeviceUpdates[deviceName.String()]
		assert.True(t, exists)
		assert.Equal(t, deviceName.String(), deviceUpdate.DeviceName)

		// Verify system summary
		require.NotNil(t, update.SystemSummary)
		assert.Equal(t, 1, update.SystemSummary.TotalDevices)
		assert.Greater(t, update.SystemSummary.TotalThroughput, uint64(0))
		assert.NotEmpty(t, update.SystemSummary.SystemHealth.Status)

		// Verify update metadata
		assert.NotEmpty(t, update.UpdateMetadata.UpdateType)
		assert.Greater(t, update.UpdateMetadata.UpdateDuration, time.Duration(0))
		assert.Equal(t, 1, update.UpdateMetadata.RecordsProcessed)
	})

	t.Run("caching functionality", func(t *testing.T) {
		// First call should cache the data
		start := time.Now()
		metrics1, err := service.GetRealTimeMetrics(ctx, deviceName)
		require.NoError(t, err)
		duration1 := time.Since(start)

		// Second call should use cache (should be faster)
		start = time.Now()
		metrics2, err := service.GetRealTimeMetrics(ctx, deviceName)
		require.NoError(t, err)
		duration2 := time.Since(start)

		// Verify cached data is returned
		assert.Equal(t, metrics1.Timestamp, metrics2.Timestamp)
		
		// Second call should be faster (from cache)
		assert.Less(t, duration2, duration1)
	})

	t.Run("multiple devices dashboard update", func(t *testing.T) {
		// Create another device
		device2, _ := tc.NewDeviceName("eth1")
		
		// Store some data for device2
		stats := &DeviceStatistics{
			DeviceName: device2.String(),
			Timestamp:  now,
			QdiscStats: []QdiscStatistics{
				{
					Handle: "1:",
					Type:   "fq",
					Stats: netlink.QdiscStats{
						BytesSent:   500000,
						PacketsSent: 500,
					},
				},
			},
			LinkStats: LinkStatistics{
				RxBytes: 1000000,
				TxBytes: 800000,
				RxRate:  50000000,
				TxRate:  40000000,
			},
		}
		
		err := historicalDataService.StoreRawData(ctx, device2, stats)
		require.NoError(t, err)

		deviceNames := []tc.DeviceName{deviceName, device2}
		update, err := service.GetDashboardUpdate(ctx, deviceNames)
		require.NoError(t, err)

		// Verify both devices are included
		assert.Len(t, update.DeviceUpdates, 2)
		assert.Contains(t, update.DeviceUpdates, deviceName.String())
		assert.Contains(t, update.DeviceUpdates, device2.String())

		// Verify system summary reflects multiple devices
		assert.Equal(t, 2, update.SystemSummary.TotalDevices)
	})
}

func TestDashboardDataService_EdgeCases(t *testing.T) {
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	historicalDataService := NewHistoricalDataService(timeSeriesStore)
	performanceMetricsService := NewPerformanceMetricsService(historicalDataService)
	service := NewDashboardDataService(historicalDataService, performanceMetricsService)
	
	ctx := context.Background()

	t.Run("no data available", func(t *testing.T) {
		emptyDevice, _ := tc.NewDeviceName("empty0")
		
		metrics, err := service.GetRealTimeMetrics(ctx, emptyDevice)
		require.NoError(t, err)

		// Should return default metrics
		assert.Equal(t, emptyDevice.String(), metrics.DeviceName)
		assert.Equal(t, "unknown", metrics.OverallHealth.Status)
		assert.Equal(t, float64(0), metrics.OverallHealth.Score)

		// Should handle trends gracefully
		trends, err := service.GetTrendData(ctx, emptyDevice, 10*time.Minute)
		require.NoError(t, err)
		assert.Equal(t, float64(0), trends.TrendConfidence)
	})

	t.Run("single data point", func(t *testing.T) {
		singleDevice, _ := tc.NewDeviceName("single0")
		
		// Store single data point
		stats := &DeviceStatistics{
			DeviceName: singleDevice.String(),
			Timestamp:  time.Now(),
			QdiscStats: []QdiscStatistics{
				{
					Handle: "1:",
					Type:   "htb",
					Stats: netlink.QdiscStats{
						BytesSent:   1000,
						PacketsSent: 10,
					},
				},
			},
			LinkStats: LinkStatistics{
				RxBytes: 2000,
				TxBytes: 1500,
				RxRate:  100000,
				TxRate:  80000,
			},
		}

		err := historicalDataService.StoreRawData(ctx, singleDevice, stats)
		require.NoError(t, err)

		metrics, err := service.GetRealTimeMetrics(ctx, singleDevice)
		require.NoError(t, err)

		// Should handle single data point gracefully
		assert.Greater(t, metrics.OverallHealth.Score, float64(0))
		assert.Equal(t, uint64(100000), metrics.ThroughputMetrics.CurrentRxRate)

		// Trends should indicate stable with low confidence
		trends, err := service.GetTrendData(ctx, singleDevice, 5*time.Minute)
		require.NoError(t, err)
		assert.Equal(t, float64(0), trends.TrendConfidence)
	})

	t.Run("high error rates", func(t *testing.T) {
		errorDevice, _ := tc.NewDeviceName("error0")
		
		// Create data with high error rates
		stats := &DeviceStatistics{
			DeviceName: errorDevice.String(),
			Timestamp:  time.Now(),
			QdiscStats: []QdiscStatistics{
				{
					Handle: "1:",
					Type:   "htb",
					Stats: netlink.QdiscStats{
						BytesSent:    1000000,
						PacketsSent:  1000,
						BytesDropped: 100000, // 10% packet loss
						Overlimits:   500,
					},
				},
			},
			LinkStats: LinkStatistics{
				RxBytes:   2000000,
				TxBytes:   1500000,
				RxPackets: 2000,
				TxPackets: 1500,
				RxErrors:  200, // 10% error rate
				TxErrors:  150,
				RxRate:    100000,
				TxRate:    80000,
			},
		}

		err := historicalDataService.StoreRawData(ctx, errorDevice, stats)
		require.NoError(t, err)

		metrics, err := service.GetRealTimeMetrics(ctx, errorDevice)
		require.NoError(t, err)

		// Should detect high error conditions
		assert.Greater(t, metrics.QualityMetrics.PacketLossRate, float64(5))
		assert.Greater(t, metrics.QualityMetrics.ErrorRate, float64(5))
		assert.Less(t, metrics.OverallHealth.Score, float64(70))
		assert.Contains(t, []string{"warning", "critical"}, metrics.OverallHealth.Status)
		assert.NotEmpty(t, metrics.OverallHealth.Issues)
	})

	t.Run("high utilization", func(t *testing.T) {
		highUtilDevice, _ := tc.NewDeviceName("highutil0")
		
		// Create data with high bandwidth utilization
		highRate := uint64(900000000) // 900 Mbps (90% of 1 Gbps)
		
		stats := &DeviceStatistics{
			DeviceName: highUtilDevice.String(),
			Timestamp:  time.Now(),
			QdiscStats: []QdiscStatistics{
				{
					Handle: "1:",
					Type:   "htb",
					Stats: netlink.QdiscStats{
						BytesSent:   10000000,
						PacketsSent: 10000,
					},
				},
			},
			LinkStats: LinkStatistics{
				RxBytes:   20000000,
				TxBytes:   15000000,
				RxPackets: 20000,
				TxPackets: 15000,
				RxRate:    highRate,
				TxRate:    highRate * 3 / 4,
			},
		}

		err := historicalDataService.StoreRawData(ctx, highUtilDevice, stats)
		require.NoError(t, err)

		metrics, err := service.GetRealTimeMetrics(ctx, highUtilDevice)
		require.NoError(t, err)

		// Should detect high utilization
		assert.Greater(t, metrics.ResourceUtilization.BandwidthUtilization, float64(80))
		assert.Greater(t, metrics.ThroughputMetrics.UtilizationRx, float64(80))
		assert.Less(t, metrics.OverallHealth.Score, float64(80))
	})
}

func TestDashboardDataService_Calculations(t *testing.T) {
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	historicalDataService := NewHistoricalDataService(timeSeriesStore)
	performanceMetricsService := NewPerformanceMetricsService(historicalDataService)
	service := NewDashboardDataService(historicalDataService, performanceMetricsService)

	t.Run("throughput calculations", func(t *testing.T) {
		// Create test data with known values
		historical := []*timeseries.AggregatedData{
			{
				LinkStats: timeseries.AggregatedLinkStats{
					AvgRxRate:      50000000,  // 50 Mbps
					AvgTxRate:      40000000,  // 40 Mbps
					TotalRxBytes:   5000000,
					TotalTxBytes:   4000000,
				},
			},
			{
				LinkStats: timeseries.AggregatedLinkStats{
					AvgRxRate:      60000000,  // 60 Mbps (20% increase)
					AvgTxRate:      48000000,  // 48 Mbps (20% increase)
					TotalRxBytes:   6000000,
					TotalTxBytes:   4800000,
				},
			},
		}

		latest := historical[1]
		metrics := service.calculateThroughputMetrics(latest, historical)

		assert.Equal(t, uint64(60000000), metrics.CurrentRxRate)
		assert.Equal(t, uint64(48000000), metrics.CurrentTxRate)
		assert.Equal(t, uint64(60000000), metrics.PeakRxRate)
		assert.Equal(t, uint64(48000000), metrics.PeakTxRate)

		// Check utilization (60 Mbps / 1000 Mbps = 6%)
		assert.InDelta(t, 6.0, metrics.UtilizationRx, 0.1)
		assert.InDelta(t, 4.8, metrics.UtilizationTx, 0.1)

		// Check rate changes (20% increase)
		assert.InDelta(t, 20.0, metrics.RateChangeRx, 0.1)
		assert.InDelta(t, 20.0, metrics.RateChangeTx, 0.1)
	})

	t.Run("quality calculations", func(t *testing.T) {
		latest := &timeseries.AggregatedData{
			QdiscStats: []timeseries.AggregatedQdiscStats{
				{
					TotalPackets:    1000,
					TotalDrops:      50,    // 5% packet loss
					TotalOverlimits: 20,    // 2% overlimits
				},
			},
			LinkStats: timeseries.AggregatedLinkStats{
				TotalRxPackets: 2000,
				TotalTxPackets: 1500,
				TotalRxErrors:  10,     // 0.5% error rate
				TotalTxErrors:  5,
			},
		}

		metrics := service.calculateQualityMetrics(latest)

		assert.InDelta(t, 5.0, metrics.PacketLossRate, 0.1)
		assert.InDelta(t, 2.0, metrics.OverlimitRate, 0.1)
		
		// Error rate: (10+5)/(2000+1500) = 15/3500 = 0.43%
		assert.InDelta(t, 0.43, metrics.ErrorRate, 0.1)
	})

	t.Run("health indicator calculations", func(t *testing.T) {
		// Test excellent health
		excellentMetrics := &RealTimeMetrics{
			QualityMetrics: QualityMetrics{
				PacketLossRate: 0.1,
				ErrorRate:      0.05,
			},
			ResourceUtilization: ResourceUtilizationMetrics{
				BandwidthUtilization: 50.0,
				CriticalResources:    []string{},
			},
		}

		health := service.calculateHealthIndicator(excellentMetrics)
		assert.Greater(t, health.Score, float64(90))
		assert.Equal(t, "excellent", health.Status)
		assert.Empty(t, health.Issues)

		// Test critical health
		criticalMetrics := &RealTimeMetrics{
			QualityMetrics: QualityMetrics{
				PacketLossRate: 5.0,   // High packet loss
				ErrorRate:      2.0,   // High error rate
			},
			ResourceUtilization: ResourceUtilizationMetrics{
				BandwidthUtilization: 95.0, // High utilization
				CriticalResources:    []string{"qdisc1", "qdisc2"},
			},
		}

		health = service.calculateHealthIndicator(criticalMetrics)
		assert.Less(t, health.Score, float64(50))
		assert.Equal(t, "critical", health.Status)
		assert.NotEmpty(t, health.Issues)
	})

	t.Run("trend calculations", func(t *testing.T) {
		// Create historical data with clear trend
		historical := make([]*timeseries.AggregatedData, 5)
		for i := 0; i < 5; i++ {
			historical[i] = &timeseries.AggregatedData{
				LinkStats: timeseries.AggregatedLinkStats{
					AvgRxRate: uint64(50000000 + i*10000000), // Increasing trend
					AvgTxRate: uint64(40000000 + i*8000000),
				},
			}
		}

		trend := service.calculateThroughputTrend(historical)

		assert.Equal(t, "increasing", trend.Direction)
		assert.Greater(t, trend.Magnitude, float64(0))
		assert.Greater(t, trend.Confidence, float64(0))
		assert.Equal(t, float64(90000000), trend.StartValue) // 50M + 40M
		assert.InDelta(t, float64(162000000), trend.CurrentValue, 1000000) // 90M + 72M with tolerance
	})
}

func TestDashboardDataService_LiveUpdates(t *testing.T) {
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	historicalDataService := NewHistoricalDataService(timeSeriesStore)
	performanceMetricsService := NewPerformanceMetricsService(historicalDataService)
	service := NewDashboardDataService(historicalDataService, performanceMetricsService)
	
	deviceName, _ := tc.NewDeviceName("eth0")

	t.Run("live updates with callback", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		var updateCount int
		var lastUpdate *DashboardUpdate
		done := make(chan error, 1)

		updateCallback := func(update *DashboardUpdate) {
			updateCount++
			lastUpdate = update
		}

		// Start live updates in background
		go func() {
			err := service.StartLiveUpdates(ctx, []tc.DeviceName{deviceName}, updateCallback)
			done <- err
		}()

		// Wait for some updates then cancel
		time.Sleep(1500 * time.Millisecond)
		cancel()

		// Wait for goroutine to complete
		err := <-done
		
		// Should have context cancellation or deadline exceeded
		require.True(t, err == context.Canceled || err == context.DeadlineExceeded, "Expected context cancellation, got: %v", err)

		// Should have received at least one update
		assert.Greater(t, updateCount, 0)
		if lastUpdate != nil {
			assert.NotEmpty(t, lastUpdate.UpdateID)
		}
	})
}

func TestDashboardDataService_SystemSummary(t *testing.T) {
	service := &DashboardDataService{}

	t.Run("system summary generation", func(t *testing.T) {
		// Create mock device updates
		deviceUpdates := map[string]*RealTimeMetrics{
			"eth0": {
				OverallHealth: HealthIndicator{
					Status: "excellent",
					Score:  95.0,
					Issues: []string{},
				},
				ThroughputMetrics: ThroughputMetrics{
					CurrentRxRate: 100000000,
					CurrentTxRate: 80000000,
				},
				ErrorCounts: ErrorCountMetrics{
					TotalErrors: 10,
				},
			},
			"eth1": {
				OverallHealth: HealthIndicator{
					Status: "warning",
					Score:  65.0,
					Issues: []string{"High utilization"},
				},
				ThroughputMetrics: ThroughputMetrics{
					CurrentRxRate: 200000000,
					CurrentTxRate: 160000000,
				},
				ErrorCounts: ErrorCountMetrics{
					TotalErrors: 50,
				},
			},
			"eth2": {
				OverallHealth: HealthIndicator{
					Status: "critical",
					Score:  30.0,
					Issues: []string{"High packet loss", "High errors"},
				},
				ThroughputMetrics: ThroughputMetrics{
					CurrentRxRate: 50000000,
					CurrentTxRate: 40000000,
				},
				ErrorCounts: ErrorCountMetrics{
					TotalErrors: 100,
				},
			},
		}

		summary := service.generateSystemSummary(deviceUpdates)

		// Verify device counts
		assert.Equal(t, 3, summary.TotalDevices)
		assert.Equal(t, 1, summary.HealthyDevices)   // eth0
		assert.Equal(t, 1, summary.WarningDevices)   // eth1
		assert.Equal(t, 1, summary.CriticalDevices)  // eth2

		// Verify aggregated metrics
		expectedThroughput := uint64(100000000 + 80000000 + 200000000 + 160000000 + 50000000 + 40000000)
		assert.Equal(t, expectedThroughput, summary.TotalThroughput)
		assert.Equal(t, uint64(160), summary.TotalErrors) // 10 + 50 + 100

		// Verify system health calculation
		expectedAvgScore := (95.0 + 65.0 + 30.0) / 3 // 63.33
		assert.InDelta(t, expectedAvgScore, summary.SystemHealth.Score, 1.0)
		assert.Equal(t, "warning", summary.SystemHealth.Status) // 63.33 falls in warning range

		// Verify issues deduplication
		assert.Contains(t, summary.TopIssues, "High utilization")
		assert.Contains(t, summary.TopIssues, "High packet loss")
		assert.Contains(t, summary.TopIssues, "High errors")
	})
}

func TestDashboardDataService_AlertGeneration(t *testing.T) {
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	historicalDataService := NewHistoricalDataService(timeSeriesStore)
	performanceMetricsService := NewPerformanceMetricsService(historicalDataService)
	service := NewDashboardDataService(historicalDataService, performanceMetricsService)
	
	ctx := context.Background()
	deviceName, _ := tc.NewDeviceName("alert-test")

	t.Run("alert generation for critical conditions", func(t *testing.T) {
		// Create data that should trigger alerts
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
						BytesDropped: 30000, // 3% packet loss (should trigger alert)
					},
				},
			},
			LinkStats: LinkStatistics{
				RxBytes:   20000000,
				TxBytes:   15000000,
				RxPackets: 20000,
				TxPackets: 15000,
				RxRate:    980000000, // 98% utilization (should trigger alert)
				TxRate:    950000000,
			},
		}

		err := historicalDataService.StoreRawData(ctx, deviceName, stats)
		require.NoError(t, err)

		alerts := service.collectGlobalAlerts(ctx, []tc.DeviceName{deviceName})

		// Should generate alerts for critical conditions
		assert.NotEmpty(t, alerts)

		// Check for packet loss alert
		foundPacketLossAlert := false
		foundUtilizationAlert := false

		for _, alert := range alerts {
			if alert.Category == "performance" && alert.Title == "High Packet Loss" {
				foundPacketLossAlert = true
				assert.Equal(t, "critical", alert.Level)
				assert.Greater(t, alert.MetricValue, 2.0)
			}
			if alert.Category == "capacity" && alert.Title == "High Bandwidth Utilization" {
				foundUtilizationAlert = true
				assert.Equal(t, "warning", alert.Level)
				assert.Greater(t, alert.MetricValue, 95.0)
			}
		}

		assert.True(t, foundPacketLossAlert, "Should generate packet loss alert")
		assert.True(t, foundUtilizationAlert, "Should generate utilization alert")
	})
}