package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Server struct {
	Address string `yaml:"address"`
	Port    string `yaml:"port"`
}

type Config struct {
	Server       Server
	DatabasePath string `yaml:"database_path"`
	Workers      int    `yaml:"workers"`
}

func InitConfig() (*Config, error) {
	file, err := os.ReadFile("config/config.yaml")
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
