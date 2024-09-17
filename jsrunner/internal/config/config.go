package config

import (
	"log/slog"
	"strings"
)

func ApplyMode(mode string) error {
	switch strings.ToLower(mode) {
	case "debug":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	return nil
}
