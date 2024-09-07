package config

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	// Port to listen for incoming connections
	Port int `env:"PORT, default=8080"`
	// Mode to run the application in
	//
	// Debug mode is supported, others default to release
	//
	// Debug sets slog level to debug
	Mode string `env:"MODE, default=debug"`
	// Username to use when building and running code
	// Does not support users that require a password to login
	//
	// If not provided, same user is assumed
	RunAs string `env:"RUN_AS"`

	// directory to place executables
	RunDir string `env:"RUN_DIR, default=/tmp/gorunner/runtime"`

	isDebug bool
}

func (c Config) IsDebug() bool {
	return c.isDebug
}

func FromEnv(ctx context.Context) (*Config, error) {
	var c Config
	if err := envconfig.Process(ctx, &c); err != nil {
		return nil, fmt.Errorf("processing env: %w", err)
	}

	return &c, nil
}

func (c *Config) Apply() error {
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("cannot use %d as port", c.Port)
	}

	switch strings.ToLower(c.Mode) {
	case "debug":
		c.isDebug = true
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	return nil
}
