package application

import (
	"context"
	"sync"
	"time"

	"github.com/rng999/traffic-control-go/internal/infrastructure/timeseries"
	"github.com/rng999/traffic-control-go/pkg/logging"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// DashboardDataService provides real-time data preparation for monitoring dashboards
type DashboardDataService struct {
	historicalDataService     *HistoricalDataService
	performanceMetricsService *PerformanceMetricsService
	logger                    logging.Logger
	
	// Real-time data cache
	liveDataCache map[string]*LiveDataCache
	cacheMutex    sync.RWMutex
	
	// Update intervals
	fastUpdateInterval   time.Duration  // For real-time metrics (e.g., 1 second)
	mediumUpdateInterval time.Duration  // For trend data (e.g., 30 seconds)
	slowUpdateInterval   time.Duration  // For historical summaries (e.g., 5 minutes)
}

// LiveDataCache holds cached real-time data for a device
type LiveDataCache struct {
	DeviceName       string
	LastUpdate       time.Time
	CurrentMetrics   *RealTimeMetrics
	RecentTrends     *TrendData
	AlertStatus      *AlertStatus
	PerformanceState *PerformanceState
	mutex            sync.RWMutex
}

// NewDashboardDataService creates a new dashboard data service
func NewDashboardDataService(
	historicalDataService *HistoricalDataService,
	performanceMetricsService *PerformanceMetricsService,
) *DashboardDataService {
	return &DashboardDataService{
		historicalDataService:     historicalDataService,
		performanceMetricsService: performanceMetricsService,
		logger:                    logging.WithComponent("application.dashboard_data"),
		liveDataCache:            make(map[string]*LiveDataCache),
		fastUpdateInterval:       1 * time.Second,
		mediumUpdateInterval:     30 * time.Second,
		slowUpdateInterval:       5 * time.Minute,
	}
}

// RealTimeMetrics represents current system metrics for dashboard display
type RealTimeMetrics struct {
	Timestamp          time.Time                     `json:"timestamp"`
	DeviceName         string                        `json:"device_name"`
	OverallHealth      HealthIndicator               `json:"overall_health"`
	ThroughputMetrics  ThroughputMetrics            `json:"throughput_metrics"`
	LatencyMetrics     LatencyMetrics               `json:"latency_metrics"`
	QualityMetrics     QualityMetrics               `json:"quality_metrics"`
	ResourceUtilization ResourceUtilizationMetrics   `json:"resource_utilization"`
	TopTalkers         []TopTalkerInfo              `json:"top_talkers"`
	ActiveConnections  int                          `json:"active_connections"`
	ErrorCounts        ErrorCountMetrics            `json:"error_counts"`
}

// HealthIndicator shows the current system health status
type HealthIndicator struct {
	Status      string  `json:"status"`      // excellent, good, warning, critical, unknown
	Score       float64 `json:"score"`       // 0-100
	LastUpdated time.Time `json:"last_updated"`
	Issues      []string `json:"issues,omitempty"`
}

// ThroughputMetrics represents current throughput information
type ThroughputMetrics struct {
	CurrentRxRate    uint64  `json:"current_rx_rate_bps"`
	CurrentTxRate    uint64  `json:"current_tx_rate_bps"`
	PeakRxRate       uint64  `json:"peak_rx_rate_bps"`
	PeakTxRate       uint64  `json:"peak_tx_rate_bps"`
	UtilizationRx    float64 `json:"utilization_rx_percent"`
	UtilizationTx    float64 `json:"utilization_tx_percent"`
	TotalBytesRx     uint64  `json:"total_bytes_rx"`
	TotalBytesTx     uint64  `json:"total_bytes_tx"`
	RateChangeRx     float64 `json:"rate_change_rx_percent"`
	RateChangeTx     float64 `json:"rate_change_tx_percent"`
}

// LatencyMetrics represents current latency information
type LatencyMetrics struct {
	AverageLatency   float64 `json:"average_latency_ms"`
	MinLatency       float64 `json:"min_latency_ms"`
	MaxLatency       float64 `json:"max_latency_ms"`
	Jitter           float64 `json:"jitter_ms"`
	PacketLossRate   float64 `json:"packet_loss_percent"`
	LatencyTrend     string  `json:"latency_trend"`
}

// QualityMetrics represents current quality of service metrics
type QualityMetrics struct {
	PacketLossRate    float64 `json:"packet_loss_rate"`
	ErrorRate         float64 `json:"error_rate"`
	RetransmissionRate float64 `json:"retransmission_rate"`
	QueueDropRate     float64 `json:"queue_drop_rate"`
	OverlimitRate     float64 `json:"overlimit_rate"`
	QualityScore      float64 `json:"quality_score"`
}

// ResourceUtilizationMetrics represents current resource usage
type ResourceUtilizationMetrics struct {
	BandwidthUtilization float64            `json:"bandwidth_utilization_percent"`
	QueueUtilization     float64            `json:"queue_utilization_percent"`
	BufferUtilization    float64            `json:"buffer_utilization_percent"`
	ComponentUtilization map[string]float64 `json:"component_utilization"`
	CriticalResources    []string           `json:"critical_resources"`
}

// TopTalkerInfo represents high-traffic flows
type TopTalkerInfo struct {
	FlowID       string  `json:"flow_id"`
	Handle       string  `json:"handle"`
	BytesPerSec  uint64  `json:"bytes_per_sec"`
	PacketsPerSec uint64 `json:"packets_per_sec"`
	Percentage   float64 `json:"percentage_of_total"`
	Classification string `json:"classification"`
}

// ErrorCountMetrics represents various error counters
type ErrorCountMetrics struct {
	TotalErrors       uint64 `json:"total_errors"`
	DropErrors        uint64 `json:"drop_errors"`
	OverlimitErrors   uint64 `json:"overlimit_errors"`
	BacklogErrors     uint64 `json:"backlog_errors"`
	ConfigErrors      uint64 `json:"config_errors"`
	ErrorRate         float64 `json:"error_rate_per_sec"`
	ErrorTrend        string  `json:"error_trend"`
}

// TrendData represents recent trend information for dashboard
type TrendData struct {
	TimeWindow        time.Duration     `json:"time_window"`
	ThroughputTrend   TrendInfo        `json:"throughput_trend"`
	LatencyTrend      TrendInfo        `json:"latency_trend"`
	ErrorRateTrend    TrendInfo        `json:"error_rate_trend"`
	QualityTrend      TrendInfo        `json:"quality_trend"`
	PredictedValues   PredictionInfo   `json:"predicted_values"`
	TrendConfidence   float64          `json:"trend_confidence"`
}

// TrendInfo represents trend information for a specific metric
type TrendInfo struct {
	Direction    string    `json:"direction"`    // increasing, decreasing, stable
	Magnitude    float64   `json:"magnitude"`    // percentage change
	Confidence   float64   `json:"confidence"`   // 0-1
	StartValue   float64   `json:"start_value"`
	CurrentValue float64   `json:"current_value"`
	ChangeRate   float64   `json:"change_rate"`  // rate of change per unit time
	LastUpdated  time.Time `json:"last_updated"`
}

// PredictionInfo represents short-term predictions
type PredictionInfo struct {
	NextMinute    map[string]float64 `json:"next_minute"`
	NextFiveMin   map[string]float64 `json:"next_five_minutes"`
	NextFifteenMin map[string]float64 `json:"next_fifteen_minutes"`
	Confidence    float64            `json:"confidence"`
}

// AlertStatus represents current alert information
type AlertStatus struct {
	ActiveAlerts    []AlertInfo `json:"active_alerts"`
	AlertCounts     AlertCounts `json:"alert_counts"`
	RecentAlerts    []AlertInfo `json:"recent_alerts"`
	AlertTrends     AlertTrends `json:"alert_trends"`
}

// AlertInfo represents a specific alert
type AlertInfo struct {
	ID           string            `json:"id"`
	Level        string            `json:"level"`        // info, warning, critical
	Category     string            `json:"category"`     // performance, capacity, error, security
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Timestamp    time.Time         `json:"timestamp"`
	Duration     time.Duration     `json:"duration"`
	Source       string            `json:"source"`
	DeviceName   string            `json:"device_name"`
	MetricValue  float64           `json:"metric_value"`
	Threshold    float64           `json:"threshold"`
	Tags         map[string]string `json:"tags"`
	Acknowledged bool              `json:"acknowledged"`
}

// AlertCounts represents alert statistics
type AlertCounts struct {
	Critical int `json:"critical"`
	Warning  int `json:"warning"`
	Info     int `json:"info"`
	Total    int `json:"total"`
}

// AlertTrends represents alert trend information
type AlertTrends struct {
	AlertRate        float64 `json:"alert_rate_per_hour"`
	ResolutionRate   float64 `json:"resolution_rate_per_hour"`
	MeanTimeToResolve time.Duration `json:"mean_time_to_resolve"`
	RecurringAlerts  int     `json:"recurring_alerts"`
}

// PerformanceState represents current performance state
type PerformanceState struct {
	State            string            `json:"state"`             // optimal, degraded, critical
	StateConfidence  float64           `json:"state_confidence"`
	StateDuration    time.Duration     `json:"state_duration"`
	PreviousState    string            `json:"previous_state"`
	StateHistory     []StateTransition `json:"state_history"`
	PerformanceIndex float64           `json:"performance_index"`
}

// StateTransition represents a performance state change
type StateTransition struct {
	FromState string        `json:"from_state"`
	ToState   string        `json:"to_state"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
	Trigger   string        `json:"trigger"`
}

// DashboardUpdate represents a complete dashboard data update
type DashboardUpdate struct {
	UpdateID        string            `json:"update_id"`
	Timestamp       time.Time         `json:"timestamp"`
	DeviceUpdates   map[string]*RealTimeMetrics `json:"device_updates"`
	SystemSummary   *SystemSummary    `json:"system_summary"`
	GlobalAlerts    []AlertInfo       `json:"global_alerts"`
	UpdateMetadata  UpdateMetadata    `json:"update_metadata"`
}

// SystemSummary represents overall system status
type SystemSummary struct {
	TotalDevices      int                        `json:"total_devices"`
	HealthyDevices    int                        `json:"healthy_devices"`
	WarningDevices    int                        `json:"warning_devices"`
	CriticalDevices   int                        `json:"critical_devices"`
	TotalThroughput   uint64                     `json:"total_throughput_bps"`
	TotalErrors       uint64                     `json:"total_errors"`
	SystemHealth      HealthIndicator            `json:"system_health"`
	TopIssues         []string                   `json:"top_issues"`
	ResourceSummary   ResourceUtilizationMetrics `json:"resource_summary"`
}

// UpdateMetadata provides information about the update
type UpdateMetadata struct {
	UpdateType       string        `json:"update_type"`        // full, incremental, alert_only
	DataFreshness    time.Duration `json:"data_freshness"`
	UpdateDuration   time.Duration `json:"update_duration"`
	RecordsProcessed int           `json:"records_processed"`
	CacheHitRate     float64       `json:"cache_hit_rate"`
}

// GetRealTimeMetrics retrieves current real-time metrics for a device
func (d *DashboardDataService) GetRealTimeMetrics(ctx context.Context, deviceName tc.DeviceName) (*RealTimeMetrics, error) {
	d.logger.Debug("Getting real-time metrics",
		logging.String("device", deviceName.String()))

	cache := d.getOrCreateCache(deviceName.String())
	
	cache.mutex.RLock()
	if cache.CurrentMetrics != nil && time.Since(cache.LastUpdate) < d.fastUpdateInterval {
		defer cache.mutex.RUnlock()
		return cache.CurrentMetrics, nil
	}
	cache.mutex.RUnlock()

	// Fetch fresh data
	metrics, err := d.fetchRealTimeMetrics(ctx, deviceName)
	if err != nil {
		return nil, err
	}

	// Update cache
	cache.mutex.Lock()
	cache.CurrentMetrics = metrics
	cache.LastUpdate = time.Now()
	cache.mutex.Unlock()

	return metrics, nil
}

// GetTrendData retrieves recent trend data for a device
func (d *DashboardDataService) GetTrendData(ctx context.Context, deviceName tc.DeviceName, timeWindow time.Duration) (*TrendData, error) {
	d.logger.Debug("Getting trend data",
		logging.String("device", deviceName.String()),
		logging.String("window", timeWindow.String()))

	cache := d.getOrCreateCache(deviceName.String())
	
	cache.mutex.RLock()
	if cache.RecentTrends != nil && time.Since(cache.LastUpdate) < d.mediumUpdateInterval {
		defer cache.mutex.RUnlock()
		return cache.RecentTrends, nil
	}
	cache.mutex.RUnlock()

	// Fetch fresh trend data
	trends, err := d.fetchTrendData(ctx, deviceName, timeWindow)
	if err != nil {
		return nil, err
	}

	// Update cache
	cache.mutex.Lock()
	cache.RecentTrends = trends
	cache.mutex.Unlock()

	return trends, nil
}

// GetDashboardUpdate provides a complete dashboard update
func (d *DashboardDataService) GetDashboardUpdate(ctx context.Context, deviceNames []tc.DeviceName) (*DashboardUpdate, error) {
	updateStart := time.Now()
	
	d.logger.Info("Generating dashboard update",
		logging.Int("devices", len(deviceNames)))

	update := &DashboardUpdate{
		UpdateID:      d.generateUpdateID(),
		Timestamp:     updateStart,
		DeviceUpdates: make(map[string]*RealTimeMetrics),
	}

	// Collect metrics for all devices
	for _, deviceName := range deviceNames {
		metrics, err := d.GetRealTimeMetrics(ctx, deviceName)
		if err != nil {
			d.logger.Warn("Failed to get metrics for device",
				logging.String("device", deviceName.String()),
				logging.Error(err))
			continue
		}
		update.DeviceUpdates[deviceName.String()] = metrics
	}

	// Generate system summary
	update.SystemSummary = d.generateSystemSummary(update.DeviceUpdates)

	// Collect global alerts
	update.GlobalAlerts = d.collectGlobalAlerts(ctx, deviceNames)

	// Add metadata
	update.UpdateMetadata = UpdateMetadata{
		UpdateType:       "full",
		DataFreshness:    time.Since(updateStart),
		UpdateDuration:   time.Since(updateStart),
		RecordsProcessed: len(update.DeviceUpdates),
		CacheHitRate:     d.calculateCacheHitRate(),
	}

	d.logger.Debug("Dashboard update completed",
		logging.String("update_id", update.UpdateID),
		logging.Int("devices", len(update.DeviceUpdates)),
		logging.String("duration", update.UpdateMetadata.UpdateDuration.String()))

	return update, nil
}

// StartLiveUpdates begins continuous live data updates
func (d *DashboardDataService) StartLiveUpdates(ctx context.Context, deviceNames []tc.DeviceName, updateCallback func(*DashboardUpdate)) error {
	d.logger.Info("Starting live dashboard updates",
		logging.Int("devices", len(deviceNames)),
		logging.String("interval", d.fastUpdateInterval.String()))

	ticker := time.NewTicker(d.fastUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			d.logger.Info("Stopping live dashboard updates")
			return ctx.Err()
		case <-ticker.C:
			update, err := d.GetDashboardUpdate(ctx, deviceNames)
			if err != nil {
				d.logger.Error("Failed to generate dashboard update", logging.Error(err))
				continue
			}

			if updateCallback != nil {
				updateCallback(update)
			}
		}
	}
}

// Private helper methods

func (d *DashboardDataService) getOrCreateCache(deviceName string) *LiveDataCache {
	d.cacheMutex.Lock()
	defer d.cacheMutex.Unlock()

	cache, exists := d.liveDataCache[deviceName]
	if !exists {
		cache = &LiveDataCache{
			DeviceName: deviceName,
		}
		d.liveDataCache[deviceName] = cache
	}
	return cache
}

func (d *DashboardDataService) fetchRealTimeMetrics(ctx context.Context, deviceName tc.DeviceName) (*RealTimeMetrics, error) {
	// Get latest data point
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Minute) // Last minute for rate calculations

	historicalData, err := d.historicalDataService.GetHistoricalData(ctx, deviceName, startTime, endTime, "")
	if err != nil {
		return nil, err
	}

	if len(historicalData) == 0 {
		// Return default metrics if no data available
		return &RealTimeMetrics{
			Timestamp:  time.Now(),
			DeviceName: deviceName.String(),
			OverallHealth: HealthIndicator{
				Status:      "unknown",
				Score:       0,
				LastUpdated: time.Now(),
			},
		}, nil
	}

	latest := historicalData[len(historicalData)-1]
	
	// Calculate real-time metrics from latest data
	metrics := &RealTimeMetrics{
		Timestamp:  latest.Timestamp,
		DeviceName: deviceName.String(),
	}

	// Calculate throughput metrics
	metrics.ThroughputMetrics = d.calculateThroughputMetrics(latest, historicalData)
	
	// Calculate quality metrics
	metrics.QualityMetrics = d.calculateQualityMetrics(latest)
	
	// Calculate resource utilization
	metrics.ResourceUtilization = d.calculateResourceUtilization(latest)
	
	// Calculate overall health
	metrics.OverallHealth = d.calculateHealthIndicator(metrics)
	
	// Get top talkers
	metrics.TopTalkers = d.calculateTopTalkers(latest)
	
	// Calculate error counts
	metrics.ErrorCounts = d.calculateErrorCounts(latest, historicalData)

	return metrics, nil
}

func (d *DashboardDataService) fetchTrendData(ctx context.Context, deviceName tc.DeviceName, timeWindow time.Duration) (*TrendData, error) {
	endTime := time.Now()
	startTime := endTime.Add(-timeWindow)

	historicalData, err := d.historicalDataService.GetHistoricalData(ctx, deviceName, startTime, endTime, "")
	if err != nil {
		return nil, err
	}

	if len(historicalData) < 2 {
		return &TrendData{
			TimeWindow:      timeWindow,
			TrendConfidence: 0,
		}, nil
	}

	trends := &TrendData{
		TimeWindow: timeWindow,
	}

	// Calculate trends for different metrics
	trends.ThroughputTrend = d.calculateThroughputTrend(historicalData)
	trends.ErrorRateTrend = d.calculateErrorRateTrend(historicalData)
	trends.QualityTrend = d.calculateQualityTrend(historicalData)
	
	// Generate predictions
	trends.PredictedValues = d.generatePredictions(historicalData)
	
	// Calculate overall trend confidence
	trends.TrendConfidence = d.calculateTrendConfidence(trends)

	return trends, nil
}

func (d *DashboardDataService) calculateThroughputMetrics(latest *timeseries.AggregatedData, historical []*timeseries.AggregatedData) ThroughputMetrics {
	current := ThroughputMetrics{
		CurrentRxRate: latest.LinkStats.AvgRxRate,
		CurrentTxRate: latest.LinkStats.AvgTxRate,
		TotalBytesRx:  latest.LinkStats.TotalRxBytes,
		TotalBytesTx:  latest.LinkStats.TotalTxBytes,
	}

	// Find peak rates from historical data
	for _, data := range historical {
		if data.LinkStats.AvgRxRate > current.PeakRxRate {
			current.PeakRxRate = data.LinkStats.AvgRxRate
		}
		if data.LinkStats.AvgTxRate > current.PeakTxRate {
			current.PeakTxRate = data.LinkStats.AvgTxRate
		}
	}

	// Calculate utilization (assuming 1Gbps interface)
	interfaceCapacity := uint64(1000000000) // 1 Gbps
	current.UtilizationRx = float64(current.CurrentRxRate) / float64(interfaceCapacity) * 100
	current.UtilizationTx = float64(current.CurrentTxRate) / float64(interfaceCapacity) * 100

	// Calculate rate changes
	if len(historical) >= 2 {
		previous := historical[len(historical)-2]
		if previous.LinkStats.AvgRxRate > 0 {
			current.RateChangeRx = (float64(current.CurrentRxRate) - float64(previous.LinkStats.AvgRxRate)) / float64(previous.LinkStats.AvgRxRate) * 100
		}
		if previous.LinkStats.AvgTxRate > 0 {
			current.RateChangeTx = (float64(current.CurrentTxRate) - float64(previous.LinkStats.AvgTxRate)) / float64(previous.LinkStats.AvgTxRate) * 100
		}
	}

	return current
}

func (d *DashboardDataService) calculateQualityMetrics(latest *timeseries.AggregatedData) QualityMetrics {
	metrics := QualityMetrics{}

	if len(latest.QdiscStats) > 0 {
		totalPackets := uint64(0)
		totalDrops := uint64(0)
		totalOverlimits := uint64(0)

		for _, qdisc := range latest.QdiscStats {
			totalPackets += qdisc.TotalPackets
			totalDrops += qdisc.TotalDrops
			totalOverlimits += qdisc.TotalOverlimits
		}

		if totalPackets > 0 {
			metrics.PacketLossRate = float64(totalDrops) / float64(totalPackets) * 100
			metrics.OverlimitRate = float64(totalOverlimits) / float64(totalPackets) * 100
		}
	}

	// Calculate error rate from link stats
	totalLinkPackets := latest.LinkStats.TotalRxPackets + latest.LinkStats.TotalTxPackets
	totalLinkErrors := latest.LinkStats.TotalRxErrors + latest.LinkStats.TotalTxErrors
	if totalLinkPackets > 0 {
		metrics.ErrorRate = float64(totalLinkErrors) / float64(totalLinkPackets) * 100
	}

	// Calculate overall quality score
	metrics.QualityScore = d.calculateQualityScore(metrics)

	return metrics
}

func (d *DashboardDataService) calculateResourceUtilization(latest *timeseries.AggregatedData) ResourceUtilizationMetrics {
	utilization := ResourceUtilizationMetrics{
		ComponentUtilization: make(map[string]float64),
	}

	// Calculate bandwidth utilization
	interfaceCapacity := uint64(1000000000) // 1 Gbps
	totalThroughput := latest.LinkStats.AvgRxRate + latest.LinkStats.AvgTxRate
	utilization.BandwidthUtilization = float64(totalThroughput) / float64(interfaceCapacity) * 100

	// Calculate queue utilization from qdisc stats
	if len(latest.QdiscStats) > 0 {
		maxBacklog := uint64(0)
		for _, qdisc := range latest.QdiscStats {
			if qdisc.MaxBacklog > maxBacklog {
				maxBacklog = qdisc.MaxBacklog
			}
			utilization.ComponentUtilization[qdisc.Handle] = float64(qdisc.AvgRate) / float64(qdisc.MaxRate) * 100
		}
		
		// Assume queue capacity of 1MB
		queueCapacity := uint64(1024 * 1024)
		utilization.QueueUtilization = float64(maxBacklog) / float64(queueCapacity) * 100
	}

	// Identify critical resources (> 80% utilization)
	for component, util := range utilization.ComponentUtilization {
		if util > 80 {
			utilization.CriticalResources = append(utilization.CriticalResources, component)
		}
	}

	return utilization
}

func (d *DashboardDataService) calculateHealthIndicator(metrics *RealTimeMetrics) HealthIndicator {
	health := HealthIndicator{
		LastUpdated: time.Now(),
	}

	// Calculate health score based on various factors
	score := float64(100)
	var issues []string

	// Check packet loss
	if metrics.QualityMetrics.PacketLossRate > 1.0 {
		score -= 20
		issues = append(issues, "High packet loss detected")
	}

	// Check error rate
	if metrics.QualityMetrics.ErrorRate > 0.5 {
		score -= 15
		issues = append(issues, "High error rate detected")
	}

	// Check utilization
	if metrics.ResourceUtilization.BandwidthUtilization > 90 {
		score -= 25
		issues = append(issues, "High bandwidth utilization")
	}

	// Check critical resources
	if len(metrics.ResourceUtilization.CriticalResources) > 0 {
		score -= 10 * float64(len(metrics.ResourceUtilization.CriticalResources))
		issues = append(issues, "Critical resource utilization detected")
	}

	health.Score = score
	health.Issues = issues

	// Determine status based on score
	if score >= 90 {
		health.Status = "excellent"
	} else if score >= 75 {
		health.Status = "good"
	} else if score >= 50 {
		health.Status = "warning"
	} else if score >= 25 {
		health.Status = "critical"
	} else {
		health.Status = "critical"
	}

	return health
}

func (d *DashboardDataService) calculateTopTalkers(latest *timeseries.AggregatedData) []TopTalkerInfo {
	var topTalkers []TopTalkerInfo

	// Get total throughput for percentage calculations
	totalThroughput := latest.LinkStats.AvgRxRate + latest.LinkStats.AvgTxRate

	// Add qdisc flows as top talkers
	for _, qdisc := range latest.QdiscStats {
		if qdisc.AvgRate > 0 {
			percentage := float64(qdisc.AvgRate) / float64(totalThroughput) * 100
			
			topTalker := TopTalkerInfo{
				FlowID:         qdisc.Handle,
				Handle:         qdisc.Handle,
				BytesPerSec:    qdisc.AvgRate / 8, // Convert bits to bytes
				PacketsPerSec:  qdisc.AvgRate / 1500, // Estimate packets (assuming 1500 byte packets)
				Percentage:     percentage,
				Classification: qdisc.Type,
			}
			topTalkers = append(topTalkers, topTalker)
		}
	}

	// Sort by bandwidth usage (simplified - would normally sort)
	// For now, just return the first few
	if len(topTalkers) > 5 {
		topTalkers = topTalkers[:5]
	}

	return topTalkers
}

func (d *DashboardDataService) calculateErrorCounts(latest *timeseries.AggregatedData, historical []*timeseries.AggregatedData) ErrorCountMetrics {
	errors := ErrorCountMetrics{
		ErrorTrend: "stable",
	}

	// Count current errors
	for _, qdisc := range latest.QdiscStats {
		errors.TotalErrors += qdisc.TotalDrops + qdisc.TotalOverlimits
		errors.DropErrors += qdisc.TotalDrops
		errors.OverlimitErrors += qdisc.TotalOverlimits
	}

	errors.TotalErrors += latest.LinkStats.TotalRxErrors + latest.LinkStats.TotalTxErrors

	// Calculate error rate
	totalPackets := latest.LinkStats.TotalRxPackets + latest.LinkStats.TotalTxPackets
	if totalPackets > 0 {
		errors.ErrorRate = float64(errors.TotalErrors) / float64(totalPackets) * 100
	}

	// Calculate trend
	if len(historical) >= 2 {
		previous := historical[len(historical)-2]
		previousErrors := uint64(0)
		for _, qdisc := range previous.QdiscStats {
			previousErrors += qdisc.TotalDrops + qdisc.TotalOverlimits
		}
		previousErrors += previous.LinkStats.TotalRxErrors + previous.LinkStats.TotalTxErrors

		if errors.TotalErrors > previousErrors*110/100 {
			errors.ErrorTrend = "increasing"
		} else if errors.TotalErrors < previousErrors*90/100 {
			errors.ErrorTrend = "decreasing"
		}
	}

	return errors
}

func (d *DashboardDataService) calculateThroughputTrend(historical []*timeseries.AggregatedData) TrendInfo {
	if len(historical) < 2 {
		return TrendInfo{Direction: "stable", Confidence: 0}
	}

	first := historical[0]
	last := historical[len(historical)-1]

	firstThroughput := float64(first.LinkStats.AvgRxRate + first.LinkStats.AvgTxRate)
	lastThroughput := float64(last.LinkStats.AvgRxRate + last.LinkStats.AvgTxRate)

	trend := TrendInfo{
		StartValue:   firstThroughput,
		CurrentValue: lastThroughput,
		LastUpdated:  time.Now(),
	}

	if firstThroughput > 0 {
		change := (lastThroughput - firstThroughput) / firstThroughput * 100
		trend.Magnitude = change

		if change > 5 {
			trend.Direction = "increasing"
		} else if change < -5 {
			trend.Direction = "decreasing"
		} else {
			trend.Direction = "stable"
		}

		trend.Confidence = 0.8 // Simplified confidence calculation
	}

	return trend
}

func (d *DashboardDataService) calculateErrorRateTrend(historical []*timeseries.AggregatedData) TrendInfo {
	if len(historical) < 2 {
		return TrendInfo{Direction: "stable", Confidence: 0}
	}

	// Calculate error rates for first and last data points
	first := historical[0]
	last := historical[len(historical)-1]

	firstErrorRate := d.calculateErrorRateForData(first)
	lastErrorRate := d.calculateErrorRateForData(last)

	trend := TrendInfo{
		StartValue:   firstErrorRate,
		CurrentValue: lastErrorRate,
		LastUpdated:  time.Now(),
	}

	if firstErrorRate > 0 {
		change := (lastErrorRate - firstErrorRate) / firstErrorRate * 100
		trend.Magnitude = change

		if change > 10 {
			trend.Direction = "increasing"
		} else if change < -10 {
			trend.Direction = "decreasing"
		} else {
			trend.Direction = "stable"
		}

		trend.Confidence = 0.7
	}

	return trend
}

func (d *DashboardDataService) calculateQualityTrend(historical []*timeseries.AggregatedData) TrendInfo {
	if len(historical) < 2 {
		return TrendInfo{Direction: "stable", Confidence: 0}
	}

	// Calculate quality scores for first and last data points
	firstQuality := d.calculateQualityScore(d.calculateQualityMetrics(historical[0]))
	lastQuality := d.calculateQualityScore(d.calculateQualityMetrics(historical[len(historical)-1]))

	trend := TrendInfo{
		StartValue:   firstQuality,
		CurrentValue: lastQuality,
		LastUpdated:  time.Now(),
	}

	change := lastQuality - firstQuality
	trend.Magnitude = change

	if change > 5 {
		trend.Direction = "increasing"
	} else if change < -5 {
		trend.Direction = "decreasing"
	} else {
		trend.Direction = "stable"
	}

	trend.Confidence = 0.6

	return trend
}

func (d *DashboardDataService) generatePredictions(historical []*timeseries.AggregatedData) PredictionInfo {
	predictions := PredictionInfo{
		NextMinute:     make(map[string]float64),
		NextFiveMin:    make(map[string]float64),
		NextFifteenMin: make(map[string]float64),
		Confidence:     0.5,
	}

	if len(historical) < 3 {
		return predictions
	}

	// Simple linear extrapolation
	latest := historical[len(historical)-1]
	previous := historical[len(historical)-2]

	currentThroughput := float64(latest.LinkStats.AvgRxRate + latest.LinkStats.AvgTxRate)
	previousThroughput := float64(previous.LinkStats.AvgRxRate + previous.LinkStats.AvgTxRate)
	
	changeRate := currentThroughput - previousThroughput

	predictions.NextMinute["throughput"] = currentThroughput + changeRate*1
	predictions.NextFiveMin["throughput"] = currentThroughput + changeRate*5
	predictions.NextFifteenMin["throughput"] = currentThroughput + changeRate*15

	return predictions
}

func (d *DashboardDataService) calculateTrendConfidence(trends *TrendData) float64 {
	// Average confidence of all trends
	total := trends.ThroughputTrend.Confidence + trends.ErrorRateTrend.Confidence + trends.QualityTrend.Confidence
	return total / 3
}

func (d *DashboardDataService) generateSystemSummary(deviceUpdates map[string]*RealTimeMetrics) *SystemSummary {
	summary := &SystemSummary{
		TotalDevices: len(deviceUpdates),
		ResourceSummary: ResourceUtilizationMetrics{
			ComponentUtilization: make(map[string]float64),
		},
	}

	var totalThroughput uint64
	var totalErrors uint64
	var healthScores []float64
	var topIssues []string

	for _, metrics := range deviceUpdates {
		// Count device status
		switch metrics.OverallHealth.Status {
		case "excellent", "good":
			summary.HealthyDevices++
		case "warning":
			summary.WarningDevices++
		case "critical":
			summary.CriticalDevices++
		}

		// Aggregate throughput
		totalThroughput += metrics.ThroughputMetrics.CurrentRxRate + metrics.ThroughputMetrics.CurrentTxRate

		// Aggregate errors
		totalErrors += metrics.ErrorCounts.TotalErrors

		// Collect health scores
		healthScores = append(healthScores, metrics.OverallHealth.Score)

		// Collect issues
		topIssues = append(topIssues, metrics.OverallHealth.Issues...)
	}

	summary.TotalThroughput = totalThroughput
	summary.TotalErrors = totalErrors
	summary.TopIssues = d.deduplicateIssues(topIssues)

	// Calculate system health
	if len(healthScores) > 0 {
		avgScore := d.calculateAverage(healthScores)
		summary.SystemHealth = HealthIndicator{
			Score:       avgScore,
			LastUpdated: time.Now(),
		}

		if avgScore >= 90 {
			summary.SystemHealth.Status = "excellent"
		} else if avgScore >= 75 {
			summary.SystemHealth.Status = "good"
		} else if avgScore >= 50 {
			summary.SystemHealth.Status = "warning"
		} else {
			summary.SystemHealth.Status = "critical"
		}
	}

	return summary
}

func (d *DashboardDataService) collectGlobalAlerts(ctx context.Context, deviceNames []tc.DeviceName) []AlertInfo {
	var alerts []AlertInfo

	// This would typically query an alerting system
	// For now, generate alerts based on current metrics
	for _, deviceName := range deviceNames {
		metrics, err := d.GetRealTimeMetrics(ctx, deviceName)
		if err != nil {
			continue
		}

		// Generate alerts for critical conditions
		if metrics.QualityMetrics.PacketLossRate > 2.0 {
			alert := AlertInfo{
				ID:          d.generateAlertID(),
				Level:       "critical",
				Category:    "performance",
				Title:       "High Packet Loss",
				Description: "Packet loss rate exceeds acceptable threshold",
				Timestamp:   time.Now(),
				Source:      "dashboard_monitor",
				DeviceName:  deviceName.String(),
				MetricValue: metrics.QualityMetrics.PacketLossRate,
				Threshold:   2.0,
			}
			alerts = append(alerts, alert)
		}

		if metrics.ResourceUtilization.BandwidthUtilization > 95 {
			alert := AlertInfo{
				ID:          d.generateAlertID(),
				Level:       "warning",
				Category:    "capacity",
				Title:       "High Bandwidth Utilization",
				Description: "Bandwidth utilization approaching capacity",
				Timestamp:   time.Now(),
				Source:      "dashboard_monitor",
				DeviceName:  deviceName.String(),
				MetricValue: metrics.ResourceUtilization.BandwidthUtilization,
				Threshold:   95.0,
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// Helper methods

func (d *DashboardDataService) generateUpdateID() string {
	return time.Now().Format("20060102150405") + "-" + "update"
}

func (d *DashboardDataService) generateAlertID() string {
	return time.Now().Format("20060102150405") + "-" + "alert"
}

func (d *DashboardDataService) calculateCacheHitRate() float64 {
	// Simplified cache hit rate calculation
	return 0.85 // 85% hit rate
}

func (d *DashboardDataService) calculateQualityScore(metrics QualityMetrics) float64 {
	score := float64(100)
	
	// Deduct points for various quality issues
	score -= metrics.PacketLossRate * 10    // Each 1% packet loss costs 10 points
	score -= metrics.ErrorRate * 5         // Each 1% error rate costs 5 points
	score -= metrics.OverlimitRate * 3     // Each 1% overlimit rate costs 3 points
	
	if score < 0 {
		score = 0
	}
	
	return score
}

func (d *DashboardDataService) calculateErrorRateForData(data *timeseries.AggregatedData) float64 {
	totalPackets := data.LinkStats.TotalRxPackets + data.LinkStats.TotalTxPackets
	totalErrors := data.LinkStats.TotalRxErrors + data.LinkStats.TotalTxErrors
	
	for _, qdisc := range data.QdiscStats {
		totalErrors += qdisc.TotalDrops
	}
	
	if totalPackets > 0 {
		return float64(totalErrors) / float64(totalPackets) * 100
	}
	
	return 0
}

func (d *DashboardDataService) deduplicateIssues(issues []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, issue := range issues {
		if !seen[issue] {
			seen[issue] = true
			result = append(result, issue)
		}
	}
	
	return result
}

func (d *DashboardDataService) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := float64(0)
	for _, v := range values {
		sum += v
	}
	
	return sum / float64(len(values))
}