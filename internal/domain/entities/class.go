package entities

import (
	"fmt"

	"github.com/rng999/traffic-control-go/pkg/tc"
)

// ClassID represents a unique identifier for a traffic class
type ClassID struct {
	device tc.DeviceName
	handle tc.Handle
}

// NewClassID creates a new ClassID
func NewClassID(device tc.DeviceName, handle tc.Handle) ClassID {
	return ClassID{device: device, handle: handle}
}

// String returns the string representation of ClassID
func (id ClassID) String() string {
	return fmt.Sprintf("%s:%s", id.device, id.handle)
}

// Device returns the device name
func (id ClassID) Device() tc.DeviceName {
	return id.device
}

// Class represents a traffic class entity
type Class struct {
	id       ClassID
	parent   tc.Handle
	name     string      // Human-readable name
	priority *Priority   // Priority must be explicitly set
	depth    int         // Hierarchy depth (0 = root level)
	children []tc.Handle // Child class handles
}

// Priority represents the priority level of a class (0-7, where 0 is highest priority)
type Priority int

// NewClass creates a new Class entity
func NewClass(device tc.DeviceName, handle tc.Handle, parent tc.Handle, name string, priority Priority) *Class {
	return &Class{
		id:       NewClassID(device, handle),
		parent:   parent,
		name:     name,
		priority: &priority,
		depth:    0, // Will be calculated based on hierarchy
		children: make([]tc.Handle, 0),
	}
}

// ID returns the class ID
func (c *Class) ID() ClassID {
	return c.id
}

// Handle returns the class handle
func (c *Class) Handle() tc.Handle {
	return c.id.handle
}

// Parent returns the parent handle
func (c *Class) Parent() tc.Handle {
	return c.parent
}

// Name returns the human-readable name
func (c *Class) Name() string {
	return c.name
}

// Priority returns the priority
func (c *Class) Priority() *Priority {
	return c.priority
}

// SetPriority sets the priority
func (c *Class) SetPriority(p Priority) {
	c.priority = &p
}

// SetName updates the class name
func (c *Class) SetName(name string) error {
	if name == "" {
		return fmt.Errorf("class name cannot be empty")
	}
	c.name = name
	return nil
}

// SetParent updates the parent handle (use with caution - should validate hierarchy)
func (c *Class) SetParent(parent tc.Handle) {
	c.parent = parent
}

// Depth returns the hierarchy depth
func (c *Class) Depth() int {
	return c.depth
}

// SetDepth sets the hierarchy depth
func (c *Class) SetDepth(depth int) {
	c.depth = depth
}

// Children returns the child class handles
func (c *Class) Children() []tc.Handle {
	return c.children
}

// AddChild adds a child class handle
func (c *Class) AddChild(childHandle tc.Handle) {
	c.children = append(c.children, childHandle)
}

// RemoveChild removes a child class handle
func (c *Class) RemoveChild(childHandle tc.Handle) {
	for i, child := range c.children {
		if child.String() == childHandle.String() {
			c.children = append(c.children[:i], c.children[i+1:]...)
			break
		}
	}
}

// HasChildren returns true if the class has child classes
func (c *Class) HasChildren() bool {
	return len(c.children) > 0
}

// IsLeaf returns true if the class has no children
func (c *Class) IsLeaf() bool {
	return len(c.children) == 0
}

// HTBClass represents an HTB-specific traffic class
type HTBClass struct {
	*Class
	rate     tc.Bandwidth
	ceil     tc.Bandwidth
	burst    uint32
	cburst   uint32
	quantum  uint32 // Quantum for borrowing (bytes)
	overhead uint32 // Packet overhead calculation (bytes)
	mpu      uint32 // Minimum packet unit (bytes)
	mtu      uint32 // Maximum transmission unit (bytes)
	prio     uint32 // Internal HTB priority (0-7)
}

// NewHTBClass creates a new HTB class
func NewHTBClass(device tc.DeviceName, handle tc.Handle, parent tc.Handle, name string, priority Priority) *HTBClass {
	class := NewClass(device, handle, parent, name, priority)
	return &HTBClass{
		Class: class,
	}
}

// SetRate sets the guaranteed rate
func (h *HTBClass) SetRate(rate tc.Bandwidth) {
	h.rate = rate
}

// Rate returns the guaranteed rate
func (h *HTBClass) Rate() tc.Bandwidth {
	return h.rate
}

// SetCeil sets the maximum rate
func (h *HTBClass) SetCeil(ceil tc.Bandwidth) {
	h.ceil = ceil
}

