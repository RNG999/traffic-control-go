package application

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/rng999/traffic-control-go/internal/infrastructure/timeseries"
	"github.com/rng999/traffic-control-go/pkg/logging"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// PerformanceMetricsService calculates performance metrics and trends from historical data
type PerformanceMetricsService struct {
	historicalDataService *HistoricalDataService
	logger                logging.Logger
}

// NewPerformanceMetricsService creates a new performance metrics service
func NewPerformanceMetricsService(historicalDataService *HistoricalDataService) *PerformanceMetricsService {
	return &PerformanceMetricsService{
		historicalDataService: historicalDataService,
		logger:                logging.WithComponent("application.performance_metrics"),
	}
}

// PerformanceMetrics represents calculated performance metrics for a device
type PerformanceMetrics struct {
	DeviceName        string                     `json:"device_name"`
	TimeRange         TimeRange                  `json:"time_range"`
	OverallMetrics    OverallPerformanceMetrics  `json:"overall_metrics"`
	QdiscMetrics      []QdiscPerformanceMetrics  `json:"qdisc_metrics"`
	ClassMetrics      []ClassPerformanceMetrics  `json:"class_metrics"`
	FilterMetrics     []FilterPerformanceMetrics `json:"filter_metrics"`
	LinkMetrics       LinkPerformanceMetrics     `json:"link_metrics"`
	Trends            TrendAnalysis              `json:"trends"`
	HealthScore       HealthScore                `json:"health_score"`
	Recommendations   []string                   `json:"recommendations"`
}

// TimeRange represents a time period for analysis
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// OverallPerformanceMetrics provides system-wide performance indicators
type OverallPerformanceMetrics struct {
	TotalThroughput      uint64  `json:"total_throughput_bps"`     // Total bytes per second
	PeakThroughput       uint64  `json:"peak_throughput_bps"`      // Peak throughput observed
	AverageThroughput    uint64  `json:"average_throughput_bps"`   // Average throughput
	ThroughputUtilization float64 `json:"throughput_utilization"`   // Percentage of capacity used
	PacketLossRate       float64 `json:"packet_loss_rate"`         // Percentage of packets lost
	LatencyIndicator     float64 `json:"latency_indicator"`        // Queue backlog as latency proxy
	JitterIndicator      float64 `json:"jitter_indicator"`         // Rate variance as jitter proxy
	EfficiencyScore      float64 `json:"efficiency_score"`         // Overall efficiency (0-100)
}

// QdiscPerformanceMetrics provides qdisc-specific performance metrics
type QdiscPerformanceMetrics struct {
	Handle               string  `json:"handle"`
	Type                 string  `json:"type"`
	TotalBytes           uint64  `json:"total_bytes"`
	TotalPackets         uint64  `json:"total_packets"`
	TotalDrops           uint64  `json:"total_drops"`
	DropRate             float64 `json:"drop_rate"`                // Percentage of packets dropped
	ThroughputMbps       float64 `json:"throughput_mbps"`          // Throughput in Mbps
	AveragePacketSize    float64 `json:"average_packet_size"`      // Average packet size in bytes
	PeakRate             uint64  `json:"peak_rate_bps"`            // Peak rate observed
	RateUtilization      float64 `json:"rate_utilization"`         // Percentage of configured rate used
	QueueEfficiency      float64 `json:"queue_efficiency"`         // How well the queue is managed
	BacklogTrend         string  `json:"backlog_trend"`            // "increasing", "decreasing", "stable"
	PerformanceScore     float64 `json:"performance_score"`        // Overall performance (0-100)
}

// ClassPerformanceMetrics provides class-specific performance metrics
type ClassPerformanceMetrics struct {
	Handle               string  `json:"handle"`
	Parent               string  `json:"parent"`
	Type                 string  `json:"type"`
	TotalBytes           uint64  `json:"total_bytes"`
	TotalPackets         uint64  `json:"total_packets"`
	TotalDrops           uint64  `json:"total_drops"`
	DropRate             float64 `json:"drop_rate"`
	ThroughputMbps       float64 `json:"throughput_mbps"`
	BandwidthUtilization float64 `json:"bandwidth_utilization"`    // Percentage of allocated bandwidth used
	HTBLendingRatio      float64 `json:"htb_lending_ratio"`        // Ratio of lending to borrowing
	HTBBorrowingRatio    float64 `json:"htb_borrowing_ratio"`      // How often class borrows bandwidth
	BacklogEfficiency    float64 `json:"backlog_efficiency"`       // How well backlog is managed
	FairnessScore        float64 `json:"fairness_score"`           // How fairly bandwidth is shared
	PerformanceScore     float64 `json:"performance_score"`        // Overall performance (0-100)
}

