package main

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config общая конфигурация приложения
type Config struct {
	Logger    LoggerConfig    `yaml:"logger"`
	Storage   StorageConfig   `yaml:"storage"`
	DB        DBConfig        `yaml:"database"`
	Kafka     KafkaConfig     `yaml:"kafka"`
	Scheduler SchedulerConfig `yaml:"scheduler"`
}

// LoggerConfig конфигурация логгера
type LoggerConfig struct {
	Level string `yaml:"level"`
}

// StorageConfig конфигурация хранилища
type StorageConfig struct {
	Type string `yaml:"type"`
}

// DBConfig конфигурация базы данных
type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

// KafkaConfig конфигурация Kafka
type KafkaConfig struct {
	BootstrapServers []string `yaml:"bootstrap_servers"`
	Topic            string   `yaml:"topic"`
}

// SchedulerConfig конфигурация scheduler
type SchedulerConfig struct {
	ScanInterval time.Duration `yaml:"scan_interval"`
}

// NewConfig загружает конфигурацию из файла
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
