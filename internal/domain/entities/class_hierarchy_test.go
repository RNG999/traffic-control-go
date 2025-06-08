package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/pkg/tc"
)

func TestClassHierarchy_AddClass(t *testing.T) {
	tests := []struct {
		name      string
		maxDepth  int
		classes   []*Class
		expectErr bool
		errMsg    string
	}{
		{
			name:     "add root class",
			maxDepth: 3,
			classes: []*Class{
				createTestClass("1:1", "1:0", "root", 0),
			},
			expectErr: false,
		},
		{
			name:     "add child class",
			maxDepth: 3,
			classes: []*Class{
				createTestClass("1:1", "1:0", "root", 0),
				createTestClass("1:10", "1:1", "child", 1),
			},
			expectErr: false,
		},
		{
			name:     "exceed max depth",
			maxDepth: 1,
			classes: []*Class{
				createTestClass("1:1", "1:0", "root", 0),
				createTestClass("1:10", "1:1", "child1", 1),
				createTestClass("1:20", "1:10", "child2", 2), // Should fail - exceeds maxDepth of 1
			},
			expectErr: true,
			errMsg:    "adding class would exceed maximum depth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := NewClassHierarchy(tt.maxDepth)

			var err error
			for i, class := range tt.classes {
				err = ch.AddClass(class)
				if tt.expectErr && i == len(tt.classes)-1 {
					// Expect error on last class
					break
				}
				require.NoError(t, err, "Failed to add class %d", i)
			}

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClassHierarchy_CircularDependency(t *testing.T) {
	ch := NewClassHierarchy(5)

	// Create a linear hierarchy: 1:1 -> 1:10 -> 1:20
	root := createTestClass("1:1", "1:0", "root", 0)
	child1 := createTestClass("1:10", "1:1", "child1", 1)
	child2 := createTestClass("1:20", "1:10", "child2", 2)

	require.NoError(t, ch.AddClass(root))
	require.NoError(t, ch.AddClass(child1))
	require.NoError(t, ch.AddClass(child2))

	// Try to create a cycle: make 1:1 (root) a child of 1:20
	cyclicClass := createTestClass("1:1", "1:20", "cyclic", 3)
	err := ch.AddClass(cyclicClass)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestClassHierarchy_RemoveClass(t *testing.T) {
	ch := NewClassHierarchy(5)

	// Create hierarchy: 1:1 -> 1:10 -> [1:20, 1:21]
	root := createTestClass("1:1", "1:0", "root", 0)
	child1 := createTestClass("1:10", "1:1", "child1", 1)
	child2 := createTestClass("1:20", "1:10", "child2", 2)
	child3 := createTestClass("1:21", "1:10", "child3", 2)

	require.NoError(t, ch.AddClass(root))
	require.NoError(t, ch.AddClass(child1))
	require.NoError(t, ch.AddClass(child2))
	require.NoError(t, ch.AddClass(child3))

	// Verify initial state
	children := ch.GetChildren(tc.MustParseHandle("1:10"))
	assert.Len(t, children, 2)

	// Remove child1 (should remove all its descendants)
	err := ch.RemoveClass(tc.MustParseHandle("1:10"))
	require.NoError(t, err)

	// Verify removal
	children = ch.GetChildren(tc.MustParseHandle("1:1"))
	assert.Len(t, children, 0)

	children = ch.GetChildren(tc.MustParseHandle("1:10"))
	assert.Len(t, children, 0)
}

func TestClassHierarchy_CalculateDepth(t *testing.T) {
	ch := NewClassHierarchy(5)

	// Create hierarchy: 1:1 -> 1:10 -> 1:20
	root := createTestClass("1:1", "1:0", "root", 0)
	child1 := createTestClass("1:10", "1:1", "child1", 1)
	child2 := createTestClass("1:20", "1:10", "child2", 2)

	require.NoError(t, ch.AddClass(root))
	require.NoError(t, ch.AddClass(child1))
	require.NoError(t, ch.AddClass(child2))

	tests := []struct {
		handle        string
		expectedDepth int
	}{
		{"1:1", 0},
		{"1:10", 1},
		{"1:20", 2},
	}

	for _, tt := range tests {
		t.Run(tt.handle, func(t *testing.T) {
			depth, err := ch.CalculateDepth(tc.MustParseHandle(tt.handle))
			require.NoError(t, err)
			assert.Equal(t, tt.expectedDepth, depth)
		})
	}
}

func TestClassHierarchy_GetDescendants(t *testing.T) {
	ch := NewClassHierarchy(5)

	// Create hierarchy: 1:1 -> 1:10 -> [1:20, 1:21] and 1:1 -> 1:11
	root := createTestClass("1:1", "1:0", "root", 0)
	child1 := createTestClass("1:10", "1:1", "child1", 1)
	child2 := createTestClass("1:11", "1:1", "child2", 1)
	grandchild1 := createTestClass("1:20", "1:10", "grandchild1", 2)
	grandchild2 := createTestClass("1:21", "1:10", "grandchild2", 2)

	require.NoError(t, ch.AddClass(root))
	require.NoError(t, ch.AddClass(child1))
	require.NoError(t, ch.AddClass(child2))
	require.NoError(t, ch.AddClass(grandchild1))
	require.NoError(t, ch.AddClass(grandchild2))

	// Get all descendants of root
	descendants := ch.GetDescendants(tc.MustParseHandle("1:1"))
	assert.Len(t, descendants, 4) // 1:10, 1:11, 1:20, 1:21

	// Get descendants of 1:10
	descendants = ch.GetDescendants(tc.MustParseHandle("1:10"))
	assert.Len(t, descendants, 2) // 1:20, 1:21

	// Get descendants of leaf node
	descendants = ch.GetDescendants(tc.MustParseHandle("1:20"))
	assert.Len(t, descendants, 0)
}

func TestClassHierarchy_GetAncestors(t *testing.T) {
	ch := NewClassHierarchy(5)

	// Create hierarchy: 1:1 -> 1:10 -> 1:20
	root := createTestClass("1:1", "1:0", "root", 0)
	child1 := createTestClass("1:10", "1:1", "child1", 1)
	child2 := createTestClass("1:20", "1:10", "child2", 2)

	require.NoError(t, ch.AddClass(root))
	require.NoError(t, ch.AddClass(child1))
	require.NoError(t, ch.AddClass(child2))

	// Get ancestors of deepest node
	ancestors := ch.GetAncestors(tc.MustParseHandle("1:20"))
	assert.Len(t, ancestors, 2) // 1:10, 1:1
	assert.Equal(t, "1:10", ancestors[0].String())
	assert.Equal(t, "1:1", ancestors[1].String())

	// Get ancestors of middle node
	ancestors = ch.GetAncestors(tc.MustParseHandle("1:10"))
	assert.Len(t, ancestors, 1) // 1:1
	assert.Equal(t, "1:1", ancestors[0].String())

	// Get ancestors of root
	ancestors = ch.GetAncestors(tc.MustParseHandle("1:1"))
	assert.Len(t, ancestors, 0)
}

func TestClassHierarchy_ValidateHierarchy(t *testing.T) {
	ch := NewClassHierarchy(5)

	// Create valid hierarchy
	root := createTestClass("1:1", "1:0", "root", 0)
	child1 := createTestClass("1:10", "1:1", "child1", 1)
	child2 := createTestClass("1:20", "1:10", "child2", 2)

	require.NoError(t, ch.AddClass(root))
	require.NoError(t, ch.AddClass(child1))
	require.NoError(t, ch.AddClass(child2))

	// Validation should pass
	err := ch.ValidateHierarchy()
	assert.NoError(t, err)
}

func TestClass_HierarchyMethods(t *testing.T) {
	device, _ := tc.NewDeviceName("eth0")
	handle := tc.MustParseHandle("1:10")
	parent := tc.MustParseHandle("1:1")

	class := NewClass(device, handle, parent, "test-class", Priority(1))

	// Test initial state
	assert.Equal(t, 0, class.Depth())
	assert.False(t, class.HasChildren())
	assert.True(t, class.IsLeaf())
	assert.Len(t, class.Children(), 0)

	// Add children
	child1 := tc.MustParseHandle("1:20")
	child2 := tc.MustParseHandle("1:21")

	class.AddChild(child1)
	class.AddChild(child2)

	assert.True(t, class.HasChildren())
	assert.False(t, class.IsLeaf())
	assert.Len(t, class.Children(), 2)

	// Remove a child
	class.RemoveChild(child1)
	assert.Len(t, class.Children(), 1)
	assert.Equal(t, child2.String(), class.Children()[0].String())

	// Set depth
	class.SetDepth(2)
	assert.Equal(t, 2, class.Depth())
}

// Helper function to create test classes
func createTestClass(handle, parent, name string, priority int) *Class {
	device, _ := tc.NewDeviceName("eth0")
	h := tc.MustParseHandle(handle)
	p := tc.MustParseHandle(parent)
	return NewClass(device, h, p, name, Priority(priority))
}
