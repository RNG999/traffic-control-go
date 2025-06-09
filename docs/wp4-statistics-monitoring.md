# WP4: Statistics and Monitoring System

## Overview

The WP4 Statistics and Monitoring System provides comprehensive data collection, analysis, and reporting capabilities for the traffic control library. It implements a sophisticated time-series database, performance metrics calculation, trend analysis, and real-time monitoring features.

## Architecture

### Core Components

1. **Time-Series Storage** (`timeseries/`)
   - SQLite and in-memory implementations
   - Automatic data aggregation (minute/hour/day/week/month)
   - Configurable retention policies
   - WAL mode optimization for SQLite

2. **Historical Data Service** (`application/historical_data_service.go`)
   - Raw data storage and retrieval
   - Data aggregation management
   - Data cleanup and retention
   - Data quality assessment

3. **Performance Metrics Service** (`application/performance_metrics_service.go`)
   - Comprehensive performance analysis
   - Health scoring algorithms
   - Trend detection and prediction
   - Component-level metrics

4. **Statistics Export Service** (`application/statistics_export_service.go`)
   - Multi-format exports (JSON, CSV, Prometheus)
   - Data integrity verification
   - Size estimation and validation

5. **Statistics Reporting Service** (`application/statistics_reporting_service.go`)
   - Advanced report generation
   - Anomaly detection
   - Correlation analysis
   - Forecasting capabilities

6. **Dashboard Data Service** (`application/dashboard_data_service.go`)
   - Real-time metrics preparation
   - Live dashboard updates
   - Alert generation
   - Caching for performance

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "time"
    
    "github.com/rng999/traffic-control-go/internal/application"
    "github.com/rng999/traffic-control-go/internal/infrastructure/timeseries"
    "github.com/rng999/traffic-control-go/pkg/tc"
)

func main() {
    ctx := context.Background()
    
    // Initialize storage
    store := timeseries.NewMemoryTimeSeriesStore()
    defer store.Close()
    
    // Initialize services
    historical := application.NewHistoricalDataService(store)
    performance := application.NewPerformanceMetricsService(historical)
    
    deviceName, _ := tc.NewDeviceName("eth0")
    
    // Store statistics data
    stats := &application.DeviceStatistics{
        DeviceName: deviceName.String(),
        Timestamp:  time.Now(),
        // ... populate with actual statistics
    }
    
    err := historical.StoreRawData(ctx, deviceName, stats)
    if err != nil {
        panic(err)
    }
    
    // Calculate performance metrics
    timeRange := application.TimeRange{
        Start: time.Now().Add(-1 * time.Hour),
        End:   time.Now(),
    }
    
    metrics, err := performance.CalculatePerformanceMetrics(ctx, deviceName, timeRange)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Health Score: %.1f\n", metrics.HealthScore.OverallScore)
}
```

### Using SQLite Storage

```go
// For production use, prefer SQLite storage
store, err := timeseries.NewSQLiteTimeSeriesStore("/path/to/database.db")
if err != nil {
    panic(err)
}
defer store.Close()
```

## Key Features

### 1. Time-Series Data Storage

- **Dual Storage Options**: In-memory (development) and SQLite (production)
- **Automatic Aggregation**: Data is automatically aggregated at multiple intervals
- **Retention Policies**: Configurable data retention with automatic cleanup
- **Performance Optimized**: WAL mode, connection pooling, and caching

```go
// Aggregate data by hour
err := historicalService.AggregateData(ctx, deviceName, 
    timeseries.IntervalHour, startTime, endTime)

// Clean up old data (keep last 7 days)
err = historicalService.CleanupOldData(ctx, deviceName, 7*24*time.Hour)
```

### 2. Performance Metrics

The system calculates comprehensive performance metrics including:

- **Overall Health Score** (0-100)
- **Throughput Analysis** (current, peak, utilization)
- **Quality Metrics** (packet loss, error rates, efficiency)
- **Component-level Metrics** (per qdisc, class, filter)
- **Trend Analysis** (increasing, decreasing, stable)

```go
metrics, err := performanceService.CalculatePerformanceMetrics(ctx, deviceName, timeRange)

fmt.Printf("Health: %s (%.1f)\n", metrics.HealthScore.HealthStatus, metrics.HealthScore.OverallScore)
fmt.Printf("Throughput: %d bps\n", metrics.OverallMetrics.TotalThroughput)
fmt.Printf("Efficiency: %.1f\n", metrics.OverallMetrics.EfficiencyScore)
```

### 3. Data Export

Export data in multiple formats for integration with external systems:

```go
exportService := application.NewStatisticsExportService(historical, performance)

// Export as JSON
jsonOptions := application.ExportOptions{
    Format:         application.FormatJSON,
    DeviceName:     deviceName,
    StartTime:      startTime,
    EndTime:        endTime,
    IncludeMetrics: true,
}

