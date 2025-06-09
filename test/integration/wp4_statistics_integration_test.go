package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rng999/traffic-control-go/internal/application"
	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	"github.com/rng999/traffic-control-go/internal/infrastructure/timeseries"
	"github.com/rng999/traffic-control-go/pkg/tc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWP4StatisticsIntegration(t *testing.T) {
	// Skip integration tests in short mode
	if testing.Short() {
		t.Skip("Skipping WP4 statistics integration tests in short mode")
	}

	// Create temporary directory for SQLite database
	tmpDir, err := os.MkdirTemp("", "wp4_integration_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test_statistics.db")

	// Set up the complete WP4 stack
	timeSeriesStore, err := timeseries.NewSQLiteTimeSeriesStore(dbPath)
	require.NoError(t, err)
	defer timeSeriesStore.Close()

	historicalDataService := application.NewHistoricalDataService(timeSeriesStore)
	performanceMetricsService := application.NewPerformanceMetricsService(historicalDataService)
	exportService := application.NewStatisticsExportService(historicalDataService, performanceMetricsService)
	reportingService := application.NewStatisticsReportingService(historicalDataService, performanceMetricsService, exportService)
	dashboardDataService := application.NewDashboardDataService(historicalDataService, performanceMetricsService)

	ctx := context.Background()
	deviceNames := []tc.DeviceName{}
	
	// Create multiple test devices
	for i := 0; i < 3; i++ {
		deviceName, _ := tc.NewDeviceName(fmt.Sprintf("eth%d", i))
		deviceNames = append(deviceNames, deviceName)
	}

	t.Run("end-to-end data flow", func(t *testing.T) {
		// Phase 1: Generate comprehensive test data
		now := time.Now().Truncate(time.Minute)
		totalDataPoints := 0

		for deviceIdx, deviceName := range deviceNames {
			// Generate 24 hours of data (every 10 minutes = 144 data points per device)
			for hour := 0; hour < 24; hour++ {
				for minute := 0; minute < 60; minute += 10 {
					timestamp := now.Add(-time.Duration(23-hour)*time.Hour - time.Duration(minute)*time.Minute)
					
					// Create realistic data with patterns and variations
					baseMultiplier := float64(deviceIdx + 1)
					timeMultiplier := float64(hour) / 24.0 // Simulate daily pattern
					variation := 1.0 + 0.3*timeMultiplier  // 30% variation throughout day

					stats := &application.DeviceStatistics{
						DeviceName: deviceName.String(),
						Timestamp:  timestamp,
						QdiscStats: []application.QdiscStatistics{
							{
								Handle: "1:",
								Type:   "htb",
								Stats: netlink.QdiscStats{
									BytesSent:    uint64(1000000 * baseMultiplier * variation),
									PacketsSent:  uint64(1000 * baseMultiplier * variation),
									BytesDropped: uint64(5 * baseMultiplier * (1 + timeMultiplier/2)),
									Overlimits:   uint64(10 * baseMultiplier * (1 + timeMultiplier/3)),
									Requeues:     uint64(2 * baseMultiplier),
								},
							},
							{
								Handle: "2:",
								Type:   "fq",
								Stats: netlink.QdiscStats{
									BytesSent:   uint64(500000 * baseMultiplier * variation),
									PacketsSent: uint64(500 * baseMultiplier * variation),
								},
							},
						},
						ClassStats: []application.ClassStatistics{
							{
								Handle: "1:10",
								Parent: "1:",
								Type:   "htb",
								Stats: netlink.ClassStats{
									BytesSent:      uint64(750000 * baseMultiplier * variation),
									PacketsSent:    uint64(750 * baseMultiplier * variation),
									BytesDropped:   uint64(3 * baseMultiplier),
									Overlimits:     uint64(6 * baseMultiplier),
									RateBPS:        uint64(100000 * baseMultiplier),
									BacklogBytes:   uint64(200 * baseMultiplier),
									BacklogPackets: uint64(20 * baseMultiplier),
								},
								DetailedStats: &netlink.DetailedClassStats{
									HTBStats: &netlink.HTBClassStats{
										Lends:   uint32(15 * baseMultiplier),
										Borrows: uint32(8 * baseMultiplier),
										Giants:  uint32(2 * baseMultiplier),
										Tokens:  uint32(1500 * baseMultiplier),
										CTokens: uint32(750 * baseMultiplier),
									},
								},
							},
						},
						FilterStats: []application.FilterStatistics{
							{
								Handle:   "800:100",
								Parent:   "1:",
								Priority: 100,
								FlowID:   "1:10",
								Stats: application.FilterStats{
									Matches:    uint64(400 * baseMultiplier * variation),
									Bytes:      uint64(400000 * baseMultiplier * variation),
									Packets:    uint64(400 * baseMultiplier * variation),
									Rate:       uint64(40000 * baseMultiplier),
									PacketRate: uint64(40 * baseMultiplier),
								},
							},
						},
						LinkStats: application.LinkStatistics{
							RxBytes:   uint64(3000000 * baseMultiplier * variation),
							TxBytes:   uint64(2500000 * baseMultiplier * variation),
							RxPackets: uint64(3000 * baseMultiplier * variation),
							TxPackets: uint64(2500 * baseMultiplier * variation),
							RxErrors:  uint64(2 * baseMultiplier),
							TxErrors:  uint64(1 * baseMultiplier),
							RxDropped: uint64(3 * baseMultiplier),
							TxDropped: uint64(2 * baseMultiplier),
							RxRate:    uint64(300000 * baseMultiplier * variation),
							TxRate:    uint64(250000 * baseMultiplier * variation),
						},
					}

					err := historicalDataService.StoreRawData(ctx, deviceName, stats)
					require.NoError(t, err)
					totalDataPoints++
				}
			}
		}

		t.Logf("Generated %d data points across %d devices", totalDataPoints, len(deviceNames))

		// Phase 2: Test data aggregation
		t.Run("data aggregation", func(t *testing.T) {
			for _, deviceName := range deviceNames {
				// Perform aggregation for different intervals
				err := historicalDataService.AggregateData(ctx, deviceName, timeseries.IntervalHour, 
					now.Add(-24*time.Hour), now)
				require.NoError(t, err)

				err = historicalDataService.AggregateData(ctx, deviceName, timeseries.IntervalDay, 
					now.Add(-24*time.Hour), now)
				require.NoError(t, err)

				// Verify aggregated data exists
				hourlyData, err := historicalDataService.GetHistoricalData(ctx, deviceName, 
					now.Add(-24*time.Hour), now, timeseries.IntervalHour)
				require.NoError(t, err)
				assert.Greater(t, len(hourlyData), 0, "Should have hourly aggregated data")

				dailyData, err := historicalDataService.GetHistoricalData(ctx, deviceName, 
					now.Add(-24*time.Hour), now, timeseries.IntervalDay)
				require.NoError(t, err)
				assert.Greater(t, len(dailyData), 0, "Should have daily aggregated data")
			}
		})

		// Phase 3: Test performance metrics calculation
		t.Run("performance metrics", func(t *testing.T) {
			for _, deviceName := range deviceNames {
				timeRange := application.TimeRange{
					Start: now.Add(-24 * time.Hour),
					End:   now,
				}

				metrics, err := performanceMetricsService.CalculatePerformanceMetrics(ctx, deviceName, timeRange)
				require.NoError(t, err)

				// Verify metrics structure
				assert.Equal(t, deviceName.String(), metrics.DeviceName)
				assert.Greater(t, metrics.OverallMetrics.TotalThroughput, uint64(0))
				assert.GreaterOrEqual(t, metrics.HealthScore.OverallScore, float64(0))
				assert.LessOrEqual(t, metrics.HealthScore.OverallScore, float64(100))
				assert.NotEmpty(t, metrics.QdiscMetrics)
				assert.NotEmpty(t, metrics.ClassMetrics)
				assert.NotEmpty(t, metrics.FilterMetrics)
				assert.Contains(t, []string{"increasing", "decreasing", "stable"}, metrics.Trends.ThroughputTrend)

				t.Logf("Device %s: Health Score %.1f, Status %s, Throughput %d bps", 
					deviceName.String(), metrics.HealthScore.OverallScore, 
					metrics.HealthScore.HealthStatus, metrics.OverallMetrics.TotalThroughput)
			}
		})

		// Phase 4: Test all export formats
		t.Run("statistics export", func(t *testing.T) {
			for _, format := range []application.ExportFormat{
				application.FormatJSON,
				application.FormatCSV,
				application.FormatPrometheus,
			} {
				t.Run(string(format), func(t *testing.T) {
					deviceName := deviceNames[0] // Test with first device

					options := application.ExportOptions{
						Format:         format,
						DeviceName:     deviceName,
						StartTime:      now.Add(-6 * time.Hour),
						EndTime:        now,
						IncludeRawData: true,
						IncludeMetrics: true,
					}

					if format == application.FormatPrometheus {
						options.PrometheusPrefix = "integration_test_"
						options.PrometheusLabels = map[string]string{
							"test_env": "integration",
						}
					}

					result, err := exportService.ExportStatistics(ctx, options)
					require.NoError(t, err)

					assert.Equal(t, format, result.Format)
					assert.Greater(t, result.DataPoints, 0)
					assert.Greater(t, result.Size, int64(0))
					assert.NotEmpty(t, result.Data)

					t.Logf("Exported %d data points in %s format (%d bytes)", 
						result.DataPoints, format, result.Size)
				})
			}
		})

		// Phase 5: Test comprehensive reporting
		t.Run("statistics reporting", func(t *testing.T) {
			deviceName := deviceNames[0]
			timeRange := application.TimeRange{
				Start: now.Add(-12 * time.Hour),
				End:   now,
			}

			// Test different report types
			reportTypes := []application.ReportType{
				application.ReportTypeSummary,
				application.ReportTypeDetailed,
				application.ReportTypeTrendAnalysis,
				application.ReportTypeCapacityPlanning,
				application.ReportTypePerformanceAudit,
			}

			for _, reportType := range reportTypes {
				t.Run(string(reportType), func(t *testing.T) {
					options := application.ReportOptions{
						Type:          reportType,
						DeviceName:    deviceName,
						TimeRange:     timeRange,
						Format:        application.FormatJSON,
						IncludeCharts: true,
					}

					report, err := reportingService.GenerateReport(ctx, options)
					require.NoError(t, err)

					assert.Equal(t, deviceName.String(), report.Metadata.DeviceName)
					assert.Equal(t, reportType, report.Metadata.ReportType)
					assert.Greater(t, report.Metadata.DataPoints, 0)
					assert.NotNil(t, report.ExecutiveSummary)

					switch reportType {
					case application.ReportTypeDetailed, application.ReportTypePerformanceAudit:
						assert.NotNil(t, report.PerformanceMetrics)
						assert.NotNil(t, report.TrendAnalysis)
						assert.NotEmpty(t, report.Charts)
					case application.ReportTypeTrendAnalysis:
						assert.NotNil(t, report.TrendAnalysis)
						assert.NotEmpty(t, report.TrendAnalysis.OverallTrends)
					case application.ReportTypeCapacityPlanning:
						assert.NotNil(t, report.CapacityAnalysis)
						assert.NotEmpty(t, report.CapacityAnalysis.GrowthProjections)
					}

					assert.NotEmpty(t, report.Recommendations)

					t.Logf("Generated %s report with %d data points, health score: %.1f", 
						reportType, report.Metadata.DataPoints, 
						report.ExecutiveSummary.HealthScore)
				})
			}
		})

		// Phase 6: Test real-time dashboard data
		t.Run("dashboard data", func(t *testing.T) {
			// Test individual device metrics
			for _, deviceName := range deviceNames {
				metrics, err := dashboardDataService.GetRealTimeMetrics(ctx, deviceName)
				require.NoError(t, err)

				assert.Equal(t, deviceName.String(), metrics.DeviceName)
				assert.NotZero(t, metrics.Timestamp)
				assert.NotEmpty(t, metrics.OverallHealth.Status)
				assert.GreaterOrEqual(t, metrics.OverallHealth.Score, float64(0))
				assert.LessOrEqual(t, metrics.OverallHealth.Score, float64(100))

				// Test trend data
				trends, err := dashboardDataService.GetTrendData(ctx, deviceName, 2*time.Hour)
				require.NoError(t, err)
				assert.Equal(t, 2*time.Hour, trends.TimeWindow)
			}

			// Test multi-device dashboard update
			update, err := dashboardDataService.GetDashboardUpdate(ctx, deviceNames)
			require.NoError(t, err)

			assert.Len(t, update.DeviceUpdates, len(deviceNames))
			assert.NotNil(t, update.SystemSummary)
			assert.Equal(t, len(deviceNames), update.SystemSummary.TotalDevices)
			assert.Greater(t, update.SystemSummary.TotalThroughput, uint64(0))

			t.Logf("Dashboard update: %d devices, system health: %s (%.1f)", 
				update.SystemSummary.TotalDevices, 
				update.SystemSummary.SystemHealth.Status,
				update.SystemSummary.SystemHealth.Score)
		})

		// Phase 7: Test data cleanup and retention
		t.Run("data cleanup", func(t *testing.T) {
			deviceName := deviceNames[0]

			// Get initial data count
			initialSummary, err := historicalDataService.GetDataSummary(ctx, deviceName)
			require.NoError(t, err)
			initialCount := initialSummary.TotalDataPoints

			// Cleanup data older than 12 hours
			retentionPolicy := timeseries.RetentionPolicy{
				RawDataRetention: 12 * time.Hour,
			}
			err = historicalDataService.CleanupOldData(ctx, deviceName, retentionPolicy)
			require.NoError(t, err)

			// Verify cleanup
			finalSummary, err := historicalDataService.GetDataSummary(ctx, deviceName)
			require.NoError(t, err)
			finalCount := finalSummary.TotalDataPoints

			assert.Less(t, finalCount, initialCount, "Data count should decrease after cleanup")

			t.Logf("Cleanup: %d -> %d data points (removed %d old entries)", 
				initialCount, finalCount, initialCount-finalCount)
		})
	})

	t.Run("concurrent operations", func(t *testing.T) {
		// Test concurrent read/write operations
		deviceName := deviceNames[0]
		now := time.Now()

		// Start concurrent operations
		errChan := make(chan error, 10)

		// Concurrent writes
		go func() {
			for i := 0; i < 5; i++ {
				stats := &application.DeviceStatistics{
					DeviceName: deviceName.String(),
					Timestamp:  now.Add(-time.Duration(i) * time.Minute),
					QdiscStats: []application.QdiscStatistics{
						{
							Handle: "test:",
							Type:   "htb",
							Stats:  netlink.QdiscStats{BytesSent: uint64(i * 1000)},
						},
					},
				}
				err := historicalDataService.StoreRawData(ctx, deviceName, stats)
				errChan <- err
			}
		}()

		// Concurrent reads
		go func() {
			for i := 0; i < 5; i++ {
				_, err := historicalDataService.GetHistoricalData(ctx, deviceName, 
					now.Add(-1*time.Hour), now, "")
				errChan <- err
			}
		}()

		// Wait for all operations
		for i := 0; i < 10; i++ {
			err := <-errChan
			assert.NoError(t, err)
		}
	})

	t.Run("performance under load", func(t *testing.T) {
		deviceName := deviceNames[0]
		startTime := time.Now()

		// Generate large batch of operations
		const batchSize = 100
		
		for i := 0; i < batchSize; i++ {
			stats := &application.DeviceStatistics{
				DeviceName: deviceName.String(),
				Timestamp:  time.Now().Add(-time.Duration(i) * time.Second),
				QdiscStats: []application.QdiscStatistics{
					{
						Handle: "perf:",
						Type:   "htb",
						Stats:  netlink.QdiscStats{BytesSent: uint64(i * 1000)},
					},
				},
			}
			err := historicalDataService.StoreRawData(ctx, deviceName, stats)
			require.NoError(t, err)
		}

		duration := time.Since(startTime)
		throughput := float64(batchSize) / duration.Seconds()
		
		t.Logf("Performance: %d operations in %v (%.1f ops/sec)", 
			batchSize, duration, throughput)

		// Performance should be reasonable (at least 50 ops/sec)
		assert.Greater(t, throughput, float64(50), "Should maintain reasonable throughput")
	})

	t.Run("error handling and recovery", func(t *testing.T) {
		deviceName := deviceNames[0]

		// Test invalid data handling
		invalidStats := &application.DeviceStatistics{
			DeviceName: "", // Invalid empty device name
			Timestamp:  time.Time{}, // Invalid zero timestamp
		}
		err := historicalDataService.StoreRawData(ctx, deviceName, invalidStats)
		assert.Error(t, err, "Should reject invalid data")

		// Test invalid time ranges
		_, err = historicalDataService.GetHistoricalData(ctx, deviceName, 
			time.Now(), time.Now().Add(-1*time.Hour), "") // Start after end
		assert.Error(t, err, "Should reject invalid time range")

		// Test non-existent device
		nonExistentDevice, _ := tc.NewDeviceName("nonexistent999")
		summary, err := historicalDataService.GetDataSummary(ctx, nonExistentDevice)
		require.NoError(t, err)
		assert.False(t, summary.HasData, "Should handle non-existent device gracefully")
	})
}

