# Logging System Documentation

The traffic-control-go project includes a comprehensive structured logging system built on top of Uber's Zap library. This document describes how to use and configure the logging system.

## Overview

The logging system provides:

- **Structured Logging**: JSON and console output formats with structured fields
- **Multiple Log Levels**: DEBUG, INFO, WARN, ERROR, FATAL
- **Context-Aware**: Automatic context propagation for devices, classes, operations
- **Configurable**: Environment variables, configuration files, and programmatic setup
- **Performance**: Built on Zap for high-performance logging
- **Integration**: Seamlessly integrated throughout the codebase

## Quick Start

### Basic Usage

```go
import "github.com/rng999/traffic-control-go/pkg/logging"

// Initialize logging (done automatically in main)
logging.InitializeDevelopment()

// Basic logging
logging.Info("Traffic control started")
logging.Error("Configuration failed", logging.Error(err))

// Context-aware logging
logger := logging.WithComponent(logging.ComponentAPI).
    WithDevice("eth0").
    WithOperation(logging.OperationCreateClass)

logger.Info("Creating traffic class", 
    logging.String("class_name", "web-traffic"),
    logging.String("bandwidth", "100Mbps"),
)
```

### Environment Configuration

Set these environment variables to configure logging:

```bash
# Log level (debug, info, warn, error, fatal)
export TC_LOG_LEVEL=info

# Output format (json, console)
export TC_LOG_FORMAT=json

# Output destinations (comma-separated)
export TC_LOG_OUTPUTS=stdout,/var/log/traffic-control.log

# Development mode (true/false)
export TC_LOG_DEVELOPMENT=false

# Enable sampling for high-volume scenarios (true/false)
export TC_LOG_SAMPLING=true
```

## Configuration

### Configuration File

Create a JSON configuration file:

```json
{
  "level": "info",
  "format": "json",
  "output_paths": ["stdout", "/var/log/traffic-control.log"],
  "development": false,
  "sampling_enabled": true,
  "component_levels": {
    "api": "info",
    "domain": "debug",
    "infrastructure": "warn",
    "netlink": "debug"
  }
}
```

Load it in your application:

```go
err := logging.InitializeFromFile("config/logging.json")
if err != nil {
    log.Fatalf("Failed to initialize logging: %v", err)
}
```

### Programmatic Configuration

```go
config := logging.Config{
    Level:           logging.LevelInfo,
    Format:          "json",
    OutputPaths:     []string{"stdout", "/tmp/app.log"},
    Development:     false,
    SamplingEnabled: true,
}

err := logging.Initialize(config)
```

### Predefined Configurations

```go
// Development: debug level, console format, verbose output
logging.InitializeDevelopment()

// Production: info level, JSON format, sampling enabled
logging.InitializeProduction()

// Default: info level, console format, basic setup
logging.InitializeDefault()
```

## Structured Fields

### Standard Fields

```go
// Basic field types
logging.String("key", "value")
logging.Int("count", 42)
logging.Int64("bytes", 1024)
logging.Float64("ratio", 0.95)
logging.Bool("enabled", true)
logging.Error(err)
logging.Duration("elapsed", time.Since(start))
```

### Traffic Control Context

```go
// Device context
logger.WithDevice("eth0")

// Traffic class context
logger.WithClass("web-traffic")

// Operation context
logger.WithOperation(logging.OperationCreateClass)

// Bandwidth context
logger.WithBandwidth("100Mbps")

// Priority context
logger.WithPriority(1)

// Component context
logger.WithComponent(logging.ComponentAPI)
```

### Chaining Context

```go
logger := logging.GetLogger().
    WithComponent(logging.ComponentAPI).
    WithDevice("eth0").
    WithOperation(logging.OperationCreateClass).
    WithClass("web-traffic")

logger.Info("Applying traffic class configuration")
```

## Log Levels

### Level Hierarchy

- **FATAL**: System cannot continue, will exit
- **ERROR**: Error conditions that need attention
- **WARN**: Warning conditions, potentially problematic
- **INFO**: General information about program execution
- **DEBUG**: Detailed information for debugging

### Usage Guidelines

