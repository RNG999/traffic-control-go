package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/pkg/tc"
)

func TestClass_SetName(t *testing.T) {
	device, _ := tc.NewDeviceName("eth0")
	class := NewClass(device, tc.MustParseHandle("1:10"), tc.MustParseHandle("1:0"), "original", Priority(1))

	// Test valid name change
	err := class.SetName("new-name")
	assert.NoError(t, err)
	assert.Equal(t, "new-name", class.Name())

	// Test empty name (should fail)
	err = class.SetName("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
	// Original name should be preserved
	assert.Equal(t, "new-name", class.Name())
}

func TestClassHierarchy_ModifyClass(t *testing.T) {
	ch := NewClassHierarchy(5)

	// Create test hierarchy: 1:1 -> 1:10 -> 1:20
	root := createTestClass("1:1", "1:0", "root", 0)
	child1 := createTestClass("1:10", "1:1", "child1", 1)
	child2 := createTestClass("1:20", "1:10", "child2", 2)

	require.NoError(t, ch.AddClass(root))
	require.NoError(t, ch.AddClass(child1))
	require.NoError(t, ch.AddClass(child2))

	tests := []struct {
		name          string
		handle        string
		modifications ClassModifications
		expectErr     bool
		validateFn    func(t *testing.T, ch *ClassHierarchy)
	}{
		{
			name:   "modify name only",
			handle: "1:10",
			modifications: ClassModifications{
				Name: stringPtr("new-child1-name"),
			},
			expectErr: false,
			validateFn: func(t *testing.T, ch *ClassHierarchy) {
				class := ch.classes[tc.MustParseHandle("1:10")]
				assert.Equal(t, "new-child1-name", class.Name())
			},
		},
		{
			name:   "modify priority only",
			handle: "1:20",
			modifications: ClassModifications{
				Priority: priorityPtr(Priority(5)),
			},
			expectErr: false,
			validateFn: func(t *testing.T, ch *ClassHierarchy) {
				class := ch.classes[tc.MustParseHandle("1:20")]
				assert.Equal(t, Priority(5), *class.Priority())
			},
		},
		{
			name:   "move class to new parent",
			handle: "1:20",
			modifications: ClassModifications{
				NewParent: handlePtr(tc.MustParseHandle("1:1")), // Move from 1:10 to 1:1
			},
			expectErr: false,
			validateFn: func(t *testing.T, ch *ClassHierarchy) {
				class := ch.classes[tc.MustParseHandle("1:20")]
				assert.Equal(t, "1:1", class.Parent().String())
				assert.Equal(t, 1, class.Depth()) // Should be at depth 1 (1:1 is at depth 0, so child is at depth 1)

				// Verify parent-child relationships
				children := ch.GetChildren(tc.MustParseHandle("1:1"))
				assert.Contains(t, handleStrings(children), "1:20")

				children = ch.GetChildren(tc.MustParseHandle("1:10"))
				assert.NotContains(t, handleStrings(children), "1:20")
			},
		},
		{
			name:   "modify multiple properties",
			handle: "1:10",
			modifications: ClassModifications{
				Name:     stringPtr("renamed-child"),
				Priority: priorityPtr(Priority(3)),
			},
			expectErr: false,
			validateFn: func(t *testing.T, ch *ClassHierarchy) {
				class := ch.classes[tc.MustParseHandle("1:10")]
				assert.Equal(t, "renamed-child", class.Name())
				assert.Equal(t, Priority(3), *class.Priority())
			},
		},
		{
			name:   "modify non-existent class",
			handle: "1:99",
			modifications: ClassModifications{
				Name: stringPtr("new-name"),
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ch.ModifyClass(tc.MustParseHandle(tt.handle), tt.modifications)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validateFn != nil {
					tt.validateFn(t, ch)
				}
			}
		})
	}
}

