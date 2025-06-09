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

func TestPerformanceMetricsService(t *testing.T) {
	// Set up test dependencies
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	historicalDataService := NewHistoricalDataService(timeSeriesStore)
	service := NewPerformanceMetricsService(historicalDataService)
	
	ctx := context.Background()
	deviceName, _ := tc.NewDeviceName("eth0")

	t.Run("calculate performance metrics with no data", func(t *testing.T) {
		timeRange := TimeRange{
			Start: time.Now().Add(-1 * time.Hour),
			End:   time.Now(),
		}

		metrics, err := service.CalculatePerformanceMetrics(ctx, deviceName, timeRange)
		require.NoError(t, err)
		
		assert.Equal(t, deviceName.String(), metrics.DeviceName)
		assert.Equal(t, "unknown", metrics.HealthScore.HealthStatus)
		assert.Contains(t, metrics.Recommendations[0], "No historical data available")
	})

	t.Run("calculate performance metrics with sample data", func(t *testing.T) {
		// Create and store sample historical data
		now := time.Now()
		
		// Store multiple data points to enable trend analysis
		for i := 0; i < 10; i++ {
			timestamp := now.Add(-time.Duration(i) * time.Minute * 10)
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
							BytesSent:      uint64(500000 * (i + 1)),
							PacketsSent:    uint64(500 * (i + 1)),
							BytesDropped:   uint64(2 * i),
							Overlimits:     uint64(5 * i),
							RateBPS:        uint64(50000 + i*5000),
							BacklogBytes:   uint64(100 + i*10),
							BacklogPackets: uint64(10 + i),
						},
						DetailedStats: &netlink.DetailedClassStats{
							HTBStats: &netlink.HTBClassStats{
								Lends:   uint32(10 + i),
								Borrows: uint32(5 + i/2),
								Giants:  uint32(i),
								Tokens:  uint32(1000 + i*100),
								CTokens: uint32(500 + i*50),
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
							Matches:    uint64(250 * (i + 1)),
							Bytes:      uint64(250000 * (i + 1)),
							Packets:    uint64(250 * (i + 1)),
							Rate:       uint64(25000 + i*2500),
							PacketRate: uint64(25 + i*2),
						},
					},
				},
				LinkStats: LinkStatistics{
					RxBytes:   uint64(2000000 * (i + 1)),
					TxBytes:   uint64(1500000 * (i + 1)),
					RxPackets: uint64(2000 * (i + 1)),
					TxPackets: uint64(1500 * (i + 1)),
					RxErrors:  uint64(i), // Minimal errors
					TxErrors:  uint64(i / 2),
					RxDropped: uint64(i),
					TxDropped: uint64(i / 2),
					RxRate:    uint64(200000 + i*20000),
					TxRate:    uint64(150000 + i*15000),
				},
			}

			err := historicalDataService.StoreRawData(ctx, deviceName, stats)
			require.NoError(t, err)
		}

		timeRange := TimeRange{
			Start: now.Add(-2 * time.Hour),
			End:   now,
		}

		metrics, err := service.CalculatePerformanceMetrics(ctx, deviceName, timeRange)
		require.NoError(t, err)

		// Verify basic structure
		assert.Equal(t, deviceName.String(), metrics.DeviceName)
		assert.Equal(t, timeRange, metrics.TimeRange)

		// Verify overall metrics
		assert.Greater(t, metrics.OverallMetrics.TotalThroughput, uint64(0))
		assert.Greater(t, metrics.OverallMetrics.PeakThroughput, uint64(0))
		assert.Greater(t, metrics.OverallMetrics.AverageThroughput, uint64(0))
		assert.GreaterOrEqual(t, metrics.OverallMetrics.EfficiencyScore, float64(0))
		assert.LessOrEqual(t, metrics.OverallMetrics.EfficiencyScore, float64(100))

		// Verify qdisc metrics
		require.Len(t, metrics.QdiscMetrics, 1)
		qdiscMetric := metrics.QdiscMetrics[0]
		assert.Equal(t, "1:", qdiscMetric.Handle)
		assert.Equal(t, "htb", qdiscMetric.Type)
		assert.Greater(t, qdiscMetric.TotalBytes, uint64(0))
		assert.Greater(t, qdiscMetric.TotalPackets, uint64(0))
		assert.GreaterOrEqual(t, qdiscMetric.DropRate, float64(0))
		assert.Greater(t, qdiscMetric.ThroughputMbps, float64(0))
		assert.Greater(t, qdiscMetric.AveragePacketSize, float64(0))
		assert.GreaterOrEqual(t, qdiscMetric.PerformanceScore, float64(0))
		assert.LessOrEqual(t, qdiscMetric.PerformanceScore, float64(100))

		// Verify class metrics
		require.Len(t, metrics.ClassMetrics, 1)
		classMetric := metrics.ClassMetrics[0]
		assert.Equal(t, "1:10", classMetric.Handle)
		assert.Equal(t, "1:", classMetric.Parent)
		assert.Equal(t, "htb", classMetric.Type)
		assert.Greater(t, classMetric.TotalBytes, uint64(0))
		assert.GreaterOrEqual(t, classMetric.HTBLendingRatio, float64(0))
		assert.LessOrEqual(t, classMetric.HTBLendingRatio, float64(1))
		assert.GreaterOrEqual(t, classMetric.HTBBorrowingRatio, float64(0))
		assert.LessOrEqual(t, classMetric.HTBBorrowingRatio, float64(1))

		// Verify filter metrics
		require.Len(t, metrics.FilterMetrics, 1)
		filterMetric := metrics.FilterMetrics[0]
		assert.Equal(t, "800:100", filterMetric.Handle)
		assert.Equal(t, "1:", filterMetric.Parent)
		assert.Equal(t, uint16(100), filterMetric.Priority)
		assert.Equal(t, "1:10", filterMetric.FlowID)
		assert.Greater(t, filterMetric.TotalMatches, uint64(0))
		assert.Greater(t, filterMetric.MatchRate, float64(0))

		// Verify link metrics
		assert.Greater(t, metrics.LinkMetrics.RxThroughputMbps, float64(0))
		assert.Greater(t, metrics.LinkMetrics.TxThroughputMbps, float64(0))
		assert.GreaterOrEqual(t, metrics.LinkMetrics.ErrorRate, float64(0))
		assert.GreaterOrEqual(t, metrics.LinkMetrics.DropRate, float64(0))
		assert.GreaterOrEqual(t, metrics.LinkMetrics.DuplexEfficiency, float64(0))
		assert.LessOrEqual(t, metrics.LinkMetrics.DuplexEfficiency, float64(100))

		// Verify trends
		assert.Contains(t, []string{"increasing", "decreasing", "stable"}, metrics.Trends.ThroughputTrend)
		assert.Contains(t, []string{"increasing", "decreasing", "stable"}, metrics.Trends.DropRateTrend)
		assert.GreaterOrEqual(t, metrics.Trends.TrendConfidence, float64(0))
		assert.LessOrEqual(t, metrics.Trends.TrendConfidence, float64(1))

		// Verify health score
		assert.GreaterOrEqual(t, metrics.HealthScore.OverallScore, float64(0))
		assert.LessOrEqual(t, metrics.HealthScore.OverallScore, float64(100))
		assert.Contains(t, []string{"excellent", "good", "fair", "poor", "critical", "unknown"}, metrics.HealthScore.HealthStatus)
		assert.NotEmpty(t, metrics.HealthScore.ComponentScores)

		// Verify recommendations
		assert.NotEmpty(t, metrics.Recommendations)
	})

	t.Run("calculate real-time metrics", func(t *testing.T) {
		// Store recent data
		now := time.Now()
		stats := &DeviceStatistics{
			DeviceName: deviceName.String(),
			Timestamp:  now.Add(-5 * time.Minute),
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
			ClassStats:  []ClassStatistics{},
			FilterStats: []FilterStatistics{},
			LinkStats: LinkStatistics{
				RxBytes: 2000000,
				TxBytes: 1500000,
				RxRate:  200000,
				TxRate:  150000,
			},
		}

		err := historicalDataService.StoreRawData(ctx, deviceName, stats)
		require.NoError(t, err)

		metrics, err := service.CalculateRealTimeMetrics(ctx, deviceName)
		require.NoError(t, err)

		assert.Equal(t, deviceName.String(), metrics.DeviceName)
		// Time range should be approximately 1 hour
		timeDiff := metrics.TimeRange.End.Sub(metrics.TimeRange.Start)
		assert.True(t, timeDiff >= 55*time.Minute && timeDiff <= 65*time.Minute)
	})
}

