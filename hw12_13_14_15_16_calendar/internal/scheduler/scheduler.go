package scheduler

import (
	"context"
	"time"

	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/app"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/kafka"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/logger"
)

type Config struct {
	ScanInterval time.Duration `yaml:"scanInterval"`
	KafkaTopic   string        `yaml:"kafkaTopic"`
}

type Scheduler struct {
	logger     logger.Iface
	storage    app.Storage
	producer   kafka.ProducerInterface
	config     Config
	cutoffTime time.Time
}

func New(logger logger.Iface, storage app.Storage, producer kafka.ProducerInterface, config Config) *Scheduler {
	return &Scheduler{
		logger:     logger,
		storage:    storage,
		producer:   producer,
		config:     config,
		cutoffTime: time.Now().AddDate(-1, 0, 0),
	}
}

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
			
			if err := s.runOnce(ctx, time.Now().UTC()); err != nil {
				s.logger.Error("scheduler error", "error", err)
			}
		}
	}
}

func (s *Scheduler) runOnce(ctx context.Context, now time.Time) error {
	s.logger.Info("--- Starting scheduler scan ---", "scan_time_utc", now)

	events, err := s.storage.GetEventsForNotification(ctx, now)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		s.logger.Info("No events to notify found.")
		return nil
	}

	s.logger.Info("Found events for notification", "count", len(events))

	for _, event := range events {
		notification := &kafka.Notification{
			EventID:   event.ID,
			Title:     event.Title,
			StartTime: event.StartTime,
			UserID:    event.UserID,
		}

		topic := s.config.KafkaTopic
		if topic == "" {
			topic = "calendar_notifications" 
		}

		if err := s.producer.Send(ctx, topic, notification); err != nil {
			s.logger.Error("failed to send notification", "event_id", event.ID, "user_id", event.UserID, "error", err)
		} else {
			s.logger.Info("SENT notification", "event_id", event.ID, "user_id", event.UserID)
		}
	}

	if err := s.storage.DeleteOldEvents(ctx, s.cutoffTime); err != nil {
		s.logger.Error("failed to delete old events", "error", err)
	}

	return nil
}
