package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Marattttt/portfolio/gorunner/internal/config"
)

func main() {
	conf, err := config.FromEnv(context.Background())
	if err != nil {
		slog.Error("Could not create config from env", slog.String("err", err.Error()))
		os.Exit(1)
	}

	slog.Info("Config parsed from env", slog.Any("conf", conf))

	if err := conf.Apply(); err != nil {
		slog.Error("Could not apply config", slog.String("err", err.Error()))
		os.Exit(1)
	}

	slog.Info("Config applied successfully")
}
