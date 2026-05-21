package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config общая конфигурация приложения
type Config struct {
	Logger LoggerConfig `yaml:"logger"`
	DB     DBConfig     `yaml:"database"`
	Kafka  KafkaConfig  `yaml:"kafka"`
	Storer StorerConfig `yaml:"storer"`
}

// LoggerConfig конфигурация логгера
type LoggerConfig struct {
	Level string `yaml:"level"`
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

// StorerConfig конфигурация storer
type StorerConfig struct {
	// Дополнительные настройки при необходимости
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
