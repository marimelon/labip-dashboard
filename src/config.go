package main

import (
	"github.com/caarlos0/env/v8"
)

type Config struct {
	Port          string `env:"PORT" envDefault:"8080"`
	InterfaceName string `env:"NET_INTERFACE" envDefault:""`
	PVEEndpoint   string `env:"PVE_ENDPOINT" envDefault:""`
	PVEApiUser    string `env:"PVE_API_USER" envDefault:""`
	PVEApiKey     string `env:"PVE_API_KEY" envDefault:""`

	NeburaUser     string `env:"NEBURA_USER" envDefault:""`
	NeburaPass     string `env:"NEBURA_PASSWORD" envDefault:""`
	NeburaEndpoint string `env:"NEBURA_ENDPOINT" envDefault:""`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
