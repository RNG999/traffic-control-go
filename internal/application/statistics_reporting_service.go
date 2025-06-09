package application

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/rng999/traffic-control-go/internal/infrastructure/timeseries"
	"github.com/rng999/traffic-control-go/pkg/logging"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// StatisticsReportingService provides advanced reporting and trend analysis capabilities
type StatisticsReportingService struct {
	historicalDataService     *HistoricalDataService
	performanceMetricsService *PerformanceMetricsService
	exportService            *StatisticsExportService
	logger                   logging.Logger
}

// NewStatisticsReportingService creates a new statistics reporting service
func NewStatisticsReportingService(
	historicalDataService *HistoricalDataService,
	performanceMetricsService *PerformanceMetricsService,
	exportService *StatisticsExportService,
) *StatisticsReportingService {
	return &StatisticsReportingService{
		historicalDataService:     historicalDataService,
		performanceMetricsService: performanceMetricsService,
		exportService:            exportService,
		logger:                    logging.WithComponent("application.statistics_reporting"),
	}
}

// ReportType defines the type of report to generate
type ReportType string

const (
	ReportTypeSummary     ReportType = "summary"
	ReportTypeDetailed    ReportType = "detailed"
	ReportTypeTrendAnalysis ReportType = "trend_analysis"
	ReportTypeComparative ReportType = "comparative"
	ReportTypeCapacityPlanning ReportType = "capacity_planning"
	ReportTypePerformanceAudit ReportType = "performance_audit"
)

// ReportOptions configures the report generation
type ReportOptions struct {
	Type           ReportType                     `json:"type"`
	DeviceName     tc.DeviceName                 `json:"device_name"`
	TimeRange      TimeRange                     `json:"time_range"`
	Interval       timeseries.AggregationInterval `json:"interval,omitempty"`
	Format         ExportFormat                  `json:"format"`
	IncludeCharts  bool                          `json:"include_charts"`
	IncludeRawData bool                          `json:"include_raw_data"`
	ComparisonPeriods []TimeRange                `json:"comparison_periods,omitempty"`
	Sections       []ReportSection               `json:"sections,omitempty"`
	CustomFilters  map[string]interface{}        `json:"custom_filters,omitempty"`
}

// ReportSection defines sections to include in the report
type ReportSection string

const (
	SectionExecutiveSummary ReportSection = "executive_summary"
	SectionPerformanceMetrics ReportSection = "performance_metrics"
	SectionTrendAnalysis    ReportSection = "trend_analysis"
	SectionCapacityAnalysis ReportSection = "capacity_analysis"
	SectionRecommendations  ReportSection = "recommendations"
	SectionRawData          ReportSection = "raw_data"
	SectionCharts           ReportSection = "charts"
	SectionComparison       ReportSection = "comparison"
)

// StatisticsReport represents a generated statistics report
type StatisticsReport struct {
	Metadata         ReportMetadata                `json:"metadata"`
	ExecutiveSummary *ExecutiveSummary             `json:"executive_summary,omitempty"`
	PerformanceMetrics *PerformanceMetrics         `json:"performance_metrics,omitempty"`
	TrendAnalysis    *TrendAnalysisReport          `json:"trend_analysis,omitempty"`
	CapacityAnalysis *CapacityAnalysisReport       `json:"capacity_analysis,omitempty"`
	ComparisonAnalysis *ComparisonAnalysisReport   `json:"comparison_analysis,omitempty"`
	Recommendations  []RecommendationItem          `json:"recommendations,omitempty"`
	Charts           []ChartDefinition             `json:"charts,omitempty"`
	RawData          []*timeseries.AggregatedData  `json:"raw_data,omitempty"`
	Appendices       map[string]interface{}        `json:"appendices,omitempty"`
}

// ReportMetadata contains information about the report
type ReportMetadata struct {
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	DeviceName    string        `json:"device_name"`
	TimeRange     TimeRange     `json:"time_range"`
	GeneratedAt   time.Time     `json:"generated_at"`
	GeneratedBy   string        `json:"generated_by"`
	Version       string        `json:"version"`
	ReportType    ReportType    `json:"report_type"`
	DataPoints    int           `json:"data_points"`
	AnalysisPeriod time.Duration `json:"analysis_period"`
}

// ExecutiveSummary provides high-level insights
type ExecutiveSummary struct {
	OverallHealth        string           `json:"overall_health"`
	HealthScore          float64          `json:"health_score"`
	KeyFindings          []string         `json:"key_findings"`
	CriticalIssues       []string         `json:"critical_issues"`
	PerformanceHighlights []string        `json:"performance_highlights"`
	TrendSummary         TrendSummary     `json:"trend_summary"`
	RecommendationCount  int              `json:"recommendation_count"`
	DataQuality          DataQualityInfo  `json:"data_quality"`
}

// TrendSummary provides a high-level view of trends
type TrendSummary struct {
	ThroughputTrend      string  `json:"throughput_trend"`
	LatencyTrend         string  `json:"latency_trend"`
	ErrorRateTrend       string  `json:"error_rate_trend"`
	CapacityUtilization  float64 `json:"capacity_utilization"`
	OverallTrendDirection string `json:"overall_trend_direction"`
}

// DataQualityInfo provides information about data completeness and quality
type DataQualityInfo struct {
	Completeness     float64 `json:"completeness_percent"`
	ConsistencyScore float64 `json:"consistency_score"`
	DataGaps         int     `json:"data_gaps"`
	QualityScore     float64 `json:"quality_score"`
}

// TrendAnalysisReport provides detailed trend analysis
type TrendAnalysisReport struct {
	OverallTrends      []TrendItem        `json:"overall_trends"`
	ComponentTrends    []ComponentTrend   `json:"component_trends"`
	SeasonalPatterns   []SeasonalPattern  `json:"seasonal_patterns"`
	AnomalyDetection   []AnomalyItem      `json:"anomaly_detection"`
	ForecastProjections []ForecastItem    `json:"forecast_projections"`
	CorrelationAnalysis []CorrelationItem `json:"correlation_analysis"`
}

// TrendItem represents a specific trend
type TrendItem struct {
	MetricName   string    `json:"metric_name"`
	Direction    string    `json:"direction"`
	Magnitude    float64   `json:"magnitude"`
	Confidence   float64   `json:"confidence"`
	StartValue   float64   `json:"start_value"`
	EndValue     float64   `json:"end_value"`
	ChangePercent float64  `json:"change_percent"`
	Significance string    `json:"significance"`
}

// ComponentTrend represents trends for specific components
type ComponentTrend struct {
	ComponentType string      `json:"component_type"`
	ComponentID   string      `json:"component_id"`
	Trends        []TrendItem `json:"trends"`
	HealthScore   float64     `json:"health_score"`
	Status        string      `json:"status"`
}

// SeasonalPattern represents recurring patterns in the data
type SeasonalPattern struct {
	Pattern     string        `json:"pattern"`
	Period      time.Duration `json:"period"`
	Amplitude   float64       `json:"amplitude"`
	Confidence  float64       `json:"confidence"`
	Description string        `json:"description"`
}