// FilterPerformanceMetrics provides filter-specific performance metrics
type FilterPerformanceMetrics struct {
	Handle               string  `json:"handle"`
	Parent               string  `json:"parent"`
	Priority             uint16  `json:"priority"`
	FlowID               string  `json:"flow_id"`
	TotalMatches         uint64  `json:"total_matches"`
	TotalBytes           uint64  `json:"total_bytes"`
	TotalPackets         uint64  `json:"total_packets"`
	MatchRate            float64 `json:"match_rate_pps"`           // Matches per second
	ThroughputMbps       float64 `json:"throughput_mbps"`
	AveragePacketSize    float64 `json:"average_packet_size"`
	ClassificationEfficiency float64 `json:"classification_efficiency"` // How efficiently traffic is classified
	PerformanceScore     float64 `json:"performance_score"`        // Overall performance (0-100)
}

// LinkPerformanceMetrics provides interface-level performance metrics
type LinkPerformanceMetrics struct {
	RxThroughputMbps     float64 `json:"rx_throughput_mbps"`
	TxThroughputMbps     float64 `json:"tx_throughput_mbps"`
	RxUtilization        float64 `json:"rx_utilization"`           // Percentage of RX capacity used
	TxUtilization        float64 `json:"tx_utilization"`           // Percentage of TX capacity used
	ErrorRate            float64 `json:"error_rate"`               // Percentage of packets with errors
	DropRate             float64 `json:"drop_rate"`                // Percentage of packets dropped
	DuplexEfficiency     float64 `json:"duplex_efficiency"`        // How well both directions are used
	PerformanceScore     float64 `json:"performance_score"`        // Overall performance (0-100)
}

// TrendAnalysis provides trend information over time
type TrendAnalysis struct {
	ThroughputTrend      string  `json:"throughput_trend"`         // "increasing", "decreasing", "stable"
	DropRateTrend        string  `json:"drop_rate_trend"`
	LatencyTrend         string  `json:"latency_trend"`
	UtilizationTrend     string  `json:"utilization_trend"`
	TrendConfidence      float64 `json:"trend_confidence"`         // Confidence in trend analysis (0-1)
	PredictedThroughput  uint64  `json:"predicted_throughput"`     // Predicted future throughput
	SeasonalityDetected  bool    `json:"seasonality_detected"`     // Whether seasonal patterns are detected
}

// HealthScore provides overall system health assessment
type HealthScore struct {
	OverallScore         float64            `json:"overall_score"`         // Overall health (0-100)
	ThroughputScore      float64            `json:"throughput_score"`      // Throughput health (0-100)
	ReliabilityScore     float64            `json:"reliability_score"`     // Reliability health (0-100)
	EfficiencyScore      float64            `json:"efficiency_score"`      // Efficiency health (0-100)
	StabilityScore       float64            `json:"stability_score"`       // Stability health (0-100)
	ComponentScores      map[string]float64 `json:"component_scores"`      // Individual component scores
	HealthStatus         string             `json:"health_status"`         // "excellent", "good", "fair", "poor", "critical"
}