```go
// DEBUG: Detailed information for troubleshooting
logger.Debug("Validating traffic class configuration",
    logging.String("class_name", className),
    logging.String("guaranteed_bw", guaranteedBw),
    logging.String("max_bw", maxBw),
)

// INFO: General operational information
logger.Info("Traffic class created successfully",
    logging.String("class_name", className),
    logging.String("device", deviceName),
)

// WARN: Potentially problematic conditions
logger.Warn("Class bandwidth exceeds recommendation",
    logging.String("class_name", className),
    logging.String("bandwidth", bandwidth),
    logging.String("recommendation", "100Mbps"),
)

// ERROR: Error conditions that need attention
logger.Error("Failed to apply traffic control configuration",
    logging.Error(err),
    logging.String("device", deviceName),
)

// FATAL: Critical errors that require program termination
logger.Fatal("Cannot initialize netlink interface",
    logging.Error(err),
)
```

## Components and Operations

### Component Constants

```go
const (
    ComponentAPI           = "api"
    ComponentDomain        = "domain"
    ComponentInfrastructure = "infrastructure"
    ComponentCommands      = "commands"
    ComponentQueries       = "queries"
    ComponentNetlink       = "netlink"
    ComponentEventStore    = "eventstore"
    ComponentValidation    = "validation"
    ComponentConfig        = "config"
)
```

### Operation Constants

```go
const (
    OperationCreateClass   = "create_class"
    OperationDeleteClass   = "delete_class"
    OperationUpdateClass   = "update_class"
    OperationCreateQdisc   = "create_qdisc"
    OperationDeleteQdisc   = "delete_qdisc"
    OperationCreateFilter  = "create_filter"
    OperationDeleteFilter  = "delete_filter"
    OperationValidation    = "validation"
    OperationConfigLoad    = "config_load"
    OperationConfigSave    = "config_save"
    OperationApplyConfig   = "apply_config"
    OperationNetlinkCall   = "netlink_call"
    OperationEventStore    = "event_store"
)
```

## Integration Patterns

### API Layer

```go
func (tc *TrafficController) Apply() error {
    logger := tc.logger.WithOperation(logging.OperationApplyConfig)
    logger.Info("Starting traffic control configuration application",
        logging.Int("class_count", len(tc.classes)),
    )
    
    if err := tc.validate(); err != nil {
        logger.Error("Configuration validation failed", logging.Error(err))
        return err
    }
    
    logger.Info("Configuration applied successfully")
    return nil
}
```

### Infrastructure Layer

```go
func (a *RealNetlinkAdapter) AddQdisc(device valueobjects.DeviceName, config QdiscConfig) types.Result[Unit] {
    logger := a.logger.WithDevice(device.String()).WithOperation(logging.OperationCreateQdisc)
    logger.Info("Adding qdisc", 
        logging.String("qdisc_type", string(config.Type)),
        logging.String("handle", config.Handle.String()),
    )
    
    // Implementation...
    
    if err != nil {
        logger.Error("Failed to add qdisc", logging.Error(err))
        return types.Failure[Unit](err)
    }
    
    logger.Debug("Qdisc added successfully")
    return types.Success(Unit{})
}
```

### Domain Layer

```go
func (a *TrafficControlAggregate) CreateClass(cmd CreateClassCommand) error {
    logger := logging.WithComponent(logging.ComponentDomain).
        WithDevice(a.deviceName.String()).
        WithClass(cmd.ClassName)
    
    logger.Debug("Creating traffic class in domain",
        logging.String("guaranteed_bw", cmd.GuaranteedBandwidth.String()),
    )
    
    // Domain logic...
    
    logger.Info("Traffic class created in domain")
    return nil
}
```

## Log Output Examples

### Console Format (Development)

```
2025-05-31T02:32:22.360Z	info	logging/logger.go:193	Creating new traffic controller	{"component": "api", "device": "eth0"}
2025-05-31T02:32:22.360Z	warn	logging/logger.go:199	Traffic class missing priority	{"component": "api", "device": "eth0", "class_name": "database-traffic", "validation_error": "missing_priority"}
```

### JSON Format (Production)

```json
{
  "level": "info",
  "timestamp": "2025-05-31T02:33:35.758Z",
  "caller": "logging/logger.go:193",
  "message": "Creating new traffic controller",
  "component": "api",
  "device": "eth0"
}
```

## Performance Considerations

### Sampling

Enable sampling in high-volume scenarios:

```go
config := logging.ProductionConfig()
config.SamplingEnabled = true
```

### Structured Fields

Prefer structured fields over string formatting:

