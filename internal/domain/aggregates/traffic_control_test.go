package aggregates

import (
	"testing"

	"github.com/rng999/traffic-control-go/pkg/tc"
	"github.com/rng999/traffic-control-go/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestTrafficControlAggregate_Immutability(t *testing.T) {
	// Create initial aggregate
	deviceName, _ := tc.NewDeviceName("eth0")
	original := NewTrafficControlAggregate(deviceName)

	// Test WithHTBQdisc immutability
	handle := tc.NewHandle(1, 0)
	defaultClass := tc.NewHandle(1, 10)

	result := original.WithHTBQdisc(handle, defaultClass)
	assert.True(t, result.IsSuccess())

	newAggregate := result.Value()

	// Verify original is unchanged
	assert.Empty(t, original.GetQdiscs())
	assert.Equal(t, 0, original.Version())
	assert.Empty(t, original.GetUncommittedEvents())

	// Verify new aggregate has the qdisc
	assert.Len(t, newAggregate.GetQdiscs(), 1)
	assert.Equal(t, 1, newAggregate.Version())
	assert.Len(t, newAggregate.GetUncommittedEvents(), 1)

	// Verify they are different instances
	assert.NotSame(t, original, newAggregate)
}

func TestTrafficControlAggregate_WithHTBClass_Immutability(t *testing.T) {
	deviceName, _ := tc.NewDeviceName("eth0")
	aggregate := NewTrafficControlAggregate(deviceName)

	// First add a qdisc
	rootHandle := tc.NewHandle(1, 0)
	defaultClass := tc.NewHandle(1, 30)

	aggregateWithQdisc := aggregate.WithHTBQdisc(rootHandle, defaultClass).Value()

	// Then add a class
	classHandle := tc.NewHandle(1, 10)
	rate := tc.Mbps(100)
	ceil := tc.Mbps(200)

	result := aggregateWithQdisc.WithHTBClass(rootHandle, classHandle, "TestClass", rate, ceil)
	assert.True(t, result.IsSuccess())

	finalAggregate := result.Value()

	// Verify aggregateWithQdisc is unchanged
	assert.Empty(t, aggregateWithQdisc.GetClasses())
	assert.Equal(t, 1, aggregateWithQdisc.Version())

	// Verify final aggregate has both qdisc and class
	assert.Len(t, finalAggregate.GetQdiscs(), 1)
	assert.Len(t, finalAggregate.GetClasses(), 1)
	assert.Equal(t, 2, finalAggregate.Version())
}

func TestTrafficControlAggregate_FunctionalChaining(t *testing.T) {
	deviceName, _ := tc.NewDeviceName("eth0")
	aggregate := NewTrafficControlAggregate(deviceName)

	// Define operations as functions
	addRootQdisc := func(ag *TrafficControlAggregate) types.Result[*TrafficControlAggregate] {
		return ag.WithHTBQdisc(tc.NewHandle(1, 0), tc.NewHandle(1, 30))
	}

	addClass1 := func(ag *TrafficControlAggregate) types.Result[*TrafficControlAggregate] {
		rate := tc.Mbps(100)
		ceil := tc.Mbps(200)
		return ag.WithHTBClass(tc.NewHandle(1, 0), tc.NewHandle(1, 10), "Class1", rate, ceil)
	}

	addClass2 := func(ag *TrafficControlAggregate) types.Result[*TrafficControlAggregate] {
		rate := tc.Mbps(50)
		ceil := tc.Mbps(100)
		return ag.WithHTBClass(tc.NewHandle(1, 0), tc.NewHandle(1, 20), "Class2", rate, ceil)
	}

	// Apply operations using functional composition
	result := aggregate.WithOperations(addRootQdisc, addClass1, addClass2)
	assert.True(t, result.IsSuccess())

	finalAggregate := result.Value()

	// Verify the final state
	assert.Len(t, finalAggregate.GetQdiscs(), 1)
	assert.Len(t, finalAggregate.GetClasses(), 2)
	assert.Equal(t, 3, finalAggregate.Version())

	// Verify original is unchanged
	assert.Empty(t, aggregate.GetQdiscs())
	assert.Empty(t, aggregate.GetClasses())
	assert.Equal(t, 0, aggregate.Version())
}

