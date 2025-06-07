package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	"github.com/rng999/traffic-control-go/internal/projections"
	"github.com/rng999/traffic-control-go/internal/queries/models"
	"github.com/rng999/traffic-control-go/pkg/logging"
	"github.com/rng999/traffic-control-go/pkg/tc"
	"github.com/rng999/traffic-control-go/pkg/types"
)

// StatisticsQueryService provides TC statistics collection functionality for queries
type StatisticsQueryService struct {
	netlinkAdapter netlink.Adapter
	readModelStore projections.ReadModelStore
	logger         logging.Logger
}

// NewStatisticsQueryService creates a new statistics query service
func NewStatisticsQueryService(netlinkAdapter netlink.Adapter, readModelStore projections.ReadModelStore) *StatisticsQueryService {
	return &StatisticsQueryService{
		netlinkAdapter: netlinkAdapter,
		readModelStore: readModelStore,
		logger:         logging.WithComponent("queries.statistics"),
	}
}

// DeviceStatistics represents statistics for a device (simplified for queries)
type DeviceStatistics struct {
	DeviceName  string             `json:"device_name"`
	Timestamp   time.Time          `json:"timestamp"`
	QdiscStats  []QdiscStatistics  `json:"qdisc_stats"`
	ClassStats  []ClassStatistics  `json:"class_stats"`
	FilterStats []FilterStatistics `json:"filter_stats"`
	LinkStats   LinkStatistics     `json:"link_stats"`
}

// QdiscStatistics represents qdisc statistics with metadata
type QdiscStatistics struct {
	Handle        string                      `json:"handle"`
	Type          string                      `json:"type"`
	Stats         netlink.QdiscStats          `json:"stats"`
	DetailedStats *netlink.DetailedQdiscStats `json:"detailed_stats,omitempty"`
}

// ClassStatistics represents class statistics with metadata
type ClassStatistics struct {
	Handle        string                      `json:"handle"`
	Parent        string                      `json:"parent"`
	Name          string                      `json:"name"`
	Stats         netlink.ClassStats          `json:"stats"`
	DetailedStats *netlink.DetailedClassStats `json:"detailed_stats,omitempty"`
}

// FilterStatistics represents filter statistics with metadata
type FilterStatistics struct {
	Parent     string `json:"parent"`
	Priority   uint16 `json:"priority"`
	Protocol   string `json:"protocol"`
	MatchCount int    `json:"match_count"`
}

// LinkStatistics represents network interface statistics
type LinkStatistics struct {
	RxBytes   uint64 `json:"rx_bytes"`
	TxBytes   uint64 `json:"tx_bytes"`
	RxPackets uint64 `json:"rx_packets"`
	TxPackets uint64 `json:"tx_packets"`
	RxErrors  uint64 `json:"rx_errors"`
	TxErrors  uint64 `json:"tx_errors"`
	RxDropped uint64 `json:"rx_dropped"`
	TxDropped uint64 `json:"tx_dropped"`
}

