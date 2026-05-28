package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Logger  LoggerConf  `yaml:"logger"`
	HTTP    HTTPConf    `yaml:"http"`
	Storage StorageConf `yaml:"storage"`
	DB      DBConf      `yaml:"db"`
}

type LoggerConf struct {
	Level string `yaml:"level"`
}

type HTTPConf struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type StorageConf struct {
	Type string `yaml:"type"` // "memory" или "sql"
}

type DBConf struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func NewConfig(configPath string) (Config, error) {
	config := Config{
		Logger: LoggerConf{
			Level: "info",
		},
		HTTP: HTTPConf{
			Host: "127.0.0.1",
			Port: 8080,
		},
		Storage: StorageConf{
			Type: "memory",
		},
		DB: DBConf{
			Host:     "localhost",
			Port:     5432,
			User:     "calendar",
			Password: "calendar",
			Database: "calendar",
		},
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}