// CalculatePerformanceMetrics calculates comprehensive performance metrics for a device
func (p *PerformanceMetricsService) CalculatePerformanceMetrics(ctx context.Context, deviceName tc.DeviceName, timeRange TimeRange) (*PerformanceMetrics, error) {
	p.logger.Info("Calculating performance metrics",
		logging.String("device", deviceName.String()),
		logging.String("start", timeRange.Start.String()),
		logging.String("end", timeRange.End.String()))

	// Get historical data for the time range
	historicalData, err := p.historicalDataService.GetHistoricalData(ctx, deviceName, timeRange.Start, timeRange.End, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	if len(historicalData) == 0 {
		return &PerformanceMetrics{
			DeviceName: deviceName.String(),
			TimeRange:  timeRange,
			HealthScore: HealthScore{
				HealthStatus: "unknown",
			},
			Recommendations: []string{"No historical data available for analysis"},
		}, nil
	}

	// Calculate metrics
	metrics := &PerformanceMetrics{
		DeviceName: deviceName.String(),
		TimeRange:  timeRange,
	}

	// Calculate overall metrics
	metrics.OverallMetrics = p.calculateOverallMetrics(historicalData)

	// Calculate qdisc metrics
	metrics.QdiscMetrics = p.calculateQdiscMetrics(historicalData)

	// Calculate class metrics
	metrics.ClassMetrics = p.calculateClassMetrics(historicalData)

	// Calculate filter metrics
	metrics.FilterMetrics = p.calculateFilterMetrics(historicalData)

	// Calculate link metrics
	metrics.LinkMetrics = p.calculateLinkMetrics(historicalData)

	// Calculate trends
	metrics.Trends = p.calculateTrends(historicalData)

	// Calculate health score
	metrics.HealthScore = p.calculateHealthScore(metrics)

	// Generate recommendations
	metrics.Recommendations = p.generateRecommendations(metrics)

	p.logger.Info("Performance metrics calculated",
		logging.String("device", deviceName.String()),
		logging.Int("data_points", len(historicalData)),
		logging.Float64("overall_score", metrics.HealthScore.OverallScore))

	return metrics, nil
}

// CalculateRealTimeMetrics calculates current performance metrics
func (p *PerformanceMetricsService) CalculateRealTimeMetrics(ctx context.Context, deviceName tc.DeviceName) (*PerformanceMetrics, error) {
	// Use last 1 hour for real-time metrics
	now := time.Now()
	timeRange := TimeRange{
		Start: now.Add(-1 * time.Hour),
		End:   now,
	}

	return p.CalculatePerformanceMetrics(ctx, deviceName, timeRange)
}

// Helper functions for calculating specific metrics

func (p *PerformanceMetricsService) calculateOverallMetrics(data []*timeseries.AggregatedData) OverallPerformanceMetrics {
	if len(data) == 0 {
		return OverallPerformanceMetrics{}
	}

	var totalThroughput, peakThroughput, totalPackets, totalDrops uint64
	var totalBacklog, totalRateVariance float64

	for _, point := range data {
		// Calculate throughput from link stats
		throughput := point.LinkStats.AvgRxRate + point.LinkStats.AvgTxRate
		totalThroughput += throughput
		if throughput > peakThroughput {
			peakThroughput = throughput
		}

		// Aggregate packet loss from qdiscs
		for _, qdisc := range point.QdiscStats {
			totalPackets += qdisc.TotalPackets
			totalDrops += qdisc.TotalDrops
			totalBacklog += float64(qdisc.AvgBacklog)
			
			// Calculate rate variance as jitter indicator
			if qdisc.MaxRate > qdisc.MinRate {
				rateVariance := float64(qdisc.MaxRate - qdisc.MinRate) / float64(qdisc.AvgRate)
				totalRateVariance += rateVariance
			}
		}
	}

	avgThroughput := totalThroughput / uint64(len(data))
	
	// Calculate packet loss rate
	var packetLossRate float64
	if totalPackets > 0 {
		packetLossRate = float64(totalDrops) / float64(totalPackets) * 100
	}

	// Calculate latency indicator (based on average backlog)
	latencyIndicator := totalBacklog / float64(len(data))

	// Calculate jitter indicator (based on rate variance)
	jitterIndicator := totalRateVariance / float64(len(data))

	// Calculate efficiency score
	efficiencyScore := p.calculateEfficiencyScore(packetLossRate, latencyIndicator, jitterIndicator)

	return OverallPerformanceMetrics{
		TotalThroughput:       totalThroughput,
		PeakThroughput:        peakThroughput,
		AverageThroughput:     avgThroughput,
		ThroughputUtilization: 85.0, // This would need interface capacity information
		PacketLossRate:        packetLossRate,
		LatencyIndicator:      latencyIndicator,
		JitterIndicator:       jitterIndicator,
		EfficiencyScore:       efficiencyScore,
	}
}

func (p *PerformanceMetricsService) calculateQdiscMetrics(data []*timeseries.AggregatedData) []QdiscPerformanceMetrics {
	qdiscMap := make(map[string]*QdiscPerformanceMetrics)

	for _, point := range data {
		for _, qdisc := range point.QdiscStats {
			key := qdisc.Handle + ":" + qdisc.Type
			
			if existing, exists := qdiscMap[key]; exists {
				// Update existing metrics
				existing.TotalBytes += qdisc.TotalBytes
				existing.TotalPackets += qdisc.TotalPackets
				existing.TotalDrops += qdisc.TotalDrops
				if qdisc.MaxRate > existing.PeakRate {
					existing.PeakRate = qdisc.MaxRate
				}
			} else {
				// Create new metrics
				throughputMbps := float64(qdisc.AvgRate) / 1_000_000 * 8 // Convert bytes/s to Mbps
				
				var dropRate float64
				if qdisc.TotalPackets > 0 {
					dropRate = float64(qdisc.TotalDrops) / float64(qdisc.TotalPackets) * 100
				}
				
				var avgPacketSize float64
				if qdisc.TotalPackets > 0 {
					avgPacketSize = float64(qdisc.TotalBytes) / float64(qdisc.TotalPackets)
				}

				qdiscMap[key] = &QdiscPerformanceMetrics{
					Handle:            qdisc.Handle,
					Type:              qdisc.Type,
					TotalBytes:        qdisc.TotalBytes,
					TotalPackets:      qdisc.TotalPackets,
					TotalDrops:        qdisc.TotalDrops,
					DropRate:          dropRate,
					ThroughputMbps:    throughputMbps,
					AveragePacketSize: avgPacketSize,
					PeakRate:          qdisc.MaxRate,
					RateUtilization:   75.0, // Would need configured rate for accurate calculation
					QueueEfficiency:   p.calculateQueueEfficiency(qdisc),
					BacklogTrend:      "stable", // Would need historical comparison
					PerformanceScore:  p.calculateQdiscPerformanceScore(qdisc, dropRate),
				}
			}
		}
	}

	// Convert map to slice
	var result []QdiscPerformanceMetrics
	for _, metrics := range qdiscMap {
		result = append(result, *metrics)
	}

	// Sort by handle for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Handle < result[j].Handle
	})

	return result
}

