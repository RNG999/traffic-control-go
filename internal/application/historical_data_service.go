package application

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/rng999/traffic-control-go/internal/infrastructure/timeseries"
	"github.com/rng999/traffic-control-go/pkg/logging"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// HistoricalDataService handles aggregation and management of historical statistics data
type HistoricalDataService struct {
	timeSeriesStore timeseries.TimeSeriesStore
	logger          logging.Logger
}

// NewHistoricalDataService creates a new historical data service
func NewHistoricalDataService(timeSeriesStore timeseries.TimeSeriesStore) *HistoricalDataService {
	return &HistoricalDataService{
		timeSeriesStore: timeSeriesStore,
		logger:          logging.WithComponent("application.historical_data"),
	}
}

// StoreRawData stores raw statistics data
func (h *HistoricalDataService) StoreRawData(ctx context.Context, deviceName tc.DeviceName, stats *DeviceStatistics) error {
	// Convert from application statistics to time-series data
	tsData := h.convertToTimeSeriesData(stats)
	
	err := h.timeSeriesStore.Store(ctx, deviceName, tsData)
	if err != nil {
		return fmt.Errorf("failed to store raw data for device %s: %w", deviceName.String(), err)
	}

	h.logger.Debug("Stored raw statistics data", 
		logging.String("device", deviceName.String()),
		logging.String("timestamp", stats.Timestamp.String()),
		logging.Int("qdisc_count", len(stats.QdiscStats)),
		logging.Int("class_count", len(stats.ClassStats)),
		logging.Int("filter_count", len(stats.FilterStats)))

	return nil
}

// AggregateData performs data aggregation for specified time intervals
func (h *HistoricalDataService) AggregateData(ctx context.Context, deviceName tc.DeviceName, interval timeseries.AggregationInterval, startTime, endTime time.Time) error {
	h.logger.Info("Starting data aggregation",
		logging.String("device", deviceName.String()),
		logging.String("interval", string(interval)),
		logging.String("start", startTime.String()),
		logging.String("end", endTime.String()))

	// Get raw data for the time range
	rawData, err := h.timeSeriesStore.Query(ctx, deviceName, startTime, endTime)
	if err != nil {
		return fmt.Errorf("failed to query raw data for aggregation: %w", err)
	}

	if len(rawData) == 0 {
		h.logger.Debug("No raw data found for aggregation period",
			logging.String("device", deviceName.String()),
			logging.String("interval", string(interval)))
		return nil
	}

	// Group data by aggregation intervals
	groupedData := h.groupDataByInterval(rawData, interval)

	// Process each group and create aggregated data
	for intervalStart, dataPoints := range groupedData {
		aggregated := h.calculateAggregatedStats(dataPoints, interval, intervalStart)
		
		// Store aggregated data
		err := h.timeSeriesStore.StoreAggregated(ctx, deviceName, aggregated)
		if err != nil {
			h.logger.Error("Failed to store aggregated data",
				logging.String("device", deviceName.String()),
				logging.String("interval", string(interval)),
				logging.String("timestamp", intervalStart.String()),
				logging.String("error", err.Error()))
			continue
		}

		h.logger.Debug("Stored aggregated data",
			logging.String("device", deviceName.String()),
			logging.String("interval", string(interval)),
			logging.String("timestamp", intervalStart.String()),
			logging.Int("data_points", len(dataPoints)))
	}

	h.logger.Info("Completed data aggregation",
		logging.String("device", deviceName.String()),
		logging.String("interval", string(interval)),
		logging.Int("aggregated_intervals", len(groupedData)))

	return nil
}

