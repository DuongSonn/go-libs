package logger

import (
	"log/slog"
)

// Config holds the configuration for the logger
type Config struct {
	// Level defines the minimum log level that will be output
	// Valid values: debug, info, warn, error
	Level string `json:"level" yaml:"level" default:"info"`
	// Output defines where logs will be written
	// Valid values: stdout, stderr, or a file path
	Output string `json:"output" yaml:"output" default:"stdout"`
	// Format defines the output format
	// Valid values: json, text
	Format string `json:"format" yaml:"format" default:"json"`
	// HideFields specifies field names that should be hidden from logs
	HideFields []string `json:"hide_fields" yaml:"hide_fields"`
}

// GetLevel converts the string level to slog.Level
func (c *Config) GetLevel() slog.Level {
	switch c.Level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
