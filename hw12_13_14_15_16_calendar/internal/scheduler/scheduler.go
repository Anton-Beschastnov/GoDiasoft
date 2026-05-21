package scheduler

import (
	"context"
	"time"

	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/app"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/kafka"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/logger"
)

// Config конфигурация scheduler
type Config struct {
	ScanInterval time.Duration `yaml:"scan_interval"`
	KafkaTopic   string        `yaml:"kafka_topic"`
}

// Scheduler процесс для отправки уведомлений
type Scheduler struct {
	logger     logger.LoggerI
	storage    app.Storage
	producer   kafka.ProducerInterface
	config     Config
	cutoffTime time.Time
}

// New создает новый scheduler
func New(logger logger.LoggerI, storage app.Storage, producer kafka.ProducerInterface, config Config) *Scheduler {
	return &Scheduler{
		logger:     logger,
		storage:    storage,
		producer:   producer,
		config:     config,
		cutoffTime: time.Now().AddDate(-1, 0, 0),
	}
}

// Run запускает scheduler
func (s *Scheduler) Run(ctx context.Context) error {
	s.logger.Info("scheduler is running", "scan_interval", s.config.ScanInterval, "kafka_topic", s.config.KafkaTopic)

	ticker := time.NewTicker(s.config.ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("scheduler stopped")
			return nil
		case <-ticker.C:
			if err := s.runOnce(ctx); err != nil {
				s.logger.Error("scheduler error", "error", err)
			}
		}
	}
}

// runOnce выполняет один цикл обработки
func (s *Scheduler) runOnce(ctx context.Context) error {
	now := time.Now()

	// Получаем события для уведомления
	events, err := s.storage.GetEventsForNotification(ctx, now)
	if err != nil {
		return err
	}

	// Отправляем уведомления в Kafka
	for _, event := range events {
		notification := &kafka.Notification{
			EventID:   event.ID,
			Title:     event.Title,
			StartTime: event.StartTime,
			UserID:    event.UserID,
		}

		topic := s.config.KafkaTopic
		if topic == "" {
			topic = "calendar_notifications" // default topic
		}

		if err := s.producer.Send(ctx, topic, notification); err != nil {
			s.logger.Error("failed to send notification", "event_id", event.ID, "user_id", event.UserID, "error", err)
		} else {
			s.logger.Info("sent notification", "event_id", event.ID, "user_id", event.UserID)
		}
	}

	// Удаляем старые события
	if err := s.storage.DeleteOldEvents(ctx, s.cutoffTime); err != nil {
		s.logger.Error("failed to delete old events", "error", err)
	}

	return nil
}
