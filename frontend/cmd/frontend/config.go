package main

import (
	"context"

	"github.com/Marattttt/portfolio/frontend/internal/runners"
	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	Runners runners.Config `env:", prefix=RUN_"`
}

func createConfig(ctx context.Context) (*Config, error) {
	var conf Config

	if err := envconfig.Process(ctx, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