func TestPerformanceMetricsService_MetricCalculations(t *testing.T) {
	service := &PerformanceMetricsService{
		logger: nil, // Not needed for these tests
	}

	t.Run("calculate efficiency score", func(t *testing.T) {
		// Perfect conditions
		score := service.calculateEfficiencyScore(0, 0, 0)
		assert.Equal(t, float64(100), score)

		// High packet loss
		score = service.calculateEfficiencyScore(5.0, 0, 0)
		assert.LessOrEqual(t, score, float64(70)) // Should lose at least 30 points

		// High latency
		score = service.calculateEfficiencyScore(0, 5000, 0)
		assert.LessOrEqual(t, score, float64(75)) // Should lose points for latency

		// High jitter
		score = service.calculateEfficiencyScore(0, 0, 2.0)
		assert.LessOrEqual(t, score, float64(85)) // Should lose points for jitter

		// Very bad conditions
		score = service.calculateEfficiencyScore(10.0, 10000, 5.0)
		assert.LessOrEqual(t, score, float64(30)) // Should be very low
	})

	t.Run("calculate queue efficiency", func(t *testing.T) {
		// Good queue performance
		qdisc := timeseries.AggregatedQdiscStats{
			MaxBacklog:     1000,
			MaxQueueLength: 10,
		}
		efficiency := service.calculateQueueEfficiency(qdisc)
		assert.Greater(t, efficiency, float64(90))

		// Poor queue performance
		qdisc = timeseries.AggregatedQdiscStats{
			MaxBacklog:     50000,
			MaxQueueLength: 500,
		}
		efficiency = service.calculateQueueEfficiency(qdisc)
		assert.LessOrEqual(t, efficiency, float64(50))
	})

	t.Run("calculate qdisc performance score", func(t *testing.T) {
		// Good performance
		qdisc := timeseries.AggregatedQdiscStats{
			TotalPackets:    1000,
			TotalOverlimits: 10,
			MaxBacklog:      1000,
			MaxQueueLength:  10,
		}
		score := service.calculateQdiscPerformanceScore(qdisc, 0.1)
		assert.Greater(t, score, float64(80))

		// Poor performance
		qdisc = timeseries.AggregatedQdiscStats{
			TotalPackets:    1000,
			TotalOverlimits: 500,
			MaxBacklog:      50000,
			MaxQueueLength:  500,
		}
		score = service.calculateQdiscPerformanceScore(qdisc, 10.0)
		assert.Less(t, score, float64(50))
	})

	t.Run("calculate fairness score", func(t *testing.T) {
		// Balanced lending/borrowing
		class := timeseries.AggregatedClassStats{
			TotalLends:   50,
			TotalBorrows: 50,
		}
		score := service.calculateFairnessScore(class)
		assert.Greater(t, score, float64(90))

		// Unbalanced - all lending
		class = timeseries.AggregatedClassStats{
			TotalLends:   100,
			TotalBorrows: 0,
		}
		score = service.calculateFairnessScore(class)
		assert.Less(t, score, float64(60))

		// No sharing needed
		class = timeseries.AggregatedClassStats{
			TotalLends:   0,
			TotalBorrows: 0,
		}
		score = service.calculateFairnessScore(class)
		assert.Equal(t, float64(100), score)
	})

	t.Run("calculate classification efficiency", func(t *testing.T) {
		// Good classification
		filter := timeseries.AggregatedFilterStats{
			TotalMatches: 1000,
			TotalBytes:   1500000, // 1500 bytes per match
		}
		efficiency := service.calculateClassificationEfficiency(filter)
		assert.Equal(t, float64(100), efficiency)

		// No matches
		filter = timeseries.AggregatedFilterStats{
			TotalMatches: 0,
			TotalBytes:   0,
		}
		efficiency = service.calculateClassificationEfficiency(filter)
		assert.Equal(t, float64(0), efficiency)

		// Unusual packet sizes (too small or too large)
		filter = timeseries.AggregatedFilterStats{
			TotalMatches: 1000,
			TotalBytes:   30000, // 30 bytes per match - too small
		}
		efficiency = service.calculateClassificationEfficiency(filter)
		assert.Equal(t, float64(75), efficiency) // Default moderate efficiency
	})

	t.Run("calculate link performance score", func(t *testing.T) {
		// Excellent link performance
		score := service.calculateLinkPerformanceScore(0, 0, 90)
		assert.Greater(t, score, float64(95))

		// Poor link performance
		score = service.calculateLinkPerformanceScore(1.0, 5.0, 20)
		assert.Less(t, score, float64(50))
	})

	t.Run("calculate health scores", func(t *testing.T) {
		// Test throughput score
		overall := OverallPerformanceMetrics{
			ThroughputUtilization: 80.0,
			EfficiencyScore:       90.0,
		}
		score := service.calculateThroughputScore(overall)
		assert.Equal(t, float64(72), score) // 80 * 0.9

		// Test reliability score
		overall = OverallPerformanceMetrics{
			PacketLossRate:   2.0,
			LatencyIndicator: 5000,
		}
		score = service.calculateReliabilityScore(overall)
		assert.Less(t, score, float64(60)) // Should lose points for both issues

		// Test stability score
		trends := TrendAnalysis{
			ThroughputTrend:  "increasing",
			DropRateTrend:    "increasing",
			LatencyTrend:     "stable",
			TrendConfidence:  0.8,
		}
		score = service.calculateStabilityScore(trends)
		assert.Less(t, score, float64(70)) // Should lose points for negative trends
	})
}