// Ceil returns the maximum rate
func (h *HTBClass) Ceil() tc.Bandwidth {
	return h.ceil
}

// SetBurst sets the burst size
func (h *HTBClass) SetBurst(burst uint32) {
	h.burst = burst
}

// Burst returns the burst size
func (h *HTBClass) Burst() uint32 {
	return h.burst
}

// SetCburst sets the ceil burst size
func (h *HTBClass) SetCburst(cburst uint32) {
	h.cburst = cburst
}

// Cburst returns the ceil burst size
func (h *HTBClass) Cburst() uint32 {
	return h.cburst
}

// CalculateBurst calculates appropriate burst size based on rate
func (h *HTBClass) CalculateBurst() uint32 {
	// Basic calculation: rate_bps / 8 * 0.01s (10ms timer)
	// Multiply by 10 for safety margin
	bytesPerSecond := h.rate.BitsPerSecond() / 8
	// Avoid floating point: 0.01 * 10 = 0.1 = 1/10
	burstValue := bytesPerSecond / 10
	if burstValue > 0xFFFFFFFF {
		return 0xFFFFFFFF // Cap at maximum uint32 value
	}
	return uint32(burstValue)
}

// CalculateCburst calculates appropriate cburst size based on ceil
func (h *HTBClass) CalculateCburst() uint32 {
	// Use ceil if set, otherwise use rate
	bandwidth := h.ceil
	if bandwidth.BitsPerSecond() == 0 {
		bandwidth = h.rate
	}

	bytesPerSecond := bandwidth.BitsPerSecond() / 8
	// Avoid floating point: 0.01 * 10 = 0.1 = 1/10
	cburstValue := bytesPerSecond / 10
	if cburstValue > 0xFFFFFFFF {
		return 0xFFFFFFFF // Cap at maximum uint32 value
	}
	return uint32(cburstValue)
}

// Enhanced HTB methods with comprehensive parameter support

// ClassHierarchy provides utilities for managing class hierarchies and validation (WP2 implementation)
type ClassHierarchy struct {
	maxDepth    int                       // Maximum allowed hierarchy depth
	classes     map[tc.Handle]*Class      // Map of handle to class for quick lookups
	parentMap   map[tc.Handle]tc.Handle   // Map of child handle to parent handle
	childrenMap map[tc.Handle][]tc.Handle // Map of parent handle to children handles
}

// NewClassHierarchy creates a new ClassHierarchy manager
func NewClassHierarchy(maxDepth int) *ClassHierarchy {
	return &ClassHierarchy{
		maxDepth:    maxDepth,
		classes:     make(map[tc.Handle]*Class),
		parentMap:   make(map[tc.Handle]tc.Handle),
		childrenMap: make(map[tc.Handle][]tc.Handle),
	}
}

// AddClass adds a class to the hierarchy
func (ch *ClassHierarchy) AddClass(class *Class) error {
	handle := class.Handle()
	parent := class.Parent()

	// Validate depth if parent exists and is not root
	if !parent.IsRoot() {
		// Check if parent exists in hierarchy
		if _, exists := ch.classes[parent]; !exists {
			return fmt.Errorf("parent class %s not found in hierarchy", parent)
		}

		depth, err := ch.CalculateDepth(parent)
		if err != nil {
			return fmt.Errorf("invalid parent: %w", err)
		}
		if depth+1 > ch.maxDepth {
			return fmt.Errorf("adding class would exceed maximum depth %d", ch.maxDepth)
		}
		class.SetDepth(depth + 1)
	} else {
		class.SetDepth(0)
	}

	// Check for circular dependencies
	if err := ch.validateNoCycle(handle, parent); err != nil {
		return err
	}

	// Add to maps
	ch.classes[handle] = class
	if !parent.IsRoot() {
		ch.parentMap[handle] = parent
		ch.childrenMap[parent] = append(ch.childrenMap[parent], handle)
	}

	return nil
}