result, err := exportService.ExportStatistics(ctx, jsonOptions)

// Export as Prometheus metrics
promOptions := application.ExportOptions{
    Format:           application.FormatPrometheus,
    DeviceName:       deviceName,
    StartTime:        startTime,
    EndTime:          endTime,
    PrometheusPrefix: "tc_",
    PrometheusLabels: map[string]string{"env": "prod"},
}
```

### 4. Advanced Reporting

Generate comprehensive reports with trend analysis and recommendations:

```go
reportingService := application.NewStatisticsReportingService(historical, performance, exportService)

options := application.ReportOptions{
    Type:          application.ReportTypeDetailed,
    DeviceName:    deviceName,
    TimeRange:     timeRange,
    Format:        application.FormatJSON,
    IncludeCharts: true,
}

report, err := reportingService.GenerateReport(ctx, options)

// Access different report sections
fmt.Printf("Health Score: %.1f\n", report.ExecutiveSummary.HealthScore)
fmt.Printf("Anomalies: %d\n", len(report.TrendAnalysis.AnomalyDetection))
fmt.Printf("Recommendations: %d\n", len(report.Recommendations))
```

**Report Types:**
- `ReportTypeSummary`: Executive overview
- `ReportTypeDetailed`: Comprehensive analysis
- `ReportTypeTrendAnalysis`: Focus on trends and forecasting
- `ReportTypeCapacityPlanning`: Capacity and growth analysis
- `ReportTypeComparative`: Compare multiple time periods
- `ReportTypePerformanceAudit`: Complete performance audit

### 5. Real-time Dashboard Data

Prepare data for real-time monitoring dashboards:

```go
dashboardService := application.NewDashboardDataService(historical, performance)

// Get real-time metrics
metrics, err := dashboardService.GetRealTimeMetrics(ctx, deviceName)

// Get trend data for charts
trends, err := dashboardService.GetTrendData(ctx, deviceName, 2*time.Hour)

// Get complete dashboard update
devices := []tc.DeviceName{deviceName}
update, err := dashboardService.GetDashboardUpdate(ctx, devices)

// Start live updates with callback
updateCallback := func(update *application.DashboardUpdate) {
    // Process live update
    fmt.Printf("Live update: %d devices\n", len(update.DeviceUpdates))
}

err = dashboardService.StartLiveUpdates(ctx, devices, updateCallback)
```

## Configuration

### Retention Policies

```go
policy := timeseries.RetentionPolicy{
    RawDataRetention:    7 * 24 * time.Hour,     // 7 days
    MinuteDataRetention: 30 * 24 * time.Hour,    // 30 days  
    HourDataRetention:   90 * 24 * time.Hour,    // 90 days
    DayDataRetention:    365 * 24 * time.Hour,   // 1 year
    WeekDataRetention:   2 * 365 * 24 * time.Hour, // 2 years
}
```

### SQLite Optimization

The SQLite implementation includes several optimizations:

- **WAL Mode**: Write-Ahead Logging for better concurrency
- **Connection Pooling**: Efficient connection management
- **Prepared Statements**: Improved query performance
- **Batch Operations**: Bulk inserts for aggregation
- **Indexing**: Optimized indexes for time-series queries

## Performance Metrics Explained

### Health Score Calculation

The health score (0-100) is calculated based on:

- **Throughput Score** (25%): Based on utilization and efficiency
- **Reliability Score** (25%): Based on packet loss and latency
- **Efficiency Score** (25%): Based on queue performance and resource usage
- **Stability Score** (25%): Based on trend consistency and variability

### Component Metrics

**Qdisc Metrics:**
- Drop rate percentage
- Throughput in Mbps
- Queue efficiency
- Rate utilization
- Performance score

**Class Metrics:**
- Bandwidth utilization
- HTB lending/borrowing ratios
- Fairness score
- Token bucket efficiency

**Filter Metrics:**
- Match rate and efficiency
- Classification accuracy
- Bandwidth allocation

**Link Metrics:**
- Duplex efficiency
- Error rates
- Drop rates
- Utilization percentages

## Trend Analysis Features

### Anomaly Detection

The system uses statistical methods to detect anomalies:

- **2-sigma and 3-sigma outlier detection**
- **Moving average deviation analysis**
- **Trend change point detection**
- **Seasonal pattern recognition**

### Forecasting

Simple linear extrapolation and more sophisticated methods:

- **Linear regression for trend prediction**
- **Seasonal pattern projection**
- **Confidence intervals for predictions**
- **Multiple time horizon forecasts**

### Correlation Analysis

Identifies relationships between metrics:

- **Pearson correlation coefficients**
- **Cross-metric trend analysis**
- **Causality identification**
- **Performance factor correlation**

## Integration Examples

### Prometheus Integration

```bash
# Scrape endpoint
curl http://localhost:8080/metrics