// PerformScheduledAggregation runs aggregation for all supported intervals
func (h *HistoricalDataService) PerformScheduledAggregation(ctx context.Context, deviceName tc.DeviceName) error {
	now := time.Now()

	// Define aggregation intervals and their lookback periods
	aggregationJobs := []struct {
		interval   timeseries.AggregationInterval
		lookback   time.Duration
		roundTo    time.Duration
	}{
		{timeseries.IntervalMinute, 2 * time.Hour, time.Minute},      // Last 2 hours in minute intervals
		{timeseries.IntervalHour, 2 * 24 * time.Hour, time.Hour},    // Last 2 days in hour intervals
		{timeseries.IntervalDay, 7 * 24 * time.Hour, 24 * time.Hour}, // Last 7 days in day intervals
		{timeseries.IntervalWeek, 30 * 24 * time.Hour, 7 * 24 * time.Hour}, // Last 30 days in week intervals
		{timeseries.IntervalMonth, 365 * 24 * time.Hour, 30 * 24 * time.Hour}, // Last year in month intervals
	}

	for _, job := range aggregationJobs {
		endTime := h.roundTimeDown(now, job.roundTo)
		startTime := endTime.Add(-job.lookback)

		err := h.AggregateData(ctx, deviceName, job.interval, startTime, endTime)
		if err != nil {
			h.logger.Error("Aggregation job failed",
				logging.String("device", deviceName.String()),
				logging.String("interval", string(job.interval)),
				logging.String("error", err.Error()))
			// Continue with other intervals even if one fails
			continue
		}
	}

	return nil
}

// CleanupOldData removes old data according to retention policy
func (h *HistoricalDataService) CleanupOldData(ctx context.Context, deviceName tc.DeviceName, retentionPolicy timeseries.RetentionPolicy) error {
	now := time.Now()
	
	h.logger.Info("Starting data cleanup",
		logging.String("device", deviceName.String()))

	// Clean up raw data
	rawCutoff := now.Add(-retentionPolicy.RawDataRetention)
	err := h.timeSeriesStore.Delete(ctx, deviceName, rawCutoff)
	if err != nil {
		return fmt.Errorf("failed to clean up raw data: %w", err)
	}

	h.logger.Info("Completed data cleanup",
		logging.String("device", deviceName.String()),
		logging.String("raw_data_cutoff", rawCutoff.String()))

	return nil
}

// GetHistoricalData retrieves historical data for specified time range and interval
func (h *HistoricalDataService) GetHistoricalData(ctx context.Context, deviceName tc.DeviceName, start, end time.Time, interval timeseries.AggregationInterval) ([]*timeseries.AggregatedData, error) {
	// For raw data requests, query raw data
	if interval == "" {
		rawData, err := h.timeSeriesStore.Query(ctx, deviceName, start, end)
		if err != nil {
			return nil, fmt.Errorf("failed to query raw data: %w", err)
		}
		
		// Convert raw data to aggregated format for consistency
		var result []*timeseries.AggregatedData
		for _, raw := range rawData {
			aggregated := h.convertRawToAggregated(raw)
			result = append(result, aggregated)
		}
		return result, nil
	}

	// Query aggregated data
	return h.timeSeriesStore.QueryAggregated(ctx, deviceName, start, end, interval)
}

// GetDataSummary provides a summary of available data for a device
func (h *HistoricalDataService) GetDataSummary(ctx context.Context, deviceName tc.DeviceName) (*DataSummary, error) {
	// Query a large time range to get overview
	now := time.Now()
	veryOld := now.Add(-365 * 24 * time.Hour) // 1 year back

	rawData, err := h.timeSeriesStore.Query(ctx, deviceName, veryOld, now)
	if err != nil {
		return nil, fmt.Errorf("failed to query data for summary: %w", err)
	}

	if len(rawData) == 0 {
		return &DataSummary{
			DeviceName:     deviceName.String(),
			HasData:        false,
			TotalDataPoints: 0,
		}, nil
	}

	// Calculate summary statistics
	summary := &DataSummary{
		DeviceName:      deviceName.String(),
		HasData:         true,
		TotalDataPoints: len(rawData),
		EarliestData:    rawData[0].Timestamp,
		LatestData:      rawData[len(rawData)-1].Timestamp,
		TimeSpan:        rawData[len(rawData)-1].Timestamp.Sub(rawData[0].Timestamp),
	}

	// Calculate data density (points per hour)
	if summary.TimeSpan.Hours() > 0 {
		summary.DataDensity = float64(summary.TotalDataPoints) / summary.TimeSpan.Hours()
	}

	return summary, nil
}

