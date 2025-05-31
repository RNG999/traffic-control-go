// Package logging provides a structured logging interface and implementation
// for the traffic control system.
package logging

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger defines the interface for structured logging throughout the application
type Logger interface {
	// Standard log levels
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	// Context-aware logging
	WithContext(ctx context.Context) Logger
	WithFields(fields ...Field) Logger

	// Traffic Control specific context methods
	WithDevice(deviceName string) Logger
	WithClass(className string) Logger
	WithOperation(operation string) Logger
	WithBandwidth(bandwidth string) Logger
	WithPriority(priority int) Logger

	// Component-specific loggers
	WithComponent(component string) Logger

	// Sync flushes any buffered log entries
	Sync() error
}

// Field represents a structured logging field
type Field struct {
	Key   string
	Value interface{}
}

// Convenience field constructors
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

func Duration(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Component constants for structured logging
const (
	ComponentAPI            = "api"
	ComponentDomain         = "domain"
	ComponentInfrastructure = "infrastructure"
	ComponentCommands       = "commands"
	ComponentQueries        = "queries"
	ComponentNetlink        = "netlink"
	ComponentEventStore     = "eventstore"
	ComponentValidation     = "validation"
	ComponentConfig         = "config"
)

// Operation constants for traffic control operations
const (
	OperationCreateClass  = "create_class"
	OperationDeleteClass  = "delete_class"
	OperationUpdateClass  = "update_class"
	OperationCreateQdisc  = "create_qdisc"
	OperationDeleteQdisc  = "delete_qdisc"
	OperationCreateFilter = "create_filter"
	OperationDeleteFilter = "delete_filter"
	OperationValidation   = "validation"
	OperationConfigLoad   = "config_load"
	OperationConfigSave   = "config_save"
	OperationApplyConfig  = "apply_config"
	OperationNetlinkCall  = "netlink_call"
	OperationEventStore   = "event_store"
)

// zapLogger implements the Logger interface using Uber's zap library
type zapLogger struct {
	zap    *zap.Logger
	sugar  *zap.SugaredLogger
	fields []Field
}

// NewLogger creates a new logger instance with the given configuration
func NewLogger(config Config) (Logger, error) {
	zapConfig, err := buildZapConfig(config)
	if err != nil {
		return nil, err
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return &zapLogger{
		zap:   logger,
		sugar: logger.Sugar(),
	}, nil
}

// buildZapConfig creates a zap configuration from our Config
func buildZapConfig(config Config) (zap.Config, error) {
	var level zapcore.Level
	switch config.Level {
	case LevelDebug:
		level = zapcore.DebugLevel
	case LevelInfo:
		level = zapcore.InfoLevel
	case LevelWarn:
		level = zapcore.WarnLevel
	case LevelError:
		level = zapcore.ErrorLevel
	case LevelFatal:
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	zapConfig := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: config.Development,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: config.Format,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      config.OutputPaths,
		ErrorOutputPaths: []string{"stderr"},
	}

	return zapConfig, nil
}

// convertFields converts our Field slice to zap fields
func (l *zapLogger) convertFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	return zapFields
}

// Debug logs a debug message with optional fields
func (l *zapLogger) Debug(msg string, fields ...Field) {
	allFields := append(l.fields, fields...)
	l.zap.Debug(msg, l.convertFields(allFields)...)
}

// Info logs an info message with optional fields
func (l *zapLogger) Info(msg string, fields ...Field) {
	allFields := append(l.fields, fields...)
	l.zap.Info(msg, l.convertFields(allFields)...)
}

// Warn logs a warning message with optional fields
func (l *zapLogger) Warn(msg string, fields ...Field) {
	allFields := append(l.fields, fields...)
	l.zap.Warn(msg, l.convertFields(allFields)...)
}

// Error logs an error message with optional fields
func (l *zapLogger) Error(msg string, fields ...Field) {
	allFields := append(l.fields, fields...)
	l.zap.Error(msg, l.convertFields(allFields)...)
}

// Fatal logs a fatal message with optional fields and exits
func (l *zapLogger) Fatal(msg string, fields ...Field) {
	allFields := append(l.fields, fields...)
	l.zap.Fatal(msg, l.convertFields(allFields)...)
}

// WithContext returns a logger with context information
func (l *zapLogger) WithContext(ctx context.Context) Logger {
	// Extract any context values if needed in the future
	return l
}

// WithFields returns a logger with additional fields
func (l *zapLogger) WithFields(fields ...Field) Logger {
	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)

	return &zapLogger{
		zap:    l.zap,
		sugar:  l.sugar,
		fields: newFields,
	}
}

// WithDevice returns a logger with device context
func (l *zapLogger) WithDevice(deviceName string) Logger {
	return l.WithFields(String("device", deviceName))
}

// WithClass returns a logger with class context
func (l *zapLogger) WithClass(className string) Logger {
	return l.WithFields(String("class", className))
}

// WithOperation returns a logger with operation context
func (l *zapLogger) WithOperation(operation string) Logger {
	return l.WithFields(String("operation", operation))
}

// WithBandwidth returns a logger with bandwidth context
func (l *zapLogger) WithBandwidth(bandwidth string) Logger {
	return l.WithFields(String("bandwidth", bandwidth))
}

// WithPriority returns a logger with priority context
func (l *zapLogger) WithPriority(priority int) Logger {
	return l.WithFields(Int("priority", priority))
}

// WithComponent returns a logger with component context
func (l *zapLogger) WithComponent(component string) Logger {
	return l.WithFields(String("component", component))
}

// Sync flushes any buffered log entries
func (l *zapLogger) Sync() error {
	return l.zap.Sync()
}