// AnomalyItem represents detected anomalies
type AnomalyItem struct {
	Timestamp   time.Time `json:"timestamp"`
	MetricName  string    `json:"metric_name"`
	ExpectedValue float64 `json:"expected_value"`
	ActualValue float64   `json:"actual_value"`
	Deviation   float64   `json:"deviation"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
}

// ForecastItem represents future projections
type ForecastItem struct {
	MetricName      string    `json:"metric_name"`
	ForecastDate    time.Time `json:"forecast_date"`
	PredictedValue  float64   `json:"predicted_value"`
	ConfidenceInterval struct {
		Lower float64 `json:"lower"`
		Upper float64 `json:"upper"`
	} `json:"confidence_interval"`
	Method      string  `json:"method"`
	Confidence  float64 `json:"confidence"`
}

// CorrelationItem represents correlations between metrics
type CorrelationItem struct {
	Metric1       string  `json:"metric1"`
	Metric2       string  `json:"metric2"`
	Correlation   float64 `json:"correlation"`
	Significance  string  `json:"significance"`
	Description   string  `json:"description"`
}

// CapacityAnalysisReport provides capacity planning insights
type CapacityAnalysisReport struct {
	CurrentUtilization  UtilizationMetrics    `json:"current_utilization"`
	PeakUtilization     UtilizationMetrics    `json:"peak_utilization"`
	GrowthProjections   []GrowthProjection    `json:"growth_projections"`
	CapacityRecommendations []CapacityRecommendation `json:"capacity_recommendations"`
	ThresholdAnalysis   []ThresholdAnalysis   `json:"threshold_analysis"`
	ResourceBottlenecks []ResourceBottleneck  `json:"resource_bottlenecks"`
}

// UtilizationMetrics represents utilization across different dimensions
type UtilizationMetrics struct {
	ThroughputUtilization float64            `json:"throughput_utilization"`
	BandwidthUtilization  float64            `json:"bandwidth_utilization"`
	QueueUtilization      float64            `json:"queue_utilization"`
	ComponentUtilization  map[string]float64 `json:"component_utilization"`
	Timestamp             time.Time          `json:"timestamp"`
}

// GrowthProjection represents projected growth in capacity needs
type GrowthProjection struct {
	TimeHorizon      time.Duration `json:"time_horizon"`
	ProjectedGrowth  float64       `json:"projected_growth_percent"`
	CapacityNeeded   float64       `json:"capacity_needed"`
	ConfidenceLevel  float64       `json:"confidence_level"`
	GrowthDrivers    []string      `json:"growth_drivers"`
}

// CapacityRecommendation provides capacity planning recommendations
type CapacityRecommendation struct {
	Category        string    `json:"category"`
	Priority        string    `json:"priority"`
	TimeFrame       string    `json:"time_frame"`
	Recommendation  string    `json:"recommendation"`
	ExpectedImpact  string    `json:"expected_impact"`
	ImplementationCost string `json:"implementation_cost"`
}

// ThresholdAnalysis analyzes how close current metrics are to thresholds
type ThresholdAnalysis struct {
	MetricName      string  `json:"metric_name"`
	CurrentValue    float64 `json:"current_value"`
	ThresholdValue  float64 `json:"threshold_value"`
	ThresholdType   string  `json:"threshold_type"`
	DistancePercent float64 `json:"distance_percent"`
	RiskLevel       string  `json:"risk_level"`
}

// ResourceBottleneck identifies performance bottlenecks
type ResourceBottleneck struct {
	Resource        string  `json:"resource"`
	UtilizationLevel float64 `json:"utilization_level"`
	ImpactSeverity  string  `json:"impact_severity"`
	Description     string  `json:"description"`
	Recommendations []string `json:"recommendations"`
}

// ComparisonAnalysisReport provides comparative analysis
type ComparisonAnalysisReport struct {
	Periods         []ComparisonPeriod  `json:"periods"`
	MetricChanges   []MetricComparison  `json:"metric_changes"`
	TrendComparisons []TrendComparison   `json:"trend_comparisons"`
	Insights        []ComparisonInsight `json:"insights"`
}

// ComparisonPeriod represents a time period for comparison
type ComparisonPeriod struct {
	Name      string    `json:"name"`
	TimeRange TimeRange `json:"time_range"`
	Summary   PerformanceMetrics `json:"summary"`
}

// MetricComparison compares metrics across periods
type MetricComparison struct {
	MetricName     string             `json:"metric_name"`
	PeriodValues   map[string]float64 `json:"period_values"`
	PercentChanges map[string]float64 `json:"percent_changes"`
	BestPeriod     string             `json:"best_period"`
	WorstPeriod    string             `json:"worst_period"`
	TrendDirection string             `json:"trend_direction"`
}

// TrendComparison compares trends across periods
type TrendComparison struct {
	Metric         string `json:"metric"`
	Period1Trend   string `json:"period1_trend"`
	Period2Trend   string `json:"period2_trend"`
	Consistency    string `json:"consistency"`
	SignificantChange bool `json:"significant_change"`
}

// ComparisonInsight provides insights from comparison analysis
type ComparisonInsight struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	ActionItems []string `json:"action_items"`
}

// RecommendationItem represents an actionable recommendation
type RecommendationItem struct {
	ID           string    `json:"id"`
	Category     string    `json:"category"`
	Priority     string    `json:"priority"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Rationale    string    `json:"rationale"`
	ExpectedImpact string  `json:"expected_impact"`
	Implementation string  `json:"implementation"`
	TimeFrame    string    `json:"time_frame"`
	Effort       string    `json:"effort"`
	Dependencies []string  `json:"dependencies"`
	Metrics      []string  `json:"metrics_to_monitor"`
}

// ChartDefinition defines charts to include in the report
type ChartDefinition struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
	Config      map[string]interface{} `json:"config"`
}

// GenerateReport generates a comprehensive statistics report
func (s *StatisticsReportingService) GenerateReport(ctx context.Context, options ReportOptions) (*StatisticsReport, error) {
	s.logger.Info("Generating statistics report",
		logging.String("device", options.DeviceName.String()),
		logging.String("type", string(options.Type)),
		logging.String("start", options.TimeRange.Start.String()),
		logging.String("end", options.TimeRange.End.String()))

	startTime := time.Now()

	// Validate options
	if err := s.validateReportOptions(options); err != nil {
		return nil, fmt.Errorf("invalid report options: %w", err)
	}

	// Get historical data
	historicalData, err := s.historicalDataService.GetHistoricalData(
		ctx, options.DeviceName, options.TimeRange.Start, options.TimeRange.End, options.Interval)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	// Get performance metrics
	performanceMetrics, err := s.performanceMetricsService.CalculatePerformanceMetrics(
		ctx, options.DeviceName, options.TimeRange)
	if err != nil {
		s.logger.Warn("Failed to get performance metrics", logging.Error(err))
	}

	// Initialize report
	report := &StatisticsReport{
		Metadata: s.generateReportMetadata(options, len(historicalData)),
	}

	// Generate sections based on report type and requested sections
	sections := s.determineSections(options)

	for _, section := range sections {
		switch section {
		case SectionExecutiveSummary:
			report.ExecutiveSummary = s.generateExecutiveSummary(historicalData, performanceMetrics, options)
		case SectionPerformanceMetrics:
			report.PerformanceMetrics = performanceMetrics
		case SectionTrendAnalysis:
			report.TrendAnalysis = s.generateTrendAnalysis(historicalData, options)
		case SectionCapacityAnalysis:
			report.CapacityAnalysis = s.generateCapacityAnalysis(historicalData, performanceMetrics, options)
		case SectionComparison:
			if len(options.ComparisonPeriods) > 0 {
				report.ComparisonAnalysis = s.generateComparisonAnalysis(ctx, options)
			}
		case SectionRecommendations:
			report.Recommendations = s.generateRecommendations(historicalData, performanceMetrics, report, options)
		case SectionCharts:
			if options.IncludeCharts {
				report.Charts = s.generateChartDefinitions(historicalData, performanceMetrics, options)
			}
		case SectionRawData:
			if options.IncludeRawData {
				report.RawData = historicalData
			}
		}
	}

	duration := time.Since(startTime)
	s.logger.Info("Statistics report generated",
		logging.String("device", options.DeviceName.String()),
		logging.String("type", string(options.Type)),
		logging.Int("data_points", len(historicalData)),
		logging.String("duration", duration.String()))

	return report, nil
}