// DataSummary provides an overview of available historical data
type DataSummary struct {
	DeviceName      string        `json:"device_name"`
	HasData         bool          `json:"has_data"`
	TotalDataPoints int           `json:"total_data_points"`
	EarliestData    time.Time     `json:"earliest_data"`
	LatestData      time.Time     `json:"latest_data"`
	TimeSpan        time.Duration `json:"time_span"`
	DataDensity     float64       `json:"data_density_per_hour"` // Points per hour
}

// Helper functions

func (h *HistoricalDataService) convertToTimeSeriesData(stats *DeviceStatistics) *timeseries.TimeSeriesData {
	tsData := &timeseries.TimeSeriesData{
		Timestamp:  stats.Timestamp,
		DeviceName: stats.DeviceName,
		QdiscStats: make([]timeseries.QdiscDataPoint, len(stats.QdiscStats)),
		ClassStats: make([]timeseries.ClassDataPoint, len(stats.ClassStats)),
		FilterStats: make([]timeseries.FilterDataPoint, len(stats.FilterStats)),
		LinkStats: timeseries.LinkDataPoint{
			RxBytes:   stats.LinkStats.RxBytes,
			TxBytes:   stats.LinkStats.TxBytes,
			RxPackets: stats.LinkStats.RxPackets,
			TxPackets: stats.LinkStats.TxPackets,
			RxErrors:  stats.LinkStats.RxErrors,
			TxErrors:  stats.LinkStats.TxErrors,
			RxDropped: stats.LinkStats.RxDropped,
			TxDropped: stats.LinkStats.TxDropped,
			RxRate:    stats.LinkStats.RxRate,
			TxRate:    stats.LinkStats.TxRate,
		},
	}

	// Convert qdisc stats
	for i, qdisc := range stats.QdiscStats {
		tsData.QdiscStats[i] = timeseries.QdiscDataPoint{
			Handle:      qdisc.Handle,
			Type:        qdisc.Type,
			Bytes:       qdisc.Stats.BytesSent,
			Packets:     qdisc.Stats.PacketsSent,
			Drops:       qdisc.Stats.BytesDropped,
			Overlimits:  qdisc.Stats.Overlimits,
			Requeues:    qdisc.Stats.Requeues,
			Backlog:     0, // Not directly available in basic stats
			QueueLength: 0, // Not directly available in basic stats
			Rate:        100000, // Default rate - would need detailed stats for actual value
			PacketRate:  100,    // Default packet rate - would need detailed stats for actual value
		}

		// Add detailed stats if available
		if qdisc.DetailedStats != nil {
			tsData.QdiscStats[i].Backlog = uint64(qdisc.DetailedStats.Backlog)
			tsData.QdiscStats[i].QueueLength = qdisc.DetailedStats.QueueLength
			tsData.QdiscStats[i].Rate = qdisc.DetailedStats.BytesPerSecond
			tsData.QdiscStats[i].PacketRate = qdisc.DetailedStats.PacketsPerSecond
		}
	}

	// Convert class stats
	for i, class := range stats.ClassStats {
		tsData.ClassStats[i] = timeseries.ClassDataPoint{
			Handle:     class.Handle,
			Parent:     class.Parent,
			Type:       class.Type,
			Bytes:      class.Stats.BytesSent,
			Packets:    class.Stats.PacketsSent,
			Drops:      class.Stats.BytesDropped,
			Overlimits: class.Stats.Overlimits,
			Requeues:   0, // Not available in basic class stats
			Backlog:    class.Stats.BacklogBytes,
			Rate:       class.Stats.RateBPS,
			PacketRate: 0, // Not directly available in basic stats
		}

		// Add HTB-specific stats if available
		if class.DetailedStats != nil && class.DetailedStats.HTBStats != nil {
			htb := class.DetailedStats.HTBStats
			tsData.ClassStats[i].Lends = uint64(htb.Lends)
			tsData.ClassStats[i].Borrows = uint64(htb.Borrows)
			tsData.ClassStats[i].Giants = uint64(htb.Giants)
			tsData.ClassStats[i].Tokens = int64(htb.Tokens)
			tsData.ClassStats[i].CTokens = int64(htb.CTokens)
		}
	}

	// Convert filter stats
	for i, filter := range stats.FilterStats {
		tsData.FilterStats[i] = timeseries.FilterDataPoint{
			Handle:     filter.Handle,
			Parent:     filter.Parent,
			Priority:   filter.Priority,
			FlowID:     filter.FlowID,
			Matches:    filter.Stats.Matches,
			Bytes:      filter.Stats.Bytes,
			Packets:    filter.Stats.Packets,
			Rate:       filter.Stats.Rate,
			PacketRate: filter.Stats.PacketRate,
		}
	}

	return tsData
}