func TestClassHierarchy_MoveClass(t *testing.T) {
	ch := NewClassHierarchy(3)

	// Create test hierarchy: 1:1 -> [1:10, 1:11] and 1:10 -> 1:20
	root := createTestClass("1:1", "1:0", "root", 0)
	child1 := createTestClass("1:10", "1:1", "child1", 1)
	child2 := createTestClass("1:11", "1:1", "child2", 1)
	grandchild := createTestClass("1:20", "1:10", "grandchild", 2)

	require.NoError(t, ch.AddClass(root))
	require.NoError(t, ch.AddClass(child1))
	require.NoError(t, ch.AddClass(child2))
	require.NoError(t, ch.AddClass(grandchild))

	tests := []struct {
		name       string
		handle     string
		newParent  string
		expectErr  bool
		errMsg     string
		validateFn func(t *testing.T, ch *ClassHierarchy)
	}{
		{
			name:      "move child to sibling",
			handle:    "1:20",
			newParent: "1:11",
			expectErr: false,
			validateFn: func(t *testing.T, ch *ClassHierarchy) {
				class := ch.classes[tc.MustParseHandle("1:20")]
				assert.Equal(t, "1:11", class.Parent().String())
				assert.Equal(t, 2, class.Depth()) // 1:11 is at depth 1, so child is at depth 2
			},
		},
		{
			name:      "move to root",
			handle:    "1:10",
			newParent: "1:", // Root handle representation
			expectErr: false,
			validateFn: func(t *testing.T, ch *ClassHierarchy) {
				class := ch.classes[tc.MustParseHandle("1:10")]
				assert.Equal(t, "1:", class.Parent().String())
				assert.Equal(t, 0, class.Depth())
			},
		},
		{
			name:      "circular dependency prevention",
			handle:    "1:1",
			newParent: "1:20", // Try to make root a child of its great-grandchild
			expectErr: true,
			errMsg:    "circular dependency",
		},
		{
			name:      "move to non-existent parent",
			handle:    "1:20",
			newParent: "1:99",
			expectErr: true,
			errMsg:    "not found in hierarchy",
		},
		{
			name:      "prevent circular dependency after previous moves",
			handle:    "1:11",
			newParent: "1:20", // This should fail because in previous test 1:20 was moved under 1:11
			expectErr: true,
			errMsg:    "circular dependency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ch.MoveClass(tc.MustParseHandle(tt.handle), tc.MustParseHandle(tt.newParent))

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				if tt.validateFn != nil {
					tt.validateFn(t, ch)
				}
			}
		})
	}
}

