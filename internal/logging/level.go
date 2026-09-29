package logging

import (
	"fmt"
	"log/slog"
	"strings"
)

func GetLevel(cfgLevel string) (slog.Level, error) {
	level := slog.LevelDebug

	switch strings.ToLower(cfgLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info", "trace":
		level = slog.LevelInfo
	case "error":
		level = slog.LevelError
	case "warn", "warning":
		level = slog.LevelWarn
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level %q", level)
	}

	return level, nil
}
