package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

// Logger wraps slog.Logger to provide additional functionality
type Logger struct {
	*slog.Logger
	hideFields map[string]struct{} // For O(1) lookup of fields to hide
}

// New creates a new logger with the given configuration
func New(cfg Config) *Logger {
	var output io.Writer
	switch cfg.Output {
	case "stderr":
		output = os.Stderr
	case "":
		fallthrough
	case "stdout":
		output = os.Stdout
	default:
		// Assume it's a file path
		file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open log file %s: %v\n", cfg.Output, err)
			output = os.Stdout
		} else {
			output = file
		}
	}

	// Convert hideFields slice to map for O(1) lookup
	hideFields := make(map[string]struct{})
	for _, field := range cfg.HideFields {
		hideFields[field] = struct{}{}
	}

	// Create handler based on format
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     cfg.GetLevel(),
		AddSource: true, // Always show source as per requirement
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Format time as MM/DD/YYYY HH:mm:ss
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					return slog.String(a.Key, t.Format("01/02/2006 15:04:05"))
				}
			}

			// Hide attributes as per requirement
			if len(groups) > 0 {
				return slog.Attr{}
			}

			// Mask specific fields if they're in the hideFields list
			if _, exists := hideFields[a.Key]; exists {
				return slog.String(a.Key, "***")
			}

			return a
		},
	}

	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	return &Logger{
		Logger:     slog.New(handler),
		hideFields: hideFields,
	}
}

// With returns a new Logger with the given attributes
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger:     l.Logger.With(args...),
		hideFields: l.hideFields,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.Logger.DebugContext(ctx, msg, args...)
}

// Info logs an info message
func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.Logger.InfoContext(ctx, msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.Logger.WarnContext(ctx, msg, args...)
}

// Error logs an error message
func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.Logger.ErrorContext(ctx, msg, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(ctx context.Context, msg string, args ...any) {
	l.Logger.ErrorContext(ctx, msg, args...)
	os.Exit(1)
}
