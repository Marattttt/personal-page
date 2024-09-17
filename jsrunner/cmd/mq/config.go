package main

import (
	"context"
	"fmt"

	"github.com/Marattttt/personal-page/jsrunner/internal/config"
	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	MQ      MQConfig      `env:", prefix=MQ_"`
	Runtime RuntimeConfig `env:", prefix=RUNTIME_"`
	Mode    string        `env:"MODE, default=debug"`
}

func (conf Config) Apply() error {
	return config.ApplyMode(conf.Mode)
}

type MQConfig struct {
	Addr     string `env:"ADDR, default=localhost:5672"`
	User     string `env:"USER, default=guest"`
	Password string `env:"PASS, default=guest"`
	RecvQ    string `env:"RECVQ, default=jsrunner"`
	RespondQ string `env:"RESPQ, default=jsrunner-response"`
}

func (m MQConfig) URL() string {
	return fmt.Sprintf("amqp://%s:%s@%s", m.User, m.Password, m.Addr)
}

type RuntimeConfig struct {
	RunAs     *string `env:"USERNAME, noinit"`
	RunAsPass *string `env:"PASS, noinit"`
	Dir       string  `env:"DIR, default=./runtimedir"`
}

func CreateConfig(ctx context.Context) (*Config, error) {
	var conf Config
	if err := envconfig.Process(ctx, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
