package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rng999/traffic-control-go/internal/application"
	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	"github.com/rng999/traffic-control-go/internal/infrastructure/timeseries"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

func main() {
	fmt.Println("WP4 Statistics and Monitoring System Demo")
	fmt.Println("=========================================")

	// Set up the statistics system
	ctx := context.Background()
	
	// Use in-memory storage for demo (in production, use SQLite)
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()

	// Initialize services
	historicalDataService := application.NewHistoricalDataService(timeSeriesStore)
	performanceMetricsService := application.NewPerformanceMetricsService(historicalDataService)
	exportService := application.NewStatisticsExportService(historicalDataService, performanceMetricsService)
	reportingService := application.NewStatisticsReportingService(historicalDataService, performanceMetricsService, exportService)
	dashboardDataService := application.NewDashboardDataService(historicalDataService, performanceMetricsService)

	deviceName, _ := tc.NewDeviceName("eth0")

	// Demo 1: Generate and store sample statistics data
	fmt.Println("\n1. Generating Sample Statistics Data")
	fmt.Println("-----------------------------------")
	
	if err := generateSampleData(ctx, historicalDataService, deviceName); err != nil {
		log.Fatal("Failed to generate sample data:", err)
	}
	fmt.Println("✓ Generated 24 hours of sample statistics data")

	// Demo 2: Data aggregation
	fmt.Println("\n2. Data Aggregation")
	fmt.Println("------------------")
	
	if err := demonstrateAggregation(ctx, historicalDataService, deviceName); err != nil {
		log.Fatal("Failed to demonstrate aggregation:", err)
	}

	// Demo 3: Performance metrics calculation
	fmt.Println("\n3. Performance Metrics Calculation")
	fmt.Println("---------------------------------")
	
	if err := demonstratePerformanceMetrics(ctx, performanceMetricsService, deviceName); err != nil {
		log.Fatal("Failed to demonstrate performance metrics:", err)
	}

	// Demo 4: Statistics export
	fmt.Println("\n4. Statistics Export")
	fmt.Println("-------------------")
	
	if err := demonstrateExport(ctx, exportService, deviceName); err != nil {
		log.Fatal("Failed to demonstrate export:", err)
	}

	// Demo 5: Report generation
	fmt.Println("\n5. Report Generation")
	fmt.Println("-------------------")
	
	if err := demonstrateReporting(ctx, reportingService, deviceName); err != nil {
		log.Fatal("Failed to demonstrate reporting:", err)
	}

	// Demo 6: Real-time dashboard data
	fmt.Println("\n6. Real-time Dashboard Data")
	fmt.Println("--------------------------")
	
	if err := demonstrateDashboard(ctx, dashboardDataService, deviceName); err != nil {
		log.Fatal("Failed to demonstrate dashboard:", err)
	}

	fmt.Println("\nDemo completed successfully!")
	fmt.Println("Check the generated files in the current directory for exported data.")
}