// GetDeviceStatistics retrieves complete statistics for a device
func (s *StatisticsQueryService) GetDeviceStatistics(ctx context.Context, deviceName string) (*DeviceStatistics, error) {
	s.logger.Info("Getting device statistics",
		logging.String("device", deviceName))

	// Get configuration from read model
	var readModel projections.TrafficControlReadModel
	modelID := fmt.Sprintf("tc:%s", deviceName)

	if err := s.readModelStore.Get(ctx, "traffic-control", modelID, &readModel); err != nil {
		s.logger.Warn("No configuration found for device",
			logging.String("device", deviceName),
			logging.Error(err))
		// Continue anyway - we can still get raw statistics
	}

	device, err := tc.NewDevice(deviceName)
	if err != nil {
		return nil, fmt.Errorf("invalid device name: %w", err)
	}

	stats := &DeviceStatistics{
		DeviceName:  deviceName,
		Timestamp:   time.Now(),
		QdiscStats:  make([]QdiscStatistics, 0),
		ClassStats:  make([]ClassStatistics, 0),
		FilterStats: make([]FilterStatistics, 0),
	}

	// Get qdisc statistics
	for _, qdisc := range readModel.Qdiscs {
		_, err := parseHandle(qdisc.Handle)
		if err != nil {
			s.logger.Warn("Invalid qdisc handle",
				logging.String("handle", qdisc.Handle),
				logging.Error(err))
			continue
		}

		// Get basic stats from netlink
		qdiscInfo := s.netlinkAdapter.GetQdiscs(device)
		if qdiscInfo.IsSuccess() {
			for _, info := range qdiscInfo.Value() {
				if info.Handle.String() == qdisc.Handle {
					qdiscStat := QdiscStatistics{
						Handle: qdisc.Handle,
						Type:   qdisc.Type,
						Stats:  info.Statistics,
					}

					// TODO: Add detailed statistics through adapter interface
					// For now, basic statistics from netlink are sufficient

					stats.QdiscStats = append(stats.QdiscStats, qdiscStat)
					break
				}
			}
		}
	}

	// Get class statistics
	for _, class := range readModel.Classes {
		_, err := parseHandle(class.Handle)
		if err != nil {
			s.logger.Warn("Invalid class handle",
				logging.String("handle", class.Handle),
				logging.Error(err))
			continue
		}

		// Get basic stats from netlink
		classInfo := s.netlinkAdapter.GetClasses(device)
		if classInfo.IsSuccess() {
			for _, info := range classInfo.Value() {
				if info.Handle.String() == class.Handle {
					classStat := ClassStatistics{
						Handle: class.Handle,
						Parent: class.Parent,
						Name:   class.Name,
						Stats:  info.Statistics,
					}

					// TODO: Add detailed statistics through adapter interface
					// For now, basic statistics from netlink are sufficient

					stats.ClassStats = append(stats.ClassStats, classStat)
					break
				}
			}
		}
	}

	// Get filter statistics (simplified)
	for _, filter := range readModel.Filters {
		filterStat := FilterStatistics{
			Parent:     filter.Parent,
			Priority:   filter.Priority,
			Protocol:   filter.Protocol,
			MatchCount: len(filter.Matches),
		}
		stats.FilterStats = append(stats.FilterStats, filterStat)
	}

	s.logger.Info("Device statistics collected",
		logging.String("device", deviceName),
		logging.Int("qdiscs", len(stats.QdiscStats)),
		logging.Int("classes", len(stats.ClassStats)),
		logging.Int("filters", len(stats.FilterStats)))

	return stats, nil
}

// GetRealtimeStatistics gets real-time statistics without read model
func (s *StatisticsQueryService) GetRealtimeStatistics(ctx context.Context, deviceName string) (*DeviceStatistics, error) {
	device, err := tc.NewDevice(deviceName)
	if err != nil {
		return nil, fmt.Errorf("invalid device name: %w", err)
	}

	stats := &DeviceStatistics{
		DeviceName:  deviceName,
		Timestamp:   time.Now(),
		QdiscStats:  make([]QdiscStatistics, 0),
		ClassStats:  make([]ClassStatistics, 0),
		FilterStats: make([]FilterStatistics, 0),
	}

	// Get all qdiscs directly from netlink
	qdiscResult := s.netlinkAdapter.GetQdiscs(device)
	if qdiscResult.IsSuccess() {
		for _, info := range qdiscResult.Value() {
			qdiscStat := QdiscStatistics{
				Handle: info.Handle.String(),
				Type:   info.Type.String(),
				Stats:  info.Statistics,
			}
			stats.QdiscStats = append(stats.QdiscStats, qdiscStat)
		}
	}

	// Get all classes directly from netlink
	classResult := s.netlinkAdapter.GetClasses(device)
	if classResult.IsSuccess() {
		for _, info := range classResult.Value() {
			classStat := ClassStatistics{
				Handle: info.Handle.String(),
				Parent: info.Parent.String(),
				Stats:  info.Statistics,
			}
			stats.ClassStats = append(stats.ClassStats, classStat)
		}
	}

	return stats, nil
}

// parseHandle converts string handle to valueobject
func parseHandle(handleStr string) (tc.Handle, error) {
	var major, minor uint16
	n, err := fmt.Sscanf(handleStr, "%x:%x", &major, &minor)
	if err != nil || n != 2 {
		return tc.Handle{}, fmt.Errorf("invalid handle format: %s", handleStr)
	}
	return tc.NewHandle(major, minor), nil
}

// GetDeviceStatisticsHandler handles queries for device statistics
type GetDeviceStatisticsHandler struct {
	statisticsService *StatisticsQueryService
}

// NewGetDeviceStatisticsHandler creates a new handler
func NewGetDeviceStatisticsHandler(statisticsService *StatisticsQueryService) *GetDeviceStatisticsHandler {
	return &GetDeviceStatisticsHandler{
		statisticsService: statisticsService,
	}
}

// Handle processes the query
func (h *GetDeviceStatisticsHandler) Handle(ctx context.Context, query *models.GetDeviceStatisticsQuery) types.Result[models.DeviceStatisticsView] {
	stats, err := h.statisticsService.GetDeviceStatistics(ctx, query.DeviceName().String())
	if err != nil {
		return types.Failure[models.DeviceStatisticsView](fmt.Errorf("failed to get device statistics: %w", err))
	}

	// Convert to view model
	view := convertDeviceStatisticsToView(stats)
	return types.Success(view)
}