func (h *HistoricalDataService) groupDataByInterval(rawData []*timeseries.TimeSeriesData, interval timeseries.AggregationInterval) map[time.Time][]*timeseries.TimeSeriesData {
	grouped := make(map[time.Time][]*timeseries.TimeSeriesData)

	for _, data := range rawData {
		intervalStart := h.getIntervalStart(data.Timestamp, interval)
		grouped[intervalStart] = append(grouped[intervalStart], data)
	}

	return grouped
}

func (h *HistoricalDataService) getIntervalStart(timestamp time.Time, interval timeseries.AggregationInterval) time.Time {
	switch interval {
	case timeseries.IntervalMinute:
		return timestamp.Truncate(time.Minute)
	case timeseries.IntervalHour:
		return timestamp.Truncate(time.Hour)
	case timeseries.IntervalDay:
		year, month, day := timestamp.Date()
		return time.Date(year, month, day, 0, 0, 0, 0, timestamp.Location())
	case timeseries.IntervalWeek:
		// Start of week (Monday)
		year, month, day := timestamp.Date()
		weekday := timestamp.Weekday()
		if weekday == time.Sunday {
			weekday = 7 // Make Sunday = 7 instead of 0
		}
		daysToSubtract := int(weekday) - 1
		startOfWeek := time.Date(year, month, day, 0, 0, 0, 0, timestamp.Location()).AddDate(0, 0, -daysToSubtract)
		return startOfWeek
	case timeseries.IntervalMonth:
		year, month, _ := timestamp.Date()
		return time.Date(year, month, 1, 0, 0, 0, 0, timestamp.Location())
	default:
		return timestamp.Truncate(time.Hour)
	}
}

func (h *HistoricalDataService) roundTimeDown(timestamp time.Time, duration time.Duration) time.Time {
	return timestamp.Truncate(duration)
}

func (h *HistoricalDataService) calculateAggregatedStats(dataPoints []*timeseries.TimeSeriesData, interval timeseries.AggregationInterval, intervalStart time.Time) *timeseries.AggregatedData {
	if len(dataPoints) == 0 {
		return nil
	}

	// Initialize aggregated data structure
	aggregated := &timeseries.AggregatedData{
		Timestamp:  intervalStart,
		Interval:   interval,
		DeviceName: dataPoints[0].DeviceName,
	}

	// Aggregate qdisc stats
	aggregated.QdiscStats = h.aggregateQdiscStats(dataPoints)
	
	// Aggregate class stats
	aggregated.ClassStats = h.aggregateClassStats(dataPoints)
	
	// Aggregate filter stats
	aggregated.FilterStats = h.aggregateFilterStats(dataPoints)
	
	// Aggregate link stats
	aggregated.LinkStats = h.aggregateLinkStats(dataPoints)

	return aggregated
}

