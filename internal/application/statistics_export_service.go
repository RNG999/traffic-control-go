package application

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/rng999/traffic-control-go/internal/infrastructure/timeseries"
	"github.com/rng999/traffic-control-go/pkg/logging"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// StatisticsExportService handles exporting statistics data in various formats
type StatisticsExportService struct {
	historicalDataService   *HistoricalDataService
	performanceMetricsService *PerformanceMetricsService
	logger                  logging.Logger
}

// NewStatisticsExportService creates a new statistics export service
func NewStatisticsExportService(
	historicalDataService *HistoricalDataService,
	performanceMetricsService *PerformanceMetricsService,
) *StatisticsExportService {
	return &StatisticsExportService{
		historicalDataService:     historicalDataService,
		performanceMetricsService: performanceMetricsService,
		logger:                    logging.WithComponent("application.statistics_export"),
	}
}

// ExportFormat defines supported export formats
type ExportFormat string

const (
	FormatJSON       ExportFormat = "json"
	FormatCSV        ExportFormat = "csv"
	FormatPrometheus ExportFormat = "prometheus"
)

// ExportOptions configures the export operation
type ExportOptions struct {
	Format          ExportFormat                 `json:"format"`
	DeviceName      tc.DeviceName               `json:"device_name"`
	StartTime       time.Time                   `json:"start_time"`
	EndTime         time.Time                   `json:"end_time"`
	Interval        timeseries.AggregationInterval `json:"interval,omitempty"`
	IncludeRawData  bool                        `json:"include_raw_data"`
	IncludeMetrics  bool                        `json:"include_metrics"`
	MetricsOnly     bool                        `json:"metrics_only"`
	Compressed      bool                        `json:"compressed"`
	// Prometheus-specific options
	PrometheusLabels map[string]string          `json:"prometheus_labels,omitempty"`
	PrometheusPrefix string                     `json:"prometheus_prefix,omitempty"`
}

// ExportResult contains the exported data and metadata
type ExportResult struct {
	Format      ExportFormat  `json:"format"`
	DeviceName  string        `json:"device_name"`
	TimeRange   TimeRange     `json:"time_range"`
	DataPoints  int           `json:"data_points"`
	Size        int64         `json:"size_bytes"`
	GeneratedAt time.Time     `json:"generated_at"`
	Data        []byte        `json:"data,omitempty"`
	Checksum    string        `json:"checksum,omitempty"`
}

// JSONExportData represents the structure for JSON exports
type JSONExportData struct {
	Metadata        ExportMetadata                    `json:"metadata"`
	RawData         []*timeseries.AggregatedData     `json:"raw_data,omitempty"`
	PerformanceMetrics *PerformanceMetrics           `json:"performance_metrics,omitempty"`
	Summary         *DataSummary                     `json:"summary,omitempty"`
}

// ExportMetadata contains metadata about the export
type ExportMetadata struct {
	DeviceName    string        `json:"device_name"`
	TimeRange     TimeRange     `json:"time_range"`
	Interval      string        `json:"interval,omitempty"`
	ExportedAt    time.Time     `json:"exported_at"`
	Format        ExportFormat  `json:"format"`
	DataPoints    int           `json:"data_points"`
	Version       string        `json:"version"`
}

// CSVColumn defines a column in CSV export
type CSVColumn struct {
	Name        string
	Type        string
	Description string
}