// GetRealtimeStatisticsHandler handles queries for realtime statistics
type GetRealtimeStatisticsHandler struct {
	statisticsService *StatisticsQueryService
}

// NewGetRealtimeStatisticsHandler creates a new handler
func NewGetRealtimeStatisticsHandler(statisticsService *StatisticsQueryService) *GetRealtimeStatisticsHandler {
	return &GetRealtimeStatisticsHandler{
		statisticsService: statisticsService,
	}
}

// Handle processes the query
func (h *GetRealtimeStatisticsHandler) Handle(ctx context.Context, query *models.GetRealtimeStatisticsQuery) types.Result[models.DeviceStatisticsView] {
	stats, err := h.statisticsService.GetRealtimeStatistics(ctx, query.DeviceName().String())
	if err != nil {
		return types.Failure[models.DeviceStatisticsView](fmt.Errorf("failed to get realtime statistics: %w", err))
	}

	// Convert to view model
	view := convertDeviceStatisticsToView(stats)
	return types.Success(view)
}

// Helper function to convert DeviceStatistics to models.DeviceStatisticsView
func convertDeviceStatisticsToView(stats *DeviceStatistics) models.DeviceStatisticsView {
	view := models.DeviceStatisticsView{
		DeviceName:  stats.DeviceName,
		Timestamp:   stats.Timestamp.Format(time.RFC3339),
		QdiscStats:  make([]models.QdiscStatisticsView, 0, len(stats.QdiscStats)),
		ClassStats:  make([]models.ClassStatisticsView, 0, len(stats.ClassStats)),
		FilterStats: make([]models.FilterStatisticsView, 0, len(stats.FilterStats)),
		LinkStats: models.LinkStatisticsView{
			RxBytes:   stats.LinkStats.RxBytes,
			TxBytes:   stats.LinkStats.TxBytes,
			RxPackets: stats.LinkStats.RxPackets,
			TxPackets: stats.LinkStats.TxPackets,
			RxErrors:  stats.LinkStats.RxErrors,
			TxErrors:  stats.LinkStats.TxErrors,
			RxDropped: stats.LinkStats.RxDropped,
			TxDropped: stats.LinkStats.TxDropped,
		},
	}

	// Convert qdisc statistics
	for _, qdisc := range stats.QdiscStats {
		qdiscView := models.QdiscStatisticsView{
			Handle:        qdisc.Handle,
			Type:          qdisc.Type,
			BytesSent:     qdisc.Stats.BytesSent,
			PacketsSent:   qdisc.Stats.PacketsSent,
			BytesDropped:  qdisc.Stats.BytesDropped,
			Overlimits:    qdisc.Stats.Overlimits,
			Requeues:      qdisc.Stats.Requeues,
			DetailedStats: make(map[string]interface{}),
		}

		if qdisc.DetailedStats != nil {
			qdiscView.Backlog = qdisc.DetailedStats.Backlog
			qdiscView.QueueLength = qdisc.DetailedStats.QueueLength
			qdiscView.DetailedStats["backlog_bytes"] = qdisc.DetailedStats.BacklogBytes
			qdiscView.DetailedStats["bytes_per_second"] = qdisc.DetailedStats.BytesPerSecond
			qdiscView.DetailedStats["packets_per_second"] = qdisc.DetailedStats.PacketsPerSecond
			if qdisc.DetailedStats.HTBStats != nil {
				qdiscView.DetailedStats["htb_direct_packets"] = qdisc.DetailedStats.HTBStats.DirectPackets
				qdiscView.DetailedStats["htb_version"] = qdisc.DetailedStats.HTBStats.Version
			}
		}

		view.QdiscStats = append(view.QdiscStats, qdiscView)
	}

	// Convert class statistics
	for _, class := range stats.ClassStats {
		classView := models.ClassStatisticsView{
			Handle:         class.Handle,
			Parent:         class.Parent,
			Name:           class.Name,
			BytesSent:      class.Stats.BytesSent,
			PacketsSent:    class.Stats.PacketsSent,
			BytesDropped:   class.Stats.BytesDropped,
			Overlimits:     class.Stats.Overlimits,
			BacklogBytes:   class.Stats.BacklogBytes,
			BacklogPackets: class.Stats.BacklogPackets,
			RateBPS:        class.Stats.RateBPS,
			DetailedStats:  make(map[string]interface{}),
		}

		if class.DetailedStats != nil && class.DetailedStats.HTBStats != nil {
			classView.DetailedStats["htb_lends"] = class.DetailedStats.HTBStats.Lends
			classView.DetailedStats["htb_borrows"] = class.DetailedStats.HTBStats.Borrows
			classView.DetailedStats["htb_giants"] = class.DetailedStats.HTBStats.Giants
			classView.DetailedStats["htb_tokens"] = class.DetailedStats.HTBStats.Tokens
			classView.DetailedStats["htb_ctokens"] = class.DetailedStats.HTBStats.CTokens
			classView.DetailedStats["htb_rate"] = class.DetailedStats.HTBStats.Rate
			classView.DetailedStats["htb_ceil"] = class.DetailedStats.HTBStats.Ceil
			classView.DetailedStats["htb_level"] = class.DetailedStats.HTBStats.Level
		}

		view.ClassStats = append(view.ClassStats, classView)
	}

	// Convert filter statistics
	for _, filter := range stats.FilterStats {
		filterView := models.FilterStatisticsView{
			Parent:     filter.Parent,
			Priority:   filter.Priority,
			Protocol:   filter.Protocol,
			MatchCount: filter.MatchCount,
		}
		view.FilterStats = append(view.FilterStats, filterView)
	}

	return view
}

