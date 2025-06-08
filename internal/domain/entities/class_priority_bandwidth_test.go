package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/pkg/tc"
)

func TestClassHierarchy_ApplyPriorityInheritance(t *testing.T) {
	tests := []struct {
		name       string
		rule       PriorityInheritanceRule
		setupFn    func() *ClassHierarchy
		validateFn func(t *testing.T, ch *ClassHierarchy)
	}{
		{
			name: "inherit parent priority exactly",
			rule: InheritParentPriority,
			setupFn: func() *ClassHierarchy {
				ch := NewClassHierarchy(5)
				root := createTestClass("1:1", "1:0", "root", 2)
				child1 := createTestClass("1:10", "1:1", "child1", 5)
				child2 := createTestClass("1:20", "1:10", "grandchild", 7)

				require.NoError(t, ch.AddClass(root))
				require.NoError(t, ch.AddClass(child1))
				require.NoError(t, ch.AddClass(child2))
				return ch
			},
			validateFn: func(t *testing.T, ch *ClassHierarchy) {
				// Child should inherit root's priority (2)
				child1 := ch.classes[tc.MustParseHandle("1:10")]
				assert.Equal(t, Priority(2), *child1.Priority())

				// Grandchild should inherit child's inherited priority (2)
				child2 := ch.classes[tc.MustParseHandle("1:20")]
				assert.Equal(t, Priority(2), *child2.Priority())
			},
		},
		{
			name: "inherit parent priority plus one",
			rule: InheritParentPlusOne,
			setupFn: func() *ClassHierarchy {
				ch := NewClassHierarchy(5)
				root := createTestClass("1:1", "1:0", "root", 1)
				child1 := createTestClass("1:10", "1:1", "child1", 5)
				child2 := createTestClass("1:20", "1:10", "grandchild", 7)

				require.NoError(t, ch.AddClass(root))
				require.NoError(t, ch.AddClass(child1))
				require.NoError(t, ch.AddClass(child2))
				return ch
			},
			validateFn: func(t *testing.T, ch *ClassHierarchy) {
				// Child should inherit root's priority + 1 (1 + 1 = 2)
				child1 := ch.classes[tc.MustParseHandle("1:10")]
				assert.Equal(t, Priority(2), *child1.Priority())

				// Grandchild should inherit child's inherited priority + 1 (2 + 1 = 3)
				child2 := ch.classes[tc.MustParseHandle("1:20")]
				assert.Equal(t, Priority(3), *child2.Priority())
			},
		},
		{
			name: "priority capped at 7",
			rule: InheritParentPlusOne,
			setupFn: func() *ClassHierarchy {
				ch := NewClassHierarchy(5)
				root := createTestClass("1:1", "1:0", "root", 7)
				child1 := createTestClass("1:10", "1:1", "child1", 3)

				require.NoError(t, ch.AddClass(root))
				require.NoError(t, ch.AddClass(child1))
				return ch
			},
			validateFn: func(t *testing.T, ch *ClassHierarchy) {
				// Child should inherit capped priority (7 + 1 = 8, capped to 7)
				child1 := ch.classes[tc.MustParseHandle("1:10")]
				assert.Equal(t, Priority(7), *child1.Priority())
			},
		},
		{
			name: "no inheritance keeps original priorities",
			rule: NoInheritance,
			setupFn: func() *ClassHierarchy {
				ch := NewClassHierarchy(5)
				root := createTestClass("1:1", "1:0", "root", 1)
				child1 := createTestClass("1:10", "1:1", "child1", 5)

				require.NoError(t, ch.AddClass(root))
				require.NoError(t, ch.AddClass(child1))
				return ch
			},
			validateFn: func(t *testing.T, ch *ClassHierarchy) {
				// Child should keep its original priority
				child1 := ch.classes[tc.MustParseHandle("1:10")]
				assert.Equal(t, Priority(5), *child1.Priority())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := tt.setupFn()
			err := ch.ApplyPriorityInheritance(tt.rule)
			require.NoError(t, err)
			tt.validateFn(t, ch)
		})
	}
}

