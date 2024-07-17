package main

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Environment string `envconfig:"ENVIRONMENT" default:"development"`
	Hostname    string `envconfig:"HOSTNAME" default:"localhost"`
	Port        string `envconfig:"PORT" default:"3000"`
	DatabaseUrl string `envconfig:"DATABASE_URL" default:"postgres://postgres:postgres@localhost:5432/ecommerce"`
	JwtSecret   string `envconfig:"JWT_SECRET" default:"supersecret"`
}

func GetConfig() (Config, error) {
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