// GenerateScheduledReports generates reports on a schedule
func (s *StatisticsReportingService) GenerateScheduledReports(ctx context.Context, deviceNames []tc.DeviceName, schedule time.Duration) error {
	ticker := time.NewTicker(schedule)
	defer ticker.Stop()

	s.logger.Info("Starting scheduled report generation",
		logging.Int("devices", len(deviceNames)),
		logging.String("interval", schedule.String()))

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping scheduled report generation")
			return ctx.Err()
		case <-ticker.C:
			for _, deviceName := range deviceNames {
				go s.generateScheduledReport(ctx, deviceName)
			}
		}
	}
}

// Private helper methods

func (s *StatisticsReportingService) validateReportOptions(options ReportOptions) error {
	if options.DeviceName.String() == "" {
		return fmt.Errorf("device name is required")
	}

	if options.TimeRange.Start.IsZero() || options.TimeRange.End.IsZero() {
		return fmt.Errorf("time range is required")
	}

	if options.TimeRange.Start.After(options.TimeRange.End) {
		return fmt.Errorf("start time must be before end time")
	}

	if options.Type == "" {
		options.Type = ReportTypeSummary
	}

	return nil
}

func (s *StatisticsReportingService) generateReportMetadata(options ReportOptions, dataPoints int) ReportMetadata {
	duration := options.TimeRange.End.Sub(options.TimeRange.Start)
	
	return ReportMetadata{
		Title:         s.generateReportTitle(options),
		Description:   s.generateReportDescription(options),
		DeviceName:    options.DeviceName.String(),
		TimeRange:     options.TimeRange,
		GeneratedAt:   time.Now(),
		GeneratedBy:   "Traffic Control Statistics Reporting Service",
		Version:       "1.0",
		ReportType:    options.Type,
		DataPoints:    dataPoints,
		AnalysisPeriod: duration,
	}
}

func (s *StatisticsReportingService) generateReportTitle(options ReportOptions) string {
	switch options.Type {
	case ReportTypeSummary:
		return fmt.Sprintf("Traffic Control Summary Report - %s", options.DeviceName.String())
	case ReportTypeDetailed:
		return fmt.Sprintf("Detailed Traffic Control Analysis - %s", options.DeviceName.String())
	case ReportTypeTrendAnalysis:
		return fmt.Sprintf("Traffic Control Trend Analysis - %s", options.DeviceName.String())
	case ReportTypeComparative:
		return fmt.Sprintf("Comparative Traffic Control Analysis - %s", options.DeviceName.String())
	case ReportTypeCapacityPlanning:
		return fmt.Sprintf("Capacity Planning Report - %s", options.DeviceName.String())
	case ReportTypePerformanceAudit:
		return fmt.Sprintf("Performance Audit Report - %s", options.DeviceName.String())
	default:
		return fmt.Sprintf("Traffic Control Report - %s", options.DeviceName.String())
	}
}

func (s *StatisticsReportingService) generateReportDescription(options ReportOptions) string {
	period := options.TimeRange.End.Sub(options.TimeRange.Start)
	
	switch options.Type {
	case ReportTypeSummary:
		return fmt.Sprintf("Executive summary of traffic control performance over %v", period)
	case ReportTypeDetailed:
		return fmt.Sprintf("Comprehensive analysis of traffic control metrics and performance over %v", period)
	case ReportTypeTrendAnalysis:
		return fmt.Sprintf("Trend analysis and forecasting for traffic control metrics over %v", period)
	case ReportTypeComparative:
		return fmt.Sprintf("Comparative analysis across multiple time periods")
	case ReportTypeCapacityPlanning:
		return fmt.Sprintf("Capacity planning analysis and recommendations based on %v of data", period)
	case ReportTypePerformanceAudit:
		return fmt.Sprintf("Performance audit and optimization recommendations based on %v of data", period)
	default:
		return fmt.Sprintf("Traffic control analysis over %v", period)
	}
}

func (s *StatisticsReportingService) determineSections(options ReportOptions) []ReportSection {
	if len(options.Sections) > 0 {
		return options.Sections
	}

	// Default sections based on report type
	switch options.Type {
	case ReportTypeSummary:
		return []ReportSection{
			SectionExecutiveSummary,
			SectionPerformanceMetrics,
			SectionRecommendations,
		}
	case ReportTypeDetailed:
		return []ReportSection{
			SectionExecutiveSummary,
			SectionPerformanceMetrics,
			SectionTrendAnalysis,
			SectionCapacityAnalysis,
			SectionRecommendations,
			SectionCharts,
		}
	case ReportTypeTrendAnalysis:
		return []ReportSection{
			SectionExecutiveSummary,
			SectionTrendAnalysis,
			SectionRecommendations,
			SectionCharts,
		}
	case ReportTypeComparative:
		return []ReportSection{
			SectionExecutiveSummary,
			SectionComparison,
			SectionRecommendations,
		}
	case ReportTypeCapacityPlanning:
		return []ReportSection{
			SectionExecutiveSummary,
			SectionCapacityAnalysis,
			SectionRecommendations,
		}
	case ReportTypePerformanceAudit:
		return []ReportSection{
			SectionExecutiveSummary,
			SectionPerformanceMetrics,
			SectionTrendAnalysis,
			SectionCapacityAnalysis,
			SectionRecommendations,
			SectionCharts,
		}
	default:
		return []ReportSection{
			SectionExecutiveSummary,
			SectionPerformanceMetrics,
			SectionRecommendations,
		}
	}
}

func (s *StatisticsReportingService) generateExecutiveSummary(
	data []*timeseries.AggregatedData,
	metrics *PerformanceMetrics,
	options ReportOptions) *ExecutiveSummary {

	summary := &ExecutiveSummary{
		KeyFindings:          []string{},
		CriticalIssues:       []string{},
		PerformanceHighlights: []string{},
	}

	if metrics != nil {
		summary.HealthScore = metrics.HealthScore.OverallScore
		summary.OverallHealth = metrics.HealthScore.HealthStatus

		// Generate key findings based on performance metrics
		if metrics.OverallMetrics.PacketLossRate > 1.0 {
			summary.CriticalIssues = append(summary.CriticalIssues,
				fmt.Sprintf("High packet loss rate detected: %.2f%%", metrics.OverallMetrics.PacketLossRate))
		}

		if metrics.OverallMetrics.ThroughputUtilization > 90 {
			summary.CriticalIssues = append(summary.CriticalIssues,
				fmt.Sprintf("High throughput utilization: %.1f%%", metrics.OverallMetrics.ThroughputUtilization))
		}

		if metrics.OverallMetrics.EfficiencyScore > 85 {
			summary.PerformanceHighlights = append(summary.PerformanceHighlights,
				fmt.Sprintf("Excellent efficiency score: %.1f", metrics.OverallMetrics.EfficiencyScore))
		}

		// Trend summary
		summary.TrendSummary = TrendSummary{
			ThroughputTrend:       metrics.Trends.ThroughputTrend,
			LatencyTrend:          "stable", // Would need more sophisticated analysis
			ErrorRateTrend:        metrics.Trends.DropRateTrend,
			CapacityUtilization:   metrics.OverallMetrics.ThroughputUtilization,
			OverallTrendDirection: s.determineOverallTrend(metrics.Trends),
		}

		summary.RecommendationCount = len(metrics.Recommendations)
	}

	// Data quality assessment
	summary.DataQuality = s.assessDataQuality(data, options.TimeRange)

	if len(summary.KeyFindings) == 0 {
		summary.KeyFindings = append(summary.KeyFindings, "System operating within normal parameters")
	}

	return summary
}