// ExportStatistics exports statistics data in the specified format
func (s *StatisticsExportService) ExportStatistics(ctx context.Context, options ExportOptions) (*ExportResult, error) {
	s.logger.Info("Starting statistics export",
		logging.String("device", options.DeviceName.String()),
		logging.String("format", string(options.Format)),
		logging.String("start", options.StartTime.String()),
		logging.String("end", options.EndTime.String()))

	startTime := time.Now()

	// Validate options
	if err := s.validateExportOptions(options); err != nil {
		return nil, fmt.Errorf("invalid export options: %w", err)
	}

	// Export based on format
	var data []byte
	var dataPoints int
	var err error

	switch options.Format {
	case FormatJSON:
		data, dataPoints, err = s.exportJSON(ctx, options)
	case FormatCSV:
		data, dataPoints, err = s.exportCSV(ctx, options)
	case FormatPrometheus:
		data, dataPoints, err = s.exportPrometheus(ctx, options)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", options.Format)
	}

	if err != nil {
		return nil, fmt.Errorf("export failed: %w", err)
	}

	// Calculate checksum for data integrity
	checksum := s.calculateChecksum(data)

	duration := time.Since(startTime)
	s.logger.Info("Statistics export completed",
		logging.String("device", options.DeviceName.String()),
		logging.String("format", string(options.Format)),
		logging.Int("data_points", dataPoints),
		logging.Int64("size_bytes", int64(len(data))),
		logging.String("duration", duration.String()))

	return &ExportResult{
		Format:      options.Format,
		DeviceName:  options.DeviceName.String(),
		TimeRange:   TimeRange{Start: options.StartTime, End: options.EndTime},
		DataPoints:  dataPoints,
		Size:        int64(len(data)),
		GeneratedAt: time.Now(),
		Data:        data,
		Checksum:    checksum,
	}, nil
}

// ExportToWriter exports statistics directly to a writer
func (s *StatisticsExportService) ExportToWriter(ctx context.Context, writer io.Writer, options ExportOptions) (*ExportResult, error) {
	result, err := s.ExportStatistics(ctx, options)
	if err != nil {
		return nil, err
	}

	_, err = writer.Write(result.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to write export data: %w", err)
	}

	// Don't include data in result when writing to external writer
	result.Data = nil

	return result, nil
}