func TestPerformanceMetricsService_TrendAnalysis(t *testing.T) {
	service := &PerformanceMetricsService{
		logger: nil,
	}

	t.Run("trend analysis with insufficient data", func(t *testing.T) {
		// Single data point
		data := []*timeseries.AggregatedData{
			{
				LinkStats: timeseries.AggregatedLinkStats{
					AvgRxRate: 100000,
					AvgTxRate: 80000,
				},
			},
		}

		trends := service.calculateTrends(data)
		assert.Equal(t, "stable", trends.ThroughputTrend)
		assert.Equal(t, "stable", trends.DropRateTrend)
		assert.Equal(t, float64(0), trends.TrendConfidence)
	})

	t.Run("trend analysis with increasing throughput", func(t *testing.T) {
		data := []*timeseries.AggregatedData{
			{
				LinkStats: timeseries.AggregatedLinkStats{
					AvgRxRate: 100000,
					AvgTxRate: 80000,
				},
				QdiscStats: []timeseries.AggregatedQdiscStats{
					{TotalDrops: 10, TotalPackets: 1000},
				},
			},
			{
				LinkStats: timeseries.AggregatedLinkStats{
					AvgRxRate: 120000, // 20% increase
					AvgTxRate: 96000,  // 20% increase
				},
				QdiscStats: []timeseries.AggregatedQdiscStats{
					{TotalDrops: 12, TotalPackets: 1200},
				},
			},
		}

		trends := service.calculateTrends(data)
		assert.Equal(t, "increasing", trends.ThroughputTrend)
		assert.Greater(t, trends.TrendConfidence, float64(0))
		assert.Greater(t, trends.PredictedThroughput, uint64(0))
	})

	t.Run("trend analysis with decreasing throughput", func(t *testing.T) {
		data := []*timeseries.AggregatedData{
			{
				LinkStats: timeseries.AggregatedLinkStats{
					AvgRxRate: 100000,
					AvgTxRate: 80000,
				},
				QdiscStats: []timeseries.AggregatedQdiscStats{
					{TotalDrops: 10, TotalPackets: 1000},
				},
			},
			{
				LinkStats: timeseries.AggregatedLinkStats{
					AvgRxRate: 80000, // 20% decrease
					AvgTxRate: 64000, // 20% decrease
				},
				QdiscStats: []timeseries.AggregatedQdiscStats{
					{TotalDrops: 15, TotalPackets: 800},
				},
			},
		}

		trends := service.calculateTrends(data)
		assert.Equal(t, "decreasing", trends.ThroughputTrend)
		assert.Equal(t, "increasing", trends.DropRateTrend) // Drop rate increased from 1% to 1.875%
	})
}