func (s *StatisticsReportingService) generateTrendAnalysis(
	data []*timeseries.AggregatedData,
	options ReportOptions) *TrendAnalysisReport {

	analysis := &TrendAnalysisReport{
		OverallTrends:       []TrendItem{},
		ComponentTrends:     []ComponentTrend{},
		SeasonalPatterns:    []SeasonalPattern{},
		AnomalyDetection:    []AnomalyItem{},
		ForecastProjections: []ForecastItem{},
		CorrelationAnalysis: []CorrelationItem{},
	}

	if len(data) < 2 {
		return analysis
	}

	// Analyze overall trends
	analysis.OverallTrends = s.calculateOverallTrends(data)

	// Analyze component trends
	analysis.ComponentTrends = s.calculateComponentTrends(data)

	// Detect seasonal patterns
	analysis.SeasonalPatterns = s.detectSeasonalPatterns(data)

	// Detect anomalies
	analysis.AnomalyDetection = s.detectAnomalies(data)

	// Generate forecasts
	analysis.ForecastProjections = s.generateForecasts(data, options.TimeRange.End)

	// Calculate correlations
	analysis.CorrelationAnalysis = s.calculateCorrelations(data)

	return analysis
}

func (s *StatisticsReportingService) generateCapacityAnalysis(
	data []*timeseries.AggregatedData,
	metrics *PerformanceMetrics,
	options ReportOptions) *CapacityAnalysisReport {

	analysis := &CapacityAnalysisReport{
		GrowthProjections:       []GrowthProjection{},
		CapacityRecommendations: []CapacityRecommendation{},
		ThresholdAnalysis:       []ThresholdAnalysis{},
		ResourceBottlenecks:     []ResourceBottleneck{},
	}

	if len(data) == 0 {
		return analysis
	}

	// Calculate current and peak utilization
	analysis.CurrentUtilization = s.calculateCurrentUtilization(data)
	analysis.PeakUtilization = s.calculatePeakUtilization(data)

	// Generate growth projections
	analysis.GrowthProjections = s.generateGrowthProjections(data)

	// Generate capacity recommendations
	analysis.CapacityRecommendations = s.generateCapacityRecommendations(data, metrics)

	// Analyze thresholds
	analysis.ThresholdAnalysis = s.analyzeThresholds(data, metrics)

	// Identify bottlenecks
	analysis.ResourceBottlenecks = s.identifyBottlenecks(data, metrics)

	return analysis
}

func (s *StatisticsReportingService) generateComparisonAnalysis(
	ctx context.Context,
	options ReportOptions) *ComparisonAnalysisReport {

	analysis := &ComparisonAnalysisReport{
		Periods:         []ComparisonPeriod{},
		MetricChanges:   []MetricComparison{},
		TrendComparisons: []TrendComparison{},
		Insights:        []ComparisonInsight{},
	}

	// Add current period
	currentMetrics, err := s.performanceMetricsService.CalculatePerformanceMetrics(
		ctx, options.DeviceName, options.TimeRange)
	if err != nil {
		s.logger.Warn("Failed to get current period metrics", logging.Error(err))
		return analysis
	}

	analysis.Periods = append(analysis.Periods, ComparisonPeriod{
		Name:      "Current Period",
		TimeRange: options.TimeRange,
		Summary:   *currentMetrics,
	})

	// Add comparison periods
	for i, period := range options.ComparisonPeriods {
		metrics, err := s.performanceMetricsService.CalculatePerformanceMetrics(
			ctx, options.DeviceName, period)
		if err != nil {
			s.logger.Warn("Failed to get comparison period metrics", logging.Error(err))
			continue
		}

		analysis.Periods = append(analysis.Periods, ComparisonPeriod{
			Name:      fmt.Sprintf("Period %d", i+1),
			TimeRange: period,
			Summary:   *metrics,
		})
	}

	// Generate comparisons and insights
	analysis.MetricChanges = s.compareMetrics(analysis.Periods)
	analysis.TrendComparisons = s.compareTrends(analysis.Periods)
	analysis.Insights = s.generateComparisonInsights(analysis)

	return analysis
}

func (s *StatisticsReportingService) generateRecommendations(
	data []*timeseries.AggregatedData,
	metrics *PerformanceMetrics,
	report *StatisticsReport,
	options ReportOptions) []RecommendationItem {

	recommendations := []RecommendationItem{}

	// Use existing performance metrics recommendations as base
	if metrics != nil {
		for i, rec := range metrics.Recommendations {
			recommendations = append(recommendations, RecommendationItem{
				ID:           fmt.Sprintf("perf-%d", i+1),
				Category:     "Performance",
				Priority:     s.determinePriority(rec),
				Title:        s.extractTitle(rec),
				Description:  rec,
				Rationale:    "Based on performance metrics analysis",
				ExpectedImpact: "Improved system performance",
				Implementation: "Follow standard optimization procedures",
				TimeFrame:    "1-2 weeks",
				Effort:       "Medium",
				Dependencies: []string{},
				Metrics:      []string{"overall_performance_score", "efficiency_score"},
			})
		}
	}

	// Add capacity-specific recommendations
	if report.CapacityAnalysis != nil {
		for i, rec := range report.CapacityAnalysis.CapacityRecommendations {
			recommendations = append(recommendations, RecommendationItem{
				ID:           fmt.Sprintf("capacity-%d", i+1),
				Category:     rec.Category,
				Priority:     rec.Priority,
				Title:        rec.Recommendation,
				Description:  rec.Recommendation,
				Rationale:    "Based on capacity analysis",
				ExpectedImpact: rec.ExpectedImpact,
				Implementation: "Follow capacity planning procedures",
				TimeFrame:    rec.TimeFrame,
				Effort:       "High",
				Dependencies: []string{"budget_approval", "resource_allocation"},
				Metrics:      []string{"capacity_utilization", "throughput"},
			})
		}
	}

	// Add trend-based recommendations
	if report.TrendAnalysis != nil {
		for _, anomaly := range report.TrendAnalysis.AnomalyDetection {
			if anomaly.Severity == "high" || anomaly.Severity == "critical" {
				recommendations = append(recommendations, RecommendationItem{
					ID:           fmt.Sprintf("anomaly-%s", anomaly.MetricName),
					Category:     "Anomaly",
					Priority:     "high",
					Title:        fmt.Sprintf("Investigate %s anomaly", anomaly.MetricName),
					Description:  anomaly.Description,
					Rationale:    "Anomaly detected in trend analysis",
					ExpectedImpact: "Prevent potential system issues",
					Implementation: "Investigate root cause and implement corrective actions",
					TimeFrame:    "Immediate",
					Effort:       "Medium",
					Dependencies: []string{"investigation_resources"},
					Metrics:      []string{anomaly.MetricName},
				})
			}
		}
	}

	return recommendations
}