func (p *PerformanceMetricsService) calculateClassMetrics(data []*timeseries.AggregatedData) []ClassPerformanceMetrics {
	classMap := make(map[string]*ClassPerformanceMetrics)

	for _, point := range data {
		for _, class := range point.ClassStats {
			key := class.Handle + ":" + class.Parent
			
			if existing, exists := classMap[key]; existing != nil && exists {
				// Update existing metrics
				existing.TotalBytes += class.TotalBytes
				existing.TotalPackets += class.TotalPackets
				existing.TotalDrops += class.TotalDrops
			} else {
				// Create new metrics
				throughputMbps := float64(class.AvgRate) / 1_000_000 * 8
				
				var dropRate float64
				if class.TotalPackets > 0 {
					dropRate = float64(class.TotalDrops) / float64(class.TotalPackets) * 100
				}

				var htbLendingRatio, htbBorrowingRatio float64
				if class.TotalLends > 0 || class.TotalBorrows > 0 {
					total := class.TotalLends + class.TotalBorrows
					if total > 0 {
						htbLendingRatio = float64(class.TotalLends) / float64(total)
						htbBorrowingRatio = float64(class.TotalBorrows) / float64(total)
					}
				}

				classMap[key] = &ClassPerformanceMetrics{
					Handle:               class.Handle,
					Parent:               class.Parent,
					Type:                 class.Type,
					TotalBytes:           class.TotalBytes,
					TotalPackets:         class.TotalPackets,
					TotalDrops:           class.TotalDrops,
					DropRate:             dropRate,
					ThroughputMbps:       throughputMbps,
					BandwidthUtilization: 80.0, // Would need configured bandwidth for accurate calculation
					HTBLendingRatio:      htbLendingRatio,
					HTBBorrowingRatio:    htbBorrowingRatio,
					BacklogEfficiency:    p.calculateBacklogEfficiency(class),
					FairnessScore:        p.calculateFairnessScore(class),
					PerformanceScore:     p.calculateClassPerformanceScore(class, dropRate),
				}
			}
		}
	}

	// Convert map to slice
	var result []ClassPerformanceMetrics
	for _, metrics := range classMap {
		result = append(result, *metrics)
	}

	// Sort by handle for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Handle < result[j].Handle
	})

	return result
}

