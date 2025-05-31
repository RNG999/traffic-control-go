package logging

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitialize(t *testing.T) {
	// Reset global logger for test isolation
	defer func() {
		globalLogger = nil
		initOnce = sync.Once{}
	}()

	config := DefaultConfig()
	err := Initialize(config)
	assert.NoError(t, err)

	logger := GetLogger()
	assert.NotNil(t, logger)
}

func TestInitializeFromFile(t *testing.T) {
	// Reset global logger for test isolation
	defer func() {
		globalLogger = nil
		initOnce = sync.Once{}
	}()

	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test_config.json")

	configContent := `{
		"level": "debug",
		"format": "console",
		"output_paths": ["stdout"],
		"development": true
	}`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	err = InitializeFromFile(configFile)
	assert.NoError(t, err)

	logger := GetLogger()
	assert.NotNil(t, logger)
}

func TestInitializeFromEnv(t *testing.T) {
	// Reset global logger for test isolation
	defer func() {
		globalLogger = nil
		initOnce = sync.Once{}
	}()

	// Save original env vars
	originalLevel := os.Getenv("TC_LOG_LEVEL")
	defer func() {
		if originalLevel == "" {
			_ = os.Unsetenv("TC_LOG_LEVEL")
		} else {
			_ = os.Setenv("TC_LOG_LEVEL", originalLevel)
		}
	}()

	_ = os.Setenv("TC_LOG_LEVEL", "debug")

	err := InitializeFromEnv()
	assert.NoError(t, err)

	logger := GetLogger()
	assert.NotNil(t, logger)
}

func TestInitializeDefault(t *testing.T) {
	// Reset global logger for test isolation
	defer func() {
		globalLogger = nil
		initOnce = sync.Once{}
	}()

	err := InitializeDefault()
	assert.NoError(t, err)

	logger := GetLogger()
	assert.NotNil(t, logger)
}

func TestInitializeDevelopment(t *testing.T) {
	// Reset global logger for test isolation
	defer func() {
		globalLogger = nil
		initOnce = sync.Once{}
	}()

	err := InitializeDevelopment()
	assert.NoError(t, err)

	logger := GetLogger()
	assert.NotNil(t, logger)
}

func TestInitializeProduction(t *testing.T) {
	// Reset global logger for test isolation
	defer func() {
		globalLogger = nil
		initOnce = sync.Once{}
	}()

	err := InitializeProduction()
	assert.NoError(t, err)

	logger := GetLogger()
	assert.NotNil(t, logger)
}

func TestGetLoggerAutoInitialize(t *testing.T) {
	// Reset global logger for test isolation
	defer func() {
		globalLogger = nil
		initOnce = sync.Once{}
	}()

	// Get logger without initialization - should auto-initialize
	logger := GetLogger()
	assert.NotNil(t, logger)

	// Second call should return the same instance
	logger2 := GetLogger()
	assert.Equal(t, logger, logger2)
}

func TestSetLogger(t *testing.T) {
	// Reset global logger for test isolation
	defer func() {
		globalLogger = nil
		initOnce = sync.Once{}
	}()

	mockLogger := &MockLogger{}
	SetLogger(mockLogger)

	logger := GetLogger()
	assert.Equal(t, mockLogger, logger)
}

func TestGlobalLoggerMethods(t *testing.T) {
	// Reset global logger for test isolation
	defer func() {
		globalLogger = nil
		initOnce = sync.Once{}
	}()

	// Initialize with development config to avoid issues
	err := InitializeDevelopment()
	require.NoError(t, err)

	// Test global convenience methods - they should not panic
	assert.NotPanics(t, func() {
		Debug("debug message", String("key", "debug"))
		Info("info message", String("key", "info"))
		Warn("warn message", String("key", "warn"))
		ErrorLog("error message", String("key", "error"))
	})

	// Test global context methods
	assert.NotPanics(t, func() {
		WithComponent(ComponentAPI).Info("api message")
		WithDevice("eth0").Info("device message")
		WithClass("web-traffic").Info("class message")
		WithOperation(OperationCreateClass).Info("operation message")
		WithFields(String("custom", "field")).Info("custom field message")
	})
}

func TestGlobalLoggerContextMethods(t *testing.T) {
	// Reset global logger for test isolation
	defer func() {
		globalLogger = nil
		initOnce = sync.Once{}
	}()

	err := InitializeDevelopment()
	require.NoError(t, err)

	// Test that context methods return loggers
	componentLogger := WithComponent(ComponentAPI)
	assert.NotNil(t, componentLogger)

	deviceLogger := WithDevice("eth0")
	assert.NotNil(t, deviceLogger)

	classLogger := WithClass("test-class")
	assert.NotNil(t, classLogger)

	operationLogger := WithOperation(OperationCreateClass)
	assert.NotNil(t, operationLogger)

	fieldsLogger := WithFields(String("test", "value"))
	assert.NotNil(t, fieldsLogger)

	// Test chaining
	chainedLogger := WithComponent(ComponentAPI).
		WithDevice("eth0").
		WithOperation(OperationCreateClass)
	assert.NotNil(t, chainedLogger)
}

func TestGlobalSync(t *testing.T) {
	// Reset global logger for test isolation
	defer func() {
		globalLogger = nil
		initOnce = sync.Once{}
	}()

	err := InitializeDevelopment()
	require.NoError(t, err)

	// Test that Sync doesn't panic
	assert.NotPanics(t, func() {
		_ = Sync()
	})
}

func TestConcurrentAccess(t *testing.T) {
	// Reset global logger for test isolation
	defer func() {
		globalLogger = nil
		initOnce = sync.Once{}
	}()

	// Test concurrent initialization and access
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()

			// Each goroutine tries to get the logger
			logger := GetLogger()
			assert.NotNil(t, logger)

			// And use it
			logger.Info("concurrent test")
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestInitializeOnce(t *testing.T) {
	// Reset global logger for test isolation
	defer func() {
		globalLogger = nil
		initOnce = sync.Once{}
	}()

	// Initialize multiple times - should only happen once
	config1 := DevelopmentConfig()
	config2 := ProductionConfig()

	err1 := Initialize(config1)
	err2 := Initialize(config2)

	assert.NoError(t, err1)
	assert.NoError(t, err2) // Should not return an error even though it's already initialized

	logger := GetLogger()
	assert.NotNil(t, logger)
}