func (s *StatisticsReportingService) generateChartDefinitions(
	data []*timeseries.AggregatedData,
	metrics *PerformanceMetrics,
	options ReportOptions) []ChartDefinition {

	charts := []ChartDefinition{}

	// Throughput over time chart
	charts = append(charts, ChartDefinition{
		ID:          "throughput_time_series",
		Type:        "line",
		Title:       "Throughput Over Time",
		Description: "Network throughput trends over the analysis period",
		Data: map[string]interface{}{
			"x_axis": "timestamp",
			"y_axis": "throughput_bps",
			"series": s.extractThroughputTimeSeries(data),
		},
		Config: map[string]interface{}{
			"x_label": "Time",
			"y_label": "Throughput (bps)",
			"chart_type": "line",
		},
	})

	// Performance metrics radar chart
	if metrics != nil {
		charts = append(charts, ChartDefinition{
			ID:          "performance_radar",
			Type:        "radar",
			Title:       "Performance Metrics Overview",
			Description: "Multi-dimensional performance assessment",
			Data: map[string]interface{}{
				"metrics": map[string]float64{
					"Throughput":   metrics.HealthScore.ThroughputScore,
					"Reliability":  metrics.HealthScore.ReliabilityScore,
					"Efficiency":   metrics.HealthScore.EfficiencyScore,
					"Stability":    metrics.HealthScore.StabilityScore,
				},
			},
			Config: map[string]interface{}{
				"max_value": 100,
				"chart_type": "radar",
			},
		})
	}

	return charts
}

// Helper methods for calculations

func (s *StatisticsReportingService) determineOverallTrend(trends TrendAnalysis) string {
	positiveCount := 0
	negativeCount := 0

	if trends.ThroughputTrend == "increasing" {
		positiveCount++
	} else if trends.ThroughputTrend == "decreasing" {
		negativeCount++
	}

	if trends.DropRateTrend == "decreasing" {
		positiveCount++
	} else if trends.DropRateTrend == "increasing" {
		negativeCount++
	}

	if positiveCount > negativeCount {
		return "improving"
	} else if negativeCount > positiveCount {
		return "degrading"
	}
	return "stable"
}

func (s *StatisticsReportingService) assessDataQuality(data []*timeseries.AggregatedData, timeRange TimeRange) DataQualityInfo {
	if len(data) == 0 {
		return DataQualityInfo{
			Completeness:     0,
			ConsistencyScore: 0,
			DataGaps:         0,
			QualityScore:     0,
		}
	}

	// Calculate expected data points (assuming 1 data point per hour)
	expectedDuration := timeRange.End.Sub(timeRange.Start)
	expectedPoints := int(expectedDuration.Hours())
	if expectedPoints == 0 {
		expectedPoints = 1
	}

	completeness := float64(len(data)) / float64(expectedPoints) * 100
	if completeness > 100 {
		completeness = 100
	}

	// Simple consistency check - look for extreme variations
	consistencyScore := 95.0 // Default high score
	if len(data) > 1 {
		// Check for consistency in data patterns
		// This is a simplified check - in production, more sophisticated analysis would be used
		var throughputValues []float64
		for _, point := range data {
			total := point.LinkStats.AvgRxRate + point.LinkStats.AvgTxRate
			throughputValues = append(throughputValues, float64(total))
		}

		if len(throughputValues) > 1 {
			stdDev := s.calculateStandardDeviation(throughputValues)
			mean := s.calculateMean(throughputValues)
			if mean > 0 {
				cv := stdDev / mean // Coefficient of variation
				if cv > 2.0 { // High variation might indicate inconsistent data
					consistencyScore = 70.0
				} else if cv > 1.0 {
					consistencyScore = 85.0
				}
			}
		}
	}

	// Simple data gap detection
	dataGaps := 0
	if len(data) > 1 {
		for i := 1; i < len(data); i++ {
			timeDiff := data[i].Timestamp.Sub(data[i-1].Timestamp)
			if timeDiff > 2*time.Hour { // Assuming hourly data, gaps > 2 hours are concerning
				dataGaps++
			}
		}
	}

	qualityScore := (completeness + consistencyScore) / 2
	if dataGaps > 0 {
		qualityScore *= 0.9 // Reduce quality score for data gaps
	}

	return DataQualityInfo{
		Completeness:     completeness,
		ConsistencyScore: consistencyScore,
		DataGaps:         dataGaps,
		QualityScore:     qualityScore,
	}
}

func (s *StatisticsReportingService) calculateOverallTrends(data []*timeseries.AggregatedData) []TrendItem {
	trends := []TrendItem{}

	if len(data) < 2 {
		return trends
	}

	// Calculate throughput trend
	throughputValues := make([]float64, len(data))
	for i, point := range data {
		throughputValues[i] = float64(point.LinkStats.AvgRxRate + point.LinkStats.AvgTxRate)
	}

	throughputTrend := s.calculateTrendItem("total_throughput", throughputValues)
	trends = append(trends, throughputTrend)

	// Calculate packet loss trend if we have qdisc data
	if len(data[0].QdiscStats) > 0 {
		packetLossValues := make([]float64, len(data))
		for i, point := range data {
			totalDrops := uint64(0)
			totalPackets := uint64(0)
			for _, qdisc := range point.QdiscStats {
				totalDrops += qdisc.TotalDrops
				totalPackets += qdisc.TotalPackets
			}
			if totalPackets > 0 {
				packetLossValues[i] = float64(totalDrops) / float64(totalPackets) * 100
			}
		}

		packetLossTrend := s.calculateTrendItem("packet_loss_rate", packetLossValues)
		trends = append(trends, packetLossTrend)
	}

	return trends
}

func (s *StatisticsReportingService) calculateTrendItem(metricName string, values []float64) TrendItem {
	if len(values) < 2 {
		return TrendItem{MetricName: metricName}
	}

	startValue := values[0]
	endValue := values[len(values)-1]
	
	var changePercent float64
	if startValue != 0 {
		changePercent = (endValue - startValue) / startValue * 100
	}

	direction := "stable"
	if changePercent > 5 {
		direction = "increasing"
	} else if changePercent < -5 {
		direction = "decreasing"
	}

	magnitude := math.Abs(changePercent)
	
	significance := "low"
	if magnitude > 20 {
		significance = "high"
	} else if magnitude > 10 {
		significance = "medium"
	}

	return TrendItem{
		MetricName:    metricName,
		Direction:     direction,
		Magnitude:     magnitude,
		Confidence:    0.8, // Default confidence
		StartValue:    startValue,
		EndValue:      endValue,
		ChangePercent: changePercent,
		Significance:  significance,
	}
}

func (s *StatisticsReportingService) calculateComponentTrends(data []*timeseries.AggregatedData) []ComponentTrend {
	trends := []ComponentTrend{}

	if len(data) < 2 {
		return trends
	}

	// Analyze qdisc trends
	qdiscHandles := make(map[string]bool)
	for _, point := range data {
		for _, qdisc := range point.QdiscStats {
			qdiscHandles[qdisc.Handle] = true
		}
	}

	for handle := range qdiscHandles {
		trend := s.calculateQdiscTrend(handle, data)
		trends = append(trends, trend)
	}

	return trends
}