func TestWP4MemoryTimeSeriesIntegration(t *testing.T) {
	// Test the complete stack with memory time series store
	if testing.Short() {
		t.Skip("Skipping memory integration tests in short mode")
	}

	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()

	historicalDataService := application.NewHistoricalDataService(timeSeriesStore)
	
	ctx := context.Background()
	deviceName, _ := tc.NewDeviceName("memory_test")

	// Test rapid data generation and processing
	t.Run("memory store rapid operations", func(t *testing.T) {
		const rapidBatchSize = 50
		startTime := time.Now()

		// Rapid data insertion
		for i := 0; i < rapidBatchSize; i++ {
			stats := &application.DeviceStatistics{
				DeviceName: deviceName.String(),
				Timestamp:  time.Now().Add(-time.Duration(i) * time.Second),
				QdiscStats: []application.QdiscStatistics{
					{
						Handle: "rapid:",
						Type:   "fq",
						Stats:  netlink.QdiscStats{BytesSent: uint64(i * 2000)},
					},
				},
			}
			err := historicalDataService.StoreRawData(ctx, deviceName, stats)
			require.NoError(t, err)
		}

		// Immediate retrieval
		data, err := historicalDataService.GetHistoricalData(ctx, deviceName, 
			startTime.Add(-1*time.Minute), time.Now().Add(1*time.Minute), "")
		require.NoError(t, err)
		assert.Equal(t, rapidBatchSize, len(data))

		duration := time.Since(startTime)
		t.Logf("Memory store: %d operations in %v", rapidBatchSize, duration)
	})
}