// GetExportSchema returns the schema/structure for a given export format
func (s *StatisticsExportService) GetExportSchema(format ExportFormat) (interface{}, error) {
	switch format {
	case FormatJSON:
		return s.getJSONSchema(), nil
	case FormatCSV:
		return s.getCSVSchema(), nil
	case FormatPrometheus:
		return s.getPrometheusSchema(), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// Private methods

func (s *StatisticsExportService) validateExportOptions(options ExportOptions) error {
	if options.DeviceName.String() == "" {
		return fmt.Errorf("device name is required")
	}

	if options.StartTime.IsZero() || options.EndTime.IsZero() {
		return fmt.Errorf("start time and end time are required")
	}

	if options.StartTime.After(options.EndTime) {
		return fmt.Errorf("start time must be before end time")
	}

	// Validate time range (prevent exports that are too large)
	maxRange := 7 * 24 * time.Hour // 7 days
	if options.EndTime.Sub(options.StartTime) > maxRange {
		return fmt.Errorf("time range cannot exceed %v", maxRange)
	}

	if options.Format == "" {
		return fmt.Errorf("export format is required")
	}

	// Prometheus-specific validation
	if options.Format == FormatPrometheus {
		if options.PrometheusPrefix == "" {
			options.PrometheusPrefix = "tc_" // Default prefix
		}
		// Validate Prometheus metric name conventions
		if !isValidPrometheusName(options.PrometheusPrefix) {
			return fmt.Errorf("invalid Prometheus prefix: %s", options.PrometheusPrefix)
		}
	}

	return nil
}

func (s *StatisticsExportService) exportJSON(ctx context.Context, options ExportOptions) ([]byte, int, error) {
	exportData := &JSONExportData{
		Metadata: ExportMetadata{
			DeviceName: options.DeviceName.String(),
			TimeRange:  TimeRange{Start: options.StartTime, End: options.EndTime},
			Interval:   string(options.Interval),
			ExportedAt: time.Now(),
			Format:     FormatJSON,
			Version:    "1.0",
		},
	}

	var dataPoints int

	// Export raw/aggregated data if requested
	if !options.MetricsOnly {
		historicalData, err := s.historicalDataService.GetHistoricalData(
			ctx, options.DeviceName, options.StartTime, options.EndTime, options.Interval)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get historical data: %w", err)
		}

		exportData.RawData = historicalData
		dataPoints = len(historicalData)
		exportData.Metadata.DataPoints = dataPoints

		// Include summary
		summary, err := s.historicalDataService.GetDataSummary(ctx, options.DeviceName)
		if err != nil {
			s.logger.Warn("Failed to get data summary", logging.Error(err))
		} else {
			exportData.Summary = summary
		}
	}

	// Export performance metrics if requested
	if options.IncludeMetrics || options.MetricsOnly {
		timeRange := TimeRange{Start: options.StartTime, End: options.EndTime}
		metrics, err := s.performanceMetricsService.CalculatePerformanceMetrics(
			ctx, options.DeviceName, timeRange)
		if err != nil {
			s.logger.Warn("Failed to calculate performance metrics", logging.Error(err))
		} else {
			exportData.PerformanceMetrics = metrics
		}
	}

	// Marshal to JSON
	var data []byte
	var err error
	if options.Compressed {
		data, err = json.Marshal(exportData)
	} else {
		data, err = json.MarshalIndent(exportData, "", "  ")
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return data, dataPoints, nil
}

func (s *StatisticsExportService) exportCSV(ctx context.Context, options ExportOptions) ([]byte, int, error) {
	// Get historical data
	historicalData, err := s.historicalDataService.GetHistoricalData(
		ctx, options.DeviceName, options.StartTime, options.EndTime, options.Interval)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get historical data: %w", err)
	}

	if len(historicalData) == 0 {
		return []byte("No data available for the specified time range\n"), 0, nil
	}

	// Create CSV buffer
	var csvData strings.Builder
	writer := csv.NewWriter(&csvData)

	// Write header
	header := s.getCSVHeader()
	if err := writer.Write(header); err != nil {
		return nil, 0, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	dataPoints := 0
	for _, data := range historicalData {
		rows := s.convertToCSVRows(data)
		for _, row := range rows {
			if err := writer.Write(row); err != nil {
				return nil, 0, fmt.Errorf("failed to write CSV row: %w", err)
			}
			dataPoints++
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, 0, fmt.Errorf("CSV writer error: %w", err)
	}

	return []byte(csvData.String()), dataPoints, nil
}

func (s *StatisticsExportService) exportPrometheus(ctx context.Context, options ExportOptions) ([]byte, int, error) {
	// Get historical data
	historicalData, err := s.historicalDataService.GetHistoricalData(
		ctx, options.DeviceName, options.StartTime, options.EndTime, options.Interval)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get historical data: %w", err)
	}

	// Get performance metrics if available
	var metrics *PerformanceMetrics
	if options.IncludeMetrics {
		timeRange := TimeRange{Start: options.StartTime, End: options.EndTime}
		metrics, err = s.performanceMetricsService.CalculatePerformanceMetrics(
			ctx, options.DeviceName, timeRange)
		if err != nil {
			s.logger.Warn("Failed to calculate performance metrics for Prometheus export", logging.Error(err))
		}
	}

	// Convert to Prometheus format
	prometheusData := s.convertToPrometheusFormat(historicalData, metrics, options)

	return []byte(prometheusData), len(historicalData), nil
}

func (s *StatisticsExportService) getCSVHeader() []string {
	return []string{
		"timestamp",
		"device_name",
		"interval",
		"component_type", // qdisc, class, filter, link
		"component_handle",
		"component_parent",
		"metric_name",
		"metric_value",
		"metric_unit",
	}
}

func (s *StatisticsExportService) convertToCSVRows(data *timeseries.AggregatedData) [][]string {
	var rows [][]string
	timestamp := data.Timestamp.Format(time.RFC3339)
	deviceName := data.DeviceName
	interval := string(data.Interval)

	// Qdisc stats
	for _, qdisc := range data.QdiscStats {
		rows = append(rows, []string{timestamp, deviceName, interval, "qdisc", qdisc.Handle, "", "total_bytes", strconv.FormatUint(qdisc.TotalBytes, 10), "bytes"})
		rows = append(rows, []string{timestamp, deviceName, interval, "qdisc", qdisc.Handle, "", "total_packets", strconv.FormatUint(qdisc.TotalPackets, 10), "packets"})
		rows = append(rows, []string{timestamp, deviceName, interval, "qdisc", qdisc.Handle, "", "total_drops", strconv.FormatUint(qdisc.TotalDrops, 10), "packets"})
		rows = append(rows, []string{timestamp, deviceName, interval, "qdisc", qdisc.Handle, "", "avg_rate", strconv.FormatUint(qdisc.AvgRate, 10), "bps"})
		rows = append(rows, []string{timestamp, deviceName, interval, "qdisc", qdisc.Handle, "", "max_rate", strconv.FormatUint(qdisc.MaxRate, 10), "bps"})
		rows = append(rows, []string{timestamp, deviceName, interval, "qdisc", qdisc.Handle, "", "max_backlog", strconv.FormatUint(qdisc.MaxBacklog, 10), "bytes"})
	}

	// Class stats
	for _, class := range data.ClassStats {
		rows = append(rows, []string{timestamp, deviceName, interval, "class", class.Handle, class.Parent, "total_bytes", strconv.FormatUint(class.TotalBytes, 10), "bytes"})
		rows = append(rows, []string{timestamp, deviceName, interval, "class", class.Handle, class.Parent, "total_packets", strconv.FormatUint(class.TotalPackets, 10), "packets"})
		rows = append(rows, []string{timestamp, deviceName, interval, "class", class.Handle, class.Parent, "total_drops", strconv.FormatUint(class.TotalDrops, 10), "packets"})
		rows = append(rows, []string{timestamp, deviceName, interval, "class", class.Handle, class.Parent, "avg_rate", strconv.FormatUint(class.AvgRate, 10), "bps"})
		rows = append(rows, []string{timestamp, deviceName, interval, "class", class.Handle, class.Parent, "total_lends", strconv.FormatUint(class.TotalLends, 10), "count"})
		rows = append(rows, []string{timestamp, deviceName, interval, "class", class.Handle, class.Parent, "total_borrows", strconv.FormatUint(class.TotalBorrows, 10), "count"})
	}

	// Filter stats
	for _, filter := range data.FilterStats {
		rows = append(rows, []string{timestamp, deviceName, interval, "filter", filter.Handle, filter.Parent, "total_matches", strconv.FormatUint(filter.TotalMatches, 10), "matches"})
		rows = append(rows, []string{timestamp, deviceName, interval, "filter", filter.Handle, filter.Parent, "total_bytes", strconv.FormatUint(filter.TotalBytes, 10), "bytes"})
		rows = append(rows, []string{timestamp, deviceName, interval, "filter", filter.Handle, filter.Parent, "total_packets", strconv.FormatUint(filter.TotalPackets, 10), "packets"})
		rows = append(rows, []string{timestamp, deviceName, interval, "filter", filter.Handle, filter.Parent, "avg_rate", strconv.FormatUint(filter.AvgRate, 10), "bps"})
	}

	// Link stats
	link := data.LinkStats
	rows = append(rows, []string{timestamp, deviceName, interval, "link", "interface", "", "total_rx_bytes", strconv.FormatUint(link.TotalRxBytes, 10), "bytes"})
	rows = append(rows, []string{timestamp, deviceName, interval, "link", "interface", "", "total_tx_bytes", strconv.FormatUint(link.TotalTxBytes, 10), "bytes"})
	rows = append(rows, []string{timestamp, deviceName, interval, "link", "interface", "", "total_rx_packets", strconv.FormatUint(link.TotalRxPackets, 10), "packets"})
	rows = append(rows, []string{timestamp, deviceName, interval, "link", "interface", "", "total_tx_packets", strconv.FormatUint(link.TotalTxPackets, 10), "packets"})
	rows = append(rows, []string{timestamp, deviceName, interval, "link", "interface", "", "avg_rx_rate", strconv.FormatUint(link.AvgRxRate, 10), "bps"})
	rows = append(rows, []string{timestamp, deviceName, interval, "link", "interface", "", "avg_tx_rate", strconv.FormatUint(link.AvgTxRate, 10), "bps"})

	return rows
}

func (s *StatisticsExportService) convertToPrometheusFormat(
	historicalData []*timeseries.AggregatedData,
	metrics *PerformanceMetrics,
	options ExportOptions) string {

	var prometheus strings.Builder
	prefix := options.PrometheusPrefix
	if prefix == "" {
		prefix = "tc_"
	}

	// Common labels
	baseLabels := fmt.Sprintf(`device="%s"`, options.DeviceName.String())
	for key, value := range options.PrometheusLabels {
		baseLabels += fmt.Sprintf(`,%s="%s"`, key, value)
	}

	// Write comments and type definitions
	prometheus.WriteString("# HELP tc_qdisc_bytes_total Total bytes processed by qdisc\n")
	prometheus.WriteString("# TYPE tc_qdisc_bytes_total counter\n")
	prometheus.WriteString("# HELP tc_qdisc_packets_total Total packets processed by qdisc\n")
	prometheus.WriteString("# TYPE tc_qdisc_packets_total counter\n")
	prometheus.WriteString("# HELP tc_qdisc_drops_total Total packets dropped by qdisc\n")
	prometheus.WriteString("# TYPE tc_qdisc_drops_total counter\n")
	prometheus.WriteString("# HELP tc_qdisc_rate_bps Current rate in bits per second\n")
	prometheus.WriteString("# TYPE tc_qdisc_rate_bps gauge\n")

	// Export latest data point for current metrics
	if len(historicalData) > 0 {
		latest := historicalData[len(historicalData)-1]
		timestamp := latest.Timestamp.Unix() * 1000 // Convert to milliseconds

		// Qdisc metrics
		for _, qdisc := range latest.QdiscStats {
			labels := fmt.Sprintf(`%s,handle="%s",type="%s"`, baseLabels, qdisc.Handle, qdisc.Type)
			prometheus.WriteString(fmt.Sprintf("%sqdisc_bytes_total{%s} %d %d\n", prefix, labels, qdisc.TotalBytes, timestamp))
			prometheus.WriteString(fmt.Sprintf("%sqdisc_packets_total{%s} %d %d\n", prefix, labels, qdisc.TotalPackets, timestamp))
			prometheus.WriteString(fmt.Sprintf("%sqdisc_drops_total{%s} %d %d\n", prefix, labels, qdisc.TotalDrops, timestamp))
			prometheus.WriteString(fmt.Sprintf("%sqdisc_rate_bps{%s} %d %d\n", prefix, labels, qdisc.AvgRate, timestamp))
			prometheus.WriteString(fmt.Sprintf("%sqdisc_backlog_bytes{%s} %d %d\n", prefix, labels, qdisc.MaxBacklog, timestamp))
		}

		// Class metrics
		for _, class := range latest.ClassStats {
			labels := fmt.Sprintf(`%s,handle="%s",parent="%s",type="%s"`, baseLabels, class.Handle, class.Parent, class.Type)
			prometheus.WriteString(fmt.Sprintf("%sclass_bytes_total{%s} %d %d\n", prefix, labels, class.TotalBytes, timestamp))
			prometheus.WriteString(fmt.Sprintf("%sclass_packets_total{%s} %d %d\n", prefix, labels, class.TotalPackets, timestamp))
			prometheus.WriteString(fmt.Sprintf("%sclass_drops_total{%s} %d %d\n", prefix, labels, class.TotalDrops, timestamp))
			prometheus.WriteString(fmt.Sprintf("%sclass_rate_bps{%s} %d %d\n", prefix, labels, class.AvgRate, timestamp))
			prometheus.WriteString(fmt.Sprintf("%sclass_lends_total{%s} %d %d\n", prefix, labels, class.TotalLends, timestamp))
			prometheus.WriteString(fmt.Sprintf("%sclass_borrows_total{%s} %d %d\n", prefix, labels, class.TotalBorrows, timestamp))
		}

		// Filter metrics
		for _, filter := range latest.FilterStats {
			labels := fmt.Sprintf(`%s,handle="%s",parent="%s",flow_id="%s"`, baseLabels, filter.Handle, filter.Parent, filter.FlowID)
			prometheus.WriteString(fmt.Sprintf("%sfilter_matches_total{%s} %d %d\n", prefix, labels, filter.TotalMatches, timestamp))
			prometheus.WriteString(fmt.Sprintf("%sfilter_bytes_total{%s} %d %d\n", prefix, labels, filter.TotalBytes, timestamp))
			prometheus.WriteString(fmt.Sprintf("%sfilter_packets_total{%s} %d %d\n", prefix, labels, filter.TotalPackets, timestamp))
			prometheus.WriteString(fmt.Sprintf("%sfilter_rate_bps{%s} %d %d\n", prefix, labels, filter.AvgRate, timestamp))
		}

		// Link metrics
		link := latest.LinkStats
		prometheus.WriteString(fmt.Sprintf("%slink_rx_bytes_total{%s} %d %d\n", prefix, baseLabels, link.TotalRxBytes, timestamp))
		prometheus.WriteString(fmt.Sprintf("%slink_tx_bytes_total{%s} %d %d\n", prefix, baseLabels, link.TotalTxBytes, timestamp))
		prometheus.WriteString(fmt.Sprintf("%slink_rx_packets_total{%s} %d %d\n", prefix, baseLabels, link.TotalRxPackets, timestamp))
		prometheus.WriteString(fmt.Sprintf("%slink_tx_packets_total{%s} %d %d\n", prefix, baseLabels, link.TotalTxPackets, timestamp))
		prometheus.WriteString(fmt.Sprintf("%slink_rx_rate_bps{%s} %d %d\n", prefix, baseLabels, link.AvgRxRate, timestamp))
		prometheus.WriteString(fmt.Sprintf("%slink_tx_rate_bps{%s} %d %d\n", prefix, baseLabels, link.AvgTxRate, timestamp))
	}

	// Export performance metrics if available
	if metrics != nil {
		timestamp := time.Now().Unix() * 1000
		
		// Overall health score
		prometheus.WriteString(fmt.Sprintf("%shealth_score{%s} %.2f %d\n", prefix, baseLabels, metrics.HealthScore.OverallScore, timestamp))
		prometheus.WriteString(fmt.Sprintf("%sthroughput_utilization_percent{%s} %.2f %d\n", prefix, baseLabels, metrics.OverallMetrics.ThroughputUtilization, timestamp))
		prometheus.WriteString(fmt.Sprintf("%spacket_loss_rate_percent{%s} %.2f %d\n", prefix, baseLabels, metrics.OverallMetrics.PacketLossRate, timestamp))
		prometheus.WriteString(fmt.Sprintf("%sefficiency_score{%s} %.2f %d\n", prefix, baseLabels, metrics.OverallMetrics.EfficiencyScore, timestamp))

		// Component health scores
		for _, qdisc := range metrics.QdiscMetrics {
			labels := fmt.Sprintf(`%s,handle="%s",type="%s"`, baseLabels, qdisc.Handle, qdisc.Type)
			prometheus.WriteString(fmt.Sprintf("%sqdisc_performance_score{%s} %.2f %d\n", prefix, labels, qdisc.PerformanceScore, timestamp))
			prometheus.WriteString(fmt.Sprintf("%sqdisc_drop_rate_percent{%s} %.2f %d\n", prefix, labels, qdisc.DropRate, timestamp))
		}

		for _, class := range metrics.ClassMetrics {
			labels := fmt.Sprintf(`%s,handle="%s",parent="%s",type="%s"`, baseLabels, class.Handle, class.Parent, class.Type)
			prometheus.WriteString(fmt.Sprintf("%sclass_performance_score{%s} %.2f %d\n", prefix, labels, class.PerformanceScore, timestamp))
			prometheus.WriteString(fmt.Sprintf("%sclass_bandwidth_utilization_percent{%s} %.2f %d\n", prefix, labels, class.BandwidthUtilization, timestamp))
		}
	}

	return prometheus.String()
}

func (s *StatisticsExportService) getJSONSchema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"metadata": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"device_name":  map[string]interface{}{"type": "string"},
					"time_range":   map[string]interface{}{"type": "object"},
					"interval":     map[string]interface{}{"type": "string"},
					"exported_at":  map[string]interface{}{"type": "string", "format": "date-time"},
					"format":       map[string]interface{}{"type": "string"},
					"data_points":  map[string]interface{}{"type": "integer"},
					"version":      map[string]interface{}{"type": "string"},
				},
			},
			"raw_data": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{"type": "object"},
			},
			"performance_metrics": map[string]interface{}{"type": "object"},
			"summary":            map[string]interface{}{"type": "object"},
		},
	}
}