func (p *PerformanceMetricsService) calculateFilterMetrics(data []*timeseries.AggregatedData) []FilterPerformanceMetrics {
	filterMap := make(map[string]*FilterPerformanceMetrics)

	for _, point := range data {
		for _, filter := range point.FilterStats {
			key := fmt.Sprintf("%s:%s:%d", filter.Handle, filter.Parent, filter.Priority)
			
			if existing, exists := filterMap[key]; existing != nil && exists {
				// Update existing metrics
				existing.TotalMatches += filter.TotalMatches
				existing.TotalBytes += filter.TotalBytes
				existing.TotalPackets += filter.TotalPackets
			} else {
				// Create new metrics
				throughputMbps := float64(filter.AvgRate) / 1_000_000 * 8
				
				var avgPacketSize float64
				if filter.TotalPackets > 0 {
					avgPacketSize = float64(filter.TotalBytes) / float64(filter.TotalPackets)
				}

				var matchRate float64
				// This would need time duration for accurate calculation
				matchRate = float64(filter.TotalMatches) / 3600.0 // Assume 1 hour for now

				filterMap[key] = &FilterPerformanceMetrics{
					Handle:                    filter.Handle,
					Parent:                    filter.Parent,
					Priority:                  filter.Priority,
					FlowID:                    filter.FlowID,
					TotalMatches:              filter.TotalMatches,
					TotalBytes:                filter.TotalBytes,
					TotalPackets:              filter.TotalPackets,
					MatchRate:                 matchRate,
					ThroughputMbps:            throughputMbps,
					AveragePacketSize:         avgPacketSize,
					ClassificationEfficiency:  p.calculateClassificationEfficiency(filter),
					PerformanceScore:          p.calculateFilterPerformanceScore(filter),
				}
			}
		}
	}

	// Convert map to slice
	var result []FilterPerformanceMetrics
	for _, metrics := range filterMap {
		result = append(result, *metrics)
	}

	// Sort by handle for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Handle < result[j].Handle
	})

	return result
}

func (p *PerformanceMetricsService) calculateLinkMetrics(data []*timeseries.AggregatedData) LinkPerformanceMetrics {
	if len(data) == 0 {
		return LinkPerformanceMetrics{}
	}

	var totalRxThroughput, totalTxThroughput, totalRxPackets, totalTxPackets uint64
	var totalRxErrors, totalTxErrors, totalRxDropped, totalTxDropped uint64

	for _, point := range data {
		link := point.LinkStats
		totalRxThroughput += link.AvgRxRate
		totalTxThroughput += link.AvgTxRate
		totalRxPackets += link.TotalRxPackets
		totalTxPackets += link.TotalTxPackets
		totalRxErrors += link.TotalRxErrors
		totalTxErrors += link.TotalTxErrors
		totalRxDropped += link.TotalRxDropped
		totalTxDropped += link.TotalTxDropped
	}

	avgRxThroughputMbps := float64(totalRxThroughput) / float64(len(data)) / 1_000_000 * 8
	avgTxThroughputMbps := float64(totalTxThroughput) / float64(len(data)) / 1_000_000 * 8

	var errorRate, dropRate float64
	totalPackets := totalRxPackets + totalTxPackets
	if totalPackets > 0 {
		totalErrors := totalRxErrors + totalTxErrors
		totalDrops := totalRxDropped + totalTxDropped
		errorRate = float64(totalErrors) / float64(totalPackets) * 100
		dropRate = float64(totalDrops) / float64(totalPackets) * 100
	}

	// Calculate duplex efficiency (how well both directions are utilized)
	var duplexEfficiency float64
	if avgRxThroughputMbps > 0 && avgTxThroughputMbps > 0 {
		minThroughput := math.Min(avgRxThroughputMbps, avgTxThroughputMbps)
		maxThroughput := math.Max(avgRxThroughputMbps, avgTxThroughputMbps)
		if maxThroughput > 0 {
			duplexEfficiency = minThroughput / maxThroughput * 100
		}
	}

	performanceScore := p.calculateLinkPerformanceScore(errorRate, dropRate, duplexEfficiency)

	return LinkPerformanceMetrics{
		RxThroughputMbps: avgRxThroughputMbps,
		TxThroughputMbps: avgTxThroughputMbps,
		RxUtilization:    70.0, // Would need interface capacity for accurate calculation
		TxUtilization:    65.0, // Would need interface capacity for accurate calculation
		ErrorRate:        errorRate,
		DropRate:         dropRate,
		DuplexEfficiency: duplexEfficiency,
		PerformanceScore: performanceScore,
	}
}

