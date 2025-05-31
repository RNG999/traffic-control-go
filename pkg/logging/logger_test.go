package logging

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		hasErr bool
	}{
		{
			name:   "default config",
			config: DefaultConfig(),
			hasErr: false,
		},
		{
			name:   "development config",
			config: DevelopmentConfig(),
			hasErr: false,
		},
		{
			name:   "production config",
			config: ProductionConfig(),
			hasErr: false,
		},
		{
			name: "json format",
			config: Config{
				Level:       LevelInfo,
				Format:      "json",
				OutputPaths: []string{"stdout"},
			},
			hasErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)
			if tt.hasErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
			}
		})
	}
}

func TestLoggerLevels(t *testing.T) {
	// Create a logger for testing
	config := Config{
		Level:       LevelDebug,
		Format:      "json",
		OutputPaths: []string{"/dev/stdout"}, // This will be redirected in production
		Development: true,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Test different log levels
	logger.Debug("debug message", String("key", "debug_value"))
	logger.Info("info message", String("key", "info_value"))
	logger.Warn("warn message", String("key", "warn_value"))
	logger.Error("error message", String("key", "error_value"))

	// Test that the logger doesn't panic on normal operations
	assert.NotPanics(t, func() {
		logger.Debug("test debug")
		logger.Info("test info")
		logger.Warn("test warn")
		logger.Error("test error")
	})
}

func TestLoggerWithFields(t *testing.T) {
	config := DevelopmentConfig()
	logger, err := NewLogger(config)
	require.NoError(t, err)

	// Test field methods
	loggerWithFields := logger.WithFields(
		String("device", "eth0"),
		Int("priority", 1),
		Bool("enabled", true),
	)

	assert.NotNil(t, loggerWithFields)

	// Test context methods
	deviceLogger := logger.WithDevice("eth0")
	classLogger := logger.WithClass("web-traffic")
	operationLogger := logger.WithOperation(OperationCreateClass)
	componentLogger := logger.WithComponent(ComponentAPI)

	assert.NotNil(t, deviceLogger)
	assert.NotNil(t, classLogger)
	assert.NotNil(t, operationLogger)
	assert.NotNil(t, componentLogger)

	// Test chaining
	chainedLogger := logger.
		WithComponent(ComponentAPI).
		WithDevice("eth0").
		WithOperation(OperationCreateClass)

	assert.NotNil(t, chainedLogger)

	// Test that chained logger can log without panic
	assert.NotPanics(t, func() {
		chainedLogger.Info("test message")
	})
}

func TestFieldConstructors(t *testing.T) {
	tests := []struct {
		name     string
		field    Field
		expected Field
	}{
		{
			name:     "string field",
			field:    String("key", "value"),
			expected: Field{Key: "key", Value: "value"},
		},
		{
			name:     "int field",
			field:    Int("count", 42),
			expected: Field{Key: "count", Value: 42},
		},
		{
			name:     "bool field",
			field:    Bool("enabled", true),
			expected: Field{Key: "enabled", Value: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.field)
		})
	}
	
	// Test error field separately since we can't easily compare errors
	t.Run("error field", func(t *testing.T) {
		testErr := errors.New("test error")
		field := Error(testErr)
		assert.Equal(t, "error", field.Key)
		assert.Equal(t, testErr, field.Value)
	})
}

func TestTrafficControlSpecificMethods(t *testing.T) {
	config := DevelopmentConfig()
	logger, err := NewLogger(config)
	require.NoError(t, err)

	tests := []struct {
		name   string
		method func() Logger
	}{
		{
			name:   "WithDevice",
			method: func() Logger { return logger.WithDevice("eth0") },
		},
		{
			name:   "WithClass",
			method: func() Logger { return logger.WithClass("web-traffic") },
		},
		{
			name:   "WithOperation",
			method: func() Logger { return logger.WithOperation(OperationCreateClass) },
		},
		{
			name:   "WithBandwidth",
			method: func() Logger { return logger.WithBandwidth("100Mbps") },
		},
		{
			name:   "WithPriority",
			method: func() Logger { return logger.WithPriority(1) },
		},
		{
			name:   "WithComponent",
			method: func() Logger { return logger.WithComponent(ComponentAPI) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contextLogger := tt.method()
			assert.NotNil(t, contextLogger)

			// Test that the logger can log without panic
			assert.NotPanics(t, func() {
				contextLogger.Info("test message")
			})
		})
	}
}

