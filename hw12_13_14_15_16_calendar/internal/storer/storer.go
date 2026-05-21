package storer

import (
	"context"
	"fmt"
	"time"

	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/app"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/kafka"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/storage"
)

// Config конфигурация storer
type Config struct {
	StorageType string      `yaml:"storage_type"`
	DB          DBConfig    `yaml:"database"`
	Kafka       KafkaConfig `yaml:"kafka"`
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

// Storer процесс для сохранения уведомлений
type Storer struct {
	logger   logger.Logger
	storage  app.Storage
	consumer kafka.ConsumerInterface
	config   Config
}

// New создает новый storer
func New(logger logger.Logger, storage app.Storage, consumer kafka.ConsumerInterface, config Config) *Storer {
	return &Storer{
		logger:   logger,
		storage:  storage,
		consumer: consumer,
		config:   config,
	}
}

// Run запускает storer
func (s *Storer) Run(ctx context.Context) error {
	s.logger.Info("storer is running...")

	// Обработчик уведомлений
	handler := func(notification *kafka.Notification) error {
		return s.saveNotification(ctx, notification)
	}

	return s.consumer.Consume(ctx, s.config.Kafka.Topic, handler)
}

// saveNotification сохраняет уведомление в базу данных
func (s *Storer) saveNotification(ctx context.Context, notification *kafka.Notification) error {
	dbNotification := &storage.Notification{
		ID:        notification.EventID,
		EventID:   notification.EventID,
		Title:     notification.Title,
		StartTime: notification.StartTime,
		UserID:    notification.UserID,
		CreatedAt: time.Now(),
	}

	if err := s.storage.SaveNotification(ctx, dbNotification); err != nil {
		return fmt.Errorf("failed to save notification: %w", err)
	}

	s.logger.Info("notification saved",
		"event_id", notification.EventID,
		"user_id", notification.UserID,
	)
	return nil
}