func TestClassHierarchy_DeleteClass(t *testing.T) {
	tests := []struct {
		name        string
		strategy    DeletionStrategy
		setupFn     func() *ClassHierarchy
		deleteClass string
		expectErr   bool
		errMsg      string
		validateFn  func(t *testing.T, ch *ClassHierarchy)
	}{
		{
			name:     "delete cascade - removes all descendants",
			strategy: DeleteCascade,
			setupFn: func() *ClassHierarchy {
				ch := NewClassHierarchy(5)
				// Create: 1:1 -> 1:10 -> [1:20, 1:21]
				require.NoError(t, ch.AddClass(createTestClass("1:1", "1:0", "root", 0)))
				require.NoError(t, ch.AddClass(createTestClass("1:10", "1:1", "child", 1)))
				require.NoError(t, ch.AddClass(createTestClass("1:20", "1:10", "grandchild1", 2)))
				require.NoError(t, ch.AddClass(createTestClass("1:21", "1:10", "grandchild2", 2)))
				return ch
			},
			deleteClass: "1:10",
			expectErr:   false,
			validateFn: func(t *testing.T, ch *ClassHierarchy) {
				// 1:10 and all its descendants should be gone
				assert.Nil(t, ch.classes[tc.MustParseHandle("1:10")])
				assert.Nil(t, ch.classes[tc.MustParseHandle("1:20")])
				assert.Nil(t, ch.classes[tc.MustParseHandle("1:21")])
				// Root should still exist
				assert.NotNil(t, ch.classes[tc.MustParseHandle("1:1")])
			},
		},
		{
			name:     "delete promote children - moves children to parent",
			strategy: DeletePromoteChildren,
			setupFn: func() *ClassHierarchy {
				ch := NewClassHierarchy(5)
				require.NoError(t, ch.AddClass(createTestClass("1:1", "1:0", "root", 0)))
				require.NoError(t, ch.AddClass(createTestClass("1:10", "1:1", "child", 1)))
				require.NoError(t, ch.AddClass(createTestClass("1:20", "1:10", "grandchild1", 2)))
				require.NoError(t, ch.AddClass(createTestClass("1:21", "1:10", "grandchild2", 2)))
				return ch
			},
			deleteClass: "1:10",
			expectErr:   false,
			validateFn: func(t *testing.T, ch *ClassHierarchy) {
				// 1:10 should be gone
				assert.Nil(t, ch.classes[tc.MustParseHandle("1:10")])
				// Children should now be under 1:1
				assert.Equal(t, "1:1", ch.classes[tc.MustParseHandle("1:20")].Parent().String())
				assert.Equal(t, "1:1", ch.classes[tc.MustParseHandle("1:21")].Parent().String())
				// Children should be in parent's children list
				children := ch.GetChildren(tc.MustParseHandle("1:1"))
				childStrings := handleStrings(children)
				assert.Contains(t, childStrings, "1:20")
				assert.Contains(t, childStrings, "1:21")
			},
		},
		{
			name:     "delete orphan children - moves children to root",
			strategy: DeleteOrphanChildren,
			setupFn: func() *ClassHierarchy {
				ch := NewClassHierarchy(5)
				require.NoError(t, ch.AddClass(createTestClass("1:1", "1:0", "root", 0)))
				require.NoError(t, ch.AddClass(createTestClass("1:10", "1:1", "child", 1)))
				require.NoError(t, ch.AddClass(createTestClass("1:20", "1:10", "grandchild1", 2)))
				require.NoError(t, ch.AddClass(createTestClass("1:21", "1:10", "grandchild2", 2)))
				return ch
			},
			deleteClass: "1:10",
			expectErr:   false,
			validateFn: func(t *testing.T, ch *ClassHierarchy) {
				// 1:10 should be gone
				assert.Nil(t, ch.classes[tc.MustParseHandle("1:10")])
				// Children should now be under root (1:)
				assert.Equal(t, "1:", ch.classes[tc.MustParseHandle("1:20")].Parent().String())
				assert.Equal(t, "1:", ch.classes[tc.MustParseHandle("1:21")].Parent().String())
			},
		},
		{
			name:     "delete fail if children - fails when class has children",
			strategy: DeleteFailIfChildren,
			setupFn: func() *ClassHierarchy {
				ch := NewClassHierarchy(5)
				require.NoError(t, ch.AddClass(createTestClass("1:1", "1:0", "root", 0)))
				require.NoError(t, ch.AddClass(createTestClass("1:10", "1:1", "child", 1)))
				require.NoError(t, ch.AddClass(createTestClass("1:20", "1:10", "grandchild", 2)))
				return ch
			},
			deleteClass: "1:10",
			expectErr:   true,
			errMsg:      "has 1 children",
		},
		{
			name:     "delete fail if children - succeeds when class has no children",
			strategy: DeleteFailIfChildren,
			setupFn: func() *ClassHierarchy {
				ch := NewClassHierarchy(5)
				require.NoError(t, ch.AddClass(createTestClass("1:1", "1:0", "root", 0)))
				require.NoError(t, ch.AddClass(createTestClass("1:10", "1:1", "child", 1)))
				return ch
			},
			deleteClass: "1:10",
			expectErr:   false,
			validateFn: func(t *testing.T, ch *ClassHierarchy) {
				// 1:10 should be gone
				assert.Nil(t, ch.classes[tc.MustParseHandle("1:10")])
				// Root should still exist
				assert.NotNil(t, ch.classes[tc.MustParseHandle("1:1")])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := tt.setupFn()

			err := ch.DeleteClass(tc.MustParseHandle(tt.deleteClass), tt.strategy)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				if tt.validateFn != nil {
					tt.validateFn(t, ch)
				}
			}
		})
	}
}

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}

func priorityPtr(p Priority) *Priority {
	return &p
}

func handlePtr(h tc.Handle) *tc.Handle {
	return &h
}

func handleStrings(handles []tc.Handle) []string {
	var result []string
	for _, h := range handles {
		result = append(result, h.String())
	}
	return result
}