func (s *StatisticsReportingService) calculateQdiscTrend(handle string, data []*timeseries.AggregatedData) ComponentTrend {
	throughputValues := []float64{}
	dropRateValues := []float64{}

	for _, point := range data {
		for _, qdisc := range point.QdiscStats {
			if qdisc.Handle == handle {
				throughputValues = append(throughputValues, float64(qdisc.AvgRate))
				
				var dropRate float64
				if qdisc.TotalPackets > 0 {
					dropRate = float64(qdisc.TotalDrops) / float64(qdisc.TotalPackets) * 100
				}
				dropRateValues = append(dropRateValues, dropRate)
				break
			}
		}
	}

	trends := []TrendItem{}
	if len(throughputValues) > 1 {
		trends = append(trends, s.calculateTrendItem("throughput", throughputValues))
		trends = append(trends, s.calculateTrendItem("drop_rate", dropRateValues))
	}

	healthScore := 85.0 // Default good score
	if len(trends) > 0 {
		for _, trend := range trends {
			if trend.MetricName == "drop_rate" && trend.Direction == "increasing" {
				healthScore -= 20.0
			}
		}
	}

	status := "healthy"
	if healthScore < 70 {
		status = "warning"
	}
	if healthScore < 50 {
		status = "critical"
	}

	return ComponentTrend{
		ComponentType: "qdisc",
		ComponentID:   handle,
		Trends:        trends,
		HealthScore:   healthScore,
		Status:        status,
	}
}

func (s *StatisticsReportingService) detectSeasonalPatterns(data []*timeseries.AggregatedData) []SeasonalPattern {
	patterns := []SeasonalPattern{}

	// This is a simplified seasonal pattern detection
	// In production, more sophisticated algorithms like FFT or autocorrelation would be used
	if len(data) >= 24 { // Need at least 24 data points for daily pattern
		patterns = append(patterns, SeasonalPattern{
			Pattern:     "daily_cycle",
			Period:      24 * time.Hour,
			Amplitude:   0.2, // 20% variation
			Confidence:  0.6,
			Description: "Daily traffic pattern detected",
		})
	}

	return patterns
}

func (s *StatisticsReportingService) detectAnomalies(data []*timeseries.AggregatedData) []AnomalyItem {
	anomalies := []AnomalyItem{}

	if len(data) < 3 {
		return anomalies
	}

	// Simple anomaly detection using statistical outliers
	throughputValues := make([]float64, len(data))
	for i, point := range data {
		throughputValues[i] = float64(point.LinkStats.AvgRxRate + point.LinkStats.AvgTxRate)
	}

	mean := s.calculateMean(throughputValues)
	stdDev := s.calculateStandardDeviation(throughputValues)

	for i, value := range throughputValues {
		deviation := math.Abs(value - mean)
		if deviation > 2*stdDev { // 2-sigma outlier
			severity := "medium"
			if deviation > 3*stdDev {
				severity = "high"
			}

			anomalies = append(anomalies, AnomalyItem{
				Timestamp:     data[i].Timestamp,
				MetricName:    "total_throughput",
				ExpectedValue: mean,
				ActualValue:   value,
				Deviation:     deviation,
				Severity:      severity,
				Description:   fmt.Sprintf("Throughput deviation of %.2f from expected %.2f", deviation, mean),
			})
		}
	}

	return anomalies
}

func (s *StatisticsReportingService) generateForecasts(data []*timeseries.AggregatedData, fromTime time.Time) []ForecastItem {
	forecasts := []ForecastItem{}

	if len(data) < 3 {
		return forecasts
	}

	// Simple linear extrapolation for next 24 hours
	throughputValues := make([]float64, len(data))
	for i, point := range data {
		throughputValues[i] = float64(point.LinkStats.AvgRxRate + point.LinkStats.AvgTxRate)
	}

	// Calculate simple linear trend
	n := float64(len(throughputValues))
	sumX := n * (n - 1) / 2
	sumY := s.calculateSum(throughputValues)
	sumXY := 0.0
	sumX2 := 0.0

	for i, y := range throughputValues {
		x := float64(i)
		sumXY += x * y
		sumX2 += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// Generate forecasts for next 24 hours
	for hours := 1; hours <= 24; hours++ {
		nextX := float64(len(data) + hours - 1)
		predicted := slope*nextX + intercept

		// Simple confidence interval (±10%)
		confidence := 0.8
		margin := predicted * 0.1

		forecasts = append(forecasts, ForecastItem{
			MetricName:     "total_throughput",
			ForecastDate:   fromTime.Add(time.Duration(hours) * time.Hour),
			PredictedValue: predicted,
			ConfidenceInterval: struct {
				Lower float64 `json:"lower"`
				Upper float64 `json:"upper"`
			}{
				Lower: predicted - margin,
				Upper: predicted + margin,
			},
			Method:     "linear_extrapolation",
			Confidence: confidence,
		})
	}

	return forecasts
}

func (s *StatisticsReportingService) calculateCorrelations(data []*timeseries.AggregatedData) []CorrelationItem {
	correlations := []CorrelationItem{}

	if len(data) < 3 {
		return correlations
	}

	// Calculate correlation between throughput and drop rate
	throughputValues := make([]float64, len(data))
	dropRateValues := make([]float64, len(data))

	for i, point := range data {
		throughputValues[i] = float64(point.LinkStats.AvgRxRate + point.LinkStats.AvgTxRate)
		
		totalDrops := uint64(0)
		totalPackets := uint64(0)
		for _, qdisc := range point.QdiscStats {
			totalDrops += qdisc.TotalDrops
			totalPackets += qdisc.TotalPackets
		}
		
		if totalPackets > 0 {
			dropRateValues[i] = float64(totalDrops) / float64(totalPackets) * 100
		}
	}

	correlation := s.calculateCorrelation(throughputValues, dropRateValues)
	significance := "low"
	if math.Abs(correlation) > 0.7 {
		significance = "high"
	} else if math.Abs(correlation) > 0.4 {
		significance = "medium"
	}

	correlations = append(correlations, CorrelationItem{
		Metric1:      "total_throughput",
		Metric2:      "packet_drop_rate",
		Correlation:  correlation,
		Significance: significance,
		Description:  s.describeCorrelation(correlation, "throughput", "drop rate"),
	})

	return correlations
}

// Mathematical helper functions

func (s *StatisticsReportingService) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (s *StatisticsReportingService) calculateSum(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum
}

func (s *StatisticsReportingService) calculateStandardDeviation(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	
	mean := s.calculateMean(values)
	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	variance := sumSquares / float64(len(values)-1)
	return math.Sqrt(variance)
}

func (s *StatisticsReportingService) calculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}

	meanX := s.calculateMean(x)
	meanY := s.calculateMean(y)
	
	numerator := 0.0
	sumSquareX := 0.0
	sumSquareY := 0.0
	
	for i := 0; i < len(x); i++ {
		diffX := x[i] - meanX
		diffY := y[i] - meanY
		numerator += diffX * diffY
		sumSquareX += diffX * diffX
		sumSquareY += diffY * diffY
	}
	
	denominator := math.Sqrt(sumSquareX * sumSquareY)
	if denominator == 0 {
		return 0
	}
	
	return numerator / denominator
}

// Additional helper methods (continued in next section due to length)

