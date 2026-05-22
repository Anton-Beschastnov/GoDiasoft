package main

import (
	"os"

	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/scheduler"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Logger    LoggerConfig     `yaml:"logger"`
	Storage   StorageConfig    `yaml:"storage"`
	DB        DBConfig         `yaml:"database"`
	Kafka     KafkaConfig      `yaml:"kafka"`
	Scheduler scheduler.Config `yaml:"scheduler"`
}

type LoggerConfig struct {
	Level string `yaml:"level"`
}

type StorageConfig struct {
	Type string `yaml:"type"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type KafkaConfig struct {
	BootstrapServers []string `yaml:"bootstrapServers"`
	Topic            string   `yaml:"topic"`
}

func NewConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
