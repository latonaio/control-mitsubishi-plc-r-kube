package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Server   Server
	IsDebug  bool
	YamlPath string
}

type Server struct {
	Addr             string `envconfig:"ADDRESS"`
	Port             string `envconfig:"PORT"`
	ShutdownWaitTime int    `envconfig:"SHUTDOWN_WAIT_TIME"`
}

func New() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
