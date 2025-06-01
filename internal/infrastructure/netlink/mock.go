package netlink

import (
	"context"
	"fmt"
	"sync"

	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
	"github.com/rng999/traffic-control-go/pkg/types"
)

// MockAdapter is a mock implementation for testing
type MockAdapter struct {
	mu      sync.RWMutex
	qdiscs  map[string]map[valueobjects.Handle]QdiscInfo // device -> handle -> qdisc
	classes map[string]map[valueobjects.Handle]ClassInfo // device -> handle -> class
	filters map[string][]FilterInfo                      // device -> filters
}

// NewMockAdapter creates a new mock adapter
func NewMockAdapter() *MockAdapter {
	return &MockAdapter{
		qdiscs:  make(map[string]map[valueobjects.Handle]QdiscInfo),
		classes: make(map[string]map[valueobjects.Handle]ClassInfo),
		filters: make(map[string][]FilterInfo),
	}
}

// AddQdisc adds a qdisc (new interface)
func (m *MockAdapter) AddQdisc(ctx context.Context, qdisc *entities.Qdisc) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	deviceStr := qdisc.Device().String()

	// Initialize device map if needed
	if _, exists := m.qdiscs[deviceStr]; !exists {
		m.qdiscs[deviceStr] = make(map[valueobjects.Handle]QdiscInfo)
	}

	// Check if qdisc already exists
	if _, exists := m.qdiscs[deviceStr][qdisc.Handle()]; exists {
		return fmt.Errorf("qdisc %s already exists on device %s", qdisc.Handle(), qdisc.Device())
	}

	// Add the qdisc
	m.qdiscs[deviceStr][qdisc.Handle()] = QdiscInfo{
		Handle:     qdisc.Handle(),
		Parent:     qdisc.Parent(),
		Type:       valueobjects.QdiscType(qdisc.Type()),
		Statistics: QdiscStats{},
	}

	return nil
}

// DeleteQdisc deletes a qdisc
func (m *MockAdapter) DeleteQdisc(device valueobjects.DeviceName, handle valueobjects.Handle) types.Result[Unit] {
	m.mu.Lock()
	defer m.mu.Unlock()

	deviceStr := device.String()

	if qdiscs, exists := m.qdiscs[deviceStr]; exists {
		if _, qdiscExists := qdiscs[handle]; qdiscExists {
			delete(qdiscs, handle)
			return types.Success(Unit{})
		}
	}

	return types.Failure[Unit](fmt.Errorf("qdisc %s not found on device %s", handle, device))
}

// GetQdiscs returns all qdiscs for a device
func (m *MockAdapter) GetQdiscs(device valueobjects.DeviceName) types.Result[[]QdiscInfo] {
	m.mu.RLock()
	defer m.mu.RUnlock()

	deviceStr := device.String()

	var result []QdiscInfo
	if qdiscs, exists := m.qdiscs[deviceStr]; exists {
		for _, qdisc := range qdiscs {
			result = append(result, qdisc)
		}
	}

	return types.Success(result)
}

// AddClass adds a class (new interface)
func (m *MockAdapter) AddClass(ctx context.Context, class *entities.Class) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	deviceStr := class.Device().String()

	// Initialize device map if needed
	if _, exists := m.classes[deviceStr]; !exists {
		m.classes[deviceStr] = make(map[valueobjects.Handle]ClassInfo)
	}

	// Check if class already exists
	if _, exists := m.classes[deviceStr][class.Handle()]; exists {
		return fmt.Errorf("class %s already exists on device %s", class.Handle(), class.Device())
	}

	// Add the class
	m.classes[deviceStr][class.Handle()] = ClassInfo{
		Handle:     class.Handle(),
		Parent:     class.Parent(),
		Type:       valueobjects.QdiscType("htb"), // Default to HTB for classes
		Statistics: ClassStats{},
	}

	return nil
}

// DeleteClass deletes a class
func (m *MockAdapter) DeleteClass(device valueobjects.DeviceName, handle valueobjects.Handle) types.Result[Unit] {
	m.mu.Lock()
	defer m.mu.Unlock()

	deviceStr := device.String()

	if classes, exists := m.classes[deviceStr]; exists {
		if _, classExists := classes[handle]; classExists {
			delete(classes, handle)
			return types.Success(Unit{})
		}
	}

	return types.Failure[Unit](fmt.Errorf("class %s not found on device %s", handle, device))
}