func TestLoggerConstants(t *testing.T) {
	// Test component constants
	components := []string{
		ComponentAPI,
		ComponentDomain,
		ComponentInfrastructure,
		ComponentCommands,
		ComponentQueries,
		ComponentNetlink,
		ComponentEventStore,
		ComponentValidation,
		ComponentConfig,
	}

	for _, component := range components {
		assert.NotEmpty(t, component)
	}

	// Test operation constants
	operations := []string{
		OperationCreateClass,
		OperationDeleteClass,
		OperationUpdateClass,
		OperationCreateQdisc,
		OperationDeleteQdisc,
		OperationCreateFilter,
		OperationDeleteFilter,
		OperationValidation,
		OperationConfigLoad,
		OperationConfigSave,
		OperationApplyConfig,
		OperationNetlinkCall,
		OperationEventStore,
	}

	for _, operation := range operations {
		assert.NotEmpty(t, operation)
	}
}

func TestLoggerSync(t *testing.T) {
	config := DefaultConfig()
	logger, err := NewLogger(config)
	require.NoError(t, err)

	// Test that Sync doesn't panic and returns no error for stdout
	err = logger.Sync()
	// Note: Sync might return an error for stdout on some systems, which is expected
	// We just verify it doesn't panic
	assert.NotPanics(t, func() {
		logger.Sync()
	})
}

// TestMockLogger tests the logger interface with a mock implementation
func TestMockLogger(t *testing.T) {
	mock := &MockLogger{}
	
	// Test that mock implements the interface
	var _ Logger = (*MockLogger)(nil)
	
	// Test basic operations
	mock.Info("test message")
	mock.Debug("debug message")
	mock.Warn("warn message")
	mock.Error("error message")
	
	// Test context methods
	deviceLogger := mock.WithDevice("eth0")
	assert.NotNil(t, deviceLogger)
	
	classLogger := mock.WithClass("test-class")
	assert.NotNil(t, classLogger)
}

// MockLogger is a simple mock implementation for testing
type MockLogger struct {
	messages []string
	fields   []Field
}

func (m *MockLogger) Debug(msg string, fields ...Field) {
	m.messages = append(m.messages, "DEBUG: "+msg)
	m.fields = append(m.fields, fields...)
}

func (m *MockLogger) Info(msg string, fields ...Field) {
	m.messages = append(m.messages, "INFO: "+msg)
	m.fields = append(m.fields, fields...)
}

func (m *MockLogger) Warn(msg string, fields ...Field) {
	m.messages = append(m.messages, "WARN: "+msg)
	m.fields = append(m.fields, fields...)
}

func (m *MockLogger) Error(msg string, fields ...Field) {
	m.messages = append(m.messages, "ERROR: "+msg)
	m.fields = append(m.fields, fields...)
}

func (m *MockLogger) Fatal(msg string, fields ...Field) {
	m.messages = append(m.messages, "FATAL: "+msg)
	m.fields = append(m.fields, fields...)
}

func (m *MockLogger) WithContext(ctx context.Context) Logger {
	return m
}

func (m *MockLogger) WithFields(fields ...Field) Logger {
	newMock := &MockLogger{
		messages: m.messages,
		fields:   append(m.fields, fields...),
	}
	return newMock
}

func (m *MockLogger) WithDevice(deviceName string) Logger {
	return m.WithFields(String("device", deviceName))
}

func (m *MockLogger) WithClass(className string) Logger {
	return m.WithFields(String("class", className))
}

func (m *MockLogger) WithOperation(operation string) Logger {
	return m.WithFields(String("operation", operation))
}

func (m *MockLogger) WithBandwidth(bandwidth string) Logger {
	return m.WithFields(String("bandwidth", bandwidth))
}

func (m *MockLogger) WithPriority(priority int) Logger {
	return m.WithFields(Int("priority", priority))
}

func (m *MockLogger) WithComponent(component string) Logger {
	return m.WithFields(String("component", component))
}

func (m *MockLogger) Sync() error {
	return nil
}