func generateSampleData(ctx context.Context, service *application.HistoricalDataService, deviceName tc.DeviceName) error {
	now := time.Now().Truncate(time.Minute)
	
	// Generate 24 hours of data (every 10 minutes)
	for hour := 0; hour < 24; hour++ {
		for minute := 0; minute < 60; minute += 10 {
			timestamp := now.Add(-time.Duration(23-hour)*time.Hour - time.Duration(minute)*time.Minute)
			
			// Create realistic data with daily patterns
			dailyFactor := 0.5 + 0.5*float64(hour)/24.0 // Higher activity during day
			variation := 1.0 + 0.2*(float64(minute)/60.0) // Small time-based variation
			
			stats := &application.DeviceStatistics{
				DeviceName: deviceName.String(),
				Timestamp:  timestamp,
				QdiscStats: []application.QdiscStatistics{
					{
						Handle: "1:",
						Type:   "htb",
						Stats: netlink.QdiscStats{
							BytesSent:    uint64(1000000 * dailyFactor * variation),
							PacketsSent:  uint64(1000 * dailyFactor * variation),
							BytesDropped: uint64(5 * dailyFactor),
							Overlimits:   uint64(10 * dailyFactor),
							Requeues:     uint64(2),
						},
					},
					{
						Handle: "2:",
						Type:   "fq",
						Stats: netlink.QdiscStats{
							BytesSent:   uint64(500000 * dailyFactor * variation),
							PacketsSent: uint64(500 * dailyFactor * variation),
						},
					},
				},
				ClassStats: []application.ClassStatistics{
					{
						Handle: "1:10",
						Parent: "1:",
						Type:   "htb",
						Stats: netlink.ClassStats{
							BytesSent:      uint64(750000 * dailyFactor * variation),
							PacketsSent:    uint64(750 * dailyFactor * variation),
							BytesDropped:   uint64(3 * dailyFactor),
							Overlimits:     uint64(6 * dailyFactor),
							RateBPS:        uint64(100000 * dailyFactor),
							BacklogBytes:   uint64(200),
							BacklogPackets: 20,
						},
						DetailedStats: &netlink.DetailedClassStats{
							HTBStats: &netlink.HTBClassStats{
								Lends:   15,
								Borrows: 8,
								Giants:  2,
								Tokens:  1500,
								CTokens: 750,
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
							Matches:    uint64(400 * dailyFactor * variation),
							Bytes:      uint64(400000 * dailyFactor * variation),
							Packets:    uint64(400 * dailyFactor * variation),
							Rate:       uint64(40000 * dailyFactor),
							PacketRate: uint64(40 * dailyFactor),
						},
					},
				},
				LinkStats: application.LinkStatistics{
					RxBytes:   uint64(3000000 * dailyFactor * variation),
					TxBytes:   uint64(2500000 * dailyFactor * variation),
					RxPackets: uint64(3000 * dailyFactor * variation),
					TxPackets: uint64(2500 * dailyFactor * variation),
					RxErrors:  2,
					TxErrors:  1,
					RxDropped: 3,
					TxDropped: 2,
					RxRate:    uint64(300000 * dailyFactor * variation),
					TxRate:    uint64(250000 * dailyFactor * variation),
				},
			}

			if err := service.StoreRawData(ctx, deviceName, stats); err != nil {
				return fmt.Errorf("failed to store data: %w", err)
			}
		}
	}
	
	return nil
}

func demonstrateAggregation(ctx context.Context, service *application.HistoricalDataService, deviceName tc.DeviceName) error {
	now := time.Now()
	
	// Perform hourly aggregation
	err := service.AggregateData(ctx, deviceName, timeseries.IntervalHour, 
		now.Add(-24*time.Hour), now)
	if err != nil {
		return fmt.Errorf("hourly aggregation failed: %w", err)
	}
	fmt.Println("✓ Hourly aggregation completed")

	// Perform daily aggregation
	err = service.AggregateData(ctx, deviceName, timeseries.IntervalDay, 
		now.Add(-24*time.Hour), now)
	if err != nil {
		return fmt.Errorf("daily aggregation failed: %w", err)
	}
	fmt.Println("✓ Daily aggregation completed")

	// Get aggregated data summary
	summary, err := service.GetDataSummary(ctx, deviceName)
	if err != nil {
		return fmt.Errorf("failed to get data summary: %w", err)
	}
	
	fmt.Printf("  Data Summary:\n")
	fmt.Printf("  - Total data points: %d\n", summary.TotalDataPoints)
	fmt.Printf("  - Time span: %v\n", summary.TimeSpan)
	fmt.Printf("  - Total bytes sent: %d\n", summary.TotalBytesSent)
	fmt.Printf("  - Total packets sent: %d\n", summary.TotalPacketsSent)
	
	return nil
}

func demonstratePerformanceMetrics(ctx context.Context, service *application.PerformanceMetricsService, deviceName tc.DeviceName) error {
	timeRange := application.TimeRange{
		Start: time.Now().Add(-6 * time.Hour),
		End:   time.Now(),
	}

	metrics, err := service.CalculatePerformanceMetrics(ctx, deviceName, timeRange)
	if err != nil {
		return fmt.Errorf("failed to calculate performance metrics: %w", err)
	}

	fmt.Printf("✓ Performance metrics calculated\n")
	fmt.Printf("  Overall Health Score: %.1f (%s)\n", 
		metrics.HealthScore.OverallScore, metrics.HealthScore.HealthStatus)
	fmt.Printf("  Total Throughput: %d bps\n", metrics.OverallMetrics.TotalThroughput)
	fmt.Printf("  Efficiency Score: %.1f\n", metrics.OverallMetrics.EfficiencyScore)
	fmt.Printf("  Packet Loss Rate: %.3f%%\n", metrics.OverallMetrics.PacketLossRate)
	fmt.Printf("  Throughput Trend: %s\n", metrics.Trends.ThroughputTrend)
	fmt.Printf("  Drop Rate Trend: %s\n", metrics.Trends.DropRateTrend)
	
	fmt.Printf("  Component Metrics:\n")
	for _, qdisc := range metrics.QdiscMetrics {
		fmt.Printf("    Qdisc %s (%s): %.1f Mbps, %.3f%% drop rate\n", 
			qdisc.Handle, qdisc.Type, qdisc.ThroughputMbps, qdisc.DropRate)
	}
	
	fmt.Printf("  Recommendations (%d):\n", len(metrics.Recommendations))
	for i, rec := range metrics.Recommendations {
		if i < 3 { // Show first 3 recommendations
			fmt.Printf("    - %s\n", rec)
		}
	}
	
	return nil
}

func demonstrateExport(ctx context.Context, service *application.StatisticsExportService, deviceName tc.DeviceName) error {
	now := time.Now()
	
	// Export in JSON format
	jsonOptions := application.ExportOptions{
		Format:         application.FormatJSON,
		DeviceName:     deviceName,
		StartTime:      now.Add(-2 * time.Hour),
		EndTime:        now,
		IncludeRawData: true,
		IncludeMetrics: true,
	}

	jsonResult, err := service.ExportStatistics(ctx, jsonOptions)
	if err != nil {
		return fmt.Errorf("JSON export failed: %w", err)
	}

	// Write to file
	if err := os.WriteFile("demo_statistics.json", jsonResult.Data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}
	fmt.Printf("✓ JSON export: %d data points, %d bytes -> demo_statistics.json\n", 
		jsonResult.DataPoints, jsonResult.Size)

	// Export in CSV format
	csvOptions := application.ExportOptions{
		Format:     application.FormatCSV,
		DeviceName: deviceName,
		StartTime:  now.Add(-1 * time.Hour),
		EndTime:    now,
	}

	csvResult, err := service.ExportStatistics(ctx, csvOptions)
	if err != nil {
		return fmt.Errorf("CSV export failed: %w", err)
	}

	// Write to file
	if err := os.WriteFile("demo_statistics.csv", csvResult.Data, 0644); err != nil {
		return fmt.Errorf("failed to write CSV file: %w", err)
	}
	fmt.Printf("✓ CSV export: %d data points, %d bytes -> demo_statistics.csv\n", 
		csvResult.DataPoints, csvResult.Size)

	// Export in Prometheus format
	prometheusOptions := application.ExportOptions{
		Format:           application.FormatPrometheus,
		DeviceName:       deviceName,
		StartTime:        now.Add(-30 * time.Minute),
		EndTime:          now,
		IncludeMetrics:   true,
		PrometheusPrefix: "demo_tc_",
		PrometheusLabels: map[string]string{
			"environment": "demo",
			"region":      "local",
		},
	}

	prometheusResult, err := service.ExportStatistics(ctx, prometheusOptions)
	if err != nil {
		return fmt.Errorf("Prometheus export failed: %w", err)
	}

	// Write to file
	if err := os.WriteFile("demo_statistics.prom", prometheusResult.Data, 0644); err != nil {
		return fmt.Errorf("failed to write Prometheus file: %w", err)
	}
	fmt.Printf("✓ Prometheus export: %d data points, %d bytes -> demo_statistics.prom\n", 
		prometheusResult.DataPoints, prometheusResult.Size)

	return nil
}

func demonstrateReporting(ctx context.Context, service *application.StatisticsReportingService, deviceName tc.DeviceName) error {
	timeRange := application.TimeRange{
		Start: time.Now().Add(-6 * time.Hour),
		End:   time.Now(),
	}

	// Generate a comprehensive report
	options := application.ReportOptions{
		Type:          application.ReportTypeDetailed,
		DeviceName:    deviceName,
		TimeRange:     timeRange,
		Format:        application.FormatJSON,
		IncludeCharts: true,
	}

	report, err := service.GenerateReport(ctx, options)
	if err != nil {
		return fmt.Errorf("report generation failed: %w", err)
	}

	fmt.Printf("✓ Detailed report generated\n")
	fmt.Printf("  Title: %s\n", report.Metadata.Title)
	fmt.Printf("  Data Points: %d\n", report.Metadata.DataPoints)
	fmt.Printf("  Analysis Period: %v\n", report.Metadata.AnalysisPeriod)
	
	if report.ExecutiveSummary != nil {
		fmt.Printf("  Executive Summary:\n")
		fmt.Printf("    Health Score: %.1f (%s)\n", 
			report.ExecutiveSummary.HealthScore, report.ExecutiveSummary.OverallHealth)
		fmt.Printf("    Key Findings: %d\n", len(report.ExecutiveSummary.KeyFindings))
		fmt.Printf("    Critical Issues: %d\n", len(report.ExecutiveSummary.CriticalIssues))
		fmt.Printf("    Data Quality Score: %.1f\n", report.ExecutiveSummary.DataQuality.QualityScore)
	}
	
	if report.TrendAnalysis != nil {
		fmt.Printf("  Trend Analysis:\n")
		fmt.Printf("    Overall Trends: %d\n", len(report.TrendAnalysis.OverallTrends))
		fmt.Printf("    Anomalies Detected: %d\n", len(report.TrendAnalysis.AnomalyDetection))
		fmt.Printf("    Forecast Projections: %d\n", len(report.TrendAnalysis.ForecastProjections))
	}
	
	fmt.Printf("  Recommendations: %d\n", len(report.Recommendations))
	fmt.Printf("  Charts: %d\n", len(report.Charts))

	return nil
}

func demonstrateDashboard(ctx context.Context, service *application.DashboardDataService, deviceName tc.DeviceName) error {
	// Get real-time metrics
	metrics, err := service.GetRealTimeMetrics(ctx, deviceName)
	if err != nil {
		return fmt.Errorf("failed to get real-time metrics: %w", err)
	}

	fmt.Printf("✓ Real-time metrics retrieved\n")
	fmt.Printf("  Overall Health: %s (%.1f)\n", 
		metrics.OverallHealth.Status, metrics.OverallHealth.Score)
	fmt.Printf("  Current Throughput: Rx %.1f Mbps, Tx %.1f Mbps\n", 
		float64(metrics.ThroughputMetrics.CurrentRxRate)/1e6, 
		float64(metrics.ThroughputMetrics.CurrentTxRate)/1e6)
	fmt.Printf("  Utilization: Rx %.1f%%, Tx %.1f%%\n", 
		metrics.ThroughputMetrics.UtilizationRx, metrics.ThroughputMetrics.UtilizationTx)
	fmt.Printf("  Quality Score: %.1f\n", metrics.QualityMetrics.QualityScore)
	fmt.Printf("  Error Count: %d (%s trend)\n", 
		metrics.ErrorCounts.TotalErrors, metrics.ErrorCounts.ErrorTrend)

	// Get trend data
	trends, err := service.GetTrendData(ctx, deviceName, 2*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to get trend data: %w", err)
	}

	fmt.Printf("  Trend Analysis (2h window):\n")
	fmt.Printf("    Throughput: %s (%.1f%% change)\n", 
		trends.ThroughputTrend.Direction, trends.ThroughputTrend.Magnitude)
	fmt.Printf("    Error Rate: %s (%.1f%% change)\n", 
		trends.ErrorRateTrend.Direction, trends.ErrorRateTrend.Magnitude)
	fmt.Printf("    Trend Confidence: %.1f%%\n", trends.TrendConfidence*100)

	// Get dashboard update
	update, err := service.GetDashboardUpdate(ctx, []tc.DeviceName{deviceName})
	if err != nil {
		return fmt.Errorf("failed to get dashboard update: %w", err)
	}

	fmt.Printf("  Dashboard Update:\n")
	fmt.Printf("    Update ID: %s\n", update.UpdateID)
	fmt.Printf("    Devices: %d\n", len(update.DeviceUpdates))
	fmt.Printf("    Global Alerts: %d\n", len(update.GlobalAlerts))
	fmt.Printf("    Update Duration: %v\n", update.UpdateMetadata.UpdateDuration)

	return nil
}