func (s *StatisticsExportService) getCSVSchema() []CSVColumn {
	return []CSVColumn{
		{Name: "timestamp", Type: "string", Description: "ISO 8601 timestamp"},
		{Name: "device_name", Type: "string", Description: "Network device name"},
		{Name: "interval", Type: "string", Description: "Aggregation interval"},
		{Name: "component_type", Type: "string", Description: "Type of component (qdisc, class, filter, link)"},
		{Name: "component_handle", Type: "string", Description: "Component handle/identifier"},
		{Name: "component_parent", Type: "string", Description: "Parent component handle"},
		{Name: "metric_name", Type: "string", Description: "Name of the metric"},
		{Name: "metric_value", Type: "string", Description: "Metric value"},
		{Name: "metric_unit", Type: "string", Description: "Unit of measurement"},
	}
}

func (s *StatisticsExportService) getPrometheusSchema() interface{} {
	return map[string]interface{}{
		"format": "Prometheus exposition format",
		"metrics": []map[string]interface{}{
			{
				"name":        "tc_qdisc_bytes_total",
				"type":        "counter",
				"description": "Total bytes processed by qdisc",
				"labels":      []string{"device", "handle", "type"},
			},
			{
				"name":        "tc_qdisc_packets_total",
				"type":        "counter",
				"description": "Total packets processed by qdisc",
				"labels":      []string{"device", "handle", "type"},
			},
			{
				"name":        "tc_health_score",
				"type":        "gauge",
				"description": "Overall system health score (0-100)",
				"labels":      []string{"device"},
			},
		},
	}
}