func TestClassHierarchy_CalculateBandwidthDistribution(t *testing.T) {
	tests := []struct {
		name       string
		setupFn    func() (*ClassHierarchy, tc.Handle, tc.Bandwidth)
		expectErr  bool
		validateFn func(t *testing.T, dist *BandwidthDistribution)
	}{
		{
			name: "no children - all bandwidth available",
			setupFn: func() (*ClassHierarchy, tc.Handle, tc.Bandwidth) {
				ch := NewClassHierarchy(5)
				root := createTestClass("1:1", "1:0", "root", 0)
				require.NoError(t, ch.AddClass(root))

				parentRate := tc.MustParseBandwidth("1000000bps") // 1 Mbps
				return ch, tc.MustParseHandle("1:1"), parentRate
			},
			validateFn: func(t *testing.T, dist *BandwidthDistribution) {
				assert.Equal(t, uint64(1000000), dist.TotalRate.BitsPerSecond())
				assert.Equal(t, uint64(0), dist.AllocatedRate.BitsPerSecond())
				assert.Equal(t, uint64(1000000), dist.AvailableRate.BitsPerSecond())
				assert.Equal(t, 0.0, dist.OversubscriptionRatio)
				assert.Len(t, dist.ChildAllocations, 0)
			},
		},
		{
			name: "sufficient bandwidth for all children",
			setupFn: func() (*ClassHierarchy, tc.Handle, tc.Bandwidth) {
				ch := NewClassHierarchy(5)

				// Create HTB classes with bandwidth settings
				device, _ := tc.NewDeviceName("eth0")

				// Root class with 1 Mbps
				root := NewHTBClass(device, tc.MustParseHandle("1:1"), tc.MustParseHandle("1:0"), "root", Priority(0))
				root.SetRate(tc.MustParseBandwidth("1000000bps"))
				root.SetCeil(tc.MustParseBandwidth("1000000bps"))

				// Child classes requesting less than available
				child1 := NewHTBClass(device, tc.MustParseHandle("1:10"), tc.MustParseHandle("1:1"), "child1", Priority(1))
				child1.SetRate(tc.MustParseBandwidth("300000bps")) // 300 Kbps

				child2 := NewHTBClass(device, tc.MustParseHandle("1:20"), tc.MustParseHandle("1:1"), "child2", Priority(1))
				child2.SetRate(tc.MustParseBandwidth("400000bps")) // 400 Kbps

				// Add classes to hierarchy and register HTB classes
				require.NoError(t, ch.AddClass(root.Class))
				require.NoError(t, ch.AddClass(child1.Class))
				require.NoError(t, ch.AddClass(child2.Class))

				ch.RegisterHTBClass(tc.MustParseHandle("1:1"), root)
				ch.RegisterHTBClass(tc.MustParseHandle("1:10"), child1)
				ch.RegisterHTBClass(tc.MustParseHandle("1:20"), child2)

				parentRate := tc.MustParseBandwidth("1000000bps")
				return ch, tc.MustParseHandle("1:1"), parentRate
			},
			validateFn: func(t *testing.T, dist *BandwidthDistribution) {
				assert.Equal(t, uint64(1000000), dist.TotalRate.BitsPerSecond())
				assert.Equal(t, uint64(700000), dist.AllocatedRate.BitsPerSecond()) // 300k + 400k
				assert.Equal(t, uint64(300000), dist.AvailableRate.BitsPerSecond()) // 1000k - 700k
				assert.Equal(t, 0.7, dist.OversubscriptionRatio)
				assert.Len(t, dist.ChildAllocations, 2)
			},
		},
		{
			name: "priority-based allocation",
			setupFn: func() (*ClassHierarchy, tc.Handle, tc.Bandwidth) {
				ch := NewClassHierarchy(5)
				device, _ := tc.NewDeviceName("eth0")

				// Root class with limited bandwidth
				root := NewHTBClass(device, tc.MustParseHandle("1:1"), tc.MustParseHandle("1:0"), "root", Priority(0))
				root.SetRate(tc.MustParseBandwidth("500000bps")) // 500 Kbps

				// High priority child
				child1 := NewHTBClass(device, tc.MustParseHandle("1:10"), tc.MustParseHandle("1:1"), "high-priority", Priority(0))
				child1.SetRate(tc.MustParseBandwidth("300000bps"))

				// Low priority child
				child2 := NewHTBClass(device, tc.MustParseHandle("1:20"), tc.MustParseHandle("1:1"), "low-priority", Priority(2))
				child2.SetRate(tc.MustParseBandwidth("400000bps"))

				require.NoError(t, ch.AddClass(root.Class))
				require.NoError(t, ch.AddClass(child1.Class))
				require.NoError(t, ch.AddClass(child2.Class))

				ch.RegisterHTBClass(tc.MustParseHandle("1:1"), root)
				ch.RegisterHTBClass(tc.MustParseHandle("1:10"), child1)
				ch.RegisterHTBClass(tc.MustParseHandle("1:20"), child2)

				parentRate := tc.MustParseBandwidth("500000bps")
				return ch, tc.MustParseHandle("1:1"), parentRate
			},
			validateFn: func(t *testing.T, dist *BandwidthDistribution) {
				// High priority should get full allocation first
				assert.Contains(t, dist.ChildAllocations, tc.MustParseHandle("1:10"))
				assert.Contains(t, dist.ChildAllocations, tc.MustParseHandle("1:20"))

				// Check that high priority got its full allocation
				highPriorityAlloc := dist.ChildAllocations[tc.MustParseHandle("1:10")]
				assert.Equal(t, uint64(300000), highPriorityAlloc.BitsPerSecond())

				// Low priority should get remaining bandwidth
				lowPriorityAlloc := dist.ChildAllocations[tc.MustParseHandle("1:20")]
				assert.Equal(t, uint64(200000), lowPriorityAlloc.BitsPerSecond()) // 500k - 300k
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch, handle, parentRate := tt.setupFn()

			dist, err := ch.CalculateBandwidthDistribution(handle, parentRate)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.validateFn(t, dist)
			}
		})
	}
}

