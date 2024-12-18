package config

import (
	"context"
	"time"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	Port int `env:"PORT, default=8080"`

	DBConfig   `env:", prefix=DB_"`
	AuthConfig `env:", prefix=AUTH_"`
}

type DBConfig struct {
	PostgresURL      string `env:"POSTGRES_URL"`
	MigrationsSource string `env:"MIGRATIONS_URL, default=file://internal/db/migrations"`
}

type AuthConfig struct {
	Issuer           string        `env:"ISSUER, default=maratbakasov.com"`
	AccessSecret     string        `env:"ACCESS_SECRET"`
	RefreshSecret    string        `env:"REFRESH_SECRET"`
	AccessValidTime  time.Duration `env:"ACCESS_VALID_FOR, default=1h"`
	RefreshValidTime time.Duration `env:"REFRESH_VALID_FOR, default=72h"`
}

func ConfigFromEnv(ctx context.Context) (*Config, error) {
	var conf Config

	if err := envconfig.Process(ctx, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
