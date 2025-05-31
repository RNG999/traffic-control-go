package logging

import (
	"sync"
)

var (
	// globalLogger holds the global logger instance
	globalLogger Logger
	// initOnce ensures the global logger is initialized only once
	initOnce sync.Once
	// mutex protects access to the global logger
	mutex sync.RWMutex
)

// Initialize sets up the global logger with the provided configuration
func Initialize(config Config) error {
	var err error
	initOnce.Do(func() {
		mutex.Lock()
		defer mutex.Unlock()
		
		globalLogger, err = NewLogger(config)
	})
	return err
}

// InitializeFromFile initializes the global logger from a configuration file
func InitializeFromFile(filename string) error {
	config, err := LoadConfigFromFile(filename)
	if err != nil {
		return err
	}
	return Initialize(config)
}

// InitializeFromEnv initializes the global logger from environment variables
func InitializeFromEnv() error {
	config := LoadConfigFromEnv()
	return Initialize(config)
}

// InitializeDefault initializes the global logger with default configuration
func InitializeDefault() error {
	config := DefaultConfig()
	return Initialize(config)
}

// InitializeDevelopment initializes the global logger with development configuration
func InitializeDevelopment() error {
	config := DevelopmentConfig()
	return Initialize(config)
}

// InitializeProduction initializes the global logger with production configuration
func InitializeProduction() error {
	config := ProductionConfig()
	return Initialize(config)
}

// GetLogger returns the global logger instance
func GetLogger() Logger {
	mutex.RLock()
	defer mutex.RUnlock()
	
	if globalLogger == nil {
		// Auto-initialize with default config if not already initialized
		mutex.RUnlock()
		mutex.Lock()
		if globalLogger == nil {
			config := DefaultConfig()
			logger, err := NewLogger(config)
			if err != nil {
				panic("Failed to initialize default logger: " + err.Error())
			}
			globalLogger = logger
		}
		mutex.Unlock()
		mutex.RLock()
	}
	
	return globalLogger
}

// SetLogger sets the global logger instance (useful for testing)
func SetLogger(logger Logger) {
	mutex.Lock()
	defer mutex.Unlock()
	globalLogger = logger
}

// Sync flushes any buffered log entries from the global logger
func Sync() error {
	return GetLogger().Sync()
}

// Global convenience methods that delegate to the global logger

// Debug logs a debug message using the global logger
func Debug(msg string, fields ...Field) {
	GetLogger().Debug(msg, fields...)
}

// Info logs an info message using the global logger
func Info(msg string, fields ...Field) {
	GetLogger().Info(msg, fields...)
}

// Warn logs a warning message using the global logger
func Warn(msg string, fields ...Field) {
	GetLogger().Warn(msg, fields...)
}

// ErrorLog logs an error message using the global logger
func ErrorLog(msg string, fields ...Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal logs a fatal message using the global logger and exits
func Fatal(msg string, fields ...Field) {
	GetLogger().Fatal(msg, fields...)
}

// WithComponent returns a logger with component context from the global logger
func WithComponent(component string) Logger {
	return GetLogger().WithComponent(component)
}

// WithDevice returns a logger with device context from the global logger
func WithDevice(deviceName string) Logger {
	return GetLogger().WithDevice(deviceName)
}

// WithClass returns a logger with class context from the global logger
func WithClass(className string) Logger {
	return GetLogger().WithClass(className)
}

// WithOperation returns a logger with operation context from the global logger
func WithOperation(operation string) Logger {
	return GetLogger().WithOperation(operation)
}

// WithFields returns a logger with additional fields from the global logger
func WithFields(fields ...Field) Logger {
	return GetLogger().WithFields(fields...)
}