func (s *StatisticsExportService) calculateChecksum(data []byte) string {
	// Simple checksum - in production, use SHA256 or similar
	sum := uint32(0)
	for _, b := range data {
		sum += uint32(b)
	}
	return fmt.Sprintf("%08x", sum)
}

func isValidPrometheusName(name string) bool {
	if len(name) == 0 {
		return false
	}
	
	// Prometheus metric names must match [a-zA-Z_:][a-zA-Z0-9_:]*
	for i, r := range name {
		if i == 0 {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || r == ':') {
				return false
			}
		} else {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == ':') {
				return false
			}
		}
	}
	return true
}

// GetSupportedFormats returns a list of supported export formats
func (s *StatisticsExportService) GetSupportedFormats() []ExportFormat {
	return []ExportFormat{
		FormatJSON,
		FormatCSV,
		FormatPrometheus,
	}
}

// ValidateFormat checks if the given format is supported
func (s *StatisticsExportService) ValidateFormat(format ExportFormat) error {
	supportedFormats := s.GetSupportedFormats()
	for _, supported := range supportedFormats {
		if format == supported {
			return nil
		}
	}
	return fmt.Errorf("unsupported format: %s, supported formats: %v", format, supportedFormats)
}

// EstimateExportSize estimates the size of an export operation
func (s *StatisticsExportService) EstimateExportSize(ctx context.Context, options ExportOptions) (int64, error) {
	// Get data summary to estimate size
	summary, err := s.historicalDataService.GetDataSummary(ctx, options.DeviceName)
	if err != nil {
		return 0, fmt.Errorf("failed to get data summary: %w", err)
	}

	// Rough estimation based on format
	var estimatedSize int64
	switch options.Format {
	case FormatJSON:
		// JSON is typically verbose, estimate ~2KB per data point
		estimatedSize = int64(summary.TotalDataPoints) * 2048
	case FormatCSV:
		// CSV is more compact, estimate ~500 bytes per data point
		estimatedSize = int64(summary.TotalDataPoints) * 512
	case FormatPrometheus:
		// Prometheus format varies, estimate ~200 bytes per metric
		estimatedSize = int64(summary.TotalDataPoints) * 200
	default:
		estimatedSize = int64(summary.TotalDataPoints) * 1024 // Default estimate
	}

	// Add overhead for metadata
	estimatedSize += 4096

	return estimatedSize, nil
}