```go
// Good: Structured and searchable
logger.Info("Class configuration applied",
    logging.String("class_name", className),
    logging.String("bandwidth", bandwidth),
)

// Avoid: String formatting loses structure
logger.Info(fmt.Sprintf("Class %s applied with bandwidth %s", className, bandwidth))
```

### Conditional Logging

For expensive operations, check log level:

```go
if logger.IsDebugEnabled() {
    expensiveData := computeExpensiveDebugInfo()
    logger.Debug("Expensive debug info", logging.String("data", expensiveData))
}
```

## Log Aggregation

The structured JSON format is compatible with common log aggregation systems:

### ELK Stack (Elasticsearch, Logstash, Kibana)

```json
{
  "level": "info",
  "timestamp": "2025-05-31T02:33:35.758Z",
  "component": "api",
  "device": "eth0",
  "operation": "create_class",
  "class_name": "web-traffic",
  "message": "Traffic class created successfully"
}
```

### Prometheus/Grafana

Use structured fields for metrics and alerting:

- `component`: Service component
- `device`: Network device
- `operation`: Operation type
- `validation_error`: Error category

### Fluentd/Fluent Bit

The JSON format works directly with Fluentd parsers.

## Testing

### Mock Logger

Use the provided mock logger for testing:

```go
func TestMyFunction(t *testing.T) {
    mock := &logging.MockLogger{}
    
    // Use mock in your tests
    myFunction(mock)
    
    // Verify logging behavior
    assert.Contains(t, mock.Messages, "Expected log message")
}
```

### Test Configuration

Use test-specific configuration:

```go
func TestWithLogging(t *testing.T) {
    config := logging.Config{
        Level:       logging.LevelDebug,
        Format:      "console",
        OutputPaths: []string{"/dev/null"}, // Suppress output in tests
    }
    
    logger, _ := logging.NewLogger(config)
    logging.SetLogger(logger)
    
    // Your test code here
}
```

## Best Practices

### 1. Use Context-Aware Logging

Always include relevant context:

```go
// Good
logger := logging.WithComponent(logging.ComponentAPI).
    WithDevice(deviceName).
    WithOperation(logging.OperationCreateClass)

// Better - add more specific context
logger := logger.WithClass(className).WithPriority(priority)
```

### 2. Log at Appropriate Levels

- Use DEBUG for detailed troubleshooting information
- Use INFO for normal operational events
- Use WARN for recoverable issues
- Use ERROR for serious problems
- Use FATAL only for unrecoverable errors

### 3. Include Error Context

```go
if err != nil {
    logger.Error("Operation failed",
        logging.Error(err),
        logging.String("operation", "create_class"),
        logging.String("class_name", className),
    )
    return err
}
```

### 4. Use Structured Fields

Make logs searchable and analyzable:

```go
// Good: Searchable
logger.Info("Bandwidth limit applied",
    logging.String("device", "eth0"),
    logging.String("class", "web-traffic"),
    logging.String("bandwidth", "100Mbps"),
)

// Avoid: Hard to search
logger.Info("Applied 100Mbps bandwidth limit to web-traffic class on eth0")
```

### 5. Consistent Field Names

Use consistent field names across the application:

- `device`: Network device name
- `class_name` or `class`: Traffic class name
- `operation`: Operation being performed
- `bandwidth`: Bandwidth values
- `priority`: Traffic priority

## Configuration Examples

See the `examples/` directory for complete configuration examples:

- `logging-config.json`: Basic production configuration
- `logging-config-development.json`: Development configuration
- `logging-config-production.json`: Production configuration with sampling

## Troubleshooting

### Common Issues

1. **Logs not appearing**: Check log level and output paths
2. **Permission denied**: Ensure write permissions for log files
3. **Performance issues**: Enable sampling or increase log level
4. **Format issues**: Verify JSON syntax in configuration files

### Debug Configuration

Enable debug logging for troubleshooting:

```bash
export TC_LOG_LEVEL=debug
export TC_LOG_FORMAT=console
export TC_LOG_DEVELOPMENT=true
```

## Future Enhancements

Planned logging improvements:

1. **Log rotation**: Automatic log file rotation
2. **Remote logging**: Direct integration with log aggregation services
3. **Metrics integration**: Automatic metrics generation from logs
4. **Structured tracing**: Distributed tracing support
5. **Configuration hot-reload**: Dynamic configuration updates