package config

import (
	"log/slog"
	"sync"
)

var (
	mu     sync.RWMutex
	logger *slog.Logger
)

func SetLogger(l *slog.Logger) {
	if l == nil {
		l = slog.Default()
	}

	mu.Lock()
	defer mu.Unlock()

	logger = l.With("module", "core")
}

func ComponentLogger(component string) *slog.Logger {
	return logger.With("component", component)
}
