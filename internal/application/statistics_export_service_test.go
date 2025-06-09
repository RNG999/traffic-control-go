package application

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	"github.com/rng999/traffic-control-go/internal/infrastructure/timeseries"
	"github.com/rng999/traffic-control-go/pkg/tc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatisticsExportService(t *testing.T) {
	// Set up test dependencies
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	historicalDataService := NewHistoricalDataService(timeSeriesStore)
	performanceMetricsService := NewPerformanceMetricsService(historicalDataService)
	service := NewStatisticsExportService(historicalDataService, performanceMetricsService)
	
	ctx := context.Background()
	deviceName, _ := tc.NewDeviceName("eth0")

	// Create test data
	now := time.Now().Truncate(time.Minute)
	
	// Store sample data
	for i := 0; i < 5; i++ {
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
						BytesDropped: uint64(5 * (i + 1)),
						Overlimits:   uint64(10 * (i + 1)),
						Requeues:     uint64(2 * (i + 1)),
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
						BytesDropped:   uint64(2 * (i + 1)),
						Overlimits:     uint64(5 * (i + 1)),
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
				RxErrors:  uint64(i),
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

	t.Run("export JSON format", func(t *testing.T) {
		options := ExportOptions{
			Format:         FormatJSON,
			DeviceName:     deviceName,
			StartTime:      now.Add(-1 * time.Hour),
			EndTime:        now,
			IncludeRawData: true,
			IncludeMetrics: true,
		}

		result, err := service.ExportStatistics(ctx, options)
		require.NoError(t, err)

		assert.Equal(t, FormatJSON, result.Format)
		assert.Equal(t, deviceName.String(), result.DeviceName)
		assert.Greater(t, result.DataPoints, 0)
		assert.Greater(t, result.Size, int64(0))
		assert.NotEmpty(t, result.Data)
		assert.NotEmpty(t, result.Checksum)

		// Verify JSON structure
		var exportData JSONExportData
		err = json.Unmarshal(result.Data, &exportData)
		require.NoError(t, err)

		assert.Equal(t, deviceName.String(), exportData.Metadata.DeviceName)
		assert.Equal(t, FormatJSON, exportData.Metadata.Format)
		assert.NotEmpty(t, exportData.Metadata.Version)
		assert.NotEmpty(t, exportData.RawData)
		assert.NotNil(t, exportData.PerformanceMetrics)
		assert.NotNil(t, exportData.Summary)
	})

	t.Run("export CSV format", func(t *testing.T) {
		options := ExportOptions{
			Format:     FormatCSV,
			DeviceName: deviceName,
			StartTime:  now.Add(-1 * time.Hour),
			EndTime:    now,
		}

		result, err := service.ExportStatistics(ctx, options)
		require.NoError(t, err)

		assert.Equal(t, FormatCSV, result.Format)
		assert.Greater(t, result.DataPoints, 0)
		assert.Greater(t, result.Size, int64(0))

		// Verify CSV structure
		csvReader := csv.NewReader(strings.NewReader(string(result.Data)))
		records, err := csvReader.ReadAll()
		require.NoError(t, err)

		// Should have header + data rows
		assert.Greater(t, len(records), 1)

		// Verify header
		expectedHeader := []string{
			"timestamp", "device_name", "interval", "component_type",
			"component_handle", "component_parent", "metric_name",
			"metric_value", "metric_unit",
		}
		assert.Equal(t, expectedHeader, records[0])

		// Verify data rows contain expected device name
		for i := 1; i < len(records); i++ {
			assert.Equal(t, deviceName.String(), records[i][1])
		}
	})

	t.Run("export Prometheus format", func(t *testing.T) {
		options := ExportOptions{
			Format:           FormatPrometheus,
			DeviceName:       deviceName,
			StartTime:        now.Add(-1 * time.Hour),
			EndTime:          now,
			IncludeMetrics:   true,
			PrometheusPrefix: "test_tc_",
			PrometheusLabels: map[string]string{
				"environment": "test",
				"region":      "us-west",
			},
		}

		result, err := service.ExportStatistics(ctx, options)
		require.NoError(t, err)

		assert.Equal(t, FormatPrometheus, result.Format)
		assert.Greater(t, result.Size, int64(0))

		prometheusData := string(result.Data)
		
		// Verify Prometheus format
		assert.Contains(t, prometheusData, "# HELP")
		assert.Contains(t, prometheusData, "# TYPE")
		assert.Contains(t, prometheusData, "test_tc_qdisc_bytes_total")
		assert.Contains(t, prometheusData, fmt.Sprintf(`device="%s"`, deviceName.String()))
		assert.Contains(t, prometheusData, `environment="test"`)
		assert.Contains(t, prometheusData, `region="us-west"`)

		// Verify metric naming
		assert.Contains(t, prometheusData, "test_tc_qdisc_bytes_total")
		assert.Contains(t, prometheusData, "test_tc_class_bytes_total")
		assert.Contains(t, prometheusData, "test_tc_filter_matches_total")
		assert.Contains(t, prometheusData, "test_tc_link_rx_bytes_total")
	})

	t.Run("export metrics only", func(t *testing.T) {
		options := ExportOptions{
			Format:      FormatJSON,
			DeviceName:  deviceName,
			StartTime:   now.Add(-1 * time.Hour),
			EndTime:     now,
			MetricsOnly: true,
		}

		result, err := service.ExportStatistics(ctx, options)
		require.NoError(t, err)

		var exportData JSONExportData
		err = json.Unmarshal(result.Data, &exportData)
		require.NoError(t, err)

		// Should not include raw data when metrics only
		assert.Empty(t, exportData.RawData)
		assert.NotNil(t, exportData.PerformanceMetrics)
	})

	t.Run("export with aggregation interval", func(t *testing.T) {
		// First aggregate data
		err := historicalDataService.AggregateData(ctx, deviceName, timeseries.IntervalHour, 
			now.Add(-2*time.Hour), now)
		require.NoError(t, err)

		options := ExportOptions{
			Format:     FormatJSON,
			DeviceName: deviceName,
			StartTime:  now.Add(-2 * time.Hour),
			EndTime:    now,
			Interval:   timeseries.IntervalHour,
		}

		result, err := service.ExportStatistics(ctx, options)
		require.NoError(t, err)

		var exportData JSONExportData
		err = json.Unmarshal(result.Data, &exportData)
		require.NoError(t, err)

		assert.Equal(t, string(timeseries.IntervalHour), exportData.Metadata.Interval)
	})
}

