package storer

import (
	"context"
	"fmt"
	"time"

	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/app"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/kafka"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/metrics"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/storage"
)

// Config конфигурация storer.
type Config struct {
	KafkaTopic string `yaml:"kafkaTopic"`
}

// Storer процесс для сохранения уведомлений.
type Storer struct {
	logger   logger.Iface
	storage  app.Storage
	consumer kafka.ConsumerInterface
	config   Config
}

// New создает новый storer.
func New(logger logger.Iface, storage app.Storage, consumer kafka.ConsumerInterface, config Config) *Storer {
	return &Storer{
		logger:   logger,
		storage:  storage,
		consumer: consumer,
		config:   config,
	}
}

// Run запускает storer.
func (s *Storer) Run(ctx context.Context) error {
	topic := s.config.KafkaTopic
	if topic == "" {
		topic = "calendar_notifications" // default topic
	}
	s.logger.Info("storer is running", "kafka_topic", topic)

	// Обработчик уведомлений.
	handler := func(notification *kafka.Notification) error {
		return s.saveNotification(ctx, notification)
	}

	return s.consumer.Consume(ctx, topic, handler)
}

// saveNotification сохраняет уведомление в базу данных.
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
		metrics.NotificationsSaved.WithLabelValues("error").Inc()
		return fmt.Errorf("failed to save notification: %w", err)
	}

	metrics.NotificationsSaved.WithLabelValues("success").Inc()
	s.logger.Info("notification saved",
		"event_id", notification.EventID,
		"user_id", notification.UserID,
	)
	return nil
}