func TestClassHierarchy_ValidateBandwidthConstraints(t *testing.T) {
	tests := []struct {
		name      string
		setupFn   func() *ClassHierarchy
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid constraints",
			setupFn: func() *ClassHierarchy {
				ch := NewClassHierarchy(5)
				device, _ := tc.NewDeviceName("eth0")

				root := NewHTBClass(device, tc.MustParseHandle("1:1"), tc.MustParseHandle("1:0"), "root", Priority(0))
				root.SetRate(tc.MustParseBandwidth("500000bps"))
				root.SetCeil(tc.MustParseBandwidth("1000000bps"))

				child := NewHTBClass(device, tc.MustParseHandle("1:10"), tc.MustParseHandle("1:1"), "child", Priority(1))
				child.SetRate(tc.MustParseBandwidth("300000bps"))
				child.SetCeil(tc.MustParseBandwidth("800000bps"))

				require.NoError(t, ch.AddClass(root.Class))
				require.NoError(t, ch.AddClass(child.Class))

				ch.RegisterHTBClass(tc.MustParseHandle("1:1"), root)
				ch.RegisterHTBClass(tc.MustParseHandle("1:10"), child)

				return ch
			},
			expectErr: false,
		},
		{
			name: "rate exceeds ceil",
			setupFn: func() *ClassHierarchy {
				ch := NewClassHierarchy(5)
				device, _ := tc.NewDeviceName("eth0")

				root := NewHTBClass(device, tc.MustParseHandle("1:1"), tc.MustParseHandle("1:0"), "root", Priority(0))
				root.SetRate(tc.MustParseBandwidth("1000000bps")) // Rate higher than ceil
				root.SetCeil(tc.MustParseBandwidth("500000bps"))

				require.NoError(t, ch.AddClass(root.Class))
				ch.RegisterHTBClass(tc.MustParseHandle("1:1"), root)

				return ch
			},
			expectErr: true,
			errMsg:    "rate (1.0Mbps) exceeds ceil (500.0Kbps)",
		},
		{
			name: "child ceil exceeds parent ceil",
			setupFn: func() *ClassHierarchy {
				ch := NewClassHierarchy(5)
				device, _ := tc.NewDeviceName("eth0")

				root := NewHTBClass(device, tc.MustParseHandle("1:1"), tc.MustParseHandle("1:0"), "root", Priority(0))
				root.SetRate(tc.MustParseBandwidth("500000bps"))
				root.SetCeil(tc.MustParseBandwidth("800000bps"))

				child := NewHTBClass(device, tc.MustParseHandle("1:10"), tc.MustParseHandle("1:1"), "child", Priority(1))
				child.SetRate(tc.MustParseBandwidth("300000bps"))
				child.SetCeil(tc.MustParseBandwidth("1000000bps")) // Exceeds parent ceil

				require.NoError(t, ch.AddClass(root.Class))
				require.NoError(t, ch.AddClass(child.Class))

				ch.RegisterHTBClass(tc.MustParseHandle("1:1"), root)
				ch.RegisterHTBClass(tc.MustParseHandle("1:10"), child)

				return ch
			},
			expectErr: true,
			errMsg:    "ceil (1.0Mbps) exceeds parent ceil (800.0Kbps)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := tt.setupFn()
			err := ch.ValidateBandwidthConstraints()

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClassHierarchy_GetBandwidthUtilization(t *testing.T) {
	ch := NewClassHierarchy(5)
	device, _ := tc.NewDeviceName("eth0")

	// Create a hierarchy with bandwidth settings
	root := NewHTBClass(device, tc.MustParseHandle("1:1"), tc.MustParseHandle("1:0"), "root", Priority(0))
	root.SetRate(tc.MustParseBandwidth("1000000bps"))

	child1 := NewHTBClass(device, tc.MustParseHandle("1:10"), tc.MustParseHandle("1:1"), "child1", Priority(1))
	child1.SetRate(tc.MustParseBandwidth("400000bps"))

	child2 := NewHTBClass(device, tc.MustParseHandle("1:20"), tc.MustParseHandle("1:1"), "child2", Priority(1))
	child2.SetRate(tc.MustParseBandwidth("300000bps"))

	require.NoError(t, ch.AddClass(root.Class))
	require.NoError(t, ch.AddClass(child1.Class))
	require.NoError(t, ch.AddClass(child2.Class))

	ch.RegisterHTBClass(tc.MustParseHandle("1:1"), root)
	ch.RegisterHTBClass(tc.MustParseHandle("1:10"), child1)
	ch.RegisterHTBClass(tc.MustParseHandle("1:20"), child2)

	utilization := ch.GetBandwidthUtilization()

	// Should have utilization for classes with bandwidth set
	assert.Contains(t, utilization, tc.MustParseHandle("1:1"))
	assert.Contains(t, utilization, tc.MustParseHandle("1:10"))
	assert.Contains(t, utilization, tc.MustParseHandle("1:20"))

	// Root should show children allocation
	rootUtil := utilization[tc.MustParseHandle("1:1")]
	assert.Equal(t, uint64(1000000), rootUtil.TotalRate.BitsPerSecond())
	assert.Len(t, rootUtil.ChildAllocations, 2)
}

func TestPriorityGroup_Sorting(t *testing.T) {
	ch := NewClassHierarchy(5)

	// Create classes with different priorities
	class0 := createTestClass("1:10", "1:1", "high", 0) // Highest priority
	class1 := createTestClass("1:20", "1:1", "medium", 1)
	class7 := createTestClass("1:30", "1:1", "low", 7) // Lowest priority
	class2 := createTestClass("1:40", "1:1", "medium2", 2)

	children := []tc.Handle{
		tc.MustParseHandle("1:10"),
		tc.MustParseHandle("1:20"),
		tc.MustParseHandle("1:30"),
		tc.MustParseHandle("1:40"),
	}

	ch.classes = map[tc.Handle]*Class{
		tc.MustParseHandle("1:10"): class0,
		tc.MustParseHandle("1:20"): class1,
		tc.MustParseHandle("1:30"): class7,
		tc.MustParseHandle("1:40"): class2,
	}

	groups := ch.groupChildrenByPriority(children)

	// Should be sorted by priority (0 first, then 1, 2, 7)
	require.Len(t, groups, 4)
	assert.Equal(t, Priority(0), groups[0].Priority)
	assert.Equal(t, Priority(1), groups[1].Priority)
	assert.Equal(t, Priority(2), groups[2].Priority)
	assert.Equal(t, Priority(7), groups[3].Priority)

	// Verify class grouping
	assert.Contains(t, groups[0].Classes, tc.MustParseHandle("1:10"))
	assert.Contains(t, groups[1].Classes, tc.MustParseHandle("1:20"))
	assert.Contains(t, groups[2].Classes, tc.MustParseHandle("1:40"))
	assert.Contains(t, groups[3].Classes, tc.MustParseHandle("1:30"))
}