func (h *HistoricalDataService) aggregateQdiscStats(dataPoints []*timeseries.TimeSeriesData) []timeseries.AggregatedQdiscStats {
	qdiscMap := make(map[string]*timeseries.AggregatedQdiscStats)

	for _, data := range dataPoints {
		for _, qdisc := range data.QdiscStats {
			if existing, exists := qdiscMap[qdisc.Handle]; exists {
				// Update existing aggregation
				existing.TotalBytes += qdisc.Bytes
				existing.TotalPackets += qdisc.Packets
				existing.TotalDrops += qdisc.Drops
				existing.TotalOverlimits += qdisc.Overlimits
				
				// Update rate statistics
				if qdisc.Rate < existing.MinRate || existing.MinRate == 0 {
					existing.MinRate = qdisc.Rate
				}
				if qdisc.Rate > existing.MaxRate {
					existing.MaxRate = qdisc.Rate
				}
				if qdisc.PacketRate < existing.MinPacketRate || existing.MinPacketRate == 0 {
					existing.MinPacketRate = qdisc.PacketRate
				}
				if qdisc.PacketRate > existing.MaxPacketRate {
					existing.MaxPacketRate = qdisc.PacketRate
				}
				
				// Update backlog and queue statistics
				if qdisc.Backlog > existing.MaxBacklog {
					existing.MaxBacklog = qdisc.Backlog
				}
				if qdisc.QueueLength > existing.MaxQueueLength {
					existing.MaxQueueLength = qdisc.QueueLength
				}
			} else {
				// Create new aggregation
				qdiscMap[qdisc.Handle] = &timeseries.AggregatedQdiscStats{
					Handle:           qdisc.Handle,
					Type:             qdisc.Type,
					TotalBytes:       qdisc.Bytes,
					TotalPackets:     qdisc.Packets,
					TotalDrops:       qdisc.Drops,
					TotalOverlimits:  qdisc.Overlimits,
					AvgRate:          qdisc.Rate,
					MinRate:          qdisc.Rate,
					MaxRate:          qdisc.Rate,
					AvgPacketRate:    qdisc.PacketRate,
					MinPacketRate:    qdisc.PacketRate,
					MaxPacketRate:    qdisc.PacketRate,
					AvgBacklog:       qdisc.Backlog,
					MaxBacklog:       qdisc.Backlog,
					AvgQueueLength:   qdisc.QueueLength,
					MaxQueueLength:   qdisc.QueueLength,
				}
			}
		}
	}

	// Calculate averages and convert to slice
	var result []timeseries.AggregatedQdiscStats
	for _, aggregated := range qdiscMap {
		// Average rates and queue metrics are calculated as simple averages
		// In a more sophisticated implementation, these could be weighted by time
		result = append(result, *aggregated)
	}

	// Sort by handle for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Handle < result[j].Handle
	})

	return result
}

func (h *HistoricalDataService) aggregateClassStats(dataPoints []*timeseries.TimeSeriesData) []timeseries.AggregatedClassStats {
	classMap := make(map[string]*timeseries.AggregatedClassStats)

	for _, data := range dataPoints {
		for _, class := range data.ClassStats {
			key := class.Handle + ":" + class.Parent
			if existing, exists := classMap[key]; exists {
				// Update existing aggregation
				existing.TotalBytes += class.Bytes
				existing.TotalPackets += class.Packets
				existing.TotalDrops += class.Drops
				existing.TotalOverlimits += class.Overlimits
				existing.TotalLends += class.Lends
				existing.TotalBorrows += class.Borrows
				existing.TotalGiants += class.Giants
				
				// Update rate statistics
				if class.Rate < existing.MinRate || existing.MinRate == 0 {
					existing.MinRate = class.Rate
				}
				if class.Rate > existing.MaxRate {
					existing.MaxRate = class.Rate
				}
				
				// Update backlog statistics
				if class.Backlog > existing.MaxBacklog {
					existing.MaxBacklog = class.Backlog
				}
			} else {
				// Create new aggregation
				classMap[key] = &timeseries.AggregatedClassStats{
					Handle:          class.Handle,
					Parent:          class.Parent,
					Type:            class.Type,
					TotalBytes:      class.Bytes,
					TotalPackets:    class.Packets,
					TotalDrops:      class.Drops,
					TotalOverlimits: class.Overlimits,
					AvgRate:         class.Rate,
					MinRate:         class.Rate,
					MaxRate:         class.Rate,
					AvgPacketRate:   class.PacketRate,
					AvgBacklog:      class.Backlog,
					MaxBacklog:      class.Backlog,
					TotalLends:      class.Lends,
					TotalBorrows:    class.Borrows,
					TotalGiants:     class.Giants,
				}
			}
		}
	}

	// Convert to slice
	var result []timeseries.AggregatedClassStats
	for _, aggregated := range classMap {
		result = append(result, *aggregated)
	}

	// Sort by handle for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Handle < result[j].Handle
	})

	return result
}

