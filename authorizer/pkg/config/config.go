package config

import "time"

type Config struct {
	AuthConfig
}

type AuthConfig struct {
	Issuer           string        `env:"ISSUER, default=maratbakasov.com"`
	AccessSecret     string        `env:"ACCESS_SECRET"`
	RefreshSecret    string        `env:"REFRESH_SECRET"`
	AccessValidTime  time.Duration `env:"ACCESS_VALID_FOR, default=6h"`
	RefreshValidTime time.Duration `env:"REFRESH_VALID_FOR, default=3d"`
}