// GetQdiscStatisticsHandler handles queries for qdisc statistics
type GetQdiscStatisticsHandler struct {
	netlinkAdapter netlink.Adapter
}

// NewGetQdiscStatisticsHandler creates a new handler
func NewGetQdiscStatisticsHandler(netlinkAdapter netlink.Adapter) *GetQdiscStatisticsHandler {
	return &GetQdiscStatisticsHandler{
		netlinkAdapter: netlinkAdapter,
	}
}

// Handle processes the query
func (h *GetQdiscStatisticsHandler) Handle(ctx context.Context, query *models.GetQdiscStatisticsQuery) types.Result[models.QdiscStatisticsView] {
	// Get qdisc information from netlink
	qdiscs := h.netlinkAdapter.GetQdiscs(query.DeviceName())
	if !qdiscs.IsSuccess() {
		return types.Failure[models.QdiscStatisticsView](fmt.Errorf("failed to get qdiscs: %w", qdiscs.Error()))
	}

	// Find the specific qdisc
	for _, qdisc := range qdiscs.Value() {
		if qdisc.Handle.String() == query.Handle().String() {
			view := models.QdiscStatisticsView{
				Handle:        qdisc.Handle.String(),
				Type:          qdisc.Type.String(),
				BytesSent:     qdisc.Statistics.BytesSent,
				PacketsSent:   qdisc.Statistics.PacketsSent,
				BytesDropped:  qdisc.Statistics.BytesDropped,
				Overlimits:    qdisc.Statistics.Overlimits,
				Requeues:      qdisc.Statistics.Requeues,
				DetailedStats: make(map[string]interface{}),
			}

			// TODO: Add detailed statistics collection through adapter interface

			return types.Success(view)
		}
	}

	return types.Failure[models.QdiscStatisticsView](fmt.Errorf("qdisc %s not found on device %s", query.Handle(), query.DeviceName()))
}

// GetClassStatisticsHandler handles queries for class statistics
type GetClassStatisticsHandler struct {
	netlinkAdapter netlink.Adapter
}

// NewGetClassStatisticsHandler creates a new handler
func NewGetClassStatisticsHandler(netlinkAdapter netlink.Adapter) *GetClassStatisticsHandler {
	return &GetClassStatisticsHandler{
		netlinkAdapter: netlinkAdapter,
	}
}

// Handle processes the query
func (h *GetClassStatisticsHandler) Handle(ctx context.Context, query *models.GetClassStatisticsQuery) types.Result[models.ClassStatisticsView] {
	// Get class information from netlink
	classes := h.netlinkAdapter.GetClasses(query.DeviceName())
	if !classes.IsSuccess() {
		return types.Failure[models.ClassStatisticsView](fmt.Errorf("failed to get classes: %w", classes.Error()))
	}

	// Find the specific class
	for _, class := range classes.Value() {
		if class.Handle.String() == query.Handle().String() {
			view := models.ClassStatisticsView{
				Handle:         class.Handle.String(),
				Parent:         class.Parent.String(),
				Name:           "", // Name not available from netlink directly
				BytesSent:      class.Statistics.BytesSent,
				PacketsSent:    class.Statistics.PacketsSent,
				BytesDropped:   class.Statistics.BytesDropped,
				Overlimits:     class.Statistics.Overlimits,
				BacklogBytes:   class.Statistics.BacklogBytes,
				BacklogPackets: class.Statistics.BacklogPackets,
				RateBPS:        class.Statistics.RateBPS,
				DetailedStats:  make(map[string]interface{}),
			}

			// TODO: Add detailed statistics collection through adapter interface

			return types.Success(view)
		}
	}

	return types.Failure[models.ClassStatisticsView](fmt.Errorf("class %s not found on device %s", query.Handle(), query.DeviceName()))
}
