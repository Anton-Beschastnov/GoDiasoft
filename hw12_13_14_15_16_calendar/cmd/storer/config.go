package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Logger LoggerConfig `yaml:"logger"`
	DB     DBConfig     `yaml:"database"`
	Kafka  KafkaConfig  `yaml:"kafka"`
	Storer StorerConfig `yaml:"storer"`
}

type LoggerConfig struct {
	Level string `yaml:"level"`
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

type StorerConfig struct {
	KafkaTopic string `yaml:"kafkaTopic"`
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
