package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	t.Run("map over integers", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		result := Map(input, func(x int) int { return x * 2 })
		expected := []int{2, 4, 6, 8, 10}

		assert.Equal(t, expected, result)
	})

	t.Run("map to different type", func(t *testing.T) {
		input := []int{1, 2, 3}
		result := Map(input, func(x int) string {
			if x%2 == 0 {
				return "even"
			}
			return "odd"
		})
		expected := []string{"odd", "even", "odd"}

		assert.Equal(t, expected, result)
	})

	t.Run("map empty slice", func(t *testing.T) {
		input := []int{}
		result := Map(input, func(x int) int { return x * 2 })

		assert.Empty(t, result)
	})
}

func TestFilter(t *testing.T) {
	t.Run("filter even numbers", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5, 6}
		result := Filter(input, func(x int) bool { return x%2 == 0 })
		expected := []int{2, 4, 6}

		assert.Equal(t, expected, result)
	})

	t.Run("filter none match", func(t *testing.T) {
		input := []int{1, 3, 5}
		result := Filter(input, func(x int) bool { return x%2 == 0 })

		assert.Empty(t, result)
	})

	t.Run("filter all match", func(t *testing.T) {
		input := []int{2, 4, 6}
		result := Filter(input, func(x int) bool { return x%2 == 0 })

		assert.Equal(t, input, result)
	})
}

func TestReduce(t *testing.T) {
	t.Run("sum integers", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		result := Reduce(input, 0, func(acc, x int) int { return acc + x })

		assert.Equal(t, 15, result)
	})

	t.Run("concatenate strings", func(t *testing.T) {
		input := []string{"hello", " ", "world"}
		result := Reduce(input, "", func(acc, x string) string { return acc + x })

		assert.Equal(t, "hello world", result)
	})

	t.Run("reduce empty slice", func(t *testing.T) {
		input := []int{}
		result := Reduce(input, 42, func(acc, x int) int { return acc + x })

		assert.Equal(t, 42, result)
	})
}

func TestFind(t *testing.T) {
	t.Run("find existing element", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		result := Find(input, func(x int) bool { return x == 3 })

		assert.True(t, result.IsSome())
		assert.Equal(t, 3, result.Value())
	})

	t.Run("find first matching element", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 3, 5}
		result := Find(input, func(x int) bool { return x == 3 })

		assert.True(t, result.IsSome())
		assert.Equal(t, 3, result.Value())
	})

	t.Run("find non-existing element", func(t *testing.T) {
		input := []int{1, 2, 4, 5}
		result := Find(input, func(x int) bool { return x == 3 })

		assert.True(t, result.IsNone())
	})

	t.Run("find in empty slice", func(t *testing.T) {
		input := []int{}
		result := Find(input, func(x int) bool { return x == 3 })

		assert.True(t, result.IsNone())
	})
}

func TestUpdateOrAppend(t *testing.T) {
	t.Run("update existing element", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		result := UpdateOrAppend(input, 99, func(x int) bool { return x == 3 })
		expected := []int{1, 2, 99, 4, 5}

		assert.Equal(t, expected, result)
		// Verify original is unchanged
		assert.Equal(t, []int{1, 2, 3, 4, 5}, input)
	})

	t.Run("append when no match", func(t *testing.T) {
		input := []int{1, 2, 4, 5}
		result := UpdateOrAppend(input, 99, func(x int) bool { return x == 3 })
		expected := []int{1, 2, 4, 5, 99}

		assert.Equal(t, expected, result)
		// Verify original is unchanged
		assert.Equal(t, []int{1, 2, 4, 5}, input)
	})

	t.Run("update first matching element", func(t *testing.T) {
		input := []int{1, 3, 2, 3, 5}
		result := UpdateOrAppend(input, 99, func(x int) bool { return x == 3 })
		expected := []int{1, 99, 2, 3, 5}

		assert.Equal(t, expected, result)
	})
}

func TestRemoveIf(t *testing.T) {
	t.Run("remove even numbers", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5, 6}
		result := RemoveIf(input, func(x int) bool { return x%2 == 0 })
		expected := []int{1, 3, 5}

		assert.Equal(t, expected, result)
	})

	t.Run("remove none", func(t *testing.T) {
		input := []int{1, 3, 5}
		result := RemoveIf(input, func(x int) bool { return x%2 == 0 })

		assert.Equal(t, input, result)
	})

	t.Run("remove all", func(t *testing.T) {
		input := []int{2, 4, 6}
		result := RemoveIf(input, func(x int) bool { return x%2 == 0 })

		assert.Empty(t, result)
	})
}

func TestFunctionalChaining(t *testing.T) {
	// Test chaining functional operations
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	result := Filter(input, func(x int) bool { return x%2 == 0 })     // [2, 4, 6, 8, 10]
	result = Map(result, func(x int) int { return x * 2 })            // [4, 8, 12, 16, 20]
	sum := Reduce(result, 0, func(acc, x int) int { return acc + x }) // 60

	assert.Equal(t, 60, sum)

	// Test finding in a filtered and mapped collection
	found := Find(result, func(x int) bool { return x > 15 })
	assert.True(t, found.IsSome())
	assert.Equal(t, 16, found.Value())
}
