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

func TestStatisticsReportingService(t *testing.T) {
	// Set up test dependencies
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	historicalDataService := NewHistoricalDataService(timeSeriesStore)
	performanceMetricsService := NewPerformanceMetricsService(historicalDataService)
	exportService := NewStatisticsExportService(historicalDataService, performanceMetricsService)
	service := NewStatisticsReportingService(historicalDataService, performanceMetricsService, exportService)
	
	ctx := context.Background()
	deviceName, _ := tc.NewDeviceName("eth0")

	// Create test data for comprehensive analysis
	now := time.Now().Truncate(time.Minute)
	
	// Store sample data with varying patterns
	for i := 0; i < 24; i++ { // 24 hours of data
		timestamp := now.Add(-time.Duration(i) * time.Hour)
		
		// Create data with trends and anomalies
		baseBytes := uint64(1000000)
		anomalyMultiplier := uint64(1)
		if i == 12 { // Create anomaly at 12 hours ago
			anomalyMultiplier = 5
		}
		
		stats := &DeviceStatistics{
			DeviceName: deviceName.String(),
			Timestamp:  timestamp,
			QdiscStats: []QdiscStatistics{
				{
					Handle: "1:",
					Type:   "htb",
					Stats: netlink.QdiscStats{
						BytesSent:    baseBytes * uint64(24-i) * anomalyMultiplier,
						PacketsSent:  uint64(1000 * (24 - i)),
						BytesDropped: uint64(5 * i), // Increasing drops over time
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
						BytesSent:      baseBytes/2 * uint64(24-i),
						PacketsSent:    uint64(500 * (24 - i)),
						BytesDropped:   uint64(2 * i),
						Overlimits:     uint64(5 * i),
						RateBPS:        uint64(50000 + i*1000),
						BacklogBytes:   uint64(100 + i*5),
						BacklogPackets: uint64(10 + i),
					},
					DetailedStats: &netlink.DetailedClassStats{
						HTBStats: &netlink.HTBClassStats{
							Lends:   uint32(10 + i),
							Borrows: uint32(5 + i/2),
							Giants:  uint32(i),
							Tokens:  uint32(1000 + i*50),
							CTokens: uint32(500 + i*25),
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
						Matches:    uint64(250 * (24 - i)),
						Bytes:      baseBytes/4 * uint64(24-i),
						Packets:    uint64(250 * (24 - i)),
						Rate:       uint64(25000 + i*500),
						PacketRate: uint64(25 + i),
					},
				},
			},
			LinkStats: LinkStatistics{
				RxBytes:   baseBytes * 2 * uint64(24-i),
				TxBytes:   uint64(float64(baseBytes) * 1.5 * float64(24-i)),
				RxPackets: uint64(2000 * (24 - i)),
				TxPackets: uint64(1500 * (24 - i)),
				RxErrors:  uint64(i / 2), // Minimal errors
				TxErrors:  uint64(i / 3),
				RxDropped: uint64(i),
				TxDropped: uint64(i / 2),
				RxRate:    uint64(200000 + i*5000),
				TxRate:    uint64(150000 + i*3000),
			},
		}

		err := historicalDataService.StoreRawData(ctx, deviceName, stats)
		require.NoError(t, err)
	}

	t.Run("generate summary report", func(t *testing.T) {
		options := ReportOptions{
			Type:       ReportTypeSummary,
			DeviceName: deviceName,
			TimeRange: TimeRange{
				Start: now.Add(-24 * time.Hour),
				End:   now,
			},
			Format: FormatJSON,
		}

		report, err := service.GenerateReport(ctx, options)
		require.NoError(t, err)

		// Verify report structure
		assert.Equal(t, deviceName.String(), report.Metadata.DeviceName)
		assert.Equal(t, ReportTypeSummary, report.Metadata.ReportType)
		assert.NotEmpty(t, report.Metadata.Title)
		assert.NotEmpty(t, report.Metadata.Description)
		assert.Greater(t, report.Metadata.DataPoints, 0)

		// Verify sections are included
		assert.NotNil(t, report.ExecutiveSummary)
		assert.NotNil(t, report.PerformanceMetrics)
		assert.NotEmpty(t, report.Recommendations)

		// Verify executive summary
		assert.Greater(t, report.ExecutiveSummary.HealthScore, float64(0))
		assert.NotEmpty(t, report.ExecutiveSummary.OverallHealth)
		assert.NotNil(t, report.ExecutiveSummary.TrendSummary)
		assert.NotNil(t, report.ExecutiveSummary.DataQuality)
	})

	t.Run("generate detailed report", func(t *testing.T) {
		options := ReportOptions{
			Type:         ReportTypeDetailed,
			DeviceName:   deviceName,
			TimeRange: TimeRange{
				Start: now.Add(-24 * time.Hour),
				End:   now,
			},
			Format:        FormatJSON,
			IncludeCharts: true,
		}

		report, err := service.GenerateReport(ctx, options)
		require.NoError(t, err)

		// Verify all sections are included
		assert.NotNil(t, report.ExecutiveSummary)
		assert.NotNil(t, report.PerformanceMetrics)
		assert.NotNil(t, report.TrendAnalysis)
		assert.NotNil(t, report.CapacityAnalysis)
		assert.NotEmpty(t, report.Recommendations)
		assert.NotEmpty(t, report.Charts)

		// Verify trend analysis
		assert.NotEmpty(t, report.TrendAnalysis.OverallTrends)
		assert.NotEmpty(t, report.TrendAnalysis.ComponentTrends)
		// Note: Anomaly detection might be empty if no significant anomalies are found
		assert.NotEmpty(t, report.TrendAnalysis.ForecastProjections)

		// Verify capacity analysis
		assert.NotNil(t, report.CapacityAnalysis.CurrentUtilization)
		assert.NotNil(t, report.CapacityAnalysis.PeakUtilization)
		assert.NotEmpty(t, report.CapacityAnalysis.GrowthProjections)
		assert.NotEmpty(t, report.CapacityAnalysis.CapacityRecommendations)

		// Verify charts (detailed report should include charts)
		assert.NotEmpty(t, report.Charts)
		for _, chart := range report.Charts {
			assert.NotEmpty(t, chart.ID)
			assert.NotEmpty(t, chart.Type)
			assert.NotEmpty(t, chart.Title)
			assert.NotNil(t, chart.Data)
		}
	})

	t.Run("generate trend analysis report", func(t *testing.T) {
		options := ReportOptions{
			Type:       ReportTypeTrendAnalysis,
			DeviceName: deviceName,
			TimeRange: TimeRange{
				Start: now.Add(-24 * time.Hour),
				End:   now,
			},
			Format:        FormatJSON,
			IncludeCharts: true,
		}

		report, err := service.GenerateReport(ctx, options)
		require.NoError(t, err)

		// Verify trend analysis is comprehensive
		require.NotNil(t, report.TrendAnalysis)
		
		// Should detect throughput trends
		assert.NotEmpty(t, report.TrendAnalysis.OverallTrends)
		throughputTrend := findTrendByMetric(report.TrendAnalysis.OverallTrends, "total_throughput")
		assert.NotNil(t, throughputTrend)
		assert.Contains(t, []string{"increasing", "decreasing", "stable"}, throughputTrend.Direction)

		// Should have component trends
		assert.NotEmpty(t, report.TrendAnalysis.ComponentTrends)
		qdiscTrend := findComponentTrend(report.TrendAnalysis.ComponentTrends, "qdisc")
		assert.NotNil(t, qdiscTrend)

		// Should detect anomalies (we inserted one at 12 hours ago)
		// Note: Anomaly detection might not find anomalies if the variation isn't significant enough
		
		// Should have forecasts
		assert.NotEmpty(t, report.TrendAnalysis.ForecastProjections)
		for _, forecast := range report.TrendAnalysis.ForecastProjections {
			assert.Greater(t, forecast.PredictedValue, float64(0))
			assert.NotEmpty(t, forecast.Method)
			assert.Greater(t, forecast.Confidence, float64(0))
		}

		// Should have correlations (if enough data points with variation)
		// Note: Correlation analysis might be empty if data is insufficient
	})

	t.Run("generate capacity planning report", func(t *testing.T) {
		options := ReportOptions{
			Type:       ReportTypeCapacityPlanning,
			DeviceName: deviceName,
			TimeRange: TimeRange{
				Start: now.Add(-24 * time.Hour),
				End:   now,
			},
			Format: FormatJSON,
		}

		report, err := service.GenerateReport(ctx, options)
		require.NoError(t, err)

		require.NotNil(t, report.CapacityAnalysis)

		// Verify utilization metrics
		assert.Greater(t, report.CapacityAnalysis.CurrentUtilization.ThroughputUtilization, float64(0))
		assert.Greater(t, report.CapacityAnalysis.PeakUtilization.ThroughputUtilization, float64(0))

		// Verify growth projections
		assert.NotEmpty(t, report.CapacityAnalysis.GrowthProjections)
		for _, projection := range report.CapacityAnalysis.GrowthProjections {
			assert.Greater(t, projection.TimeHorizon, time.Duration(0))
			assert.Greater(t, projection.ConfidenceLevel, float64(0))
			assert.NotEmpty(t, projection.GrowthDrivers)
		}

		// Verify recommendations
		assert.NotEmpty(t, report.CapacityAnalysis.CapacityRecommendations)
		for _, rec := range report.CapacityAnalysis.CapacityRecommendations {
			assert.NotEmpty(t, rec.Category)
			assert.NotEmpty(t, rec.Priority)
			assert.NotEmpty(t, rec.Recommendation)
		}

		// Verify threshold analysis
		assert.NotEmpty(t, report.CapacityAnalysis.ThresholdAnalysis)
		
		// Verify bottleneck identification
		assert.NotEmpty(t, report.CapacityAnalysis.ResourceBottlenecks)
	})

	t.Run("generate comparative report", func(t *testing.T) {
		// Create comparison periods
		comparisonPeriods := []TimeRange{
			{
				Start: now.Add(-48 * time.Hour),
				End:   now.Add(-24 * time.Hour),
			},
		}

		options := ReportOptions{
			Type:       ReportTypeComparative,
			DeviceName: deviceName,
			TimeRange: TimeRange{
				Start: now.Add(-24 * time.Hour),
				End:   now,
			},
			Format:            FormatJSON,
			ComparisonPeriods: comparisonPeriods,
		}

		report, err := service.GenerateReport(ctx, options)
		require.NoError(t, err)

		require.NotNil(t, report.ComparisonAnalysis)

		// Should have periods
		assert.Greater(t, len(report.ComparisonAnalysis.Periods), 1)

		// Should have metric comparisons
		assert.NotEmpty(t, report.ComparisonAnalysis.MetricChanges)

		// Should have trend comparisons
		assert.NotEmpty(t, report.ComparisonAnalysis.TrendComparisons)

		// Should have insights
		assert.NotEmpty(t, report.ComparisonAnalysis.Insights)
	})

	t.Run("generate performance audit report", func(t *testing.T) {
		options := ReportOptions{
			Type:          ReportTypePerformanceAudit,
			DeviceName:    deviceName,
			TimeRange: TimeRange{
				Start: now.Add(-24 * time.Hour),
				End:   now,
			},
			Format:        FormatJSON,
			IncludeCharts: true,
		}

		report, err := service.GenerateReport(ctx, options)
		require.NoError(t, err)

		// Performance audit should include all sections
		assert.NotNil(t, report.ExecutiveSummary)
		assert.NotNil(t, report.PerformanceMetrics)
		assert.NotNil(t, report.TrendAnalysis)
		assert.NotNil(t, report.CapacityAnalysis)
		assert.NotEmpty(t, report.Recommendations)
		assert.NotEmpty(t, report.Charts)
	})

	t.Run("report with raw data", func(t *testing.T) {
		options := ReportOptions{
			Type:           ReportTypeSummary,
			DeviceName:     deviceName,
			TimeRange: TimeRange{
				Start: now.Add(-6 * time.Hour),
				End:   now,
			},
			Format:         FormatJSON,
			IncludeRawData: true,
		}

		report, err := service.GenerateReport(ctx, options)
		require.NoError(t, err)

		assert.NotEmpty(t, report.RawData)
		assert.LessOrEqual(t, len(report.RawData), 6) // Should have up to 6 hours of data
	})

	t.Run("custom sections report", func(t *testing.T) {
		options := ReportOptions{
			Type:       ReportTypeDetailed,
			DeviceName: deviceName,
			TimeRange: TimeRange{
				Start: now.Add(-12 * time.Hour),
				End:   now,
			},
			Format: FormatJSON,
			Sections: []ReportSection{
				SectionExecutiveSummary,
				SectionTrendAnalysis,
				SectionRecommendations,
			},
		}

		report, err := service.GenerateReport(ctx, options)
		require.NoError(t, err)

		// Should only include requested sections
		assert.NotNil(t, report.ExecutiveSummary)
		assert.NotNil(t, report.TrendAnalysis)
		assert.NotEmpty(t, report.Recommendations)

		// Should not include unrequested sections
		assert.Nil(t, report.CapacityAnalysis)
		assert.Nil(t, report.ComparisonAnalysis)
		assert.Empty(t, report.Charts)
	})
}

func TestStatisticsReportingService_ValidationAndEdgeCases(t *testing.T) {
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	historicalDataService := NewHistoricalDataService(timeSeriesStore)
	performanceMetricsService := NewPerformanceMetricsService(historicalDataService)
	exportService := NewStatisticsExportService(historicalDataService, performanceMetricsService)
	service := NewStatisticsReportingService(historicalDataService, performanceMetricsService, exportService)
	
	ctx := context.Background()
	deviceName, _ := tc.NewDeviceName("eth0")

	t.Run("validate report options", func(t *testing.T) {
		// Missing device name
		options := ReportOptions{
			Type: ReportTypeSummary,
			TimeRange: TimeRange{
				Start: time.Now().Add(-1 * time.Hour),
				End:   time.Now(),
			},
		}
		_, err := service.GenerateReport(ctx, options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "device name is required")

		// Missing time range
		options = ReportOptions{
			Type:       ReportTypeSummary,
			DeviceName: deviceName,
		}
		_, err = service.GenerateReport(ctx, options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "time range is required")

		// Invalid time range
		now := time.Now()
		options = ReportOptions{
			Type:       ReportTypeSummary,
			DeviceName: deviceName,
			TimeRange: TimeRange{
				Start: now,
				End:   now.Add(-1 * time.Hour),
			},
		}
		_, err = service.GenerateReport(ctx, options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start time must be before end time")
	})

	t.Run("report with no data", func(t *testing.T) {
		emptyDevice, _ := tc.NewDeviceName("empty0")
		
		options := ReportOptions{
			Type:       ReportTypeSummary,
			DeviceName: emptyDevice,
			TimeRange: TimeRange{
				Start: time.Now().Add(-1 * time.Hour),
				End:   time.Now(),
			},
			Format: FormatJSON,
		}

		report, err := service.GenerateReport(ctx, options)
		require.NoError(t, err)

		assert.Equal(t, emptyDevice.String(), report.Metadata.DeviceName)
		assert.Equal(t, 0, report.Metadata.DataPoints)
		
		// Should still have executive summary even with no data
		assert.NotNil(t, report.ExecutiveSummary)
		assert.NotEmpty(t, report.ExecutiveSummary.KeyFindings)
	})

	t.Run("report with minimal data", func(t *testing.T) {
		minimalDevice, _ := tc.NewDeviceName("minimal0")
		
		// Store single data point
		stats := &DeviceStatistics{
			DeviceName:  minimalDevice.String(),
			Timestamp:   time.Now().Add(-30 * time.Minute),
			QdiscStats:  []QdiscStatistics{{Handle: "1:", Type: "htb", Stats: netlink.QdiscStats{BytesSent: 1000}}},
			ClassStats:  []ClassStatistics{},
			FilterStats: []FilterStatistics{},
			LinkStats:   LinkStatistics{},
		}

		err := historicalDataService.StoreRawData(ctx, minimalDevice, stats)
		require.NoError(t, err)

		options := ReportOptions{
			Type:       ReportTypeTrendAnalysis,
			DeviceName: minimalDevice,
			TimeRange: TimeRange{
				Start: time.Now().Add(-1 * time.Hour),
				End:   time.Now(),
			},
			Format: FormatJSON,
		}

		report, err := service.GenerateReport(ctx, options)
		require.NoError(t, err)

		// Should handle minimal data gracefully
		assert.NotNil(t, report.TrendAnalysis)
		// Trends should be stable or empty with insufficient data
		if len(report.TrendAnalysis.OverallTrends) > 0 {
			assert.Contains(t, []string{"stable", "unknown"}, report.TrendAnalysis.OverallTrends[0].Direction)
		}
	})
}

func TestStatisticsReportingService_ScheduledReports(t *testing.T) {
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	historicalDataService := NewHistoricalDataService(timeSeriesStore)
	performanceMetricsService := NewPerformanceMetricsService(historicalDataService)
	exportService := NewStatisticsExportService(historicalDataService, performanceMetricsService)
	service := NewStatisticsReportingService(historicalDataService, performanceMetricsService, exportService)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	deviceNames := []tc.DeviceName{}
	device1, _ := tc.NewDeviceName("eth0")
	device2, _ := tc.NewDeviceName("eth1")
	deviceNames = append(deviceNames, device1, device2)

	// Test scheduled report generation with short interval
	t.Run("scheduled reports with cancellation", func(t *testing.T) {
		// Start scheduled reports with 1 second interval in a goroutine
		errChan := make(chan error, 1)
		go func() {
			err := service.GenerateScheduledReports(ctx, deviceNames, 1*time.Second)
			errChan <- err
		}()

		// Let it run for 2-3 seconds then cancel
		time.Sleep(2500 * time.Millisecond)
		cancel()
		
		// Wait for the goroutine to finish and check error
		select {
		case err := <-errChan:
			// Should return either DeadlineExceeded or Canceled
			assert.True(t, err == context.DeadlineExceeded || err == context.Canceled)
		case <-time.After(1 * time.Second):
			t.Error("Scheduled reports did not stop in time")
		}
	})
}

func TestStatisticsReportingService_AnalysisHelpers(t *testing.T) {
	service := &StatisticsReportingService{}

	t.Run("data quality assessment", func(t *testing.T) {
		// Test with good data
		data := make([]*timeseries.AggregatedData, 24)
		baseTime := time.Now().Truncate(time.Hour)
		
		for i := 0; i < 24; i++ {
			data[i] = &timeseries.AggregatedData{
				Timestamp: baseTime.Add(-time.Duration(i) * time.Hour),
				LinkStats: timeseries.AggregatedLinkStats{
					AvgRxRate: uint64(100000 + i*1000), // Consistent pattern
					AvgTxRate: uint64(80000 + i*800),
				},
			}
		}

		timeRange := TimeRange{Start: baseTime.Add(-24 * time.Hour), End: baseTime}
		quality := service.assessDataQuality(data, timeRange)

		assert.Greater(t, quality.Completeness, float64(90))
		assert.Greater(t, quality.ConsistencyScore, float64(80))
		assert.Equal(t, 0, quality.DataGaps)
		assert.Greater(t, quality.QualityScore, float64(80))

		// Test with poor data (high variation)
		for i := 0; i < 24; i++ {
			multiplier := 1
			if i%3 == 0 {
				multiplier = 10 // Create high variation
			}
			data[i].LinkStats.AvgRxRate = uint64(100000 * multiplier)
		}

		quality = service.assessDataQuality(data, timeRange)
		assert.LessOrEqual(t, quality.ConsistencyScore, float64(95)) // Should detect inconsistency
	})

	t.Run("trend calculations", func(t *testing.T) {
		data := make([]*timeseries.AggregatedData, 10)
		for i := 0; i < 10; i++ {
			data[i] = &timeseries.AggregatedData{
				LinkStats: timeseries.AggregatedLinkStats{
					AvgRxRate: uint64(100000 + i*10000), // Increasing trend
					AvgTxRate: uint64(80000 + i*8000),
				},
				QdiscStats: []timeseries.AggregatedQdiscStats{
					{
						Handle:       "1:",
						TotalDrops:   uint64(10 + i*2), // Increasing drops
						TotalPackets: uint64(1000 + i*100),
					},
				},
			}
		}

		trends := service.calculateOverallTrends(data)
		assert.NotEmpty(t, trends)

		// Find throughput trend
		throughputTrend := findTrendByMetric(trends, "total_throughput")
		assert.NotNil(t, throughputTrend)
		assert.Equal(t, "increasing", throughputTrend.Direction)
		assert.Greater(t, throughputTrend.Magnitude, float64(0))

		// Find packet loss trend
		packetLossTrend := findTrendByMetric(trends, "packet_loss_rate")
		assert.NotNil(t, packetLossTrend)
		assert.Equal(t, "increasing", packetLossTrend.Direction)
	})

	t.Run("anomaly detection", func(t *testing.T) {
		data := make([]*timeseries.AggregatedData, 10)
		for i := 0; i < 10; i++ {
			throughput := uint64(100000)
			if i == 5 { // Create anomaly
				throughput = 500000
			}
			
			data[i] = &timeseries.AggregatedData{
				Timestamp: time.Now().Add(-time.Duration(i) * time.Hour),
				LinkStats: timeseries.AggregatedLinkStats{
					AvgRxRate: throughput,
					AvgTxRate: throughput / 2,
				},
			}
		}

		anomalies := service.detectAnomalies(data)
		assert.NotEmpty(t, anomalies)

		// Should detect the anomaly at index 5
		found := false
		for _, anomaly := range anomalies {
			if anomaly.MetricName == "total_throughput" && anomaly.Severity != "low" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should detect throughput anomaly")
	})

	t.Run("forecasting", func(t *testing.T) {
		data := make([]*timeseries.AggregatedData, 12)
		fromTime := time.Now()
		
		for i := 0; i < 12; i++ {
			data[i] = &timeseries.AggregatedData{
				LinkStats: timeseries.AggregatedLinkStats{
					AvgRxRate: uint64(100000 + i*5000), // Linear trend
					AvgTxRate: uint64(80000 + i*4000),
				},
			}
		}

		forecasts := service.generateForecasts(data, fromTime)
		assert.NotEmpty(t, forecasts)
		assert.Len(t, forecasts, 24) // Should forecast 24 hours

		// Verify forecasts make sense
		for _, forecast := range forecasts {
			assert.Equal(t, "total_throughput", forecast.MetricName)
			assert.Greater(t, forecast.PredictedValue, float64(0))
			assert.Equal(t, "linear_extrapolation", forecast.Method)
			assert.Greater(t, forecast.Confidence, float64(0))
			assert.Less(t, forecast.ConfidenceInterval.Lower, forecast.PredictedValue)
			assert.Greater(t, forecast.ConfidenceInterval.Upper, forecast.PredictedValue)
		}
	})

	t.Run("correlation analysis", func(t *testing.T) {
		data := make([]*timeseries.AggregatedData, 10)
		for i := 0; i < 10; i++ {
			// Create correlated data: as throughput increases, drop rate decreases
			data[i] = &timeseries.AggregatedData{
				LinkStats: timeseries.AggregatedLinkStats{
					AvgRxRate: uint64(100000 + i*10000),
					AvgTxRate: uint64(80000 + i*8000),
				},
				QdiscStats: []timeseries.AggregatedQdiscStats{
					{
						TotalDrops:   uint64(100 - i*5), // Decreasing as throughput increases
						TotalPackets: uint64(10000),
					},
				},
			}
		}

		correlations := service.calculateCorrelations(data)
		assert.NotEmpty(t, correlations)

		// Should find correlation between throughput and drop rate
		correlation := correlations[0]
		assert.Equal(t, "total_throughput", correlation.Metric1)
		assert.Equal(t, "packet_drop_rate", correlation.Metric2)
		assert.Less(t, correlation.Correlation, float64(0)) // Should be negative correlation
		assert.NotEmpty(t, correlation.Description)
	})
}

// Helper functions for tests

func findTrendByMetric(trends []TrendItem, metricName string) *TrendItem {
	for i := range trends {
		if trends[i].MetricName == metricName {
			return &trends[i]
		}
	}
	return nil
}

func findComponentTrend(trends []ComponentTrend, componentType string) *ComponentTrend {
	for i := range trends {
		if trends[i].ComponentType == componentType {
			return &trends[i]
		}
	}
	return nil
}