func TestPerformanceMetricsService_Recommendations(t *testing.T) {
	service := &PerformanceMetricsService{
		logger: nil,
	}

	t.Run("generate recommendations for healthy system", func(t *testing.T) {
		metrics := &PerformanceMetrics{
			HealthScore: HealthScore{
				OverallScore: 95.0,
			},
			OverallMetrics: OverallPerformanceMetrics{
				PacketLossRate: 0.1,
			},
			QdiscMetrics: []QdiscPerformanceMetrics{
				{
					Handle:          "1:",
					DropRate:        0.5,
					RateUtilization: 70.0,
				},
			},
			ClassMetrics: []ClassPerformanceMetrics{
				{
					Handle:            "1:10",
					HTBBorrowingRatio: 0.3,
				},
			},
			LinkMetrics: LinkPerformanceMetrics{
				ErrorRate: 0.01,
			},
			Trends: TrendAnalysis{
				ThroughputTrend:  "stable",
				DropRateTrend:    "stable",
				TrendConfidence:  0.8,
			},
		}

		recommendations := service.generateRecommendations(metrics)
		assert.Contains(t, recommendations[0], "performing well")
	})

	t.Run("generate recommendations for problematic system", func(t *testing.T) {
		metrics := &PerformanceMetrics{
			HealthScore: HealthScore{
				OverallScore: 45.0,
			},
			OverallMetrics: OverallPerformanceMetrics{
				PacketLossRate: 5.0,
			},
			QdiscMetrics: []QdiscPerformanceMetrics{
				{
					Handle:          "1:",
					DropRate:        3.0,
					RateUtilization: 95.0,
				},
			},
			ClassMetrics: []ClassPerformanceMetrics{
				{
					Handle:            "1:10",
					HTBBorrowingRatio: 0.9,
				},
			},
			LinkMetrics: LinkPerformanceMetrics{
				ErrorRate: 0.5,
			},
			Trends: TrendAnalysis{
				ThroughputTrend:  "increasing",
				DropRateTrend:    "increasing",
				TrendConfidence:  0.8,
			},
		}

		recommendations := service.generateRecommendations(metrics)
		
		// Should have multiple recommendations
		assert.Greater(t, len(recommendations), 3)
		
		// Should identify specific issues
		foundSystemIssue := false
		foundPacketLoss := false
		foundQdiscIssue := false
		foundClassIssue := false
		foundLinkIssue := false
		foundTrendIssue := false
		
		for _, rec := range recommendations {
			if contains(rec, "System performance is below acceptable") {
				foundSystemIssue = true
			}
			if contains(rec, "High packet loss detected") {
				foundPacketLoss = true
			}
			if contains(rec, "has high drop rate") {
				foundQdiscIssue = true
			}
			if contains(rec, "near capacity") {
				foundQdiscIssue = true
			}
			if contains(rec, "frequently borrows bandwidth") {
				foundClassIssue = true
			}
			if contains(rec, "High error rate detected") {
				foundLinkIssue = true
			}
			if contains(rec, "Drop rate is increasing") {
				foundTrendIssue = true
			}
		}
		
		assert.True(t, foundSystemIssue, "Should recommend system review")
		assert.True(t, foundPacketLoss, "Should identify packet loss issue")
		assert.True(t, foundQdiscIssue, "Should identify qdisc issues")
		assert.True(t, foundClassIssue, "Should identify class borrowing issue")
		assert.True(t, foundLinkIssue, "Should identify link error issue")
		assert.True(t, foundTrendIssue, "Should identify increasing drop rate trend")
	})
}

