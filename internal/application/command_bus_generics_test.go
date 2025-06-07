package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rng999/traffic-control-go/internal/commands/models"
	"github.com/rng999/traffic-control-go/internal/infrastructure/eventstore"
	"github.com/rng999/traffic-control-go/internal/infrastructure/netlink"
	"github.com/rng999/traffic-control-go/pkg/logging"
)

// TestGenericCommandHandler tests the new generic command handler interface
func TestGenericCommandHandler(t *testing.T) {
	t.Run("command_bus_should_work_with_generic_handlers", func(t *testing.T) {
		// Setup
		eventStore := eventstore.NewMemoryEventStoreWithContext()
		netlinkAdapter := netlink.NewMockAdapter()
		logger := logging.WithComponent("application")
		service := NewTrafficControlService(eventStore, netlinkAdapter, logger)

		// This test should fail initially as we haven't implemented generics yet
		// It demonstrates what we want to achieve

		// Create a command that should be handled with type safety
		command := &models.CreateHTBQdiscCommand{
			DeviceName:   "eth0",
			Handle:       "1:0",
			DefaultClass: "1:999",
		}

		// We want to be able to execute commands with compile-time type safety
		// This should not require runtime type assertions
		ctx := context.Background()
		err := service.commandBus.ExecuteTypedCommand(ctx, command)

		// For now, this will fail because ExecuteTyped doesn't exist
		// We'll implement it with generics
		assert.NoError(t, err)
	})

	t.Run("generic_handler_should_have_compile_time_type_safety", func(t *testing.T) {
		// This test demonstrates the type safety we want to achieve
		
		// Mock generic handler that should work with specific command types
		handler := &MockGenericHTBQdiscHandler{}
		
		// Should be able to register with compile-time type checking
		// This will fail initially but shows our target design
		command := &models.CreateHTBQdiscCommand{
			DeviceName:   "eth0", 
			Handle:       "1:0",
			DefaultClass: "1:999",
		}

		ctx := context.Background()
		
		// The handler should receive the exact command type, not interface{}
		err := handler.HandleTyped(ctx, command)
		require.NoError(t, err)
		
		// Verify the handler received the correct type
		assert.True(t, handler.receivedCorrectType)
	})

	t.Run("generic_command_bus_should_handle_all_qdisc_types", func(t *testing.T) {
		// Setup
		eventStore := eventstore.NewMemoryEventStoreWithContext()
		netlinkAdapter := netlink.NewMockAdapter()
		logger := logging.WithComponent("application")
		service := NewTrafficControlService(eventStore, netlinkAdapter, logger)

		ctx := context.Background()

		// First create an HTB qdisc (required for HTB classes)
		htbQdiscCommand := &models.CreateHTBQdiscCommand{
			DeviceName:   "eth0",
			Handle:       "1:0",
			DefaultClass: "1:999",
		}
		err := service.commandBus.ExecuteTypedCommand(ctx, htbQdiscCommand)
		assert.NoError(t, err)

		// Test HTB class command (now that parent HTB qdisc exists)
		htbClassCommand := &models.CreateHTBClassCommand{
			DeviceName: "eth0",
			Parent:     "1:0",
			ClassID:    "1:1",
			Rate:       "50Mbps",
			Ceil:       "100Mbps",
		}
		err = service.commandBus.ExecuteTypedCommand(ctx, htbClassCommand)
		assert.NoError(t, err)

		// Test Filter command (now that target class exists)
		filterCommand := &models.CreateFilterCommand{
			DeviceName: "eth0",
			Parent:     "1:0",
			Priority:   100,
			Protocol:   "ip",
			FlowID:     "1:1",
			Match:      map[string]string{"src": "192.168.1.0/24"},
		}
		err = service.commandBus.ExecuteTypedCommand(ctx, filterCommand)
		assert.NoError(t, err)

		// Test TBF qdisc command (on different device to avoid conflicts)
		tbfCommand := &models.CreateTBFQdiscCommand{
			DeviceName: "eth1",
			Handle:     "1:0",
			Rate:       "100Mbps",
			Buffer:     1000,
			Limit:      2000,
			Burst:      3000,
		}
		err = service.commandBus.ExecuteTypedCommand(ctx, tbfCommand)
		assert.NoError(t, err)

		// Test PRIO qdisc command (on different device to avoid conflicts)
		prioCommand := &models.CreatePRIOQdiscCommand{
			DeviceName: "eth2",
			Handle:     "1:0",
			Bands:      3,
			Priomap:    []uint8{1, 2, 2, 2, 1, 2, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1},
		}
		err = service.commandBus.ExecuteTypedCommand(ctx, prioCommand)
		assert.NoError(t, err)

		// Test FQ_CODEL qdisc command (on different device to avoid conflicts)
		fqcodelCommand := &models.CreateFQCODELQdiscCommand{
			DeviceName: "eth3",
			Handle:     "1:0",
			Limit:      1000,
			Flows:      1024,
			Target:     5000,
			Interval:   100000,
			Quantum:    1514,
			ECN:        true,
		}
		err = service.commandBus.ExecuteTypedCommand(ctx, fqcodelCommand)
		assert.NoError(t, err)
	})
}

// MockGenericHTBQdiscHandler is a mock that demonstrates the generic interface we want
type MockGenericHTBQdiscHandler struct {
	receivedCorrectType bool
}

// HandleTyped demonstrates the generic handler method we want to implement
func (h *MockGenericHTBQdiscHandler) HandleTyped(ctx context.Context, command *models.CreateHTBQdiscCommand) error {
	// This method receives the specific command type, not interface{}
	h.receivedCorrectType = true
	
	// Validate that we have the expected fields without type assertions
	if command.DeviceName == "" || command.Handle == "" {
		return assert.AnError
	}
	
	return nil
}

// Test that shows current interface{} usage has type assertion overhead
func TestCurrentInterfaceUsageProblems(t *testing.T) {
	t.Run("current_handler_requires_type_assertions", func(t *testing.T) {
		// This demonstrates the current problem with interface{} usage
		
		handler := &CurrentStyleHandler{}
		
		// Current implementation requires passing interface{}
		var command interface{} = &models.CreateHTBQdiscCommand{
			DeviceName:   "eth0",
			Handle:       "1:0", 
			DefaultClass: "1:999",
		}
		
		ctx := context.Background()
		err := handler.Handle(ctx, command)
		assert.NoError(t, err)
		
		// This shows the type assertion that should be eliminated
		assert.True(t, handler.performedTypeAssertion)
	})
}

// CurrentStyleHandler demonstrates the current interface{} approach
type CurrentStyleHandler struct {
	performedTypeAssertion bool
}

// Handle uses the current interface{} approach (what we want to replace)
func (h *CurrentStyleHandler) Handle(ctx context.Context, command interface{}) error {
	// This type assertion is what we want to eliminate with generics
	cmd, ok := command.(*models.CreateHTBQdiscCommand)
	if !ok {
		return assert.AnError
	}
	
	h.performedTypeAssertion = true
	
	// Rest of the logic...
	if cmd.DeviceName == "" {
		return assert.AnError
	}
	
	return nil
}