func TestStatisticsExportService_Validation(t *testing.T) {
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	historicalDataService := NewHistoricalDataService(timeSeriesStore)
	performanceMetricsService := NewPerformanceMetricsService(historicalDataService)
	service := NewStatisticsExportService(historicalDataService, performanceMetricsService)
	
	ctx := context.Background()
	deviceName, _ := tc.NewDeviceName("eth0")
	now := time.Now()

	t.Run("validate export options", func(t *testing.T) {
		// Missing device name
		options := ExportOptions{
			Format:    FormatJSON,
			StartTime: now.Add(-1 * time.Hour),
			EndTime:   now,
		}
		_, err := service.ExportStatistics(ctx, options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "device name is required")

		// Missing time range
		options = ExportOptions{
			Format:     FormatJSON,
			DeviceName: deviceName,
		}
		_, err = service.ExportStatistics(ctx, options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start time and end time are required")

		// Invalid time range
		options = ExportOptions{
			Format:     FormatJSON,
			DeviceName: deviceName,
			StartTime:  now,
			EndTime:    now.Add(-1 * time.Hour),
		}
		_, err = service.ExportStatistics(ctx, options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start time must be before end time")

		// Time range too large
		options = ExportOptions{
			Format:     FormatJSON,
			DeviceName: deviceName,
			StartTime:  now.Add(-8 * 24 * time.Hour),
			EndTime:    now,
		}
		_, err = service.ExportStatistics(ctx, options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "time range cannot exceed")

		// Missing format
		options = ExportOptions{
			DeviceName: deviceName,
			StartTime:  now.Add(-1 * time.Hour),
			EndTime:    now,
		}
		_, err = service.ExportStatistics(ctx, options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "export format is required")

		// Unsupported format
		options = ExportOptions{
			Format:     ExportFormat("xml"),
			DeviceName: deviceName,
			StartTime:  now.Add(-1 * time.Hour),
			EndTime:    now,
		}
		_, err = service.ExportStatistics(ctx, options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported export format")
	})

	t.Run("validate Prometheus options", func(t *testing.T) {
		// Invalid Prometheus prefix
		options := ExportOptions{
			Format:           FormatPrometheus,
			DeviceName:       deviceName,
			StartTime:        now.Add(-1 * time.Hour),
			EndTime:          now,
			PrometheusPrefix: "123invalid",
		}
		_, err := service.ExportStatistics(ctx, options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid Prometheus prefix")
	})
}

func TestStatisticsExportService_EmptyData(t *testing.T) {
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	historicalDataService := NewHistoricalDataService(timeSeriesStore)
	performanceMetricsService := NewPerformanceMetricsService(historicalDataService)
	service := NewStatisticsExportService(historicalDataService, performanceMetricsService)
	
	ctx := context.Background()
	deviceName, _ := tc.NewDeviceName("eth0")
	now := time.Now()

	t.Run("export with no data", func(t *testing.T) {
		options := ExportOptions{
			Format:     FormatJSON,
			DeviceName: deviceName,
			StartTime:  now.Add(-1 * time.Hour),
			EndTime:    now,
		}

		result, err := service.ExportStatistics(ctx, options)
		require.NoError(t, err)

		assert.Equal(t, 0, result.DataPoints)
		
		var exportData JSONExportData
		err = json.Unmarshal(result.Data, &exportData)
		require.NoError(t, err)

		assert.Empty(t, exportData.RawData)
	})

	t.Run("CSV export with no data", func(t *testing.T) {
		options := ExportOptions{
			Format:     FormatCSV,
			DeviceName: deviceName,
			StartTime:  now.Add(-1 * time.Hour),
			EndTime:    now,
		}

		result, err := service.ExportStatistics(ctx, options)
		require.NoError(t, err)

		assert.Contains(t, string(result.Data), "No data available")
	})
}

func TestStatisticsExportService_Utilities(t *testing.T) {
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	historicalDataService := NewHistoricalDataService(timeSeriesStore)
	performanceMetricsService := NewPerformanceMetricsService(historicalDataService)
	service := NewStatisticsExportService(historicalDataService, performanceMetricsService)
	
	ctx := context.Background()
	deviceName, _ := tc.NewDeviceName("eth0")

	t.Run("get supported formats", func(t *testing.T) {
		formats := service.GetSupportedFormats()
		assert.Contains(t, formats, FormatJSON)
		assert.Contains(t, formats, FormatCSV)
		assert.Contains(t, formats, FormatPrometheus)
	})

	t.Run("validate format", func(t *testing.T) {
		err := service.ValidateFormat(FormatJSON)
		assert.NoError(t, err)

		err = service.ValidateFormat(ExportFormat("invalid"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
	})

	t.Run("get export schema", func(t *testing.T) {
		// JSON schema
		schema, err := service.GetExportSchema(FormatJSON)
		require.NoError(t, err)
		assert.NotNil(t, schema)

		// CSV schema
		schema, err = service.GetExportSchema(FormatCSV)
		require.NoError(t, err)
		csvSchema, ok := schema.([]CSVColumn)
		assert.True(t, ok)
		assert.Greater(t, len(csvSchema), 0)

		// Prometheus schema
		schema, err = service.GetExportSchema(FormatPrometheus)
		require.NoError(t, err)
		assert.NotNil(t, schema)

		// Invalid format
		_, err = service.GetExportSchema(ExportFormat("invalid"))
		assert.Error(t, err)
	})

	t.Run("estimate export size", func(t *testing.T) {
		options := ExportOptions{
			Format:     FormatJSON,
			DeviceName: deviceName,
			StartTime:  time.Now().Add(-1 * time.Hour),
			EndTime:    time.Now(),
		}

		size, err := service.EstimateExportSize(ctx, options)
		require.NoError(t, err)
		assert.Greater(t, size, int64(0))

		// Different formats should have different size estimates
		options.Format = FormatCSV
		csvSize, err := service.EstimateExportSize(ctx, options)
		require.NoError(t, err)

		options.Format = FormatPrometheus
		promSize, err := service.EstimateExportSize(ctx, options)
		require.NoError(t, err)

		// All sizes should be at least the overhead
		assert.GreaterOrEqual(t, size, int64(4096))
		assert.GreaterOrEqual(t, csvSize, int64(4096))
		assert.GreaterOrEqual(t, promSize, int64(4096))
	})
}

func TestStatisticsExportService_ExportToWriter(t *testing.T) {
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	historicalDataService := NewHistoricalDataService(timeSeriesStore)
	performanceMetricsService := NewPerformanceMetricsService(historicalDataService)
	service := NewStatisticsExportService(historicalDataService, performanceMetricsService)
	
	ctx := context.Background()
	deviceName, _ := tc.NewDeviceName("eth0")

	t.Run("export to writer", func(t *testing.T) {
		var output strings.Builder
		
		options := ExportOptions{
			Format:     FormatJSON,
			DeviceName: deviceName,
			StartTime:  time.Now().Add(-1 * time.Hour),
			EndTime:    time.Now(),
		}

		result, err := service.ExportToWriter(ctx, &output, options)
		require.NoError(t, err)

		assert.Greater(t, output.Len(), 0)
		assert.Nil(t, result.Data) // Data should not be included in result when writing to writer
		assert.Greater(t, result.Size, int64(0))
	})
}

func TestPrometheusNameValidation(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid name", "tc_metric", true},
		{"valid with underscore", "tc_metric_name", true},
		{"valid with colon", "tc:metric", true},
		{"valid mixed case", "Tc_Metric", true},
		{"invalid starts with number", "1tc_metric", false},
		{"invalid character", "tc-metric", false},
		{"empty string", "", false},
		{"single character", "t", true},
		{"single underscore", "_", true},
		{"single colon", ":", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidPrometheusName(tc.input)
			assert.Equal(t, tc.expected, result, "Input: %s", tc.input)
		})
	}
}

func TestStatisticsExportService_CSVGeneration(t *testing.T) {
	// Test CSV header generation
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	historicalDataService := NewHistoricalDataService(timeSeriesStore)
	performanceMetricsService := NewPerformanceMetricsService(historicalDataService)
	service := NewStatisticsExportService(historicalDataService, performanceMetricsService)

	header := service.getCSVHeader()
	expectedColumns := []string{
		"timestamp", "device_name", "interval", "component_type",
		"component_handle", "component_parent", "metric_name",
		"metric_value", "metric_unit",
	}
	assert.Equal(t, expectedColumns, header)

	// Test CSV row conversion
	testData := &timeseries.AggregatedData{
		Timestamp:  time.Now(),
		DeviceName: "test-device",
		Interval:   timeseries.IntervalHour,
		QdiscStats: []timeseries.AggregatedQdiscStats{
			{
				Handle:      "1:",
				TotalBytes:  1000000,
				TotalPackets: 1000,
				TotalDrops:  5,
				AvgRate:     100000,
				MaxRate:     150000,
				MaxBacklog:  500,
			},
		},
		ClassStats: []timeseries.AggregatedClassStats{
			{
				Handle:       "1:10",
				Parent:       "1:",
				TotalBytes:   500000,
				TotalPackets: 500,
				TotalDrops:   2,
				AvgRate:      50000,
				TotalLends:   10,
				TotalBorrows: 5,
			},
		},
		FilterStats: []timeseries.AggregatedFilterStats{
			{
				Handle:       "800:100",
				Parent:       "1:",
				TotalMatches: 250,
				TotalBytes:   250000,
				TotalPackets: 250,
				AvgRate:      25000,
			},
		},
		LinkStats: timeseries.AggregatedLinkStats{
			TotalRxBytes:   2000000,
			TotalTxBytes:   1500000,
			TotalRxPackets: 2000,
			TotalTxPackets: 1500,
			AvgRxRate:      200000,
			AvgTxRate:      150000,
		},
	}

	rows := service.convertToCSVRows(testData)
	assert.Greater(t, len(rows), 0)

	// Verify all rows have the correct number of columns
	for _, row := range rows {
		assert.Len(t, row, len(header))
		assert.Equal(t, "test-device", row[1]) // device_name column
		assert.Equal(t, string(timeseries.IntervalHour), row[2]) // interval column
	}

	// Verify qdisc metrics are present
	qdiscRows := filterRowsByComponentType(rows, "qdisc")
	assert.Greater(t, len(qdiscRows), 0)

	// Verify class metrics are present
	classRows := filterRowsByComponentType(rows, "class")
	assert.Greater(t, len(classRows), 0)

	// Verify filter metrics are present
	filterRows := filterRowsByComponentType(rows, "filter")
	assert.Greater(t, len(filterRows), 0)

	// Verify link metrics are present
	linkRows := filterRowsByComponentType(rows, "link")
	assert.Greater(t, len(linkRows), 0)
}

func filterRowsByComponentType(rows [][]string, componentType string) [][]string {
	var filtered [][]string
	for _, row := range rows {
		if len(row) > 3 && row[3] == componentType {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func TestStatisticsExportService_Checksum(t *testing.T) {
	timeSeriesStore := timeseries.NewMemoryTimeSeriesStore()
	defer timeSeriesStore.Close()
	
	historicalDataService := NewHistoricalDataService(timeSeriesStore)
	performanceMetricsService := NewPerformanceMetricsService(historicalDataService)
	service := NewStatisticsExportService(historicalDataService, performanceMetricsService)

	// Test checksum calculation
	data1 := []byte("test data")
	checksum1 := service.calculateChecksum(data1)
	assert.NotEmpty(t, checksum1)

	// Same data should produce same checksum
	checksum2 := service.calculateChecksum(data1)
	assert.Equal(t, checksum1, checksum2)

	// Different data should produce different checksum
	data2 := []byte("different test data")
	checksum3 := service.calculateChecksum(data2)
	assert.NotEqual(t, checksum1, checksum3)
}