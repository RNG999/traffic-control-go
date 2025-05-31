package logging

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, LevelInfo, config.Level)
	assert.Equal(t, "console", config.Format)
	assert.Equal(t, []string{"stdout"}, config.OutputPaths)
	assert.False(t, config.Development)
	assert.False(t, config.SamplingEnabled)
}

func TestDevelopmentConfig(t *testing.T) {
	config := DevelopmentConfig()

	assert.Equal(t, LevelDebug, config.Level)
	assert.Equal(t, "console", config.Format)
	assert.Equal(t, []string{"stdout"}, config.OutputPaths)
	assert.True(t, config.Development)
	assert.False(t, config.SamplingEnabled)
}

func TestProductionConfig(t *testing.T) {
	config := ProductionConfig()

	assert.Equal(t, LevelInfo, config.Level)
	assert.Equal(t, "json", config.Format)
	assert.Equal(t, []string{"stdout"}, config.OutputPaths)
	assert.False(t, config.Development)
	assert.True(t, config.SamplingEnabled)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid default config",
			config:    DefaultConfig(),
			wantError: false,
		},
		{
			name: "invalid log level",
			config: Config{
				Level:       Level("invalid"),
				Format:      "console",
				OutputPaths: []string{"stdout"},
			},
			wantError: true,
			errorMsg:  "invalid log level",
		},
		{
			name: "invalid format",
			config: Config{
				Level:       LevelInfo,
				Format:      "invalid",
				OutputPaths: []string{"stdout"},
			},
			wantError: true,
			errorMsg:  "invalid log format",
		},
		{
			name: "empty output paths",
			config: Config{
				Level:       LevelInfo,
				Format:      "console",
				OutputPaths: []string{},
			},
			wantError: true,
			errorMsg:  "at least one output path must be specified",
		},
		{
			name: "valid file output",
			config: Config{
				Level:       LevelInfo,
				Format:      "json",
				OutputPaths: []string{"/tmp/test.log"},
			},
			wantError: false,
		},
		{
			name: "invalid component level",
			config: Config{
				Level:       LevelInfo,
				Format:      "console",
				OutputPaths: []string{"stdout"},
				ComponentLevels: map[string]Level{
					"api": Level("invalid"),
				},
			},
			wantError: true,
			errorMsg:  "invalid log level for component",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "logging.json")

	config := Config{
		Level:       LevelDebug,
		Format:      "json",
		OutputPaths: []string{"stdout", "/tmp/app.log"},
		Development: true,
		ComponentLevels: map[string]Level{
			"api":    LevelInfo,
			"domain": LevelDebug,
		},
	}

	// Write config to file
	data, err := json.MarshalIndent(config, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(configFile, data, 0644)
	require.NoError(t, err)

	// Load config from file
	loadedConfig, err := LoadConfigFromFile(configFile)
	require.NoError(t, err)

	assert.Equal(t, config.Level, loadedConfig.Level)
	assert.Equal(t, config.Format, loadedConfig.Format)
	assert.Equal(t, config.OutputPaths, loadedConfig.OutputPaths)
	assert.Equal(t, config.Development, loadedConfig.Development)
	assert.Equal(t, config.ComponentLevels, loadedConfig.ComponentLevels)
}

func TestLoadConfigFromFileErrors(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
		wantErr  bool
	}{
		{
			name:     "file not found",
			filename: "nonexistent.json",
			wantErr:  true,
		},
		{
			name:     "invalid json",
			filename: "invalid.json",
			content:  `{"level": "info", "format":}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.content != "" {
				tmpDir := t.TempDir()
				filename := filepath.Join(tmpDir, tt.filename)
				err := os.WriteFile(filename, []byte(tt.content), 0644)
				require.NoError(t, err)
				tt.filename = filename
			}

			_, err := LoadConfigFromFile(tt.filename)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Save original env vars
	originalVars := map[string]string{
		"TC_LOG_LEVEL":       os.Getenv("TC_LOG_LEVEL"),
		"TC_LOG_FORMAT":      os.Getenv("TC_LOG_FORMAT"),
		"TC_LOG_OUTPUTS":     os.Getenv("TC_LOG_OUTPUTS"),
		"TC_LOG_DEVELOPMENT": os.Getenv("TC_LOG_DEVELOPMENT"),
		"TC_LOG_SAMPLING":    os.Getenv("TC_LOG_SAMPLING"),
	}

	// Clean up after test
	defer func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	tests := []struct {
		name     string
		envVars  map[string]string
		expected Config
	}{
		{
			name:     "no env vars",
			envVars:  map[string]string{},
			expected: DefaultConfig(),
		},
		{
			name: "all env vars set",
			envVars: map[string]string{
				"TC_LOG_LEVEL":       "debug",
				"TC_LOG_FORMAT":      "json",
				"TC_LOG_OUTPUTS":     "stdout,/tmp/app.log",
				"TC_LOG_DEVELOPMENT": "true",
				"TC_LOG_SAMPLING":    "true",
			},
			expected: Config{
				Level:           LevelDebug,
				Format:          "json",
				OutputPaths:     []string{"stdout", "/tmp/app.log"},
				Development:     true,
				SamplingEnabled: true,
			},
		},
		{
			name: "partial env vars",
			envVars: map[string]string{
				"TC_LOG_LEVEL":  "warn",
				"TC_LOG_FORMAT": "console",
			},
			expected: Config{
				Level:           LevelWarn,
				Format:          "console",
				OutputPaths:     []string{"stdout"}, // default
				Development:     false,              // default
				SamplingEnabled: false,              // default
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars first
			for key := range originalVars {
				os.Unsetenv(key)
			}

			// Set test env vars
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			config := LoadConfigFromEnv()

			assert.Equal(t, tt.expected.Level, config.Level)
			assert.Equal(t, tt.expected.Format, config.Format)
			assert.Equal(t, tt.expected.OutputPaths, config.OutputPaths)
			assert.Equal(t, tt.expected.Development, config.Development)
			assert.Equal(t, tt.expected.SamplingEnabled, config.SamplingEnabled)
		})
	}
}

func TestConfigString(t *testing.T) {
	config := Config{
		Level:           LevelInfo,
		Format:          "json",
		OutputPaths:     []string{"stdout"},
		Development:     false,
		SamplingEnabled: true,
	}

	str := config.String()
	assert.Contains(t, str, "Level=info")
	assert.Contains(t, str, "Format=json")
	assert.Contains(t, str, "Outputs=[stdout]")
	assert.Contains(t, str, "Development=false")
	assert.Contains(t, str, "Sampling=true")
}

func TestConfigComponentLevels(t *testing.T) {
	config := DefaultConfig()

	// Test setting component level
	config.SetComponentLevel("api", LevelDebug)
	config.SetComponentLevel("domain", LevelWarn)

	// Test getting component levels
	assert.Equal(t, LevelDebug, config.GetComponentLevel("api"))
	assert.Equal(t, LevelWarn, config.GetComponentLevel("domain"))
	assert.Equal(t, config.Level, config.GetComponentLevel("nonexistent"))

	// Test validation with component levels
	err := config.Validate()
	assert.NoError(t, err)
}

func TestLevelConstants(t *testing.T) {
	levels := []Level{
		LevelDebug,
		LevelInfo,
		LevelWarn,
		LevelError,
		LevelFatal,
	}

	for _, level := range levels {
		assert.NotEmpty(t, level)
		assert.IsType(t, Level(""), level)
	}
}