func TestPerformanceMetricsService_ThroughputPrediction(t *testing.T) {
	service := &PerformanceMetricsService{
		logger: nil,
	}

	t.Run("predict throughput with linear trend", func(t *testing.T) {
		// Create data with clear linear trend
		data := make([]*timeseries.AggregatedData, 12)
		for i := 0; i < 12; i++ {
			data[i] = &timeseries.AggregatedData{
				LinkStats: timeseries.AggregatedLinkStats{
					// Throughput increases by 10000 each step
					AvgRxRate: uint64(100000 + i*10000),
					AvgTxRate: uint64(80000 + i*8000),
				},
			}
		}

		predicted := service.predictThroughput(data)
		
		// Should predict next value in sequence
		expectedRx := uint64(100000 + 12*10000) // 220000
		expectedTx := uint64(80000 + 12*8000)   // 176000
		expectedTotal := expectedRx + expectedTx // 396000
		
		// Allow some tolerance for floating point calculations
		assert.InDelta(t, expectedTotal, predicted, 5000)
	})

	t.Run("predict throughput with insufficient data", func(t *testing.T) {
		// Single data point
		data := []*timeseries.AggregatedData{
			{
				LinkStats: timeseries.AggregatedLinkStats{
					AvgRxRate: 100000,
					AvgTxRate: 80000,
				},
			},
		}

		predicted := service.predictThroughput(data)
		assert.Equal(t, uint64(0), predicted)
	})
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}