package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOption_Some(t *testing.T) {
	option := Some(42)

	assert.True(t, option.IsSome())
	assert.False(t, option.IsNone())
	assert.Equal(t, 42, option.Value())
	assert.Equal(t, 42, option.GetOrElse(0))
}

func TestOption_None(t *testing.T) {
	option := None[int]()

	assert.False(t, option.IsSome())
	assert.True(t, option.IsNone())
	assert.Equal(t, 42, option.GetOrElse(42))

	// Value() should panic for None
	assert.Panics(t, func() {
		option.Value()
	})
}

func TestOption_OptionFromPtr(t *testing.T) {
	t.Run("non-nil pointer", func(t *testing.T) {
		value := 42
		option := OptionFromPtr(&value)

		assert.True(t, option.IsSome())
		assert.Equal(t, 42, option.Value())
	})

	t.Run("nil pointer", func(t *testing.T) {
		var ptr *int = nil
		option := OptionFromPtr(ptr)

		assert.True(t, option.IsNone())
		assert.Equal(t, 42, option.GetOrElse(42))
	})
}

func TestOption_Map(t *testing.T) {
	t.Run("map over Some", func(t *testing.T) {
		option := Some(5)
		result := option.Map(func(x int) int { return x * 2 })

		assert.True(t, result.IsSome())
		assert.Equal(t, 10, result.Value())
	})

	t.Run("map over None", func(t *testing.T) {
		option := None[int]()
		result := option.Map(func(x int) int { return x * 2 })

		assert.True(t, result.IsNone())
	})
}

func TestOption_FlatMap(t *testing.T) {
	t.Run("flatMap over Some returning Some", func(t *testing.T) {
		option := Some(5)
		result := option.FlatMap(func(x int) Option[int] {
			if x > 0 {
				return Some(x * 2)
			}
			return None[int]()
		})

		assert.True(t, result.IsSome())
		assert.Equal(t, 10, result.Value())
	})

	t.Run("flatMap over Some returning None", func(t *testing.T) {
		option := Some(-5)
		result := option.FlatMap(func(x int) Option[int] {
			if x > 0 {
				return Some(x * 2)
			}
			return None[int]()
		})

		assert.True(t, result.IsNone())
	})

	t.Run("flatMap over None", func(t *testing.T) {
		option := None[int]()
		result := option.FlatMap(func(x int) Option[int] {
			return Some(x * 2)
		})

		assert.True(t, result.IsNone())
	})
}

func TestOption_Filter(t *testing.T) {
	t.Run("filter Some with passing predicate", func(t *testing.T) {
		option := Some(10)
		result := option.Filter(func(x int) bool { return x > 5 })

		assert.True(t, result.IsSome())
		assert.Equal(t, 10, result.Value())
	})

	t.Run("filter Some with failing predicate", func(t *testing.T) {
		option := Some(3)
		result := option.Filter(func(x int) bool { return x > 5 })

		assert.True(t, result.IsNone())
	})

	t.Run("filter None", func(t *testing.T) {
		option := None[int]()
		result := option.Filter(func(x int) bool { return x > 5 })

		assert.True(t, result.IsNone())
	})
}

func TestOption_Match(t *testing.T) {
	t.Run("match Some", func(t *testing.T) {
		option := Some(42)
		var result int
		var executed bool

		option.Match(
			func(value int) {
				result = value
				executed = true
			},
			func() {
				executed = false
			},
		)

		assert.True(t, executed)
		assert.Equal(t, 42, result)
	})

	t.Run("match None", func(t *testing.T) {
		option := None[int]()
		var executed bool

		option.Match(
			func(value int) {
				executed = false
			},
			func() {
				executed = true
			},
		)

		assert.True(t, executed)
	})
}

func TestOption_ToPtr(t *testing.T) {
	t.Run("Some to pointer", func(t *testing.T) {
		option := Some(42)
		ptr := option.ToPtr()

		assert.NotNil(t, ptr)
		assert.Equal(t, 42, *ptr)

		// Verify it's a copy, not the original
		*ptr = 100
		assert.Equal(t, 42, option.Value()) // Original should be unchanged
	})

	t.Run("None to pointer", func(t *testing.T) {
		option := None[int]()
		ptr := option.ToPtr()

		assert.Nil(t, ptr)
	})
}

func TestOption_Chaining(t *testing.T) {
	// Test functional chaining
	result := Some(5).
		Map(func(x int) int { return x * 2 }).
		Filter(func(x int) bool { return x > 5 }).
		FlatMap(func(x int) Option[int] { return Some(x + 1) }).
		GetOrElse(0)

	assert.Equal(t, 11, result)

	// Test chaining that results in None
	result2 := Some(2).
		Map(func(x int) int { return x * 2 }).
		Filter(func(x int) bool { return x > 5 }).
		FlatMap(func(x int) Option[int] { return Some(x + 1) }).
		GetOrElse(999)

	assert.Equal(t, 999, result2)
}