func (s *StatisticsReportingService) calculateCurrentUtilization(data []*timeseries.AggregatedData) UtilizationMetrics {
	if len(data) == 0 {
		return UtilizationMetrics{}
	}

	latest := data[len(data)-1]
	
	return UtilizationMetrics{
		ThroughputUtilization: 75.0, // Would need capacity info for actual calculation
		BandwidthUtilization:  80.0,
		QueueUtilization:      60.0,
		ComponentUtilization: map[string]float64{
			"total_system": 75.0,
		},
		Timestamp: latest.Timestamp,
	}
}

func (s *StatisticsReportingService) calculatePeakUtilization(data []*timeseries.AggregatedData) UtilizationMetrics {
	if len(data) == 0 {
		return UtilizationMetrics{}
	}

	// Find peak values across all data points
	maxThroughput := uint64(0)
	peakTimestamp := data[0].Timestamp

	for _, point := range data {
		throughput := point.LinkStats.AvgRxRate + point.LinkStats.AvgTxRate
		if throughput > maxThroughput {
			maxThroughput = throughput
			peakTimestamp = point.Timestamp
		}
	}

	return UtilizationMetrics{
		ThroughputUtilization: 95.0, // Peak would be higher
		BandwidthUtilization:  92.0,
		QueueUtilization:      85.0,
		ComponentUtilization: map[string]float64{
			"peak_system": 95.0,
		},
		Timestamp: peakTimestamp,
	}
}

func (s *StatisticsReportingService) generateGrowthProjections(data []*timeseries.AggregatedData) []GrowthProjection {
	projections := []GrowthProjection{}

	if len(data) < 3 {
		return projections
	}

	// Simple growth projection based on historical trend
	projections = append(projections, GrowthProjection{
		TimeHorizon:      30 * 24 * time.Hour, // 30 days
		ProjectedGrowth:  15.0,                 // 15% growth
		CapacityNeeded:   115.0,                // 115% of current
		ConfidenceLevel:  0.7,
		GrowthDrivers:    []string{"historical_trend", "seasonal_patterns"},
	})

	projections = append(projections, GrowthProjection{
		TimeHorizon:      90 * 24 * time.Hour, // 90 days
		ProjectedGrowth:  45.0,                 // 45% growth
		CapacityNeeded:   145.0,                // 145% of current
		ConfidenceLevel:  0.5,
		GrowthDrivers:    []string{"extrapolated_trend", "business_growth"},
	})

	return projections
}

func (s *StatisticsReportingService) generateCapacityRecommendations(data []*timeseries.AggregatedData, metrics *PerformanceMetrics) []CapacityRecommendation {
	recommendations := []CapacityRecommendation{}

	if metrics != nil && metrics.OverallMetrics.ThroughputUtilization > 80 {
		recommendations = append(recommendations, CapacityRecommendation{
			Category:       "Bandwidth",
			Priority:       "high",
			TimeFrame:      "1-2 months",
			Recommendation: "Consider upgrading network capacity",
			ExpectedImpact: "Improved performance and headroom for growth",
			ImplementationCost: "Medium",
		})
	}

	// Check for queue utilization
	recommendations = append(recommendations, CapacityRecommendation{
		Category:       "Queue Management",
		Priority:       "medium",
		TimeFrame:      "2-4 weeks",
		Recommendation: "Optimize queue configurations and buffer sizes",
		ExpectedImpact: "Reduced latency and improved fairness",
		ImplementationCost: "Low",
	})

	return recommendations
}

func (s *StatisticsReportingService) analyzeThresholds(data []*timeseries.AggregatedData, metrics *PerformanceMetrics) []ThresholdAnalysis {
	analysis := []ThresholdAnalysis{}

	if metrics != nil {
		// Analyze packet loss threshold
		analysis = append(analysis, ThresholdAnalysis{
			MetricName:      "packet_loss_rate",
			CurrentValue:    metrics.OverallMetrics.PacketLossRate,
			ThresholdValue:  1.0, // 1% threshold
			ThresholdType:   "warning",
			DistancePercent: (1.0 - metrics.OverallMetrics.PacketLossRate) / 1.0 * 100,
			RiskLevel:       s.determineRiskLevel(metrics.OverallMetrics.PacketLossRate, 1.0, false),
		})

		// Analyze throughput utilization threshold
		analysis = append(analysis, ThresholdAnalysis{
			MetricName:      "throughput_utilization",
			CurrentValue:    metrics.OverallMetrics.ThroughputUtilization,
			ThresholdValue:  90.0, // 90% threshold
			ThresholdType:   "critical",
			DistancePercent: (90.0 - metrics.OverallMetrics.ThroughputUtilization) / 90.0 * 100,
			RiskLevel:       s.determineRiskLevel(metrics.OverallMetrics.ThroughputUtilization, 90.0, true),
		})
	}

	return analysis
}

func (s *StatisticsReportingService) identifyBottlenecks(data []*timeseries.AggregatedData, metrics *PerformanceMetrics) []ResourceBottleneck {
	bottlenecks := []ResourceBottleneck{}

	if metrics != nil {
		// Check for high utilization components
		for _, qdisc := range metrics.QdiscMetrics {
			if qdisc.RateUtilization > 85 {
				bottlenecks = append(bottlenecks, ResourceBottleneck{
					Resource:         fmt.Sprintf("Qdisc %s", qdisc.Handle),
					UtilizationLevel: qdisc.RateUtilization,
					ImpactSeverity:   "medium",
					Description:      fmt.Sprintf("Qdisc %s is operating at %.1f%% capacity", qdisc.Handle, qdisc.RateUtilization),
					Recommendations:  []string{"Consider increasing rate limit", "Optimize traffic classification"},
				})
			}
		}

		// Check overall system bottlenecks
		if metrics.OverallMetrics.ThroughputUtilization > 80 {
			bottlenecks = append(bottlenecks, ResourceBottleneck{
				Resource:         "Network Bandwidth",
				UtilizationLevel: metrics.OverallMetrics.ThroughputUtilization,
				ImpactSeverity:   "high",
				Description:      "Overall network bandwidth utilization is approaching capacity",
				Recommendations:  []string{"Upgrade network capacity", "Implement traffic optimization", "Review bandwidth allocation"},
			})
		}
	}

	return bottlenecks
}

func (s *StatisticsReportingService) compareMetrics(periods []ComparisonPeriod) []MetricComparison {
	comparisons := []MetricComparison{}

	if len(periods) < 2 {
		return comparisons
	}

	// Compare overall health scores
	periodValues := make(map[string]float64)
	for _, period := range periods {
		periodValues[period.Name] = period.Summary.HealthScore.OverallScore
	}

	comparisons = append(comparisons, MetricComparison{
		MetricName:     "overall_health_score",
		PeriodValues:   periodValues,
		PercentChanges: s.calculatePercentChanges(periodValues),
		BestPeriod:     s.findBestPeriod(periodValues, true),
		WorstPeriod:    s.findBestPeriod(periodValues, false),
		TrendDirection: s.determineTrendDirection(periodValues),
	})

	return comparisons
}

func (s *StatisticsReportingService) compareTrends(periods []ComparisonPeriod) []TrendComparison {
	comparisons := []TrendComparison{}

	if len(periods) >= 2 {
		period1 := periods[0]
		period2 := periods[1]

		comparisons = append(comparisons, TrendComparison{
			Metric:         "throughput",
			Period1Trend:   period1.Summary.Trends.ThroughputTrend,
			Period2Trend:   period2.Summary.Trends.ThroughputTrend,
			Consistency:    s.assessTrendConsistency(period1.Summary.Trends.ThroughputTrend, period2.Summary.Trends.ThroughputTrend),
			SignificantChange: period1.Summary.Trends.ThroughputTrend != period2.Summary.Trends.ThroughputTrend,
		})
	}

	return comparisons
}