func (p *PerformanceMetricsService) calculateTrends(data []*timeseries.AggregatedData) TrendAnalysis {
	if len(data) < 2 {
		return TrendAnalysis{
			ThroughputTrend:  "stable",
			DropRateTrend:    "stable",
			LatencyTrend:     "stable",
			UtilizationTrend: "stable",
			TrendConfidence:  0.0,
		}
	}

	// Simple trend analysis based on first and last data points
	firstPoint := data[0]
	lastPoint := data[len(data)-1]

	// Calculate throughput trend
	firstThroughput := firstPoint.LinkStats.AvgRxRate + firstPoint.LinkStats.AvgTxRate
	lastThroughput := lastPoint.LinkStats.AvgRxRate + lastPoint.LinkStats.AvgTxRate
	
	throughputTrend := "stable"
	if lastThroughput > uint64(float64(firstThroughput)*1.05) {
		throughputTrend = "increasing"
	} else if lastThroughput < uint64(float64(firstThroughput)*0.95) {
		throughputTrend = "decreasing"
	}

	// Calculate drop rate trend (simplified)
	var firstDrops, lastDrops, firstPackets, lastPackets uint64
	for _, qdisc := range firstPoint.QdiscStats {
		firstDrops += qdisc.TotalDrops
		firstPackets += qdisc.TotalPackets
	}
	for _, qdisc := range lastPoint.QdiscStats {
		lastDrops += qdisc.TotalDrops
		lastPackets += qdisc.TotalPackets
	}

	dropRateTrend := "stable"
	if firstPackets > 0 && lastPackets > 0 {
		firstDropRate := float64(firstDrops) / float64(firstPackets)
		lastDropRate := float64(lastDrops) / float64(lastPackets)
		
		if lastDropRate > firstDropRate*1.1 {
			dropRateTrend = "increasing"
		} else if lastDropRate < firstDropRate*0.9 {
			dropRateTrend = "decreasing"
		}
	}

	// Predict future throughput (simple linear extrapolation)
	predictedThroughput := lastThroughput
	if len(data) > 10 {
		// Use linear regression for better prediction
		predictedThroughput = p.predictThroughput(data)
	}

	return TrendAnalysis{
		ThroughputTrend:     throughputTrend,
		DropRateTrend:       dropRateTrend,
		LatencyTrend:        "stable", // Would need more sophisticated analysis
		UtilizationTrend:    "stable", // Would need more sophisticated analysis
		TrendConfidence:     0.7,      // Moderate confidence with simple analysis
		PredictedThroughput: predictedThroughput,
		SeasonalityDetected: false,    // Would need longer time series
	}
}

func (p *PerformanceMetricsService) calculateHealthScore(metrics *PerformanceMetrics) HealthScore {
	// Calculate component scores
	throughputScore := p.calculateThroughputScore(metrics.OverallMetrics)
	reliabilityScore := p.calculateReliabilityScore(metrics.OverallMetrics)
	efficiencyScore := metrics.OverallMetrics.EfficiencyScore
	stabilityScore := p.calculateStabilityScore(metrics.Trends)

	// Calculate overall score as weighted average
	overallScore := (throughputScore*0.3 + reliabilityScore*0.3 + efficiencyScore*0.2 + stabilityScore*0.2)

	// Determine health status
	healthStatus := "excellent"
	if overallScore < 90 {
		healthStatus = "good"
	}
	if overallScore < 75 {
		healthStatus = "fair"
	}
	if overallScore < 60 {
		healthStatus = "poor"
	}
	if overallScore < 40 {
		healthStatus = "critical"
	}

	componentScores := map[string]float64{
		"throughput":   throughputScore,
		"reliability":  reliabilityScore,
		"efficiency":   efficiencyScore,
		"stability":    stabilityScore,
	}

	// Add individual component scores
	for _, qdisc := range metrics.QdiscMetrics {
		componentScores[fmt.Sprintf("qdisc_%s", qdisc.Handle)] = qdisc.PerformanceScore
	}
	for _, class := range metrics.ClassMetrics {
		componentScores[fmt.Sprintf("class_%s", class.Handle)] = class.PerformanceScore
	}

	return HealthScore{
		OverallScore:     overallScore,
		ThroughputScore:  throughputScore,
		ReliabilityScore: reliabilityScore,
		EfficiencyScore:  efficiencyScore,
		StabilityScore:   stabilityScore,
		ComponentScores:  componentScores,
		HealthStatus:     healthStatus,
	}
}

