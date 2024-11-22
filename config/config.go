package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Port string     `yaml:"port"`
	JWT  JwtOptions `yaml:"jwt"`
}

type JwtOptions struct {
	AccessTokenLifetimeMinutes int `yaml:"access_token_lifetime_minutes"`
	RefreshTokenLifetimeDays   int `yaml:"refresh_token_lifetime_days"`
}

func ReadConfig(src string) (*Config, error) {
	file, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	decoder := yaml.NewDecoder(file)
	cfg := &Config{}
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
