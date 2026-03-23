package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// DefaultLevel returns the default log level (info).
func DefaultLevel() slog.Level {
	return slog.LevelInfo
}

// Config configures the logger.
type Config struct {
	Level      slog.Level
	LogToFile  bool
	LogDir     string
	LogFile    string
	JSONOutput bool
}

// NewLogger creates a slog.Logger with stdout handler and optionally a file handler.
func NewLogger(cfg Config) (*slog.Logger, io.Closer, error) {
	var handlers []slog.Handler

	// Stdout handler
	stdoutLevel := cfg.Level
	stdoutOpts := &slog.HandlerOptions{Level: stdoutLevel}
	var stdoutHandler slog.Handler
	if cfg.JSONOutput {
		stdoutHandler = slog.NewJSONHandler(os.Stdout, stdoutOpts)
	} else {
		stdoutHandler = slog.NewTextHandler(os.Stdout, stdoutOpts)
	}
	handlers = append(handlers, stdoutHandler)

	// File handler (optional)
	var file *os.File
	if cfg.LogToFile && cfg.LogDir != "" && cfg.LogFile != "" {
		if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
			return nil, nil, fmt.Errorf("create log dir: %w", err)
		}
		path := filepath.Join(cfg.LogDir, cfg.LogFile)
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, nil, fmt.Errorf("open log file: %w", err)
		}
		file = f
		fileHandler := slog.NewJSONHandler(file, &slog.HandlerOptions{Level: cfg.Level})
		handlers = append(handlers, fileHandler)
	}

	handler := slog.New(multiHandler(handlers))
	return handler, file, nil
}

// multiHandler forwards records to multiple handlers.
type multiHandler struct {
	handlers []slog.Handler
}

func multiHandler(handlers []slog.Handler) slog.Handler {
	return &multiHandler{handlers: handlers}
}

func (h *multiHandler) Enabled(level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(r slog.Record) error {
	var lastErr error
	for _, handler := range h.handlers {
		if handler.Enabled(r.Level) {
			if err := handler.Handle(r); err != nil {
				lastErr = err
			}
		}
	}
	return lastErr
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{handlers: newHandlers}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &multiHandler{handlers: newHandlers}
}