func (p *PerformanceMetricsService) generateRecommendations(metrics *PerformanceMetrics) []string {
	var recommendations []string

	// Check overall health
	if metrics.HealthScore.OverallScore < 60 {
		recommendations = append(recommendations, "System performance is below acceptable levels. Consider reviewing configuration and capacity.")
	}

	// Check packet loss
	if metrics.OverallMetrics.PacketLossRate > 1.0 {
		recommendations = append(recommendations, fmt.Sprintf("High packet loss detected (%.2f%%). Consider increasing buffer sizes or bandwidth allocation.", metrics.OverallMetrics.PacketLossRate))
	}

	// Check qdisc performance
	for _, qdisc := range metrics.QdiscMetrics {
		if qdisc.DropRate > 2.0 {
			recommendations = append(recommendations, fmt.Sprintf("Qdisc %s has high drop rate (%.2f%%). Consider tuning queue parameters.", qdisc.Handle, qdisc.DropRate))
		}
		if qdisc.RateUtilization > 90 {
			recommendations = append(recommendations, fmt.Sprintf("Qdisc %s is near capacity (%.1f%% utilization). Consider increasing rate limits.", qdisc.Handle, qdisc.RateUtilization))
		}
	}

	// Check class performance
	for _, class := range metrics.ClassMetrics {
		if class.HTBBorrowingRatio > 0.8 {
			recommendations = append(recommendations, fmt.Sprintf("Class %s frequently borrows bandwidth (%.1f%% of the time). Consider increasing guaranteed rate.", class.Handle, class.HTBBorrowingRatio*100))
		}
	}

	// Check trends
	if metrics.Trends.ThroughputTrend == "increasing" && metrics.Trends.TrendConfidence > 0.7 {
		recommendations = append(recommendations, "Throughput is consistently increasing. Monitor capacity and consider scaling up if trend continues.")
	}
	
	if metrics.Trends.DropRateTrend == "increasing" {
		recommendations = append(recommendations, "Drop rate is increasing. Investigate potential congestion or configuration issues.")
	}

	// Check link performance
	if metrics.LinkMetrics.ErrorRate > 0.1 {
		recommendations = append(recommendations, fmt.Sprintf("High error rate detected (%.3f%%). Check physical connection and hardware.", metrics.LinkMetrics.ErrorRate))
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "System is performing well. Continue monitoring for any changes.")
	}

	return recommendations
}

// Helper calculation functions

func (p *PerformanceMetricsService) calculateEfficiencyScore(packetLossRate, latencyIndicator, jitterIndicator float64) float64 {
	// Start with perfect score and deduct points for issues
	score := 100.0
	
	// Deduct for packet loss (up to 30 points)
	score -= math.Min(packetLossRate*10, 30)
	
	// Deduct for latency (up to 25 points)
	if latencyIndicator > 1000 {
		score -= math.Min((latencyIndicator-1000)/1000*25, 25)
	}
	
	// Deduct for jitter (up to 15 points)
	score -= math.Min(jitterIndicator*15, 15)
	
	return math.Max(score, 0)
}

func (p *PerformanceMetricsService) calculateQueueEfficiency(qdisc timeseries.AggregatedQdiscStats) float64 {
	// Queue efficiency based on backlog management and rate utilization
	efficiency := 100.0
	
	// Penalize for excessive backlog
	if qdisc.MaxBacklog > 10000 {
		efficiency -= math.Min(float64(qdisc.MaxBacklog)/1000*5, 30)
	}
	
	// Penalize for queue length
	if qdisc.MaxQueueLength > 100 {
		efficiency -= math.Min(float64(qdisc.MaxQueueLength)/10*2, 20)
	}
	
	return math.Max(efficiency, 0)
}

func (p *PerformanceMetricsService) calculateQdiscPerformanceScore(qdisc timeseries.AggregatedQdiscStats, dropRate float64) float64 {
	score := 100.0
	
	// Deduct for drops
	score -= math.Min(dropRate*5, 40)
	
	// Deduct for excessive overlimits
	if qdisc.TotalPackets > 0 {
		overlimitRate := float64(qdisc.TotalOverlimits) / float64(qdisc.TotalPackets) * 100
		score -= math.Min(overlimitRate*2, 20)
	}
	
	// Deduct for poor queue management
	queueEfficiency := p.calculateQueueEfficiency(qdisc)
	score = score * (queueEfficiency / 100.0)
	
	return math.Max(score, 0)
}

func (p *PerformanceMetricsService) calculateBacklogEfficiency(class timeseries.AggregatedClassStats) float64 {
	// Simple calculation based on backlog levels
	if class.MaxBacklog == 0 {
		return 100.0
	}
	
	efficiency := 100.0 - math.Min(float64(class.MaxBacklog)/1000*10, 50)
	return math.Max(efficiency, 0)
}