func TestTrafficControlAggregate_Chain(t *testing.T) {
	deviceName, _ := tc.NewDeviceName("eth0")
	aggregate := NewTrafficControlAggregate(deviceName)

	// Use Chain for single operation
	result := aggregate.Chain(func(ag *TrafficControlAggregate) types.Result[*TrafficControlAggregate] {
		return ag.WithHTBQdisc(tc.NewHandle(1, 0), tc.NewHandle(1, 10))
	})

	assert.True(t, result.IsSuccess())
	assert.Len(t, result.Value().GetQdiscs(), 1)

	// Original should be unchanged
	assert.Empty(t, aggregate.GetQdiscs())
}

func TestTrafficControlAggregate_ErrorPropagation(t *testing.T) {
	deviceName, _ := tc.NewDeviceName("eth0")
	aggregate := NewTrafficControlAggregate(deviceName)

	// Operation that will fail (duplicate qdisc)
	addQdisc := func(ag *TrafficControlAggregate) types.Result[*TrafficControlAggregate] {
		return ag.WithHTBQdisc(tc.NewHandle(1, 0), tc.NewHandle(1, 10))
	}

	addDuplicateQdisc := func(ag *TrafficControlAggregate) types.Result[*TrafficControlAggregate] {
		return ag.WithHTBQdisc(tc.NewHandle(1, 0), tc.NewHandle(1, 20)) // Same handle
	}

	// Apply operations - should fail on second operation
	result := aggregate.WithOperations(addQdisc, addDuplicateQdisc)
	assert.True(t, result.IsFailure())
	assert.Contains(t, result.Error().Error(), "already exists")

	// Original should be unchanged
	assert.Empty(t, aggregate.GetQdiscs())
}

func TestTrafficControlAggregate_BusinessRuleValidation(t *testing.T) {
	deviceName, _ := tc.NewDeviceName("eth0")
	aggregate := NewTrafficControlAggregate(deviceName)

	t.Run("invalid qdisc handle", func(t *testing.T) {
		// Non-root handle should fail
		invalidHandle := tc.NewHandle(1, 1) // Minor != 0
		result := aggregate.WithHTBQdisc(invalidHandle, tc.NewHandle(1, 10))

		assert.True(t, result.IsFailure())
		assert.Contains(t, result.Error().Error(), "root qdisc handle must have minor = 0")
	})

	t.Run("class with invalid parent", func(t *testing.T) {
		// Try to add class without parent qdisc
		nonExistentParent := tc.NewHandle(2, 0)
		classHandle := tc.NewHandle(1, 10)
		rate := tc.Mbps(100)
		ceil := tc.Mbps(200)

		result := aggregate.WithHTBClass(nonExistentParent, classHandle, "TestClass", rate, ceil)

		assert.True(t, result.IsFailure())
		assert.Contains(t, result.Error().Error(), "does not exist")
	})

	t.Run("class with ceil < rate", func(t *testing.T) {
		// First add qdisc
		aggregateWithQdisc := aggregate.WithHTBQdisc(tc.NewHandle(1, 0), tc.NewHandle(1, 30)).Value()

		// Try to add class with ceil < rate
		rate := tc.Mbps(200)
		ceil := tc.Mbps(100) // ceil < rate

		result := aggregateWithQdisc.WithHTBClass(tc.NewHandle(1, 0), tc.NewHandle(1, 10), "TestClass", rate, ceil)

		assert.True(t, result.IsFailure())
		assert.Contains(t, result.Error().Error(), "cannot be less than rate")
	})
}