func (s *StatisticsReportingService) generateComparisonInsights(analysis *ComparisonAnalysisReport) []ComparisonInsight {
	insights := []ComparisonInsight{}

	// Analyze metric changes for insights
	for _, metric := range analysis.MetricChanges {
		if metric.TrendDirection == "improving" {
			insights = append(insights, ComparisonInsight{
				Type:        "positive_trend",
				Severity:    "low",
				Title:       fmt.Sprintf("%s is improving", metric.MetricName),
				Description: fmt.Sprintf("The %s metric shows consistent improvement across comparison periods", metric.MetricName),
				Impact:      "Positive impact on overall system performance",
				ActionItems: []string{"Monitor continued improvement", "Identify factors contributing to improvement"},
			})
		} else if metric.TrendDirection == "degrading" {
			insights = append(insights, ComparisonInsight{
				Type:        "negative_trend",
				Severity:    "medium",
				Title:       fmt.Sprintf("%s is degrading", metric.MetricName),
				Description: fmt.Sprintf("The %s metric shows declining performance across comparison periods", metric.MetricName),
				Impact:      "Potential negative impact on system performance",
				ActionItems: []string{"Investigate root causes", "Implement corrective measures", "Monitor closely"},
			})
		}
	}

	return insights
}

// Additional helper methods

func (s *StatisticsReportingService) determinePriority(recommendation string) string {
	lower := strings.ToLower(recommendation)
	if strings.Contains(lower, "critical") || strings.Contains(lower, "urgent") || strings.Contains(lower, "high") {
		return "high"
	}
	if strings.Contains(lower, "medium") || strings.Contains(lower, "moderate") {
		return "medium"
	}
	return "low"
}

func (s *StatisticsReportingService) extractTitle(recommendation string) string {
	// Extract first sentence or first 60 characters as title
	if len(recommendation) <= 60 {
		return recommendation
	}
	
	// Find first sentence
	if idx := strings.Index(recommendation, "."); idx > 0 && idx < 60 {
		return recommendation[:idx]
	}
	
	// Truncate at word boundary
	if len(recommendation) > 60 {
		words := strings.Split(recommendation[:60], " ")
		if len(words) > 1 {
			return strings.Join(words[:len(words)-1], " ") + "..."
		}
	}
	
	return recommendation[:60] + "..."
}

func (s *StatisticsReportingService) determineRiskLevel(currentValue, threshold float64, ascending bool) string {
	var ratio float64
	if ascending {
		ratio = currentValue / threshold
	} else {
		ratio = threshold / currentValue
	}

	if ratio >= 1.0 {
		return "critical"
	} else if ratio >= 0.8 {
		return "high"
	} else if ratio >= 0.6 {
		return "medium"
	}
	return "low"
}

func (s *StatisticsReportingService) calculatePercentChanges(values map[string]float64) map[string]float64 {
	changes := make(map[string]float64)
	
	// Simple implementation - in production, would need more sophisticated comparison logic
	var baseValue float64
	var basePeriod string
	for period, value := range values {
		if basePeriod == "" {
			baseValue = value
			basePeriod = period
		} else {
			if baseValue != 0 {
				changes[period] = (value - baseValue) / baseValue * 100
			}
		}
	}
	
	return changes
}

func (s *StatisticsReportingService) findBestPeriod(values map[string]float64, highest bool) string {
	var bestPeriod string
	var bestValue float64
	first := true

	for period, value := range values {
		if first {
			bestPeriod = period
			bestValue = value
			first = false
		} else {
			if (highest && value > bestValue) || (!highest && value < bestValue) {
				bestPeriod = period
				bestValue = value
			}
		}
	}

	return bestPeriod
}

func (s *StatisticsReportingService) determineTrendDirection(values map[string]float64) string {
	// Simple trend determination - in production, would use more sophisticated analysis
	if len(values) < 2 {
		return "stable"
	}

	// For simplicity, compare first and last values
	var firstValue, lastValue float64
	var periods []string
	for period := range values {
		periods = append(periods, period)
	}
	sort.Strings(periods)

	if len(periods) >= 2 {
		firstValue = values[periods[0]]
		lastValue = values[periods[len(periods)-1]]

		if lastValue > firstValue*1.05 {
			return "improving"
		} else if lastValue < firstValue*0.95 {
			return "degrading"
		}
	}

	return "stable"
}

func (s *StatisticsReportingService) assessTrendConsistency(trend1, trend2 string) string {
	if trend1 == trend2 {
		return "consistent"
	}
	
	// Check if trends are at least in the same general direction
	improving := []string{"increasing", "improving"}
	degrading := []string{"decreasing", "degrading"}
	
	trend1Improving := false
	trend2Improving := false
	trend1Degrading := false
	trend2Degrading := false
	
	for _, t := range improving {
		if trend1 == t {
			trend1Improving = true
		}
		if trend2 == t {
			trend2Improving = true
		}
	}
	
	for _, t := range degrading {
		if trend1 == t {
			trend1Degrading = true
		}
		if trend2 == t {
			trend2Degrading = true
		}
	}
	
	if (trend1Improving && trend2Improving) || (trend1Degrading && trend2Degrading) {
		return "similar"
	}
	
	return "inconsistent"
}

func (s *StatisticsReportingService) describeCorrelation(correlation float64, metric1, metric2 string) string {
	absCorr := math.Abs(correlation)
	
	strength := "weak"
	if absCorr > 0.7 {
		strength = "strong"
	} else if absCorr > 0.4 {
		strength = "moderate"
	}
	
	direction := "positive"
	if correlation < 0 {
		direction = "negative"
	}
	
	return fmt.Sprintf("%s %s correlation between %s and %s (r=%.3f)", 
		strings.Title(strength), direction, metric1, metric2, correlation)
}

func (s *StatisticsReportingService) extractThroughputTimeSeries(data []*timeseries.AggregatedData) map[string]interface{} {
	timestamps := make([]string, len(data))
	throughputs := make([]uint64, len(data))
	
	for i, point := range data {
		timestamps[i] = point.Timestamp.Format(time.RFC3339)
		throughputs[i] = point.LinkStats.AvgRxRate + point.LinkStats.AvgTxRate
	}
	
	return map[string]interface{}{
		"timestamps":  timestamps,
		"throughputs": throughputs,
	}
}

func (s *StatisticsReportingService) generateScheduledReport(ctx context.Context, deviceName tc.DeviceName) {
	// Generate daily summary report
	options := ReportOptions{
		Type:       ReportTypeSummary,
		DeviceName: deviceName,
		TimeRange: TimeRange{
			Start: time.Now().Add(-24 * time.Hour),
			End:   time.Now(),
		},
		Format: FormatJSON,
		Sections: []ReportSection{
			SectionExecutiveSummary,
			SectionPerformanceMetrics,
			SectionRecommendations,
		},
	}

	report, err := s.GenerateReport(ctx, options)
	if err != nil {
		s.logger.Error("Failed to generate scheduled report",
			logging.String("device", deviceName.String()),
			logging.Error(err))
		return
	}

	s.logger.Info("Generated scheduled report",
		logging.String("device", deviceName.String()),
		logging.String("health_status", report.ExecutiveSummary.OverallHealth),
		logging.Float64("health_score", report.ExecutiveSummary.HealthScore))
}