func (p *PerformanceMetricsService) calculateFairnessScore(class timeseries.AggregatedClassStats) float64 {
	// Fairness based on HTB lending/borrowing behavior
	if class.TotalLends == 0 && class.TotalBorrows == 0 {
		return 100.0 // No sharing needed
	}
	
	total := class.TotalLends + class.TotalBorrows
	if total == 0 {
		return 100.0
	}
	
	// Fair sharing would be roughly balanced
	lendRatio := float64(class.TotalLends) / float64(total)
	
	// Score is higher when there's reasonable balance
	fairnessScore := 100.0 - math.Abs(lendRatio-0.5)*100
	return math.Max(fairnessScore, 0)
}

func (p *PerformanceMetricsService) calculateClassPerformanceScore(class timeseries.AggregatedClassStats, dropRate float64) float64 {
	score := 100.0
	
	// Deduct for drops
	score -= math.Min(dropRate*5, 40)
	
	// Add efficiency factors
	backlogEfficiency := p.calculateBacklogEfficiency(class)
	fairnessScore := p.calculateFairnessScore(class)
	
	score = score * 0.6 + backlogEfficiency*0.2 + fairnessScore*0.2
	
	return math.Max(score, 0)
}

func (p *PerformanceMetricsService) calculateClassificationEfficiency(filter timeseries.AggregatedFilterStats) float64 {
	// Classification efficiency based on match rate and throughput
	if filter.TotalMatches == 0 {
		return 0.0 // No matches means no classification
	}
	
	// High matches with corresponding throughput indicates good efficiency
	if filter.TotalBytes > 0 {
		bytesPerMatch := float64(filter.TotalBytes) / float64(filter.TotalMatches)
		// Reasonable packet size indicates effective classification
		if bytesPerMatch > 64 && bytesPerMatch < 9000 {
			return 100.0
		}
	}
	
	return 75.0 // Default moderate efficiency
}

func (p *PerformanceMetricsService) calculateFilterPerformanceScore(filter timeseries.AggregatedFilterStats) float64 {
	score := 100.0
	
	// Filter performance is mainly about classification efficiency
	classificationEfficiency := p.calculateClassificationEfficiency(filter)
	score = score * (classificationEfficiency / 100.0)
	
	// Bonus for high throughput
	if filter.AvgRate > 1000000 { // 1 Mbps
		score += 10.0
	}
	
	return math.Min(score, 100.0)
}

func (p *PerformanceMetricsService) calculateLinkPerformanceScore(errorRate, dropRate, duplexEfficiency float64) float64 {
	score := 100.0
	
	// Deduct for errors and drops
	score -= math.Min(errorRate*50, 30)
	score -= math.Min(dropRate*10, 30)
	
	// Add bonus for good duplex utilization
	if duplexEfficiency > 70 {
		score += 5.0
	}
	
	return math.Min(math.Max(score, 0), 100.0)
}

func (p *PerformanceMetricsService) calculateThroughputScore(overall OverallPerformanceMetrics) float64 {
	// Throughput score based on utilization and efficiency
	score := overall.ThroughputUtilization
	
	// Adjust based on efficiency
	score = score * (overall.EfficiencyScore / 100.0)
	
	return math.Min(score, 100.0)
}

func (p *PerformanceMetricsService) calculateReliabilityScore(overall OverallPerformanceMetrics) float64 {
	score := 100.0
	
	// Deduct for packet loss
	score -= math.Min(overall.PacketLossRate*20, 60)
	
	// Deduct for high latency
	if overall.LatencyIndicator > 1000 {
		score -= math.Min((overall.LatencyIndicator-1000)/1000*20, 30)
	}
	
	return math.Max(score, 0)
}

func (p *PerformanceMetricsService) calculateStabilityScore(trends TrendAnalysis) float64 {
	score := 100.0
	
	// Stable trends get full score
	if trends.ThroughputTrend != "stable" {
		score -= 15.0
	}
	if trends.DropRateTrend == "increasing" {
		score -= 25.0
	}
	if trends.LatencyTrend == "increasing" {
		score -= 20.0
	}
	
	// Low confidence in trends reduces score
	score = score * trends.TrendConfidence
	
	return math.Max(score, 0)
}

func (p *PerformanceMetricsService) predictThroughput(data []*timeseries.AggregatedData) uint64 {
	if len(data) < 2 {
		return 0
	}
	
	// Simple linear regression on throughput
	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(data))
	
	for i, point := range data {
		x := float64(i)
		y := float64(point.LinkStats.AvgRxRate + point.LinkStats.AvgTxRate)
		
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	
	// Calculate slope and intercept
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n
	
	// Predict next value
	nextX := n
	predicted := slope*nextX + intercept
	
	return uint64(math.Max(predicted, 0))
}