// RemoveClass removes a class and its descendants from the hierarchy
func (ch *ClassHierarchy) RemoveClass(handle tc.Handle) error {
	class, exists := ch.classes[handle]
	if !exists {
		return fmt.Errorf("class %s not found", handle)
	}

	// Remove all descendants first
	children := ch.childrenMap[handle]
	for _, child := range children {
		if err := ch.RemoveClass(child); err != nil {
			return fmt.Errorf("failed to remove child %s: %w", child, err)
		}
	}

	// Remove from parent's children list
	parent := class.Parent()
	if !parent.IsRoot() {
		parentChildren := ch.childrenMap[parent]
		for i, child := range parentChildren {
			if child.String() == handle.String() {
				ch.childrenMap[parent] = append(parentChildren[:i], parentChildren[i+1:]...)
				break
			}
		}
	}

	// Remove from maps
	delete(ch.classes, handle)
	delete(ch.parentMap, handle)
	delete(ch.childrenMap, handle)

	return nil
}

// CalculateDepth calculates the depth of a class in the hierarchy
func (ch *ClassHierarchy) CalculateDepth(handle tc.Handle) (int, error) {
	// If the handle is root, depth is 0
	if handle.IsRoot() {
		return 0, nil
	}

	// Get the class to check its parent
	class, exists := ch.classes[handle]
	if !exists {
		return 0, fmt.Errorf("class %s not found in hierarchy", handle)
	}

	// If the class parent is root, depth is 0 (first level under root)
	if class.Parent().IsRoot() {
		return 0, nil
	}

	visited := make(map[tc.Handle]bool)
	depth := 0
	current := handle

	for !current.IsRoot() {
		if visited[current] {
			return 0, fmt.Errorf("circular dependency detected involving class %s", current)
		}
		visited[current] = true

		classObj, exists := ch.classes[current]
		if !exists {
			return 0, fmt.Errorf("class %s not found in hierarchy", current)
		}

		parent := classObj.Parent()
		if parent.IsRoot() {
			break // We're at level 0 (direct child of root)
		}

		current = parent
		depth++

		if depth > ch.maxDepth {
			return 0, fmt.Errorf("depth exceeds maximum allowed depth %d", ch.maxDepth)
		}
	}

	return depth, nil
}

// validateNoCycle checks for circular dependencies when adding a class
func (ch *ClassHierarchy) validateNoCycle(childHandle, parentHandle tc.Handle) error {
	if parentHandle.IsRoot() {
		return nil
	}

	// Check if the proposed parent is actually a descendant of the child
	visited := make(map[tc.Handle]bool)
	current := parentHandle

	for !current.IsRoot() {
		if current.String() == childHandle.String() {
			return fmt.Errorf("circular dependency: %s cannot be parent of %s", parentHandle, childHandle)
		}

		if visited[current] {
			return fmt.Errorf("circular dependency detected in existing hierarchy at %s", current)
		}
		visited[current] = true

		parent, exists := ch.parentMap[current]
		if !exists {
			break
		}
		current = parent
	}

	return nil
}

// GetChildren returns all direct children of a class
func (ch *ClassHierarchy) GetChildren(handle tc.Handle) []tc.Handle {
	return ch.childrenMap[handle]
}

// GetDescendants returns all descendants (children, grandchildren, etc.) of a class
func (ch *ClassHierarchy) GetDescendants(handle tc.Handle) []tc.Handle {
	var descendants []tc.Handle
	children := ch.childrenMap[handle]

	for _, child := range children {
		descendants = append(descendants, child)
		descendants = append(descendants, ch.GetDescendants(child)...)
	}

	return descendants
}

// GetAncestors returns all ancestors (parent, grandparent, etc.) of a class
func (ch *ClassHierarchy) GetAncestors(handle tc.Handle) []tc.Handle {
	var ancestors []tc.Handle
	current := handle

	for !current.IsRoot() {
		parent, exists := ch.parentMap[current]
		if !exists {
			break
		}
		ancestors = append(ancestors, parent)
		current = parent
	}

	return ancestors
}