// GetClasses returns all classes for a device
func (m *MockAdapter) GetClasses(device valueobjects.DeviceName) types.Result[[]ClassInfo] {
	m.mu.RLock()
	defer m.mu.RUnlock()

	deviceStr := device.String()

	var result []ClassInfo
	if classes, exists := m.classes[deviceStr]; exists {
		for _, class := range classes {
			result = append(result, class)
		}
	}

	return types.Success(result)
}

// AddFilter adds a filter (new interface)
func (m *MockAdapter) AddFilter(ctx context.Context, filter *entities.Filter) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	deviceStr := filter.Device().String()

	// Initialize device filter slice if needed
	if _, exists := m.filters[deviceStr]; !exists {
		m.filters[deviceStr] = make([]FilterInfo, 0)
	}

	// Add the filter
	filterInfo := FilterInfo{
		Parent:   filter.Parent(),
		Priority: filter.ID().Priority(),
		Protocol: filter.Protocol(),
		Handle:   filter.ID().Handle(),
		FlowID:   filter.FlowID(),
		Matches:  make([]MatchInfo, 0),
	}

	// Convert matches
	for _, match := range filter.Matches() {
		filterInfo.Matches = append(filterInfo.Matches, MatchInfo{
			Type:  match.Type(),
			Value: match.String(),
		})
	}

	m.filters[deviceStr] = append(m.filters[deviceStr], filterInfo)
	return nil
}

// DeleteFilter deletes a filter
func (m *MockAdapter) DeleteFilter(device valueobjects.DeviceName, parent valueobjects.Handle, priority uint16, handle valueobjects.Handle) types.Result[Unit] {
	m.mu.Lock()
	defer m.mu.Unlock()

	deviceStr := device.String()

	if filters, exists := m.filters[deviceStr]; exists {
		for i, filter := range filters {
			if filter.Parent == parent && filter.Priority == priority && filter.Handle == handle {
				// Remove filter from slice
				m.filters[deviceStr] = append(filters[:i], filters[i+1:]...)
				return types.Success(Unit{})
			}
		}
	}

	return types.Failure[Unit](fmt.Errorf("filter not found on device %s", device))
}

// GetFilters returns all filters for a device
func (m *MockAdapter) GetFilters(device valueobjects.DeviceName) types.Result[[]FilterInfo] {
	m.mu.RLock()
	defer m.mu.RUnlock()

	deviceStr := device.String()

	var result []FilterInfo
	if filters, exists := m.filters[deviceStr]; exists {
		result = append(result, filters...)
	}

	return types.Success(result)
}

// GetLinkStatistics returns link statistics for a device
func (m *MockAdapter) GetLinkStatistics(device valueobjects.DeviceName) types.Result[LinkStats] {
	// Return mock link statistics
	return types.Success(LinkStats{
		RxBytes:   1000000,
		TxBytes:   2000000,
		RxPackets: 10000,
		TxPackets: 20000,
		RxErrors:  0,
		TxErrors:  0,
		RxDropped: 5,
		TxDropped: 3,
	})
}

// SetQdiscStatistics sets mock statistics for a qdisc (for testing)
func (m *MockAdapter) SetQdiscStatistics(device valueobjects.DeviceName, handle valueobjects.Handle, stats QdiscStats) {
	m.mu.Lock()
	defer m.mu.Unlock()

	deviceStr := device.String()
	
	if qdiscs, exists := m.qdiscs[deviceStr]; exists {
		if qdisc, qdiscExists := qdiscs[handle]; qdiscExists {
			qdisc.Statistics = stats
			qdiscs[handle] = qdisc
		}
	}
}

// SetClassStatistics sets mock statistics for a class (for testing)
func (m *MockAdapter) SetClassStatistics(device valueobjects.DeviceName, handle valueobjects.Handle, stats ClassStats) {
	m.mu.Lock()
	defer m.mu.Unlock()

	deviceStr := device.String()
	
	if classes, exists := m.classes[deviceStr]; exists {
		if class, classExists := classes[handle]; classExists {
			class.Statistics = stats
			classes[handle] = class
		}
	}
}