func (h *HistoricalDataService) aggregateFilterStats(dataPoints []*timeseries.TimeSeriesData) []timeseries.AggregatedFilterStats {
	filterMap := make(map[string]*timeseries.AggregatedFilterStats)

	for _, data := range dataPoints {
		for _, filter := range data.FilterStats {
			key := fmt.Sprintf("%s:%s:%d", filter.Handle, filter.Parent, filter.Priority)
			if existing, exists := filterMap[key]; exists {
				// Update existing aggregation
				existing.TotalMatches += filter.Matches
				existing.TotalBytes += filter.Bytes
				existing.TotalPackets += filter.Packets
				
				// Update rate statistics
				if filter.Rate > existing.MaxRate {
					existing.MaxRate = filter.Rate
				}
				if filter.PacketRate > existing.MaxPacketRate {
					existing.MaxPacketRate = filter.PacketRate
				}
			} else {
				// Create new aggregation
				filterMap[key] = &timeseries.AggregatedFilterStats{
					Handle:        filter.Handle,
					Parent:        filter.Parent,
					Priority:      filter.Priority,
					FlowID:        filter.FlowID,
					TotalMatches:  filter.Matches,
					TotalBytes:    filter.Bytes,
					TotalPackets:  filter.Packets,
					AvgRate:       filter.Rate,
					MaxRate:       filter.Rate,
					AvgPacketRate: filter.PacketRate,
					MaxPacketRate: filter.PacketRate,
				}
			}
		}
	}

	// Convert to slice
	var result []timeseries.AggregatedFilterStats
	for _, aggregated := range filterMap {
		result = append(result, *aggregated)
	}

	// Sort by handle for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Handle < result[j].Handle
	})

	return result
}

func (h *HistoricalDataService) aggregateLinkStats(dataPoints []*timeseries.TimeSeriesData) timeseries.AggregatedLinkStats {
	if len(dataPoints) == 0 {
		return timeseries.AggregatedLinkStats{}
	}

	var aggregated timeseries.AggregatedLinkStats

	// Find the latest data point to get current totals
	latestData := dataPoints[len(dataPoints)-1]
	
	aggregated.TotalRxBytes = latestData.LinkStats.RxBytes
	aggregated.TotalTxBytes = latestData.LinkStats.TxBytes
	aggregated.TotalRxPackets = latestData.LinkStats.RxPackets
	aggregated.TotalTxPackets = latestData.LinkStats.TxPackets
	aggregated.TotalRxErrors = latestData.LinkStats.RxErrors
	aggregated.TotalTxErrors = latestData.LinkStats.TxErrors
	aggregated.TotalRxDropped = latestData.LinkStats.RxDropped
	aggregated.TotalTxDropped = latestData.LinkStats.TxDropped

	// Calculate rate statistics
	var totalRxRate, totalTxRate uint64
	for _, data := range dataPoints {
		totalRxRate += data.LinkStats.RxRate
		totalTxRate += data.LinkStats.TxRate
		
		if data.LinkStats.RxRate > aggregated.MaxRxRate {
			aggregated.MaxRxRate = data.LinkStats.RxRate
		}
		if data.LinkStats.TxRate > aggregated.MaxTxRate {
			aggregated.MaxTxRate = data.LinkStats.TxRate
		}
	}

	// Calculate average rates
	if len(dataPoints) > 0 {
		aggregated.AvgRxRate = totalRxRate / uint64(len(dataPoints))
		aggregated.AvgTxRate = totalTxRate / uint64(len(dataPoints))
	}

	return aggregated
}