// ValidateHierarchy validates the entire hierarchy for consistency
func (ch *ClassHierarchy) ValidateHierarchy() error {
	// Check all classes for valid depth and no cycles
	for handle := range ch.classes {
		if _, err := ch.CalculateDepth(handle); err != nil {
			return fmt.Errorf("validation failed for class %s: %w", handle, err)
		}
	}

	// Check that all parent-child relationships are bidirectional
	for child, parent := range ch.parentMap {
		parentChildren := ch.childrenMap[parent]
		found := false
		for _, c := range parentChildren {
			if c.String() == child.String() {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("inconsistent hierarchy: child %s not found in parent %s children list", child, parent)
		}
	}

	return nil
}

// ModifyClass updates a class with new properties
func (ch *ClassHierarchy) ModifyClass(handle tc.Handle, modifications ClassModifications) error {
	class, exists := ch.classes[handle]
	if !exists {
		return fmt.Errorf("class %s not found", handle)
	}

	// Apply name modification if provided
	if modifications.Name != nil {
		if err := class.SetName(*modifications.Name); err != nil {
			return fmt.Errorf("failed to set name: %w", err)
		}
	}

	// Apply priority modification if provided
	if modifications.Priority != nil {
		class.SetPriority(*modifications.Priority)
	}

	// Apply parent modification if provided (most complex)
	if modifications.NewParent != nil {
		if err := ch.MoveClass(handle, *modifications.NewParent); err != nil {
			return fmt.Errorf("failed to move class: %w", err)
		}
	}

	return nil
}

// MoveClass moves a class to a new parent, validating the hierarchy
func (ch *ClassHierarchy) MoveClass(handle, newParent tc.Handle) error {
	class, exists := ch.classes[handle]
	if !exists {
		return fmt.Errorf("class %s not found", handle)
	}

	oldParent := class.Parent()

	// Don't move if already at the correct parent
	if oldParent.String() == newParent.String() {
		return nil
	}

	// Validate the new parent exists (unless it's root)
	if !newParent.IsRoot() {
		if _, exists := ch.classes[newParent]; !exists {
			return fmt.Errorf("new parent %s not found in hierarchy", newParent)
		}
	}

	// Check for circular dependency
	if err := ch.validateNoCycle(handle, newParent); err != nil {
		return err
	}

	// Calculate new depth and validate
	var newDepth int
	if newParent.IsRoot() {
		newDepth = 0
	} else {
		parentDepth, err := ch.CalculateDepth(newParent)
		if err != nil {
			return fmt.Errorf("failed to calculate parent depth: %w", err)
		}
		newDepth = parentDepth + 1
	}

	if newDepth > ch.maxDepth {
		return fmt.Errorf("moving class would exceed maximum depth %d", ch.maxDepth)
	}

	// Validate that all descendants would still be within depth limits
	descendants := ch.GetDescendants(handle)
	for _, descendant := range descendants {
		descendantClass := ch.classes[descendant]
		descendantDepthFromHandle := descendantClass.Depth() - class.Depth()
		if newDepth+descendantDepthFromHandle > ch.maxDepth {
			return fmt.Errorf("moving class would cause descendant %s to exceed maximum depth %d", descendant, ch.maxDepth)
		}
	}

	// Remove from old parent's children list
	if !oldParent.IsRoot() {
		ch.removeFromChildrenList(oldParent, handle)
	}

	// Update the class's parent
	class.SetParent(newParent)
	class.SetDepth(newDepth)

	// Update internal maps
	if !newParent.IsRoot() {
		ch.parentMap[handle] = newParent
		ch.childrenMap[newParent] = append(ch.childrenMap[newParent], handle)
	} else {
		delete(ch.parentMap, handle)
	}

	// Update depths of all descendants
	ch.updateDescendantDepths(handle, newDepth)

	return nil
}

// removeFromChildrenList removes a child from a parent's children list
func (ch *ClassHierarchy) removeFromChildrenList(parent, child tc.Handle) {
	children := ch.childrenMap[parent]
	for i, c := range children {
		if c.String() == child.String() {
			ch.childrenMap[parent] = append(children[:i], children[i+1:]...)
			break
		}
	}
}

// updateDescendantDepths recursively updates the depth of all descendants
func (ch *ClassHierarchy) updateDescendantDepths(handle tc.Handle, parentDepth int) {
	children := ch.childrenMap[handle]
	for _, child := range children {
		childClass := ch.classes[child]
		childClass.SetDepth(parentDepth + 1)
		ch.updateDescendantDepths(child, parentDepth+1)
	}
}

// DeleteClass removes a class from the hierarchy with different strategies for handling children
func (ch *ClassHierarchy) DeleteClass(handle tc.Handle, strategy DeletionStrategy) error {
	class, exists := ch.classes[handle]
	if !exists {
		return fmt.Errorf("class %s not found", handle)
	}

	children := ch.GetChildren(handle)

	switch strategy {
	case DeleteCascade:
		// Remove all descendants first
		return ch.RemoveClass(handle) // Uses existing cascade logic

	case DeletePromoteChildren:
		// Move all children to this class's parent
		parent := class.Parent()
		for _, child := range children {
			if err := ch.MoveClass(child, parent); err != nil {
				return fmt.Errorf("failed to promote child %s: %w", child, err)
			}
		}
		// Now remove the class (it should have no children)
		return ch.removeClassOnly(handle)

	case DeleteOrphanChildren:
		// Move children to root (make them top-level classes)
		rootHandle := tc.NewHandle(1, 0) // Create the root handle (1:)
		for _, child := range children {
			if err := ch.MoveClass(child, rootHandle); err != nil {
				return fmt.Errorf("failed to orphan child %s: %w", child, err)
			}
		}
		// Now remove the class
		return ch.removeClassOnly(handle)

	case DeleteFailIfChildren:
		// Fail if the class has children
		if len(children) > 0 {
			return fmt.Errorf("cannot delete class %s: has %d children", handle, len(children))
		}
		return ch.removeClassOnly(handle)

	default:
		return fmt.Errorf("unknown deletion strategy: %v", strategy)
	}
}

// removeClassOnly removes a single class without affecting children (assumes children are already handled)
func (ch *ClassHierarchy) removeClassOnly(handle tc.Handle) error {
	class, exists := ch.classes[handle]
	if !exists {
		return fmt.Errorf("class %s not found", handle)
	}

	// Remove from parent's children list
	parent := class.Parent()
	if !parent.IsRoot() {
		ch.removeFromChildrenList(parent, handle)
	}

	// Remove from maps
	delete(ch.classes, handle)
	delete(ch.parentMap, handle)
	delete(ch.childrenMap, handle)

	return nil
}

// ApplyPriorityInheritance applies priority inheritance rules to all classes in the hierarchy
func (ch *ClassHierarchy) ApplyPriorityInheritance(rule PriorityInheritanceRule) error {
	if rule == NoInheritance {
		return nil // Nothing to do
	}

	// Process classes level by level to ensure parents are processed before children
	for depth := 0; depth <= ch.maxDepth; depth++ {
		for handle, class := range ch.classes {
			if class.Depth() == depth && !class.Parent().IsRoot() {
				parent, exists := ch.classes[class.Parent()]
				if !exists {
					return fmt.Errorf("parent class %s not found for %s", class.Parent(), handle)
				}

				if parent.Priority() == nil {
					continue // Skip if parent has no priority set
				}

				var newPriority Priority
				switch rule {
				case InheritParentPriority:
					newPriority = *parent.Priority()
				case InheritParentPlusOne:
					newPriority = *parent.Priority() + 1
					if newPriority > 7 { // HTB priority is 0-7
						newPriority = 7
					}
				}

				class.SetPriority(newPriority)
			}
		}
	}

	return nil
}

// CalculateBandwidthDistribution calculates how bandwidth should be distributed among child classes
func (ch *ClassHierarchy) CalculateBandwidthDistribution(parentHandle tc.Handle, parentRate tc.Bandwidth) (*BandwidthDistribution, error) {
	children := ch.GetChildren(parentHandle)
	if len(children) == 0 {
		return &BandwidthDistribution{
			TotalRate:             parentRate,
			AllocatedRate:         tc.MustParseBandwidth("0bps"),
			AvailableRate:         parentRate,
			ChildAllocations:      make(map[tc.Handle]tc.Bandwidth),
			OversubscriptionRatio: 0.0,
		}, nil
	}

	// Group children by priority
	priorityGroups := ch.groupChildrenByPriority(children)

	// Calculate total demand per priority group
	for i := range priorityGroups {
		group := &priorityGroups[i]
		totalDemand := tc.MustParseBandwidth("0bps")
		for _, childHandle := range group.Classes {
			if htbClass := ch.getHTBClass(childHandle); htbClass != nil {
				rate := htbClass.Rate()
				if rate.BitsPerSecond() > 0 {
					totalDemand = tc.MustParseBandwidth(fmt.Sprintf("%dbps",
						totalDemand.BitsPerSecond()+rate.BitsPerSecond()))
				}
			}
		}
		group.TotalDemand = totalDemand
	}

	// Allocate bandwidth by priority (higher priority first)
	allocations := make(map[tc.Handle]tc.Bandwidth)
	remainingRate := parentRate
	totalAllocated := tc.MustParseBandwidth("0bps")

	for _, group := range priorityGroups {
		if remainingRate.BitsPerSecond() <= 0 {
			break // No more bandwidth available
		}

		if group.TotalDemand.BitsPerSecond() == 0 {
			continue // No demand in this priority group
		}

		// Distribute available bandwidth among classes in this priority group
		availableForGroup := remainingRate
		if group.TotalDemand.BitsPerSecond() <= availableForGroup.BitsPerSecond() {
			// Enough bandwidth for full allocation
			for _, childHandle := range group.Classes {
				if htbClass := ch.getHTBClass(childHandle); htbClass != nil {
					rate := htbClass.Rate()
					if rate.BitsPerSecond() > 0 {
						allocations[childHandle] = rate
						totalAllocated = tc.MustParseBandwidth(fmt.Sprintf("%dbps",
							totalAllocated.BitsPerSecond()+rate.BitsPerSecond()))
						remainingRate = tc.MustParseBandwidth(fmt.Sprintf("%dbps",
							remainingRate.BitsPerSecond()-rate.BitsPerSecond()))
					}
				}
			}
		} else {
			// Not enough bandwidth - proportional allocation
			for _, childHandle := range group.Classes {
				if htbClass := ch.getHTBClass(childHandle); htbClass != nil {
					rate := htbClass.Rate()
					if rate.BitsPerSecond() > 0 && group.TotalDemand.BitsPerSecond() > 0 {
						proportion := float64(rate.BitsPerSecond()) / float64(group.TotalDemand.BitsPerSecond())
						allocated := uint64(float64(availableForGroup.BitsPerSecond()) * proportion)
						allocatedBandwidth := tc.MustParseBandwidth(fmt.Sprintf("%dbps", allocated))
						allocations[childHandle] = allocatedBandwidth
						totalAllocated = tc.MustParseBandwidth(fmt.Sprintf("%dbps",
							totalAllocated.BitsPerSecond()+allocated))
					}
				}
			}
			remainingRate = tc.MustParseBandwidth("0bps") // All available bandwidth used
		}
	}

	// Calculate oversubscription ratio
	oversubscriptionRatio := 0.0
	if parentRate.BitsPerSecond() > 0 {
		oversubscriptionRatio = float64(totalAllocated.BitsPerSecond()) / float64(parentRate.BitsPerSecond())
	}

	return &BandwidthDistribution{
		TotalRate:             parentRate,
		AllocatedRate:         totalAllocated,
		AvailableRate:         tc.MustParseBandwidth(fmt.Sprintf("%dbps", parentRate.BitsPerSecond()-totalAllocated.BitsPerSecond())),
		ChildAllocations:      allocations,
		OversubscriptionRatio: oversubscriptionRatio,
	}, nil
}

// groupChildrenByPriority groups child classes by their priority levels
func (ch *ClassHierarchy) groupChildrenByPriority(children []tc.Handle) []PriorityGroup {
	priorityMap := make(map[Priority][]tc.Handle)

	for _, childHandle := range children {
		child, exists := ch.classes[childHandle]
		if !exists || child.Priority() == nil {
			continue
		}

		priority := *child.Priority()
		priorityMap[priority] = append(priorityMap[priority], childHandle)
	}

	// Convert to sorted slice (priority 0 is highest, so sort ascending)
	var groups []PriorityGroup
	for priority := Priority(0); priority <= 7; priority++ {
		if classes, exists := priorityMap[priority]; exists {
			groups = append(groups, PriorityGroup{
				Priority: priority,
				Classes:  classes,
			})
		}
	}

	return groups
}

// ClassWithBandwidth represents a class that has bandwidth information
type ClassWithBandwidth interface {
	Rate() tc.Bandwidth
	Ceil() tc.Bandwidth
}

// htbClasses stores a mapping of handles to HTB class instances for bandwidth calculations
var htbClasses = make(map[tc.Handle]*HTBClass)

// RegisterHTBClass registers an HTB class instance for bandwidth calculations
func (ch *ClassHierarchy) RegisterHTBClass(handle tc.Handle, htbClass *HTBClass) {
	htbClasses[handle] = htbClass
}

// getHTBClass returns the HTB class if the handle points to an HTB class
func (ch *ClassHierarchy) getHTBClass(handle tc.Handle) *HTBClass {
	return htbClasses[handle]
}

// UnregisterHTBClass removes an HTB class from the bandwidth calculation registry
func (ch *ClassHierarchy) UnregisterHTBClass(handle tc.Handle) {
	delete(htbClasses, handle)
}

// ValidateBandwidthConstraints validates that bandwidth allocations are consistent across the hierarchy
func (ch *ClassHierarchy) ValidateBandwidthConstraints() error {
	for handle, class := range ch.classes {
		if htbClass := ch.getHTBClass(handle); htbClass != nil {
			// Validate that rate <= ceil
			if htbClass.Rate().BitsPerSecond() > htbClass.Ceil().BitsPerSecond() && htbClass.Ceil().BitsPerSecond() > 0 {
				return fmt.Errorf("class %s: rate (%s) exceeds ceil (%s)",
					handle, htbClass.Rate(), htbClass.Ceil())
			}

			// Validate against parent constraints
			if !class.Parent().IsRoot() {
				if parentHTB := ch.getHTBClass(class.Parent()); parentHTB != nil {
					// Child's ceil should not exceed parent's ceil
					if htbClass.Ceil().BitsPerSecond() > parentHTB.Ceil().BitsPerSecond() && parentHTB.Ceil().BitsPerSecond() > 0 {
						return fmt.Errorf("class %s: ceil (%s) exceeds parent ceil (%s)",
							handle, htbClass.Ceil(), parentHTB.Ceil())
					}
				}
			}
		}
	}

	return nil
}

// GetBandwidthUtilization calculates bandwidth utilization statistics for the hierarchy
func (ch *ClassHierarchy) GetBandwidthUtilization() map[tc.Handle]*BandwidthDistribution {
	utilization := make(map[tc.Handle]*BandwidthDistribution)

	for handle := range ch.classes {
		if htbClass := ch.getHTBClass(handle); htbClass != nil && htbClass.Rate().BitsPerSecond() > 0 {
			distribution, err := ch.CalculateBandwidthDistribution(handle, htbClass.Rate())
			if err == nil {
				utilization[handle] = distribution
			}
		}
	}

	return utilization
}

// ClassModifications represents the modifications that can be applied to a class
type ClassModifications struct {
	Name      *string    // New name for the class
	Priority  *Priority  // New priority for the class
	NewParent *tc.Handle // New parent handle (for moving the class)
}

// DeletionStrategy defines how to handle children when deleting a class
type DeletionStrategy int

const (
	DeleteCascade         DeletionStrategy = iota // Delete the class and all its descendants
	DeletePromoteChildren                         // Move children to the deleted class's parent
	DeleteOrphanChildren                          // Move children to become top-level classes
	DeleteFailIfChildren                          // Fail if the class has children
)

// PriorityInheritanceRule defines how child classes inherit priority from parents
type PriorityInheritanceRule int

const (
	InheritParentPriority PriorityInheritanceRule = iota // Child inherits exact parent priority
	InheritParentPlusOne                                 // Child inherits parent priority + 1 (lower priority)
	NoInheritance                                        // Child keeps its own priority
)

// BandwidthDistribution represents bandwidth allocation information for a class
type BandwidthDistribution struct {
	TotalRate             tc.Bandwidth               // Total rate available for distribution
	AllocatedRate         tc.Bandwidth               // Rate already allocated to children
	AvailableRate         tc.Bandwidth               // Remaining rate available
	ChildAllocations      map[tc.Handle]tc.Bandwidth // Rate allocated to each child
	OversubscriptionRatio float64                    // Ratio of allocated to available (> 1.0 means oversubscribed)
}

// PriorityGroup represents classes with the same priority level for bandwidth distribution
type PriorityGroup struct {
	Priority    Priority
	Classes     []tc.Handle
	TotalDemand tc.Bandwidth // Sum of all classes' requested rates in this priority group
}
// SetQuantum sets the quantum for borrowing
func (h *HTBClass) SetQuantum(quantum uint32) {
	h.quantum = quantum
}

// Quantum returns the quantum for borrowing
func (h *HTBClass) Quantum() uint32 {
	return h.quantum
}

// SetOverhead sets the packet overhead
func (h *HTBClass) SetOverhead(overhead uint32) {
	h.overhead = overhead
}

// Overhead returns the packet overhead
func (h *HTBClass) Overhead() uint32 {
	return h.overhead
}

// SetMPU sets the minimum packet unit
func (h *HTBClass) SetMPU(mpu uint32) {
	h.mpu = mpu
}

// MPU returns the minimum packet unit
func (h *HTBClass) MPU() uint32 {
	return h.mpu
}

// SetMTU sets the maximum transmission unit
func (h *HTBClass) SetMTU(mtu uint32) {
	h.mtu = mtu
}

// MTU returns the maximum transmission unit
func (h *HTBClass) MTU() uint32 {
	return h.mtu
}

// SetHTBPrio sets the internal HTB priority
func (h *HTBClass) SetHTBPrio(prio uint32) {
	h.prio = prio
}

// HTBPrio returns the internal HTB priority
func (h *HTBClass) HTBPrio() uint32 {
	return h.prio
}

// CalculateQuantum calculates appropriate quantum based on rate
func (h *HTBClass) CalculateQuantum() uint32 {
	// Quantum calculation: rate_bps / 8 / HZ
	// Standard Linux HZ is typically 1000, so quantum = rate_bytes_per_second / 1000
	// Minimum quantum is typically 1000 bytes, maximum is 60000 bytes
	const (
		MinQuantum = 1000  // Minimum quantum (1KB)
		MaxQuantum = 60000 // Maximum quantum (60KB)
		HZ         = 1000  // Linux timer frequency
	)

	if h.rate.BitsPerSecond() == 0 {
		return MinQuantum
	}

	bytesPerSecond := h.rate.BitsPerSecond() / 8
	// Prevent integer overflow in conversion
	quantumCalc := bytesPerSecond / HZ
	var quantum uint32
	if quantumCalc > 0xFFFFFFFF {
		quantum = 0xFFFFFFFF
	} else {
		quantum = uint32(quantumCalc) // #nosec G115 - bounds checked above
	}

	// Ensure quantum is within reasonable bounds
	if quantum < MinQuantum {
		return MinQuantum
	}
	if quantum > MaxQuantum {
		return MaxQuantum
	}

	return quantum
}

// CalculateEnhancedBurst calculates burst with MTU and overhead considerations
func (h *HTBClass) CalculateEnhancedBurst() uint32 {
	// Enhanced burst calculation considering MTU, overhead, and timer resolution
	const TimerResolutionMS = 64 // Linux timer resolution in milliseconds

	if h.rate.BitsPerSecond() == 0 {
		return 1600 // Default minimum burst
	}

	// Calculate burst for timer resolution period
	bytesPerSecond := h.rate.BitsPerSecond() / 8
	// Prevent integer overflow in conversion
	burstCalc := bytesPerSecond * TimerResolutionMS / 1000
	var burstBytes uint32
	if burstCalc > 0xFFFFFFFF {
		burstBytes = 0xFFFFFFFF
	} else {
		burstBytes = uint32(burstCalc) // #nosec G115 - bounds checked above
	}

	// Add overhead consideration
	if h.overhead > 0 {
		// Assume average packet size for overhead calculation
		avgPacketSize := uint32(1500) // Standard Ethernet MTU
		if h.mtu > 0 {
			avgPacketSize = h.mtu
		}

		packetsPerBurst := burstBytes / avgPacketSize
		if packetsPerBurst == 0 {
			packetsPerBurst = 1
		}

		overheadTotal := packetsPerBurst * h.overhead
		burstBytes += overheadTotal
	}

	// Ensure minimum burst considering MTU
	minBurst := uint32(1600) // Default minimum
	if h.mtu > 0 {
		minBurst = h.mtu * 2 // At least 2 MTU-sized packets
	}

	if burstBytes < minBurst {
		burstBytes = minBurst
	}

	// burstBytes is already uint32, so no need to cap (would be caught at conversion)

	return burstBytes
}

// CalculateEnhancedCburst calculates cburst with advanced parameters
func (h *HTBClass) CalculateEnhancedCburst() uint32 {
	// Use ceil if set, otherwise use rate
	bandwidth := h.ceil
	if bandwidth.BitsPerSecond() == 0 {
		bandwidth = h.rate
	}

	// Temporarily store original rate to calculate cburst with ceil
	originalRate := h.rate
	h.rate = bandwidth

	cburst := h.CalculateEnhancedBurst()

	// Restore original rate
	h.rate = originalRate

	return cburst
}

// ApplyDefaultParameters applies sensible defaults for HTB parameters
func (h *HTBClass) ApplyDefaultParameters() {
	// Set quantum if not already set
	if h.quantum == 0 {
		h.quantum = h.CalculateQuantum()
	}

	// Set default MTU if not set
	if h.mtu == 0 {
		h.mtu = 1500 // Standard Ethernet MTU
	}

	// Set default MPU if not set
	if h.mpu == 0 {
		h.mpu = 64 // Minimum Ethernet frame payload
	}

	// Set default overhead if not set
	if h.overhead == 0 {
		h.overhead = 4 // Basic Ethernet overhead estimate
	}

	// Calculate burst and cburst using enhanced algorithms
	if h.burst == 0 {
		h.burst = h.CalculateEnhancedBurst()
	}

	if h.cburst == 0 {
		h.cburst = h.CalculateEnhancedCburst()
	}
>>>>>>> main
}