# Example metrics
tc_qdisc_bytes_total{device="eth0",handle="1:",type="htb"} 1000000 1640995200000
tc_qdisc_packets_total{device="eth0",handle="1:",type="htb"} 1000 1640995200000
tc_health_score{device="eth0"} 85.5 1640995200000
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Traffic Control Monitoring",
    "panels": [
      {
        "title": "Health Score",
        "type": "stat",
        "targets": [
          {
            "expr": "tc_health_score",
            "legendFormat": "{{device}}"
          }
        ]
      },
      {
        "title": "Throughput",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(tc_qdisc_bytes_total[5m]) * 8",
            "legendFormat": "{{device}} - {{handle}}"
          }
        ]
      }
    ]
  }
}
```

### REST API Integration

```go
// HTTP handler for metrics endpoint
func metricsHandler(w http.ResponseWriter, r *http.Request) {
    deviceName, _ := tc.NewDeviceName(r.URL.Query().Get("device"))
    
    options := application.ExportOptions{
        Format:     application.FormatPrometheus,
        DeviceName: deviceName,
        StartTime:  time.Now().Add(-1 * time.Hour),
        EndTime:    time.Now(),
    }
    
    result, err := exportService.ExportStatistics(ctx, options)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "text/plain")
    w.Write(result.Data)
}
```

## Testing

### Unit Tests

```bash
# Run all WP4 tests
go test ./internal/application/...-run TestWP4

# Run specific service tests
go test ./internal/application -run TestHistoricalDataService
go test ./internal/application -run TestPerformanceMetricsService
```

### Integration Tests

```bash
# Run integration tests (requires more time)
go test ./test/integration -run TestWP4Statistics
```

### Benchmarks

```bash
# Run performance benchmarks
go test -bench=. ./internal/infrastructure/timeseries
```

## Best Practices

### 1. Storage Selection

- **Development**: Use `NewMemoryTimeSeriesStore()` for fast iteration
- **Production**: Use `NewSQLiteTimeSeriesStore()` for persistence
- **High-scale**: Consider implementing custom time-series backends

### 2. Data Retention

- Configure retention policies based on storage capacity and requirements
- More granular data for recent periods, aggregated data for historical analysis
- Regular cleanup to manage storage growth

### 3. Performance Optimization

- Use aggregated data for long time range queries
- Implement caching for frequently accessed metrics
- Batch operations when storing multiple data points
- Monitor database size and query performance

### 4. Error Handling

```go
// Always handle errors appropriately
if err := service.StoreRawData(ctx, device, stats); err != nil {
    log.Printf("Failed to store statistics: %v", err)
    // Implement retry logic or alert mechanisms
}
```

### 5. Monitoring the Monitor

- Monitor the statistics system itself for performance issues
- Set up alerts for storage capacity and query latency
- Regular health checks on the time-series database

## Troubleshooting

### Common Issues

1. **High Memory Usage**: Check retention policies and aggregation settings
2. **Slow Queries**: Verify indexes and consider query optimization
3. **Storage Growth**: Monitor disk usage and cleanup policies
4. **Missing Data**: Check data ingestion and storage errors

### Debug Mode

```go
// Enable detailed logging
logger := logging.WithLevel(logging.DEBUG)
service := application.NewHistoricalDataService(store).WithLogger(logger)
```

### Performance Monitoring

```go
// Monitor operation latency
start := time.Now()
err := service.StoreRawData(ctx, device, stats)
duration := time.Since(start)

if duration > 100*time.Millisecond {
    log.Printf("Slow storage operation: %v", duration)
}
```

## Future Enhancements

### Planned Features

1. **Real-time Streaming**: WebSocket support for live dashboard updates
2. **Machine Learning**: Advanced anomaly detection and predictive analytics
3. **Distributed Storage**: Support for distributed time-series databases
4. **Advanced Visualizations**: Built-in charting and dashboard capabilities
5. **Rule Engine**: Custom alerting rules and automated responses

### Extension Points

The system is designed for extensibility:

- **Custom Aggregation Functions**: Implement domain-specific aggregations
- **Additional Export Formats**: Add support for new export formats
- **Advanced Analytics**: Integrate machine learning models
- **Custom Dashboards**: Build application-specific dashboard components

## API Reference

For detailed API documentation, see the individual service documentation:

- [Historical Data Service API](./api/historical-data-service.md)
- [Performance Metrics Service API](./api/performance-metrics-service.md)
- [Statistics Export Service API](./api/statistics-export-service.md)
- [Statistics Reporting Service API](./api/statistics-reporting-service.md)
- [Dashboard Data Service API](./api/dashboard-data-service.md)