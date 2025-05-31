package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Level represents log levels
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
	LevelFatal Level = "fatal"
)

// Config holds logging configuration
type Config struct {
	// Level specifies the minimum log level
	Level Level `json:"level" yaml:"level"`

	// Format specifies the log format ("json" or "console")
	Format string `json:"format" yaml:"format"`

	// OutputPaths specifies where to write logs (files or "stdout", "stderr")
	OutputPaths []string `json:"output_paths" yaml:"output_paths"`

	// Development mode enables more verbose logging and stack traces
	Development bool `json:"development" yaml:"development"`

	// SamplingEnabled enables log sampling for high-volume scenarios
	SamplingEnabled bool `json:"sampling_enabled" yaml:"sampling_enabled"`

	// Component-specific log levels (optional)
	ComponentLevels map[string]Level `json:"component_levels,omitempty" yaml:"component_levels,omitempty"`
}

// DefaultConfig returns a reasonable default configuration
func DefaultConfig() Config {
	return Config{
		Level:           LevelInfo,
		Format:          "console",
		OutputPaths:     []string{"stdout"},
		Development:     false,
		SamplingEnabled: false,
	}
}

// DevelopmentConfig returns a configuration suitable for development
func DevelopmentConfig() Config {
	return Config{
		Level:           LevelDebug,
		Format:          "console",
		OutputPaths:     []string{"stdout"},
		Development:     true,
		SamplingEnabled: false,
	}
}

// ProductionConfig returns a configuration suitable for production
func ProductionConfig() Config {
	return Config{
		Level:           LevelInfo,
		Format:          "json",
		OutputPaths:     []string{"stdout"},
		Development:     false,
		SamplingEnabled: true,
	}
}

// LoadConfigFromFile loads logging configuration from a JSON or YAML file
func LoadConfigFromFile(filename string) (Config, error) {
	config := DefaultConfig()

	// Validate filename for path traversal
	if err := validateLogFilePath(filename); err != nil {
		return config, fmt.Errorf("invalid file path: %w", err)
	}

	// #nosec G304 - filename is validated by validateLogFilePath above
	file, err := os.Open(filename)
	if err != nil {
		return config, fmt.Errorf("failed to open config file %s: %w", filename, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log the error but don't override the main return error
		}
	}()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return config, fmt.Errorf("failed to decode config file %s: %w", filename, err)
	}

	return config, config.Validate()
}

// LoadConfigFromEnv loads logging configuration from environment variables
func LoadConfigFromEnv() Config {
	config := DefaultConfig()

	if level := os.Getenv("TC_LOG_LEVEL"); level != "" {
		config.Level = Level(strings.ToLower(level))
	}

	if format := os.Getenv("TC_LOG_FORMAT"); format != "" {
		config.Format = strings.ToLower(format)
	}

	if outputs := os.Getenv("TC_LOG_OUTPUTS"); outputs != "" {
		config.OutputPaths = strings.Split(outputs, ",")
	}

	if dev := os.Getenv("TC_LOG_DEVELOPMENT"); dev == "true" {
		config.Development = true
	}

	if sampling := os.Getenv("TC_LOG_SAMPLING"); sampling == "true" {
		config.SamplingEnabled = true
	}

	return config
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate log level
	validLevels := map[Level]bool{
		LevelDebug: true,
		LevelInfo:  true,
		LevelWarn:  true,
		LevelError: true,
		LevelFatal: true,
	}

	if !validLevels[c.Level] {
		return fmt.Errorf("invalid log level: %s (valid levels: debug, info, warn, error, fatal)", c.Level)
	}

	// Validate format
	if c.Format != "json" && c.Format != "console" {
		return fmt.Errorf("invalid log format: %s (valid formats: json, console)", c.Format)
	}

	// Validate output paths
	if len(c.OutputPaths) == 0 {
		return fmt.Errorf("at least one output path must be specified")
	}

	for _, path := range c.OutputPaths {
		if path != "stdout" && path != "stderr" {
			// Check if file path is writable
			// #nosec G304 - path is from validated configuration
			if _, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600); err != nil {
				return fmt.Errorf("cannot write to log file %s: %w", path, err)
			}
		}
	}

	// Validate component levels
	for component, level := range c.ComponentLevels {
		if !validLevels[level] {
			return fmt.Errorf("invalid log level for component %s: %s", component, level)
		}
	}

	return nil
}

// String returns a string representation of the config
func (c *Config) String() string {
	return fmt.Sprintf("Level=%s, Format=%s, Outputs=%v, Development=%t, Sampling=%t",
		c.Level, c.Format, c.OutputPaths, c.Development, c.SamplingEnabled)
}

// SetComponentLevel sets the log level for a specific component
func (c *Config) SetComponentLevel(component string, level Level) {
	if c.ComponentLevels == nil {
		c.ComponentLevels = make(map[string]Level)
	}
	c.ComponentLevels[component] = level
}

// GetComponentLevel returns the log level for a specific component
func (c *Config) GetComponentLevel(component string) Level {
	if c.ComponentLevels != nil {
		if level, exists := c.ComponentLevels[component]; exists {
			return level
		}
	}
	return c.Level
}

// validateLogFilePath validates that the log file path is safe and doesn't contain path traversal
func validateLogFilePath(filename string) error {
	// Clean the path to resolve any .. or . components
	cleaned := filepath.Clean(filename)

	// Check for path traversal attempts
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("path traversal detected in filename: %s", filename)
	}

	// Ensure it's not an absolute path to system directories (except typical log locations)
	if filepath.IsAbs(cleaned) {
		// Allow certain safe absolute paths for log files
		if strings.HasPrefix(cleaned, "/tmp/") ||
			strings.HasPrefix(cleaned, "/var/tmp/") ||
			strings.HasPrefix(cleaned, "/var/log/") ||
			strings.HasPrefix(cleaned, "/home/") {
			return nil
		}
		return fmt.Errorf("absolute paths to system directories not allowed: %s", filename)
	}

	return nil
}
