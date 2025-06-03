package application

import (
	"context"
	"fmt"
	"time"

	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	"github.com/rng999/traffic-control-go/internal/projections"
	"github.com/rng999/traffic-control-go/pkg/logging"
)

// StatisticsService provides TC statistics collection functionality
type StatisticsService struct {
	netlinkAdapter netlink.Adapter
	readModelStore projections.ReadModelStore
	logger         logging.Logger
}

// NewStatisticsService creates a new statistics service
func NewStatisticsService(netlinkAdapter netlink.Adapter, readModelStore projections.ReadModelStore) *StatisticsService {
	return &StatisticsService{
		netlinkAdapter: netlinkAdapter,
		readModelStore: readModelStore,
		logger:         logging.WithComponent("application.statistics"),
	}
}

// DeviceStatistics represents statistics for a device
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
func (s *StatisticsService) GetDeviceStatistics(ctx context.Context, deviceName string) (*DeviceStatistics, error) {
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

	device, err := valueobjects.NewDevice(deviceName)
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

					// Try to get detailed stats - simplified for compilation
					// TODO: Implement proper adapter wrapper access
					// if adapter, ok := s.netlinkAdapter.(*netlink.AdapterWrapper); ok {
					//     if realAdapter, ok := adapter.RealAdapter().(*netlink.RealNetlinkAdapter); ok {
					//         detailed := realAdapter.GetDetailedQdiscStats(device, handle)
					//         if detailed.IsSuccess() {
					//             qdiscStat.DetailedStats = &detailed.Value()
					//         }
					//     }
					// }

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

					// Try to get detailed stats - simplified for compilation
					// TODO: Implement proper adapter wrapper access
					// if adapter, ok := s.netlinkAdapter.(*netlink.AdapterWrapper); ok {
					//     if realAdapter, ok := adapter.RealAdapter().(*netlink.RealNetlinkAdapter); ok {
					//         detailed := realAdapter.GetDetailedClassStats(device, handle)
					//         if detailed.IsSuccess() {
					//             classStat.DetailedStats = &detailed.Value()
					//         }
					//     }
					// }

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
func (s *StatisticsService) GetRealtimeStatistics(ctx context.Context, deviceName string) (*DeviceStatistics, error) {
	device, err := valueobjects.NewDevice(deviceName)
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

// MonitorStatistics continuously monitors statistics
func (s *StatisticsService) MonitorStatistics(ctx context.Context, deviceName string, interval time.Duration, callback func(*DeviceStatistics)) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.logger.Info("Starting statistics monitoring",
		logging.String("device", deviceName),
		logging.String("interval", interval.String()))

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping statistics monitoring",
				logging.String("device", deviceName))
			return ctx.Err()
		case <-ticker.C:
			stats, err := s.GetDeviceStatistics(ctx, deviceName)
			if err != nil {
				s.logger.Error("Failed to get statistics",
					logging.String("device", deviceName),
					logging.Error(err))
				continue
			}
			callback(stats)
		}
	}
}
