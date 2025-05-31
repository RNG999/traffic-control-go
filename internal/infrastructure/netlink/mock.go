package netlink

import (
	"fmt"
	"sync"

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

// AddQdisc adds a qdisc
func (m *MockAdapter) AddQdisc(device valueobjects.DeviceName, config QdiscConfig) types.Result[Unit] {
	m.mu.Lock()
	defer m.mu.Unlock()

	deviceStr := device.String()

	// Initialize device map if needed
	if _, exists := m.qdiscs[deviceStr]; !exists {
		m.qdiscs[deviceStr] = make(map[valueobjects.Handle]QdiscInfo)
	}

	// Check if qdisc already exists
	if _, exists := m.qdiscs[deviceStr][config.Handle]; exists {
		return types.Failure[Unit](fmt.Errorf("qdisc %s already exists on device %s", config.Handle, device))
	}

	// Add the qdisc
	m.qdiscs[deviceStr][config.Handle] = QdiscInfo{
		Handle:     config.Handle,
		Parent:     config.Parent,
		Type:       config.Type,
		Statistics: QdiscStats{},
	}

	return types.Success(Unit{})
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

// AddClass adds a class
func (m *MockAdapter) AddClass(device valueobjects.DeviceName, config ClassConfig) types.Result[Unit] {
	m.mu.Lock()
	defer m.mu.Unlock()

	deviceStr := device.String()

	// Initialize device map if needed
	if _, exists := m.classes[deviceStr]; !exists {
		m.classes[deviceStr] = make(map[valueobjects.Handle]ClassInfo)
	}

	// Check if class already exists
	if _, exists := m.classes[deviceStr][config.Handle]; exists {
		return types.Failure[Unit](fmt.Errorf("class %s already exists on device %s", config.Handle, device))
	}

	// Verify parent exists (either qdisc or class)
	parentExists := false
	if qdiscs, hasQdiscs := m.qdiscs[deviceStr]; hasQdiscs {
		if _, exists := qdiscs[config.Parent]; exists {
			parentExists = true
		}
	}
	if !parentExists {
		if classes, hasClasses := m.classes[deviceStr]; hasClasses {
			if _, exists := classes[config.Parent]; exists {
				parentExists = true
			}
		}
	}

	if !parentExists {
		return types.Failure[Unit](fmt.Errorf("parent %s not found on device %s", config.Parent, device))
	}

	// Add the class
	m.classes[deviceStr][config.Handle] = ClassInfo{
		Handle:     config.Handle,
		Parent:     config.Parent,
		Type:       config.Type,
		Statistics: ClassStats{},
	}

	return types.Success(Unit{})
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

// AddFilter adds a filter
func (m *MockAdapter) AddFilter(device valueobjects.DeviceName, config FilterConfig) types.Result[Unit] {
	m.mu.Lock()
	defer m.mu.Unlock()

	deviceStr := device.String()

	// Initialize device filters if needed
	if _, exists := m.filters[deviceStr]; !exists {
		m.filters[deviceStr] = make([]FilterInfo, 0)
	}

	// Check for duplicate
	for _, existing := range m.filters[deviceStr] {
		if existing.Parent.Equals(config.Parent) &&
			existing.Priority == config.Priority &&
			existing.Handle.Equals(config.Handle) {
			return types.Failure[Unit](fmt.Errorf("filter already exists on device %s", device))
		}
	}

	// Add the filter
	m.filters[deviceStr] = append(m.filters[deviceStr], FilterInfo(config))

	return types.Success(Unit{})
}

// DeleteFilter deletes a filter
func (m *MockAdapter) DeleteFilter(device valueobjects.DeviceName, parent valueobjects.Handle, priority uint16, handle valueobjects.Handle) types.Result[Unit] {
	m.mu.Lock()
	defer m.mu.Unlock()

	deviceStr := device.String()

	if filters, exists := m.filters[deviceStr]; exists {
		for i, filter := range filters {
			if filter.Parent.Equals(parent) &&
				filter.Priority == priority &&
				filter.Handle.Equals(handle) {
				// Remove the filter
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

	result := []FilterInfo{}
	if filters, exists := m.filters[deviceStr]; exists {
		result = append(result, filters...)
	}

	return types.Success(result)
}