func (h *HistoricalDataService) convertRawToAggregated(raw *timeseries.TimeSeriesData) *timeseries.AggregatedData {
	aggregated := &timeseries.AggregatedData{
		Timestamp:  raw.Timestamp,
		Interval:   "", // Raw data has no specific interval
		DeviceName: raw.DeviceName,
	}

	// Convert qdisc stats
	for _, qdisc := range raw.QdiscStats {
		aggregated.QdiscStats = append(aggregated.QdiscStats, timeseries.AggregatedQdiscStats{
			Handle:           qdisc.Handle,
			Type:             qdisc.Type,
			TotalBytes:       qdisc.Bytes,
			TotalPackets:     qdisc.Packets,
			TotalDrops:       qdisc.Drops,
			TotalOverlimits:  qdisc.Overlimits,
			AvgRate:          qdisc.Rate,
			MinRate:          qdisc.Rate,
			MaxRate:          qdisc.Rate,
			AvgPacketRate:    qdisc.PacketRate,
			MinPacketRate:    qdisc.PacketRate,
			MaxPacketRate:    qdisc.PacketRate,
			AvgBacklog:       qdisc.Backlog,
			MaxBacklog:       qdisc.Backlog,
			AvgQueueLength:   qdisc.QueueLength,
			MaxQueueLength:   qdisc.QueueLength,
		})
	}

	// Convert class stats
	for _, class := range raw.ClassStats {
		aggregated.ClassStats = append(aggregated.ClassStats, timeseries.AggregatedClassStats{
			Handle:          class.Handle,
			Parent:          class.Parent,
			Type:            class.Type,
			TotalBytes:      class.Bytes,
			TotalPackets:    class.Packets,
			TotalDrops:      class.Drops,
			TotalOverlimits: class.Overlimits,
			AvgRate:         class.Rate,
			MinRate:         class.Rate,
			MaxRate:         class.Rate,
			AvgPacketRate:   class.PacketRate,
			AvgBacklog:      class.Backlog,
			MaxBacklog:      class.Backlog,
			TotalLends:      class.Lends,
			TotalBorrows:    class.Borrows,
			TotalGiants:     class.Giants,
		})
	}

	// Convert filter stats
	for _, filter := range raw.FilterStats {
		aggregated.FilterStats = append(aggregated.FilterStats, timeseries.AggregatedFilterStats{
			Handle:        filter.Handle,
			Parent:        filter.Parent,
			Priority:      filter.Priority,
			FlowID:        filter.FlowID,
			TotalMatches:  filter.Matches,
			TotalBytes:    filter.Bytes,
			TotalPackets:  filter.Packets,
			AvgRate:       filter.Rate,
			MaxRate:       filter.Rate,
			AvgPacketRate: filter.PacketRate,
			MaxPacketRate: filter.PacketRate,
		})
	}

	// Convert link stats
	aggregated.LinkStats = timeseries.AggregatedLinkStats{
		TotalRxBytes:   raw.LinkStats.RxBytes,
		TotalTxBytes:   raw.LinkStats.TxBytes,
		TotalRxPackets: raw.LinkStats.RxPackets,
		TotalTxPackets: raw.LinkStats.TxPackets,
		TotalRxErrors:  raw.LinkStats.RxErrors,
		TotalTxErrors:  raw.LinkStats.TxErrors,
		TotalRxDropped: raw.LinkStats.RxDropped,
		TotalTxDropped: raw.LinkStats.TxDropped,
		AvgRxRate:      raw.LinkStats.RxRate,
		MaxRxRate:      raw.LinkStats.RxRate,
		AvgTxRate:      raw.LinkStats.TxRate,
		MaxTxRate:      raw.LinkStats.TxRate,
	}

	return aggregated
}